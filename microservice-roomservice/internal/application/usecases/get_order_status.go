package usecases

import (
	"context"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
)

// GetOrderStatus retrieves an order by ID
type GetOrderStatus struct {
	orderRepo repositories.OrderRepository
}

// NewGetOrderStatus creates a new get order status use case
func NewGetOrderStatus(orderRepo repositories.OrderRepository) *GetOrderStatus {
	return &GetOrderStatus{
		orderRepo: orderRepo,
	}
}

// ByID retrieves an order by ID
func (uc *GetOrderStatus) ByID(ctx context.Context, id string) (*entities.Order, error) {
	return uc.orderRepo.FindByID(ctx, id)
}

// ByRoomID retrieves all orders for a room
func (uc *GetOrderStatus) ByRoomID(ctx context.Context, roomID string) ([]entities.Order, error) {
	return uc.orderRepo.FindByRoomID(ctx, roomID)
}

// ByReservationID retrieves all orders for a reservation
func (uc *GetOrderStatus) ByReservationID(ctx context.Context, reservationID string) ([]entities.Order, error) {
	return uc.orderRepo.FindByReservationID(ctx, reservationID)
}
