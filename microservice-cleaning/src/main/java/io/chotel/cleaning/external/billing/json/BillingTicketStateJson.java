package io.chotel.cleaning.external.billing.json;

import com.fasterxml.jackson.annotation.JsonProperty;
import io.chotel.cleaning.external.billing.model.BillingTicketState;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AccessLevel;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;

@Getter
@Accessors(fluent = true)
@Serdeable
@RequiredArgsConstructor(access = AccessLevel.PRIVATE)
public enum BillingTicketStateJson {
    @JsonProperty("Pending") PENDING(BillingTicketState.PENDING),
    @JsonProperty("Paying") PAYING(BillingTicketState.PAYING),
    @JsonProperty("Paid") PAID(BillingTicketState.PAID),
    @JsonProperty("Canceled") CANCELED(BillingTicketState.CANCELED);

    private final BillingTicketState domainState;

    public static BillingTicketStateJson of(BillingTicketState domainState) {
        if (domainState == null) {
            return null;
        }

        return switch (domainState) {
            case PENDING -> PENDING;
            case PAYING -> PAYING;
            case PAID -> PAID;
            case CANCELED -> CANCELED;
        };
    }
}
