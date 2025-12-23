package usecases

import (
	"context"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/services"
)

// UpdateOrderStatus handles updating order status
type UpdateOrderStatus struct {
	orderService *services.OrderService
}

// NewUpdateOrderStatus creates a new update order status use case
func NewUpdateOrderStatus(orderService *services.OrderService) *UpdateOrderStatus {
	return &UpdateOrderStatus{
		orderService: orderService,
	}
}

// Execute updates the status of an order
func (uc *UpdateOrderStatus) Execute(ctx context.Context, orderID string, status entities.OrderStatus) error {
	return uc.orderService.UpdateOrderStatus(ctx, orderID, status)
}
