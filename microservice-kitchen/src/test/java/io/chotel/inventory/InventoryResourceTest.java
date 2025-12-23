package io.chotel.inventory;

import io.chotel.inventory.dto.InventoryResponse;
import io.quarkus.test.InjectMock;
import io.quarkus.test.common.http.TestHTTPEndpoint;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import jakarta.ws.rs.NotFoundException;
import java.time.Instant;
import java.util.Arrays;
import java.util.List;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.*;
import static org.mockito.ArgumentMatchers.any;

@QuarkusTest
@TestHTTPEndpoint(InventoryResource.class)
class InventoryResourceTest {

    @InjectMock
    InventoryService inventoryService;

    private static final String BURGER_ID = "burger";
    private static final String FRIES_ID = "fries";
    private InventoryResponse burgerResponse;
    private InventoryResponse friesResponse;

    @BeforeEach
    void setUp() {
        burgerResponse = new InventoryResponse(BURGER_ID, 50L, Instant.now());
        friesResponse = new InventoryResponse(FRIES_ID, 100L, Instant.now());
    }

    @Test
    void testUpdateStock_Success() {
        Mockito.when(inventoryService.updateStock(any())).thenReturn(burgerResponse);

        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "menuItemId": "burger",
                    "quantity": 50
                }
                """)
            .when().post()
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("menuItemId", is(BURGER_ID))
            .body("currentStock", is(50));
    }

    @Test
    void testUpdateStock_ValidationFailure_MissingMenuItemId() {
        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "quantity": 50
                }
                """)
            .when().post()
            .then()
            .statusCode(400);
    }

    @Test
    void testUpdateStock_ValidationFailure_NegativeQuantity() {
        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "menuItemId": "burger",
                    "quantity": -10
                }
                """)
            .when().post()
            .then()
            .statusCode(400);
    }

    @Test
    void testUpdateStock_ValidationFailure_ZeroQuantity() {
        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "menuItemId": "burger",
                    "quantity": 0
                }
                """)
            .when().post()
            .then()
            .statusCode(400);
    }

    @Test
    void testGetStock_Success() {
        Mockito.when(inventoryService.getStock(BURGER_ID)).thenReturn(burgerResponse);

        given()
            .pathParam("menuItemId", BURGER_ID)
            .when().get("/{menuItemId}")
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("menuItemId", is(BURGER_ID))
            .body("currentStock", is(50));
    }

    @Test
    void testGetStock_NotFound() {
        Mockito.when(inventoryService.getStock(BURGER_ID))
            .thenThrow(new NotFoundException("Menu item not found in inventory: " + BURGER_ID));

        given()
            .pathParam("menuItemId", BURGER_ID)
            .when().get("/{menuItemId}")
            .then()
            .statusCode(404);
    }

    @Test
    void testListAllInventory_Success() {
        List<InventoryResponse> inventory = Arrays.asList(burgerResponse, friesResponse);
        Mockito.when(inventoryService.listAllInventory()).thenReturn(inventory);

        given()
            .when().get()
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("$", hasSize(2))
            .body("[0].menuItemId", is(BURGER_ID))
            .body("[0].currentStock", is(50))
            .body("[1].menuItemId", is(FRIES_ID))
            .body("[1].currentStock", is(100));
    }

    @Test
    void testListAllInventory_EmptyList() {
        Mockito.when(inventoryService.listAllInventory()).thenReturn(List.of());

        given()
            .when().get()
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("$", hasSize(0));
    }

    @Test
    void testBulkUpdateStock_Success() {
        List<InventoryResponse> responses = Arrays.asList(burgerResponse, friesResponse);
        Mockito.when(inventoryService.bulkUpdateStock(any())).thenReturn(responses);

        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "items": [
                        {"menuItemId": "burger", "quantity": 50},
                        {"menuItemId": "fries", "quantity": 100}
                    ]
                }
                """)
            .when().post("/bulk")
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("$", hasSize(2))
            .body("[0].menuItemId", is(BURGER_ID))
            .body("[1].menuItemId", is(FRIES_ID));
    }

    @Test
    void testBulkUpdateStock_ValidationFailure_EmptyList() {
        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "items": []
                }
                """)
            .when().post("/bulk")
            .then()
            .statusCode(400);
    }

    @Test
    void testBulkUpdateStock_ValidationFailure_InvalidItem() {
        given()
            .contentType(ContentType.JSON)
            .body("""
                {
                    "items": [
                        {"menuItemId": "burger", "quantity": 50},
                        {"menuItemId": "fries", "quantity": -10}
                    ]
                }
                """)
            .when().post("/bulk")
            .then()
            .statusCode(400);
    }

    @Test
    void testAddStock_Success() {
        InventoryResponse updatedResponse = new InventoryResponse(BURGER_ID, 75L, Instant.now());
        Mockito.when(inventoryService.addStock(any())).thenReturn(updatedResponse);

        given()
            .pathParam("menuItemId", BURGER_ID)
            .contentType(ContentType.JSON)
            .body("""
                {
                    "menuItemId": "burger",
                    "quantity": 25
                }
                """)
            .when().post("/{menuItemId}/add")
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("menuItemId", is(BURGER_ID))
            .body("currentStock", is(75));
    }

    @Test
    void testAddStock_PathParameterMismatch() {
        given()
            .pathParam("menuItemId", BURGER_ID)
            .contentType(ContentType.JSON)
            .body("""
                {
                    "menuItemId": "fries",
                    "quantity": 25
                }
                """)
            .when().post("/{menuItemId}/add")
            .then()
            .statusCode(400)
            .body(containsString("Menu item ID in path must match request body"));
    }

    @Test
    void testAddStock_ValidationFailure_NegativeQuantity() {
        given()
            .pathParam("menuItemId", BURGER_ID)
            .contentType(ContentType.JSON)
            .body("""
                {
                    "menuItemId": "burger",
                    "quantity": -25
                }
                """)
            .when().post("/{menuItemId}/add")
            .then()
            .statusCode(400);
    }
}
