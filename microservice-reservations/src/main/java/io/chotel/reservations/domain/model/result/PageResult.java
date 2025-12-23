package io.chotel.reservations.domain.model.result;

import java.util.List;
import java.util.function.Function;

public record PageResult<T>(
        List<T> contents,
        long totalElements
) {
    public <S> PageResult<S> map(Function<? super T, S> mapper) {
        return new PageResult<>(
                contents.stream().map(mapper).toList(),
                totalElements
        );
    }
}
