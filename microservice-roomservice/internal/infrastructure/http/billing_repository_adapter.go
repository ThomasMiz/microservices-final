package http

import (
	"context"

	"microservice-roomservice/internal/domain/repositories"

	"github.com/shopspring/decimal"
)

// BillingRepositoryAdapter implements the BillingRepository interface using the HTTP client
type BillingRepositoryAdapter struct {
	client *BillingClient
}

// NewBillingRepositoryAdapter creates a new billing repository adapter
func NewBillingRepositoryAdapter(client *BillingClient) repositories.BillingRepository {
	return &BillingRepositoryAdapter{
		client: client,
	}
}

// CreateTicket creates a new billing ticket
func (a *BillingRepositoryAdapter) CreateTicket(ctx context.Context, req repositories.CreateTicketRequest) (*repositories.BillingTicket, error) {
	ticket, err := a.client.CreateTicket(ctx, req.FolderID, req.Total, req.ItemID, req.Description)
	if err != nil {
		return nil, err
	}

	total, err := decimal.NewFromString(ticket.Total)
	if err != nil {
		total = decimal.Zero
	}

	description := ""
	if ticket.Description != nil {
		description = *ticket.Description
	}

	return &repositories.BillingTicket{
		ID:          ticket.ID,
		FolderID:    ticket.Folder.ID,
		Total:       total,
		ItemID:      ticket.ItemID,
		Description: description,
		State:       repositories.BillingTicketState(ticket.State),
		CreatedAt:   ticket.CreatedAt,
		ClosedAt:    ticket.ClosedAt,
	}, nil
}

// DeleteTicket deletes a billing ticket by ID
func (a *BillingRepositoryAdapter) DeleteTicket(ctx context.Context, ticketID string) error {
	return a.client.DeleteTicket(ctx, ticketID)
}
