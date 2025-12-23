package io.chotel.order;

import io.chotel.order.dto.Item;
import io.chotel.order.dto.Order;
import io.chotel.order.dto.OrderEvent;
import io.chotel.order.dto.OrderStatus;
import io.chotel.order.messaging.RedisOrderEventProducer;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import jakarta.ws.rs.NotFoundException;
import java.time.Instant;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

@QuarkusTest
class OrderServiceTest {

    @InjectMock
    OrderRepository orderRepository;

    @InjectMock
    InventoryRepository inventoryRepository;

    @InjectMock
    RedisOrderEventProducer orderEventProducer;

    OrderService orderService;

    private Order testOrder;
    private static final String ORDER_ID = "test-order-123";

    @BeforeEach
    void setUp() {
        orderService = new OrderService();
        orderService.orderRepository = orderRepository;
        orderService.inventoryRepository = inventoryRepository;
        orderService.orderEventProducer = orderEventProducer;

        testOrder = new Order(ORDER_ID, "room-123", "reservation-456", Arrays.asList(
            new Item("burger", 2, "10.00"),
            new Item("fries", 1, "5.00")
        ), "15.00", "PENDING", Instant.now());
    }

    @Test
    void testGetOrder_WhenOrderExists() {
        Mockito.when(orderRepository.getOrder(ORDER_ID)).thenReturn(Optional.of(testOrder));

        Optional<Order> result = orderService.getOrder(ORDER_ID);

        assertTrue(result.isPresent());
        assertEquals(testOrder, result.get());
        Mockito.verify(orderRepository).getOrder(ORDER_ID);
    }

    @Test
    void testGetOrder_WhenOrderDoesNotExist() {
        Mockito.when(orderRepository.getOrder(ORDER_ID)).thenReturn(Optional.empty());

        Optional<Order> result = orderService.getOrder(ORDER_ID);

        assertFalse(result.isPresent());
        Mockito.verify(orderRepository).getOrder(ORDER_ID);
    }

    @Test
    void testGetOrders() {
        List<Order> expectedOrders = Arrays.asList(testOrder, new Order("order-456", "room-456", "reservation-456", Arrays.asList(), "0.00", "PENDING", Instant.now()));
        Mockito.when(orderRepository.getOrders()).thenReturn(expectedOrders);

        List<Order> result = orderService.getOrders();

        assertEquals(expectedOrders, result);
        Mockito.verify(orderRepository).getOrders();
    }

    @Test
    void testUpdateOrder_Success() {
        Mockito.when(orderRepository.getOrder(ORDER_ID)).thenReturn(Optional.of(testOrder));
        Mockito.doNothing().when(orderEventProducer).publish(Mockito.any(OrderEvent.class));

        orderService.updateOrder(ORDER_ID, OrderStatus.COMPLETED);

        Mockito.verify(orderRepository).getOrder(ORDER_ID);
        Mockito.verify(orderEventProducer).publish(Mockito.any(OrderEvent.class));
        Mockito.verify(orderRepository).deleteOrder(ORDER_ID);
    }

    @Test
    void testUpdateOrder_OrderNotFound() {
        Mockito.when(orderRepository.getOrder(ORDER_ID)).thenReturn(Optional.empty());

        assertThrows(NotFoundException.class, () -> 
            orderService.updateOrder(ORDER_ID, OrderStatus.COMPLETED)
        );

        Mockito.verify(orderRepository).getOrder(ORDER_ID);
        Mockito.verify(orderRepository, Mockito.never()).deleteOrder(Mockito.anyString());
    }

    @Test
    void testUpdateOrder_CancelledStatus() {
        Mockito.when(orderRepository.getOrder(ORDER_ID)).thenReturn(Optional.of(testOrder));

        orderService.updateOrder(ORDER_ID, OrderStatus.CANCELLED);

        Mockito.verify(orderRepository).getOrder(ORDER_ID);
        // Order event should not be sent for PENDING -> CANCELLED transition
        Mockito.verify(orderEventProducer, Mockito.never()).publish(Mockito.any(OrderEvent.class));
        Mockito.verify(orderRepository).deleteOrder(ORDER_ID);
    }
}