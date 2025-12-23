package io.chotel.reservations.domain.model.request;

import io.chotel.reservations.domain.model.RoomCategory;

import java.math.BigDecimal;

public record CreateRoomRequest(
        String number,
        boolean active,
        String name,
        String description,
        Integer maxCapacity,
        RoomCategory category,
        BigDecimal hourlyPrice
) {}
