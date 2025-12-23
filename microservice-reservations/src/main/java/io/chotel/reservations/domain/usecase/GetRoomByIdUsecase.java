package io.chotel.reservations.domain.usecase;

import io.chotel.reservations.domain.model.RoomWithAvailability;

import java.time.OffsetDateTime;

public interface GetRoomByIdUsecase {

    RoomWithAvailability execute(
            long id,
            OffsetDateTime availabilityFrom,
            OffsetDateTime availabilityTo
    );
}
