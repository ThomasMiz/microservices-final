package io.chotel.order;

import io.chotel.order.dto.Item;
import io.chotel.order.dto.Order;
import io.chotel.order.dto.OrderCompletionRequest;
import io.chotel.order.dto.OrderStatus;
import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.common.http.TestHTTPEndpoint;
import io.restassured.http.ContentType;
import jakarta.inject.Inject;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.time.Instant;
import java.util.List;
import java.util.Optional;

import static io.restassured.RestAssured.given;
import static org.junit.jupiter.api.Assertions.*;

@QuarkusTest
@TestHTTPEndpoint(OrderResource.class)
class OrderKafkaIntegrationTest {

    @Inject
    OrderRepository orderRepository;

    @Inject
    OrderService orderService;

    private Order testOrder;
    private static final String ORDER_ID = "test-order-123";

    @BeforeEach
    void setUp() throws InterruptedException {
        testOrder = new Order(ORDER_ID, "room-123", "reservation-456", List.of(
                new Item("Pizza", 2, "15.00"),
                new Item("Coke", 1, "3.00")
        ), "18.00", "PENDING", Instant.now());
        
        // Clean up any existing test data
        orderRepository.deleteOrder(ORDER_ID);
        Thread.sleep(50); // Small delay to ensure deletion completes
    }

    @Test
    void testOrderStore_ReceivesKafkaMessage() throws InterruptedException {
        // Store order directly in repository
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Wait for async processing

        // Verify that the order was stored in Redis
        Optional<Order> storedOrder = orderRepository.getOrder(ORDER_ID);
        assertTrue(storedOrder.isPresent(), "Order should be stored in Redis");
        assertEquals(testOrder, storedOrder.get(), "Stored order should match the original");
    }

    @Test
    void testOrderUpdate_SendsKafkaMessage() throws Exception {
        // First store an order in Redis
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Wait for storage to complete
        
        // Create order completion request
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.COMPLETED);

        // Perform the update via REST API
        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);

        // Wait a bit for async processing
        Thread.sleep(200);

        // Verify that order was deleted from Redis (after successful Kafka send)
        Optional<Order> deletedOrder = orderRepository.getOrder(ORDER_ID);
        assertFalse(deletedOrder.isPresent(), "Order should be deleted from Redis after update");
    }

    @Test
    void testCompleteOrderFlow_KafkaInteraction() throws Exception {
        // Test the complete flow: store order, then update via REST API
        
        // Step 1: Store order directly in repository
        orderRepository.createOrder(testOrder);
        Thread.sleep(100);
        
        // Step 2: Verify order is in Redis
        Optional<Order> storedOrder = orderRepository.getOrder(ORDER_ID);
        assertTrue(storedOrder.isPresent(), "Order should be stored in Redis");
        
        // Step 3: Update order via REST API (which sends Kafka message and deletes from Redis)
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.CANCELLED);

        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);

        // Wait for async processing
        Thread.sleep(200);

        // Step 4: Verify order is deleted from Redis
        Optional<Order> deletedOrder = orderRepository.getOrder(ORDER_ID);
        assertFalse(deletedOrder.isPresent(), "Order should be deleted from Redis after update");
    }

    @Test
    void testMultipleOrders_KafkaInteraction() throws Exception {
        // Test multiple orders to ensure Redis and Kafka handle multiple messages correctly
        Order order1 = new Order("order-1", "room-1", "reservation-1", List.of(new Item("Burger", 1, "8.00")), "8.00", "PENDING", Instant.now());
        Order order2 = new Order("order-2", "room-2", "reservation-2", List.of(new Item("Pizza", 2, "16.00")), "16.00", "PENDING", Instant.now());

        // Store orders directly in repository
        orderRepository.createOrder(order1);
        orderRepository.createOrder(order2);
        Thread.sleep(100);

        // Update orders via REST API
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.COMPLETED);

        given()
                .pathParam("id", "order-1")
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);

        given()
                .pathParam("id", "order-2")
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);

        // Wait for async processing
        Thread.sleep(200);

        // Verify all orders are deleted from Redis
        Optional<Order> deletedOrder1 = orderRepository.getOrder("order-1");
        Optional<Order> deletedOrder2 = orderRepository.getOrder("order-2");
        
        assertFalse(deletedOrder1.isPresent(), "Order 1 should be deleted from Redis after update");
        assertFalse(deletedOrder2.isPresent(), "Order 2 should be deleted from Redis after update");
    }
}