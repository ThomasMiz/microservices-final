package entities_test

import (
	"testing"

	"microservice-roomservice/internal/domain/entities"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status entities.OrderStatus
		want   bool
	}{
		{"pending is valid", entities.OrderStatusPending, true},
		{"preparing is valid", entities.OrderStatusPreparing, true},
		{"completed is valid", entities.OrderStatusCompleted, true},
		{"cancelled is valid", entities.OrderStatusCancelled, true},
		{"invalid status", entities.OrderStatus("invalid"), false},
		{"empty status", entities.OrderStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestNewOrderItem(t *testing.T) {
	t.Run("successfully creates order item with valid data", func(t *testing.T) {
		id := "item-1"
		menuItemID := "menu-1"
		quantity := 2
		price := decimal.NewFromFloat(12.99)

		item, err := entities.NewOrderItem(id, menuItemID, quantity, price)

		require.NoError(t, err)
		assert.Equal(t, id, item.ID)
		assert.Equal(t, menuItemID, item.MenuItemID)
		assert.Equal(t, quantity, item.Quantity)
		assert.True(t, price.Equal(item.Price))
	})

	t.Run("fails with empty ID", func(t *testing.T) {
		item, err := entities.NewOrderItem("", "menu-1", 2, decimal.NewFromFloat(12.99))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "order item ID cannot be empty", err.Error())
	})

	t.Run("fails with empty menu item ID", func(t *testing.T) {
		item, err := entities.NewOrderItem("item-1", "", 2, decimal.NewFromFloat(12.99))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item ID cannot be empty", err.Error())
	})

	t.Run("fails with zero quantity", func(t *testing.T) {
		item, err := entities.NewOrderItem("item-1", "menu-1", 0, decimal.NewFromFloat(12.99))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "quantity must be greater than zero", err.Error())
	})

	t.Run("fails with negative quantity", func(t *testing.T) {
		item, err := entities.NewOrderItem("item-1", "menu-1", -1, decimal.NewFromFloat(12.99))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "quantity must be greater than zero", err.Error())
	})

	t.Run("fails with zero price", func(t *testing.T) {
		item, err := entities.NewOrderItem("item-1", "menu-1", 2, decimal.Zero)

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "price must be greater than zero", err.Error())
	})

	t.Run("fails with negative price", func(t *testing.T) {
		item, err := entities.NewOrderItem("item-1", "menu-1", 2, decimal.NewFromFloat(-5.00))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "price must be greater than zero", err.Error())
	})
}

func TestOrderItem_GetTotal(t *testing.T) {
	tests := []struct {
		name     string
		price    decimal.Decimal
		quantity int
		want     string
	}{
		{"single item", decimal.NewFromFloat(10.00), 1, "10"},
		{"multiple items", decimal.NewFromFloat(12.50), 3, "37.5"},
		{"decimal precision", decimal.NewFromFloat(9.99), 2, "19.98"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := entities.NewOrderItem("item-1", "menu-1", tt.quantity, tt.price)
			total := item.GetTotal()
			expected, _ := decimal.NewFromString(tt.want)
			assert.True(t, expected.Equal(total))
		})
	}
}

func TestNewOrder(t *testing.T) {
	t.Run("successfully creates order with valid data", func(t *testing.T) {
		id := "order-1"
		roomID := "ROOM-001"
		reservationID := "reservation-1"
		item1, _ := entities.NewOrderItem("item-1", "menu-1", 2, decimal.NewFromFloat(10.00))
		item2, _ := entities.NewOrderItem("item-2", "menu-2", 1, decimal.NewFromFloat(15.00))
		items := []entities.OrderItem{*item1, *item2}

		order, err := entities.NewOrder(id, roomID, reservationID, items)

		require.NoError(t, err)
		assert.Equal(t, id, order.ID)
		assert.Equal(t, roomID, order.RoomID)
		assert.Equal(t, reservationID, order.ReservationID)
		assert.Len(t, order.Items, 2)
		assert.True(t, decimal.NewFromFloat(35.00).Equal(order.TotalPrice))
		assert.Equal(t, entities.OrderStatusPending, order.Status)
		assert.False(t, order.CreatedAt.IsZero())
		assert.False(t, order.UpdatedAt.IsZero())
	})

	t.Run("creates order without reservation ID", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		items := []entities.OrderItem{*item}

		order, err := entities.NewOrder("order-1", "ROOM-001", "", items)

		require.NoError(t, err)
		assert.Equal(t, "", order.ReservationID)
	})

	t.Run("fails with empty order ID", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		items := []entities.OrderItem{*item}

		order, err := entities.NewOrder("", "ROOM-001", "", items)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "order ID cannot be empty", err.Error())
	})

	t.Run("fails with empty room ID", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		items := []entities.OrderItem{*item}

		order, err := entities.NewOrder("order-1", "", "", items)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "room ID cannot be empty", err.Error())
	})

	t.Run("fails with empty items", func(t *testing.T) {
		order, err := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{})

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, "order must have at least one item", err.Error())
	})

	t.Run("calculates correct total price", func(t *testing.T) {
		item1, _ := entities.NewOrderItem("item-1", "menu-1", 3, decimal.NewFromFloat(12.50))
		item2, _ := entities.NewOrderItem("item-2", "menu-2", 2, decimal.NewFromFloat(7.25))
		items := []entities.OrderItem{*item1, *item2}

		order, err := entities.NewOrder("order-1", "ROOM-001", "", items)

		require.NoError(t, err)
		// 3 * 12.50 + 2 * 7.25 = 37.50 + 14.50 = 52.00
		expected := decimal.NewFromFloat(52.00)
		assert.True(t, expected.Equal(order.TotalPrice))
	})
}

