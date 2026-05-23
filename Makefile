.PHONY: build test lint frontend-build docker-build clean run

# Build the Go binary
build: frontend-build
	go build -o sentinelmesh .

# Run all Go tests
test:
	go test -v -count=1 ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Build frontend
frontend-build:
	cd frontend && npm ci && npm run build

# Build Docker image
docker-build:
	docker build -t sentinelmesh:latest .

# Run locally
run: build
	./sentinelmesh

# Clean build artifacts
clean:
	rm -f sentinelmesh
	rm -rf frontend/dist
	rm -f sentinelmesh.db sentinelmesh.db-shm sentinelmesh.db-wal

# Install dev dependencies
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
