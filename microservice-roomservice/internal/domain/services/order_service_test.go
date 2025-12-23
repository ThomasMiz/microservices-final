package services_test

import (
	"context"
	"errors"
	"testing"

	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
	"microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/internal/testutil"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewOrderService(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	menuRepo := new(mocks.MockMenuRepository)
	reservationRepo := new(mocks.MockReservationRepository)

	service := services.NewOrderService(orderRepo, menuRepo, reservationRepo)

	assert.NotNil(t, service)
}

func TestOrderService_ValidateOrder(t *testing.T) {
	t.Run("successfully validates valid order", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		menuRepo.On("ExistsByID", mock.Anything, "menu-2").Return(true, nil)

		req := services.CreateOrderRequest{
			RoomID: 1,
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
				{MenuItemID: "menu-2", Quantity: 1},
			},
		}

		err := service.ValidateOrder(context.Background(), req)

		require.NoError(t, err)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails with empty room ID", func(t *testing.T) {
		service := services.NewOrderService(nil, nil, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(0),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		err := service.ValidateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "room ID is required", err.Error())
	})

	t.Run("fails with empty items", func(t *testing.T) {
		service := services.NewOrderService(nil, nil, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items:  []services.OrderItemRequest{},
		}

		err := service.ValidateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "order must have at least one item", err.Error())
	})

	t.Run("fails when menu item does not exist", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(false, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		err := service.ValidateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "menu item not found: menu-1", err.Error())
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails with zero quantity", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 0},
			},
		}

		err := service.ValidateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "quantity must be greater than zero", err.Error())
	})

	t.Run("fails with negative quantity", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: -1},
			},
		}

		err := service.ValidateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, "quantity must be greater than zero", err.Error())
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		repoError := errors.New("database error")
		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(false, repoError)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		err := service.ValidateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Equal(t, repoError, err)
		menuRepo.AssertExpectations(t)
	})
}

func TestOrderService_CalculateTotalPrice(t *testing.T) {
	t.Run("successfully calculates total price for single item", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuItem := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.50))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem, nil)

		items := []services.OrderItemRequest{
			{MenuItemID: "menu-1", Quantity: 3},
		}

		total, orderItems, err := service.CalculateTotalPrice(context.Background(), items)

		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(37.50).Equal(total)) // 12.50 * 3
		assert.Len(t, orderItems, 1)
		assert.Equal(t, "menu-1", orderItems[0].MenuItemID)
		assert.Equal(t, 3, orderItems[0].Quantity)
		menuRepo.AssertExpectations(t)
	})

	t.Run("successfully calculates total price for multiple items", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuItem1 := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(12.50))
		menuItem2 := testutil.CreateTestMenuItem("menu-2", "Pizza", decimal.NewFromFloat(15.00))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem1, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-2").Return(menuItem2, nil)

		items := []services.OrderItemRequest{
			{MenuItemID: "menu-1", Quantity: 2},
			{MenuItemID: "menu-2", Quantity: 1},
		}

		total, orderItems, err := service.CalculateTotalPrice(context.Background(), items)

		require.NoError(t, err)
		assert.True(t, decimal.NewFromFloat(40.00).Equal(total)) // 12.50*2 + 15.00*1
		assert.Len(t, orderItems, 2)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when menu item not found", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(nil, errors.New("not found"))

		items := []services.OrderItemRequest{
			{MenuItemID: "menu-1", Quantity: 2},
		}

		total, orderItems, err := service.CalculateTotalPrice(context.Background(), items)

		assert.Error(t, err)
		assert.True(t, total.IsZero())
		assert.Nil(t, orderItems)
		menuRepo.AssertExpectations(t)
	})

	t.Run("fails when menu item is nil", func(t *testing.T) {
		menuRepo := new(mocks.MockMenuRepository)
		service := services.NewOrderService(nil, menuRepo, nil)

		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(nil, nil)

		items := []services.OrderItemRequest{
			{MenuItemID: "menu-1", Quantity: 2},
		}

		total, orderItems, err := service.CalculateTotalPrice(context.Background(), items)

		assert.Error(t, err)
		assert.Equal(t, "menu item not found: menu-1", err.Error())
		assert.True(t, total.IsZero())
		assert.Nil(t, orderItems)
		menuRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetReservationID(t *testing.T) {
	t.Run("successfully gets reservation ID", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		mockReservation := &repositories.Reservation{
			ID:      123,
			RoomID:  1,
			GuestID: "guest-1",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		reservationID, err := service.GetReservationID(context.Background(), int64(1))

		require.NoError(t, err)
		assert.Equal(t, "123", reservationID)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("returns empty when reservation not found", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, nil)

		reservationID, err := service.GetReservationID(context.Background(), int64(1))

		require.NoError(t, err)
		assert.Equal(t, "", reservationID)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		repoError := errors.New("reservation service error")
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, repoError)

		reservationID, err := service.GetReservationID(context.Background(), int64(1))

		assert.Error(t, err)
		assert.Equal(t, "", reservationID)
		assert.Equal(t, repoError, err)
		reservationRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetActiveReservation(t *testing.T) {
	t.Run("successfully gets active reservation with billing folder", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "billing-folder-abc",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		reservation, err := service.GetActiveReservation(context.Background(), int64(1))

		require.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, int64(123), reservation.ID)
		assert.Equal(t, "billing-folder-abc", reservation.BillingFolderID)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("returns error when no active reservation found", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, nil)

		reservation, err := service.GetActiveReservation(context.Background(), int64(1))

		assert.Error(t, err)
		assert.Nil(t, reservation)
		assert.Contains(t, err.Error(), "no active reservation found for room 1")
		reservationRepo.AssertExpectations(t)
	})

	t.Run("returns error when reservation has no billing folder ID", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "", // Empty billing folder ID
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		reservation, err := service.GetActiveReservation(context.Background(), int64(1))

		assert.Error(t, err)
		assert.Nil(t, reservation)
		assert.Contains(t, err.Error(), "reservation 123 has no billing folder ID")
		reservationRepo.AssertExpectations(t)
	})

	t.Run("returns error when service fails", func(t *testing.T) {
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(nil, nil, reservationRepo)

		repoError := errors.New("reservation service error")
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, repoError)

		reservation, err := service.GetActiveReservation(context.Background(), int64(1))

		assert.Error(t, err)
		assert.Nil(t, reservation)
		assert.Contains(t, err.Error(), "failed to get reservation for room 1")
		reservationRepo.AssertExpectations(t)
	})
}

