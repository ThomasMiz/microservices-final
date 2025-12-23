package io.chotel.order.messaging;

import io.opentelemetry.context.propagation.TextMapGetter;
import java.util.Map;

/**
 * TextMapGetter implementation for extracting trace context from Redis Stream messages.
 * This enables distributed tracing across services communicating via Redis Streams.
 */
public class RedisMessageTextMapGetter implements TextMapGetter<Map<String, String>> {
    
    @Override
    public Iterable<String> keys(Map<String, String> carrier) {
        return carrier.keySet();
    }
    
    @Override
    public String get(Map<String, String> carrier, String key) {
        if (carrier == null) {
            return null;
        }
        return carrier.get(key);
    }
}
