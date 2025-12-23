package io.chotel.reservations.domain.model.result;

import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.model.request.SearchReservationsRequest;

public record SearchReservationsResult(
        SearchReservationsRequest request,
        PageResult<Reservation> page
) {}
