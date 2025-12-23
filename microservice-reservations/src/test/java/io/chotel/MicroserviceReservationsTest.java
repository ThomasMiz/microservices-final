package io.chotel;

import io.micronaut.runtime.EmbeddedApplication;
import io.micronaut.test.extensions.junit5.annotation.MicronautTest;
import jakarta.inject.Inject;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import javax.sql.DataSource;

// https://guides.micronaut.io/latest/replace-h2-with-real-database-for-testing-gradle-java.html

@MicronautTest
class MicroserviceReservationsTest {

    @Inject
    EmbeddedApplication<?> application;

    @Inject
    DataSource dataSource;

    @Test
    void testDataSourceIsPresent() {
        assertTrue(dataSource != null);
    }

    @Test
    void testItWorks() {
        assertTrue(application.isRunning());
    }

    @Test
    void testOnePlusOneEqualsTwo() {
        int result = 1 + 1;
        assertEquals(2, result);
    }
}
