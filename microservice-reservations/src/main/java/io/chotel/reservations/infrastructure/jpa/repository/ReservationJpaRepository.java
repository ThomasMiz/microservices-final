package io.chotel.reservations.infrastructure.jpa.repository;

import io.chotel.reservations.infrastructure.jpa.entity.ReservationJpaEntity;
import io.micronaut.data.annotation.Query;
import io.micronaut.data.annotation.Repository;
import io.micronaut.data.jpa.repository.JpaRepository;

import java.time.OffsetDateTime;
import java.util.Collection;
import java.util.List;

@Repository
public interface ReservationJpaRepository extends JpaRepository<ReservationJpaEntity, Long> {

    @Query("""
            FROM ReservationJpaEntity r
            WHERE r.room.id IN :roomIds AND r.startDate >= :startDate AND r.endDate <= :endDate
            ORDER BY r.room.id, r.startDate
            """)
    List<ReservationJpaEntity> findReservationsForRoomsBetweenDates(
            Collection<Long> roomIds,
            OffsetDateTime startDate,
            OffsetDateTime endDate
    );
}
