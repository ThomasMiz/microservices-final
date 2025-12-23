package io.chotel.cleaning.controller.damage.json.request;

import io.micronaut.core.annotation.Nullable;
import io.micronaut.serde.annotation.Serdeable;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Positive;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CreateDamageReportJson {

    @NotBlank
    private String roomNumber;

    @NotBlank
    private String brokenItem;

    @Nullable
    private String description;

    @Positive
    private BigDecimal fineAmount;
}
