package usecases_test

import (
	"context"
	"errors"
	"testing"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetOrderStatus(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	uc := usecases.NewGetOrderStatus(orderRepo)

	assert.NotNil(t, uc)
}

func TestGetOrderStatus_ByID(t *testing.T) {
	t.Run("successfully gets order by ID", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		expectedOrder := testutil.CreateTestOrder("order-1", "ROOM-001", "reservation-1", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(expectedOrder, nil)

		order, err := uc.ByID(context.Background(), "order-1")

		require.NoError(t, err)
		assert.Equal(t, expectedOrder, order)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when order not found", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		orderRepo.On("FindByID", mock.Anything, "order-1").Return(nil, errors.New("not found"))

		order, err := uc.ByID(context.Background(), "order-1")

		assert.Error(t, err)
		assert.Nil(t, order)
		orderRepo.AssertExpectations(t)
	})
}

func TestGetOrderStatus_ByRoomID(t *testing.T) {
	t.Run("successfully gets orders by room ID", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		order1 := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		order2 := testutil.CreateTestOrder("order-2", "ROOM-001", "", nil)
		expectedOrders := []entities.Order{*order1, *order2}
		orderRepo.On("FindByRoomID", mock.Anything, "ROOM-001").Return(expectedOrders, nil)

		orders, err := uc.ByRoomID(context.Background(), "ROOM-001")

		require.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, "order-1", orders[0].ID)
		assert.Equal(t, "order-2", orders[1].ID)
		orderRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no orders found", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		orderRepo.On("FindByRoomID", mock.Anything, "ROOM-001").Return([]entities.Order{}, nil)

		orders, err := uc.ByRoomID(context.Background(), "ROOM-001")

		require.NoError(t, err)
		assert.Empty(t, orders)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		repoError := errors.New("database error")
		orderRepo.On("FindByRoomID", mock.Anything, "ROOM-001").Return(nil, repoError)

		orders, err := uc.ByRoomID(context.Background(), "ROOM-001")

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, repoError, err)
		orderRepo.AssertExpectations(t)
	})
}

func TestGetOrderStatus_ByReservationID(t *testing.T) {
	t.Run("successfully gets orders by reservation ID", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		order1 := testutil.CreateTestOrder("order-1", "ROOM-001", "reservation-1", nil)
		order2 := testutil.CreateTestOrder("order-2", "ROOM-002", "reservation-1", nil)
		expectedOrders := []entities.Order{*order1, *order2}
		orderRepo.On("FindByReservationID", mock.Anything, "reservation-1").Return(expectedOrders, nil)

		orders, err := uc.ByReservationID(context.Background(), "reservation-1")

		require.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, "reservation-1", orders[0].ReservationID)
		assert.Equal(t, "reservation-1", orders[1].ReservationID)
		orderRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no orders found", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		orderRepo.On("FindByReservationID", mock.Anything, "reservation-1").Return([]entities.Order{}, nil)

		orders, err := uc.ByReservationID(context.Background(), "reservation-1")

		require.NoError(t, err)
		assert.Empty(t, orders)
		orderRepo.AssertExpectations(t)
	})

	t.Run("fails when repository returns error", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		uc := usecases.NewGetOrderStatus(orderRepo)

		repoError := errors.New("database error")
		orderRepo.On("FindByReservationID", mock.Anything, "reservation-1").Return(nil, repoError)

		orders, err := uc.ByReservationID(context.Background(), "reservation-1")

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, repoError, err)
		orderRepo.AssertExpectations(t)
	})
}
