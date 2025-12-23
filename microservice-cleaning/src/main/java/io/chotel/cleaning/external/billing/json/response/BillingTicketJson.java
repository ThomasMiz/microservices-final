package io.chotel.cleaning.external.billing.json.response;

import com.fasterxml.jackson.annotation.JsonProperty;
import io.chotel.cleaning.external.billing.json.BillingTicketStateJson;
import io.chotel.cleaning.external.billing.model.BillingTicket;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class BillingTicketJson {

    private String id;

    private BillingFolderJson folder;

    private BigDecimal total;

    @JsonProperty("item_id")
    private String itemId;

    private String description;

    @JsonProperty("created_at")
    private LocalDateTime createdAt;

    @JsonProperty("closed_at")
    private LocalDateTime closedAt;

    private BillingTicketStateJson state;

    public BillingTicket toDomain() {
        return new BillingTicket(
                id,
                folder.toDomain(),
                total,
                itemId,
                description,
                createdAt,
                closedAt,
                state.domainState()
        );
    }
}
