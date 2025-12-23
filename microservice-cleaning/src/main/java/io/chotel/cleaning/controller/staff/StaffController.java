package io.chotel.cleaning.controller.staff;

import io.chotel.cleaning.controller.staff.json.request.CreateCleaningStaffJson;
import io.chotel.cleaning.controller.staff.json.response.CleaningStaffJson;
import io.chotel.cleaning.controller.staff.query.SearchCleaningStaffQueryParams;
import io.chotel.cleaning.model.CleaningStaff;
import io.chotel.cleaning.repository.CleaningStaffRepository;
import io.chotel.cleaning.util.DatabaseUtils;
import io.micronaut.data.model.Page;
import io.micronaut.data.model.Pageable;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.*;
import io.micronaut.security.annotation.Secured;
import io.micronaut.security.rules.SecurityRule;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;

import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.List;

@Secured(SecurityRule.IS_AUTHENTICATED)
@Controller("/staff")
@RequiredArgsConstructor
public class StaffController {
    private final Clock clock;
    private final CleaningStaffRepository cleaningStaffRepository;

    @Get
    public HttpResponse<List<CleaningStaffJson>> searchCleaningStaff(
            @Valid @RequestBean SearchCleaningStaffQueryParams query
    ) {
        String name = DatabaseUtils.makeSearchParam(query.getName());
        Page<CleaningStaff> page = cleaningStaffRepository.search(name, Pageable.from(query.pageOrStart(), query.pageSizeOr(20)));

        List<CleaningStaffJson> body = page.getContent().stream().map(CleaningStaffJson::from).toList();
        return HttpResponse.ok(body).header("X-Total-Elements", Long.toString(page.getTotalSize()));
    }

    @Post
    public HttpResponse<CleaningStaffJson> createCleaningStaff(
            @Valid @Body CreateCleaningStaffJson body
    ) {
        CleaningStaff staff = cleaningStaffRepository.save(new CleaningStaff(body.getName(), OffsetDateTime.now(clock)));
        return HttpResponse.ok(CleaningStaffJson.from(staff));
    }

    @Get("/{id:\\d+}")
    public HttpResponse<CleaningStaffJson> getCleaningStaffById(
            @PathVariable long id
    ) {
        CleaningStaff staff = cleaningStaffRepository.findById(id).orElse(null);
        if (staff == null) {
            return HttpResponse.notFound();
        }

        return HttpResponse.ok(CleaningStaffJson.from(staff));
    }
}
