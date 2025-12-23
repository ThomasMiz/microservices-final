package io.chotel.inventory.dto;

import jakarta.validation.Valid;
import jakarta.validation.constraints.NotEmpty;
import java.util.List;

public record InventoryBulkUpdateRequest(
    @NotEmpty(message = "Items list cannot be empty")
    @Valid
    List<InventoryUpdateRequest> items
) {}
