package io.chotel.reservations.domain.exception;

public class RoomNotFoundException extends NotFoundException {
    public RoomNotFoundException() {
        super("Room not found");
    }
}
