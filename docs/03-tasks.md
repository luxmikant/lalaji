# Deliverable Tasks — B2B Kirana Shipping Charge Estimator

> **Project:** Jambotails — B2B e-Commerce Shipping Module  
> **Tech Stack:** Go (Gin), PostgreSQL, Redis  
> **Date:** March 3, 2026  
> **Status:** Ready for Development

---

## Overview

All tasks are organized into **Phases** (logical milestones).  
Each task has:
- `ID` — unique task identifier
- `Priority` — Must Have / Good to Have
- `Effort` — Small (< 2h) / Medium (2–4h) / Large (4–8h)
- `Depends On` — prerequisite task IDs

Total estimated effort: **~60–70 hours** for a single developer.

---

## Phase 0 — Project Bootstrap

> Goal: A running Go HTTP server with proper project scaffolding.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-001 | Initialize Go module (`go mod init`) and project directory structure as defined in `02-design.md` | Must Have | Small | — |
| T-002 | Add dependencies: `gin`, `viper`, `zap`, `go-playground/validator`, `pq`, `go-redis`, `testify`, `golang-jwt` | Must Have | Small | T-001 |
| T-003 | Set up `config/config.go` using Viper — load from `.env`: DB URL, Redis URL, JWT secret, port, log level | Must Have | Small | T-001 |
| T-004 | Create `cmd/server/main.go` — bootstrap config, DB pool, Redis client, router, start server | Must Have | Small | T-002, T-003 |
| T-005 | Set up `docker-compose.yml` with services: `app`, `postgres:16`, `redis:7` with volume mounts and env vars | Must Have | Small | T-001 |
| T-006 | Create `Makefile` with targets: `make run`, `make test`, `make migrate`, `make lint` | Must Have | Small | T-001 |
| T-007 | Create `.env.example` documenting all required environment variables | Must Have | Small | T-003 |

**Phase 0 Exit Criteria:** `make run` starts the Go server on port 8080 without errors; Docker Compose brings up all services.

---

## Phase 1 — Database & Migrations

> Goal: All tables exist in PostgreSQL with correct constraints and seed data.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-010 | Set up `golang-migrate` integration — migration runner in `make migrate` | Must Have | Small | T-006 |
| T-011 | Write migration `000001_init_schema.up.sql` — create all tables: `customers`, `sellers`, `products`, `warehouses`, `transport_rates`, `delivery_speed_configs`, `orders`, `shipments` | Must Have | Medium | T-010 |
| T-012 | Write migration `000001_init_schema.down.sql` — drop all tables in reverse order | Must Have | Small | T-011 |
| T-013 | Write migration `000002_seed_data.up.sql` — seed: 2 warehouses, 3 transport rates, 2 speed configs, sample sellers, products, customers | Must Have | Small | T-011 |
| T-014 | Validate: run migrations up and down, verify all constraints (FK, UNIQUE, NOT NULL, CHECK for lat/lng) work correctly | Must Have | Small | T-013 |

**Phase 1 Exit Criteria:** All tables created, seeded, and verified via `psql` or DBeaver.

---

## Phase 2 — Models & Repository Layer

> Goal: All Go struct definitions and PostgreSQL queries implemented with interface abstractions.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-020 | Create `internal/models/` — Go structs for all 8 entities with correct field types, JSON tags, and `validate` tags | Must Have | Medium | T-011 |
| T-021 | Create `internal/repositories/interfaces.go` — define all repository interfaces: `WarehouseRepository`, `CustomerRepository`, `SellerRepository`, `ProductRepository`, `TransportRateRepository`, `DeliverySpeedConfigRepository` | Must Have | Small | T-020 |
| T-022 | Implement `internal/repositories/warehouse_repo.go` — `GetByID()`, `GetAllActive()` using `database/sql` + `pq` | Must Have | Small | T-021 |
| T-023 | Implement `internal/repositories/customer_repo.go` — `GetByID()` | Must Have | Small | T-021 |
| T-024 | Implement `internal/repositories/seller_repo.go` — `GetByID()` | Must Have | Small | T-021 |
| T-025 | Implement `internal/repositories/product_repo.go` — `GetByID()`, validate seller ownership | Must Have | Small | T-021 |
| T-026 | Implement `internal/repositories/rate_config_repo.go` — `GetTransportRates()`, `GetDeliverySpeedConfig(speed)` | Must Have | Small | T-021 |
| T-027 | Implement `internal/repositories/order_repo.go` + `shipment_repo.go` — `Create()` with transaction support | Good to Have | Medium | T-021 |

