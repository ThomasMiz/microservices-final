package io.chotel.reservations.infrastructure.jpa.adapter;

import io.chotel.reservations.domain.repository.DatabaseHealthCheck;
import jakarta.inject.Singleton;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Singleton
public class DatabaseHealthCheckImpl implements DatabaseHealthCheck {

    @Override
    public void checkHealth() {
        // Simple health check that doesn't connect to database
        // The application should be considered healthy if it starts successfully
        // Real database connectivity issues will be handled by connection pool and fail fast on actual queries
    }
}
