package io.chotel.order.messaging;

import io.opentelemetry.context.propagation.TextMapSetter;
import java.util.Map;

/**
 * TextMapSetter implementation for injecting trace context into Redis Stream messages.
 * This enables distributed tracing across services communicating via Redis Streams.
 */
public class RedisMessageTextMapSetter implements TextMapSetter<Map<String, String>> {
    
    @Override
    public void set(Map<String, String> carrier, String key, String value) {
        if (carrier != null && key != null && value != null) {
            carrier.put(key, value);
        }
    }
}
