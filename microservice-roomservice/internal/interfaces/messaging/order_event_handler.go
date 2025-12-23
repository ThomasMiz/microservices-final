package messaging

import (
	"context"
	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/internal/infrastructure/messaging"
	"microservice-roomservice/pkg/logger"
)

// OrderEventProducer defines the interface for publishing order events
type OrderEventProducer interface {
	PublishOrderEvent(ctx context.Context, event messaging.OrderEvent) error
}

// OrderEventHandler handles order-related events for saga orchestration
type OrderEventHandler struct {
	updateOrderStatus *usecases.UpdateOrderStatus
	billingService    *services.BillingService
	orderRepo         repositories.OrderRepository
	producer          OrderEventProducer
}

// NewOrderEventHandler creates a new order event handler
func NewOrderEventHandler(
	updateOrderStatus *usecases.UpdateOrderStatus,
	billingService *services.BillingService,
	orderRepo repositories.OrderRepository,
	producer OrderEventProducer,
) *OrderEventHandler {
	return &OrderEventHandler{
		updateOrderStatus: updateOrderStatus,
		billingService:    billingService,
		orderRepo:         orderRepo,
		producer:          producer,
	}
}

// HandleOrderEvent handles orchestration events from the kitchen service
func (h *OrderEventHandler) HandleOrderEvent(ctx context.Context, event messaging.OrderEvent) error {
	logger.CtxInfo(ctx).Msgf("Received order event: type=%s, orderId=%s", event.EventType, event.OrderID)

	switch event.EventType {
	case "OrderAccepted":
		return h.handleOrderAccepted(ctx, event)
	case "OrderRejected":
		return h.handleOrderRejected(ctx, event)
	case "OrderPreparationCompleted":
		return h.handleOrderPreparationCompleted(ctx, event)
	case "OrderPreparationFailed":
		return h.handleOrderPreparationFailed(ctx, event)
	default:
		logger.CtxWarn(ctx).Msgf("Unknown event type: %s", event.EventType)
		return nil
	}
}

// handleOrderAccepted processes OrderAccepted event from kitchen
// Kitchen has reserved stock, now we need to create billing
func (h *OrderEventHandler) handleOrderAccepted(ctx context.Context, event messaging.OrderEvent) error {
	logger.CtxInfo(ctx).Msgf("Processing OrderAccepted for order %s", event.OrderID)

	// Get the order
	order, err := h.orderRepo.FindByID(ctx, event.OrderID)
	if err != nil {
		logger.CtxError(ctx).Err(err).Msgf("Failed to find order %s", event.OrderID)
		return err
	}

	if order == nil {
		logger.CtxWarn(ctx).Msgf("Order %s not found", event.OrderID)
		return nil
	}

	// Update status to PENDING_BILLING
	if err := order.UpdateStatus(entities.OrderStatusPendingBilling); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Failed to update order status")
		return err
	}

	if err := h.orderRepo.Update(ctx, order); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Failed to save order status update")
		return err
	}

	// Create billing ticket
	billingFolderID := event.BillingFolderID
	if billingFolderID == "" {
		// Try to get from order if not in event
		billingFolderID = order.BillingFolderID
	}

	if err := h.billingService.CreateBillingForOrder(ctx, order, billingFolderID); err != nil {
		logger.CtxError(ctx).Err(err).Msgf("Billing failed for order %s", event.OrderID)

		// Publish OrderBillingFailed event for compensation
		failedEvent := messaging.OrderEvent{
			EventType:     "OrderBillingFailed",
			OrderID:       order.ID,
			RoomID:        order.RoomID,
			ReservationID: order.ReservationID,
			Reason:        err.Error(),
			Timestamp:     order.CreatedAt,
		}

		if h.producer != nil {
			if pubErr := h.producer.PublishOrderEvent(ctx, failedEvent); pubErr != nil {
				logger.CtxError(ctx).Err(pubErr).Msg("Failed to publish OrderBillingFailed event")
			}
		}

		// Update order to cancelled
		order.Status = entities.OrderStatusCancelled
		h.orderRepo.Update(ctx, order)

		return nil // Don't retry, compensation event sent
	}

	// Billing succeeded - publish OrderConfirmed
	logger.CtxInfo(ctx).Msgf("Billing created for order %s, publishing OrderConfirmed", event.OrderID)

	// Update order status to PREPARING
	order.Status = entities.OrderStatusPreparing
	if err := h.orderRepo.Update(ctx, order); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Failed to update order to preparing")
	}

	// Publish OrderConfirmed event
	confirmedEvent := messaging.OrderEvent{
		EventType:     "OrderConfirmed",
		OrderID:       order.ID,
		RoomID:        order.RoomID,
		ReservationID: order.ReservationID,
		Timestamp:     order.CreatedAt,
	}

	if h.producer != nil {
		if err := h.producer.PublishOrderEvent(ctx, confirmedEvent); err != nil {
			logger.CtxError(ctx).Err(err).Msg("Failed to publish OrderConfirmed event")
		}
	}

	logger.CtxInfo(ctx).Msgf("Order %s confirmed and in preparation", event.OrderID)
	return nil
}

