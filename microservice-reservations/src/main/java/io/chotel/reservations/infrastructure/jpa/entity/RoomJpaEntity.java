package io.chotel.reservations.infrastructure.jpa.entity;

import io.chotel.reservations.domain.model.RoomCategory;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

/**
 * Represents a room in the hotel.
 */
@Data
@Entity
@Table(name = "room")
@NoArgsConstructor
@AllArgsConstructor
public class RoomJpaEntity {
    public static final String UNIQUE_NUMBER_WHERE_ACTIVE_CONSTRAINT = "room_unique_number_where_active";

    /**
     * This room's unique identifier.
     */
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    /**
     * The room's number. The format of this field is opaque to this entity, a typical format is to use the formula
     * {@code (floorNumber * 100) + roomNumberWithinFloor}, so, for example, the third room of the first floor would be
     * 103.
     */
    @Column(name = "number", nullable = false)
    private String number;

    /**
     * Whether this room is active or not.
     */
    @Column(name = "active", nullable = false)
    private boolean active;

    /**
     * A human-friendly name for this room.
     */
    @Column(name = "name")
    private String name;

    /**
     * A human-friendly description for this room.
     */
    @Column(name = "description")
    private String description;

    /**
     * The maximum number of guests the room can accommodate.
     */
    @Column(name = "max_capacity", nullable = false)
    private int maxCapacity;

    @Column(name = "category")
    @Enumerated(EnumType.STRING)
    private RoomCategory category;

    /**
     * The current hourly price this room goes for. Reservations may have a different price as they have what the price
     * was at the time the reservation was performed.
     */
    @Column(name = "hourly_price", precision = 20, scale = 2)
    private BigDecimal hourlyPrice;

    public RoomJpaEntity(String number, boolean active, String name, String description, int maxCapacity, RoomCategory category, BigDecimal hourlyPrice) {
        this.number = number;
        this.active = active;
        this.name = name;
        this.description = description;
        this.maxCapacity = maxCapacity;
        this.category = category;
        this.hourlyPrice = hourlyPrice;
    }
}
