package io.chotel.reservations.application.usecase;

import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.model.request.SearchReservationsRequest;
import io.chotel.reservations.domain.model.result.PageResult;
import io.chotel.reservations.domain.model.result.SearchReservationsResult;
import io.chotel.reservations.domain.repository.ReservationRepository;
import io.chotel.reservations.domain.usecase.SearchReservationsUsecase;
import io.micronaut.transaction.annotation.Transactional;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

@Singleton
@RequiredArgsConstructor
public class SearchReservationsUsecaseImpl implements SearchReservationsUsecase {
    private final ReservationRepository reservationRepository;

    @Override
    @Transactional
    public SearchReservationsResult execute(SearchReservationsRequest request) {
        PageResult<Reservation> page = reservationRepository.search(request);
        return new SearchReservationsResult(request, page);
    }
}
