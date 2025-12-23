package io.chotel.reservations.domain.usecase;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.request.CreateRoomRequest;

public interface CreateRoomUsecase {

    Room execute(CreateRoomRequest request);
}
