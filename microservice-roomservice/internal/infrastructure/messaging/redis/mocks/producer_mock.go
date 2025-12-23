package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"microservice-roomservice/internal/infrastructure/messaging"
)

// MockProducer is a mock implementation of OrderProducer
type MockProducer struct {
	mock.Mock
}

// PublishOrder mocks the PublishOrder method
func (m *MockProducer) PublishOrder(ctx context.Context, message messaging.OrderMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

// PublishOrderEvent mocks the PublishOrderEvent method for saga orchestration
func (m *MockProducer) PublishOrderEvent(ctx context.Context, event messaging.OrderEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Close mocks the Close method (optional, for completeness)
func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}
