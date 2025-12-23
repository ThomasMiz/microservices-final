package io.chotel.reservations.domain.service;

import io.chotel.reservations.domain.model.Range;
import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.repository.ReservationRepository;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

import java.time.OffsetDateTime;
import java.util.*;

@Singleton
@RequiredArgsConstructor
public class RoomAvailabilityService {
    private final ReservationRepository reservationRepository;

    private static List<Range<OffsetDateTime>> calculateAvailability(
            OffsetDateTime startDate,
            OffsetDateTime endDate,
            List<Reservation> reservations
    ) {
        if (reservations == null || reservations.isEmpty()) {
            return List.of(new Range<>(startDate, endDate));
        }

        List<Range<OffsetDateTime>> result = new ArrayList<>();

        OffsetDateTime prev = startDate;
        for (Reservation res : reservations) {
            result.add(new Range<>(prev, res.startDate()));
            prev = res.endDate();
        }
        result.add(new Range<>(prev, endDate));

        result.removeIf(range -> !range.from().isBefore(range.to()));
        return result;
    }

    public List<Range<OffsetDateTime>> getAvailabilityForRoom(
            long roomId,
            OffsetDateTime availabilityFrom,
            OffsetDateTime availabilityTo
    ) {
        return getAvailabilityForRooms(Set.of(roomId), availabilityFrom, availabilityTo).get(roomId);
    }

    public Map<Long, List<Range<OffsetDateTime>>> getAvailabilityForRooms(
            Set<Long> roomIds,
            OffsetDateTime availabilityFrom,
            OffsetDateTime availabilityTo
    ) {
        Map<Long, List<Reservation>> reservationsByRoomId = reservationRepository.findReservationsForRoomsBetweenDates(
                roomIds,
                availabilityFrom,
                availabilityTo
        );

        Map<Long, List<Range<OffsetDateTime>>> result = HashMap.newHashMap(reservationsByRoomId.size());
        reservationsByRoomId.forEach((roomId, reservations) -> {
            result.put(roomId, calculateAvailability(availabilityFrom, availabilityTo, reservations));
        });

        roomIds.forEach(roomId -> {
            result.computeIfAbsent(roomId, k -> List.of(new Range<>(availabilityFrom, availabilityTo)));
        });

        return result;
    }
}
