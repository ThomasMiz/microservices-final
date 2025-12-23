package io.chotel.cleaning.external.reservation.json.response;

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
}
