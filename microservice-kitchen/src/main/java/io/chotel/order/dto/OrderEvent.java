package io.chotel.order.dto;

import com.fasterxml.jackson.annotation.JsonAlias;
import com.fasterxml.jackson.annotation.JsonProperty;

import io.quarkus.runtime.annotations.RegisterForReflection;

import java.time.Instant;
import java.util.List;

/**
 * Represents an order event in the orchestration flow.
 * Event types:
 * - OrderRequested: Initial order request from Room Service
 * - OrderAccepted: Kitchen accepted the order (stock reserved)
 * - OrderRejected: Kitchen rejected the order (insufficient stock)
 * - OrderConfirmed: Billing succeeded, order ready for preparation
 * - OrderBillingFailed: Billing failed, need to restore stock
 * - OrderPreparationCompleted: Kitchen finished preparing the order
 * - OrderPreparationFailed: Critical failure in kitchen after confirmation
 */
@RegisterForReflection
public record OrderEvent(
    @JsonProperty("event_type")
    @JsonAlias({"eventType"})
    String eventType,
    @JsonProperty("order_id")
    @JsonAlias({"orderId"})
    String orderId,
    @JsonProperty("room_id")
    @JsonAlias({"roomId"})
    String roomId,
    @JsonProperty("reservation_id")
    @JsonAlias({"reservationId"})
    String reservationId,
    @JsonProperty("items")
    List<Item> items,
    @JsonProperty("total_price")
    @JsonAlias({"totalPrice"})
    String totalPrice,
    @JsonProperty("billing_folder_id")
    @JsonAlias({"billingFolderId"})
    String billingFolderId,
    @JsonProperty("reason")
    String reason,
    @JsonProperty("timestamp")
    Instant timestamp
) {
    public static OrderEvent orderAccepted(Order order) {
        return new OrderEvent(
            "OrderAccepted",
            order.id(),
            order.roomId(),
            order.reservationId(),
            order.items(),
            order.totalPrice(),
            null,
            null,
            Instant.now()
        );
    }

    public static OrderEvent orderRejected(Order order, String reason) {
        return new OrderEvent(
            "OrderRejected",
            order.id(),
            order.roomId(),
            order.reservationId(),
            order.items(),
            order.totalPrice(),
            null,
            reason,
            Instant.now()
        );
    }

    public static OrderEvent orderPreparationCompleted(Order order) {
        return new OrderEvent(
            "OrderPreparationCompleted",
            order.id(),
            order.roomId(),
            order.reservationId(),
            order.items(),
            order.totalPrice(),
            null,
            null,
            Instant.now()
        );
    }

    public static OrderEvent orderPreparationFailed(Order order, String reason) {
        return new OrderEvent(
            "OrderPreparationFailed",
            order.id(),
            order.roomId(),
            order.reservationId(),
            order.items(),
            order.totalPrice(),
            null,
            reason,
            Instant.now()
        );
    }

    public boolean isOrderRequested() {
        return "OrderRequested".equals(eventType);
    }

    public boolean isOrderConfirmed() {
        return "OrderConfirmed".equals(eventType);
    }

    public boolean isOrderBillingFailed() {
        return "OrderBillingFailed".equals(eventType);
    }
}
