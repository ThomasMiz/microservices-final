package io.chotel.reservations.infrastructure.jpa.entity;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.OffsetDateTime;

/**
 * Represents a period of time in which guests are occupying a certain room in the hotel.
 */
@Data
@Entity
@Table(name = "reservation")
@NoArgsConstructor
@AllArgsConstructor
public class ReservationJpaEntity {

    /**
     * This reservation's unique identifier.
     */
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    /**
     * The room being rented.
     */
    @ManyToOne
    @JoinColumn(name = "room_id")
    private RoomJpaEntity room;

    /**
     * An opaque ID identifying the guest.
     */
    @Column(name = "guest_id", nullable = false)
    private String guestId;

    /**
     * The date at which the reservation starts.
     */
    @Column(name = "start_date", nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    private OffsetDateTime startDate;

    /**
     * The date at which the reservation ends.
     */
    @Column(name = "end_date", nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    private OffsetDateTime endDate;

    /**
     * The ID of the folder in the billing microservice for expenses related to this reservation.
     */
    @Column(name = "billing_folder_id")
    private String billingFolderId;

    /**
     * The ID of the billing ticket that charges the guest for the room itself.
     */
    @Column(name = "reservation_billing_ticket")
    private String reservationBillingTicket;

    /**
     * The hourly price this room was rented for.
     */
    @Column(name = "rented_hourly_price", nullable = false, precision = 20, scale = 2)
    private BigDecimal rentedHourlyPrice;

    /**
     * The total price for the room itself (does not include extras like room service).
     */
    @Column(name = "total_price", nullable = false, precision = 20, scale = 2)
    private BigDecimal totalPrice;

    /**
     * Creates a new {@link ReservationJpaEntity} with null ID for insertion into the database.
     */
    public ReservationJpaEntity(RoomJpaEntity room, String guestId, OffsetDateTime startDate, OffsetDateTime endDate, String billingFolderId, String reservationBillingTicket, BigDecimal rentedHourlyPrice, BigDecimal totalPrice) {
        this.room = room;
        this.guestId = guestId;
        this.startDate = startDate;
        this.endDate = endDate;
        this.billingFolderId = billingFolderId;
        this.reservationBillingTicket = reservationBillingTicket;
        this.rentedHourlyPrice = rentedHourlyPrice;
        this.totalPrice = totalPrice;
    }
}