func TestOrder_UpdateStatus(t *testing.T) {
	t.Run("pending to preparing", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})

		err := order.UpdateStatus(entities.OrderStatusPreparing)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusPreparing, order.Status)
	})

	t.Run("pending to cancelled", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})

		err := order.UpdateStatus(entities.OrderStatusCancelled)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCancelled, order.Status)
	})

	t.Run("preparing to completed", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)

		err := order.UpdateStatus(entities.OrderStatusCompleted)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCompleted, order.Status)
	})

	t.Run("preparing to cancelled", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)

		err := order.UpdateStatus(entities.OrderStatusCancelled)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCancelled, order.Status)
	})

	t.Run("fails with invalid status", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})

		err := order.UpdateStatus(entities.OrderStatus("invalid"))

		assert.Error(t, err)
		assert.Equal(t, "invalid order status", err.Error())
		assert.Equal(t, entities.OrderStatusPending, order.Status)
	})

	t.Run("fails pending to completed", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})

		err := order.UpdateStatus(entities.OrderStatusCompleted)

		assert.Error(t, err)
		assert.Equal(t, "pending order can only transition to preparing or cancelled", err.Error())
		assert.Equal(t, entities.OrderStatusPending, order.Status)
	})

	t.Run("fails completed to any status", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)
		_ = order.UpdateStatus(entities.OrderStatusCompleted)

		err := order.UpdateStatus(entities.OrderStatusCancelled)

		assert.Error(t, err)
		assert.Equal(t, "completed or cancelled orders cannot be updated", err.Error())
		assert.Equal(t, entities.OrderStatusCompleted, order.Status)
	})

	t.Run("fails cancelled to any status", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusCancelled)

		err := order.UpdateStatus(entities.OrderStatusPending)

		assert.Error(t, err)
		assert.Equal(t, "completed or cancelled orders cannot be updated", err.Error())
		assert.Equal(t, entities.OrderStatusCancelled, order.Status)
	})
}

func TestOrder_CanBeCancelled(t *testing.T) {
	tests := []struct {
		name   string
		status entities.OrderStatus
		want   bool
	}{
		{"pending can be cancelled", entities.OrderStatusPending, true},
		{"preparing can be cancelled", entities.OrderStatusPreparing, true},
		{"completed cannot be cancelled", entities.OrderStatusCompleted, false},
		{"cancelled cannot be cancelled", entities.OrderStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
			order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})

			switch tt.status {
			case entities.OrderStatusPreparing:
				_ = order.UpdateStatus(entities.OrderStatusPreparing)
			case entities.OrderStatusCompleted:
				_ = order.UpdateStatus(entities.OrderStatusPreparing)
				_ = order.UpdateStatus(entities.OrderStatusCompleted)
			case entities.OrderStatusCancelled:
				_ = order.UpdateStatus(entities.OrderStatusCancelled)
			}

			assert.Equal(t, tt.want, order.CanBeCancelled())
		})
	}
}

func TestOrder_Cancel(t *testing.T) {
	t.Run("successfully cancels pending order", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})

		err := order.Cancel()

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCancelled, order.Status)
	})

	t.Run("successfully cancels preparing order", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)

		err := order.Cancel()

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCancelled, order.Status)
	})

	t.Run("fails to cancel completed order", func(t *testing.T) {
		item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)
		_ = order.UpdateStatus(entities.OrderStatusCompleted)

		err := order.Cancel()

		assert.Error(t, err)
		assert.Equal(t, "order cannot be cancelled", err.Error())
		assert.Equal(t, entities.OrderStatusCompleted, order.Status)
	})
}

func TestOrder_StatusChecks(t *testing.T) {
	item, _ := entities.NewOrderItem("item-1", "menu-1", 1, decimal.NewFromFloat(10.00))

	t.Run("IsPending", func(t *testing.T) {
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		assert.True(t, order.IsPending())
		assert.False(t, order.IsPreparing())
		assert.False(t, order.IsCompleted())
		assert.False(t, order.IsCancelled())
	})

	t.Run("IsPreparing", func(t *testing.T) {
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)
		assert.False(t, order.IsPending())
		assert.True(t, order.IsPreparing())
		assert.False(t, order.IsCompleted())
		assert.False(t, order.IsCancelled())
	})

	t.Run("IsCompleted", func(t *testing.T) {
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.UpdateStatus(entities.OrderStatusPreparing)
		_ = order.UpdateStatus(entities.OrderStatusCompleted)
		assert.False(t, order.IsPending())
		assert.False(t, order.IsPreparing())
		assert.True(t, order.IsCompleted())
		assert.False(t, order.IsCancelled())
	})

	t.Run("IsCancelled", func(t *testing.T) {
		order, _ := entities.NewOrder("order-1", "ROOM-001", "", []entities.OrderItem{*item})
		_ = order.Cancel()
		assert.False(t, order.IsPending())
		assert.False(t, order.IsPreparing())
		assert.False(t, order.IsCompleted())
		assert.True(t, order.IsCancelled())
	})
}
