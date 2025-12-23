package io.chotel.cleaning.messaging.json;

import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CheckoutMessageJson {
    private String roomNumber;
    private String guestId;
    private OffsetDateTime timestamp;
}
