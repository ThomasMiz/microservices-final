package io.chotel.inventory.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;

public record InventoryUpdateRequest(
    @NotBlank(message = "Menu item ID is required")
    String menuItemId,
    
    @Positive(message = "Quantity must be positive")
    long quantity
) {}
