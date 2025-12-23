package io.chotel.cleaning.external.billing.config;

import io.micronaut.context.annotation.ConfigurationProperties;
import io.micronaut.context.annotation.Requires;

@ConfigurationProperties(BillingApiConfig.PREFIX)
@Requires(property = BillingApiConfig.PREFIX)
public record BillingApiConfig(
        String url
) {
    public static final String PREFIX = "service.billing";
}
