package io.chotel.cleaning.controller.staff.json.response;

import io.chotel.cleaning.model.CleaningStaff;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CleaningStaffJson {
    private Long id;
    private String name;
    private OffsetDateTime createdAt;

    public void fillWith(CleaningStaff cleaningStaff) {
        this.id = cleaningStaff.getId();
        this.name = cleaningStaff.getName();
        this.createdAt = cleaningStaff.getCreatedAt();
    }

    public static CleaningStaffJson from(CleaningStaff cleaningStaff) {
        if (cleaningStaff == null) {
            return null;
        }

        CleaningStaffJson json = new CleaningStaffJson();
        json.fillWith(cleaningStaff);
        return json;
    }
}
