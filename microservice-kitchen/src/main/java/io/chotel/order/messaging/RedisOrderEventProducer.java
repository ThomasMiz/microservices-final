package io.chotel.order.messaging;

import io.chotel.order.dto.OrderEvent;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.StatusCode;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.Scope;
import io.opentelemetry.context.propagation.TextMapPropagator;
import io.quarkus.logging.Log;
import io.quarkus.redis.datasource.RedisDataSource;
import io.quarkus.redis.datasource.stream.StreamCommands;
import io.quarkus.redis.datasource.stream.XAddArgs;
import io.quarkus.runtime.Startup;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;

import java.util.HashMap;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Redis Streams producer for publishing order events to Room Service
 */
@ApplicationScoped
@Startup
public class RedisOrderEventProducer {

    @Inject
    RedisDataSource redisDataSource;

    @Inject
    ObjectMapper objectMapper;

    @Inject
    OpenTelemetry openTelemetry;

    @Inject
    Tracer tracer;

    private final RedisMessageTextMapSetter textMapSetter = new RedisMessageTextMapSetter();
    private final String stream;
    private StreamCommands<String, String, String> streamCommands;

    public RedisOrderEventProducer() {
        // Stream name will be configured via application.properties
        this.stream = System.getenv().getOrDefault("KITCHEN_RESPONSES_STREAM", "kitchen-responses");
    }

    @jakarta.annotation.PostConstruct
    void initialize() {
        streamCommands = redisDataSource.stream(String.class);
        Log.infof("Redis producer initialized for stream: %s", stream);
    }

    /**
     * Publishes an order event to Redis Streams with trace context propagation
     */
    public void publish(OrderEvent event) {
        // Create span for publishing
        Span span = tracer.spanBuilder("redis publish order event")
            .setSpanKind(SpanKind.PRODUCER)
            .setAttribute("messaging.system", "redis")
            .setAttribute("messaging.destination", stream)
            .setAttribute("messaging.operation", "publish")
            .setAttribute("messaging.message_id", event.orderId())
            .setAttribute("order.event_type", event.eventType())
            .startSpan();
        
        try (Scope scope = span.makeCurrent()) {
            String eventJson = objectMapper.writeValueAsString(event);
            
            // Create mutable map for trace context injection
            Map<String, String> values = new HashMap<>();
            values.put("order_id", event.orderId());
            values.put("event_type", event.eventType());
            values.put("data", eventJson);
            
            // Inject trace context into message
            Context current = Context.current();
            TextMapPropagator textMapPropagator = openTelemetry.getPropagators().getTextMapPropagator();
            textMapPropagator.inject(current, values, textMapSetter);
            
            // Publish to Redis
            streamCommands.xadd(stream, values);
            
            span.setStatus(StatusCode.OK);
            Log.infof("Published %s event for order %s with trace context to Redis stream %s", 
                event.eventType(), event.orderId(), stream);
        } catch (Exception e) {
            span.recordException(e);
            span.setStatus(StatusCode.ERROR, "Failed to publish event");
            Log.errorf(e, "Failed to publish event %s for order %s", 
                event.eventType(), event.orderId());
            throw new RuntimeException("Failed to publish event", e);
        } finally {
            span.end();
        }
    }
}
