package io.chotel.reservations.domain.util;

import lombok.experimental.UtilityClass;

@UtilityClass
public final class ObjectUtils {

    /**
     * Returns the first non-null value.
     */
    public static <T> T coalesce(T obj1, T obj2) {
        if (obj1 != null) {
            return obj1;
        }

        return obj2;
    }

    /**
     * Returns the first non-null value.
     */
    public static <T> T coalesce(T obj1, T obj2, T obj3) {
        if (obj1 != null) {
            return obj1;
        }

        if (obj2 != null) {
            return obj2;
        }

        return obj3;
    }

    /**
     * Returns the first non-null value.
     */
    @SafeVarargs
    public static <T> T coalesce(T... objects) {
        for (T obj : objects) {
            if (obj != null) {
                return obj;
            }
        }

        return null;
    }
}
