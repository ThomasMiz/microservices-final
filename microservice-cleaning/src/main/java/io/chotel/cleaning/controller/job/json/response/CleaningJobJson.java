package io.chotel.cleaning.controller.job.json.response;

import io.chotel.cleaning.controller.staff.json.response.CleaningStaffJson;
import io.chotel.cleaning.model.CleaningJob;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.OffsetDateTime;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class CleaningJobJson {
    private Long id;
    private String roomNumber;
    private CleaningStaffJson assignedTo;
    private OffsetDateTime createdAt;
    private OffsetDateTime startedAt;
    private OffsetDateTime finishedAt;

    public void fillWith(CleaningJob job) {
        this.id = job.getId();
        this.roomNumber = job.getRoom().getRoomNumber();
        this.assignedTo = CleaningStaffJson.from(job.getAssignedTo());
        this.createdAt = job.getCreatedAt();
        this.startedAt = job.getStartedAt();
        this.finishedAt = job.getFinishedAt();
    }

    public static CleaningJobJson from(CleaningJob job) {
        if (job == null) {
            return null;
        }

        CleaningJobJson json = new CleaningJobJson();
        json.fillWith(job);
        return json;
    }
}
