package usecases_test

import (
	"context"
	"errors"
	"testing"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
	repoMocks "microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/internal/infrastructure/messaging"
	redisMocks "microservice-roomservice/internal/infrastructure/messaging/redis/mocks"
	"microservice-roomservice/internal/testutil"
	"microservice-roomservice/pkg/config"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestBillingConfig() *config.BillingConfig {
	return &config.BillingConfig{
		BaseURL:   "http://test-billing:8080",
		ErrorRate: 0.0,
	}
}

func setupCreateOrderTest() (
	*repoMocks.MockOrderRepository,
	*repoMocks.MockMenuRepository,
	*repoMocks.MockReservationRepository,
	*repoMocks.MockBillingRepository,
	*redisMocks.MockProducer,
	*services.OrderService,
	*services.BillingService,
) {
	orderRepo := new(repoMocks.MockOrderRepository)
	menuRepo := new(repoMocks.MockMenuRepository)
	reservationRepo := new(repoMocks.MockReservationRepository)
	billingRepo := new(repoMocks.MockBillingRepository)
	producer := new(redisMocks.MockProducer)

	orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
	billingService := services.NewBillingService(billingRepo, newTestBillingConfig())

	return orderRepo, menuRepo, reservationRepo, billingRepo, producer, orderService, billingService
}

func TestNewCreateOrder(t *testing.T) {
	orderRepo, menuRepo, reservationRepo, billingRepo, producer, _, _ := setupCreateOrderTest()

	orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
	billingService := services.NewBillingService(billingRepo, newTestBillingConfig())

	uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

	assert.NotNil(t, uc)
}

func TestCreateOrder_Execute(t *testing.T) {
	t.Run("successfully creates order with all dependencies", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, _, producer, orderService, billingService := setupCreateOrderTest()

		// Setup menu item
		menuItem := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(10.00))
		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem, nil)

		// Setup reservation with billing folder ID
		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "billing-folder-abc",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		// Setup order save
		orderRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.Order")).Return(nil)

		// Setup Redis publish for OrderRequested event (saga orchestration)
		producer.On("PublishOrderEvent", mock.Anything, mock.AnythingOfType("messaging.OrderEvent")).Return(nil)

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "1", order.RoomID)
		assert.Equal(t, "123", order.ReservationID)
		assert.Equal(t, entities.OrderStatusPendingKitchen, order.Status)

		menuRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
		orderRepo.AssertExpectations(t)
		producer.AssertExpectations(t)
	})

	t.Run("fails when no active reservation exists", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, billingRepo, producer, orderService, billingService := setupCreateOrderTest()

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(nil, nil)

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "no active reservation found")

		// Verify billing was never called
		billingRepo.AssertNotCalled(t, "CreateTicket")
		// Verify order was never saved
		orderRepo.AssertNotCalled(t, "Save")
		// Verify Redis producer was never called
		producer.AssertNotCalled(t, "PublishOrderEvent")
	})

	t.Run("fails when reservation has no billing folder ID", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, billingRepo, producer, orderService, billingService := setupCreateOrderTest()

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "", // Empty!
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "has no billing folder ID")

		billingRepo.AssertNotCalled(t, "CreateTicket")
		orderRepo.AssertNotCalled(t, "Save")
		producer.AssertNotCalled(t, "PublishOrderEvent")
	})

	t.Run("succeeds even when Redis publish fails", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, _, producer, orderService, billingService := setupCreateOrderTest()

		menuItem := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(10.00))
		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem, nil)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "billing-folder-abc",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		orderRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.Order")).Return(nil)

		// Redis fails - but order should still succeed
		producer.On("PublishOrderEvent", mock.Anything, mock.Anything).Return(errors.New("redis unavailable"))

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		// Order should succeed despite Redis failure
		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "1", order.RoomID)
		assert.Equal(t, entities.OrderStatusPendingKitchen, order.Status)

		// All expectations should be met
		producer.AssertExpectations(t)
	})

	t.Run("fails validation with invalid room ID", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, billingRepo, producer, orderService, billingService := setupCreateOrderTest()

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(0), // Invalid
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "room ID is required")

		menuRepo.AssertNotCalled(t, "ExistsByID")
		reservationRepo.AssertNotCalled(t, "GetActiveReservationForRoom")
		billingRepo.AssertNotCalled(t, "CreateTicket")
		orderRepo.AssertNotCalled(t, "Save")
	})

	t.Run("fails validation with no items", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, billingRepo, producer, orderService, billingService := setupCreateOrderTest()

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items:  []services.OrderItemRequest{}, // Empty
		}

		order, err := uc.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "order must have at least one item")

		menuRepo.AssertNotCalled(t, "ExistsByID")
		reservationRepo.AssertNotCalled(t, "GetActiveReservationForRoom")
		billingRepo.AssertNotCalled(t, "CreateTicket")
		orderRepo.AssertNotCalled(t, "Save")
	})

	t.Run("fails when menu item does not exist", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, billingRepo, producer, orderService, billingService := setupCreateOrderTest()

		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(false, nil)

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "menu item not found")

		reservationRepo.AssertNotCalled(t, "GetActiveReservationForRoom")
		billingRepo.AssertNotCalled(t, "CreateTicket")
		orderRepo.AssertNotCalled(t, "Save")
	})

	t.Run("works with nil producer", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, billingRepo, _, orderService, billingService := setupCreateOrderTest()

		menuItem := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(10.00))
		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem, nil)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "billing-folder-abc",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		expectedTicket := &repositories.BillingTicket{
			ID:       "ticket-123",
			FolderID: "billing-folder-abc",
			Total:    decimal.NewFromFloat(20.00),
			ItemID:   "ROOMSERVICE-test",
			State:    repositories.BillingTicketStatePending,
		}
		billingRepo.On("CreateTicket", mock.Anything, mock.Anything).Return(expectedTicket, nil)

		orderRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.Order")).Return(nil)

		// Create use case with nil producer
		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, nil)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		// Should succeed without Redis producer
		require.NoError(t, err)
		assert.NotNil(t, order)
	})

	t.Run("correctly includes all order items in Redis message", func(t *testing.T) {
		orderRepo, menuRepo, reservationRepo, _, producer, orderService, billingService := setupCreateOrderTest()

		menuItem1 := testutil.CreateTestMenuItem("menu-1", "Burger", decimal.NewFromFloat(10.00))
		menuItem2 := testutil.CreateTestMenuItem("menu-2", "Pizza", decimal.NewFromFloat(15.00))
		menuRepo.On("ExistsByID", mock.Anything, "menu-1").Return(true, nil)
		menuRepo.On("ExistsByID", mock.Anything, "menu-2").Return(true, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(menuItem1, nil)
		menuRepo.On("FindByID", mock.Anything, "menu-2").Return(menuItem2, nil)

		mockReservation := &repositories.Reservation{
			ID:              123,
			RoomID:          1,
			GuestID:         "guest-1",
			BillingFolderID: "billing-folder-abc",
		}
		reservationRepo.On("GetActiveReservationForRoom", mock.Anything, int64(1)).Return(mockReservation, nil)

		orderRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.Order")).Return(nil)

		// Capture Redis event to verify items
		var capturedEvent messaging.OrderEvent
		producer.On("PublishOrderEvent", mock.Anything, mock.AnythingOfType("messaging.OrderEvent")).Run(func(args mock.Arguments) {
			capturedEvent = args.Get(1).(messaging.OrderEvent)
		}).Return(nil)

		uc := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)

		req := services.CreateOrderRequest{
			RoomID: int64(1),
			Items: []services.OrderItemRequest{
				{MenuItemID: "menu-1", Quantity: 2},
				{MenuItemID: "menu-2", Quantity: 1},
			},
		}

		order, err := uc.Execute(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, order)

		// Verify Redis event contains all items
		assert.Equal(t, "OrderRequested", capturedEvent.EventType)
		assert.Len(t, capturedEvent.Items, 2)
		assert.Equal(t, "1", capturedEvent.RoomID)
		assert.Equal(t, "123", capturedEvent.ReservationID)
		assert.Equal(t, "billing-folder-abc", capturedEvent.BillingFolderID)
	})
}
