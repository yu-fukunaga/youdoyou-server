.PHONY: build run seed check test clean setup

# Build all binaries
build:
	@echo "Building..."
	go build -o bin/server ./cmd/server
	go build -o bin/check ./cmd/check
	go build -o bin/seed ./cmd/seed

# Run the server
run:
	@echo "Running server..."
	go run ./cmd/server

# Run the seed tool
# Seed all: make seed
# Seed specific: make seed/basic
seed:
	@echo "Running seed for all files..."
	FIRESTORE_EMULATOR_HOST=localhost:8080 go run ./cmd/seed all

seed/%:
	@echo "Running seed for $*..."
	FIRESTORE_EMULATOR_HOST=localhost:8080 go run ./cmd/seed $*

# Run the check tool
check:
	@echo "Running check..."
	go run ./cmd/check

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin

# Setup dependencies
setup:
	@echo "Setting up..."
	go mod tidy
	go mod download

# Start Firebase emulators
emulators:
	@echo "Starting emulators..."
	cd firebase && firebase emulators:start
