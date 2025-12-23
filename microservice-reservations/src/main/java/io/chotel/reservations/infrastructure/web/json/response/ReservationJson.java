package io.chotel.reservations.infrastructure.web.json.response;

import io.chotel.reservations.domain.model.Reservation;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class ReservationJson {
    private Long id;
    private RoomJson room;
    private String guestId;
    private OffsetDateTime startDate;
    private OffsetDateTime endDate;
    private String billingFolderId;
    private String reservationBillingTicket;
    private BigDecimal rentedHourlyPrice;
    private BigDecimal totalPrice;

    public void fillWith(Reservation reservation) {
        this.id = reservation.id();
        this.room = RoomJson.from(reservation.room());
        this.guestId = reservation.guestId();
        this.startDate = reservation.startDate();
        this.endDate = reservation.endDate();
        this.billingFolderId = reservation.billingFolderId();
        this.reservationBillingTicket = reservation.reservationBillingTicket();
        this.rentedHourlyPrice = reservation.rentedHourlyPrice();
        this.totalPrice = reservation.totalPrice();
    }

    public static ReservationJson from(Reservation reservation) {
        if (reservation == null) {
            return null;
        }

        ReservationJson json = new ReservationJson();
        json.fillWith(reservation);
        return json;
    }
}
