package messaging

import "time"

// OrderMessage represents an order message sent to messaging system
type OrderMessage struct {
	OrderID       string    `json:"order_id"`
	RoomID        string    `json:"room_id"`
	ReservationID string    `json:"reservation_id"`
	Items         []Item    `json:"items"`
	TotalPrice    string    `json:"total_price"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// Item represents an item in an order message
type Item struct {
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
	Price      string `json:"price"`
}

// OrderEvent represents an orchestration event for the saga pattern
type OrderEvent struct {
	EventType       string    `json:"event_type"`
	OrderID         string    `json:"order_id"`
	RoomID          string    `json:"room_id"`
	ReservationID   string    `json:"reservation_id"`
	Items           []Item    `json:"items"`
	TotalPrice      string    `json:"total_price"`
	BillingFolderID string    `json:"billing_folder_id,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

// OrderCompletionMessage represents an order completion message
type OrderCompletionMessage struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
