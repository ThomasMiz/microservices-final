package repositories

import (
	"context"

	"github.com/shopspring/decimal"
)

// BillingRepository defines the interface for billing service integration
type BillingRepository interface {
	// CreateTicket creates a new billing ticket within a folder
	CreateTicket(ctx context.Context, req CreateTicketRequest) (*BillingTicket, error)

	// DeleteTicket deletes a billing ticket by ID
	DeleteTicket(ctx context.Context, ticketID string) error
}

// BillingFolder represents a billing folder from the billing service
type BillingFolder struct {
	ID        string
	Name      string
	CreatedAt string
	ClosedAt  *string
}

// BillingTicket represents a billing ticket from the billing service
type BillingTicket struct {
	ID          string
	FolderID    string
	Total       decimal.Decimal
	ItemID      string
	Description string
	State       BillingTicketState
	CreatedAt   string
	ClosedAt    *string
}

// BillingTicketState represents the state of a billing ticket
type BillingTicketState string

const (
	BillingTicketStatePending  BillingTicketState = "Pending"
	BillingTicketStatePaying   BillingTicketState = "Paying"
	BillingTicketStatePaid     BillingTicketState = "Paid"
	BillingTicketStateCanceled BillingTicketState = "Canceled"
)

// CreateTicketRequest represents the request to create a billing ticket
type CreateTicketRequest struct {
	FolderID    string
	Total       decimal.Decimal
	ItemID      string
	Description string
}
