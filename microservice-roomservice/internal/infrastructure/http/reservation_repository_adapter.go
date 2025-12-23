package http

import (
	"context"

	"microservice-roomservice/internal/domain/repositories"
)

// ReservationRepositoryAdapter implements the ReservationRepository interface using the HTTP client
type ReservationRepositoryAdapter struct {
	client *ReservationClient
}

// NewReservationRepositoryAdapter creates a new reservation repository adapter
func NewReservationRepositoryAdapter(client *ReservationClient) repositories.ReservationRepository {
	return &ReservationRepositoryAdapter{
		client: client,
	}
}

// GetActiveReservationForRoom retrieves the active reservation for a given room
func (a *ReservationRepositoryAdapter) GetActiveReservationForRoom(ctx context.Context, roomID int64) (*repositories.Reservation, error) {
	reservation, err := a.client.GetActiveReservationForRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return reservation.ToDomain(), nil
}

// GetReservationByID retrieves a reservation by its ID
func (a *ReservationRepositoryAdapter) GetReservationByID(ctx context.Context, reservationID int64) (*repositories.Reservation, error) {
	reservation, err := a.client.GetReservationByID(ctx, reservationID)
	if err != nil {
		return nil, err
	}
	if reservation == nil {
		return nil, nil
	}
	return reservation.ToDomain(), nil
}

// GetRoomByID retrieves a room by its ID
func (a *ReservationRepositoryAdapter) GetRoomByID(ctx context.Context, roomID int64) (*repositories.Room, error) {
	room, err := a.client.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return room.ToDomain(), nil
}

// ValidateRoom checks if a room exists and is active
func (a *ReservationRepositoryAdapter) ValidateRoom(ctx context.Context, roomID int64) (bool, error) {
	return a.client.ValidateRoom(ctx, roomID)
}
