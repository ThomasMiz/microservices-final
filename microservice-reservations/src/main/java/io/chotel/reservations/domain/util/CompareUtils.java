package io.chotel.reservations.domain.util;

import lombok.experimental.UtilityClass;

@UtilityClass
public final class CompareUtils {

    public static <T extends Comparable<? super T>> T min(T a, T b) {
        return a.compareTo(b) <= 0 ? a : b;
    }

    public static <T extends Comparable<? super T>> T max(T a, T b) {
        return a.compareTo(b) >= 0 ? a : b;
    }
}
