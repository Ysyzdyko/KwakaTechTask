.PHONY: build build-api build-worker run test clean docker-build docker-up docker-down docker-logs

# Build both API and Worker
build: build-api build-worker

# Build API
build-api:
	@echo "Building API..."
	@go build -o bin/api ./cmd/api

# Build Worker
build-worker:
	@echo "Building Worker..."
	@go build -o bin/worker ./cmd/worker

# Run API locally (requires MongoDB and RabbitMQ)
run-api: build-api
	@echo "Running API..."
	@./bin/api

# Run Worker locally (requires MongoDB and RabbitMQ)
run-worker: build-worker
	@echo "Running Worker..."
	@./bin/worker

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

# Docker commands
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

docker-up:
	@echo "Starting services..."
	@docker-compose up -d

docker-down:
	@echo "Stopping services..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

docker-restart:
	@docker-compose restart

# Setup: create credentials directory
setup:
	@echo "Creating credentials directory..."
	@mkdir -p credentials
	@echo "Please place your Google Sheets credentials.json in the credentials/ directory"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy



