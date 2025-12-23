package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"microservice-roomservice/internal/domain/repositories"
)

// MockReservationRepository is a mock implementation of ReservationRepository
type MockReservationRepository struct {
	mock.Mock
}

// GetActiveReservationForRoom mocks the GetActiveReservationForRoom method
func (m *MockReservationRepository) GetActiveReservationForRoom(ctx context.Context, roomID int64) (*repositories.Reservation, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Reservation), args.Error(1)
}

// GetReservationByID mocks the GetReservationByID method
func (m *MockReservationRepository) GetReservationByID(ctx context.Context, reservationID int64) (*repositories.Reservation, error) {
	args := m.Called(ctx, reservationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Reservation), args.Error(1)
}

// GetRoomByID mocks the GetRoomByID method
func (m *MockReservationRepository) GetRoomByID(ctx context.Context, roomID int64) (*repositories.Room, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Room), args.Error(1)
}

// ValidateRoom mocks the ValidateRoom method
func (m *MockReservationRepository) ValidateRoom(ctx context.Context, roomID int64) (bool, error) {
	args := m.Called(ctx, roomID)
	return args.Bool(0), args.Error(1)
}
