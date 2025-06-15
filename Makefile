.PHONY: help build run test clean db-up db-down db-reset test-stress test-stress-extreme

# Default target
help:
	@echo "🚀 Online Order Management System"
	@echo "================================="
	@echo "Available commands:"
	@echo "  make build          - Build the Go application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make test-api       - Test API endpoints (requires running server)"
	@echo "  make test-stress    - Stress test: 1,000 orders with 100 concurrent goroutines"
	@echo "  make test-stress-extreme - EXTREME stress test: 10,000 orders with 500 goroutines"
	@echo "  make test-debug     - Debug test: Check if order creation works properly"
	@echo "  make db-up          - Start PostgreSQL database"
	@echo "  make db-down        - Stop PostgreSQL database"
	@echo "  make db-reset       - Reset database (stop, remove, start)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make dev            - Start development environment (db + server)"

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



# Run stress benchmark
bench-stress:
	@echo "⚡ Running stress benchmark..."
	@if ! curl -s http://localhost:8080/health > /dev/null; then \
		echo "❌ Server is not running. Please start the server first with 'make run'"; \
		exit 1; \
	fi
	go test -bench=BenchmarkStressTest_OrderCreation ./test/ -benchtime=30s

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
	rm -f server.log
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

# Stress test - 1,000 orders with high concurrency
test-stress: build db-up
	@echo "🔥 Starting stress test with 1,000 orders..."
	@echo "📊 This will create 1,000 orders using 100 concurrent goroutines"
	@echo "⏳ Starting server in background..."
	@./bin/server > server.log 2>&1 & \
	SERVER_PID=$$!; \
	echo "Server PID: $$SERVER_PID"; \
	echo "⏳ Waiting for server to be ready..."; \
	for i in {1..30}; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Server is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ Server failed to start within 30 seconds"; \
			kill $$SERVER_PID 2>/dev/null || true; \
			exit 1; \
		fi; \
		sleep 1; \
	done; \
	echo "🔥 Running stress test: 1,000 orders with 100 concurrent goroutines..."; \
	go clean -testcache; \
	if go test -v ./test/ -run TestStressTest_1000Orders; then \
		echo "✅ Stress test completed successfully!"; \
		RESULT=0; \
	else \
		echo "❌ Stress test failed!"; \
		RESULT=1; \
	fi; \
	echo "🧹 Cleaning up..."; \
	kill $$SERVER_PID 2>/dev/null || true; \
	sleep 2; \
	rm -f server.log; \
	exit $$RESULT

# EXTREME stress test - 10,000 orders with very high concurrency
test-stress-extreme: build db-up
	@echo "🚨 Starting EXTREME stress test with 10,000 orders..."
	@echo "⚠️  WARNING: This test will create 10,000 orders using 500 concurrent goroutines"
	@echo "⚠️  This may take several minutes and significantly stress your system"
	@echo "⏳ Starting server in background..."
	@./bin/server > server.log 2>&1 & \
	SERVER_PID=$$!; \
	echo "Server PID: $$SERVER_PID"; \
	echo "⏳ Waiting for server to be ready..."; \
	for i in {1..30}; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Server is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ Server failed to start within 30 seconds"; \
			kill $$SERVER_PID 2>/dev/null || true; \
			exit 1; \
		fi; \
		sleep 1; \
	done; \
	echo "🚨 Running EXTREME stress test: 10,000 orders with 500 concurrent goroutines..."; \
	go clean -testcache; \
	if go test -v ./test/ -run TestStressTest_10000Orders -timeout=15m; then \
		echo "✅ EXTREME stress test completed!"; \
		RESULT=0; \
	else \
		echo "❌ EXTREME stress test failed!"; \
		RESULT=1; \
	fi; \
	echo "🧹 Cleaning up..."; \
	kill $$SERVER_PID 2>/dev/null || true; \
	sleep 2; \
	rm -f server.log; \
	exit $$RESULT

# Debug test - Check if order creation works properly
test-debug: build db-up
	@echo "🔍 Starting debug test to check order creation..."
	@echo "⏳ Starting server in background..."
	@./bin/server > server.log 2>&1 & \
	SERVER_PID=$$!; \
	echo "Server PID: $$SERVER_PID"; \
	echo "⏳ Waiting for server to be ready..."; \
	for i in {1..30}; do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✅ Server is ready!"; \
			break; \
		fi; \
		if [ $$i -eq 30 ]; then \
			echo "❌ Server failed to start within 30 seconds"; \
			kill $$SERVER_PID 2>/dev/null || true; \
			exit 1; \
		fi; \
		sleep 1; \
	done; \
	echo "🔍 Running debug tests..."; \
	if go test -v ./test/ -run TestDebugStressTest; then \
		echo "✅ Debug tests completed successfully!"; \
		RESULT=0; \
	else \
		echo "❌ Debug tests failed!"; \
		RESULT=1; \
	fi; \
	echo "🧹 Cleaning up..."; \
	kill $$SERVER_PID 2>/dev/null || true; \
	sleep 2; \
	rm -f server.log; \
	exit $$RESULT

# Run comprehensive stress testing
test-all-stress: test-stress test-stress-extreme
	@echo "🎯 All stress tests completed!" 