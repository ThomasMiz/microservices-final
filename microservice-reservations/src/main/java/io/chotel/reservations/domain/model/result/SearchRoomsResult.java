package io.chotel.reservations.domain.model.result;

import io.chotel.reservations.domain.model.RoomWithAvailability;
import io.chotel.reservations.domain.model.request.SearchRoomsRequest;

public record SearchRoomsResult(
        SearchRoomsRequest request,
        PageResult<RoomWithAvailability> page
) {}
