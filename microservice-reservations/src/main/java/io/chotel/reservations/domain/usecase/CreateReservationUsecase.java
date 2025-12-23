package io.chotel.reservations.domain.usecase;

import io.chotel.reservations.domain.model.request.CreateReservationRequest;
import io.chotel.reservations.domain.model.result.CreateReservationResult;

public interface CreateReservationUsecase {

    CreateReservationResult execute(CreateReservationRequest request);
}
