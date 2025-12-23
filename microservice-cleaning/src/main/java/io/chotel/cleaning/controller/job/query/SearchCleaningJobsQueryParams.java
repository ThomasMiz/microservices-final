package io.chotel.cleaning.controller.job.query;

import io.micronaut.core.annotation.Introspected;
import io.micronaut.core.annotation.Nullable;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@Introspected
public class SearchCleaningJobsQueryParams {
    private @Nullable String roomNumber;
    private @Nullable Long assignedTo;

    private @Nullable Integer page;
    private @Nullable Integer pageSize;

    public int pageOrStart() {
        return page == null ? 0 : page;
    }

    public int pageSizeOr(int defaultSize) {
        return pageSize == null ? defaultSize : pageSize;
    }
}
