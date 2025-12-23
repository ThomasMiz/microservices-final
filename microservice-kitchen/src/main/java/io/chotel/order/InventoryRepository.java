package io.chotel.order;

import io.opentelemetry.instrumentation.annotations.WithSpan;
import io.quarkus.logging.Log;
import io.quarkus.redis.datasource.ReactiveRedisDataSource;
import io.quarkus.redis.datasource.RedisDataSource;
import io.quarkus.redis.datasource.value.ValueCommands;
import io.smallrye.mutiny.Uni;
import jakarta.enterprise.context.ApplicationScoped;

/**
 * Repository for managing inventory stock in Redis.
 * Uses atomic operations to ensure consistency under concurrent access.
 * 
 * Stock keys follow the pattern: stock:{menuItemId}
 */
@ApplicationScoped
public class InventoryRepository {

    private static final String STOCK_KEY_PREFIX = "stock:";

    private final ValueCommands<String, Long> syncCommands;
    private final ReactiveRedisDataSource reactiveDataSource;

    public InventoryRepository(RedisDataSource redisDataSource, ReactiveRedisDataSource reactiveDataSource) {
        this.syncCommands = redisDataSource.value(Long.class);
        this.reactiveDataSource = reactiveDataSource;
    }

    /**
     * Attempts to reserve stock for a menu item atomically.
     * Uses DECRBY and checks if the result is >= 0.
     * If the result would be negative, the stock is restored.
     * 
     * @param menuItemId The menu item identifier
     * @param quantity The quantity to reserve
     * @return true if stock was successfully reserved, false if insufficient stock
     */
    @WithSpan("inventory.reserveStock")
    public boolean reserveStock(String menuItemId, int quantity) {
        String key = getStockKey(menuItemId);
        
        try {
            // Atomic decrement - returns the new value after decrementing
            Long newValue = syncCommands.decrby(key, quantity);
            
            if (newValue == null) {
                // Key doesn't exist, treat as no stock available
                Log.warnf("Stock key %s does not exist, cannot reserve", key);
                return false;
            }
            
            if (newValue < 0) {
                // Not enough stock, restore what we decremented
                Log.infof("Insufficient stock for %s. Attempted: %d, would result in: %d. Restoring.", 
                    menuItemId, quantity, newValue);
                syncCommands.incrby(key, quantity);
                return false;
            }
            
            Log.infof("Successfully reserved %d units of %s. Remaining stock: %d", 
                quantity, menuItemId, newValue);
            return true;
            
        } catch (Exception e) {
            Log.errorf(e, "Error reserving stock for %s", menuItemId);
            return false;
        }
    }

    /**
     * Restores stock for a menu item (compensation for failed billing).
     * 
     * @param menuItemId The menu item identifier
     * @param quantity The quantity to restore
     */
    @WithSpan("inventory.restoreStock")
    public void restoreStock(String menuItemId, int quantity) {
        String key = getStockKey(menuItemId);
        
        try {
            Long newValue = syncCommands.incrby(key, quantity);
            Log.infof("Restored %d units of %s. New stock: %d", quantity, menuItemId, newValue);
        } catch (Exception e) {
            Log.errorf(e, "Error restoring stock for %s", menuItemId);
        }
    }

    /**
     * Gets the current stock for a menu item.
     * 
     * @param menuItemId The menu item identifier
     * @return The current stock, or 0 if not found
     */
    @WithSpan("inventory.getStock")
    public long getStock(String menuItemId) {
        String key = getStockKey(menuItemId);
        Long value = syncCommands.get(key);
        return value != null ? value : 0;
    }

    /**
     * Sets the stock for a menu item (for initialization/admin purposes).
     * 
     * @param menuItemId The menu item identifier
     * @param quantity The quantity to set
     */
    @WithSpan("inventory.setStock")
    public void setStock(String menuItemId, long quantity) {
        String key = getStockKey(menuItemId);
        syncCommands.set(key, quantity);
        Log.infof("Set stock for %s to %d", menuItemId, quantity);
    }

    private String getStockKey(String menuItemId) {
        return STOCK_KEY_PREFIX + menuItemId;
    }
}
