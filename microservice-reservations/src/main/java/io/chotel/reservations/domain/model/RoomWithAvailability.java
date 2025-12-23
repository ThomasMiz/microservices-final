package io.chotel.reservations.domain.model;

import java.time.OffsetDateTime;
import java.util.List;

public record RoomWithAvailability(
        Room room,
        List<Range<OffsetDateTime>> availability
) {}
