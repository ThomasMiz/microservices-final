package io.chotel.reservations.domain.model;

import java.math.BigDecimal;
import java.time.OffsetDateTime;

/**
 * Represents a period of time in which guests are occupying a certain room in the hotel.
 *
 * @param id                       This reservation's unique ID.
 * @param room                     The room being rented.
 * @param guestId                  An opaque ID identifying the guest.
 * @param startDate                The date at which the reservation starts.
 * @param endDate                  The date at which the reservation ends.
 * @param billingFolderId          The ID of the folder in the billing microservice for expenses related to this
 *                                 reservation.
 * @param reservationBillingTicket The ID of the billing ticket that charges the guest for the room itself.
 * @param rentedHourlyPrice        The hourly price this room was rented for.
 * @param totalPrice               The total price for the room itself (does not include extras like room service).
 */
public record Reservation(
        Long id,
        Room room,
        String guestId,
        OffsetDateTime startDate,
        OffsetDateTime endDate,
        String billingFolderId,
        String reservationBillingTicket,
        BigDecimal rentedHourlyPrice,
        BigDecimal totalPrice
) {}
