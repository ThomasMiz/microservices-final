package io.chotel.reservations.domain.usecase;

import io.chotel.reservations.domain.model.Reservation;

public interface GetReservationByIdUsecase {

    Reservation execute(long id);
}
