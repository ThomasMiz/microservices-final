package io.chotel.reservations.domain.model.request;

import java.time.OffsetDateTime;

public record CreateReservationRequest(
        long roomId,
        String guestId,
        OffsetDateTime startDate,
        OffsetDateTime endDate
) {}
