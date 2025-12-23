package entities_test

import (
	"testing"

	"microservice-roomservice/internal/domain/entities"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMenuItem(t *testing.T) {
	t.Run("successfully creates menu item with valid data", func(t *testing.T) {
		id := "menu-1"
		description := "Burger"
		price := decimal.NewFromFloat(12.99)

		item, err := entities.NewMenuItem(id, description, price)

		require.NoError(t, err)
		assert.Equal(t, id, item.ID)
		assert.Equal(t, description, item.Description)
		assert.True(t, price.Equal(item.Price))
		assert.False(t, item.CreatedAt.IsZero())
		assert.False(t, item.UpdatedAt.IsZero())
	})

	t.Run("fails with empty ID", func(t *testing.T) {
		item, err := entities.NewMenuItem("", "Burger", decimal.NewFromFloat(12.99))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item ID cannot be empty", err.Error())
	})

	t.Run("fails with empty description", func(t *testing.T) {
		item, err := entities.NewMenuItem("menu-1", "", decimal.NewFromFloat(12.99))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item description cannot be empty", err.Error())
	})

	t.Run("fails with zero price", func(t *testing.T) {
		item, err := entities.NewMenuItem("menu-1", "Burger", decimal.Zero)

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item price must be greater than zero", err.Error())
	})

	t.Run("fails with negative price", func(t *testing.T) {
		item, err := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(-5.00))

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item price must be greater than zero", err.Error())
	})
}

func TestMenuItem_UpdateDescription(t *testing.T) {
	t.Run("successfully updates description", func(t *testing.T) {
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		oldUpdatedAt := item.UpdatedAt

		err := item.UpdateDescription("Cheeseburger")

		require.NoError(t, err)
		assert.Equal(t, "Cheeseburger", item.Description)
		assert.True(t, item.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("fails with empty description", func(t *testing.T) {
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))

		err := item.UpdateDescription("")

		assert.Error(t, err)
		assert.Equal(t, "menu item description cannot be empty", err.Error())
		assert.Equal(t, "Burger", item.Description)
	})
}

func TestMenuItem_UpdatePrice(t *testing.T) {
	t.Run("successfully updates price", func(t *testing.T) {
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		oldUpdatedAt := item.UpdatedAt
		newPrice := decimal.NewFromFloat(14.99)

		err := item.UpdatePrice(newPrice)

		require.NoError(t, err)
		assert.True(t, newPrice.Equal(item.Price))
		assert.True(t, item.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("fails with zero price", func(t *testing.T) {
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		oldPrice := item.Price

		err := item.UpdatePrice(decimal.Zero)

		assert.Error(t, err)
		assert.Equal(t, "menu item price must be greater than zero", err.Error())
		assert.True(t, oldPrice.Equal(item.Price))
	})

	t.Run("fails with negative price", func(t *testing.T) {
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		oldPrice := item.Price

		err := item.UpdatePrice(decimal.NewFromFloat(-5.00))

		assert.Error(t, err)
		assert.Equal(t, "menu item price must be greater than zero", err.Error())
		assert.True(t, oldPrice.Equal(item.Price))
	})
}

func TestNewMenu(t *testing.T) {
	t.Run("creates empty menu", func(t *testing.T) {
		menu := entities.NewMenu()

		assert.NotNil(t, menu)
		assert.Empty(t, menu.Items)
	})
}

func TestMenu_AddItem(t *testing.T) {
	t.Run("successfully adds item to menu", func(t *testing.T) {
		menu := entities.NewMenu()
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))

		menu.AddItem(*item)

		assert.Len(t, menu.Items, 1)
		assert.Equal(t, item.ID, menu.Items[0].ID)
	})

	t.Run("adds multiple items to menu", func(t *testing.T) {
		menu := entities.NewMenu()
		item1, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		item2, _ := entities.NewMenuItem("menu-2", "Pizza", decimal.NewFromFloat(15.99))

		menu.AddItem(*item1)
		menu.AddItem(*item2)

		assert.Len(t, menu.Items, 2)
	})
}

func TestMenu_FindItemByID(t *testing.T) {
	t.Run("successfully finds item by ID", func(t *testing.T) {
		menu := entities.NewMenu()
		item1, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		item2, _ := entities.NewMenuItem("menu-2", "Pizza", decimal.NewFromFloat(15.99))
		menu.AddItem(*item1)
		menu.AddItem(*item2)

		found, err := menu.FindItemByID("menu-2")

		require.NoError(t, err)
		assert.Equal(t, "menu-2", found.ID)
		assert.Equal(t, "Pizza", found.Description)
	})

	t.Run("returns error when item not found", func(t *testing.T) {
		menu := entities.NewMenu()
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		menu.AddItem(*item)

		found, err := menu.FindItemByID("non-existent")

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Equal(t, "menu item not found", err.Error())
	})
}

func TestMenu_RemoveItem(t *testing.T) {
	t.Run("successfully removes item from menu", func(t *testing.T) {
		menu := entities.NewMenu()
		item1, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		item2, _ := entities.NewMenuItem("menu-2", "Pizza", decimal.NewFromFloat(15.99))
		menu.AddItem(*item1)
		menu.AddItem(*item2)

		err := menu.RemoveItem("menu-1")

		require.NoError(t, err)
		assert.Len(t, menu.Items, 1)
		assert.Equal(t, "menu-2", menu.Items[0].ID)
	})

	t.Run("returns error when item not found", func(t *testing.T) {
		menu := entities.NewMenu()
		item, _ := entities.NewMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		menu.AddItem(*item)

		err := menu.RemoveItem("non-existent")

		assert.Error(t, err)
		assert.Equal(t, "menu item not found", err.Error())
		assert.Len(t, menu.Items, 1)
	})
}
