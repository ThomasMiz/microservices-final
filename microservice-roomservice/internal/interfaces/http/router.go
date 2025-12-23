package http

import (
	"microservice-roomservice/internal/interfaces/http/handlers"
	"microservice-roomservice/pkg/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// SetupRouter sets up the HTTP router
func SetupRouter(
	cfg *config.Config,
	menuHandler *handlers.MenuHandler,
	orderHandler *handlers.OrderHandler,
	healthHandler *handlers.HealthHandler,
) *gin.Engine {
	router := gin.Default()

	// Health checks at root level (outside BasePath for Kubernetes probes)
	router.GET("/health/live", healthHandler.LivenessProbe)
	router.GET("/health/ready", healthHandler.ReadinessProbe)

	// Use the BasePath from environment variables
	basePath := cfg.Server.BasePath
	api := router.Group(basePath)

	// Add OpenTelemetry middleware only to API group (not health checks)
	if cfg.Tracing.Enabled {
		api.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}

	{
		// Menu endpoints
		menu := api.Group("/menu")
		{
			menu.GET("", menuHandler.GetAllMenuItems)
			menu.POST("/items", menuHandler.AddMenuItem)
			menu.PUT("/items/:id", menuHandler.UpdateMenuItem)
			menu.DELETE("/items/:id", menuHandler.DeleteMenuItem)
		}

		// Order endpoints
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.GET("/room/:roomId", orderHandler.GetOrdersByRoom)
		}
	}

	return router
}
