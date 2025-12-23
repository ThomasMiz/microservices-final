package io.chotel.reservations.domain.model;

public record Range<T>(
        T from,
        T to
) {}
