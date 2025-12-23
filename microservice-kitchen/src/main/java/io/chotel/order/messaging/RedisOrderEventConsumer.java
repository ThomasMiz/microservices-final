package io.chotel.order.messaging;

import com.fasterxml.jackson.databind.ObjectMapper;

import io.chotel.order.OrderService;
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
import io.quarkus.redis.datasource.stream.*;
import io.quarkus.runtime.ShutdownEvent;
import io.quarkus.runtime.StartupEvent;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.enterprise.event.Observes;
import jakarta.inject.Inject;

import org.eclipse.microprofile.context.ManagedExecutor;

import java.time.Duration;
import java.util.List;
import java.util.Map;

/** Redis Streams consumer for receiving order requests from Room Service */
@ApplicationScoped
public class RedisOrderEventConsumer {

    @Inject
    RedisDataSource redisDataSource;

    @Inject
    OrderService orderService;

    @Inject
    ObjectMapper objectMapper;

    @Inject
    OpenTelemetry openTelemetry;

    @Inject
    Tracer tracer;

    @Inject
    ManagedExecutor managedExecutor;

    private final RedisMessageTextMapGetter textMapGetter = new RedisMessageTextMapGetter();
    private final String stream;
    private final String consumerGroup;
    private final String consumerName;
    private StreamCommands<String, String, String> streamCommands;
    private volatile boolean running = false;

    public RedisOrderEventConsumer() {
        this.stream = System.getenv().getOrDefault("KITCHEN_REQUESTS_STREAM", "kitchen-requests");
        this.consumerGroup = System.getenv().getOrDefault("KITCHEN_CONSUMER_GROUP", "kitchen-group");
        this.consumerName = "kitchen-consumer-" + System.currentTimeMillis();
    }

    void onStart(@Observes StartupEvent ev) {
        streamCommands = redisDataSource.stream(String.class);

        try {
            // 1. Use 'mkstream' to ensure the stream key exists, avoiding the crash
            XGroupCreateArgs args = new XGroupCreateArgs().mkstream();

            // 2. Create the group.
            // '0' = process all existing messages (including the ones you already sent)
            // '$' = process only new messages arriving from now on
            streamCommands.xgroupCreate(stream, consumerGroup, "0", args);

            Log.infof("Created consumer group %s for stream %s", consumerGroup, stream);

        } catch (Exception e) {
            // 3. Only ignore the error if it is specifically "BUSYGROUP" (already exists)
            if (e.getMessage().contains("BUSYGROUP")) {
                Log.infof("Consumer group %s already exists. Resuming.", consumerGroup);
            } else {
                // 4. CRITICAL: If it's any other error, CRASH or Log ERROR.
                // Do not pretend it succeeded.
                Log.errorf(e, "FATAL: Failed to create consumer group %s!", consumerGroup);
                throw new RuntimeException("Could not initialize Redis Stream", e);
            }
        }

        // ... start executor service ...
        running = true;
        managedExecutor.submit(this::consumeMessages);

        Log.infof("Redis consumer started: stream=%s, group=%s", stream, consumerGroup);
    }

    void onStop(@Observes ShutdownEvent ev) {
        running = false;
        Log.info("Redis consumer stopped");
    }

    private void consumeMessages() {
        while (running) {
            try {
                // Read messages from the stream using consumer group
                XReadGroupArgs args = new XReadGroupArgs();
                args.count(10);
                args.block(Duration.ofSeconds(1));

                List<StreamMessage<String, String, String>> messages = streamCommands.xreadgroup(
                        consumerGroup,
                        consumerName,
                        stream,
                        ">", // Read only new messages
                        args);

                for (StreamMessage<String, String, String> message : messages) {
                    processMessage(message);
                }
            } catch (Exception e) {
                if (running) {
                    Log.errorf(e, "Error reading from Redis stream");
                    try {
                        Thread.sleep(1000); // Wait before retrying
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            }
        }
    }

    private void processMessage(StreamMessage<String, String, String> message) {
        Map<String, String> payload = message.payload();
        
        // Extract trace context from message
        TextMapPropagator textMapPropagator = openTelemetry.getPropagators().getTextMapPropagator();
        Context extractedContext = textMapPropagator.extract(
            Context.current(),
            payload,
            textMapGetter
        );
        
        // Create span linked to parent trace
        Span span = tracer.spanBuilder("redis process order event")
            .setParent(extractedContext)
            .setSpanKind(SpanKind.CONSUMER)
            .setAttribute("messaging.system", "redis")
            .setAttribute("messaging.source", stream)
            .setAttribute("messaging.operation", "process")
            .startSpan();
        
        // Make both the context AND span current using Context.makeCurrent()
        Context contextWithSpan = extractedContext.with(span);
        try (Scope scope = contextWithSpan.makeCurrent()) {
            String dataJson = payload.get("data");

            if (dataJson == null) {
                Log.warn("Message missing 'data' field: " + message.id());
                streamCommands.xack(stream, consumerGroup, message.id());
                span.setStatus(StatusCode.ERROR, "Missing data field");
                return;
            }

            OrderEvent event = objectMapper.readValue(dataJson, OrderEvent.class);
            span.setAttribute("messaging.message_id", event.orderId());
            span.setAttribute("order.event_type", event.eventType());
            
            Log.infof(
                    "Received order event: type=%s, orderId=%s",
                    event.eventType(), event.orderId());

            // Process the event (this will create child spans)
            orderService.handleOrderEvent(event);

            // Acknowledge the message
            streamCommands.xack(stream, consumerGroup, message.id());
            span.setStatus(StatusCode.OK);
            Log.debugf("Acknowledged message: %s", message.id());

        } catch (Exception e) {
            span.recordException(e);
            span.setStatus(StatusCode.ERROR, "Error processing message");
            Log.errorf(e, "Error processing message: %s", message.id());
            // Don't acknowledge - message will be retried
        } finally {
            span.end();
        }
    }
}
