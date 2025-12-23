package testutil

import (
	"time"

	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateTestMenuItem creates a test menu item with default values
func CreateTestMenuItem(id, description string, price decimal.Decimal) *entities.MenuItem {
	if id == "" {
		id = uuid.New().String()
	}
	if description == "" {
		description = "Test Item"
	}
	if price.IsZero() {
		price = decimal.NewFromFloat(10.99)
	}

	item, _ := entities.NewMenuItem(id, description, price)
	return item
}

// CreateTestMenuItems creates multiple test menu items
func CreateTestMenuItems(count int) []entities.MenuItem {
	items := make([]entities.MenuItem, count)
	for i := 0; i < count; i++ {
		item := CreateTestMenuItem(
			uuid.New().String(),
			"Test Item "+string(rune(i+1)),
			decimal.NewFromFloat(float64(10+i)),
		)
		items[i] = *item
	}
	return items
}

// CreateTestOrderItem creates a test order item
func CreateTestOrderItem(menuItemID string, quantity int) *entities.OrderItem {
	if menuItemID == "" {
		menuItemID = uuid.New().String()
	}
	if quantity == 0 {
		quantity = 1
	}

	item, _ := entities.NewOrderItem(
		uuid.New().String(),
		menuItemID,
		quantity,
		decimal.NewFromFloat(15.99),
	)
	return item
}

// CreateTestOrder creates a test order with default values
func CreateTestOrder(id, roomID, reservationID string, items []entities.OrderItem) *entities.Order {
	if id == "" {
		id = uuid.New().String()
	}
	if roomID == "" {
		roomID = "ROOM-001"
	}
	if items == nil {
		item := CreateTestOrderItem("", 2)
		items = []entities.OrderItem{*item}
	}

	order, _ := entities.NewOrder(id, roomID, reservationID, items)
	return order
}

// CreateTestOrderWithStatus creates a test order with a specific status
func CreateTestOrderWithStatus(status entities.OrderStatus) *entities.Order {
	order := CreateTestOrder("", "", "", nil)
	if status != entities.OrderStatusPending {
		_ = order.UpdateStatus(status)
	}
	return order
}

// CreateTestOrderWithItems creates a test order with specific items
func CreateTestOrderWithItems(roomID string, itemCount int) *entities.Order {
	items := make([]entities.OrderItem, itemCount)
	for i := 0; i < itemCount; i++ {
		item := CreateTestOrderItem(uuid.New().String(), i+1)
		items[i] = *item
	}
	return CreateTestOrder("", roomID, "", items)
}

// TimeNow returns a fixed time for consistent testing
func TimeNow() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

// DecimalFromString creates a decimal from string, panics on error (for test convenience)
func DecimalFromString(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

// CreateTestReservation creates a test reservation with billing folder ID
func CreateTestReservation(id int64, roomID int64, billingFolderID string) *repositories.Reservation {
	if billingFolderID == "" {
		billingFolderID = "test-billing-folder-" + uuid.New().String()
	}
	return &repositories.Reservation{
		ID:              id,
		RoomID:          roomID,
		GuestID:         "test-guest-" + uuid.New().String(),
		StartDate:       TimeNow().Format(time.RFC3339),
		EndDate:         TimeNow().Add(24 * time.Hour).Format(time.RFC3339),
		BillingFolderID: billingFolderID,
	}
}

// CreateTestOrderWithBillingInfo creates a test order with billing info set
func CreateTestOrderWithBillingInfo(roomID, billingFolderID, billingTicketID string) *entities.Order {
	order := CreateTestOrder("", roomID, "", nil)
	order.SetBillingInfo(billingFolderID, billingTicketID)
	return order
}
