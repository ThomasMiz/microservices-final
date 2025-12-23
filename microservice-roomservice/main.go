package main

import (
	"database/sql"

	"microservice-roomservice/internal/interfaces/http/handlers"
	"microservice-roomservice/pkg/config"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupRouter(cfg *config.Config, healthHandler *handlers.HealthHandler) *gin.Engine {
	router := gin.Default()

	// Health checks at root level (outside BasePath for Kubernetes probes)
	router.GET("/health/live", healthHandler.LivenessProbe)
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	// Use the BasePath from environment variables
	basePath := cfg.Server.BasePath
	api := router.Group(basePath)
	{
		// Menu endpoints
		menu := api.Group("/menu")
		{
			menu.GET("", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Menu items endpoint"})
			})
		}

		// Order endpoints
		orders := api.Group("/orders")
		{
			orders.POST("", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Create order endpoint"})
			})
		}
	}

	return router
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize health handler (without database for simple demo)
	var db *sql.DB
	healthHandler := handlers.NewHealthHandler(db, nil)

	// Setup router
	r := setupRouter(cfg, healthHandler)
	r.Run(":" + cfg.Server.Port)
}
