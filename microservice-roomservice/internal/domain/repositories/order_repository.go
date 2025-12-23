package repositories

import (
	"context"
	"microservice-roomservice/internal/domain/entities"
)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	// Save saves an order
	Save(ctx context.Context, order *entities.Order) error

	// FindByID finds an order by ID
	FindByID(ctx context.Context, id string) (*entities.Order, error)

	// FindByRoomID finds all orders for a specific room
	FindByRoomID(ctx context.Context, roomID string) ([]entities.Order, error)

	// FindByReservationID finds all orders for a specific reservation
	FindByReservationID(ctx context.Context, reservationID string) ([]entities.Order, error)

	// Update updates an order
	Update(ctx context.Context, order *entities.Order) error

	// Delete deletes an order by ID
	Delete(ctx context.Context, id string) error

	// FindByStatus finds all orders with a specific status
	FindByStatus(ctx context.Context, status entities.OrderStatus) ([]entities.Order, error)
}
