package io.chotel.inventory.dto;

import java.time.Instant;

public record InventoryResponse(
    String menuItemId,
    long currentStock,
    Instant lastUpdated
) {}
