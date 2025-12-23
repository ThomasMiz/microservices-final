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

func TestNewBillingService(t *testing.T) {
	billingRepo := new(mocks.MockBillingRepository)

	service := services.NewBillingService(billingRepo, newTestBillingConfig())

	assert.NotNil(t, service)
}

func TestBillingService_CreateBillingForOrder(t *testing.T) {
	t.Run("successfully creates billing ticket in existing folder", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		billingFolderID := "folder-123"

		expectedTicket := &repositories.BillingTicket{
			ID:          "ticket-456",
			FolderID:    billingFolderID,
			Total:       order.TotalPrice,
			ItemID:      "ROOMSERVICE-order-1",
			Description: "Room service order for room 101 with 0 items",
			State:       repositories.BillingTicketStatePending,
			CreatedAt:   "2025-01-01T00:00:00Z",
		}
		billingRepo.On("CreateTicket", mock.Anything, mock.MatchedBy(func(req repositories.CreateTicketRequest) bool {
			return req.FolderID == billingFolderID && req.ItemID == "ROOMSERVICE-order-1"
		})).Return(expectedTicket, nil)

		err := service.CreateBillingForOrder(context.Background(), order, billingFolderID)

		require.NoError(t, err)
		assert.Equal(t, billingFolderID, order.BillingFolderID)
		assert.Equal(t, "ticket-456", order.BillingTicketID)
		billingRepo.AssertExpectations(t)
		// Verify no folder creation was attempted
		billingRepo.AssertNotCalled(t, "CreateFolder")
	})

	t.Run("returns error when order is nil", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		err := service.CreateBillingForOrder(context.Background(), nil, "folder-123")

		assert.Error(t, err)
		assert.Equal(t, "order cannot be nil", err.Error())
	})

	t.Run("returns error when billing folder ID is empty", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)

		err := service.CreateBillingForOrder(context.Background(), order, "")

		assert.Error(t, err)
		assert.Equal(t, "billing folder ID is required", err.Error())
	})

	t.Run("returns error when ticket creation fails", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		billingFolderID := "folder-123"

		billingRepo.On("CreateTicket", mock.Anything, mock.Anything).Return(nil, errors.New("billing service unavailable"))

		err := service.CreateBillingForOrder(context.Background(), order, billingFolderID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create billing ticket")
		billingRepo.AssertExpectations(t)
		// Verify no folder operations were attempted
		billingRepo.AssertNotCalled(t, "CreateFolder")
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("successfully creates billing for order with items", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		items := []entities.OrderItem{
			{ID: "item-1", MenuItemID: "menu-1", Quantity: 2, Price: decimal.NewFromFloat(10.00)},
			{ID: "item-2", MenuItemID: "menu-2", Quantity: 1, Price: decimal.NewFromFloat(15.00)},
		}
		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", items)
		billingFolderID := "reservation-folder-789"

		expectedTicket := &repositories.BillingTicket{
			ID:       "ticket-456",
			FolderID: billingFolderID,
			Total:    decimal.NewFromFloat(35.00),
			ItemID:   "ROOMSERVICE-order-1",
			State:    repositories.BillingTicketStatePending,
		}
		billingRepo.On("CreateTicket", mock.Anything, mock.MatchedBy(func(req repositories.CreateTicketRequest) bool {
			return req.FolderID == billingFolderID &&
				req.Total.Equal(decimal.NewFromFloat(35.00)) &&
				req.Description == "Room service order for room 101 with 2 items"
		})).Return(expectedTicket, nil)

		err := service.CreateBillingForOrder(context.Background(), order, billingFolderID)

		require.NoError(t, err)
		assert.True(t, order.HasBillingInfo())
		assert.Equal(t, billingFolderID, order.BillingFolderID)
		assert.Equal(t, "ticket-456", order.BillingTicketID)
		billingRepo.AssertExpectations(t)
	})

	t.Run("uses correct folder ID from reservation", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		reservationBillingFolderID := "res-billing-folder-abc"

		expectedTicket := &repositories.BillingTicket{
			ID:       "ticket-789",
			FolderID: reservationBillingFolderID,
			Total:    order.TotalPrice,
			ItemID:   "ROOMSERVICE-order-1",
			State:    repositories.BillingTicketStatePending,
		}
		billingRepo.On("CreateTicket", mock.Anything, mock.MatchedBy(func(req repositories.CreateTicketRequest) bool {
			return req.FolderID == reservationBillingFolderID
		})).Return(expectedTicket, nil)

		err := service.CreateBillingForOrder(context.Background(), order, reservationBillingFolderID)

		require.NoError(t, err)
		// Verify the folder ID stored in order is from reservation, not auto-generated
		assert.Equal(t, reservationBillingFolderID, order.BillingFolderID)
		billingRepo.AssertExpectations(t)
	})
}

