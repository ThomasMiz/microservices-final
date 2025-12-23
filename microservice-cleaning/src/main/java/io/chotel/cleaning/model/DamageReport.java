package io.chotel.cleaning.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.OffsetDateTime;

@Data
@Entity
@Table(name = "damage_report")
@NoArgsConstructor
@AllArgsConstructor
public class DamageReport {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    @Column(name = "room_number", nullable = false)
    private String roomNumber;

    @ManyToOne
    @JoinColumn(name = "reported_by_id", nullable = false)
    private CleaningStaff reportedBy;

    @Column(name = "item")
    private String item;

    @Column(name = "description")
    private String description;

    @Column(name = "guest_id", nullable = false)
    private String guestId;

    @Column(name = "fine_amount", precision = 20, scale = 2)
    private BigDecimal fineAmount;

    @Column(name = "billing_folder", nullable = false)
    private String billingFolder;

    @Column(name = "billing_ticket", nullable = false)
    private String billingTicket;

    @Column(name = "created_at", updatable = false)
    @Temporal(TemporalType.TIMESTAMP)
    private OffsetDateTime createdAt;
}
