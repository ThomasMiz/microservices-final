package io.chotel.reservations.domain.model;

import java.math.BigDecimal;

/**
 *
 * Represents a room in the hotel.
 *
 * @param id          This room's unique identifier.
 * @param number      The room's number. The format of this field is opaque to this entity, a typical format is to use
 *                    the formula {@code (floorNumber * 100) + roomNumberWithinFloor}, so, for example, the third room
 *                    of the first floor would be 103.
 * @param active      Whether this room is active or not.
 * @param name        A human-friendly name for this room.
 * @param description A human-friendly description for this room.
 * @param maxCapacity The maximum number of guests the room can accommodate.
 * @param hourlyPrice The current hourly price this room goes for. Reservations may have a different price as they have
 *                    what the price was at the time the reservation was performed.
 */
public record Room(
        Long id,
        String number,
        boolean active,
        String name,
        String description,
        int maxCapacity,
        RoomCategory category,
        BigDecimal hourlyPrice
) {}