**Phase 2 Exit Criteria:** All repo functions connect to DB and return correct data against seeded records. Manual `go test ./internal/repositories/...` against a test DB passes.

---

## Phase 3 — Core Business Logic (Services)

> Goal: Geo engine, transport strategy, pricing strategy, warehouse service, and shipping service all implemented and unit-tested.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-030 | Implement `internal/services/geo/haversine.go` — `Distance(lat1, lng1, lat2, lng2 float64) float64` | Must Have | Small | — |
| T-031 | Write unit tests for `haversine.go` — test known city pairs (Bangalore→Mumbai = ~981 km), edge cases: same point (0km), antipodal points | Must Have | Small | T-030 |
| T-032 | Define `internal/services/transport/strategy.go` — `TransportStrategy` interface with `Name() string` and `RatePerKmPerKg() float64` | Must Have | Small | — |
| T-033 | Implement `minivan.go`, `truck.go`, `aeroplane.go` — each implementing `TransportStrategy` | Must Have | Small | T-032 |
| T-034 | Implement `internal/services/transport/factory.go` — `NewTransportStrategy(distanceKm float64, rates []TransportRate) TransportStrategy` | Must Have | Small | T-033 |
| T-035 | Write unit tests for transport factory — dist=50 → MiniVan, dist=200 → Truck, dist=600 → Aeroplane, boundary values: dist=99.99, dist=100, dist=499.99, dist=500 | Must Have | Small | T-034 |
| T-036 | Define `internal/services/pricing/strategy.go` — `PricingStrategy` interface with `Calculate(distKm, weightKg, ratePerKmPerKg float64) PricingBreakdown` | Must Have | Small | — |
| T-037 | Implement `standard.go` and `express.go` pricing strategies | Must Have | Small | T-036 |
| T-038 | Implement `internal/services/pricing/factory.go` — `NewPricingStrategy(speed string) (PricingStrategy, error)` | Must Have | Small | T-037 |
| T-039 | Write unit tests for pricing — verify Standard: base 10 + (2×320×10=6400), Express adds (1.2×10=12). Test with multiple scenarios including zero-weight edge case | Must Have | Small | T-038 |
| T-040 | Implement `internal/services/warehouse_service.go` — `FindNearest(ctx, sellerID, productID int64) (*NearestWarehouseResult, error)` — validates seller, product, fetches all warehouses, runs Haversine loop, returns nearest | Must Have | Medium | T-022, T-024, T-025, T-030 |
| T-041 | Write unit tests for `warehouse_service.go` using mock repositories (testify/mock) — test: nearest correctly identified, seller not found, no active warehouses | Must Have | Medium | T-040 |
| T-042 | Implement `internal/services/shipping_service.go` — `CalculateCharge(ctx, warehouseID, customerID, productID int64, speed string) (*ShippingChargeResult, error)` — fetch entities, compute distance, select transport strategy, compute price | Must Have | Medium | T-023, T-025, T-026, T-030, T-034, T-038 |
| T-043 | Implement `CalculateFull(ctx, sellerID, customerID, productID int64, speed string) (*FullCalculationResult, error)` — calls warehouse_service then shipping_service | Must Have | Small | T-040, T-042 |
| T-044 | Write unit tests for `shipping_service.go` — test: valid calc with mock repos, invalid speed enum, warehouse not found, customer not found, perishable product forcing express warning | Must Have | Medium | T-042, T-043 |

**Phase 3 Exit Criteria:** `go test ./internal/services/...` passes 100% with mocked repositories. All edge cases from requirements FR-03, FR-04, FR-05 covered.

---

## Phase 4 — Cache Layer

> Goal: Redis caching integrated into warehouse and shipping services.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-050 | Set up `internal/cache/redis_client.go` — initialize `go-redis` client, ping health check on startup | Good to Have | Small | T-002 |
| T-051 | Implement `internal/cache/cache_service.go` — `Get(key) ([]byte, bool)`, `Set(key, value, ttl)`, `Delete(key)` with JSON serialization | Good to Have | Small | T-050 |
| T-052 | Integrate cache in `warehouse_service.FindNearest()` — check `nearest_wh:{sellerID}` before DB call; set on miss with TTL 10min | Good to Have | Small | T-040, T-051 |
| T-053 | Integrate cache in `shipping_service.CalculateCharge()` — check before DB call; set on miss with TTL 5min | Good to Have | Small | T-042, T-051 |
| T-054 | Implement in-memory cache (sync.Map) for rate configs (`transport_rates`, `speed_configs`) loaded at server startup — refreshed every 60min via background goroutine | Good to Have | Small | T-026 |
| T-055 | Implement graceful Redis fallback — if Redis unavailable, log warning and serve from DB without surfacing error to client | Good to Have | Small | T-052, T-053 |

