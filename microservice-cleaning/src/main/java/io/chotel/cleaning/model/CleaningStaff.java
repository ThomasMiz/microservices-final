package io.chotel.cleaning.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

@Data
@Entity
@Table(name = "cleaning_staff")
@NoArgsConstructor
@AllArgsConstructor
public class CleaningStaff {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    @Column(name = "name", nullable = false)
    private String name;

    @Column(name = "created_at", nullable = false, updatable = false)
    @Temporal(TemporalType.TIMESTAMP)
    private OffsetDateTime createdAt;

    public CleaningStaff(String name, OffsetDateTime createdAt) {
        this.name = name;
        this.createdAt = createdAt;
    }
}
