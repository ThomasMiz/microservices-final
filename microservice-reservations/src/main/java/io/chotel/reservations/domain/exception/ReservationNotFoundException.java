package io.chotel.reservations.domain.exception;

public class ReservationNotFoundException extends NotFoundException {
    public ReservationNotFoundException() {
        super("Reservation not found");
    }
}
