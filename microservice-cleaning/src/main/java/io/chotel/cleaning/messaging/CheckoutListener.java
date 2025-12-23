package io.chotel.cleaning.messaging;

import io.chotel.cleaning.messaging.json.CheckinMessageJson;
import io.chotel.cleaning.messaging.json.CheckoutMessageJson;
import io.chotel.cleaning.service.RoomOccupancyService;
import io.micronaut.configuration.kafka.annotation.KafkaListener;
import io.micronaut.configuration.kafka.annotation.OffsetReset;
import io.micronaut.configuration.kafka.annotation.Topic;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.SpanKind;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.propagation.TextMapGetter;
import jakarta.inject.Inject;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerRecord;

import java.util.HashMap;
import java.util.Map;

// https://micronaut-projects.github.io/micronaut-kafka/latest/guide/

@Slf4j
@KafkaListener(groupId = "cleaning", offsetReset = OffsetReset.EARLIEST)
@RequiredArgsConstructor
public class CheckoutListener {

    private static final TextMapGetter<Map<String, String>> KAFKA_HEADER_GETTER = 
        new TextMapGetter<>() {
            @Override
            public Iterable<String> keys(Map<String, String> carrier) {
                return carrier.keySet();
            }

            @Override
            public String get(Map<String, String> carrier, String key) {
                return carrier.get(key);
            }
        };

    private final RoomOccupancyService roomOccupancyService;
    
    @Inject
    private Tracer tracer;

    @Inject
    private final OpenTelemetry openTelemetry;

    @Topic("check-in")
    public void receiveCheckinMessage(ConsumerRecord<String, CheckinMessageJson> record) {
        // Extract trace context from Kafka headers
        Map<String, String> headers = extractHeaders(record);
        log.info("Extracted Kafka headers: {}", headers);
        
        // Log what propagators are configured
        var propagators = openTelemetry.getPropagators();
        log.info("Configured propagators: {}", propagators.getClass().getName());
        
        // Extract the remote context from headers
        Context extractedContext = propagators
            .getTextMapPropagator()
            .extract(Context.current(), headers, KAFKA_HEADER_GETTER);

        // Log the extracted span context to see if we got the trace ID
        Span extractedSpan = Span.fromContext(extractedContext);
        log.info("Extracted span context - trace_id={} span_id={} isValid={} isSampled={}", 
            extractedSpan.getSpanContext().getTraceId(), 
            extractedSpan.getSpanContext().getSpanId(),
            extractedSpan.getSpanContext().isValid(),
            extractedSpan.getSpanContext().isSampled());

        // Create a new span that is a child of the extracted context
        Span span = tracer.spanBuilder("process check-in message")
            .setParent(extractedContext)
            .setSpanKind(SpanKind.CONSUMER)
            .setAttribute("messaging.system", "kafka")
            .setAttribute("messaging.destination", "check-in")
            .setAttribute("messaging.operation", "process")
            .startSpan();

        log.info("Created span with trace_id={} span_id={} parent_span_id={}", 
            span.getSpanContext().getTraceId(), 
            span.getSpanContext().getSpanId(),
            extractedSpan.getSpanContext().getSpanId());

        // Make the span current - this ensures all child operations inherit the trace context
        try (var scope = extractedContext.with(span).makeCurrent()) {
            CheckinMessageJson body = record.value();
            log.info("Received check-in message: {}", body);
            roomOccupancyService.onCheckinEvent(body.getRoomNumber(), body.getGuestId(), body.getTimestamp());
        } catch (Exception e) {
            span.recordException(e);
            throw e;
        } finally {
            span.end();
        }
    }

    @Topic("check-out")
    public void receiveCheckoutMessage(ConsumerRecord<String, CheckoutMessageJson> record) {
        // Extract trace context from Kafka headers
        Map<String, String> headers = extractHeaders(record);
        log.info("Extracted Kafka headers: {}", headers);
        
        // Extract the remote context from headers
        Context extractedContext = openTelemetry.getPropagators()
            .getTextMapPropagator()
            .extract(Context.current(), headers, KAFKA_HEADER_GETTER);

        // Log the extracted span context to see if we got the trace ID
        Span extractedSpan = Span.fromContext(extractedContext);
        log.info("Extracted span context - trace_id={} span_id={} isValid={}", 
            extractedSpan.getSpanContext().getTraceId(), 
            extractedSpan.getSpanContext().getSpanId(),
            extractedSpan.getSpanContext().isValid());

        // Create a new span that is a child of the extracted context
        Span span = tracer.spanBuilder("process check-out message")
            .setParent(extractedContext)
            .setSpanKind(SpanKind.CONSUMER)
            .setAttribute("messaging.system", "kafka")
            .setAttribute("messaging.destination", "check-out")
            .setAttribute("messaging.operation", "process")
            .startSpan();

        log.info("Created span with trace_id={} span_id={} parent_span_id={}", 
            span.getSpanContext().getTraceId(), 
            span.getSpanContext().getSpanId(),
            extractedSpan.getSpanContext().getSpanId());

        // Make the span current - this ensures all child operations inherit the trace context
        try (var scope = extractedContext.with(span).makeCurrent()) {
            CheckoutMessageJson body = record.value();
            log.info("Received check-out message: {}", body);
            roomOccupancyService.onCheckoutEvent(body.getRoomNumber(), body.getGuestId(), body.getTimestamp());
        } catch (Exception e) {
            span.recordException(e);
            throw e;
        } finally {
            span.end();
        }
    }

    private Map<String, String> extractHeaders(ConsumerRecord<?, ?> record) {
        Map<String, String> headers = new HashMap<>();
        record.headers().forEach(header -> {
            if (header.value() != null) {
                headers.put(header.key(), new String(header.value()));
            }
        });
        return headers;
    }
}
