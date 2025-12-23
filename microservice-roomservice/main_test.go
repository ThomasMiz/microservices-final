package main

import (
	"encoding/json"
	"microservice-roomservice/internal/interfaces/http/handlers"
	"microservice-roomservice/pkg/config"

	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	// Load test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     "8080",
			BasePath: "/api/roomservice",
		},
	}

	// Create health handler without database for testing
	healthHandler := handlers.NewHealthHandler(nil, nil)

	router := setupRouter(cfg, healthHandler)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedKey    string
	}{
		{
			name:           "Liveness probe returns 200",
			path:           "/health/live",
			expectedStatus: 200,
			expectedKey:    "alive",
		},
		{
			name:           "Readiness probe returns 200",
			path:           "/health/ready",
			expectedStatus: 200,
			expectedKey:    "ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedKey, response["status"])
		})
	}
}

func TestBasePathRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:     "8080",
			BasePath: "/api/roomservice",
		},
	}

	healthHandler := handlers.NewHealthHandler(nil, nil)
	router := setupRouter(cfg, healthHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/roomservice/menu", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
