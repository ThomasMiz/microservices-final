package repositories

import "context"

// ReservationRepository defines the interface for reservation service integration
type ReservationRepository interface {
	// GetActiveReservationForRoom retrieves the active reservation for a given room
	GetActiveReservationForRoom(ctx context.Context, roomID int64) (*Reservation, error)

	// GetReservationByID retrieves a reservation by its ID
	GetReservationByID(ctx context.Context, reservationID int64) (*Reservation, error)

	// GetRoomByID retrieves a room by its ID
	GetRoomByID(ctx context.Context, roomID int64) (*Room, error)

	// ValidateRoom checks if a room exists and is active
	ValidateRoom(ctx context.Context, roomID int64) (bool, error)
}

// Reservation represents a reservation from the reservation service
type Reservation struct {
	ID                int64
	RoomID            int64
	GuestID           string
	StartDate         string
	EndDate           string
	BillingFolderID   string
	ReservationTicket string
	RentedHourlyPrice float64
	TotalPrice        float64
}

// Room represents a room from the reservation service
type Room struct {
	ID           int64
	Name         string
	Description  string
	BasePrice    float64
	Active       bool
	RoomCategory string
}
