.PHONY: build run seed check test clean setup lint secure semgrep secrets

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

# Run linter
lint:
	@echo "Running linter..."
	go tool golangci-lint run

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	go tool golangci-lint run --fix
	pre-commit run trailing-whitespace --all-files || true
	pre-commit run end-of-file-fixer --all-files || true

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin

# Security scan with semgrep
semgrep:
	@echo "Running Semgrep..."
	semgrep --config p/golang --config p/security-audit --config p/ci .

# Secret scan with gitleaks
secrets:
	@echo "Running gitleaks..."
	gitleaks detect --source .

# All security checks
secure: secrets semgrep
	@echo "Security checks completed!"

# Setup dependencies
setup:
	@echo "Setting up..."
	go mod tidy
	go mod download

# Start Firebase emulators
emulators:
	@echo "Starting emulators..."
	cd firebase && firebase emulators:start

# Deploy to Cloud Run
deploy:
	@echo "Deploying to Cloud Run..."
	./scripts/build_deploy.sh
