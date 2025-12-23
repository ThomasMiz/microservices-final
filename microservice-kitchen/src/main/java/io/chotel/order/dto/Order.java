package io.chotel.order.dto;

import java.time.Instant;
import java.util.List;

public record Order(
    String id,
    String roomId,
    String reservationId,
    List<Item> items,
    String totalPrice,
    String status,
    Instant createdAt
) {}
