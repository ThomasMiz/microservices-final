package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"microservice-roomservice/internal/domain/repositories"
)

// MockBillingRepository is a mock implementation of BillingRepository
type MockBillingRepository struct {
	mock.Mock
}

// CreateTicket mocks the CreateTicket method
func (m *MockBillingRepository) CreateTicket(ctx context.Context, req repositories.CreateTicketRequest) (*repositories.BillingTicket, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.BillingTicket), args.Error(1)
}

// DeleteTicket mocks the DeleteTicket method
func (m *MockBillingRepository) DeleteTicket(ctx context.Context, ticketID string) error {
	args := m.Called(ctx, ticketID)
	return args.Error(0)
}
