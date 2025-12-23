package io.chotel.cleaning.repository;

import io.chotel.cleaning.model.CleaningJob;
import io.micronaut.core.annotation.Nullable;
import io.micronaut.data.annotation.Query;
import io.micronaut.data.annotation.Repository;
import io.micronaut.data.jpa.repository.JpaRepository;
import io.micronaut.data.model.Page;
import io.micronaut.data.model.Pageable;

import java.util.Optional;

@Repository
public interface CleaningJobRepository extends JpaRepository<CleaningJob, Long> {

    @Query(
            value = """
                    FROM CleaningJob c
                    WHERE (:roomNumber IS NULL OR c.room.roomNumber = :roomNumber)
                    AND (:assignedTo IS NULL OR c.assignedTo.id = :assignedTo)
                    ORDER BY c.id DESC
                    """,
            countQuery = """
                    SELECT COUNT(*) FROM CleaningJob c
                    WHERE (:roomNumber IS NULL OR c.room.roomNumber = :roomNumber)
                    AND (:assignedTo IS NULL OR c.assignedTo.id = :assignedTo)
                    """)
    Page<CleaningJob> search(
            @Nullable String roomNumber,
            @Nullable Long assignedTo,
            Pageable pageable
    );

    @Query("FROM CleaningJob c WHERE c.room.roomNumber = :roomNumber ORDER BY c.id DESC")
    Optional<CleaningJob> getLatestByRoomNumber(String roomNumber);
}
