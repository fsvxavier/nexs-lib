# Makefile for PostgreSQL module

.PHONY: mocks test test-unit test-integration coverage clean

# Generate mocks for all interfaces
mocks:
	@echo "Generating mocks..."
	go generate ./db/postgresql/providers/pgx/mocks/
	@echo "Mocks generated successfully!"

# Run all tests
test: test-unit test-integration

# Run unit tests with mocks
test-unit:
	@echo "Running unit tests..."
	go test -v -tags=unit -timeout=30s ./db/postgresql/...

# Run integration tests (requires database)
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration -timeout=30s ./db/postgresql/...

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	go test -v -tags=unit -coverprofile=coverage.out ./db/postgresql/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f coverage.out coverage.html
	find . -name "*_mock.go" -delete
	@echo "Clean completed!"

# Install required tools
tools:
	@echo "Installing required tools..."
	go install github.com/golang/mock/mockgen@latest
	@echo "Tools installed successfully!"
