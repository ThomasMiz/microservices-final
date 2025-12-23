package io.chotel.cleaning.controller.damage;

import io.chotel.cleaning.controller.damage.json.request.CreateDamageReportJson;
import io.chotel.cleaning.controller.damage.json.response.DamageReportJson;
import io.chotel.cleaning.controller.json.response.MessageJson;
import io.chotel.cleaning.external.billing.BillingService;
import io.chotel.cleaning.external.billing.model.BillingTicket;
import io.chotel.cleaning.external.billing.model.CreateTicketRequest;
import io.chotel.cleaning.external.reservation.ReservationService;
import io.chotel.cleaning.external.reservation.json.response.ReservationJson;
import io.chotel.cleaning.external.reservation.query.SearchReservationsRequestQueryParams;
import io.chotel.cleaning.model.CleaningStaff;
import io.chotel.cleaning.model.DamageReport;
import io.chotel.cleaning.repository.CleaningStaffRepository;
import io.chotel.cleaning.repository.DamageReportRepository;
import io.micronaut.data.model.Sort;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.*;
import io.micronaut.scheduling.TaskExecutors;
import io.micronaut.scheduling.annotation.ExecuteOn;
import io.micronaut.security.annotation.Secured;
import io.micronaut.security.authentication.ServerAuthentication;
import io.micronaut.security.rules.SecurityRule;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.security.Principal;
import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Set;

@Slf4j
@Secured(SecurityRule.IS_AUTHENTICATED)
@Controller("/damage-reports")
@RequiredArgsConstructor
public class DamageReportController {
    private final Clock clock;
    private final ReservationService reservationService;
    private final BillingService billingService;
    private final DamageReportRepository damageReportRepository;
    private final CleaningStaffRepository cleaningStaffRepository;

    @Get
    public HttpResponse<List<DamageReportJson>> getAll() {
        List<DamageReportJson> body = damageReportRepository.findAll().stream().map(DamageReportJson::from).toList();
        return HttpResponse.ok(body);
    }

    @Get("/{id:\\d+}")
    public HttpResponse<DamageReportJson> getById(@PathVariable Long id) {
        DamageReport dr = damageReportRepository.findById(id).orElse(null);
        if (dr == null) {
            return HttpResponse.notFound();
        }

        return HttpResponse.ok(DamageReportJson.from(dr));
    }

    @Post
    @ExecuteOn(TaskExecutors.BLOCKING)
    public HttpResponse<?> create(
            Principal principal,
            @Valid @Body CreateDamageReportJson body
    ) {
        Long userId = (Long) ((ServerAuthentication) principal).getAttributes().get("userId");
        CleaningStaff reportedBy = cleaningStaffRepository.findById(userId).orElseThrow();

        List<ReservationJson> reservations = reservationService.getReservations(
                SearchReservationsRequestQueryParams.builder()
                        .roomNumbers(Set.of(body.getRoomNumber()))
                        .page(0)
                        .pageSize(1)
                        .sortBy(SearchReservationsRequestQueryParams.SortBy.ID)
                        .direction(Sort.Order.Direction.DESC)
                        .build()
        );

        if (reservations.isEmpty()) {
            log.error("Cannot create damage report: no reservations found for room {}", body.getRoomNumber());
            return HttpResponse.serverError(new MessageJson("Server error: could not find any reservations for room " + body.getRoomNumber()));
        }

        ReservationJson reservation = reservations.getFirst();
        String guestId = reservation.getGuestId();
        String billingFolderId = reservation.getBillingFolderId();

        BillingTicket ticket;
        try {
            ticket = billingService.createTicket(new CreateTicketRequest(
                    billingFolderId,
                    body.getFineAmount(),
                    "broken-" + body.getBrokenItem(),
                    "Fine for breaking the item: " + body.getBrokenItem()
            ));
        } catch (Exception e) {
            log.error("Cannot create damage report: could not create billing ticket", e);
            return HttpResponse.serverError(new MessageJson("Server error: could not create ticket with billing service"));
        }

        String billingTicketId = ticket.id();

        OffsetDateTime now = OffsetDateTime.now(clock);
        DamageReport dr = new DamageReport(null, body.getRoomNumber(), reportedBy, body.getBrokenItem(), body.getDescription(), guestId, body.getFineAmount(), billingFolderId, billingTicketId, now);
        dr = damageReportRepository.save(dr);

        return HttpResponse.created(DamageReportJson.from(dr));
    }
}
