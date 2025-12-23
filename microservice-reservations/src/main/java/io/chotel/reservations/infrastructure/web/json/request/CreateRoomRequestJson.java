package io.chotel.reservations.infrastructure.web.json.request;

import io.chotel.reservations.domain.model.RoomCategory;
import io.chotel.reservations.domain.model.request.CreateRoomRequest;
import io.micronaut.serde.annotation.Serdeable;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CreateRoomRequestJson {

    @NotBlank
    private String number;

    private boolean active = true;

    @NotBlank
    private String name;

    @NotBlank
    private String description;

    @NotNull
    @Positive
    private Integer maxCapacity;

    @NotNull
    private RoomCategory category;

    @NotNull
    @PositiveOrZero
    private BigDecimal hourlyPrice;

    public CreateRoomRequest toDomain() {
        return new CreateRoomRequest(
                number,
                active,
                name,
                description,
                maxCapacity,
                category,
                hourlyPrice
        );
    }
}
