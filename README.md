# Online Order Management System

A RESTful API built with Go for managing online orders.

## Quick Start

### 1. Prerequisites

- Go 1.22+
- Docker & Docker Compose

### 2. Setup & Run

```bash
# Clone and setup
git clone <repository-url>
cd online-order-management-system
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
POST   /api/v1/orders/bulk      # Bulk create orders
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

## Available Commands

```bash
make help       # Show all commands
make build      # Build application
make run        # Build and run server
make test       # Run tests
make test-api   # Test API endpoints
make db-up      # Start database
make db-down    # Stop database
make db-reset   # Reset database
make clean      # Clean build artifacts
```

## Environment Variables

```bash
DATABASE_URL=postgres://user:password@localhost/orderdb?sslmode=disable
PORT=8080
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
- **Use Case Layer**: Application business logic
- **Repository Layer**: Data access interfaces
- **Infrastructure Layer**: Database implementations
- **API Layer**: HTTP handlers
