.PHONY: build run test lint clean docker-up docker-down migrate-up migrate-down \
        frontend-install frontend-dev frontend-build

# ── Variables ─────────────────────────────────────────────────
APP_NAME   := shipping-service
CMD_DIR    := ./cmd/server
DB_DSN     ?= postgres://jambotails:jambotails_secret@localhost:5432/shipping_db?sslmode=disable

# ── Build & Run ──────────────────────────────────────────────
build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

run: build
	./bin/$(APP_NAME)

# ── Test ─────────────────────────────────────────────────────
test:
	go test ./... -v -count=1

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ── Lint ─────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ── Database Migrations ──────────────────────────────────────
migrate-up:
	migrate -path migrations -database "$(DB_DSN)" up

migrate-down:
	migrate -path migrations -database "$(DB_DSN)" down

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# ── Docker ───────────────────────────────────────────────────
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

docker-logs:
	docker compose logs -f app

# ── Frontend ─────────────────────────────────────────────────
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

# ── Clean ────────────────────────────────────────────────────
clean:
	rm -rf bin/ coverage.out coverage.html
	go clean -cache
