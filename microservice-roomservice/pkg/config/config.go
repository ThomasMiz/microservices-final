package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Reservation ReservationConfig
	Billing     BillingConfig
	Tracing     TracingConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port     string
	BasePath string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// RedisConfig holds Redis configuration for messaging
type RedisConfig struct {
	Addr                   string // Redis server address (host:port)
	Password               string
	KitchenRequestsStream  string // Stream for requests to Kitchen Service
	KitchenResponsesStream string // Stream for responses from Kitchen Service
	ConsumerGroupID        string
}

// ReservationConfig holds reservation service configuration
type ReservationConfig struct {
	BaseURL string
}

// BillingConfig holds billing service configuration
type BillingConfig struct {
	BaseURL   string
	ErrorRate float64 // Error simulation rate (0.0 = 0%, 1.0 = 100%)
}

// TracingConfig holds OpenTelemetry configuration
type TracingConfig struct {
	Enabled        bool
	OTLPEndpoint   string
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	// Parse billing error rate
	billingErrorRate, err := strconv.ParseFloat(getEnv("BILLING_ERROR_RATE", "0.0"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid BILLING_ERROR_RATE: %w", err)
	}
	if billingErrorRate < 0.0 || billingErrorRate > 1.0 {
		return nil, fmt.Errorf("BILLING_ERROR_RATE must be between 0.0 and 1.0")
	}

	// Parse tracing enabled flag
	tracingEnabled := getEnv("OTEL_ENABLED", "true") == "true"

	return &Config{
		Server: ServerConfig{
			Port:     getEnv("PORT", "8080"),
			BasePath: getEnv("BASE_PATH", "/api/roomservice"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "roomservice"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:                   getEnv("REDIS_ADDR", "localhost:6379"),
			Password:               getEnv("REDIS_PASSWORD", ""),
			KitchenRequestsStream:  getEnv("REDIS_KITCHEN_REQUESTS_STREAM", "kitchen-requests"),
			KitchenResponsesStream: getEnv("REDIS_KITCHEN_RESPONSES_STREAM", "kitchen-responses"),
			ConsumerGroupID:        getEnv("REDIS_CONSUMER_GROUP", "roomservice-group"),
		},
		Reservation: ReservationConfig{
			BaseURL: getEnv("RESERVATION_SERVICE_URL", "http://localhost:8081"),
		},
		Billing: BillingConfig{
			BaseURL:   getEnv("BILLING_SERVICE_URL", "http://localhost:8082"),
			ErrorRate: billingErrorRate,
		},
		Tracing: TracingConfig{
			Enabled:        tracingEnabled,
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
			ServiceName:    getEnv("OTEL_SERVICE_NAME", "microservice-roomservice"),
			ServiceVersion: getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
			Environment:    getEnv("OTEL_ENVIRONMENT", "development"),
		},
	}, nil
}

// GetDatabaseDSN returns the PostgreSQL DSN
func (c *DatabaseConfig) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
