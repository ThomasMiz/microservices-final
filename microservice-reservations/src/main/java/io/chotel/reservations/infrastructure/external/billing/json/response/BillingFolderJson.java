package io.chotel.reservations.infrastructure.external.billing.json.response;

import com.fasterxml.jackson.annotation.JsonProperty;
import io.chotel.reservations.domain.external.billing.model.BillingFolder;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class BillingFolderJson {

    private String id;

    @JsonProperty("created_at")
    private LocalDateTime createdAt;

    @JsonProperty("closed_at")
    private LocalDateTime closedAt;

    private String name;

    public BillingFolder toDomain() {
        return new BillingFolder(
                id,
                createdAt,
                closedAt,
                name
        );
    }
}
