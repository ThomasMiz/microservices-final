package io.chotel.reservations.infrastructure.web.json.response;

import io.chotel.reservations.domain.model.Range;
import io.micronaut.serde.annotation.Serdeable;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Serdeable
public class RangeJson<T> {
    private T from;
    private T to;

    public void fillWith(Range<T> range) {
        this.from = range.from();
        this.to = range.to();
    }

    public static <T> RangeJson<T> from(Range<T> range) {
        if (range == null) {
            return null;
        }

        RangeJson<T> json = new RangeJson<>();
        json.fillWith(range);
        return json;
    }
}
