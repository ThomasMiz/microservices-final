package io.chotel.reservations.domain.repository;

public interface DatabaseHealthCheck {

    /**
     * Checks that the connection to the database is working and throws an exception if any issue is detected.
     */
    void checkHealth();
}
