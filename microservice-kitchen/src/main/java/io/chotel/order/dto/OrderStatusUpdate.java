package io.chotel.order.dto;

import java.time.Instant;

public record OrderStatusUpdate(String id, String status, Instant updatedAt) {
    public OrderStatusUpdate(String id, OrderStatus status) {
        this(id, status.name(), Instant.now());
    }
}
