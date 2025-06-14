.PHONY: help build run test clean db-up db-down db-reset

# Default target
help:
	@echo "🚀 Online Order Management System"
	@echo "================================="
	@echo "Available commands:"
	@echo "  make build     - Build the Go application"
	@echo "  make run       - Run the application"
	@echo "  make test      - Run tests"
	@echo "  make test-api  - Test API endpoints (requires running server)"
	@echo "  make db-up     - Start PostgreSQL database"
	@echo "  make db-down   - Stop PostgreSQL database"
	@echo "  make db-reset  - Reset database (stop, remove, start)"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make dev       - Start development environment (db + server)"

# Build the application
build:
	@echo "🔨 Building application..."
	go mod tidy
	go build -o bin/server main.go

# Run the application
run: build
	@echo "🚀 Starting server..."
	./bin/server

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Test API endpoints
test-api:
	@echo "🔍 Testing API endpoints..."
	@if ! curl -s http://localhost:8080/health > /dev/null; then \
		echo "❌ Server is not running. Please start the server first with 'make run'"; \
		exit 1; \
	fi
	./test_api.sh

# Start PostgreSQL database
db-up:
	@echo "🐘 Starting PostgreSQL database..."
	docker compose up -d postgres
	@echo "⏳ Waiting for database to be ready..."
	@until docker compose exec postgres pg_isready -U user -d orderdb; do \
		echo "Waiting for database..."; \
		sleep 2; \
	done
	@echo "✅ Database is ready!"

# Stop PostgreSQL database
db-down:
	@echo "🛑 Stopping PostgreSQL database..."
	docker compose down

# Reset database
db-reset: db-down
	@echo "🔄 Resetting database..."
	docker compose down -v
	$(MAKE) db-up

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Development environment
dev: db-up
	@echo "🚀 Starting development environment..."
	@echo "Database is running, now start the server with 'make run' in another terminal"
	@echo "Or run 'make run' to start the server now"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "🔍 Linting code..."
	golangci-lint run

# Run all checks
check: fmt lint test
	@echo "✅ All checks passed!" 