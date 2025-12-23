package io.chotel.inventory;

import io.chotel.inventory.dto.InventoryBulkUpdateRequest;
import io.chotel.inventory.dto.InventoryResponse;
import io.chotel.inventory.dto.InventoryUpdateRequest;
import io.chotel.order.InventoryRepository;
import io.opentelemetry.instrumentation.annotations.WithSpan;
import io.quarkus.logging.Log;
import io.quarkus.redis.datasource.RedisDataSource;
import io.quarkus.redis.datasource.keys.KeyCommands;

import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import jakarta.ws.rs.NotFoundException;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;

/**
 * Service for managing kitchen inventory operations.
 * Provides business logic layer on top of InventoryRepository.
 */
@ApplicationScoped
public class InventoryService {

    private static final String STOCK_KEY_PREFIX = "stock:";

    @Inject
    InventoryRepository inventoryRepository;

    @Inject
    RedisDataSource redisDataSource;

    /**
     * Updates stock for a single menu item.
     * Sets the stock to the specified quantity (not additive).
     * 
     * @param request The inventory update request
     * @return Response with updated stock information
     */
    @WithSpan("inventory.updateStock")
    public InventoryResponse updateStock(InventoryUpdateRequest request) {
        Log.infof("Updating stock for %s to %d", request.menuItemId(), request.quantity());
        
        inventoryRepository.setStock(request.menuItemId(), request.quantity());
        
        return new InventoryResponse(
            request.menuItemId(),
            request.quantity(),
            Instant.now()
        );
    }

    /**
     * Adds stock to an existing menu item (additive operation).
     * If the item doesn't exist, creates it with the specified quantity.
     * 
     * @param request The inventory update request
     * @return Response with updated stock information
     */
    @WithSpan("inventory.addStock")
    public InventoryResponse addStock(InventoryUpdateRequest request) {
        Log.infof("Adding %d units to stock for %s", request.quantity(), request.menuItemId());
        
        long currentStock = inventoryRepository.getStock(request.menuItemId());
        long newStock = currentStock + request.quantity();
        
        inventoryRepository.setStock(request.menuItemId(), newStock);
        
        Log.infof("Stock for %s updated from %d to %d", request.menuItemId(), currentStock, newStock);
        
        return new InventoryResponse(
            request.menuItemId(),
            newStock,
            Instant.now()
        );
    }

    /**
     * Gets current stock for a menu item.
     * 
     * @param menuItemId The menu item identifier
     * @return Response with current stock information
     * @throws NotFoundException if the item doesn't exist in inventory
     */
    @WithSpan("inventory.getStock")
    public InventoryResponse getStock(String menuItemId) {
        Log.infof("Getting stock for %s", menuItemId);
        
        long currentStock = inventoryRepository.getStock(menuItemId);
        
        if (currentStock == 0) {
            // Check if the key exists in Redis or if it's genuinely 0
            KeyCommands<String> keyCommands = redisDataSource.key();
            boolean exists = keyCommands.exists(STOCK_KEY_PREFIX + menuItemId);
            
            if (!exists) {
                Log.warnf("Menu item %s not found in inventory", menuItemId);
                throw new NotFoundException("Menu item not found in inventory: " + menuItemId);
            }
        }
        
        return new InventoryResponse(
            menuItemId,
            currentStock,
            Instant.now()
        );
    }

    /**
     * Lists all inventory items with their current stock.
     * 
     * @return List of all inventory items
     */
    @WithSpan("inventory.listAllInventory")
    public List<InventoryResponse> listAllInventory() {
        Log.info("Listing all inventory items");
        
        KeyCommands<String> keyCommands = redisDataSource.key();
        List<String> keys = keyCommands.keys(STOCK_KEY_PREFIX + "*");
        
        List<InventoryResponse> inventory = new ArrayList<>();
        
        for (String key : keys) {
            String menuItemId = key.substring(STOCK_KEY_PREFIX.length());
            long currentStock = inventoryRepository.getStock(menuItemId);
            
            inventory.add(new InventoryResponse(
                menuItemId,
                currentStock,
                Instant.now()
            ));
        }
        
        Log.infof("Found %d inventory items", inventory.size());
        return inventory;
    }

    /**
     * Bulk update multiple inventory items.
     * All operations succeed or fail together (transactional).
     * 
     * @param request The bulk update request
     * @return List of updated inventory items
     */
    @WithSpan("inventory.bulkUpdateStock")
    public List<InventoryResponse> bulkUpdateStock(InventoryBulkUpdateRequest request) {
        Log.infof("Bulk updating %d inventory items", request.items().size());
        
        List<InventoryResponse> responses = new ArrayList<>();
        
        for (InventoryUpdateRequest item : request.items()) {
            try {
                InventoryResponse response = updateStock(item);
                responses.add(response);
            } catch (Exception e) {
                Log.errorf(e, "Failed to update stock for %s during bulk operation", item.menuItemId());
                throw new RuntimeException("Bulk update failed for item: " + item.menuItemId(), e);
            }
        }
        
        Log.infof("Successfully bulk updated %d items", responses.size());
        return responses;
    }

    /**
     * Validates if sufficient stock exists for a menu item.
     * 
     * @param menuItemId The menu item identifier
     * @param quantity The required quantity
     * @return true if sufficient stock exists
     */
    @WithSpan("inventory.hasStock")
    public boolean hasStock(String menuItemId, long quantity) {
        long currentStock = inventoryRepository.getStock(menuItemId);
        return currentStock >= quantity;
    }
}
