package io.chotel.cleaning.auth;

import io.chotel.cleaning.model.CleaningStaff;
import io.chotel.cleaning.repository.CleaningStaffRepository;
import io.micronaut.core.annotation.NonNull;
import io.micronaut.http.HttpRequest;
import io.micronaut.security.authentication.AuthenticationFailureReason;
import io.micronaut.security.authentication.AuthenticationRequest;
import io.micronaut.security.authentication.AuthenticationResponse;
import io.micronaut.security.authentication.provider.HttpRequestAuthenticationProvider;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;

import java.util.Map;

// https://guides.micronaut.io/latest/micronaut-security-basicauth-maven-java.html

@Singleton
@RequiredArgsConstructor
public class SimpleAfAuthenticationProvider<B> implements HttpRequestAuthenticationProvider<B> {

    // Security? No. All users have the same password :-)
    private static final String ALL_USERS_SHARE_THIS_PASSWORD = "chotardo69";

    private final CleaningStaffRepository cleaningStaffRepository;

    @Override
    public @NonNull AuthenticationResponse authenticate(
            HttpRequest<B> requestContext,
            @NonNull AuthenticationRequest<String, String> authRequest
    ) {
        if (!authRequest.getSecret().equals(ALL_USERS_SHARE_THIS_PASSWORD)) {
            return AuthenticationResponse.failure(AuthenticationFailureReason.CREDENTIALS_DO_NOT_MATCH);
        }

        CleaningStaff staff = cleaningStaffRepository.findByName(authRequest.getIdentity()).orElse(null);
        if (staff == null) {
            // We should really use CREDENTIALS_DO_NOT_MATCH but yeah I prefer this being verbose if it fails xd
            return AuthenticationResponse.failure(AuthenticationFailureReason.USER_NOT_FOUND);
        }

        return AuthenticationResponse.success(staff.getName(), Map.of("userId", staff.getId()));
    }
}
