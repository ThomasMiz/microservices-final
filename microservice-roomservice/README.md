# Room Service Microservice


A Domain-Driven Design (DDD) microservice for managing hotel room service orders, built with Go, PostgreSQL, and Kafka.


## Architecture

This service follows clean architecture with DDD principles:

- **Domain Layer**: Core business logic and entities
- **Application Layer**: Use cases and application services
- **Infrastructure Layer**: External integrations (database, Kafka, HTTP clients)
- **Interface Layer**: HTTP endpoints and message handlers

## Features

- Menu management (CRUD operations)
- Order creation and tracking
- Integration with reservation service
- Kafka-based order processing pipeline
- PostgreSQL persistence with transactional support
- RESTful API endpoints

## Project Structure

```
microservice-roomservice/
├── cmd/
│   └── main.go                           # Application entry point
├── internal/
│   ├── domain/
│   │   ├── entities/                     # Domain entities (Menu, Order)
│   │   │   ├── menu.go
│   │   │   └── order.go
│   │   ├── repositories/                 # Repository interfaces
│   │   │   ├── menu_repository.go
│   │   │   ├── order_repository.go
│   │   │   └── reservation_repository.go
│   │   └── services/                     # Domain services
│   │       ├── menu_service.go
│   │       └── order_service.go
│   ├── application/
│   │   └── usecases/                     # Application use cases
│   │       ├── create_order.go
│   │       ├── get_order_status.go
│   │       ├── manage_menu.go
│   │       └── update_order_status.go
│   ├── infrastructure/
│   │   ├── persistence/postgres/         # PostgreSQL implementations
│   │   │   ├── migrations/schema.sql
│   │   │   ├── menu_repo.go
│   │   │   └── order_repo.go
│   │   ├── messaging/kafka/              # Kafka integration
│   │   │   ├── producer.go
│   │   │   └── consumer.go
│   │   └── http/                         # External HTTP clients
│   │       └── reservation_client.go
│   └── interfaces/
│       ├── http/                         # HTTP handlers and routing
│       │   ├── handlers/
│       │   │   ├── menu_handler.go
│       │   │   └── order_handler.go
│       │   └── router.go
│       └── messaging/                    # Kafka message handlers
│           └── order_event_handler.go
└── pkg/
    ├── config/                           # Configuration management
    │   └── config.go
    └── errors/                           # Error handling
        └── errors.go
```

## API Endpoints

### Menu Management

- `GET /menu` - Get all menu items
- `POST /menu/items` - Add a new menu item
- `PUT /menu/items/:id` - Update a menu item
- `DELETE /menu/items/:id` - Delete a menu item

### Order Management

- `POST /orders` - Create a new order
- `GET /orders/:id` - Get order by ID
- `GET /orders/room/:roomId` - Get all orders for a room

### Health Check

- `GET /health` - Service health check

## Configuration

The service is configured via environment variables:

### Server Configuration
- `PORT` - HTTP server port (default: 8080)

### Database Configuration
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_NAME` - Database name (default: roomservice)
- `DB_SSLMODE` - SSL mode (default: disable)

### Kafka Configuration
- `KAFKA_BROKERS` - Comma-separated list of Kafka brokers (default: localhost:9092)
- `KAFKA_ORDER_TOPIC` - Topic for publishing orders (default: kitchen-orders)
- `KAFKA_COMPLETION_TOPIC` - Topic for order completions (default: order-completions)
- `KAFKA_CONSUMER_GROUP` - Consumer group ID (default: roomservice-group)

### External Services
- `RESERVATION_SERVICE_URL` - Base URL for reservation service (default: http://localhost:8081)

## Database Setup

Initialize the database using the schema file:

```bash
psql -U postgres -d roomservice -f internal/infrastructure/persistence/postgres/migrations/schema.sql
```

## Running the Service

### Local Development

```bash
# Build the service
go build -o bin/roomservice cmd/main.go

# Run the service
./bin/roomservice
```

### Using Docker

```bash
# Build the Docker image
docker build -t roomservice:latest .

# Run the container
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e KAFKA_BROKERS=host.docker.internal:9092 \
  roomservice:latest
```

## Business Logic Flows

### Order Creation Flow

1. Receive order request with room ID and items
2. Validate all menu items exist
3. Fetch reservation ID from reservation service (optional)
4. Calculate total price
5. Create order with PENDING status
6. Save order to database
7. Publish order to Kafka for kitchen processing
8. Return order confirmation

### Order Status Update Flow

1. Kafka consumer listens for order completion messages
2. Update order status in database
3. Log successful status update

## Domain Entities

### MenuItem
- ID, Description, Price
- Business rules: Price must be > 0

### Order
- ID, RoomID, ReservationID, Items, TotalPrice, Status
- Status: pending → preparing → completed/cancelled
- Business rules: Status transition validation

### OrderItem
- ID, MenuItemID, Quantity, Price
- Calculates total based on quantity

## Dependencies

- **Gin**: HTTP web framework
- **PostgreSQL**: Database
- **Kafka**: Message broker (segmentio/kafka-go)
- **Decimal**: Precise decimal handling (shopspring/decimal)
- **UUID**: Unique ID generation (google/uuid)

## Testing

Run tests:
```bash
go test ./...
```

## Implementation Status

All phases of the implementation plan have been completed:

- ✅ Phase 1: Foundation setup
- ✅ Phase 2: Domain layer implementation
- ✅ Phase 3: Infrastructure layer
- ✅ Phase 4: Application layer
- ✅ Phase 5: Interface layer
- ✅ Phase 6: Configuration and deployment setup

## Next Steps

Potential enhancements (as per IMPLEMENTATION_PLAN.md):

1. Order cancellation support
2. Order modification for pending orders
3. Analytics dashboard
4. Inventory integration
5. Comprehensive test suite
6. Authentication and authorization
7. Distributed tracing
8. Metrics and monitoring

## License

[Your License Here]
