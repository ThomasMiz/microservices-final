package io.chotel.reservations.infrastructure.web.json.response;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.RoomCategory;
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

    public void fillWith(Room room) {
        this.id = room.id();
        this.number = room.number();
        this.active = room.active();
        this.name = room.name();
        this.description = room.description();
        this.maxCapacity = room.maxCapacity();
        this.category = room.category();
        this.hourlyPrice = room.hourlyPrice();
    }

    public static RoomJson from(Room room) {
        if (room == null) {
            return null;
        }

        RoomJson json = new RoomJson();
        json.fillWith(room);
        return json;
    }
}
