package usecases_test

import (
	"context"
	"errors"
	"testing"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/testutil"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewManageMenu(t *testing.T) {
	menuRepo := new(mocks.MockMenuRepository)
	uc := usecases.NewManageMenu(menuRepo)

	assert.NotNil(t, uc)
}

func TestManageMenu_AddMenuItem(t *testing.T) {
	t.Run("successfully adds menu item", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		menuRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(nil)

		req := usecases.AddMenuItemRequest{
			Description: "Burger",
			Price:       decimal.NewFromFloat(12.99),
		}

		item, err := uc.AddMenuItem(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, "Burger", item.Description)
		assert.True(t, decimal.NewFromFloat(12.99).Equal(item.Price))
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails with empty description", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		req := usecases.AddMenuItemRequest{
			Description: "",
			Price:       decimal.NewFromFloat(12.99),
		}

		item, err := uc.AddMenuItem(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item description cannot be empty", err.Error())
	})

	t.Run("fails with zero price", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		req := usecases.AddMenuItemRequest{
			Description: "Burger",
			Price:       decimal.Zero,
		}

		item, err := uc.AddMenuItem(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, "menu item price must be greater than zero", err.Error())
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		repoError := errors.New("database error")
		menuRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(repoError)

		req := usecases.AddMenuItemRequest{
			Description: "Burger",
			Price:       decimal.NewFromFloat(12.99),
		}

		item, err := uc.AddMenuItem(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, repoError, err)
		menuRepo.AssertExpectations(t)
	})
}

func TestManageMenu_UpdateMenuItem(t *testing.T) {
	t.Run("successfully updates description and price", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		existingItem := testutil.CreateTestMenuItem("menu-1", "Old Description", decimal.NewFromFloat(10.00))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(existingItem, nil)
		menuRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(nil)

		req := usecases.UpdateMenuItemRequest{
			ID:          "menu-1",
			Description: "New Description",
			Price:       decimal.NewFromFloat(15.00),
		}

		item, err := uc.UpdateMenuItem(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, "New Description", item.Description)
		assert.True(t, decimal.NewFromFloat(15.00).Equal(item.Price))
		menuRepo.AssertExpectations(t)
	})

	t.Run("successfully updates only description", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		existingItem := testutil.CreateTestMenuItem("menu-1", "Old Description", decimal.NewFromFloat(10.00))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(existingItem, nil)
		menuRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(nil)

		req := usecases.UpdateMenuItemRequest{
			ID:          "menu-1",
			Description: "New Description",
			Price:       decimal.Zero,
		}

		item, err := uc.UpdateMenuItem(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, "New Description", item.Description)
		assert.True(t, decimal.NewFromFloat(10.00).Equal(item.Price))
		menuRepo.AssertExpectations(t)
	})

	t.Run("successfully updates only price", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		existingItem := testutil.CreateTestMenuItem("menu-1", "Old Description", decimal.NewFromFloat(10.00))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(existingItem, nil)
		menuRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(nil)

		req := usecases.UpdateMenuItemRequest{
			ID:          "menu-1",
			Description: "",
			Price:       decimal.NewFromFloat(20.00),
		}

		item, err := uc.UpdateMenuItem(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, "Old Description", item.Description)
		assert.True(t, decimal.NewFromFloat(20.00).Equal(item.Price))
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when item not found", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(nil, errors.New("not found"))

		req := usecases.UpdateMenuItemRequest{
			ID:          "menu-1",
			Description: "New Description",
			Price:       decimal.NewFromFloat(15.00),
		}

		item, err := uc.UpdateMenuItem(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, item)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when update repository fails", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		existingItem := testutil.CreateTestMenuItem("menu-1", "Old Description", decimal.NewFromFloat(10.00))
		repoError := errors.New("database error")
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(existingItem, nil)
		menuRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(repoError)

		req := usecases.UpdateMenuItemRequest{
			ID:          "menu-1",
			Description: "New Description",
			Price:       decimal.NewFromFloat(15.00),
		}

		item, err := uc.UpdateMenuItem(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, item)
		assert.Equal(t, repoError, err)
		menuRepo.AssertExpectations(t)
	})
}

func TestManageMenu_RemoveMenuItem(t *testing.T) {
	t.Run("successfully removes menu item", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		menuRepo.On("Delete", mock.Anything, "menu-1").Return(nil)

		err := uc.RemoveMenuItem(context.Background(), "menu-1")

		require.NoError(t, err)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		repoError := errors.New("database error")
		menuRepo.On("Delete", mock.Anything, "menu-1").Return(repoError)

		err := uc.RemoveMenuItem(context.Background(), "menu-1")

		assert.Error(t, err)
		assert.Equal(t, repoError, err)
		menuRepo.AssertExpectations(t)
	})
}

func TestManageMenu_GetMenuItem(t *testing.T) {
	t.Run("successfully gets menu item", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		expectedItem := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.99))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(expectedItem, nil)

		item, err := uc.GetMenuItem(context.Background(), "menu-1")

		require.NoError(t, err)
		assert.Equal(t, expectedItem, item)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when item not found", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(nil, errors.New("not found"))

		item, err := uc.GetMenuItem(context.Background(), "menu-1")

		assert.Error(t, err)
		assert.Nil(t, item)
		menuRepo.AssertExpectations(t)
	})
}

func TestManageMenu_GetAllMenuItems(t *testing.T) {
	t.Run("successfully gets all menu items", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		expectedItems := testutil.CreateTestMenuItems(3)
		menuRepo.On("FindAll", mock.Anything).Return(expectedItems, nil)

		items, err := uc.GetAllMenuItems(context.Background())

		require.NoError(t, err)
		assert.Len(t, items, 3)
		menuRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no items", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		menuRepo.On("FindAll", mock.Anything).Return([]entities.MenuItem{}, nil)

		items, err := uc.GetAllMenuItems(context.Background())

		require.NoError(t, err)
		assert.Empty(t, items)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		uc := usecases.NewManageMenu(menuRepo)

		repoError := errors.New("database error")
		menuRepo.On("FindAll", mock.Anything).Return(nil, repoError)

		items, err := uc.GetAllMenuItems(context.Background())

		assert.Error(t, err)
		assert.Nil(t, items)
		assert.Equal(t, repoError, err)
		menuRepo.AssertExpectations(t)
	})
}
