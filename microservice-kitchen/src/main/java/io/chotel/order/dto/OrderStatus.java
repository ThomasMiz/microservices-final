package io.chotel.order.dto;

public enum OrderStatus {
    // Kitchen internal states
    PENDING,        // Order received, stock reserved, waiting for billing confirmation
    IN_PREPARATION, // Billing confirmed, order is being prepared
    COMPLETED,      // Order preparation finished
    CANCELLED       // Order cancelled (billing failed or other reason)
}
