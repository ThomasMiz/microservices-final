package io.chotel.reservations.infrastructure.web.controller;

import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Get;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

import java.net.URI;

@Singleton
@Controller("/tiki")
@RequiredArgsConstructor
public class TikiController {

    @Get
    public HttpResponse<String> tiki() {
        if (Math.random() > 0.25) {
            return HttpResponse.ok("Taka!");
        }

        return HttpResponse.redirect(URI.create("https://youtu.be/dQw4w9WgXcQ"));
    }
}
