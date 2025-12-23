package io.chotel.reservations.domain.usecase;

import io.chotel.reservations.domain.model.request.SearchReservationsRequest;
import io.chotel.reservations.domain.model.result.SearchReservationsResult;

public interface SearchReservationsUsecase {

    SearchReservationsResult execute(SearchReservationsRequest request);
}