**Phase 4 Exit Criteria:** Repeated API calls for same inputs served from Redis (verify with Redis `MONITOR` or cache hit logs).

---

## Phase 5 — HTTP Handlers & Middleware

> Goal: All three APIs are wired up with full validation, middleware, and correct response envelopes.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-060 | Create `pkg/response/response.go` — `Success(c, data)`, `Error(c, statusCode, code, message)`, `ValidationError(c, fields)` helpers with standard envelope | Must Have | Small | — |
| T-061 | Create `pkg/errors/errors.go` — `AppError` struct: `Code string`, `Message string`, `HTTPStatusCode int` + common error vars (ErrNotFound, ErrValidation, ErrInternal) | Must Have | Small | — |
| T-062 | Create `pkg/validator/validator.go` — initialize `go-playground/validator`, register custom validators: `latitude` (-90 to 90), `longitude` (-180 to 180), `delivery_speed` (standard\|express) | Must Have | Small | — |
| T-063 | Implement `internal/middleware/request_id.go` — generate UUID v4 per request, inject into `X-Request-ID` header and Gin context | Must Have | Small | — |
| T-064 | Implement `internal/middleware/logger.go` — log method, path, status, latency, requestId using `zap` on every request | Must Have | Small | T-063 |
| T-065 | Implement `internal/middleware/recovery.go` — catch panics, log stack trace, return 500 with requestId | Must Have | Small | T-063 |
| T-066 | Implement `internal/middleware/auth.go` — validate `Authorization: Bearer <JWT>` header, reject with 401 if missing/invalid/expired | Must Have | Small | — |
| T-067 | Implement `internal/middleware/rate_limiter.go` — token bucket per IP, 100 req/min, return 429 on exceed | Good to Have | Medium | — |
| T-068 | Implement `internal/handlers/warehouse_handler.go` — parse & validate `sellerId`, `productId` from query params; call `warehouse_service.FindNearest()`; return response using envelope | Must Have | Medium | T-060, T-061, T-062, T-040 |
| T-069 | Implement `internal/handlers/shipping_handler.go` — `GetCharge()` and `Calculate()` handlers; parse params/body; validate enums; call shipping service; return response | Must Have | Medium | T-060, T-061, T-062, T-042, T-043 |
| T-070 | Implement `internal/handlers/health_handler.go` — `GET /health` returns DB ping + Redis ping status for container orchestration checks | Must Have | Small | — |
| T-071 | Wire all routes in `cmd/server/main.go` — apply middleware chain, register all route groups under `/api/v1/` | Must Have | Small | T-063, T-064, T-065, T-066, T-068, T-069, T-070 |

**Phase 5 Exit Criteria:** All three main APIs return correct responses for valid and invalid inputs. Postman/curl testing passes all cases from [01-requirements.md Error Handling table](01-requirements.md).

---

## Phase 6 — Integration Tests

> Goal: HTTP-level tests covering all API endpoints with real DB (test DB) and mocked external services.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-080 | Create `tests/integration/testhelpers.go` — spin up test Gin router, inject test DB connection, seed test data, tear down after each test | Good to Have | Medium | T-071 |
| T-081 | Write integration tests for `GET /api/v1/warehouse/nearest` — happy path, missing sellerId, invalid productId ownership, no active warehouses | Good to Have | Medium | T-080 |
| T-082 | Write integration tests for `GET /api/v1/shipping-charge` — happy path (standard + express), invalid speed, warehouse not found, customer not found | Good to Have | Medium | T-080 |
| T-083 | Write integration tests for `POST /api/v1/shipping-charge/calculate` — happy path, missing body fields, seller not found, invalid speed enum | Good to Have | Medium | T-080 |
| T-084 | Verify charge calculation correctness: use seeded data (known distances, known weights) and assert exact `shippingCharge` values against hand-computed expected outputs | Good to Have | Medium | T-081, T-082, T-083 |

**Phase 6 Exit Criteria:** `make test` runs all unit + integration tests and passes. Coverage report shows ≥ 70% on `internal/services/` and `internal/handlers/`.

---

## Phase 7 — Observability & Polish

> Goal: Production-readiness: structured logs, metrics, proper error messages everywhere.

