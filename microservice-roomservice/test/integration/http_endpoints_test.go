package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories/mocks"
	"microservice-roomservice/internal/domain/services"
	httpHandlers "microservice-roomservice/internal/interfaces/http"
	"microservice-roomservice/internal/interfaces/http/handlers"
	"microservice-roomservice/internal/testutil"
	"microservice-roomservice/pkg/config"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupMenuRouter() (*mocks.MockMenuRepository, *httptest.Server) {
	menuRepo := new(mocks.MockMenuRepository)
	manageMenu := usecases.NewManageMenu(menuRepo)
	menuHandler := handlers.NewMenuHandler(manageMenu)

	// Create test config with empty base path for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     "8080",
			BasePath: "", // Empty base path for easier testing
		},
	}

	// Create health handler without database for testing
	healthHandler := handlers.NewHealthHandler(nil, nil)

	// Create minimal router for testing
	router := httpHandlers.SetupRouter(cfg, menuHandler, nil, healthHandler)
	server := httptest.NewServer(router)

	return menuRepo, server
}

func TestMenuEndpoints_Integration(t *testing.T) {
	t.Run("GET /menu returns all menu items", func(t *testing.T) {
		menuRepo, server := setupMenuRouter()
		defer server.Close()

		items := testutil.CreateTestMenuItems(3)
		menuRepo.On("FindAll", mock.Anything).Return(items, nil)

		resp, err := http.Get(server.URL + "/menu")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.MenuItem
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		menuRepo.AssertExpectations(t)
	})

	t.Run("GET /menu returns empty list", func(t *testing.T) {
		menuRepo, server := setupMenuRouter()
		defer server.Close()

		menuRepo.On("FindAll", mock.Anything).Return([]entities.MenuItem{}, nil)

		resp, err := http.Get(server.URL + "/menu")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.MenuItem
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Empty(t, result)
		menuRepo.AssertExpectations(t)
	})

	t.Run("POST /menu/items creates new menu item", func(t *testing.T) {
		menuRepo, server := setupMenuRouter()
		defer server.Close()

		menuRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(nil)

		reqBody := map[string]string{
			"description": "Burger",
			"price":       "12.99",
		}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/menu/items",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result entities.MenuItem
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "Burger", result.Description)
		assert.True(t, decimal.NewFromFloat(12.99).Equal(result.Price))
		menuRepo.AssertExpectations(t)
	})

	t.Run("POST /menu/items fails with invalid data", func(t *testing.T) {
		_, server := setupMenuRouter()
		defer server.Close()

		reqBody := map[string]string{
			"description": "",
			"price":       "12.99",
		}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/menu/items",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("POST /menu/items fails with invalid price format", func(t *testing.T) {
		_, server := setupMenuRouter()
		defer server.Close()

		reqBody := map[string]string{
			"description": "Burger",
			"price":       "invalid",
		}
		jsonBody, _ := json.Marshal(reqBody)

		resp, err := http.Post(
			server.URL+"/menu/items",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("PUT /menu/items/:id updates menu item", func(t *testing.T) {
		menuRepo, server := setupMenuRouter()
		defer server.Close()

		existingItem := testutil.CreateTestMenuItem("menu-1", "Old Name", decimal.NewFromFloat(10.00))
		menuRepo.On("FindByID", mock.Anything, "menu-1").Return(existingItem, nil)
		menuRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.MenuItem")).Return(nil)

		reqBody := map[string]string{
			"description": "New Name",
			"price":       "15.00",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest(
			http.MethodPut,
			server.URL+"/menu/items/menu-1",
			bytes.NewBuffer(jsonBody),
		)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result entities.MenuItem
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "New Name", result.Description)
		menuRepo.AssertExpectations(t)
	})

	t.Run("DELETE /menu/items/:id deletes menu item", func(t *testing.T) {
		menuRepo, server := setupMenuRouter()
		defer server.Close()

		menuRepo.On("Delete", mock.Anything, "menu-1").Return(nil)

		req, _ := http.NewRequest(
			http.MethodDelete,
			server.URL+"/menu/items/menu-1",
			nil,
		)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		menuRepo.AssertExpectations(t)
	})
}

func setupOrderRouter() (*mocks.MockOrderRepository, *mocks.MockMenuRepository, *mocks.MockReservationRepository, *httptest.Server) {
	orderRepo := new(mocks.MockOrderRepository)
	menuRepo := new(mocks.MockMenuRepository)
	reservationRepo := new(mocks.MockReservationRepository)

	orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)

	// Create a mock producer for testing
	createOrder := &usecases.CreateOrder{}
	getOrderStatus := usecases.NewGetOrderStatus(orderRepo)
	updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)

	orderHandler := handlers.NewOrderHandler(createOrder, getOrderStatus, updateOrderStatus)

	// Create test config with empty base path for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     "8080",
			BasePath: "", // Empty base path for easier testing
		},
	}

	// Create health handler without database for testing
	healthHandler := handlers.NewHealthHandler(nil, nil)

	router := httpHandlers.SetupRouter(cfg, nil, orderHandler, healthHandler)
	server := httptest.NewServer(router)

	return orderRepo, menuRepo, reservationRepo, server
}

