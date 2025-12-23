package io.chotel.reservations.domain.model.request;

import io.chotel.reservations.domain.model.RoomCategory;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.Set;

public record SearchRoomsRequest(
        String textSearch,
        Boolean active,
        OffsetDateTime availableAfter,
        OffsetDateTime availableBefore,
        BigDecimal minPrice,
        BigDecimal maxPrice,
        Integer minCapacity,
        Integer maxCapacity,
        Set<RoomCategory> categories,
        PageRequest<SortBy> pageRequest
) {
    public enum SortBy {
        ID,
        NUMBER,
        PRICE,
        SEARCH_SIMILARITY
    }
}
