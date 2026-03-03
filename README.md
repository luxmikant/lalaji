# Jambotails — B2B Shipping Charge Estimator

A Go REST API that estimates shipping charges for Kirana (small retail) stores in a B2B e-commerce context. Given a seller, product, customer, and delivery speed, it finds the nearest warehouse and calculates the total shipping charge using the **Haversine formula**, a **Strategy Pattern** for transport modes and pricing, and **DB-configurable** rates.

---

## Quick Start

### Option A: Docker Compose (recommended)

```bash
docker compose up --build -d    # starts Postgres, Redis, migrations, API
curl http://localhost:8080/health
```

### Option B: Local (requires Postgres 16 + Redis 7 running)

```bash
cp .env.example .env            # edit DB/Redis credentials
make migrate-up                 # run migrations
make run                        # build & start on :8080
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | DB + Redis health check |
| `GET` | `/api/v1/warehouse/nearest?sellerId=1&productId=1` | Find nearest warehouse to seller |
| `GET` | `/api/v1/shipping-charge?warehouseId=1&productId=1&customerId=1&deliverySpeed=standard` | Compute shipping charge |
| `POST` | `/api/v1/shipping-charge/calculate` | End-to-end: nearest warehouse + charge |

### POST body example

```json
{
  "sellerId": 1,
  "productId": 1,
  "customerId": 1,
  "deliverySpeed": "standard"
}
```

### Response example

```json
{
  "success": true,
  "data": {
    "shippingCharge": 620.24,
    "breakdown": {
      "distanceKm": 305.12,
      "transportMode": "truck",
      "ratePerKmPerKg": 2.0,
      "billableWeightKg": 10.0,
      "baseCourierCharge": 10.0,
      "distanceCharge": 610.24,
      "expressCharge": 0,
      "totalCharge": 620.24
    },
    "nearestWarehouse": {
      "warehouseId": 1,
      "warehouseName": "Bangalore Central WH",
      "warehouseLocation": { "lat": 12.9716, "lng": 77.5946 },
      "distanceKm": 0
    }
  },
  "requestId": "abc-123"
}
```

---

## Project Structure

```
cmd/server/main.go         → HTTP server bootstrap & graceful shutdown
config/                    → Viper-based config from .env
internal/
  models/                  → Domain entities (Customer, Seller, Product, Warehouse, etc.)
  repositories/            → DB access via interfaces + Postgres implementations
  services/                → Business logic (WarehouseService, ShippingService)
    geo/                   → Haversine distance calculation
    transport/             → Transport strategy pattern (MiniVan/Truck/Aeroplane)
    pricing/               → Pricing strategy pattern (Standard/Express)
  handlers/                → HTTP handlers (warehouse, shipping, health)
  middleware/              → RequestID, Logger, Recovery, Auth, RateLimiter
  cache/                   → Redis client + CacheService with graceful degradation
pkg/
  errors/                  → AppError type with factory helpers
  response/                → Standard JSON response envelope
  validator/               → Custom validation (delivery speed, error formatting)
migrations/                → SQL up/down migration files + seed data
```

---

## Architecture Highlights

- **Strategy Pattern** — transport mode (MiniVan/Truck/Aeroplane) selected by distance range; pricing (Standard/Express) selected by speed
- **Repository Pattern** — all DB access via interfaces for easy mocking in tests
- **Haversine Formula** — great-circle distance between lat/lng coordinates
- **Billable Weight** — `max(actualWeight, volumetricWeight)` where volumetric = `L × W × H / 5000`
- **DB-configurable Rates** — transport rates, delivery speed configs, base charges all stored in DB
- **Graceful Degradation** — if Redis is down, API continues using DB directly

---

## Running Tests

```bash
make test                 # all tests, verbose
make test-cover           # with HTML coverage report
```

39 unit tests covering: Haversine, transport strategy selection (boundary conditions), pricing calculations, billable weight, and full service-level tests with mock repositories.

---

## Documentation

| # | Document | Purpose |
|---|----------|---------|
| 1 | [Requirements Mapping](docs/01-requirements.md) | FRs, NFRs, business rules, validation matrix |
| 2 | [Design Document](docs/02-design.md) | Entity design, architecture, data flow, API contracts |
| 3 | [Task Breakdown](docs/03-tasks.md) | Phased deliverables, effort estimates, dependency graph |

---

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile binary to `bin/` |
| `make run` | Build + run server |
| `make test` | Run all unit tests |
| `make test-cover` | Tests + HTML coverage |
| `make lint` | golangci-lint |
| `make migrate-up` | Apply DB migrations |
| `make migrate-down` | Rollback migrations |
| `make docker-up` | Start full stack via Docker Compose |
| `make docker-down` | Tear down Docker stack |
