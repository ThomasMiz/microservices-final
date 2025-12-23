package io.chotel.reservations.infrastructure.web.param;

import io.chotel.reservations.domain.model.RoomCategory;
import io.chotel.reservations.domain.model.request.PageRequest;
import io.chotel.reservations.domain.model.request.SearchRoomsRequest;
import io.chotel.reservations.infrastructure.web.model.SortDirectionJson;
import io.micronaut.core.annotation.Introspected;
import io.micronaut.core.annotation.Nullable;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.Set;

@Data
@NoArgsConstructor
@Introspected
public class SearchRoomsRequestQueryParams {
    private @Nullable String search;
    private @Nullable Boolean active;
    private @Nullable OffsetDateTime availableAfter;
    private @Nullable OffsetDateTime availableBefore;
    private @Nullable BigDecimal minPrice;
    private @Nullable BigDecimal maxPrice;
    private @Nullable Integer minCapacity;
    private @Nullable Integer maxCapacity;
    private @Nullable Set<RoomCategory> categories;

    private @Nullable Integer page;
    private @Nullable Integer pageSize;
    private @Nullable SortBy sortBy;
    private @Nullable SortDirectionJson direction;

    public enum SortBy {
        ID,
        NUMBER,
        PRICE,
        SIMILARITY;

        public SearchRoomsRequest.SortBy toDomain() {
            return switch (this) {
                case ID -> SearchRoomsRequest.SortBy.ID;
                case NUMBER -> SearchRoomsRequest.SortBy.NUMBER;
                case PRICE -> SearchRoomsRequest.SortBy.PRICE;
                case SIMILARITY -> SearchRoomsRequest.SortBy.SEARCH_SIMILARITY;
            };
        }
    }

    public int pageOrStart() {
        return page == null ? 0 : page;
    }

    public int pageSizeOr(int defaultSize) {
        return pageSize == null ? defaultSize : pageSize;
    }

    public SearchRoomsRequest toDomain() {
        PageRequest<SearchRoomsRequest.SortBy> pageRequest = new PageRequest<>(
                pageOrStart(),
                pageSizeOr(20),
                sortBy == null ? null : sortBy.toDomain(),
                direction == null ? null : direction.toDomain()
        );

        return new SearchRoomsRequest(
                search,
                active,
                availableAfter,
                availableBefore,
                minPrice, maxPrice,
                minCapacity,
                maxCapacity,
                categories,
                pageRequest
        );
    }
}
