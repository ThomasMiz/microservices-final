package io.chotel.cleaning.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

/**
 * Represents a cleaning job for a room that may be pending, ongoing, or finished. This state is implicit depending on
 * whether the {@link #startedAt} and {@link #finishedAt} fields are present.
 */
@Data
@Entity
@Table(name = "cleaning_job")
@NoArgsConstructor
@AllArgsConstructor
public class CleaningJob {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    @ManyToOne(optional = false)
    @JoinColumn(name = "room_number", nullable = false, updatable = false)
    private RoomOccupancy room;

    @ManyToOne
    @JoinColumn(name = "assigned_to_id")
    private CleaningStaff assignedTo;

    @Column(name = "created_at")
    private OffsetDateTime createdAt;

    @Column(name = "started_at")
    private OffsetDateTime startedAt;

    @Column(name = "finished_at")
    private OffsetDateTime finishedAt;

    /**
     * Creates a new {@link CleaningJob} with null ID for insertion into the database.
     */
    public CleaningJob(RoomOccupancy room, CleaningStaff assignedTo, OffsetDateTime createdAt, OffsetDateTime startedAt, OffsetDateTime finishedAt) {
        this.room = room;
        this.assignedTo = assignedTo;
        this.createdAt = createdAt;
        this.startedAt = startedAt;
        this.finishedAt = finishedAt;
    }
}
