package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"microservice-roomservice/internal/domain/repositories"
	"microservice-roomservice/pkg/logger"
)

// ReservationClient is an HTTP client for the reservation service
type ReservationClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewReservationClient creates a new reservation client
func NewReservationClient(baseURL string) *ReservationClient {
	return &ReservationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			// Use default transport without otelhttp wrapping
			// We'll handle tracing manually
		},
	}
}

// Reservation represents a reservation from the reservation service
type Reservation struct {
	ID                int64   `json:"id"`
	Room              Room    `json:"room"`
	GuestID           string  `json:"guestId"`
	StartDate         string  `json:"startDate"`
	EndDate           string  `json:"endDate"`
	BillingFolderID   string  `json:"billingFolderId"`
	ReservationTicket string  `json:"reservationBillingTicket"`
	RentedHourlyPrice float64 `json:"rentedHourlyPrice"`
	TotalPrice        float64 `json:"totalPrice"`
}

// Room represents a room from the reservation service
type Room struct {
	ID          int64   `json:"id"`
	Number      string  `json:"number"`
	Active      bool    `json:"active"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MaxCapacity int     `json:"maxCapacity"`
	Category    string  `json:"category"`
	HourlyPrice float64 `json:"hourlyPrice"`
}

// GetActiveReservationForRoom retrieves the active reservation for a given room
func (c *ReservationClient) GetActiveReservationForRoom(ctx context.Context, roomID int64) (*Reservation, error) {
	// Check if context has a parent span
	parentSpan := trace.SpanFromContext(ctx)
	parentSpanCtx := parentSpan.SpanContext()
	logger.CtxDebug(ctx).Msgf("ReservationClient.GetActiveReservationForRoom - Parent span context valid: %v, traceID: %s, spanID: %s",
		parentSpanCtx.IsValid(),
		parentSpanCtx.TraceID().String(),
		parentSpanCtx.SpanID().String())

	tracer := otel.Tracer("reservation-client")
	logger.CtxDebug(ctx).Msgf("ReservationClient.GetActiveReservationForRoom - Tracer type: %T", tracer)

	ctx, span := tracer.Start(ctx, "GET reservation")
	defer span.End()

	spanCtx := span.SpanContext()
	logger.CtxDebug(ctx).Msgf("ReservationClient.GetActiveReservationForRoom - Created span: traceID=%s spanID=%s isRecording=%v isValid=%v isSampled=%v",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		span.IsRecording(),
		spanCtx.IsValid(),
		spanCtx.IsSampled())

	span.SetAttributes(attribute.Int64("room.id", roomID))

	now := time.Now().Format(time.RFC3339)
	params := url.Values{}
	params.Add("roomIds", fmt.Sprintf("%d", roomID))
	params.Add("from", now)
	params.Add("to", now)
	// Sort by ID descending to get the most recent reservation first
	params.Add("sortBy", "ID")
	params.Add("direction", "DESCENDING")

	url := fmt.Sprintf("%s/reservations?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Manually inject trace context into HTTP headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reservation service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var reservations []Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservations); err != nil {
		return nil, err
	}

	if len(reservations) == 0 {
		return nil, errors.New("no active reservation found for room")
	}

	// Log warning if multiple reservations found
	if len(reservations) > 1 {
		logger.CtxWarn(ctx).Msgf("Multiple active reservations found for room %d (count: %d), using most recent (ID: %d)",
			roomID, len(reservations), reservations[0].ID)
	}

	// Return the first reservation (most recent due to ID DESC sort)
	return &reservations[0], nil
}

// GetReservationByID retrieves a reservation by its ID
func (c *ReservationClient) GetReservationByID(ctx context.Context, reservationID int64) (*Reservation, error) {
	url := fmt.Sprintf("%s/reservations/%d", c.baseURL, reservationID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reservation service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var reservation Reservation
	if err := json.NewDecoder(resp.Body).Decode(&reservation); err != nil {
		return nil, err
	}

	return &reservation, nil
}

// GetRoomByID retrieves a room by its ID
func (c *ReservationClient) GetRoomByID(ctx context.Context, roomID int64) (*Room, error) {
	url := fmt.Sprintf("%s/rooms/%d", c.baseURL, roomID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("room not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reservation service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var room Room
	if err := json.NewDecoder(resp.Body).Decode(&room); err != nil {
		return nil, err
	}

	if !room.Active {
		return nil, errors.New("room is not active")
	}

	return &room, nil
}

// ValidateRoom checks if a room exists and is active
func (c *ReservationClient) ValidateRoom(ctx context.Context, roomID int64) (bool, error) {
	room, err := c.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, errors.New("room not found")) {
			return false, nil
		}
		return false, err
	}
	return room.Active, nil
}

// ToDomain converts HTTP Reservation to domain Reservation
func (r *Reservation) ToDomain() *repositories.Reservation {
	return &repositories.Reservation{
		ID:                r.ID,
		RoomID:            r.Room.ID,
		GuestID:           r.GuestID,
		StartDate:         r.StartDate,
		EndDate:           r.EndDate,
		BillingFolderID:   r.BillingFolderID,
		ReservationTicket: r.ReservationTicket,
		RentedHourlyPrice: r.RentedHourlyPrice,
		TotalPrice:        r.TotalPrice,
	}
}

// ToDomain converts HTTP Room to domain Room
func (r *Room) ToDomain() *repositories.Room {
	return &repositories.Room{
		ID:           r.ID,
		Name:         r.Name,
		Description:  r.Description,
		BasePrice:    r.HourlyPrice,
		Active:       r.Active,
		RoomCategory: r.Category,
	}
}
