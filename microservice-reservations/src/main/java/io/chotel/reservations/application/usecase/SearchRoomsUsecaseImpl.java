package io.chotel.reservations.application.usecase;

import io.chotel.reservations.domain.model.Range;
import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.RoomWithAvailability;
import io.chotel.reservations.domain.model.request.SearchRoomsRequest;
import io.chotel.reservations.domain.model.result.PageResult;
import io.chotel.reservations.domain.model.result.SearchRoomsResult;
import io.chotel.reservations.domain.repository.RoomRepository;
import io.chotel.reservations.domain.service.RoomAvailabilityService;
import io.chotel.reservations.domain.usecase.SearchRoomsUsecase;
import io.micronaut.transaction.annotation.Transactional;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Singleton
@RequiredArgsConstructor
public class SearchRoomsUsecaseImpl implements SearchRoomsUsecase {
    private final Clock clock;

    private final RoomRepository roomRepository;
    private final RoomAvailabilityService roomAvailabilityService;

    @Override
    @Transactional
    public SearchRoomsResult execute(SearchRoomsRequest request) {
        PageResult<Room> roomPage = roomRepository.search(request);

        OffsetDateTime availabilityFrom = request.availableAfter();
        OffsetDateTime availabilityTo = request.availableBefore();

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

        Map<Long, List<Range<OffsetDateTime>>> reservationsByRoomId = roomAvailabilityService.getAvailabilityForRooms(
                roomPage.contents().stream().map(Room::id).collect(Collectors.toSet()),
                availabilityFrom,
                availabilityTo
        );

        PageResult<RoomWithAvailability> page = roomPage.map(
                room -> new RoomWithAvailability(room, reservationsByRoomId.get(room.id()))
        );

        return new SearchRoomsResult(request, page);
    }
}
