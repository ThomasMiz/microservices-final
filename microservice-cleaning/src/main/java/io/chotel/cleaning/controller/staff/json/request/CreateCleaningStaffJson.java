package io.chotel.cleaning.controller.staff.json.request;

import io.micronaut.serde.annotation.Serdeable;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CreateCleaningStaffJson {

    @NotBlank
    @Pattern(regexp = "[a-zA-Z][a-zA-Z0-9\\-_ ]+[a-zA-Z0-9]")
    private String name;
}
