package io.chotel.reservations.infrastructure.web.controller;

import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.model.request.SearchReservationsRequest;
import io.chotel.reservations.domain.model.result.CreateReservationResult;
import io.chotel.reservations.domain.model.result.PageResult;
import io.chotel.reservations.domain.model.result.SearchReservationsResult;
import io.chotel.reservations.domain.usecase.CreateReservationUsecase;
import io.chotel.reservations.domain.usecase.GetReservationByIdUsecase;
import io.chotel.reservations.domain.usecase.SearchReservationsUsecase;
import io.chotel.reservations.infrastructure.web.json.request.CreateReservationRequestJson;
import io.chotel.reservations.infrastructure.web.json.response.ReservationJson;
import io.chotel.reservations.infrastructure.web.param.SearchReservationsRequestQueryParams;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.*;
import io.micronaut.scheduling.TaskExecutors;
import io.micronaut.scheduling.annotation.ExecuteOn;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;

import java.util.List;

@Controller("/reservations")
@RequiredArgsConstructor
public class ReservationController {
    private final CreateReservationUsecase createReservationUsecase;
    private final SearchReservationsUsecase searchReservationsUsecase;
    private final GetReservationByIdUsecase getReservationByIdUsecase;

    @Post
    @ExecuteOn(TaskExecutors.BLOCKING)
    public HttpResponse<ReservationJson> createReservation(
            @Valid @Body CreateReservationRequestJson body
    ) {
        CreateReservationResult result = createReservationUsecase.execute(body.toDomain());
        Reservation reservation = result.reservation();
        ReservationJson json = ReservationJson.from(reservation);
        return HttpResponse.ok(json);
    }

    @Get
    public HttpResponse<List<ReservationJson>> searchReservations(
            @Valid @RequestBean SearchReservationsRequestQueryParams query
    ) {
        SearchReservationsRequest request = query.toDomain();
        SearchReservationsResult result = searchReservationsUsecase.execute(request);

        PageResult<Reservation> page = result.page();
        List<ReservationJson> body = page.contents().stream().map(ReservationJson::from).toList();

        return HttpResponse.ok(body).header("X-Total-Elements", Long.toString(page.totalElements()));
    }

    @Get("/{id:\\d+}")
    public HttpResponse<ReservationJson> getReservationById(
            @PathVariable long id
    ) {
        Reservation reservation = getReservationByIdUsecase.execute(id);
        ReservationJson json = ReservationJson.from(reservation);
        return HttpResponse.ok(json);
    }
}