func TestBillingService_DeleteBillingForOrder(t *testing.T) {
	t.Run("successfully deletes billing ticket for order", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		order.SetBillingInfo("folder-123", "ticket-456")

		billingRepo.On("DeleteTicket", mock.Anything, "ticket-456").Return(nil)

		service.DeleteBillingForOrder(context.Background(), order)

		billingRepo.AssertExpectations(t)
		// Verify folder is NOT deleted (it belongs to the reservation)
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("does nothing when order is nil", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		service.DeleteBillingForOrder(context.Background(), nil)

		billingRepo.AssertNotCalled(t, "DeleteTicket")
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("does nothing when order has no billing info", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)

		service.DeleteBillingForOrder(context.Background(), order)

		billingRepo.AssertNotCalled(t, "DeleteTicket")
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("logs error but continues when ticket deletion fails", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		order.SetBillingInfo("folder-123", "ticket-456")

		billingRepo.On("DeleteTicket", mock.Anything, "ticket-456").Return(errors.New("ticket not found"))

		// Should not panic, just log the error
		service.DeleteBillingForOrder(context.Background(), order)

		billingRepo.AssertExpectations(t)
		// Folder should NOT be deleted (belongs to reservation)
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("only deletes ticket not folder since folder belongs to reservation", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		service := services.NewBillingService(billingRepo, newTestBillingConfig())

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		order.SetBillingInfo("reservation-folder-999", "ticket-456")

		billingRepo.On("DeleteTicket", mock.Anything, "ticket-456").Return(nil)

		service.DeleteBillingForOrder(context.Background(), order)

		// Verify only ticket is deleted, folder is preserved (belongs to reservation)
		billingRepo.AssertExpectations(t)
		billingRepo.AssertNotCalled(t, "DeleteFolder")
	})

	t.Run("simulates billing error when error rate is configured", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		billingConfig := &config.BillingConfig{
			BaseURL:   "http://test-billing:8080",
			ErrorRate: 1.0, // 100% error rate for deterministic test
		}
		service := services.NewBillingService(billingRepo, billingConfig)

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		billingFolderID := "folder-123"

		err := service.CreateBillingForOrder(context.Background(), order, billingFolderID)

		// Should return simulated error without calling the repository
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated billing ticket creation failure")
		assert.Contains(t, err.Error(), "error rate: 1.00")
		billingRepo.AssertNotCalled(t, "CreateTicket")
	})

	t.Run("does not simulate error when error rate is 0", func(t *testing.T) {
		billingRepo := new(mocks.MockBillingRepository)
		billingConfig := &config.BillingConfig{
			BaseURL:   "http://test-billing:8080",
			ErrorRate: 0.0, // No error simulation
		}
		service := services.NewBillingService(billingRepo, billingConfig)

		order := testutil.CreateTestOrder("order-1", "101", "reservation-1", nil)
		billingFolderID := "folder-123"

		expectedTicket := &repositories.BillingTicket{
			ID:          "ticket-456",
			FolderID:    billingFolderID,
			Total:       order.TotalPrice,
			ItemID:      "ROOMSERVICE-order-1",
			Description: "Room service order for room 101 with 0 items",
			State:       repositories.BillingTicketStatePending,
			CreatedAt:   "2025-01-01T00:00:00Z",
		}
		billingRepo.On("CreateTicket", mock.Anything, mock.MatchedBy(func(req repositories.CreateTicketRequest) bool {
			return req.FolderID == billingFolderID && req.ItemID == "ROOMSERVICE-order-1"
		})).Return(expectedTicket, nil)

		err := service.CreateBillingForOrder(context.Background(), order, billingFolderID)

		assert.NoError(t, err)
		assert.Equal(t, billingFolderID, order.BillingFolderID)
		assert.Equal(t, "ticket-456", order.BillingTicketID)
		billingRepo.AssertExpectations(t)
	})
}