| Task ID | Task | Priority | Effort | Depends On |
|---|---|---|---|---|
| T-090 | Set up `zap` logger in `config/` — production JSON logger in non-dev mode, dev console logger in dev mode | Good to Have | Small | T-003 |
| T-091 | Add Prometheus metrics endpoint `GET /metrics` — expose: request count by route, latency histogram, cache hit/miss counters | Good to Have | Medium | T-071 |
| T-092 | Add DB connection pool settings to config: `MaxOpenConns=25`, `MaxIdleConns=10`, `ConnMaxLifetime=5m` — prevents connection exhaustion at scale | Must Have | Small | T-004 |
| T-093 | Review and harden all error messages — no internal error details exposed to client in 500 responses; all 4xx messages are actionable to the API caller | Must Have | Small | T-069 |
| T-094 | Add `golangci-lint` config (`.golangci.yml`) and run linter — fix all lint errors: unused vars, error handling, shadow vars | Good to Have | Small | T-071 |
| T-095 | Final Dockerfile — multi-stage build: `golang:1.22-alpine` build stage + `alpine:3.19` runtime stage; non-root user | Must Have | Small | T-001 |
| T-096 | Write README.md at project root — setup instructions, env vars, how to run with Docker, how to run tests, API examples with curl | Must Have | Small | T-071 |

**Phase 7 Exit Criteria:** `docker compose up` starts all services; APIs return responses with structured logs visible; README is complete.

---

## Summary Table

| Phase | Goal | Tasks | Must Have | Good To Have | Est. Effort |
|---|---|---|---|---|---|
| Phase 0 | Project Bootstrap | T-001 to T-007 | All | — | 4h |
| Phase 1 | Database & Migrations | T-010 to T-014 | All | — | 5h |
| Phase 2 | Models & Repository | T-020 to T-027 | T-020 to T-026 | T-027 | 8h |
| Phase 3 | Core Business Logic | T-030 to T-044 | All | — | 18h |
| Phase 4 | Cache Layer | T-050 to T-055 | — | All | 5h |
| Phase 5 | Handlers & Middleware | T-060 to T-071 | Most | T-067 | 12h |
| Phase 6 | Integration Tests | T-080 to T-084 | — | All | 8h |
| Phase 7 | Observability & Polish | T-090 to T-096 | T-092–T-093, T-095–T-096 | Rest | 6h |
| **Total** | | **49 tasks** | **~35 tasks** | **~14 tasks** | **~66h** |

---

## Dependency Graph (Critical Path)

```
T-001 (init)
  ├── T-002 (deps) ──────────────────────────────────────┐
  ├── T-003 (config) → T-004 (main.go)                   │
  ├── T-005 (docker)                                      │
  ├── T-006 (makefile) → T-010 (migrate) → T-011 (schema)│
  │                            └── T-013 (seed)          │
  │                                                       │
  └── T-020 (models) ─────────────────────────────────── ┤
        └── T-021 (interfaces)                           │
              ├── T-022 (warehouse_repo)                 │
              ├── T-023 (customer_repo)                  │
              ├── T-024 (seller_repo)                    │
              ├── T-025 (product_repo)                   │
              └── T-026 (rate_config_repo)               │
                                                         │
T-030 (haversine) → T-031 (haversine tests)             │
T-032 (transport iface) → T-033 (impls) → T-034 (factory) → T-035 (tests)
T-036 (pricing iface) → T-037 (impls) → T-038 (factory) → T-039 (tests)

T-022 + T-024 + T-025 + T-030 → T-040 (warehouse_service) → T-041 (tests)
T-023 + T-025 + T-026 + T-030 + T-034 + T-038 → T-042 (shipping_service)
T-040 + T-042 → T-043 (calculate_full) → T-044 (tests)

T-040 + T-041 → T-060 + T-061 + T-062 → T-068 (warehouse_handler)
T-042 + T-043 → T-069 (shipping_handler)

T-068 + T-069 + T-063..T-066 → T-071 (wire routes)
T-071 → T-080 → T-081, T-082, T-083, T-084
```

---

## Development Order Recommendation

For a single developer, tackle in this linear order:

```
Phase 0 (all) → Phase 1 (all) → Phase 2 (T-020 to T-026) 
→ Phase 3 (T-030 to T-044) → Phase 5 (T-060 to T-071) 
→ Phase 7 (T-092, T-093, T-095, T-096)
→ [If time allows] Phase 4 (cache) → Phase 6 (tests) → Phase 7 (rest)
```

This order ensures the **core business logic is always testable** before adding layers on top.
