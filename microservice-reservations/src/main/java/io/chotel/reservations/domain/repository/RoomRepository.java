package io.chotel.reservations.domain.repository;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.request.CreateRoomRequest;
import io.chotel.reservations.domain.model.request.SearchRoomsRequest;
import io.chotel.reservations.domain.model.result.PageResult;

import java.util.Optional;

public interface RoomRepository {

    Optional<Room> findById(Long id);

    Room createRoom(CreateRoomRequest request);

    PageResult<Room> search(SearchRoomsRequest request);
}
