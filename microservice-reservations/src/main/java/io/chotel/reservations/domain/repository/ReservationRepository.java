package io.chotel.reservations.domain.repository;

import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.model.request.SearchReservationsRequest;
import io.chotel.reservations.domain.model.result.PageResult;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public interface ReservationRepository {

    Optional<Reservation> findById(long id);

    Reservation createReservation(
            Long roomId,
            String guestId,
            OffsetDateTime startDate,
            OffsetDateTime endDate,
            BigDecimal rentedHourlyPrice,
            BigDecimal totalPrice
    );

    Reservation setBillingInfo(
            Reservation reservation,
            String billingFolderId,
            String reservationBillingTicket
    );

    void deleteReservationById(long id);

    /**
     * Finds all the reservations for the given rooms that intersect the given time range.
     *
     * @return The reservations found in a map keyed by room ID.
     */
    Map<Long, List<Reservation>> findReservationsForRoomsBetweenDates(
            Collection<Long> roomIds,
            OffsetDateTime startDate,
            OffsetDateTime endDate
    );

    PageResult<Reservation> search(SearchReservationsRequest request);
}
