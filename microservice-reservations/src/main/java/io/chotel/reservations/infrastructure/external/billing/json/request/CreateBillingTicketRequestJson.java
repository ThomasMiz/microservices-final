package io.chotel.reservations.infrastructure.external.billing.json.request;

import com.fasterxml.jackson.annotation.JsonProperty;
import io.chotel.reservations.domain.external.billing.model.CreateTicketRequest;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CreateBillingTicketRequestJson {

    @JsonProperty("folder_id")
    private String folderId;

    private BigDecimal total;

    @JsonProperty("item_id")
    private String itemId;

    private String description;

    public static CreateBillingTicketRequestJson from(CreateTicketRequest request) {
        return new CreateBillingTicketRequestJson(
                request.folderId(),
                request.total(),
                request.itemId(),
                request.description()
        );
    }
}
