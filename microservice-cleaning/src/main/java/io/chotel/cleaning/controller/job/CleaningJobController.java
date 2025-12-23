package io.chotel.cleaning.controller.job;

import io.chotel.cleaning.controller.job.json.request.AssignStaffToJobJson;
import io.chotel.cleaning.controller.job.json.response.CleaningJobJson;
import io.chotel.cleaning.controller.job.query.SearchCleaningJobsQueryParams;
import io.chotel.cleaning.controller.json.response.MessageJson;
import io.chotel.cleaning.model.CleaningJob;
import io.chotel.cleaning.model.CleaningStaff;
import io.chotel.cleaning.repository.CleaningJobRepository;
import io.chotel.cleaning.repository.CleaningStaffRepository;
import io.chotel.cleaning.service.RoomOccupancyService;
import io.micronaut.data.model.Page;
import io.micronaut.data.model.Pageable;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.*;
import io.micronaut.security.annotation.Secured;
import io.micronaut.security.authentication.ServerAuthentication;
import io.micronaut.security.rules.SecurityRule;
import io.micronaut.transaction.annotation.Transactional;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.security.Principal;
import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.List;

@Slf4j
@Secured(SecurityRule.IS_AUTHENTICATED)
@Controller("/jobs")
@RequiredArgsConstructor
public class CleaningJobController {
    private final Clock clock;
    private final CleaningStaffRepository cleaningStaffRepository;
    private final CleaningJobRepository cleaningJobRepository;
    private final RoomOccupancyService roomOccupancyService;

    @Get
    public HttpResponse<List<CleaningJobJson>> searchCleaningJobs(
            @Valid @RequestBean SearchCleaningJobsQueryParams query
    ) {
        Pageable pageable = Pageable.from(query.pageOrStart(), query.pageSizeOr(20));
        Page<CleaningJob> page = cleaningJobRepository.search(query.getRoomNumber(), query.getAssignedTo(), pageable);

        List<CleaningJobJson> body = page.getContent().stream().map(CleaningJobJson::from).toList();
        return HttpResponse.ok(body).header("X-Total-Elements", Long.toString(page.getTotalSize()));
    }

    @Get("/{id:\\d+}")
    public HttpResponse<CleaningJobJson> getCleaningJobById(
            @PathVariable Long id
    ) {
        CleaningJob job = cleaningJobRepository.findById(id).orElse(null);
        if (job == null) {
            return HttpResponse.notFound();
        }

        return HttpResponse.ok(CleaningJobJson.from(job));
    }

    @Post("/{id:\\d+}/assign")
    @Transactional
    public HttpResponse<MessageJson> assignCleaningJob(
            @PathVariable Long id,
            @Valid @Body AssignStaffToJobJson body
    ) {
        CleaningJob job = cleaningJobRepository.findById(id).orElse(null);
        if (job == null) {
            return HttpResponse.notFound();
        }

        if (job.getFinishedAt() != null) {
            return HttpResponse.badRequest(new MessageJson("Cannot re-assign an already finished job"));
        }

        CleaningStaff staff = cleaningStaffRepository.findById(body.getStaffId()).orElse(null);
        if (staff == null) {
            return HttpResponse.badRequest(new MessageJson("Cleaning staff with ID " + body.getStaffId() + " not found"));
        }

        job.setAssignedTo(staff);
        job = cleaningJobRepository.save(job);
        log.info("Assigned job ID {} to staff {} (ID {})", job.getId(), staff.getName(), staff.getId());
        return HttpResponse.ok(new MessageJson("OK"));
    }

    @Post("/{id:\\d+}/started")
    @Transactional
    public HttpResponse<MessageJson> markCleaningJobAsStarted(
            Principal principal,
            @PathVariable Long id
    ) {
        CleaningJob job = cleaningJobRepository.findById(id).orElse(null);
        if (job == null) {
            return HttpResponse.notFound();
        }

        if (job.getStartedAt() != null) {
            return HttpResponse.badRequest(new MessageJson("This job has already been started"));
        }

        if (job.getAssignedTo() == null) {
            Long userId = (Long) ((ServerAuthentication) principal).getAttributes().get("userId");
            job.setAssignedTo(cleaningStaffRepository.findById(userId).orElseThrow());
        }

        job.setStartedAt(OffsetDateTime.now(clock));
        job = cleaningJobRepository.save(job);
        log.info("Marked job ID {} as started", job.getId());
        return HttpResponse.ok(new MessageJson("OK"));
    }

    @Post("/{id:\\d+}/finished")
    @Transactional
    public HttpResponse<MessageJson> markCleaningJobAsFinished(
            Principal principal,
            @PathVariable Long id
    ) {
        CleaningJob job = cleaningJobRepository.findById(id).orElse(null);
        if (job == null) {
            return HttpResponse.notFound();
        }

        if (job.getFinishedAt() != null) {
            return HttpResponse.badRequest(new MessageJson("This job has already been finished"));
        }

        if (job.getStartedAt() == null) {
            return HttpResponse.badRequest(new MessageJson("This job has not been started yet"));
        }

        if (job.getAssignedTo() == null) {
            Long userId = (Long) ((ServerAuthentication) principal).getAttributes().get("userId");
            job.setAssignedTo(cleaningStaffRepository.findById(userId).orElseThrow());
        }

        job.setFinishedAt(OffsetDateTime.now(clock));
        job = cleaningJobRepository.save(job);
        log.info("Marked job ID {} as finished", job.getId());
        return HttpResponse.ok(new MessageJson("OK"));
    }

    @Post("/create-cleaning-jobs")
    public HttpResponse<MessageJson> createCleaningJobs() {
        roomOccupancyService.createCleaningJobs();
        return HttpResponse.ok(new MessageJson("OK"));
    }
}
