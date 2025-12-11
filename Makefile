.PHONY: run build clean test dev db docker-up docker-down migrate sqlc tidy

# === Development ===

# Run app (uses .env, expects DB to be running - local or docker)
run:
	go run ./cmd/app

# Start docker DB + run app locally
dev: db run

# Start only the database (docker)
db:
	docker-compose up db -d

# === Build ===

# Dev build
build:
	go build -o bin/app ./cmd/app

# Production build (optimized, stripped)
build-prod:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/app ./cmd/app

# Cross-compile for Linux (for deploying from Mac/Windows)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/app ./cmd/app

clean:
	rm -rf bin/ coverage.out

# === Test ===

test:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# === Docker ===

# Full docker setup (app + db)
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

# Build production image (app only, no db)
docker-build:
	docker build -t go-bkd .

# === Database ===

# Run migrations (requires golang-migrate)
migrate-up:
	migrate -path db/migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path db/migrations -database "$${DATABASE_URL}" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

# Regenerate sqlc
sqlc:
	sqlc generate

# === Misc ===

tidy:
	go mod tidy

# === Project Init ===

# Initialize new project from this starter
# Usage: make init name=github.com/username/myproject
init:
	@if [ -z "$(name)" ]; then echo "Usage: make init name=github.com/username/project"; exit 1; fi
	@echo "Renaming project to $(name)..."
	@# Update go.mod
	@sed -i '' 's|github.com/vkrishna03/go-bkd-init|$(name)|g' go.mod
	@# Update all Go imports
	@find . -name "*.go" -type f -exec sed -i '' 's|github.com/vkrishna03/go-bkd-init|$(name)|g' {} +
	@# Update docker image name in Makefile
	@sed -i '' 's|go-bkd|$(shell basename $(name))|g' Makefile
	@echo "Done! Run 'go mod tidy' to verify."
