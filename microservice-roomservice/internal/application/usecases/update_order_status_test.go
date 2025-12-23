package usecases_test

import (
	"context"
	"errors"
	"testing"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateOrderStatus(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	menuRepo := new(mocks.MockMenuRepository)
	reservationRepo := new(mocks.MockReservationRepository)
	orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
	uc := usecases.NewUpdateOrderStatus(orderService)

	assert.NotNil(t, uc)
}

func TestUpdateOrderStatus_Execute(t *testing.T) {
	t.Run("successfully updates order status from pending to preparing", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		order := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)
		orderRepo.On("Update", mock.Anything, order).Return(nil)

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatusPreparing)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusPreparing, order.Status)
		orderRepo.AssertExpectations(t)
	})

	t.Run("successfully updates order status from preparing to completed", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		order := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		_ = order.UpdateStatus(entities.OrderStatusPreparing)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)
		orderRepo.On("Update", mock.Anything, order).Return(nil)

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatusCompleted)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCompleted, order.Status)
		orderRepo.AssertExpectations(t)
	})

	t.Run("successfully cancels pending order", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		order := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)
		orderRepo.On("Update", mock.Anything, order).Return(nil)

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatusCancelled)

		require.NoError(t, err)
		assert.Equal(t, entities.OrderStatusCancelled, order.Status)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when order not found", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		orderRepo.On("FindByID", mock.Anything, "order-1").Return(nil, errors.New("not found"))

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatusPreparing)

		assert.Error(t, err)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails with invalid status transition", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		order := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatusCompleted)

		assert.Error(t, err)
		assert.Equal(t, "pending order can only transition to preparing or cancelled", err.Error())
		assert.Equal(t, entities.OrderStatusPending, order.Status)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails with invalid status", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		order := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatus("invalid"))

		assert.Error(t, err)
		assert.Equal(t, "invalid order status", err.Error())
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when update repository fails", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		uc := usecases.NewUpdateOrderStatus(orderService)

		order := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		repoError := errors.New("database error")
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(order, nil)
		orderRepo.On("Update", mock.Anything, order).Return(repoError)

		err := uc.Execute(context.Background(), "order-1", entities.OrderStatusPreparing)

		assert.Error(t, err)
		assert.Equal(t, repoError, err)
		orderRepo.AssertExpectations(t)
	})
}
