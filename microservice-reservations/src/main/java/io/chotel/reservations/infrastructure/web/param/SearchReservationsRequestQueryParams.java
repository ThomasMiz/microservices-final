package io.chotel.reservations.infrastructure.web.param;

import io.chotel.reservations.domain.model.request.PageRequest;
import io.chotel.reservations.domain.model.request.SearchReservationsRequest;
import io.chotel.reservations.infrastructure.web.model.SortDirectionJson;
import io.micronaut.core.annotation.Introspected;
import io.micronaut.core.annotation.Nullable;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;
import java.util.Set;

@Data
@NoArgsConstructor
@Introspected
public class SearchReservationsRequestQueryParams {
    private @Nullable OffsetDateTime from;
    private @Nullable OffsetDateTime to;
    private @Nullable Set<Long> roomIds;
    private @Nullable Set<String> roomNumbers;
    private @Nullable String guestId;

    private @Nullable Integer page;
    private @Nullable Integer pageSize;
    private @Nullable SortBy sortBy;
    private @Nullable SortDirectionJson direction;

    public enum SortBy {
        ID,
        STARTDATE,
        ENDDATE,
        ROOMID;

        public SearchReservationsRequest.SortBy toDomain() {
            return switch (this) {
                case ID -> SearchReservationsRequest.SortBy.ID;
                case STARTDATE -> SearchReservationsRequest.SortBy.START_DATE;
                case ENDDATE -> SearchReservationsRequest.SortBy.END_DATE;
                case ROOMID -> SearchReservationsRequest.SortBy.ROOM_ID;
            };
        }
    }

    public int pageOrStart() {
        return page == null ? 0 : page;
    }

    public int pageSizeOr(int defaultSize) {
        return pageSize == null ? defaultSize : pageSize;
    }

    public SearchReservationsRequest toDomain() {
        PageRequest<SearchReservationsRequest.SortBy> pageRequest = new PageRequest<>(
                pageOrStart(),
                pageSizeOr(20),
                sortBy == null ? null : sortBy.toDomain(),
                direction == null ? null : direction.toDomain()
        );

        return new SearchReservationsRequest(
                from,
                to,
                roomIds,
                roomNumbers,
                guestId,
                pageRequest
        );
    }
}
