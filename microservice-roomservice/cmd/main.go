package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/internal/infrastructure/http"
	"microservice-roomservice/internal/infrastructure/messaging/redis"
	"microservice-roomservice/internal/infrastructure/persistence/postgres"
	httpHandlers "microservice-roomservice/internal/interfaces/http"
	"microservice-roomservice/internal/interfaces/http/handlers"
	"microservice-roomservice/internal/interfaces/messaging"
	"microservice-roomservice/pkg/config"
	"microservice-roomservice/pkg/logger"
	"microservice-roomservice/pkg/tracing"

	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger
	logger.Init()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize OpenTelemetry tracer
	shutdown, err := tracing.InitTracer(tracing.TracerConfig{
		ServiceName:    cfg.Tracing.ServiceName,
		ServiceVersion: cfg.Tracing.ServiceVersion,
		Environment:    cfg.Tracing.Environment,
		OTLPEndpoint:   cfg.Tracing.OTLPEndpoint,
		Enabled:        cfg.Tracing.Enabled,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error().Err(err).Msg("Failed to shutdown tracer")
		}
	}()

	// Verify tracer provider is set correctly
	tp := otel.GetTracerProvider()
	logger.Debug().Msgf("[TRACE DEBUG] TracerProvider type: %T", tp)
	testTracer := tp.Tracer("test-tracer")
	logger.Debug().Msgf("[TRACE DEBUG] Test Tracer type: %T", testTracer)
	propagator := otel.GetTextMapPropagator()
	logger.Debug().Msgf("[TRACE DEBUG] TextMapPropagator type: %T", propagator)

	// Initialize database connection with OpenTelemetry instrumentation
	db, err := otelsql.Open("postgres", cfg.Database.GetDatabaseDSN(),
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize separate non-traced database connection for health checks
	var healthDB *sql.DB
	if cfg.Tracing.Enabled {
		// Create non-traced connection using standard sql.Open
		healthDB, err = sql.Open("postgres", cfg.Database.GetDatabaseDSN())
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect to health check database")
		}
		defer healthDB.Close()
	} else {
		// If tracing is disabled, use the same connection
		healthDB = db
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to ping database")
	}
	logger.Info().Msg("Successfully connected to database")

	// Run database migrations
	if err := postgres.RunMigrations(db); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Initialize repositories
	menuRepo := postgres.NewMenuRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	reservationClient := http.NewReservationClient(cfg.Reservation.BaseURL)
	reservationRepo := http.NewReservationRepositoryAdapter(reservationClient)
	billingClient := http.NewBillingClient(cfg.Billing.BaseURL)
	billingRepo := http.NewBillingRepositoryAdapter(billingClient)

	// Initialize domain services
	orderService := services.NewOrderService(orderRepo, menuRepo, reservationRepo)
	billingService := services.NewBillingService(billingRepo, &cfg.Billing)

	// Initialize Redis producer for Kitchen requests
	var producer *redis.Producer
	if cfg.Redis.Password != "" {
		producer = redis.NewProducer(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.KitchenRequestsStream)
		logger.Info().Msg("Redis producer initialized with authentication (kitchen-requests stream)")
	} else {
		producer = redis.NewProducer(cfg.Redis.Addr, "", cfg.Redis.KitchenRequestsStream)
		logger.Info().Msg("Redis producer initialized without authentication (kitchen-requests stream)")
	}
	defer producer.Close()

	// Initialize use cases
	manageMenu := usecases.NewManageMenu(menuRepo)
	createOrder := usecases.NewCreateOrder(orderService, billingService, orderRepo, producer)
	getOrderStatus := usecases.NewGetOrderStatus(orderRepo)
	updateOrderStatus := usecases.NewUpdateOrderStatus(orderService)

	// Initialize HTTP handlers
	menuHandler := handlers.NewMenuHandler(manageMenu)
	orderHandler := handlers.NewOrderHandler(createOrder, getOrderStatus, updateOrderStatus)
	healthHandler := handlers.NewHealthHandler(db, healthDB)

	// Setup router
	router := httpHandlers.SetupRouter(cfg, menuHandler, orderHandler, healthHandler)

	// Initialize Order Events consumer for Kitchen responses
	// This consumer listens for responses from Kitchen service (OrderAccepted, OrderRejected, etc.)
	orderEventHandler := messaging.NewOrderEventHandler(updateOrderStatus, billingService, orderRepo, producer)
	var orderEventConsumer *redis.OrderEventConsumer
	if cfg.Redis.Password != "" {
		orderEventConsumer = redis.NewOrderEventConsumer(
			cfg.Redis.Addr,
			cfg.Redis.Password,
			cfg.Redis.KitchenResponsesStream,
			cfg.Redis.ConsumerGroupID+"-responses",
			orderEventHandler,
		)
		logger.Info().Msg("Order Event consumer initialized with authentication (kitchen-responses stream)")
	} else {
		orderEventConsumer = redis.NewOrderEventConsumer(
			cfg.Redis.Addr,
			"",
			cfg.Redis.KitchenResponsesStream,
			cfg.Redis.ConsumerGroupID+"-responses",
			orderEventHandler,
		)
		logger.Info().Msg("Order Event consumer initialized without authentication (kitchen-responses stream)")
	}

	// Start Redis consumer in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Order Event consumer for saga orchestration
	go func() {
		logger.Info().Msg("Starting Order Event consumer (order-events)...")
		if err := orderEventConsumer.Start(ctx); err != nil {
			logger.Error().Err(err).Msg("Order Event consumer error")
		}
	}()

	// Start HTTP server
	logger.Info().Msgf("Starting HTTP server on port %s...", cfg.Server.Port)
	go func() {
		if err := router.Run(":" + cfg.Server.Port); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")
	cancel()

	// Give ongoing operations time to complete
	time.Sleep(2 * time.Second)

	// Shutdown tracer
	if err := shutdown(context.Background()); err != nil {
		logger.Error().Err(err).Msg("Error shutting down tracer")
	}

	logger.Info().Msg("Server shutdown complete")
}
