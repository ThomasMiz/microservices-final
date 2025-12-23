package io.chotel.reservations.application.usecase;

import io.chotel.reservations.domain.exception.ReservationNotFoundException;
import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.repository.ReservationRepository;
import io.chotel.reservations.domain.usecase.GetReservationByIdUsecase;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

@Singleton
@RequiredArgsConstructor
public class GetReservationByIdUsecaseImpl implements GetReservationByIdUsecase {
    private final ReservationRepository reservationRepository;

    @Override
    public Reservation execute(long id) {
        return reservationRepository.findById(id).orElseThrow(ReservationNotFoundException::new);
    }
}
