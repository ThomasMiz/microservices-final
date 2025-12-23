package io.chotel.cleaning.controller.damage.json.response;

import io.chotel.cleaning.controller.staff.json.response.CleaningStaffJson;
import io.chotel.cleaning.model.DamageReport;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class DamageReportJson {
    private Long id;
    private String roomNumber;
    private CleaningStaffJson reportedBy;
    private String item;
    private String description;
    private String guestId;
    private BigDecimal fineAmount;
    private String billingFolder;
    private String billingTicket;
    private OffsetDateTime createdAt;

    public void fillWith(DamageReport dr) {
        this.id = dr.getId();
        this.roomNumber = dr.getRoomNumber();
        this.reportedBy = CleaningStaffJson.from(dr.getReportedBy());
        this.item = dr.getItem();
        this.description = dr.getDescription();
        this.guestId = dr.getGuestId();
        this.fineAmount = dr.getFineAmount();
        this.billingFolder = dr.getBillingFolder();
        this.billingTicket = dr.getBillingTicket();
        this.createdAt = dr.getCreatedAt();
    }

    public static DamageReportJson from(DamageReport dr) {
        DamageReportJson json = new DamageReportJson();
        json.fillWith(dr);
        return json;
    }
}
