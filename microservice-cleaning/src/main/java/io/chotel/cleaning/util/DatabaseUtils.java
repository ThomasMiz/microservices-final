package io.chotel.cleaning.util;

import lombok.experimental.UtilityClass;
import org.apache.commons.lang3.StringUtils;

import java.util.regex.Pattern;

@UtilityClass
public final class DatabaseUtils {

    public static final String NON_ALPHANUMERIC_REGEX = "[^a-zA-Z0-9]+";
    public static final Pattern NON_ALPHANUMERIC_PATTERN = Pattern.compile(NON_ALPHANUMERIC_REGEX);

    public static String makeSearchParam(String search) {
        if (search == null) {
            return null;
        }


        search = StringUtils.stripAccents(search);
        search = NON_ALPHANUMERIC_PATTERN.matcher(search).replaceAll("");

        return search == null || search.isEmpty() ? null : ("%" + search + "%");
    }
}
