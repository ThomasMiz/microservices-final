package io.chotel.reservations.domain.external.billing.model;

import java.time.LocalDateTime;

public record BillingFolder(
        String id,
        LocalDateTime createdAt,
        LocalDateTime closedAt,
        String name
) {}
