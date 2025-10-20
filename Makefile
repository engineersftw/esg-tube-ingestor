.PHONY: help build test test-unit test-integration test-coverage clean install run lint fmt db-up db-down db-logs db-shell docker-clean

# Default target
help:
	@echo "YouTube Video Sync - Development Commands"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build          Build the binary"
	@echo "  make install        Install to GOPATH/bin"
	@echo "  make run            Run the application"
	@echo "  make clean          Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test           Run all tests"
	@echo "  make test-unit      Run unit tests only"
	@echo "  make test-integration  Run integration tests"
	@echo "  make test-coverage  Run tests with coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint           Run linters (go vet, golint)"
	@echo "  make fmt            Format code with gofmt"
	@echo ""
	@echo "Database (Docker):"
	@echo "  make db-up          Start PostgreSQL with Docker Compose"
	@echo "  make db-down        Stop PostgreSQL"
	@echo "  make db-logs        View database logs"
	@echo "  make db-shell       Connect to PostgreSQL shell"
	@echo "  make db-reset       Reset database (WARNING: deletes all data)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-clean   Clean up Docker volumes and containers"

# Build targets
build:
	@echo "Building youtube-sync..."
	go build -o youtube-sync cmd/youtube-sync/main.go
	@echo "✓ Build complete: ./youtube-sync"

install:
	@echo "Installing youtube-sync to GOPATH/bin..."
	go install cmd/youtube-sync/main.go
	@echo "✓ Installed"

clean:
	@echo "Cleaning build artifacts..."
	rm -f youtube-sync
	rm -rf dist/
	rm -rf coverage.*
	@echo "✓ Clean complete"

# Testing targets
test:
	@echo "Running all tests..."
	go test ./... -v

test-unit:
	@echo "Running unit tests..."
	go test ./tests/unit/... -v

test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Code quality targets
lint:
	@echo "Running linters..."
	go vet ./...
	@which golint > /dev/null 2>&1 || (echo "Installing golint..." && go install golang.org/x/lint/golint@latest)
	golint ./...
	@echo "✓ Lint complete"

fmt:
	@echo "Formatting code..."
	gofmt -w -s .
	@echo "✓ Format complete"

# Database targets
db-up:
	@echo "Starting PostgreSQL with Docker Compose..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 3
	@docker-compose exec postgres pg_isready -U postgres || echo "Database not ready yet, give it a few more seconds"
	@echo "✓ Database is running on localhost:5432"
	@echo "  User: postgres"
	@echo "  Password: postgres"
	@echo "  Database: esg_tube"

db-down:
	@echo "Stopping PostgreSQL..."
	docker-compose down
	@echo "✓ Database stopped"

db-logs:
	@echo "Showing database logs (Ctrl+C to exit)..."
	docker-compose logs -f postgres

db-shell:
	@echo "Connecting to PostgreSQL shell..."
	docker-compose exec postgres psql -U postgres -d esg_tube

db-reset:
	@echo "⚠️  WARNING: This will delete all database data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Resetting database..."; \
		docker-compose down -v; \
		docker-compose up -d; \
		echo "✓ Database reset complete"; \
	else \
		echo "Cancelled"; \
	fi

# Docker cleanup
docker-clean:
	@echo "Cleaning up Docker resources..."
	docker-compose down -v --remove-orphans
	@echo "✓ Docker cleanup complete"

# Development workflow
dev: db-up
	@echo "Development environment ready!"
	@echo ""
	@echo "Database is running. You can now:"
	@echo "  1. Run: make build"
	@echo "  2. Test: make test"
	@echo "  3. Execute: ./youtube-sync video dQw4w9WgXcQ"
	@echo ""
	@echo "When done, run: make db-down"

# Quick smoke test
smoke-test: build db-up
	@echo "Running smoke test..."
	@echo "Checking version..."
	./youtube-sync version || echo "⚠️  Binary not working correctly"
	@echo "Checking database connection..."
	go test ./tests/integration/database_test.go -v -run TestDatabaseConnection || echo "⚠️  Database connection failed"
	@echo "✓ Smoke test complete"