// handleOrderRejected processes OrderRejected event from kitchen
// Kitchen rejected due to insufficient stock
func (h *OrderEventHandler) handleOrderRejected(ctx context.Context, event messaging.OrderEvent) error {
	logger.CtxInfo(ctx).Msgf("Processing OrderRejected for order %s: %s", event.OrderID, event.Reason)

	// Get and update order status to cancelled
	order, err := h.orderRepo.FindByID(ctx, event.OrderID)
	if err != nil || order == nil {
		logger.CtxWarn(ctx).Msgf("Order %s not found for rejection", event.OrderID)
		return nil
	}

	order.Status = entities.OrderStatusCancelled
	if err := h.orderRepo.Update(ctx, order); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Failed to update order status to cancelled")
		return err
	}

	logger.CtxInfo(ctx).Msgf("Order %s cancelled due to kitchen rejection: %s", event.OrderID, event.Reason)
	return nil
}

// handleOrderPreparationCompleted processes preparation completion
func (h *OrderEventHandler) handleOrderPreparationCompleted(ctx context.Context, event messaging.OrderEvent) error {
	logger.CtxInfo(ctx).Msgf("Processing OrderPreparationCompleted for order %s", event.OrderID)

	order, err := h.orderRepo.FindByID(ctx, event.OrderID)
	if err != nil || order == nil {
		logger.CtxWarn(ctx).Msgf("Order %s not found for completion", event.OrderID)
		return nil
	}

	order.Status = entities.OrderStatusCompleted
	if err := h.orderRepo.Update(ctx, order); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Failed to update order status to completed")
		return err
	}

	logger.CtxInfo(ctx).Msgf("Order %s completed", event.OrderID)
	return nil
}

// handleOrderPreparationFailed processes critical kitchen failure after confirmation
// This requires billing cancellation for financial consistency
func (h *OrderEventHandler) handleOrderPreparationFailed(ctx context.Context, event messaging.OrderEvent) error {
	logger.CtxInfo(ctx).Msgf("Processing OrderPreparationFailed for order %s: %s", event.OrderID, event.Reason)

	order, err := h.orderRepo.FindByID(ctx, event.OrderID)
	if err != nil || order == nil {
		logger.CtxWarn(ctx).Msgf("Order %s not found for failure handling", event.OrderID)
		return nil
	}

	// Cancel billing ticket
	if order.HasBillingInfo() {
		logger.CtxInfo(ctx).Msgf("Cancelling billing for failed order %s", event.OrderID)
		h.billingService.DeleteBillingForOrder(ctx, order)
	}

	// Update order status
	order.Status = entities.OrderStatusCancelled
	if err := h.orderRepo.Update(ctx, order); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Failed to update order status to cancelled")
		return err
	}

	logger.CtxInfo(ctx).Msgf("Order %s cancelled due to preparation failure, billing reversed", event.OrderID)
	return nil
}

// HandleOrderCompletion handles legacy order completion messages
func (h *OrderEventHandler) HandleOrderCompletion(ctx context.Context, message messaging.OrderCompletionMessage) error {
	logger.CtxInfo(ctx).Msgf("Received order completion message (legacy): OrderID=%s, Status=%s", message.OrderID, message.Status)

	// Convert string status to OrderStatus
	status := entities.OrderStatus(message.Status)
	if !status.IsValid() {
		logger.CtxWarn(ctx).Msgf("Invalid order status: %s", message.Status)
		return nil // Don't retry invalid messages
	}

	// If order is cancelled, cancel the billing ticket
	if status == entities.OrderStatusCancelled {
		h.cancelBillingForOrder(ctx, message.OrderID)
	}

	// Update order status
	if err := h.updateOrderStatus.Execute(ctx, message.OrderID, status); err != nil {
		logger.CtxError(ctx).Err(err).Msg("Error updating order status")
		return err
	}

	logger.CtxInfo(ctx).Msgf("Successfully updated order %s to status %s", message.OrderID, message.Status)
	return nil
}

// cancelBillingForOrder cancels billing for a given order ID
func (h *OrderEventHandler) cancelBillingForOrder(ctx context.Context, orderID string) {
	order, err := h.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		logger.CtxError(ctx).Err(err).Msgf("Failed to retrieve order %s for billing cancellation", orderID)
		return
	}

	if order == nil {
		logger.CtxWarn(ctx).Msgf("Order %s not found for billing cancellation", orderID)
		return
	}

	if order.HasBillingInfo() {
		logger.CtxInfo(ctx).Msgf("Cancelling billing for order %s (folder: %s, ticket: %s)", orderID, order.BillingFolderID, order.BillingTicketID)
		h.billingService.DeleteBillingForOrder(ctx, order)
	} else {
		logger.CtxInfo(ctx).Msgf("Order %s has no billing info to cancel", orderID)
	}
}
