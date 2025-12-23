package io.chotel.reservations.application.usecase;

import io.chotel.reservations.domain.exception.RoomNotFoundException;
import io.chotel.reservations.domain.model.Range;
import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.RoomWithAvailability;
import io.chotel.reservations.domain.repository.RoomRepository;
import io.chotel.reservations.domain.service.RoomAvailabilityService;
import io.chotel.reservations.domain.usecase.GetRoomByIdUsecase;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.List;

@Singleton
@RequiredArgsConstructor
public class GetRoomByIdUsecaseImpl implements GetRoomByIdUsecase {
    private final Clock clock;

    private final RoomRepository roomRepository;
    private final RoomAvailabilityService roomAvailabilityService;

    @Override
    public RoomWithAvailability execute(
            long id,
            OffsetDateTime availabilityFrom,
            OffsetDateTime availabilityTo
    ) {
        if (availabilityFrom == null || availabilityTo == null) {
            OffsetDateTime now = OffsetDateTime.now(clock);

            if (availabilityFrom == null && availabilityTo == null) {
                availabilityFrom = now.minusDays(20);
                availabilityTo = now.plusDays(20);
            } else if (availabilityFrom == null) {
                availabilityFrom = availabilityTo.minusDays(40);
            } else { // availabilityTo is null
                availabilityTo = availabilityFrom.plusDays(40);
            }
        }

        Room room = roomRepository.findById(id).orElseThrow(RoomNotFoundException::new);
        List<Range<OffsetDateTime>> availability = roomAvailabilityService.getAvailabilityForRoom(
                id,
                availabilityFrom,
                availabilityTo
        );

        return new RoomWithAvailability(room, availability);
    }
}
