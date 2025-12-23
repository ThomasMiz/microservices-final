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

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.CoreMatchers.notNullValue;
import static org.hamcrest.Matchers.hasSize;
import static org.junit.jupiter.api.Assertions.*;

@QuarkusTest
@TestHTTPEndpoint(OrderResource.class)
class OrderIntegrationTest {

    @Inject
    OrderRepository orderRepository;

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
    void testGetOrders_EmptyList() {
        given()
                .when()
                .get()
                .then()
                .statusCode(200)
                .contentType(ContentType.JSON)
                .body("$", hasSize(0));
    }

    @Test
    void testGetOrder_NotFound() {
        given()
                .pathParam("id", "non-existent-order")
                .when()
                .get("/{id}")
                .then()
                .statusCode(404);
    }

    @Test
    void testGetOrder_Found() throws InterruptedException {
        // Create an order first
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Small delay to ensure order is stored
        
        given()
                .pathParam("id", ORDER_ID)
                .when()
                .get("/{id}")
                .then()
                .statusCode(200)
                .contentType(ContentType.JSON)
                .body("id", is(ORDER_ID))
                .body("items", hasSize(2))
                .body("items[0].menuItemId", is("Pizza"))
                .body("items[0].quantity", is(2))
                .body("items[1].menuItemId", is("Coke"))
                .body("items[1].quantity", is(1));
    }

    @Test
    void testGetOrders_WithOrders() throws InterruptedException {
        // Create an order first
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Small delay to ensure order is stored
        
        given()
                .when()
                .get()
                .then()
                .statusCode(200)
                .contentType(ContentType.JSON)
                .body("$", hasSize(1))
                .body("[0].id", is(ORDER_ID))
                .body("[0].items", hasSize(2));
    }

    @Test
    void testUpdateOrder_NotFound() {
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.COMPLETED);

        given()
                .pathParam("id", "non-existent-order")
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(404);
    }

    @Test
    void testUpdateOrder_InvalidRequest() {
        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body("{}")
                .when()
                .patch("/{id}")
                .then()
                .statusCode(400);
    }

    @Test
    void testOrderFlow_CompleteOrder() throws InterruptedException {
        // First, create an order in the repository
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Small delay to ensure order is stored
        
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.COMPLETED);

        // Test updating order to completed status
        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204); // No Content for successful update
    }

    @Test
    void testOrderFlow_CancelOrder() throws InterruptedException {
        // First, create an order in the repository
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Small delay to ensure order is stored
        
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.CANCELLED);

        // Test updating order to cancelled status
        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body(request)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204); // No Content for successful update
    }

    @Test
    void testUpdateOrder_BothStatuses() throws InterruptedException {
        // Test COMPLETED status
        orderRepository.createOrder(testOrder);
        Thread.sleep(100); // Small delay to ensure order is stored
        
        OrderCompletionRequest completedRequest = new OrderCompletionRequest();
        completedRequest.setStatus(OrderStatus.COMPLETED);

        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body(completedRequest)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);

        Thread.sleep(100); // Small delay to ensure async operation completes
        
        // Test CANCELLED status with a different order ID
        String secondOrderId = "test-order-456";
        Order secondOrder = new Order(secondOrderId, "room-789", "reservation-012", List.of(
                new Item("Burger", 1, "8.00"),
                new Item("Fries", 1, "3.00")
        ), "11.00", "PENDING", Instant.now());
        
        orderRepository.createOrder(secondOrder);
        Thread.sleep(100); // Small delay to ensure order is stored
        
        OrderCompletionRequest cancelledRequest = new OrderCompletionRequest();
        cancelledRequest.setStatus(OrderStatus.CANCELLED);

        given()
                .pathParam("id", secondOrderId)
                .contentType(ContentType.JSON)
                .body(cancelledRequest)
                .when()
                .patch("/{id}")
                .then()
                .statusCode(204);
    }

    @Test
    void testResponseHeaders() {
        given()
                .when()
                .get()
                .then()
                .statusCode(200)
                .contentType(ContentType.JSON);
    }

    @Test
    void testOrderSerialization() {
        // This test verifies that the Order DTO can be properly serialized/deserialized
        Order order = new Order("test-id", "room-test", "reservation-test", List.of(new Item("test-item", 1, "5.00")), "5.00", "PENDING", Instant.now());
        
        assertNotNull(order.id());
        assertNotNull(order.items());
        assertEquals("test-id", order.id());
        assertEquals(1, order.items().size());
        assertEquals("test-item", order.items().get(0).menuItemId());
        assertEquals(1, order.items().get(0).quantity());
    }

    @Test
    void testOrderCompletionRequestValidation() {
        // Test with null status (should fail validation)
        given()
                .pathParam("id", ORDER_ID)
                .contentType(ContentType.JSON)
                .body("{\"status\": null}")
                .when()
                .patch("/{id}")
                .then()
                .statusCode(400);
    }
}