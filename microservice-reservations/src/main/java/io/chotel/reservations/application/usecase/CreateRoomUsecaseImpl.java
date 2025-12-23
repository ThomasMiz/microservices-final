package io.chotel.reservations.application.usecase;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.request.CreateRoomRequest;
import io.chotel.reservations.domain.repository.RoomRepository;
import io.chotel.reservations.domain.usecase.CreateRoomUsecase;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

@Singleton
@RequiredArgsConstructor
public class CreateRoomUsecaseImpl implements CreateRoomUsecase {
    private final RoomRepository roomRepository;

    @Override
    public Room execute(CreateRoomRequest request) {
        return roomRepository.createRoom(request);
    }
}
