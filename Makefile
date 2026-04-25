.PHONY: setup dev web-dev build-web build run docker-build docker-up

# First-time setup: install all dependencies
setup:
	go mod tidy
	cd web && npm install

# Run Go backend only (for local dev alongside `make web-dev`)
dev:
	go run .

# Run React dev server with hot-reload (proxies /api to :8080)
web-dev:
	cd web && npm run dev

# Build React frontend into web/dist/
build-web:
	cd web && npm run build

# Full build: React then Go (single binary with embedded frontend)
build: build-web
	CGO_ENABLED=0 go build -o feedfarmer .

# Build & run the binary locally
run: build
	./feedfarmer

# Build Docker image
docker-build:
	docker build -t feedfarmer/feedfarmer:latest .

# Start via Docker Compose (recommended for testing)
docker-up:
	docker compose up --build
