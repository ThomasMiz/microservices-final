package io.chotel.reservations.domain.external.billing.model;

import lombok.Builder;
import lombok.Getter;
import lombok.NonNull;
import lombok.experimental.Accessors;

import java.math.BigDecimal;

@Getter
@Builder
@Accessors(fluent = true)
public class CreateTicketRequest {
    private final @NonNull String folderId;
    private final @NonNull BigDecimal total;
    private final @NonNull String itemId;
    private final String description;
}
