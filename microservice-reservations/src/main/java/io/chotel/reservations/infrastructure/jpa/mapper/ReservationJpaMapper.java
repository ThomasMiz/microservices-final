package io.chotel.reservations.infrastructure.jpa.mapper;

import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.infrastructure.jpa.entity.ReservationJpaEntity;
import lombok.experimental.UtilityClass;

@UtilityClass
public final class ReservationJpaMapper {

    public static Reservation toDomain(ReservationJpaEntity entity) {
        if (entity == null) {
            return null;
        }

        return new Reservation(
                entity.getId(),
                RoomJpaMapper.toDomain(entity.getRoom()),
                entity.getGuestId(),
                entity.getStartDate(),
                entity.getEndDate(),
                entity.getBillingFolderId(),
                entity.getReservationBillingTicket(),
                entity.getRentedHourlyPrice(),
                entity.getTotalPrice()
        );
    }

    public static ReservationJpaEntity toEntity(Reservation reservation) {
        return new ReservationJpaEntity(
                reservation.id(),
                RoomJpaMapper.toEntity(reservation.room()),
                reservation.guestId(),
                reservation.startDate(),
                reservation.endDate(),
                reservation.billingFolderId(),
                reservation.reservationBillingTicket(),
                reservation.rentedHourlyPrice(),
                reservation.totalPrice()
        );
    }
}
