package io.chotel.reservations.domain.model.request;

import java.time.OffsetDateTime;
import java.util.Set;

public record SearchReservationsRequest(
        OffsetDateTime from,
        OffsetDateTime to,
        Set<Long> roomIds,
        Set<String> roomNumbers,
        String guestId,
        PageRequest<SortBy> pageRequest
) {
    public enum SortBy {
        ID,
        START_DATE,
        END_DATE,
        ROOM_ID
    }
}
