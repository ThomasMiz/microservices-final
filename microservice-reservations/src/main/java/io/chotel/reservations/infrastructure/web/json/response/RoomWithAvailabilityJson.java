package io.chotel.reservations.infrastructure.web.json.response;

import io.chotel.reservations.domain.model.RoomWithAvailability;
import io.micronaut.serde.annotation.Serdeable;
import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;
import java.util.List;

@EqualsAndHashCode(callSuper = true)
@Data
@NoArgsConstructor
@Serdeable
public class RoomWithAvailabilityJson extends RoomJson {
    private List<RangeJson<OffsetDateTime>> availability;

    public void fillWith(RoomWithAvailability roomWithAvailability) {
        super.fillWith(roomWithAvailability.room());
        availability = roomWithAvailability.availability().stream().map(RangeJson::from).toList();
    }

    public static RoomWithAvailabilityJson from(RoomWithAvailability roomWithAvailability) {
        RoomWithAvailabilityJson json = new RoomWithAvailabilityJson();
        json.fillWith(roomWithAvailability);
        return json;
    }
}
