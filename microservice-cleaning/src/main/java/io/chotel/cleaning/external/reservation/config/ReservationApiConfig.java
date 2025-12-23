package io.chotel.cleaning.external.reservation.config;

import io.micronaut.context.annotation.ConfigurationProperties;
import io.micronaut.context.annotation.Requires;

@ConfigurationProperties(ReservationApiConfig.PREFIX)
@Requires(property = ReservationApiConfig.PREFIX)
public record ReservationApiConfig(
        String url
) {
    public static final String PREFIX = "service.reservation";
}
