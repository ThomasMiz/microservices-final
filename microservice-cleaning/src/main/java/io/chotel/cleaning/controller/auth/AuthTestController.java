package io.chotel.cleaning.controller.auth;

import io.chotel.cleaning.controller.json.response.MessageJson;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Get;
import io.micronaut.security.annotation.Secured;
import io.micronaut.security.authentication.ServerAuthentication;
import io.micronaut.security.rules.SecurityRule;

import java.security.Principal;

@Controller("/auth")
public class AuthTestController {

    @Get("/test")
    @Secured(SecurityRule.IS_AUTHENTICATED)
    public HttpResponse<MessageJson> testy(Principal principal) {
        Long userId = (Long) ((ServerAuthentication) principal).getAttributes().get("userId");
        String message = "Hello, " + principal.getName() + "! You are used ID " + userId + ".";
        return HttpResponse.ok(new MessageJson(message));
    }
}
