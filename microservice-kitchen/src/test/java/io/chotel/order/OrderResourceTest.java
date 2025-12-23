package io.chotel.order;

import io.chotel.order.dto.Item;
import io.chotel.order.dto.Order;
import io.chotel.order.dto.OrderCompletionRequest;
import io.chotel.order.dto.OrderStatus;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.time.Instant;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;
import static org.hamcrest.Matchers.hasSize;
import static org.hamcrest.Matchers.notNullValue;
import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@QuarkusTest
class OrderResourceTest {

    @InjectMock
    OrderService orderService;

    private Order testOrder;
    private static final String ORDER_ID = "test-order-123";

    @BeforeEach
    void setUp() {
        testOrder = new Order(ORDER_ID, "room-123", "reservation-456", Arrays.asList(
            new Item("burger", 2, "10.00"),
            new Item("fries", 1, "5.00")
        ), "15.00", "PENDING", Instant.now());
    }

    @Test
    void testGetOrders() {
        List<Order> expectedOrders = Arrays.asList(testOrder, new Order("order-456", "room-456", "reservation-456", Arrays.asList(), "0.00", "PENDING", Instant.now()));
        when(orderService.getOrders()).thenReturn(expectedOrders);

        given()
            .when().get("/orders")
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("", hasSize(2))
            .body("[0].id", is(ORDER_ID))
            .body("[0].items", hasSize(2))
            .body("[0].items[0].menuItemId", is("burger"))
            .body("[0].items[0].quantity", is(2))
            .body("[0].items[1].menuItemId", is("fries"))
            .body("[0].items[1].quantity", is(1));

        verify(orderService).getOrders();
    }

    @Test
    void testGetOrder_WhenOrderExists() {
        when(orderService.getOrder(ORDER_ID)).thenReturn(Optional.of(testOrder));

        given()
            .pathParam("id", ORDER_ID)
            .when().get("/orders/{id}")
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("id", is(ORDER_ID))
            .body("items", hasSize(2))
            .body("items[0].menuItemId", is("burger"))
            .body("items[0].quantity", is(2))
            .body("items[1].menuItemId", is("fries"))
            .body("items[1].quantity", is(1));

        verify(orderService).getOrder(ORDER_ID);
    }

    @Test
    void testGetOrder_WhenOrderDoesNotExist() {
        when(orderService.getOrder(ORDER_ID)).thenReturn(Optional.empty());

        given()
            .pathParam("id", ORDER_ID)
            .when().get("/orders/{id}")
            .then()
            .statusCode(404);

        verify(orderService).getOrder(ORDER_ID);
    }

    @Test
    void testUpdateOrder_WithCompletedStatus() {
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.COMPLETED);

        given()
            .pathParam("id", ORDER_ID)
            .contentType(ContentType.JSON)
            .body(request)
            .when().patch("/orders/{id}")
            .then()
            .statusCode(204);

        verify(orderService).updateOrder(ORDER_ID, OrderStatus.COMPLETED);
    }

    @Test
    void testUpdateOrder_WithCancelledStatus() {
        OrderCompletionRequest request = new OrderCompletionRequest();
        request.setStatus(OrderStatus.CANCELLED);

        given()
            .pathParam("id", ORDER_ID)
            .contentType(ContentType.JSON)
            .body(request)
            .when().patch("/orders/{id}")
            .then()
            .statusCode(204);

        verify(orderService).updateOrder(ORDER_ID, OrderStatus.CANCELLED);
    }

    @Test
    void testUpdateOrder_WithInvalidRequest() {
        // Test with null status (should fail validation)
        OrderCompletionRequest request = new OrderCompletionRequest();
        // status remains null

        given()
            .pathParam("id", ORDER_ID)
            .contentType(ContentType.JSON)
            .body(request)
            .when().patch("/orders/{id}")
            .then()
            .statusCode(400);

        verify(orderService, never()).updateOrder(anyString(), any());
    }

    @Test
    void testUpdateOrder_WithInvalidJson() {
        // Test with malformed JSON
        given()
            .pathParam("id", ORDER_ID)
            .contentType(ContentType.JSON)
            .body("{\"invalid\":\"json\"}")
            .when().patch("/orders/{id}")
            .then()
            .statusCode(400);

        verify(orderService, never()).updateOrder(anyString(), any());
    }
}