package io.chotel.reservations.domain.usecase;

import io.chotel.reservations.domain.model.request.SearchRoomsRequest;
import io.chotel.reservations.domain.model.result.SearchRoomsResult;

public interface SearchRoomsUsecase {

    SearchRoomsResult execute(SearchRoomsRequest request);
}
