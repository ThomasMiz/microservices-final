package io.chotel.cleaning.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

@Data
@Entity
@Table(name = "room_occupancy")
@NoArgsConstructor
@AllArgsConstructor
public class RoomOccupancy {

    @Id
    @Column(name = "room_number", nullable = false)
    private String roomNumber;

    @Column(name = "state", nullable = false)
    @Enumerated(EnumType.STRING)
    private State state;

    @Column(name = "current_guest_id")
    private String currentGuestId;

    @Column(name = "updated_at", nullable = false)
    @Temporal(TemporalType.TIMESTAMP)
    private OffsetDateTime updatedAt;

    public enum State {
        OCCUPIED_DIRTY,
        OCCUPIED_CLEAN,
        EMPTY_DIRTY,
        EMPTY_CLEAN
    }
}
