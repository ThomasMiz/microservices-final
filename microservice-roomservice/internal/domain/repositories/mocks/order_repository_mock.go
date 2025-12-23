package mocks

import (
	"context"
	"microservice-roomservice/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock implementation of OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

// Save mocks the Save method
func (m *MockOrderRepository) Save(ctx context.Context, order *entities.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockOrderRepository) Update(ctx context.Context, order *entities.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*entities.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Order), args.Error(1)
}

// FindByRoomID mocks the FindByRoomID method
func (m *MockOrderRepository) FindByRoomID(ctx context.Context, roomID string) ([]entities.Order, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Order), args.Error(1)
}

// FindByReservationID mocks the FindByReservationID method
func (m *MockOrderRepository) FindByReservationID(ctx context.Context, reservationID string) ([]entities.Order, error) {
	args := m.Called(ctx, reservationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Order), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockOrderRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindByStatus mocks the FindByStatus method
func (m *MockOrderRepository) FindByStatus(ctx context.Context, status entities.OrderStatus) ([]entities.Order, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Order), args.Error(1)
}
