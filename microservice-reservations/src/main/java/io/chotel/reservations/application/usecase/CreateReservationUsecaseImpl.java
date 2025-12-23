package io.chotel.reservations.application.usecase;

import io.chotel.reservations.domain.exception.RoomNotActiveException;
import io.chotel.reservations.domain.exception.RoomNotAvailableException;
import io.chotel.reservations.domain.exception.RoomNotFoundException;
import io.chotel.reservations.domain.external.billing.BillingService;
import io.chotel.reservations.domain.external.billing.model.BillingFolder;
import io.chotel.reservations.domain.external.billing.model.BillingTicket;
import io.chotel.reservations.domain.external.billing.model.CreateTicketRequest;
import io.chotel.reservations.domain.model.Reservation;
import io.chotel.reservations.domain.model.Room;
import io.chotel.reservations.domain.model.request.CreateReservationRequest;
import io.chotel.reservations.domain.model.result.CreateReservationResult;
import io.chotel.reservations.domain.repository.ReservationRepository;
import io.chotel.reservations.domain.repository.RoomRepository;
import io.chotel.reservations.domain.usecase.CreateReservationUsecase;
import io.micronaut.transaction.SynchronousTransactionManager;
import jakarta.inject.Singleton;
import jakarta.persistence.EntityManager;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.math.BigDecimal;
import java.time.Duration;
import java.util.List;
import java.util.concurrent.CompletableFuture;

@Slf4j
@Singleton
@RequiredArgsConstructor
public class CreateReservationUsecaseImpl implements CreateReservationUsecase {
    private static final int SECONDS_PER_HOUR = 3600;

    private final SynchronousTransactionManager<EntityManager> transactionManager;

    private final RoomRepository roomRepository;
    private final ReservationRepository reservationRepository;

    private final BillingService billingService;

    @Override
    public CreateReservationResult execute(CreateReservationRequest request) {
        Room room = roomRepository.findById(request.roomId()).orElseThrow(RoomNotFoundException::new);

        if (!room.active()) {
            throw new RoomNotActiveException();
        }

        // Calculate the duration of the stay in hours, rounding the minutes up
        Duration stayDuration = Duration.between(request.startDate(), request.endDate());
        long stayDurationHours = (stayDuration.toSeconds() + SECONDS_PER_HOUR - 1) / SECONDS_PER_HOUR;
        BigDecimal totalPrice = room.hourlyPrice().multiply(BigDecimal.valueOf(stayDurationHours));

        Reservation reservation = transactionManager.executeWrite(status -> {
            List<Reservation> intersectingReservations = reservationRepository.findReservationsForRoomsBetweenDates(
                    List.of(room.id()),
                    request.startDate(),
                    request.endDate()
            ).get(room.id());

            if (intersectingReservations != null && !intersectingReservations.isEmpty()) {
                throw new RoomNotAvailableException(intersectingReservations);
            }

            return reservationRepository.createReservation(
                    room.id(),
                    request.guestId(),
                    request.startDate(),
                    request.endDate(),
                    room.hourlyPrice(),
                    totalPrice
            );
        });

        BillingFolder folder = null;
        BillingTicket ticket = null;
        try {
            folder = billingService.createBillingFolder(request.guestId());

            CreateTicketRequest createTicketRequest = CreateTicketRequest.builder()
                    .folderId(folder.id())
                    .total(totalPrice)
                    .itemId("ROOM" + room.id())
                    .description(String.format("The charge for using room %s for %d hours", room.number(), stayDurationHours))
                    .build();

            ticket = billingService.createTicket(createTicketRequest);
            reservation = reservationRepository.setBillingInfo(reservation, folder.id(), ticket.id());
            return new CreateReservationResult(reservation);
        } catch (Exception e) {
            log.info("Exception occurred while creating reservation, undoing effects", e);

            // Remove the already-inserted row from the database and the folder or ticket from the billing service
            deleteReservationAsync(reservation);
            deleteFolderAndTicketAsync(folder, ticket);

            throw e;
        }
    }

    private void deleteReservationAsync(Reservation reservation) {
        final Long reservationId = reservation.id();
        CompletableFuture.runAsync(() -> {
            reservationRepository.deleteReservationById(reservationId);
        });
    }

    private void deleteFolderAndTicketAsync(BillingFolder folder, BillingTicket ticket) {
        if (folder != null || ticket != null) {
            final String ticketId = ticket == null ? null : ticket.id();
            final String folderId = folder == null ? null : folder.id();
            CompletableFuture.runAsync(() -> {
                if (ticketId != null) {
                    billingService.deleteTicket(ticketId);
                }

                if (folderId != null) {
                    billingService.deleteBillingFolder(folderId);
                }
            });
        }
    }
}
