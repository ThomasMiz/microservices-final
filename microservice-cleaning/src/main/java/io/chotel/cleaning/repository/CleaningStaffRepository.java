package io.chotel.cleaning.repository;

import io.chotel.cleaning.model.CleaningStaff;
import io.micronaut.core.annotation.Nullable;
import io.micronaut.data.annotation.Query;
import io.micronaut.data.annotation.Repository;
import io.micronaut.data.jpa.repository.JpaRepository;
import io.micronaut.data.model.Page;
import io.micronaut.data.model.Pageable;

import java.util.Optional;

@Repository
public interface CleaningStaffRepository extends JpaRepository<CleaningStaff, Long> {

    Optional<CleaningStaff> findByName(String name);

    @Query(
            value = """
                    FROM CleaningStaff c
                    WHERE (:name IS NULL OR c.name ILIKE :name)
                    ORDER BY c.id ASC
                    """,
            countQuery = """
                    SELECT COUNT(c)
                    FROM CleaningStaff c
                    WHERE (:name IS NULL OR c.name ILIKE :name)
                    """
    )
    Page<CleaningStaff> search(
            @Nullable String name,
            Pageable pageable
    );
}
