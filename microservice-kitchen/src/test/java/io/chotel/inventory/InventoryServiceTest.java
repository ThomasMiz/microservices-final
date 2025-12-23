package io.chotel.inventory;

import io.chotel.inventory.dto.InventoryBulkUpdateRequest;
import io.chotel.inventory.dto.InventoryResponse;
import io.chotel.inventory.dto.InventoryUpdateRequest;
import io.chotel.order.InventoryRepository;
import io.quarkus.redis.datasource.RedisDataSource;
import io.quarkus.redis.datasource.keys.KeyCommands;
import io.quarkus.test.InjectMock;
import io.quarkus.test.junit.QuarkusTest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import jakarta.ws.rs.NotFoundException;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;

@QuarkusTest
class InventoryServiceTest {

    @InjectMock
    InventoryRepository inventoryRepository;

    @InjectMock
    RedisDataSource redisDataSource;

    InventoryService inventoryService;

    private static final String BURGER_ID = "burger";
    private static final String FRIES_ID = "fries";
    private static final long INITIAL_STOCK = 50L;

    @BeforeEach
    void setUp() {
        inventoryService = new InventoryService();
        inventoryService.inventoryRepository = inventoryRepository;
        inventoryService.redisDataSource = redisDataSource;
    }

    @Test
    void testUpdateStock_SetsStockToSpecifiedQuantity() {
        InventoryUpdateRequest request = new InventoryUpdateRequest(BURGER_ID, INITIAL_STOCK);

        InventoryResponse response = inventoryService.updateStock(request);

        assertNotNull(response);
        assertEquals(BURGER_ID, response.menuItemId());
        assertEquals(INITIAL_STOCK, response.currentStock());
        assertNotNull(response.lastUpdated());
        
        Mockito.verify(inventoryRepository).setStock(BURGER_ID, INITIAL_STOCK);
    }

    @Test
    void testAddStock_AddsToExistingStock() {
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(30L);
        InventoryUpdateRequest request = new InventoryUpdateRequest(BURGER_ID, 20L);

        InventoryResponse response = inventoryService.addStock(request);

        assertNotNull(response);
        assertEquals(BURGER_ID, response.menuItemId());
        assertEquals(50L, response.currentStock()); // 30 + 20
        
        Mockito.verify(inventoryRepository).getStock(BURGER_ID);
        Mockito.verify(inventoryRepository).setStock(BURGER_ID, 50L);
    }

    @Test
    void testAddStock_CreatesNewItemWhenNonExistent() {
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(0L);
        InventoryUpdateRequest request = new InventoryUpdateRequest(BURGER_ID, INITIAL_STOCK);

        InventoryResponse response = inventoryService.addStock(request);

        assertNotNull(response);
        assertEquals(BURGER_ID, response.menuItemId());
        assertEquals(INITIAL_STOCK, response.currentStock());
        
        Mockito.verify(inventoryRepository).getStock(BURGER_ID);
        Mockito.verify(inventoryRepository).setStock(BURGER_ID, INITIAL_STOCK);
    }

    @Test
    void testGetStock_ReturnsStockForExistingItem() {
        KeyCommands<String> keyCommands = Mockito.mock(KeyCommands.class);
        Mockito.when(redisDataSource.key()).thenReturn(keyCommands);
        Mockito.when(keyCommands.exists("stock:" + BURGER_ID)).thenReturn(true);
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(INITIAL_STOCK);

        InventoryResponse response = inventoryService.getStock(BURGER_ID);

        assertNotNull(response);
        assertEquals(BURGER_ID, response.menuItemId());
        assertEquals(INITIAL_STOCK, response.currentStock());
        
        Mockito.verify(inventoryRepository).getStock(BURGER_ID);
    }

    @Test
    void testGetStock_ThrowsNotFoundForNonExistentItem() {
        KeyCommands<String> keyCommands = Mockito.mock(KeyCommands.class);
        Mockito.when(redisDataSource.key()).thenReturn(keyCommands);
        Mockito.when(keyCommands.exists("stock:" + BURGER_ID)).thenReturn(false);
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(0L);

        assertThrows(NotFoundException.class, () -> inventoryService.getStock(BURGER_ID));
        
        Mockito.verify(inventoryRepository).getStock(BURGER_ID);
    }

    @Test
    void testGetStock_ReturnsZeroForItemWithZeroStock() {
        KeyCommands<String> keyCommands = Mockito.mock(KeyCommands.class);
        Mockito.when(redisDataSource.key()).thenReturn(keyCommands);
        Mockito.when(keyCommands.exists("stock:" + BURGER_ID)).thenReturn(true); // Key exists
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(0L);

        InventoryResponse response = inventoryService.getStock(BURGER_ID);

        assertNotNull(response);
        assertEquals(BURGER_ID, response.menuItemId());
        assertEquals(0L, response.currentStock());
    }

