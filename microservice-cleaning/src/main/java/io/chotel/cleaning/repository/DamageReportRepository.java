package io.chotel.cleaning.repository;

import io.chotel.cleaning.model.DamageReport;
import io.micronaut.data.annotation.Repository;
import io.micronaut.data.jpa.repository.JpaRepository;

@Repository
public interface DamageReportRepository extends JpaRepository<DamageReport, Long> {
}
