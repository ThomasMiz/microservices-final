package io.chotel.cleaning.controller.staff.query;

import io.micronaut.core.annotation.Introspected;
import io.micronaut.core.annotation.Nullable;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@Introspected
public class SearchCleaningStaffQueryParams {
    private @Nullable String name;

    private @Nullable Integer page;
    private @Nullable Integer pageSize;

    public int pageOrStart() {
        return page == null ? 0 : page;
    }

    public int pageSizeOr(int defaultSize) {
        return pageSize == null ? defaultSize : pageSize;
    }
}
