package usecases

import (
	"context"
	"fmt"

	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/internal/infrastructure/messaging"
	"microservice-roomservice/pkg/logger"
)

// OrderProducer defines the interface for publishing order messages
type OrderProducer interface {
	PublishOrder(ctx context.Context, message messaging.OrderMessage) error
	PublishOrderEvent(ctx context.Context, event messaging.OrderEvent) error
}

// CreateOrder handles order creation with saga orchestration
type CreateOrder struct {
	orderService   *services.OrderService
	billingService *services.BillingService
	orderRepo      repositories.OrderRepository
	producer       OrderProducer
}

// NewCreateOrder creates a new create order use case
func NewCreateOrder(
	orderService *services.OrderService,
	billingService *services.BillingService,
	orderRepo repositories.OrderRepository,
	producer OrderProducer,
) *CreateOrder {
	return &CreateOrder{
		orderService:   orderService,
		billingService: billingService,
		orderRepo:      orderRepo,
		producer:       producer,
	}
}

// Execute creates a new order using the saga orchestration pattern.
// Flow:
// 1. Validate room and reservation
// 2. Save order locally with PENDING_KITCHEN status
// 3. Publish OrderRequested event to messaging system (Redis)
// 4. Kitchen will respond with OrderAccepted or OrderRejected
// 5. On OrderAccepted, create billing (handled by event handler)
// 6. On billing success, publish OrderConfirmed
// 7. On billing failure, publish OrderBillingFailed
func (uc *CreateOrder) Execute(ctx context.Context, req services.CreateOrderRequest) (*entities.Order, error) {
	// Step 1: Create order using domain service (includes reservation lookup)
	result, err := uc.orderService.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	order := result.Order
	reservation := result.Reservation

	// Step 2: Set billing folder ID from reservation
	order.BillingFolderID = reservation.BillingFolderID

	// Step 3: Set initial status to PENDING_KITCHEN (waiting for stock validation)
	order.Status = entities.OrderStatusPendingKitchen

	// Step 4: Save order to database with PENDING_KITCHEN status
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	logger.CtxInfo(ctx).Msgf("Order %s created with status PENDING_KITCHEN, publishing OrderRequested event", order.ID)

	// Step 5: Publish OrderRequested event to messaging system for kitchen validation
	if uc.producer != nil {
		uc.publishOrderRequestedEvent(ctx, order, reservation.BillingFolderID)
	} else {
		logger.CtxWarn(ctx).Msgf("Producer is nil, cannot publish OrderRequested event for order %s", order.ID)
	}

	return order, nil
}

// publishOrderRequestedEvent publishes an OrderRequested event
func (uc *CreateOrder) publishOrderRequestedEvent(ctx context.Context, order *entities.Order, billingFolderID string) {
	items := make([]messaging.Item, len(order.Items))
	for i, item := range order.Items {
		items[i] = messaging.Item{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      item.Price.String(),
		}
	}

	event := messaging.OrderEvent{
		EventType:       "OrderRequested",
		OrderID:         order.ID,
		RoomID:          order.RoomID,
		ReservationID:   order.ReservationID,
		Items:           items,
		TotalPrice:      order.TotalPrice.String(),
		BillingFolderID: billingFolderID,
		Timestamp:       order.CreatedAt,
	}

	if err := uc.producer.PublishOrderEvent(ctx, event); err != nil {
		logger.CtxError(ctx).Err(err).Msgf("Failed to publish OrderRequested event for order %s", order.ID)
		// Note: Order is already saved, so it will be in PENDING_KITCHEN state
		// A background job should handle retrying or timeout handling
	}
}

// Legacy method for backward compatibility
func (uc *CreateOrder) publishOrderToMessaging(ctx context.Context, order *entities.Order) {
	logger.CtxDebug(ctx).Msgf("publishOrderToMessaging - Starting publish for order %s (legacy)", order.ID)
	items := make([]messaging.Item, len(order.Items))
	for i, item := range order.Items {
		items[i] = messaging.Item{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      item.Price.String(),
		}
	}

	message := messaging.OrderMessage{
		OrderID:       order.ID,
		RoomID:        order.RoomID,
		ReservationID: order.ReservationID,
		Items:         items,
		TotalPrice:    order.TotalPrice.String(),
		Status:        string(order.Status),
		CreatedAt:     order.CreatedAt,
	}

	if err := uc.producer.PublishOrder(ctx, message); err != nil {
		logger.CtxError(ctx).Err(err).Msgf("Failed to publish order %s", order.ID)
	}
}
