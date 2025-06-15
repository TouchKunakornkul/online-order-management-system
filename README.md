# Online Order Management System

A RESTful API built with Go for managing online orders with high concurrent processing capabilities.

## Project Structure

This project leverages a **Clean Architecture** structure that promotes separation of concerns, testability, and maintainability. The architecture is organized into distinct layers with clear dependencies flowing inward.

```
online-order-management-system/
├── cmd/                           # Application entry points
├── internal/                      # 🔒 Private application code (Clean Architecture)
│   ├── api/                       # 🌐 Delivery Layer
│   │   ├── http/handler/          # HTTP handlers and DTOs
│   │   └── validation/            # Business-specific validation logic
│   ├── domain/                    # 🏛️  Domain Layer
│   │   ├── entity/                # Business entities and rules
│   │   ├── errors/                # Domain-specific error definitions
│   │   └── repository/            # Repository interfaces
│   ├── infra/db/                  # 🔧 Infrastructure Layer - Database implementations
│   ├── middleware/                # 🛡️  Cross-cutting concerns - HTTP middleware
│   └── usecase/order/             # 💼 Use Case Layer - Business logic orchestration
├── pkg/                           # 📦 Domain-agnostic utilities (reusable)
│   ├── errors/                    # Generic error handling framework
│   ├── logger/                    # Structured JSON logging
│   ├── retryutil/                 # Retry utilities with exponential backoff
│   └── validation/                # Generic validation types and utilities
├── config/                        # ⚙️  Configuration management
├── test/                          # 🧪 Stress tests and benchmarks
├── docs/                          # 📚 Auto-generated Swagger documentation
├── docker-compose.yml             # 🐳 Database setup
├── schema.sql                     # 🗄️  Database schema
└── main.go                        # 🚀 Application entry point
```

## Quick Start

### Setup & Run

```bash
# 1. Clone and setup
git clone <repository-url>
cd online-order-management-system
go mod tidy

# 2. Create a .env file
cp env.example .env
# Or create manually with the following content:
```

**Create a .env file** with the following content:

```bash
# PostgreSQL Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DBNAME=orderdb
POSTGRES_SSLMODE=disable

# Connection Pool Settings
DB_MAX_OPEN_CONNS=300
DB_MAX_IDLE_CONNS=150
DB_CONN_MAX_LIFETIME=45m
DB_CONN_MAX_IDLE_TIME=20m
DB_PING_TIMEOUT=15s

# Server Configuration
PORT=8080
GIN_MODE=debug
```

Or view the complete sample in `env.example` file.

```bash
# 3. Start database
make db-up

# 4. Run server
make run
```

The server will start on `http://localhost:8080`

### Access Swagger

Open your browser and navigate to:

```
http://localhost:8080/swagger/index.html
```

### Load Test

```bash
# Stress test: 1,000 orders with 100 concurrent goroutines
make test-stress

# EXTREME test: 10,000 orders with 500 concurrent goroutines
make test-stress-extreme
```

Expected performance: 2,000+ orders/second with 100% success rate.

## Available Commands

```bash
# Development
make help               # Show all commands
make build              # Build application
make run                # Build and run server
make test               # Run tests

# Database
make db-up              # Start PostgreSQL database
make db-down            # Stop database
make db-reset           # Reset database

# Load Testing
make test-stress        # 1,000 orders stress test
make test-stress-extreme # 10,000 orders extreme test

# Documentation
make swagger-generate   # Generate Swagger docs
make swagger-regen      # Regenerate Swagger docs

# Cleanup
make clean              # Clean build artifacts
```

## API Endpoints

```
GET    /health                  # Health check
POST   /api/v1/orders           # Create order
GET    /api/v1/orders           # List orders (page-based pagination)
GET    /api/v1/orders/:id       # Get order by ID
PUT    /api/v1/orders/:id/status # Update order status
```

### Example Usage

**Create Order:**

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe",
    "items": [
      {
        "product_name": "Laptop",
        "quantity": 1,
        "unit_price": 999.99
      }
    ]
  }'
```

**List Orders:**

```bash
# Get first page
curl "http://localhost:8080/api/v1/orders?page=1&limit=10"
```

---

**Built with Clean Architecture • High Concurrency • PostgreSQL • Swagger Documentation**