func TestOrderEndpoints_Integration(t *testing.T) {
	t.Run("GET /orders/:id returns order", func(t *testing.T) {
		orderRepo, _, _, server := setupOrderRouter()
		defer server.Close()

		expectedOrder := testutil.CreateTestOrder("order-1", "ROOM-001", "reservation-1", nil)
		orderRepo.On("FindByID", mock.Anything, "order-1").Return(expectedOrder, nil)

		resp, err := http.Get(server.URL + "/orders/order-1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result entities.Order
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "order-1", result.ID)
		assert.Equal(t, "ROOM-001", result.RoomID)
		orderRepo.AssertExpectations(t)
	})

	t.Run("GET /orders/:id returns 404 when not found", func(t *testing.T) {
		orderRepo, _, _, server := setupOrderRouter()
		defer server.Close()

		orderRepo.On("FindByID", mock.Anything, "order-1").Return(nil, assert.AnError)

		resp, err := http.Get(server.URL + "/orders/order-1")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		orderRepo.AssertExpectations(t)
	})

	t.Run("GET /orders/room/:roomId returns orders", func(t *testing.T) {
		orderRepo, _, _, server := setupOrderRouter()
		defer server.Close()

		order1 := testutil.CreateTestOrder("order-1", "ROOM-001", "", nil)
		order2 := testutil.CreateTestOrder("order-2", "ROOM-001", "", nil)
		orders := []entities.Order{*order1, *order2}
		orderRepo.On("FindByRoomID", mock.Anything, "ROOM-001").Return(orders, nil)

		resp, err := http.Get(server.URL + "/orders/room/ROOM-001")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.Order
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "ROOM-001", result[0].RoomID)
		orderRepo.AssertExpectations(t)
	})

	t.Run("GET /orders/room/:roomId returns empty list", func(t *testing.T) {
		orderRepo, _, _, server := setupOrderRouter()
		defer server.Close()

		orderRepo.On("FindByRoomID", mock.Anything, "ROOM-001").Return([]entities.Order{}, nil)

		resp, err := http.Get(server.URL + "/orders/room/ROOM-001")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []entities.Order
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Empty(t, result)
		orderRepo.AssertExpectations(t)
	})
}

func TestHealthEndpoint_Integration(t *testing.T) {
	t.Run("GET /health/live returns alive status", func(t *testing.T) {
		// Create test config with empty base path for testing
		cfg := &config.Config{
			Server: config.ServerConfig{
				Port:     "8080",
				BasePath: "", // Empty base path for easier testing
			},
		}

		// Create health handler without database for testing
		healthHandler := handlers.NewHealthHandler(nil, nil)

		router := httpHandlers.SetupRouter(cfg, nil, nil, healthHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		resp, err := http.Get(server.URL + "/health/live")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "alive", result["status"])
	})

	t.Run("GET /health/ready returns ready status", func(t *testing.T) {
		// Create test config with empty base path for testing
		cfg := &config.Config{
			Server: config.ServerConfig{
				Port:     "8080",
				BasePath: "", // Empty base path for easier testing
			},
		}

		// Create health handler without database for testing
		healthHandler := handlers.NewHealthHandler(nil, nil)

		router := httpHandlers.SetupRouter(cfg, nil, nil, healthHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		resp, err := http.Get(server.URL + "/health/ready")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ready", result["status"])
	})
}
