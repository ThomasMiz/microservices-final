package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
	"microservice-roomservice/pkg/config"
	"microservice-roomservice/pkg/logger"
)

// BillingService handles billing-related business logic
type BillingService struct {
	billingRepo repositories.BillingRepository
	config      *config.BillingConfig
	tracer      trace.Tracer
}

// NewBillingService creates a new billing service
func NewBillingService(billingRepo repositories.BillingRepository, config *config.BillingConfig) *BillingService {
	return &BillingService{
		billingRepo: billingRepo,
		config:      config,
		tracer:      otel.Tracer("microservice-billing"),
	}
}

// CreateBillingForOrder creates a billing ticket for an order in the reservation's billing folder.
// The billingFolderID must be provided from the reservation - we do not create new folders.
func (s *BillingService) CreateBillingForOrder(ctx context.Context, order *entities.Order, billingFolderID string) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}
	if billingFolderID == "" {
		return fmt.Errorf("billing folder ID is required")
	}

	// Simulate billing error if configured
	if s.config.ErrorRate > 0 {
		rand.Seed(time.Now().UnixNano())
		if rand.Float64() < s.config.ErrorRate {
			_, span := s.tracer.Start(ctx, "billing-error-simulation")
			defer span.End()

			span.SetAttributes(
				attribute.String("error.type", "simulated_billing_failure"),
				attribute.Float64("error.rate", s.config.ErrorRate),
				attribute.String("order.id", order.ID),
			)
			span.SetStatus(codes.Error, "Simulated billing ticket creation failure")

			return fmt.Errorf("simulated billing ticket creation failure (error rate: %.2f)", s.config.ErrorRate)
		}
	}

	// Build description with order details
	description := fmt.Sprintf("Room service order for room %s with %d items", order.RoomID, len(order.Items))

	// Create ticket for the order in the reservation's billing folder
	ticketReq := repositories.CreateTicketRequest{
		FolderID:    billingFolderID,
		Total:       order.TotalPrice,
		ItemID:      fmt.Sprintf("ROOMSERVICE-%s", order.ID),
		Description: description,
	}

	ticket, err := s.billingRepo.CreateTicket(ctx, ticketReq)
	if err != nil {
		return fmt.Errorf("failed to create billing ticket: %w", err)
	}

	// Update order with billing info (folder ID from reservation, ticket ID from billing service)
	order.SetBillingInfo(billingFolderID, ticket.ID)

	return nil
}

// DeleteBillingForOrder deletes billing ticket for an order (for rollback scenarios).
// Note: We only delete the ticket, not the folder, as the folder belongs to the reservation.
func (s *BillingService) DeleteBillingForOrder(ctx context.Context, order *entities.Order) {
	if order == nil || !order.HasBillingInfo() {
		return
	}

	// Delete ticket only - the folder belongs to the reservation
	if order.BillingTicketID != "" {
		if err := s.billingRepo.DeleteTicket(ctx, order.BillingTicketID); err != nil {
			logger.CtxError(ctx).Err(err).Msgf("Failed to delete billing ticket %s", order.BillingTicketID)
		}
	}
}
