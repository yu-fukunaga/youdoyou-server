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

# Create GitHub tag and release (auto-increment version)
# Usage: make ghtag        -> auto-increment patch (v1.0.0 -> v1.0.1)
# Usage: make ghtag/minor  -> auto-increment minor (v1.0.0 -> v1.1.0)
# Usage: make ghtag/major  -> auto-increment major (v1.0.0 -> v2.0.0)
ghtag:
	@./scripts/create_release.sh patch

ghtag/minor:
	@./scripts/create_release.sh minor

ghtag/major:
	@./scripts/create_release.sh major

# Deploy a release to Cloud Run
# Usage: make release           -> deploy latest tag
# Usage: make release/v1.0.0    -> deploy specific tag
release:
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null); \
	if [ -z "$$LATEST_TAG" ]; then \
		echo "Error: No tags found in repository"; \
		echo "Please create a release first with: make ghtag"; \
		exit 1; \
	fi; \
	echo "Deploying latest tag: $$LATEST_TAG"; \
	./scripts/deploy_release.sh $$LATEST_TAG

release/%:
	@./scripts/deploy_release.sh $*
