package io.chotel.reservations.infrastructure.external.billing.json.request;

import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CreateBillingFolderRequestJson {
    private String name;
}
