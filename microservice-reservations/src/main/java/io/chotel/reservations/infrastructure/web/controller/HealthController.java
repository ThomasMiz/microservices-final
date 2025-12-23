package io.chotel.reservations.infrastructure.web.controller;

import io.chotel.reservations.domain.repository.DatabaseHealthCheck;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Get;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

@Singleton
@Controller("/health")
@RequiredArgsConstructor
public class HealthController {
    private final DatabaseHealthCheck databaseHealthCheck;

    @Get("/live")
    public HttpResponse<Void> liveness() {
        try {
            databaseHealthCheck.checkHealth();
            return HttpResponse.noContent();
        } catch (Exception ignored) {
            return HttpResponse.serverError();
        }
    }

    @Get("/ready")
    public HttpResponse<Void> readiness() {
        return liveness();
    }
}
