package redis

import (
	"context"
	"encoding/json"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"microservice-roomservice/internal/infrastructure/messaging"
	"microservice-roomservice/pkg/logger"
)

// Consumer handles consuming messages from Redis Streams
type Consumer struct {
	client     *redisclient.Client
	stream     string
	group      string
	consumerID string
	handler    MessageHandler
}

// MessageHandler defines the interface for handling messages
type MessageHandler interface {
	HandleOrderCompletion(ctx context.Context, message messaging.OrderCompletionMessage) error
}

// OrderEventHandler defines the interface for handling orchestration events
type OrderEventHandler interface {
	HandleOrderEvent(ctx context.Context, event messaging.OrderEvent) error
}

// NewConsumer creates a new Redis consumer
func NewConsumer(addr, password, stream, group string, handler MessageHandler) *Consumer {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	consumerID := "roomservice-" + time.Now().Format("20060102150405")

	return &Consumer{
		client:     client,
		stream:     stream,
		group:      group,
		consumerID: consumerID,
		handler:    handler,
	}
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	logger.CtxInfo(ctx).Msgf("Starting Redis consumer for stream: %s", c.stream)

	// Create consumer group if it doesn't exist
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logger.CtxError(ctx).Err(err).Msg("Error creating consumer group")
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.CtxInfo(ctx).Msg("Stopping Redis consumer")
			return c.client.Close()
		default:
			// Read messages from the stream
			streams, err := c.client.XReadGroup(ctx, &redisclient.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumerID,
				Streams:  []string{c.stream, ">"},
				Count:    10,
				Block:    1 * time.Second,
			}).Result()

			if err != nil {
				if err == redisclient.Nil {
					// No messages available
					continue
				}
				logger.CtxError(ctx).Err(err).Msg("Error reading from stream")
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := c.processMessage(ctx, message); err != nil {
						logger.CtxError(ctx).Err(err).Msg("Error processing message")
						continue
					}

					// Acknowledge the message
					if err := c.client.XAck(ctx, c.stream, c.group, message.ID).Err(); err != nil {
						logger.CtxError(ctx).Err(err).Msg("Error acknowledging message")
					}
				}
			}
		}
	}
}

// processMessage processes a single message
func (c *Consumer) processMessage(ctx context.Context, msg redisclient.XMessage) error {
	// Extract trace context from Redis message
	carrier := &RedisMessageCarrier{values: msg.Values}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Check if we extracted a valid parent span
	parentSpan := trace.SpanFromContext(ctx)
	parentSpanCtx := parentSpan.SpanContext()
	logger.CtxDebug(ctx).Msgf("RedisConsumer.processMessage - Extracted parent span: traceID=%s spanID=%s isValid=%v",
		parentSpanCtx.TraceID().String(),
		parentSpanCtx.SpanID().String(),
		parentSpanCtx.IsValid())

	tracer := otel.Tracer("redis-consumer")
	ctx, span := tracer.Start(ctx, "redis process")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "redis"),
		attribute.String("messaging.destination", c.stream),
		attribute.String("messaging.operation", "process"),
	)

	spanCtx := span.SpanContext()
	logger.CtxDebug(ctx).Msgf("RedisConsumer.processMessage - Created span: traceID=%s spanID=%s isRecording=%v",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		span.IsRecording())

	// Extract data from message
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		span.SetStatus(codes.Error, "data field not found in message")
		return nil
	}

	var orderCompletion messaging.OrderCompletionMessage
	if err := json.Unmarshal([]byte(dataStr), &orderCompletion); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal message")
		return err
	}

	span.SetAttributes(attribute.String("messaging.message_id", orderCompletion.OrderID))

	if err := c.handler.HandleOrderCompletion(ctx, orderCompletion); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "handler failed")
		return err
	}

	span.SetStatus(codes.Ok, "message processed successfully")
	return nil
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.client.Close()
}

// OrderEventConsumer handles consuming order events from Redis for orchestration
type OrderEventConsumer struct {
	client     *redisclient.Client
	stream     string
	group      string
	consumerID string
	handler    OrderEventHandler
}

// NewOrderEventConsumer creates a new Redis consumer for order events
func NewOrderEventConsumer(addr, password, stream, group string, handler OrderEventHandler) *OrderEventConsumer {
	client := redisclient.NewClient(&redisclient.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	consumerID := "roomservice-events-" + time.Now().Format("20060102150405")

	return &OrderEventConsumer{
		client:     client,
		stream:     stream,
		group:      group,
		consumerID: consumerID,
		handler:    handler,
	}
}

// Start starts consuming order events
func (c *OrderEventConsumer) Start(ctx context.Context) error {
	logger.CtxInfo(ctx).Msgf("Starting Order Event consumer for stream: %s", c.stream)

	// Create consumer group if it doesn't exist
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logger.CtxError(ctx).Err(err).Msg("Error creating consumer group")
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.CtxInfo(ctx).Msg("Stopping Order Event consumer")
			return c.client.Close()
		default:
			// Read messages from the stream
			streams, err := c.client.XReadGroup(ctx, &redisclient.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumerID,
				Streams:  []string{c.stream, ">"},
				Count:    10,
				Block:    1 * time.Second,
			}).Result()

			if err != nil {
				if err == redisclient.Nil {
					// No messages available
					continue
				}
				logger.CtxError(ctx).Err(err).Msg("Error reading order event from stream")
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					if err := c.processOrderEvent(ctx, message); err != nil {
						logger.CtxError(ctx).Err(err).Msg("Error processing order event")
						continue
					}

					// Acknowledge the message
					if err := c.client.XAck(ctx, c.stream, c.group, message.ID).Err(); err != nil {
						logger.CtxError(ctx).Err(err).Msg("Error acknowledging order event")
					}
				}
			}
		}
	}
}

// processOrderEvent processes a single order event
func (c *OrderEventConsumer) processOrderEvent(ctx context.Context, msg redisclient.XMessage) error {
	carrier := &RedisMessageCarrier{values: msg.Values}
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	tracer := otel.Tracer("redis-consumer")
	ctx, span := tracer.Start(ctx, "redis process order event")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "redis"),
		attribute.String("messaging.destination", c.stream),
		attribute.String("messaging.operation", "process"),
	)

	// Extract data from message
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		span.SetStatus(codes.Error, "data field not found in message")
		return nil
	}

	var event messaging.OrderEvent
	if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal order event")
		return err
	}

	span.SetAttributes(
		attribute.String("messaging.message_id", event.OrderID),
		attribute.String("order.event_type", event.EventType),
	)

	logger.CtxInfo(ctx).Msgf("Processing order event: type=%s, orderId=%s", event.EventType, event.OrderID)

	if err := c.handler.HandleOrderEvent(ctx, event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "handler failed")
		return err
	}

	span.SetStatus(codes.Ok, "order event processed successfully")
	return nil
}

// Close closes the consumer
func (c *OrderEventConsumer) Close() error {
	return c.client.Close()
}
