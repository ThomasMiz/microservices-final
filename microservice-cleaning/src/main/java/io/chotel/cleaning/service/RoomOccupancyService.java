package io.chotel.cleaning.service;

import io.chotel.cleaning.model.CleaningJob;
import io.chotel.cleaning.model.RoomOccupancy;
import io.chotel.cleaning.repository.CleaningJobRepository;
import io.chotel.cleaning.repository.RoomOccupancyRepository;
import io.micronaut.scheduling.annotation.Scheduled;
import io.micronaut.transaction.annotation.Transactional;
import jakarta.annotation.PostConstruct;
import jakarta.inject.Singleton;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.time.Clock;
import java.time.OffsetDateTime;
import java.util.ArrayList;
import java.util.List;

@Slf4j
@Singleton
@RequiredArgsConstructor
public class RoomOccupancyService {
    private final Clock clock;

    private final RoomOccupancyRepository roomOccupancyRepository;
    private final CleaningJobRepository cleaningJobRepository;

    private RoomOccupancy updateRoomStateIfNewer(
            String eventType,
            String roomNumber,
            String guestId,
            OffsetDateTime timestamp,
            RoomOccupancy.State newState
    ) {
        RoomOccupancy ro = roomOccupancyRepository.findById(roomNumber).orElse(null);

        if (ro != null) {
            if (ro.getUpdatedAt() != null && ro.getUpdatedAt().isAfter(timestamp)) {
                log.warn("Ignoring {} event for room {} because the event is older than the room's last update (lastUpdate={}, eventTimestamp={})", eventType, roomNumber, ro.getUpdatedAt(), timestamp);
            }

            ro.setState(newState);
            ro.setCurrentGuestId(guestId);
            ro.setUpdatedAt(timestamp);
            ro = roomOccupancyRepository.update(ro);
            log.info("Updated state of room {} to {} at {}", ro.getRoomNumber(), ro.getState(), ro.getUpdatedAt());
        } else {
            ro = new RoomOccupancy(roomNumber, newState, guestId, timestamp);
            ro = roomOccupancyRepository.save(ro);
            log.info("Created state of room {} as {} at {}", ro.getRoomNumber(), ro.getState(), ro.getUpdatedAt());
        }

        return ro;
    }

    private CleaningJob createCleaningJobForRoom(RoomOccupancy ro) {
        // If the room already has a pending cleaning job created in the past 30 minutes, just use that
        CleaningJob latest = cleaningJobRepository.getLatestByRoomNumber(ro.getRoomNumber()).orElse(null);
        OffsetDateTime maxCleanTime = OffsetDateTime.now(clock).minusMinutes(30);
        if (latest != null && latest.getStartedAt() == null && latest.getCreatedAt().isAfter(maxCleanTime)) {
            log.debug("Not creating a new cleaning job for room {} because there's already a recent pending latest job with ID {}", ro.getRoomNumber(), latest.getId());
            return latest;
        }

        CleaningJob job = new CleaningJob(ro, null, OffsetDateTime.now(clock), null, null);
        job = cleaningJobRepository.save(job);
        log.info("Created cleaning job ID {} for room number {}", job.getId(), ro.getRoomNumber());
        return job;
    }

    @Transactional
    public void onCheckoutEvent(
            String roomNumber,
            String guestId,
            OffsetDateTime timestamp
    ) {
        RoomOccupancy ro = updateRoomStateIfNewer("checkout", roomNumber, guestId, timestamp, RoomOccupancy.State.EMPTY_DIRTY);
        createCleaningJobForRoom(ro);
    }

    @Transactional
    public void onCheckinEvent(
            String roomNumber,
            String guestId,
            OffsetDateTime timestamp
    ) {
        RoomOccupancy ro = updateRoomStateIfNewer("checkout", roomNumber, guestId, timestamp, RoomOccupancy.State.OCCUPIED_DIRTY);
        createCleaningJobForRoom(ro);
    }

    @Scheduled(cron = "0 30 4 1/1 * ?")
    @PostConstruct
    @Transactional
    public void createCleaningJobs() {
        log.info("Creating pending cleaning jobs as part of scheduled cron");
        OffsetDateTime now = OffsetDateTime.now(clock);
        OffsetDateTime latestCleaningTimeAllowed = now.minusHours(3);
        List<RoomOccupancy> rooms = roomOccupancyRepository.getRoomsNeedingCleaning(latestCleaningTimeAllowed);

        if (rooms.isEmpty()) {
            log.info("No jobs to create");
            return;
        }

        List<CleaningJob> jobs = new ArrayList<>();
        for (RoomOccupancy room : rooms) {
            jobs.add(new CleaningJob(room, null, now, null, null));
        }

        jobs = cleaningJobRepository.saveAll(jobs);
        log.info("Created {} cleaning jobs", jobs.size());
    }
}
