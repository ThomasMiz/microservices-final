package io.chotel.reservations.domain.exception;

import io.chotel.reservations.domain.model.Reservation;

import java.util.List;

public class RoomNotAvailableException extends RuntimeException {
    private final List<Reservation> intersectingReservations;

    public RoomNotAvailableException(List<Reservation> intersectingReservations) {
        super("Room not available in the requested time range");
        this.intersectingReservations = intersectingReservations;
    }
}
