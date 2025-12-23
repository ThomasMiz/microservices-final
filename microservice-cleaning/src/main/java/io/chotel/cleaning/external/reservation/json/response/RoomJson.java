package io.chotel.cleaning.external.reservation.json.response;

import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class RoomJson {
    private Long id;
    private String number;
    private boolean active;
    private String name;
    private String description;
    private int maxCapacity;
    private RoomCategory category;
    private BigDecimal hourlyPrice;
}
