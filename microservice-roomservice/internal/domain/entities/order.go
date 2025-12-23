package entities

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	// OrderStatusPendingKitchen - Order created, waiting for kitchen stock validation
	OrderStatusPendingKitchen OrderStatus = "pending_kitchen"
	// OrderStatusPendingBilling - Kitchen accepted, waiting for billing confirmation
	OrderStatusPendingBilling OrderStatus = "pending_billing"
	// OrderStatusPending - Legacy status (kept for backward compatibility)
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusPreparing - Billing confirmed, kitchen is preparing
	OrderStatusPreparing OrderStatus = "preparing"
	// OrderStatusCompleted - Order completed
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusCancelled - Order cancelled (rejected by kitchen, billing failed, etc.)
	OrderStatusCancelled OrderStatus = "cancelled"
)

// IsValid checks if the order status is valid
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPendingKitchen, OrderStatusPendingBilling, OrderStatusPending, OrderStatusPreparing, OrderStatusCompleted, OrderStatusCancelled:
		return true
	}
	return false
}

// OrderItem represents a single item in an order
type OrderItem struct {
	ID         string
	MenuItemID string
	Quantity   int
	Price      decimal.Decimal
}

// NewOrderItem creates a new order item with validation
func NewOrderItem(id, menuItemID string, quantity int, price decimal.Decimal) (*OrderItem, error) {
	if id == "" {
		return nil, errors.New("order item ID cannot be empty")
	}
	if menuItemID == "" {
		return nil, errors.New("menu item ID cannot be empty")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}
	if price.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("price must be greater than zero")
	}

	return &OrderItem{
		ID:         id,
		MenuItemID: menuItemID,
		Quantity:   quantity,
		Price:      price,
	}, nil
}

// GetTotal calculates the total price for this order item
func (o *OrderItem) GetTotal() decimal.Decimal {
	return o.Price.Mul(decimal.NewFromInt(int64(o.Quantity)))
}

// Order represents a room service order
type Order struct {
	ID              string
	RoomID          string
	ReservationID   string
	BillingFolderID string
	BillingTicketID string
	Items           []OrderItem
	TotalPrice      decimal.Decimal
	Status          OrderStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewOrder creates a new order with validation
func NewOrder(id, roomID, reservationID string, items []OrderItem) (*Order, error) {
	if id == "" {
		return nil, errors.New("order ID cannot be empty")
	}
	if roomID == "" {
		return nil, errors.New("room ID cannot be empty")
	}
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	// Calculate total price
	totalPrice := decimal.Zero
	for _, item := range items {
		totalPrice = totalPrice.Add(item.GetTotal())
	}

	now := time.Now()
	return &Order{
		ID:            id,
		RoomID:        roomID,
		ReservationID: reservationID,
		Items:         items,
		TotalPrice:    totalPrice,
		Status:        OrderStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// UpdateStatus updates the order status with validation
func (o *Order) UpdateStatus(status OrderStatus) error {
	if !status.IsValid() {
		return errors.New("invalid order status")
	}

	// Validate status transitions for new orchestration flow
	switch o.Status {
	case OrderStatusPendingKitchen:
		// Kitchen can accept (pending_billing) or reject (cancelled)
		if status != OrderStatusPendingBilling && status != OrderStatusCancelled {
			return errors.New("pending_kitchen order can only transition to pending_billing or cancelled")
		}
	case OrderStatusPendingBilling:
		// Billing can succeed (preparing) or fail (cancelled)
		if status != OrderStatusPreparing && status != OrderStatusCancelled {
			return errors.New("pending_billing order can only transition to preparing or cancelled")
		}
	case OrderStatusPending:
		// Legacy: pending can go to preparing or cancelled
		if status != OrderStatusPreparing && status != OrderStatusCancelled {
			return errors.New("pending order can only transition to preparing or cancelled")
		}
	case OrderStatusPreparing:
		if status != OrderStatusCompleted && status != OrderStatusCancelled {
			return errors.New("preparing order can only transition to completed or cancelled")
		}
	case OrderStatusCompleted, OrderStatusCancelled:
		return errors.New("completed or cancelled orders cannot be updated")
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPendingKitchen ||
		o.Status == OrderStatusPendingBilling ||
		o.Status == OrderStatusPending ||
		o.Status == OrderStatusPreparing
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if !o.CanBeCancelled() {
		return errors.New("order cannot be cancelled")
	}
	return o.UpdateStatus(OrderStatusCancelled)
}

// IsPendingKitchen checks if the order is waiting for kitchen validation
func (o *Order) IsPendingKitchen() bool {
	return o.Status == OrderStatusPendingKitchen
}

// IsPendingBilling checks if the order is waiting for billing confirmation
func (o *Order) IsPendingBilling() bool {
	return o.Status == OrderStatusPendingBilling
}

// IsPending checks if the order is pending (legacy status)
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsPreparing checks if the order is being prepared
func (o *Order) IsPreparing() bool {
	return o.Status == OrderStatusPreparing
}

// IsCompleted checks if the order is completed
func (o *Order) IsCompleted() bool {
	return o.Status == OrderStatusCompleted
}

// IsCancelled checks if the order is cancelled
func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

// SetBillingInfo sets the billing folder and ticket IDs for the order
func (o *Order) SetBillingInfo(folderID, ticketID string) {
	o.BillingFolderID = folderID
	o.BillingTicketID = ticketID
	o.UpdatedAt = time.Now()
}

// HasBillingInfo checks if the order has billing information
func (o *Order) HasBillingInfo() bool {
	return o.BillingFolderID != "" && o.BillingTicketID != ""
}