func TestOrderService_CreateOrder(t *testing.T) {
	t.Run("successfully creates order with reservation and billing folder", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(orderRepo, menuRepo, reservationRepo)

		menuItem := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(10.00))
		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem, nil)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "billing-folder-xyz",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		result, err := service.CreateOrder(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Order)
		assert.NotNil(t, result.Reservation)
		assert.Equal(t, "1", result.Order.RoomID)
		assert.Equal(t, "123", result.Order.ReservationID)
		assert.Len(t, result.Order.Items, 1)
		assert.True(t, decimal.NewFromFloat(20.00).Equal(result.Order.TotalPrice))
		assert.Equal(t, entities.OrderStatusPending, result.Order.Status)
		assert.Equal(t, "billing-folder-xyz", result.Reservation.BillingFolderID)
		menuRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("fails when no active reservation found", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(orderRepo, menuRepo, reservationRepo)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		result, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no active reservation found")
		menuRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("fails when reservation service returns error", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(orderRepo, menuRepo, reservationRepo)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, errors.New("service unavailable"))

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		result, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get reservation")
		menuRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("fails when reservation has no billing folder ID", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		service := services.NewOrderService(orderRepo, menuRepo, reservationRepo)

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "", // Missing billing folder ID
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		result, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "has no billing folder ID")
		menuRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("fails validation with empty room ID", func(t *testing.T) {
		service := services.NewOrderService(nil, nil, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(0),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		result, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "room ID is required", err.Error())
	})

	t.Run("fails validation with no items", func(t *testing.T) {
		service := services.NewOrderService(nil, nil, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items:  []services.OrderItemRequest{},
		}

		result, err := service.CreateOrder(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "order must have at least one item", err.Error())
	})
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	t.Run("successfully updates order status", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		service := services.NewOrderService(orderRepo, nil, nil)

		order := testutil.CreateTestOrder("order-1", "1", "reservation-1", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)
		orderRepo.On("Update", mock.Anything, order).Return(nil)

		err := service.UpdateOrderStatus(context.Background(), "order-1", entities.OrderStatusPreparing)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusPreparing, order.Status)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when order not found", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		service := services.NewOrderService(orderRepo, nil, nil)

		orderRepo.On("FindByID", mock.Anything, "order-1").Return(nil, errors.New("not found"))

		err := service.UpdateOrderStatus(context.Background(), "order-1", entities.OrderStatusPreparing)

		assert.Error(t, err)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when order is nil", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		service := services.NewOrderService(orderRepo, nil, nil)

		orderRepo.On("FindByID", mock.Anything, "order-1").Return(nil, nil)

		err := service.UpdateOrderStatus(context.Background(), "order-1", entities.OrderStatusPreparing)

		assert.Error(t, err)
		assert.Equal(t, "order not found", err.Error())
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails with invalid status transition", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		service := services.NewOrderService(orderRepo, nil, nil)

		order := testutil.CreateTestOrder("order-1", "1", "reservation-1", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)

		err := service.UpdateOrderStatus(context.Background(), "order-1", entities.OrderStatusCompleted)

		assert.Error(t, err)
		assert.Equal(t, "pending order can only transition to preparing or cancelled", err.Error())
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when update repository fails", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		service := services.NewOrderService(orderRepo, nil, nil)

		order := testutil.CreateTestOrder("order-1", "1", "reservation-1", nil)
		repoError := errors.New("database error")
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)
		orderRepo.On("Update", mock.Anything, order).Return(repoError)

		err := service.UpdateOrderStatus(context.Background(), "order-1", entities.OrderStatusPreparing)

		assert.Error(t, err)
		assert.Equal(t, repoError, err)
		orderRepo.AssertExpectations(t)
	})
}
