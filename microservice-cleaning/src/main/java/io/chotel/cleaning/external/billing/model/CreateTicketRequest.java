package io.chotel.cleaning.external.billing.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NonNull;
import lombok.experimental.Accessors;

import java.math.BigDecimal;

@Getter
@Builder
@Accessors(fluent = true)
@AllArgsConstructor
public class CreateTicketRequest {
    private final @NonNull String folderId;
    private final @NonNull BigDecimal total;
    private final @NonNull String itemId;
    private final String description;
}
