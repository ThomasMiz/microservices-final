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
import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.Matchers.hasSize;
import static org.junit.jupiter.api.Assertions.*;

@QuarkusTest
@TestHTTPEndpoint(OrderResource.class)
class OrderRedisIntegrationTest {

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
        orderRepository.deleteOrder("order-1");
        orderRepository.deleteOrder("order-2");
        Thread.sleep(50); // Small delay to ensure deletion completes
    }

    @Test
    void testOrderStorage_RedisPersistence() throws InterruptedException {
        // Test storing order in Redis
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Wait for async processing
        
        // Verify order was stored in Redis
        Optional<Order> storedOrder = orderRepository.getOrder(ORDER_ID);
        assertTrue(storedOrder.isPresent(), "Order should be stored in Redis");
        assertEquals(testOrder, storedOrder.get(), "Stored order should match the original");
    }

    @Test
    void testOrderRetrieval_RedisFetch() throws InterruptedException {
        // First store an order in Redis
        orderRepository.createOrder(testOrder);
        Thread.sleep(50); // Small delay to ensure storage completes
        
        // Retrieve order via service
        Optional<Order> retrievedOrder = orderService.getOrder(ORDER_ID);
        
        // Verify retrieval
        assertTrue(retrievedOrder.isPresent(), "Order should be retrievable from Redis");
        assertEquals(testOrder, retrievedOrder.get(), "Retrieved order should match the stored order");
    }

    @Test
    void testOrderList_RedisFetchAll() throws InterruptedException {
        // Store multiple orders in Redis
        Order order1 = new Order("order-1", "room-1", "reservation-1", List.of(new Item("Burger", 1, "8.00")), "8.00", "PENDING", Instant.now());
        Order order2 = new Order("order-2", "room-2", "reservation-2", List.of(new Item("Pizza", 2, "16.00")), "16.00", "PENDING", Instant.now());
        
        orderRepository.createOrder(testOrder);
        orderRepository.createOrder(order1);
        orderRepository.createOrder(order2);
        Thread.sleep(50); // Small delay to ensure storage completes
        
        // Retrieve all orders via service
        List<Order> allOrders = orderService.getOrders();
        
        // Verify retrieval
        assertEquals(3, allOrders.size(), "Should retrieve all 3 orders from Redis");
        assertTrue(allOrders.contains(testOrder), "Orders list should contain testOrder");
        assertTrue(allOrders.contains(order1), "Orders list should contain order1");
        assertTrue(allOrders.contains(order2), "Orders list should contain order2");
    }

    @Test
    void testOrderUpdate_RedisDeletion() throws InterruptedException {
        // Store order in Redis first
        orderRepository.createOrder(testOrder);
        Thread.sleep(50);
        
        // Verify order exists before update
        Optional<Order> beforeUpdate = orderService.getOrder(ORDER_ID);
        assertTrue(beforeUpdate.isPresent(), "Order should exist before update");
        
        // Update order via service (which should send Kafka message and delete from Redis)
        orderService.updateOrder(ORDER_ID, OrderStatus.COMPLETED);
        
        // Wait for async processing (Kafka send + deletion)
        Thread.sleep(300);
        
        // Verify order was deleted from Redis
        Optional<Order> afterUpdate = orderService.getOrder(ORDER_ID);
        assertFalse(afterUpdate.isPresent(), "Order should be deleted from Redis after update");
    }

    @Test
    void testOrderUpdate_RedisNotFound() {
        // Try to update an order that doesn't exist in Redis
        assertThrows(jakarta.ws.rs.NotFoundException.class, () -> 
            orderService.updateOrder("non-existent-order", OrderStatus.COMPLETED)
        );
    }

    @Test
    void testCompleteOrderFlow_RedisInteraction() throws InterruptedException {
        // Test complete flow: store -> retrieve -> update -> verify deletion
        
        // Step 1: Store order directly in repository
        orderRepository.createOrder(testOrder);
        Thread.sleep(100);
        
        // Step 2: Verify order is in Redis
        Optional<Order> storedOrder = orderService.getOrder(ORDER_ID);
        assertTrue(storedOrder.isPresent(), "Order should be stored in Redis");
        
        // Step 3: Update order (sends event, deletes from Redis)
        orderService.updateOrder(ORDER_ID, OrderStatus.CANCELLED);
        Thread.sleep(100);
        
        // Step 4: Verify order is deleted from Redis
        Optional<Order> deletedOrder = orderService.getOrder(ORDER_ID);
        assertFalse(deletedOrder.isPresent(), "Order should be deleted from Redis after update");
    }

    @Test
    void testRedisOperations_ViaRestAPI() throws InterruptedException {
        // Test Redis operations through the REST API endpoints
        
        // Store order in Redis
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Increased delay to ensure storage completes
        
        // Test GET /orders/{id}
        given()
                .pathParam("id", ORDER_ID)
                .when()
                .get("/{id}")
                .then()
                .statusCode(200)
                .contentType(ContentType.JSON)
                .body("id", is(ORDER_ID))
                .body("items", hasSize(2));
        
        // Test GET /orders
        given()
                .when()
                .get()
                .then()
                .statusCode(200)
                .contentType(ContentType.JSON)
                .body("$", hasSize(1))
                .body("[0].id", is(ORDER_ID));
        
        // Test PATCH /orders/{id} (which triggers Kafka and deletion)
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.COMPLETED);
        
        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);
        
        Thread.sleep(300); // Increased delay to ensure async processing completes
        
        // Verify order was deleted from Redis
        given()
                .pathParam("id", ORDER_ID)
                .when()
                .get("/{id}")
                .then()
                .statusCode(404);
    }
}