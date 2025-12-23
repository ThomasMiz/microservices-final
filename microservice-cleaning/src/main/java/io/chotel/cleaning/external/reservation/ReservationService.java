package io.chotel.cleaning.external.reservation;

import io.chotel.cleaning.external.reservation.config.ReservationApiConfig;
import io.chotel.cleaning.external.reservation.json.response.ReservationJson;
import io.chotel.cleaning.external.reservation.query.SearchReservationsRequestQueryParams;
import io.micronaut.core.type.Argument;
import io.micronaut.http.HttpRequest;
import io.micronaut.http.client.BlockingHttpClient;
import io.micronaut.http.client.HttpClient;
import io.micronaut.http.uri.UriBuilder;
import jakarta.inject.Singleton;

import java.net.URI;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Singleton
public class ReservationService {
    private final BlockingHttpClient httpClient;

    private final String baseUrl;
    private final String reservationsUrl;

    public ReservationService(ReservationApiConfig apiConfig, HttpClient httpClient) {
        this.httpClient = httpClient.toBlocking();

        baseUrl = apiConfig.url();
        reservationsUrl = baseUrl + "/reservations";
    }

    public List<ReservationJson> getReservations(
            SearchReservationsRequestQueryParams query
    ) {
        Map<String, Object> params = new HashMap<>();

        if (query.getFrom() != null) params.put("from", query.getFrom());
        if (query.getTo() != null) params.put("to", query.getTo());
        if (query.getRoomIds() != null && !query.getRoomIds().isEmpty()) params.put("roomIds", query.getRoomIds().stream().map(Object::toString).collect(Collectors.joining(",")));
        if (query.getRoomNumbers() != null && !query.getRoomNumbers().isEmpty()) params.put("roomNumbers", String.join(",", query.getRoomNumbers()));
        if (query.getGuestId() != null) params.put("guestId", query.getGuestId());
        if (query.getPage() != null) params.put("page", query.getPage());
        if (query.getPageSize() != null) params.put("pageSize", query.getPageSize());
        if (query.getSortBy() != null) params.put("sortBy", query.getSortBy());
        if (query.getDirection() != null) params.put("direction", query.getDirection());

        UriBuilder uriBuilder = UriBuilder.of(reservationsUrl);
        params.forEach(uriBuilder::queryParam);
        URI uri = uriBuilder.build();

        HttpRequest<?> request = HttpRequest.GET(uri);

        return httpClient.retrieve(request, Argument.listOf(ReservationJson.class));
    }
}
