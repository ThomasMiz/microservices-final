package io.chotel.reservations.domain.model.request;

/**
 * Used to specify how to paginate the results of a query.
 *
 * @param <S> The type of ordering. Must be immutable. If a query does not permit ordering, use {@link Void}.
 */
public record PageRequest<S>(
        int pageNumber,
        int pageSize,
        S sortBy,
        SortDirection sortDirection
) {
    public PageRequest {
        if (pageNumber < 0) {
            throw new IllegalArgumentException("pageNumber must be 0 or positive");
        }

        if (pageSize < 1) {
            throw new IllegalArgumentException("pageSize must be positive");
        }
    }

    public int firstElementOffset() {
        return Math.multiplyExact(pageNumber, pageSize);
    }
}
