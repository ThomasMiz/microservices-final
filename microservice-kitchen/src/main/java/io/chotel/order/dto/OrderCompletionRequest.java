
package io.chotel.order.dto;

import jakarta.validation.constraints.NotNull;

public class OrderCompletionRequest {

    @NotNull
    private OrderStatus status;

    public OrderStatus getStatus() {
        return status;
    }

    public void setStatus(OrderStatus status) {
        this.status = status;
    }
}
