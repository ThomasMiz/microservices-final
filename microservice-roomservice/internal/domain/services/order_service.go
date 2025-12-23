package services

import (
	"context"
	"errors"
	"fmt"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderService handles order-related business logic
type OrderService struct {
	orderRepo       repositories.OrderRepository
	menuRepo        repositories.MenuRepository
	reservationRepo repositories.ReservationRepository
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo repositories.OrderRepository,
	menuRepo repositories.MenuRepository,
	reservationRepo repositories.ReservationRepository,
) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		menuRepo:        menuRepo,
		reservationRepo: reservationRepo,
	}
}

// CreateOrderRequest represents the data needed to create an order
type CreateOrderRequest struct {
	RoomID int64
	Items  []OrderItemRequest
}

// OrderItemRequest represents an item in an order request
type OrderItemRequest struct {
	MenuItemID string
	Quantity   int
}

// ValidateOrder validates an order can be created
func (s *OrderService) ValidateOrder(ctx context.Context, req CreateOrderRequest) error {
	if req.RoomID <= 0 {
		return errors.New("room ID is required")
	}
	if len(req.Items) == 0 {
		return errors.New("order must have at least one item")
	}

	// Validate all menu items exist
	for _, item := range req.Items {
		exists, err := s.menuRepo.ExistsByID(ctx, item.MenuItemID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("menu item not found: " + item.MenuItemID)
		}
		if item.Quantity <= 0 {
			return errors.New("quantity must be greater than zero")
		}
	}

	return nil
}

// CalculateTotalPrice calculates the total price for an order
func (s *OrderService) CalculateTotalPrice(ctx context.Context, items []OrderItemRequest) (decimal.Decimal, []entities.OrderItem, error) {
	total := decimal.Zero
	orderItems := make([]entities.OrderItem, 0, len(items))

	for _, item := range items {
		menuItem, err := s.menuRepo.FindByID(ctx, item.MenuItemID)
		if err != nil {
			return decimal.Zero, nil, err
		}
		if menuItem == nil {
			return decimal.Zero, nil, errors.New("menu item not found: " + item.MenuItemID)
		}

		orderItem, err := entities.NewOrderItem(
			uuid.New().String(),
			item.MenuItemID,
			item.Quantity,
			menuItem.Price,
		)
		if err != nil {
			return decimal.Zero, nil, err
		}

		orderItems = append(orderItems, *orderItem)
		total = total.Add(orderItem.GetTotal())
	}

	return total, orderItems, nil
}

// GetReservationID fetches the reservation ID for a room
func (s *OrderService) GetReservationID(ctx context.Context, roomID int64) (string, error) {
	reservation, err := s.reservationRepo.GetActiveReservationForRoom(ctx, roomID)
	if err != nil {
		return "", err
	}
	if reservation == nil {
		return "", nil
	}
	return fmt.Sprintf("%d", reservation.ID), nil
}

// GetActiveReservation fetches the full active reservation for a room, including the billing folder ID.
// This is required for order creation to use the reservation's billing folder.
func (s *OrderService) GetActiveReservation(ctx context.Context, roomID int64) (*repositories.Reservation, error) {
	reservation, err := s.reservationRepo.GetActiveReservationForRoom(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation for room %d: %w", roomID, err)
	}
	if reservation == nil {
		return nil, fmt.Errorf("no active reservation found for room %d", roomID)
	}
	if reservation.BillingFolderID == "" {
		return nil, fmt.Errorf("reservation %d has no billing folder ID", reservation.ID)
	}
	return reservation, nil
}

// CreateOrderResult contains the order and reservation data needed for billing
type CreateOrderResult struct {
	Order       *entities.Order
	Reservation *repositories.Reservation
}

// CreateOrder creates a new order with all validations.
// Returns the order and reservation data needed for billing integration.
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResult, error) {
	// Validate the order
	if err := s.ValidateOrder(ctx, req); err != nil {
		return nil, err
	}

	// Get active reservation - required for billing folder
	reservation, err := s.GetActiveReservation(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	// Calculate total price and create order items
	_, orderItems, err := s.CalculateTotalPrice(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	// Create the order
	order, err := entities.NewOrder(
		uuid.New().String(),
		fmt.Sprintf("%d", req.RoomID),
		fmt.Sprintf("%d", reservation.ID),
		orderItems,
	)
	if err != nil {
		return nil, err
	}

	return &CreateOrderResult{
		Order:       order,
		Reservation: reservation,
	}, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status entities.OrderStatus) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if err := order.UpdateStatus(status); err != nil {
		return err
	}

	return s.orderRepo.Update(ctx, order)
}
