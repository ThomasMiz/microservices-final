package io.chotel.reservations.infrastructure.web.model;

import io.chotel.reservations.domain.model.request.SortDirection;

public enum SortDirectionJson {
    ASC,
    DESC;

    public SortDirection toDomain() {
        return switch (this) {
            case ASC -> SortDirection.ASCENDING;
            case DESC -> SortDirection.DESCENDING;
        };
    }
}
