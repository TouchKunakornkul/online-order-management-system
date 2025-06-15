# Online Order Management System

A RESTful API built with Go for managing online orders with high concurrent processing capabilities.

## Quick Start

### 1. Prerequisites

- Go 1.22+
- Docker & Docker Compose

### 2. Setup & Run

```bash
# Setup
go mod tidy

# Start database
make db-up

# Run server
make run
```

### 3. Test API

```bash
make test-api
```

## API Endpoints

```
GET    /health                  # Health check
POST   /api/v1/orders           # Create order
GET    /api/v1/orders           # List orders (with pagination)
GET    /api/v1/orders/:id       # Get order by ID
PUT    /api/v1/orders/:id/status # Update order status
```

## Example Usage

### Create Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe",
    "customer_email": "john@example.com",
    "items": [
      {
        "product_name": "Laptop",
        "quantity": 1,
        "unit_price": 999.99
      }
    ]
  }'
```

### Get Order

```bash
curl http://localhost:8080/api/v1/orders/1
```

### List Orders

```bash
curl "http://localhost:8080/api/v1/orders?limit=10&cursor=2025-06-14T17:47:05Z_1"
```

### Update Order Status

```bash
curl -X PUT http://localhost:8080/api/v1/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "processing"}'
```

## Stress Testing & Performance

This system is designed to handle high concurrent order creation using goroutines and database transactions.

### Run Stress Tests

```bash
# Stress test - 1,000 orders with 100 concurrent goroutines
make test-stress

# EXTREME stress test - 10,000 orders with 500 concurrent goroutines
make test-stress-extreme

# Run all stress tests
make test-all-stress

# Run stress benchmark
make bench-stress
```

### Performance Expectations

**Stress Test (1,000 orders)**:

- **Success Rate**: ≥ 90%
- **Orders Per Second**: ≥ 5 OPS
- **Concurrent Goroutines**: 100
- **Average Latency**: Variable under stress

**EXTREME Stress Test (10,000 orders)**:

- **Success Rate**: ≥ 80% (acceptable under extreme load)
- **Orders Per Second**: Variable (performance analysis provided)
- **Concurrent Goroutines**: 500
- **Test Duration**: Up to 10 minutes

### Stress Test Configuration

The stress tests create large numbers of orders simultaneously using goroutines:

```go
// 1,000 orders stress test
config := StressTestConfig{
    BaseURL:        "http://localhost:8080",
    TotalOrders:    1000,         // Total orders to create
    MaxConcurrency: 100,          // Concurrent goroutines
    RequestTimeout: 30 * time.Second,
}

// 10,000 orders EXTREME stress test
config := StressTestConfig{
    BaseURL:        "http://localhost:8080",
    TotalOrders:    10000,        // Total orders to create
    MaxConcurrency: 500,          // Concurrent goroutines
    RequestTimeout: 60 * time.Second,
}
```

For detailed stress testing documentation, see [docs/CONCURRENT_TESTING.md](docs/CONCURRENT_TESTING.md).

## Available Commands

```bash
make help               # Show all commands
make build              # Build application
make run                # Build and run server
make test               # Run tests
make test-api           # Test API endpoints
make test-stress        # Stress test: 1,000 orders with 100 concurrent goroutines
make test-stress-extreme # EXTREME stress test: 10,000 orders with 500 goroutines
make test-all-stress    # Run all stress tests
make bench-stress       # Run stress benchmarks
make db-up              # Start database
make db-down            # Stop database
make db-reset           # Reset database
make clean              # Clean build artifacts
```

## Environment Configuration

The system uses environment variables for configuration. Copy `env.example` to `.env` or set these variables:

### Database Configuration

```bash
# PostgreSQL Database Configuration
POSTGRES_HOST=localhost      # Database host (default: localhost)
POSTGRES_PORT=5432          # Database port (default: 5432)
POSTGRES_USER=user          # Database user (default: user)
POSTGRES_PASSWORD=password  # Database password (default: password)
POSTGRES_DBNAME=orderdb     # Database name (default: orderdb)
POSTGRES_SSLMODE=disable    # SSL mode (default: disable)

# Connection Pool Settings (optimized for high concurrency)
DB_MAX_OPEN_CONNS=300        # Maximum open connections (default: 300)
DB_MAX_IDLE_CONNS=150        # Maximum idle connections (default: 150)
DB_CONN_MAX_LIFETIME=45m     # Connection max lifetime (default: 45m)
DB_CONN_MAX_IDLE_TIME=20m    # Connection max idle time (default: 20m)
DB_PING_TIMEOUT=15s          # Database ping timeout (default: 15s)
```

### Server Configuration

```bash
PORT=8080                    # Server port (default: 8080)
GIN_MODE=debug              # Gin mode: debug, release
```

### Environment Presets

**Development** (lower resource usage):

```bash
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
```

**Production** (high performance):

```bash
DB_MAX_OPEN_CONNS=300
DB_MAX_IDLE_CONNS=150
```

**Extreme Load** (for 10K+ concurrent orders):

```bash
DB_MAX_OPEN_CONNS=400
DB_MAX_IDLE_CONNS=200
```

## Database Schema

**Orders Table:**

- id, customer_name, customer_email, total_amount, status, created_at, updated_at

**Order Items Table:**

- id, order_id, product_name, quantity, unit_price, total_price

**Valid Order Statuses:**

- pending, processing, completed, cancelled

## Architecture

Clean Architecture with:

- **Domain Layer**: Business entities and rules
- **Use Case Layer**: Application business logic with goroutine support
- **Repository Layer**: Data access interfaces with transaction safety
- **Infrastructure Layer**: Database implementations with connection pooling
- **API Layer**: HTTP handlers with concurrent request processing
