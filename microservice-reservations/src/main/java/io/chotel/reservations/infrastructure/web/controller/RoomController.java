package io.chotel.reservations.infrastructure.web.controller;

import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.RoomWithAvailability;
import io.chotel.reservations.domain.model.request.CreateRoomRequest;
import io.chotel.reservations.domain.model.request.SearchRoomsRequest;
import io.chotel.reservations.domain.model.result.PageResult;
import io.chotel.reservations.domain.model.result.SearchRoomsResult;
import io.chotel.reservations.domain.usecase.CreateRoomUsecase;
import io.chotel.reservations.domain.usecase.GetRoomByIdUsecase;
import io.chotel.reservations.domain.usecase.SearchRoomsUsecase;
import io.chotel.reservations.infrastructure.web.json.request.CreateRoomRequestJson;
import io.chotel.reservations.infrastructure.web.json.response.RoomJson;
import io.chotel.reservations.infrastructure.web.json.response.RoomWithAvailabilityJson;
import io.chotel.reservations.infrastructure.web.param.SearchRoomsRequestQueryParams;
import io.micronaut.core.annotation.Nullable;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.*;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;

import java.time.OffsetDateTime;
import java.util.List;

@Controller("/rooms")
@RequiredArgsConstructor
public class RoomController {
    private final CreateRoomUsecase createRoomUsecase;
    private final SearchRoomsUsecase searchRoomsUsecase;
    private final GetRoomByIdUsecase getRoomByIdUsecase;

    @Post
    public HttpResponse<RoomJson> createRoom(
            @Valid @Body CreateRoomRequestJson body
    ) {
        CreateRoomRequest request = body.toDomain();
        Room room = createRoomUsecase.execute(request);
        RoomJson json = RoomJson.from(room);
        return HttpResponse.ok(json);
    }

    @Get
    public HttpResponse<List<RoomWithAvailabilityJson>> searchRooms(
            @Valid @RequestBean SearchRoomsRequestQueryParams query
    ) {
        SearchRoomsRequest request = query.toDomain();
        SearchRoomsResult result = searchRoomsUsecase.execute(request);

        PageResult<RoomWithAvailability> page = result.page();
        List<RoomWithAvailabilityJson> body = page.contents().stream().map(RoomWithAvailabilityJson::from).toList();

        return HttpResponse.ok(body).header("X-Total-Elements", Long.toString(page.totalElements()));
    }

    @Get("/{id:\\d+}")
    public HttpResponse<RoomWithAvailabilityJson> getRoomById(
            @PathVariable long id,
            @QueryValue @Nullable OffsetDateTime availabilityFrom,
            @QueryValue @Nullable OffsetDateTime availabilityTo
    ) {
        RoomWithAvailability room = getRoomByIdUsecase.execute(id, availabilityFrom, availabilityTo);
        RoomWithAvailabilityJson json = RoomWithAvailabilityJson.from(room);
        return HttpResponse.ok(json);
    }
}
