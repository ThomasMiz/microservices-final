package io.chotel.reservations.domain.model.result;

import io.chotel.reservations.domain.model.Reservation;

public record CreateReservationResult(
        Reservation reservation
) {}
