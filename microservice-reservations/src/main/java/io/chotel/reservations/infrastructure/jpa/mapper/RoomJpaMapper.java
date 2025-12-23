package io.chotel.reservations.infrastructure.jpa.mapper;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.infrastructure.jpa.entity.RoomJpaEntity;
import lombok.experimental.UtilityClass;

@UtilityClass
public final class RoomJpaMapper {

    public static Room toDomain(RoomJpaEntity entity) {
        if (entity == null) {
            return null;
        }

        return new Room(
                entity.getId(),
                entity.getNumber(),
                entity.isActive(),
                entity.getName(),
                entity.getDescription(),
                entity.getMaxCapacity(),
                entity.getCategory(),
                entity.getHourlyPrice()
        );
    }

    public static RoomJpaEntity toEntity(Room room) {
        if (room == null) {
            return null;
        }

        return new RoomJpaEntity(
                room.id(),
                room.number(),
                room.active(),
                room.name(),
                room.description(),
                room.maxCapacity(),
                room.category(),
                room.hourlyPrice()
        );
    }
}
