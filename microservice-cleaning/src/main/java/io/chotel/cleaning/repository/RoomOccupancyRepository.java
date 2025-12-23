package io.chotel.cleaning.repository;

import io.chotel.cleaning.model.RoomOccupancy;
import io.micronaut.data.annotation.Query;
import io.micronaut.data.annotation.Repository;
import io.micronaut.data.jpa.repository.JpaRepository;

import java.time.OffsetDateTime;
import java.util.List;

@Repository
public interface RoomOccupancyRepository extends JpaRepository<RoomOccupancy, String> {

    @Query("""
            FROM RoomOccupancy r
            WHERE r.state IN ('OCCUPIED_DIRTY', 'EMPTY_DIRTY')
            AND NOT EXISTS (
                FROM CleaningJob c WHERE c.room = r AND c.createdAt >= :latestCleaningTimeAllowed
            )
            """)
    List<RoomOccupancy> getRoomsNeedingCleaning(OffsetDateTime latestCleaningTimeAllowed);
}
