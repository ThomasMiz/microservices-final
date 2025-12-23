package redis

import (
	"context"
	"encoding/json"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"microservice-roomservice/internal/infrastructure/messaging"
	"microservice-roomservice/pkg/logger"
)

// Producer handles publishing messages to Redis Streams
type Producer struct {
	client *redisclient.Client
	stream string
}

// NewProducer creates a new Redis producer
func NewProducer(addr, password string, stream string) *Producer {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &Producer{
		client: client,
		stream: stream,
	}
}

// PublishOrder publishes an order message to Redis with retry mechanism
func (p *Producer) PublishOrder(ctx context.Context, message messaging.OrderMessage) error {
	tracer := otel.Tracer("redis-producer")
	ctx, span := tracer.Start(ctx, "redis publish")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "redis"),
		attribute.String("messaging.destination", p.stream),
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.message_id", message.OrderID),
	)

	spanCtx := span.SpanContext()
	logger.CtxDebug(ctx).Msgf("RedisProducer.PublishOrder - Created span: traceID=%s spanID=%s isRecording=%v",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		span.IsRecording())

	data, err := json.Marshal(message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		return err
	}

	// Create Redis stream message with trace context
	values := map[string]interface{}{
		"order_id": message.OrderID,
		"data":     string(data),
	}

	// Inject trace context into message
	carrier := &RedisMessageCarrier{values: values}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	logger.CtxDebug(ctx).Msg("RedisProducer.PublishOrder - Injected trace context into Redis message")

	// Retry mechanism with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		_, err := p.client.XAdd(ctx, &redisclient.XAddArgs{
			Stream: p.stream,
			Values: values,
		}).Result()

		if err == nil {
			span.SetStatus(codes.Ok, "message published successfully")
			logger.CtxDebug(ctx).Msg("RedisProducer.PublishOrder - Message published successfully")
			return nil
		}
		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, "context cancelled")
			return ctx.Err()
		}

		// Exponential backoff: 100ms, 400ms, 1.6s
		backoff := time.Duration(100) * time.Millisecond * time.Duration(1<<uint(attempt))
		select {
		case <-time.After(backoff):
			continue
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, "context done")
			return ctx.Err()
		}
	}

	span.RecordError(lastErr)
	span.SetStatus(codes.Error, "failed to publish after retries")
	return lastErr
}

// PublishOrderEvent publishes an order event to Redis for orchestration
func (p *Producer) PublishOrderEvent(ctx context.Context, event messaging.OrderEvent) error {
	tracer := otel.Tracer("redis-producer")
	ctx, span := tracer.Start(ctx, "redis publish order event")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "redis"),
		attribute.String("messaging.destination", p.stream),
		attribute.String("messaging.operation", "publish"),
		attribute.String("messaging.message_id", event.OrderID),
		attribute.String("order.event_type", event.EventType),
	)

	logger.CtxDebug(ctx).Msgf("Publishing %s event for order %s", event.EventType, event.OrderID)

	data, err := json.Marshal(event)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal event")
		return err
	}

	// Create Redis stream message
	values := map[string]interface{}{
		"order_id":   event.OrderID,
		"event_type": event.EventType,
		"data":       string(data),
	}

	// Inject trace context into message
	carrier := &RedisMessageCarrier{values: values}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	// Retry mechanism with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		_, err := p.client.XAdd(ctx, &redisclient.XAddArgs{
			Stream: p.stream,
			Values: values,
		}).Result()

		if err == nil {
			span.SetStatus(codes.Ok, "event published successfully")
			logger.CtxDebug(ctx).Msgf("Published %s event for order %s", event.EventType, event.OrderID)
			return nil
		}
		lastErr = err

		if ctx.Err() != nil {
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, "context cancelled")
			return ctx.Err()
		}

		backoff := time.Duration(100) * time.Millisecond * time.Duration(1<<uint(attempt))
		select {
		case <-time.After(backoff):
			continue
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			span.SetStatus(codes.Error, "context done")
			return ctx.Err()
		}
	}

	span.RecordError(lastErr)
	span.SetStatus(codes.Error, "failed to publish event after retries")
	return lastErr
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.client.Close()
}
