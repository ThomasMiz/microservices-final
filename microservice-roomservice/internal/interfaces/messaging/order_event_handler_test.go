package messaging_test

import (
	"context"
	"errors"
	"testing"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/domain/services"
	messagingTypes "microservice-roomservice/internal/infrastructure/messaging"
	redisMocks "microservice-roomservice/internal/infrastructure/messaging/redis/mocks"
	"microservice-roomservice/internal/interfaces/messaging"
	"microservice-roomservice/internal/testutil"
	"microservice-roomservice/pkg/config"

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

func TestNewOrderEventHandler(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	menuRepo := new(mocks.MockMenuRepository)
	reservationRepo := new(mocks.MockReservationRepository)
	orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
	updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
	billingRepo := new(mocks.MockBillingRepository)
	billingService := services.NewBillingService(billingRepo, newTestBillingConfig())
	producer := new(redisMocks.MockProducer)

	handler := messaging.NewOrderEventHandler(updateOrderStatus, billingService, orderRepo, producer)

	assert.NotNil(t, handler)
}

func TestOrderEventHandler_HandleOrderCompletion(t *testing.T) {
	t.Run("successfully handles completed order", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		producer := new(redisMocks.MockProducer)

		// Create a preparing order (can transition to completed)
		order := testutil.CreateTestOrder("order-123", "101", "reservation-1", nil)
		_ = order.UpdateStatus(entities.OrderStatusPreparing) // Update to preparing status

		orderRepo.On("FindByID", mock.Anything, "order-123").Return(order, nil)
		orderRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, nil, orderRepo, producer)

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "completed",
		}

		err := handler.HandleOrderCompletion(context.Background(), message)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
	})

	t.Run("handles invalid status", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		producer := new(redisMocks.MockProducer)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, nil, orderRepo, producer)

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "invalid-status",
		}

		err := handler.HandleOrderCompletion(context.Background(), message)

		require.NoError(t, err)
		orderRepo.AssertNotCalled(t, "FindByID")
	})

	t.Run("cancels billing ticket when order is cancelled", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		billingRepo := new(mocks.MockBillingRepository)
		billingService := services.NewBillingService(billingRepo, newTestBillingConfig())
		producer := new(redisMocks.MockProducer)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, billingService, orderRepo, producer)

		// Create an order with billing info
		order := testutil.CreateTestOrder("order-123", "101", "reservation-1", nil)
		order.SetBillingInfo("folder-456", "ticket-789")

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "cancelled",
		}

		// Mock order lookup - once for billing cancellation, once for status update
		orderRepo.On("FindByID", mock.Anything, "order-123").Return(order, nil).Twice()
		orderRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
		// Only ticket is deleted, not the folder (folder belongs to reservation)
		billingRepo.On("DeleteTicket", mock.Anything, "ticket-789").Return(nil)

		err := handler.HandleOrderCompletion(context.Background(), message)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
		billingRepo.AssertExpectations(t)
		// Verify folder is NOT deleted (belongs to reservation)
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("continues even if billing ticket deletion fails", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		billingRepo := new(mocks.MockBillingRepository)
		billingService := services.NewBillingService(billingRepo, newTestBillingConfig())
		producer := new(redisMocks.MockProducer)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, billingService, orderRepo, producer)

		// Create an order with billing info
		order := testutil.CreateTestOrder("order-123", "101", "reservation-1", nil)
		order.SetBillingInfo("folder-456", "ticket-789")

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "cancelled",
		}

		// Mock order lookup - once for billing cancellation, once for status update
		orderRepo.On("FindByID", mock.Anything, "order-123").Return(order, nil).Twice()
		orderRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
		// Ticket deletion fails but order status update should still proceed
		billingRepo.On("DeleteTicket", mock.Anything, "ticket-789").Return(errors.New("billing service down"))

		err := handler.HandleOrderCompletion(context.Background(), message)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
		billingRepo.AssertExpectations(t)
		// Verify folder is NOT deleted (belongs to reservation)
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("handles order not found during billing cancellation", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		billingRepo := new(mocks.MockBillingRepository)
		billingService := services.NewBillingService(billingRepo, newTestBillingConfig())
		producer := new(redisMocks.MockProducer)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, billingService, orderRepo, producer)

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "cancelled",
		}

		// Mock order lookup for both billing cancellation and status update
		// In this scenario, the order is not found, so both billing cancellation and status update should fail
		orderRepo.On("FindByID", mock.Anything, "order-123").Return(nil, errors.New("order not found")).Twice()

		err := handler.HandleOrderCompletion(context.Background(), message)

		// We expect the handler to return an error because status update fails
		require.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
		orderRepo.AssertExpectations(t)
		billingRepo.AssertNotCalled(t, "DeleteTicket")
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("does not cancel billing for non-cancelled status", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		producer := new(redisMocks.MockProducer)

		// Mock order lookup for status update
		order := testutil.CreateTestOrder("order-123", "101", "reservation-1", nil)
		orderRepo.On("FindByID", mock.Anything, "order-123").Return(order, nil)
		orderRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, nil, orderRepo, producer)

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "cancelled",
		}

		err := handler.HandleOrderCompletion(context.Background(), message)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
	})

	t.Run("handles order with no billing info", func(t *testing.T) {
		orderRepo := new(mocks.MockOrderRepository)
		menuRepo := new(mocks.MockMenuRepository)
		reservationRepo := new(mocks.MockReservationRepository)
		orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
		updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)
		billingRepo := new(mocks.MockBillingRepository)
		billingService := services.NewBillingService(billingRepo, newTestBillingConfig())
		producer := new(redisMocks.MockProducer)

		handler := messaging.NewOrderEventHandler(updateOrderStatus, billingService, orderRepo, producer)

		// Create an order without billing info
		order := testutil.CreateTestOrder("order-123", "101", "reservation-1", nil)

		message := messagingTypes.OrderCompletionMessage{
			OrderID: "order-123",
			Status:  "cancelled",
		}

		// Mock order lookup for both billing cancellation and status update
		orderRepo.On("FindByID", mock.Anything, "order-123").Return(order, nil).Twice()
		orderRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		err := handler.HandleOrderCompletion(context.Background(), message)

		require.NoError(t, err)
		orderRepo.AssertExpectations(t)
		billingRepo.AssertNotCalled(t, "DeleteTicket")
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})
}
