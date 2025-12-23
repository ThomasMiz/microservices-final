package io.chotel.cleaning.external.reservation.query;

import io.micronaut.core.annotation.Introspected;
import io.micronaut.core.annotation.Nullable;
import io.micronaut.data.model.Sort;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;
import java.util.Set;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
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
    private @Nullable Sort.Order.Direction direction;

    public enum SortBy {
        ID,
        STARTDATE,
        ENDDATE,
        ROOMID;
    }
}
