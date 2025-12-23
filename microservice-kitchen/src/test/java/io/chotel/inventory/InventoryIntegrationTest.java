package io.chotel.inventory;

import io.chotel.order.InventoryRepository;
import io.quarkus.test.common.http.TestHTTPEndpoint;
import io.quarkus.test.junit.QuarkusTest;
import io.restassured.http.ContentType;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import jakarta.inject.Inject;

import static io.restassured.RestAssured.given;
import static org.hamcrest.Matchers.*;

/**
 * Full integration tests for Inventory endpoints with real Redis backend.
 * Tests the complete HTTP → Service → Repository → Redis flow.
 */
@QuarkusTest
@TestHTTPEndpoint(InventoryResource.class)
class InventoryIntegrationTest {

    @Inject
    InventoryRepository inventoryRepository;

    private static final String BURGER_ID = "integration-burger";
    private static final String FRIES_ID = "integration-fries";
    private static final String PIZZA_ID = "integration-pizza";

    @BeforeEach
    void setUp() throws InterruptedException {
        // Clean up test data
        inventoryRepository.setStock(BURGER_ID, 0L);
        inventoryRepository.setStock(FRIES_ID, 0L);
        inventoryRepository.setStock(PIZZA_ID, 0L);
        Thread.sleep(50); // Small delay to ensure cleanup completes
    }

    @Test
    void testFullCycle_CreateUpdateAndRetrieveInventory() throws InterruptedException {
        // 1. Create initial stock
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 50
                }
                """, BURGER_ID))
            .when().post()
            .then()
            .statusCode(200)
            .body("menuItemId", is(BURGER_ID))
            .body("currentStock", is(50));

        Thread.sleep(100); // Wait for async processing

        // 2. Verify stock was persisted in Redis
        long stockInRedis = inventoryRepository.getStock(BURGER_ID);
        assert stockInRedis == 50L : "Expected stock to be 50, but got " + stockInRedis;

        // 3. Get stock via API
        given()
            .pathParam("menuItemId", BURGER_ID)
            .when().get("/{menuItemId}")
            .then()
            .statusCode(200)
            .body("menuItemId", is(BURGER_ID))
            .body("currentStock", is(50));

        // 4. Update stock
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 75
                }
                """, BURGER_ID))
            .when().post()
            .then()
            .statusCode(200)
            .body("currentStock", is(75));

        Thread.sleep(100);

        // 5. Verify updated stock
        long updatedStock = inventoryRepository.getStock(BURGER_ID);
        assert updatedStock == 75L : "Expected updated stock to be 75, but got " + updatedStock;
    }

    @Test
    void testAddStock_AccumulatesCorrectly() throws InterruptedException {
        // 1. Set initial stock
        inventoryRepository.setStock(BURGER_ID, 30L);
        Thread.sleep(50);

        // 2. Add more stock via API
        given()
            .pathParam("menuItemId", BURGER_ID)
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 20
                }
                """, BURGER_ID))
            .when().post("/{menuItemId}/add")
            .then()
            .statusCode(200)
            .body("menuItemId", is(BURGER_ID))
            .body("currentStock", is(50)); // 30 + 20

        Thread.sleep(100);

        // 3. Verify accumulated stock in Redis
        long finalStock = inventoryRepository.getStock(BURGER_ID);
        assert finalStock == 50L : "Expected accumulated stock to be 50, but got " + finalStock;
    }

    @Test
    void testListAllInventory_ReturnsAllItems() throws InterruptedException {
        // 1. Create multiple inventory items
        inventoryRepository.setStock(BURGER_ID, 50L);
        inventoryRepository.setStock(FRIES_ID, 100L);
        inventoryRepository.setStock(PIZZA_ID, 25L);
        Thread.sleep(100);

        // 2. List all items
        given()
            .when().get()
            .then()
            .statusCode(200)
            .contentType(ContentType.JSON)
            .body("$", hasSize(greaterThanOrEqualTo(3)))
            .body("menuItemId", hasItems(BURGER_ID, FRIES_ID, PIZZA_ID));
    }

    @Test
    void testBulkUpdate_UpdatesMultipleItems() throws InterruptedException {
        // 1. Bulk update
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "items": [
                        {"menuItemId": "%s", "quantity": 60},
                        {"menuItemId": "%s", "quantity": 120}
                    ]
                }
                """, BURGER_ID, FRIES_ID))
            .when().post("/bulk")
            .then()
            .statusCode(200)
            .body("$", hasSize(2))
            .body("[0].menuItemId", is(BURGER_ID))
            .body("[0].currentStock", is(60))
            .body("[1].menuItemId", is(FRIES_ID))
            .body("[1].currentStock", is(120));

        Thread.sleep(100);

        // 2. Verify both items were updated in Redis
        long burgerStock = inventoryRepository.getStock(BURGER_ID);
        long friesStock = inventoryRepository.getStock(FRIES_ID);
        
        assert burgerStock == 60L : "Expected burger stock to be 60, but got " + burgerStock;
        assert friesStock == 120L : "Expected fries stock to be 120, but got " + friesStock;
    }

    @Test
    void testGetStock_NotFound_Returns404() {
        given()
            .pathParam("menuItemId", "non-existent-item")
            .when().get("/{menuItemId}")
            .then()
            .statusCode(404);
    }

    @Test
    void testUpdateStock_WithZeroQuantity_Returns400() {
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 0
                }
                """, BURGER_ID))
            .when().post()
            .then()
            .statusCode(400);
    }

    @Test
    void testUpdateStock_WithNegativeQuantity_Returns400() {
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": -50
                }
                """, BURGER_ID))
            .when().post()
            .then()
            .statusCode(400);
    }

    @Test
    void testBulkUpdate_WithEmptyList_Returns400() {
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
    void testAddStock_PathMismatch_Returns400() {
        given()
            .pathParam("menuItemId", BURGER_ID)
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 25
                }
                """, FRIES_ID))
            .when().post("/{menuItemId}/add")
            .then()
            .statusCode(400)
            .body(containsString("Menu item ID in path must match request body"));
    }

    @Test
    void testConcurrentUpdates_MaintainsConsistency() throws InterruptedException {
        // Set initial stock
        inventoryRepository.setStock(BURGER_ID, 100L);
        Thread.sleep(50);

        // Simulate concurrent updates (in practice, test sequentially)
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 150
                }
                """, BURGER_ID))
            .when().post()
            .then()
            .statusCode(200);

        Thread.sleep(50);

        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 200
                }
                """, BURGER_ID))
            .when().post()
            .then()
            .statusCode(200);

        Thread.sleep(100);

        // Verify final state (last write wins)
        long finalStock = inventoryRepository.getStock(BURGER_ID);
        assert finalStock == 200L : "Expected final stock to be 200, but got " + finalStock;
    }

    @Test
    void testStockPersistence_AcrossRequests() throws InterruptedException {
        // 1. Set stock
        given()
            .contentType(ContentType.JSON)
            .body(String.format("""
                {
                    "menuItemId": "%s",
                    "quantity": 88
                }
                """, PIZZA_ID))
            .when().post()
            .then()
            .statusCode(200);

        Thread.sleep(100);

        // 2. Retrieve in separate request
        given()
            .pathParam("menuItemId", PIZZA_ID)
            .when().get("/{menuItemId}")
            .then()
            .statusCode(200)
            .body("currentStock", is(88));

        // 3. Retrieve via list endpoint
        given()
            .when().get()
            .then()
            .statusCode(200)
            .body("find { it.menuItemId == '" + PIZZA_ID + "' }.currentStock", is(88));
    }
}
