package io.chotel.reservations.domain.external.billing.model;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public record BillingTicket(
        String id,
        BillingFolder folder,
        BigDecimal total,
        String itemId,
        String description,
        LocalDateTime createdAt,
        LocalDateTime closedAt,
        BillingTicketState state
) {}
