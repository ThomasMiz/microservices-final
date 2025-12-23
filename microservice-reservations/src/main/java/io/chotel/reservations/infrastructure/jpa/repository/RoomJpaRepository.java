package io.chotel.reservations.infrastructure.jpa.repository;

import io.chotel.reservations.infrastructure.jpa.entity.RoomJpaEntity;
import io.micronaut.data.annotation.Repository;
import io.micronaut.data.jpa.repository.JpaRepository;

@Repository
public interface RoomJpaRepository extends JpaRepository<RoomJpaEntity, Long> {
}
