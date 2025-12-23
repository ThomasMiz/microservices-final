package io.chotel.reservations.domain.exception;

public class RoomNotActiveException extends RuntimeException {
    public RoomNotActiveException() {
        super("Room not active");
    }
}
