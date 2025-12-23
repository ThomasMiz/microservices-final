package io.chotel.reservations.infrastructure.web.json.request;

import io.chotel.reservations.domain.model.request.CreateReservationRequest;
import io.micronaut.serde.annotation.Serdeable;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CreateReservationRequestJson {

    @NotNull
    @PositiveOrZero
    private Long roomId;

    @NotBlank
    private String guestId;

    @NotNull
    private OffsetDateTime startDate;

    @NotNull
    private OffsetDateTime endDate;

    public CreateReservationRequest toDomain() {
        return new CreateReservationRequest(
                roomId,
                guestId,
                startDate,
                endDate
        );
    }
}
