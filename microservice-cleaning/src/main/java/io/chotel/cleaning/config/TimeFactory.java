package io.chotel.cleaning.config;

import io.micronaut.context.annotation.Bean;
import io.micronaut.context.annotation.Factory;
import jakarta.inject.Singleton;

import java.time.Clock;

@Factory
public class TimeFactory {

    @Bean
    @Singleton
    public Clock clock() {
        return Clock.systemDefaultZone();
    }
}
