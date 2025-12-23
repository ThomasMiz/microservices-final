package io.chotel.order;

import io.chotel.order.dto.Item;
import io.chotel.order.dto.Order;
import io.chotel.order.dto.OrderEvent;
import io.chotel.order.dto.OrderStatus;
import io.chotel.order.messaging.RedisOrderEventProducer;
import io.opentelemetry.instrumentation.annotations.WithSpan;
import io.quarkus.logging.Log;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import jakarta.ws.rs.NotFoundException;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@ApplicationScoped
public class OrderService {

    @Inject
    OrderRepository orderRepository;

    @Inject
    InventoryRepository inventoryRepository;

    @Inject
    RedisOrderEventProducer orderEventProducer;

    /**
     * Handles incoming OrderRequested events from Room Service.
     * Attempts to reserve stock atomically and responds with OrderAccepted or OrderRejected.
     */
    @WithSpan("order.handleOrderEvent")
    public void handleOrderEvent(OrderEvent event) {
        Log.infof("Received order event: type=%s, orderId=%s", event.eventType(), event.orderId());

        try {
            if (event.isOrderRequested()) {
                handleOrderRequested(event);
            } else if (event.isOrderConfirmed()) {
                handleOrderConfirmed(event);
            } else if (event.isOrderBillingFailed()) {
                handleOrderBillingFailed(event);
            } else {
                Log.warnf("Unknown event type: %s", event.eventType());
            }
        } catch (Exception e) {
            Log.errorf(e, "Error handling order event: %s", event.orderId());
            throw e; // Re-throw to trigger Kafka retry/DLQ
        }
    }

    /**
     * Handles OrderRequested: Try to reserve stock, respond with Accept/Reject.
     */
    @WithSpan("order.handleOrderRequested")
    public void handleOrderRequested(OrderEvent event) {
        Log.infof("Processing OrderRequested for order %s", event.orderId());

        // Try to reserve stock for all items
        List<Item> reservedItems = new ArrayList<>();
        boolean allReserved = true;
        String failureReason = null;

        for (Item item : event.items()) {
            boolean reserved = inventoryRepository.reserveStock(item.menuItemId(), item.quantity());
            if (reserved) {
                reservedItems.add(item);
            } else {
                allReserved = false;
                failureReason = "Insufficient stock for item: " + item.menuItemId();
                break;
            }
        }

        if (!allReserved) {
            // Rollback any reservations we made
            for (Item item : reservedItems) {
                inventoryRepository.restoreStock(item.menuItemId(), item.quantity());
            }

            // Publish OrderRejected
            Order rejectedOrder = createOrderFromEvent(event, OrderStatus.CANCELLED);
            OrderEvent rejectedEvent = OrderEvent.orderRejected(rejectedOrder, failureReason);
            publishOrderEvent(rejectedEvent);
            Log.infof("Order %s rejected: %s", event.orderId(), failureReason);
            return;
        }

        // All stock reserved - create order in PENDING state and publish OrderAccepted
        Order order = createOrderFromEvent(event, OrderStatus.PENDING);
        orderRepository.createOrder(order);

        OrderEvent acceptedEvent = OrderEvent.orderAccepted(order);
        publishOrderEvent(acceptedEvent);
        Log.infof("Order %s accepted, stock reserved, waiting for billing confirmation", event.orderId());
    }

    /**
     * Handles OrderConfirmed: Billing succeeded, start preparation.
     */
    @WithSpan("order.handleOrderConfirmed")
    public void handleOrderConfirmed(OrderEvent event) {
        Log.infof("Processing OrderConfirmed for order %s", event.orderId());

        Optional<Order> orderOpt = orderRepository.getOrder(event.orderId());
        if (orderOpt.isEmpty()) {
            Log.warnf("Order %s not found for confirmation", event.orderId());
            return;
        }

        Order existingOrder = orderOpt.get();
        
        // Update order status to IN_PREPARATION
        Order updatedOrder = new Order(
            existingOrder.id(),
            existingOrder.roomId(),
            existingOrder.reservationId(),
            existingOrder.items(),
            existingOrder.totalPrice(),
            OrderStatus.IN_PREPARATION.name(),
            existingOrder.createdAt()
        );
        
        orderRepository.createOrder(updatedOrder); // Overwrites existing
        Log.infof("Order %s is now IN_PREPARATION", event.orderId());
    }

    /**
     * Handles OrderBillingFailed: Restore stock and cancel order.
     */
    @WithSpan("order.handleOrderBillingFailed")
    public void handleOrderBillingFailed(OrderEvent event) {
        Log.infof("Processing OrderBillingFailed for order %s", event.orderId());

        Optional<Order> orderOpt = orderRepository.getOrder(event.orderId());
        if (orderOpt.isEmpty()) {
            Log.warnf("Order %s not found for billing failure handling", event.orderId());
            return;
        }

        Order order = orderOpt.get();

        // Restore stock for all items (compensation)
        for (Item item : order.items()) {
            inventoryRepository.restoreStock(item.menuItemId(), item.quantity());
        }

        // Delete the order from Redis (it was never confirmed)
        orderRepository.deleteOrder(order.id());
        Log.infof("Order %s cancelled due to billing failure, stock restored", event.orderId());
    }

    @WithSpan("order.getOrder")
    public Optional<Order> getOrder(String id) {
        Log.info("Getting order with id " + id);
        return orderRepository.getOrder(id);
    }

    @WithSpan("order.getOrders")
    public List<Order> getOrders() {
        Log.info("Getting orders");
        return orderRepository.getOrders();
    }

    /**
     * Updates order status and notifies via Kafka.
     * Used when kitchen completes or cancels an order.
     */
    @WithSpan("order.updateOrder")
    public void updateOrder(String id, OrderStatus status) {
        Log.info("Updating order with id " + id + " to " + status);
        Order order = orderRepository.getOrder(id).orElseThrow(NotFoundException::new);

        if (status == OrderStatus.COMPLETED) {
            // Send completion event
            OrderEvent completionEvent = OrderEvent.orderPreparationCompleted(order);
            publishOrderEvent(completionEvent);
            // Delete from Redis on completion
            orderRepository.deleteOrder(id);
        } else if (status == OrderStatus.CANCELLED) {
            // If cancelled after confirmation, need to notify for billing cancellation and restore stock
            if (OrderStatus.IN_PREPARATION.name().equals(order.status())) {
                // Restore stock for all items
                for (Item item : order.items()) {
                    inventoryRepository.restoreStock(item.menuItemId(), item.quantity());
                }
                OrderEvent failureEvent = OrderEvent.orderPreparationFailed(order, "Kitchen cancelled order");
                publishOrderEvent(failureEvent);
            }
            // Delete from Redis on cancellation
            orderRepository.deleteOrder(id);
        }
    }

    private Order createOrderFromEvent(OrderEvent event, OrderStatus status) {
        return new Order(
            event.orderId(),
            event.roomId(),
            event.reservationId(),
            event.items(),
            event.totalPrice(),
            status.name(),
            Instant.now()
        );
    }

    private void publishOrderEvent(OrderEvent event) {
        try {
            orderEventProducer.publish(event);
            Log.infof("Published %s event for order %s", event.eventType(), event.orderId());
        } catch (Exception e) {
            Log.errorf(e, "Failed to publish %s event for order %s", event.eventType(), event.orderId());
        }
    }
}
