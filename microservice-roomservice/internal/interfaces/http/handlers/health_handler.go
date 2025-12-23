package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db         *sql.DB
	untracedDB *sql.DB
}

func NewHealthHandler(db *sql.DB, untracedDB *sql.DB) *HealthHandler {
	return &HealthHandler{
		db:         db,
		untracedDB: untracedDB,
	}
}

// LivenessProbe indicates if the application is alive
// This should always return 200 as long as the application is running
func (h *HealthHandler) LivenessProbe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessProbe indicates if the application is ready to serve traffic
// This checks critical dependencies like database connectivity
func (h *HealthHandler) ReadinessProbe(c *gin.Context) {
	// Check database connectivity using non-traced connection
	dbToCheck := h.untracedDB
	if dbToCheck == nil {
		dbToCheck = h.db // fallback to traced connection if untraced not available
	}

	if dbToCheck != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := dbToCheck.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"error":     "database connection failed",
				"details":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"database": "ok",
		},
	})
}