    @Test
    void testListAllInventory_ReturnsAllItems() {
        KeyCommands<String> keyCommands = Mockito.mock(KeyCommands.class);
        Mockito.when(redisDataSource.key()).thenReturn(keyCommands);
        Mockito.when(keyCommands.keys("stock:*"))
            .thenReturn(Arrays.asList("stock:burger", "stock:fries"));
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(50L);
        Mockito.when(inventoryRepository.getStock(FRIES_ID)).thenReturn(100L);

        List<InventoryResponse> inventory = inventoryService.listAllInventory();

        assertNotNull(inventory);
        assertEquals(2, inventory.size());
        
        InventoryResponse burger = inventory.stream()
            .filter(i -> i.menuItemId().equals(BURGER_ID))
            .findFirst()
            .orElseThrow();
        assertEquals(50L, burger.currentStock());
        
        InventoryResponse fries = inventory.stream()
            .filter(i -> i.menuItemId().equals(FRIES_ID))
            .findFirst()
            .orElseThrow();
        assertEquals(100L, fries.currentStock());
    }

    @Test
    void testListAllInventory_ReturnsEmptyListWhenNoItems() {
        KeyCommands<String> keyCommands = Mockito.mock(KeyCommands.class);
        Mockito.when(redisDataSource.key()).thenReturn(keyCommands);
        Mockito.when(keyCommands.keys("stock:*")).thenReturn(List.of());

        List<InventoryResponse> inventory = inventoryService.listAllInventory();

        assertNotNull(inventory);
        assertTrue(inventory.isEmpty());
    }

    @Test
    void testBulkUpdateStock_UpdatesMultipleItems() {
        InventoryUpdateRequest burger = new InventoryUpdateRequest(BURGER_ID, 50L);
        InventoryUpdateRequest fries = new InventoryUpdateRequest(FRIES_ID, 100L);
        InventoryBulkUpdateRequest bulkRequest = new InventoryBulkUpdateRequest(Arrays.asList(burger, fries));

        List<InventoryResponse> responses = inventoryService.bulkUpdateStock(bulkRequest);

        assertNotNull(responses);
        assertEquals(2, responses.size());
        assertEquals(BURGER_ID, responses.get(0).menuItemId());
        assertEquals(50L, responses.get(0).currentStock());
        assertEquals(FRIES_ID, responses.get(1).menuItemId());
        assertEquals(100L, responses.get(1).currentStock());
        
        Mockito.verify(inventoryRepository).setStock(BURGER_ID, 50L);
        Mockito.verify(inventoryRepository).setStock(FRIES_ID, 100L);
    }

    @Test
    void testBulkUpdateStock_ThrowsExceptionOnFailure() {
        InventoryUpdateRequest burger = new InventoryUpdateRequest(BURGER_ID, 50L);
        InventoryUpdateRequest fries = new InventoryUpdateRequest(FRIES_ID, 100L);
        InventoryBulkUpdateRequest bulkRequest = new InventoryBulkUpdateRequest(Arrays.asList(burger, fries));
        
        // Simulate failure on second item
        Mockito.doNothing().when(inventoryRepository).setStock(BURGER_ID, 50L);
        Mockito.doThrow(new RuntimeException("Redis error")).when(inventoryRepository).setStock(FRIES_ID, 100L);

        assertThrows(RuntimeException.class, () -> inventoryService.bulkUpdateStock(bulkRequest));
        
        Mockito.verify(inventoryRepository).setStock(BURGER_ID, 50L);
        Mockito.verify(inventoryRepository).setStock(FRIES_ID, 100L);
    }

    @Test
    void testHasStock_ReturnsTrueWhenSufficientStock() {
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(50L);

        boolean result = inventoryService.hasStock(BURGER_ID, 30L);

        assertTrue(result);
        Mockito.verify(inventoryRepository).getStock(BURGER_ID);
    }

    @Test
    void testHasStock_ReturnsFalseWhenInsufficientStock() {
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(20L);

        boolean result = inventoryService.hasStock(BURGER_ID, 30L);

        assertFalse(result);
        Mockito.verify(inventoryRepository).getStock(BURGER_ID);
    }

    @Test
    void testHasStock_ReturnsTrueWhenExactStock() {
        Mockito.when(inventoryRepository.getStock(BURGER_ID)).thenReturn(50L);

        boolean result = inventoryService.hasStock(BURGER_ID, 50L);

        assertTrue(result);
    }
}
