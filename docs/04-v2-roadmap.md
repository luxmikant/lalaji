# Lalaji v2 — Production-Grade B2B Kirana Platform

**Status:** Designed (not implemented) | **Target Timeline:** Phase 2 of project  
**Created:** March 3, 2026 | **Version:** 2.0-Planning

---

## Overview

The v1 MVP proved the core shipping charge estimation engine with three APIs and a basic demo frontend. V2 transforms this into a real, full-stack B2B procurement platform targeting Kirana shop owners. Key new capabilities:

- **User Identity & RBAC** — buyers, sellers, admins with role-based API access
- **Living Product Catalogue** — sellers upload products with images to S3; search/browse/compare
- **Order Lifecycle Management** — cart → checkout → shipping → tracking → ratings
- **Inventory Saver** — quick-reorder from saved shopping lists (time-saving for repetitive Kirana orders)
- **PostGIS-Powered Mapping** — spatial queries for warehouse coverage + real-time buyer locator
- **Admin Control Panel** — RBAC-protected dashboard for platform management
- **Hardened Error Handling** — typed domain errors, graceful degradation, invalid/missing param validation
- **Strategic Caching** — expanded Redis coverage (15+ cache keys), invalidation on mutations
- **Modern Frontend** — creative design inspired by Notion + Meesho, not Amazon/Flipkart; Tailwind + Framer Motion + Leaflet

**Preserves all v1 design patterns:** Repository, Strategy (Transport/Pricing), Factory, Clean Architecture, DI, Config-Driven Rules, Middleware Chain.

---

## Architecture & Design Decisions

### Stack
- **Backend:** Go 1.x · Gin · PostgreSQL + **PostGIS** · Redis 7
- **Frontend:** Next.js 14 · React 18 · TypeScript · Tailwind CSS 3 · Leaflet + OpenStreetMap
- **Storage:** Cloudflare R2 (S3-compatible, no egress fees for India CDN)
- **Auth:** JWT (access: 15 min) + Refresh Tokens (7 days, rotated in DB for revocability)
- **State Mgmt (Client):** Zustand (auth, cart), TanStack Query (server state, caching, refetch)

### Design Patterns

**Preserved from v1:**
- **Repository Pattern** — all DB access behind interfaces; concrete impls injected; enables mocking in tests
- **Strategy Pattern (Transport/Pricing)** — modes selected by factory functions based on distance and config
- **Factory Pattern** — `transport/factory.go`, `pricing/factory.go`, now adding `storage/factory.go` for S3 vs. local disk fallback
- **Clean Architecture** — Handler → Service → Repository; no layer leakage
- **Dependency Injection** — constructors receive interface arguments
- **Config-Driven Rules** — all rates, speeds, warehouse coverage stored in DB; no hardcoding
- **Middleware Chain** — composed: request_id → logger → recovery → auth → role_guard (new) → rate_limiter

**New Patterns:**
- **Observer Pattern** — order status changes emit events to `order_status_events` table; decouples order lifecycle from notifications (future integrations: email, SMS, WhatsApp via webhook consumers)
- **Builder Pattern** — dynamic SQL query construction for admin filters (e.g., `OrderQueryBuilder.WithStatus(...).WithCustomer(...).Build()`)
- **Decorator Pattern** — cache decorator wraps repository methods; reads from cache, falls through to DB on miss, invalidates on mutations
- **Command Pattern** — `PlaceOrderCommand` struct encapsulates full order placement logic (atomicity via DB transaction)

### Data Model Philosophy

**Inspiration:** Real Kirana procurement needs + e-commerce best practices
- **Products:** Added brand, GST codes, shelf life, barcode (EAN-13), unit of measure (kg/litre/piece/dozen), tags
- **Sellers:** Added operating hours, avg dispatch time, cancellation rate, payment mode acceptance, logo/banner
- **Customers:** Added FSSAI license (food safety), shop area, monthly order value (for credit limits)
- **Orders:** Added invoice numbers, coupons, GST calculation, PDF invoice storage, buyer/seller ratings, dispute tracking
- **Warehouses:** Temperature zones (for dairy/ice cream), serviceable pincodes, operating hours

---

## Database Schema Extension

### New Migrations (500 lines total)

#### `000003_enable_postgis.up.sql` (50 lines)
```sql
CREATE EXTENSION IF NOT EXISTS postgis;
ALTER TABLE warehouses ADD COLUMN location GEOGRAPHY(POINT, 4326) DEFAULT NULL;
ALTER TABLE customers ADD COLUMN location GEOGRAPHY(POINT, 4326) DEFAULT NULL;
ALTER TABLE sellers ADD COLUMN location GEOGRAPHY(POINT, 4326) DEFAULT NULL;
-- Backfill from existing lat/lng columns
UPDATE warehouses SET location = ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography;
UPDATE customers SET location = ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography;
UPDATE sellers SET location = ST_SetSRID(ST_MakePoint(lng, lat), 4326)::geography;
-- Spatial indexes
CREATE INDEX idx_warehouses_location ON warehouses USING GIST(location);
CREATE INDEX idx_customers_location ON customers USING GIST(location);
CREATE INDEX idx_sellers_location ON sellers USING GIST(location);
```

#### `000004_users_rbac.up.sql` (80 lines)
```sql
CREATE TYPE user_role AS ENUM ('admin', 'seller', 'buyer');
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role user_role NOT NULL,
  reference_id BIGINT, -- FK to seller_id or customer_id
  reference_type VARCHAR(50), -- 'seller' or 'customer'
  is_active BOOLEAN DEFAULT true,
  last_login_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

CREATE TABLE refresh_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  revoked_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
```

#### `000005_extend_entities.up.sql` (200 lines)
Adds columns to existing tables:
- `products`: `brand`, `unit_of_measure` (enum), `barcode_ean13`, `hsn_code`, `gst_rate_percent`, `image_url`, `thumbnail_url`, `tags` (text[]), `is_available`, `reorder_threshold`, `shelf_life_days`, `created_by_seller_id`
- `sellers`: `logo_url`, `banner_url`, `store_description`, `operating_days` (text[]), `opening_time`, `closing_time`, `accepted_payment_modes` (text[]), `avg_dispatch_hours`, `total_orders_fulfilled`, `cancellation_rate_percent`, `avg_rating` (decimal 2,1)
- `customers`: `profile_photo_url`, `alternate_phone`, `fssai_license_number`, `shop_area_sqft`, `monthly_order_value_avg`, `loyalty_points`, `credit_balance`
- `warehouses`: `serviceable_pincodes` (text[]), `operating_hours` (text), `temperature_zones` (text[]: dry/cold/frozen)
- `orders`: `invoice_number` (unique), `coupon_code`, `discount_amount`, `gst_amount`, `invoice_url`, `buyer_notes`, `seller_notes`, `rated_at`, `buyer_rating` (1-5), `seller_rating` (1-5)

#### `000006_new_features.up.sql` (150 lines)
New tables:
- `categories` — product taxonomy with parent_id (hierarchical)
- `saved_lists` (Inventory Saver) — buyer's reorder templates
- `saved_list_items` — items in saved lists
- `cart` — buyer's current shopping session
- `cart_items` — items in cart with quantity snapshots
- `product_price_history` — price change tracking for comparison dashboard
- `order_status_events` — Observer pattern: immutable event log of status transitions
- `reviews` — product/seller review table (for future star ratings)

---

## Backend Implementation (New Code)

### Models (`internal/models/`)

| File | Structs | Purpose |
|---|---|---|
| `user.go` | `User`, `RefreshToken` | Auth-related models |
| `auth.go` | `RegisterRequest`, `LoginRequest`, `TokenPair`, `Claims` | Auth request/response DTOs |
| `cart.go` | `Cart`, `CartItem` | Shopping cart models |
| `saved_list.go` | `SavedList`, `SavedListItem` | Inventory saver models |
| `category.go` | `Category` | Product category with hierarchy |
| `order_event.go` | `OrderStatusEvent` | Event log for Observer pattern |
| ~extend existing~ | — | Add new fields to Product, Seller, Customer, Warehouse, Order models |

### Repositories (`internal/repositories/`)

| Interface | Methods | Note |
|---|---|---|
| `UserRepository` | Create, GetByEmail, GetByID, UpdateLastLogin, Deactivate | New |
| `RefreshTokenRepository` | Create, GetByTokenHash, Revoke, CleanupExpired | New |
| `CartRepository` | GetActive, Create, ClearItems, Delete | New |
| `SavedListRepository` | Create, GetByCustomerID, AddItem, RemoveItem, GetItems | New |
| `ProductRepository` | *existing + new* GetByCategoryID, GetByPriceBand, Search (full-text), GetCompare, UpdateStock | Enhanced |
| `OrderRepository` | *existing + new* GetByCustomer, GetBySeller, UpdateStatus, GetEvents | Enhanced |
| `WarehouseRepository` | *existing + new* NearbyByRadius (PostGIS), GetCoverage, GetByPincodeList | Enhanced with PostGIS |
| `CategoryRepository` | GetTree, Create, Update | New |
| `OrderEventRepository` | Create, GetByOrderID | New (Observer pattern) |

### Services (`internal/services/`)

| File | Methods | Purpose |
|---|---|---|
| `auth_service.go` | Register, Login, Refresh, Logout, ValidateToken | New auth service |
| `user_service.go` | GetUserByID, GetAllUsers (admin), DeactivateUser | New user management |
| `order_service.go` | PlaceOrder (with transaction), CancelOrder, UpdateStatus, RateOrder | New order lifecycle |
| `cart_service.go` | AddItem, RemoveItem, UpdateQuantity, GetCart, Clear | New cart service |
| `saved_list_service.go` | Create, GetLists, AddItem, Reorder (one-click) | New Inventory Saver |
| `product_service.go` | Search, Compare, GetByCategory, TrackPriceHistory | Enhanced product service |
| `storage/s3_service.go` | UploadImage, DeleteFile, GetSignedURL | New S3 upload service |
| `notification_service.go` (stub) | EmitOrderEvent, Subscribe, Publish | Observer pattern; decoupled for future integrations |

### Handlers (`internal/handlers/`)

| File | Routes | Purpose |
|---|---|---|
| `auth_handler.go` | POST /api/v1/auth/{register,login,refresh,logout} | New auth endpoints |
| `user_handler.go` | GET /api/v1/admin/users, PATCH /api/v1/admin/users/:id/status | New user mgmt (admin) |
| `cart_handler.go` | POST/PUT/DELETE /api/v1/cart/{items} | New cart endpoints |
| `order_handler.go` | POST/GET /api/v1/orders, PATCH status/rate, GET /api/v1/seller/orders | New order endpoints + seller order view |
| `saved_list_handler.go` | CRUD /api/v1/saved-lists + POST reorder | New Inventory Saver |
| `product_handler.go` (enhanced) | GET /api/v1/products, POST /api/v1/seller/products, GET /api/v1/products/compare | Enhanced product APIs |
| `category_handler.go` | GET /api/v1/categories, POST /api/v1/admin/categories | New category endpoints |
| `warehouse_handler.go` (enhanced) | GET /api/v1/warehouses/nearby, GET /api/v1/warehouses/:id/coverage | Enhanced warehouse APIs |
| `admin_handler.go` | GET /api/v1/admin/dashboard, /orders, /config/* | New admin panel APIs |

### Middleware

| File | Purpose |
|---|---|
| `role_guard.go` | `RequireRole(roles ...UserRole)` — composes with auth middleware; returns 403 if not authorized |
| `error_handler.go` | Central error handler; maps `AppError` types to HTTP responses with structured error envelopes |
| `cache_interceptor.go` (optional) | Decorator middleware that wraps repo calls; can be applied per route |

### Error Handling (`pkg/errors/`)

Extend `errors.go` with:
```go
type AppError struct {
  Code       string      // e.g., "WAREHOUSE_NOT_FOUND"
  Message    string      // user-facing message
  HTTPStatus int         // 400, 404, 409, etc.
  Details    interface{} // optional debugging info
}
```

**Sentinel errors:**
- `ErrWarehouseNotFound` (404)
- `ErrDeliveryUnsupported` (422) — customer pincode not serviceable
- `ErrInvalidParam` (400)
- `ErrProductNotFound` (404)
- `ErrSellerNotVerified` (403)
- `ErrInsufficientStock` (409)
- `ErrOrderNotCancellable` (409)
- `ErrUnauthorized` (401)
- `ErrForbidden` (403)
- `ErrCartEmpty` (400)
- `ErrSelfReference` (400) — seller/buyer same entity

### Caching Strategy (`internal/cache/`)

Extend cache service with typed key builders and 15+ cache patterns:

| Key Pattern | TTL | Invalidated On |
|---|---|---|
| `nearest_wh:{sellerID}` | 10 min | warehouse update |
| `shipping_charge:{wh}:{cust}:{prod}:{speed}` | 5 min | product weight/price change |
| `products:list:{hash(filters)}` | 5 min | product create/update/delete |
| `product:{id}` | 10 min | product update |
| `product:compare:{sorted_ids}` | 3 min | any product in set changes |
| `seller:products:{sellerID}` | 5 min | seller adds/updates products |
| `categories:tree` | 1 hr | category tree update |
| `saved_lists:{customerID}` | 2 min | list create/update |
| `cart:{customerID}` | 1 min | item add/remove |
| `orders:list:{customerID}:{page}` | 2 min | new order, status change |
| `warehouse:nearby:{lat}:{lng}:{radius}` | 10 min | warehouse update |
| `admin:dashboard` | 1 min | any order event |
| `seller:orders:{sellerID}:{page}` | 2 min | order status change |
| `user:{id}` | 15 min | user profile update |
| `config:transport-rates` | 30 min | rate config update |

Cache invalidation is **eager** (delete on mutation) for correctness-critical data (orders, stock); TTL-only for read-heavy data (product list, categories).

### Config Extension (`config/config.go`)

Add `StorageConfig`:
```go
type StorageConfig struct {
  Provider      string // "s3" or "local"
  Bucket        string
  Region        string
  Endpoint      string // for Cloudflare R2
  AccessKey     string
  SecretKey     string
  MaxFileSizeMB int
}
```

---

## Frontend Implementation (Next.js 14)

### New npm Dependencies
```json
{
  "leaflet": "^1.9.x",
  "react-leaflet": "^4.x",
  "framer-motion": "^10.x",
  "recharts": "^2.x",
  "@aws-sdk/client-s3": "^3.x",
  "react-hot-toast": "^2.x",
  "zustand": "^4.x",
  "@tanstack/react-query": "^5.x",
  "react-dropzone": "^14.x",
  "date-fns": "^3.x"
}
```

### State Management

**Zustand stores** (`frontend/lib/store.ts`):
- `useAuthStore` — user, token, login/logout actions
- `useCartStore` — items, itemCount, addItem, removeItem, getTotal
- `useUIStore` — sidebarOpen, theme, notifications queue

**TanStack Query** — all server state (product list, orders, warehouses) goes through `useQuery`/`useMutation` hooks; automatic caching + refetch on focus.

### New Pages (App Router)

| Route | Auth | Component | Key Features |
|---|---|---|---|
| `/` | Public | Landing | Hero + CTA, animated products, `Seedha Source Se` tagline |
| `/login` | Public | AuthPage | Split-screen; role tabs (Buyer/Seller/Admin); JWT in httpOnly cookie |
| `/register` | Public | RegisterPage | Form validation, terms acceptance, optional GST/FSSAI input |
| `/buyer/dashboard` | Buyer | BuyerDashboard | Active orders, saved lists, quick-reorder, loyalty points |
| `/buyer/browse` | Buyer | BrowsePage | Product grid, filter sidebar (category, price, seller, stock), search bar (debounced) |
| `/buyer/compare` | Buyer | ProductComparePage | Side-by-side table (up to 4 products), price alerts, cheapest highlighted |
| `/buyer/cart` | Buyer | CartDrawer | Slide-in, item list, shipping estimate per seller group, delivery speed toggle, delete/update quantity |
| `/buyer/checkout` | Buyer | CheckoutPage | 3-step wizard: cart review → delivery details → confirm + pay |
| `/buyer/orders` | Buyer | OrdersPage | List with status badges, timeline progress bar, filter by status/date |
| `/buyer/orders/[id]` | Buyer | OrderDetailPage | Full breakdown, shipment tracking, "Rate Order" modal, invoice download |
| `/buyer/saved-lists` | Buyer | SavedListsPage | Card grid; each card shows item count + "Reorder All" CTA; manage items modal |
| `/seller/dashboard` | Seller | SellerDashboard | Stats: products, active orders, revenue, rating; quick-upload button |
| `/seller/catalog` | Seller | CatalogPage | Product table with inline edit, image upload modal (→ S3 presigned URL) |
| `/seller/orders` | Seller | SellerOrdersPage | Received orders, status dropdown update, bulk action (mark as dispatched) |
| `/seller/analytics` | Seller | SellerAnalyticsPage | Chart: orders/day, revenue trend, top products |
| `/admin/dashboard` | Admin | AdminDashboard | Recharts: GMV/day, orders/day, top sellers, order status breakdown |
| `/admin/users` | Admin | UsersPage | User list table, role badge, activate/deactivate toggle |
| `/admin/orders` | Admin | AllOrdersPage | Filter builder (status, date range, seller, customer), force-status override |
| `/admin/sellers` | Admin | SellersPage | All sellers, verification badge, rating, bulk verify/suspend |
| `/admin/warehouses` | Admin | WarehousesPage | Leaflet map with markers + coverage circles, table below, add warehouse modal |
| `/admin/config` | Admin | ConfigPage | Transport rates table (editable inline), delivery speed config cards |

### Shared Components

**Collections:**
- `<Navbar>` — role-aware links, user menu, logout
- `<CartDrawer>` — slide-in cart (zustand state)
- `<WarehouseMap>` — Leaflet component with markers/circles
- `<StatusBadge>` — order status visual (placed/shipped/delivered/cancelled with colors)
- `<PriceTag>` — ₹ formatted with `en-IN` locale, optional strikethrough original
- `<ProductCard>` — grid card with image, price, seller, "Add to Cart"
- `<OrderCard>` — compact order summary with status + CTA
- `<SavedListCard>` — list name, item count, "Reorder All" button
- `<CompareTable>` — side-by-side product comparison with price highlight
- `<LoadingSkeleton>` — Tailwind shimmer effect for placeholders
- `<EmptyState>` — Indian bazaar SVG illustration + message
- `<FilterSidebar>` — reusable filter controls (category, price slider, toggle)
- `<ErrorBoundary>` — React error fallback with user-friendly message
- `<FormInput>`, `<FormSelect>`, `<FormCheckbox>` — styled form controls

### Design System

**Brand Colors (Indian Aesthetic):**
- Primary: Saffron `#E8560A` (accent, CTAs)
- Secondary: Earthy Green `#2D6A4F` (trust, seller badges)
- Tertiary: Brass/Gold `#D4AF37` (premium, loyalty points)
- Neutral: Warm Cream `#FDF6EC` (background)
- Accent: Deep Blue `#1B3A57` (links, inputs)

**Typography:**
- Headings: `Rajdhani` (bold, geometric, numbers)
- Body: `DM Sans` (clean, accessible)
- Monospace: `JetBrains Mono` (prices, tracking IDs)

**Motion:** Framer Motion for:
- Page transitions (fadeInUp on mount)
- Cart slide-in (x: -100 → 0)
- Toast notifications (slideDown + fadeOut)
- Product hover (scaleUp, shadow elevation)
- Status timeline animation (line stroke + dot pulse)

---

## API Specifications (OpenAPI-Compatible)

### Auth Endpoints

```
POST /api/v1/auth/register
  Body: { email, password, role, sellerName?, customerName?, phone, gstNumber? }
  Response: { success, data: { accessToken, refreshToken, user } }

POST /api/v1/auth/login
  Body: { email, password }
  Response: { success, data: { accessToken, refreshToken, user } }

POST /api/v1/auth/refresh
  Body: { refreshToken }
  Response: { success, data: { accessToken, refreshToken } }

POST /api/v1/auth/logout
  Header: Authorization: Bearer <token>
  Response: { success }
```

### Product Endpoints

```
GET /api/v1/products
  Params: category?, search?, sellerId?, minPrice?, maxPrice?, page?, limit?
  Response: { success, data: { items: [Product], total, page, pageSize } }
  Cache: 5 min (key: products:list:{hash(filters)})

GET /api/v1/products/:id
  Response: { success, data: Product }
  Cache: 10 min (key: product:{id})

GET /api/v1/products/compare?ids=1,2,3
  Response: { success, data: { items: [Product], minPriceSellerId } }
  Cache: 3 min (key: product:compare:{sorted_ids})

POST /api/v1/seller/products (Seller only)
  Body: multipart/form-data { name, price, weight, image, hsn_code, gst_rate, ... }
  Response: { success, data: Product }
  Cache Invalidation: delete products:list:*, seller:products:{sellerId}

PUT /api/v1/seller/products/:id (Seller only, ownership check)
  Body: { name, price, weight, image?, ... }
  Response: { success, data: Product }

DELETE /api/v1/seller/products/:id (Seller only)
  Response: { success }
```

### Cart Endpoints

```
GET /api/v1/cart (Buyer only)
  Response: { success, data: { items: [CartItem], total, shippingEstimate } }
  Cache: 1 min

POST /api/v1/cart/items (Buyer only)
  Body: { productId, sellerId, quantity, deliverySpeed }
  Response: { success, data: CartItem }

PUT /api/v1/cart/items/:id (Buyer only)
  Body: { quantity, deliverySpeed? }
  Response: { success, data: CartItem }

DELETE /api/v1/cart/items/:id (Buyer only)
  Response: { success }
```

### Order Endpoints

```
POST /api/v1/orders (Buyer only)
  Body: { cartId, paymentMode, deliveryNotes? }
  Response: { success, data: Order }
  Side Effects: create Order + Shipment records, deduct stock, emit order_placed event, clear cart

GET /api/v1/orders (Buyer only)
  Params: status?, page?, limit?
  Response: { success, data: { items: [Order], total } }

GET /api/v1/orders/:id (Buyer or Seller or Admin)
  Response: { success, data: { order, shipment, lineItems, statusTimeline } }

PATCH /api/v1/orders/:id/cancel (Buyer only, status=placed)
  Response: { success, data: Order }

PUT /api/v1/orders/:id/rate (Buyer only, after delivered)
  Body: { buyerRating, sellerRating, review? }
  Response: { success, data: Order }

GET /api/v1/seller/orders (Seller only)
  Params: status?, page?
  Response: { success, data: [Order] }

PATCH /api/v1/seller/orders/:id/status (Seller only)
  Body: { status } (validation: can only move forward)
  Response: { success, data: Order, event: OrderStatusEvent }
```

### Saved Lists (Inventory Saver)

```
GET /api/v1/saved-lists (Buyer only)
  Response: { success, data: [SavedList] }

POST /api/v1/saved-lists (Buyer only)
  Body: { name, description? }
  Response: { success, data: SavedList }

POST /api/v1/saved-lists/:id/items (Buyer only)
  Body: { productId, sellerId, quantity, deliverySpeed? }
  Response: { success, data: SavedListItem }

POST /api/v1/saved-lists/:id/reorder (Buyer only)
  Response: { success, data: { cart, addedItems, skippedItems, message } }
  Side Effect: adds all items to cart (skips out-of-stock); returns summary

DELETE /api/v1/saved-lists/:id (Buyer only)
  Response: { success }
```

### Warehouse Endpoints

```
GET /api/v1/warehouses/nearby?lat=12.97&lng=77.59&radius_km=100 (Buyer)
  Response: { success, data: [Warehouse with distance] }
  Cache: 10 min

GET /api/v1/warehouses/:id/coverage (Public)
  Response: { success, data: { warehouseId, serviciablePincodes, coverageRadiusKm, geoJsonPolygon? } }

GET /api/v1/admin/warehouses (Admin only)
  Response: { success, data: [Warehouse with capacity %] }

POST /api/v1/admin/warehouses (Admin only)
  Body: { name, code, lat, lng, maxCapacitySqft?, servicePincodes[], ... }
  Response: { success, data: Warehouse }

PUT /api/v1/admin/warehouses/:id (Admin only)
  Response: { success, data: Warehouse }
```

### Admin Endpoints

```
GET /api/v1/admin/dashboard (Admin only)
  Response: { success, data: { totalOrders, totalGMV, topProducts, topSellers, orderStatusBreakdown, ... } }
  Cache: 1 min

GET /api/v1/admin/config/transport-rates (Admin only)
  Response: { success, data: [TransportRate] }

POST /api/v1/admin/config/transport-rates (Admin only)
  Body: { distanceBandMinKm, distanceBandMaxKm, transportMode, ratePerKmPerKg, effectiveFrom, effectiveTo }
  Response: { success, data: TransportRate }

GET /api/v1/admin/config/delivery-speeds (Admin only)
  Response: { success, data: [DeliverySpeedConfig] }

PUT /api/v1/admin/config/delivery-speeds/:id (Admin only)
  Body: { baseCharge, extraChargePerKg }
  Response: { success, data: DeliverySpeedConfig }
```

---

## Testing Strategy

### Unit Tests
- **Auth Service** — register, login, logout; password hashing; token expiry
- **Order Service** — place order transaction; stock deduction; event emission
- **Product Service** — search/filter; comparison logic; price history tracking
- **Cache Service** — get/set/delete operations; TTL handling
- **Validators** — GST number, pincode, phone (Indian format)

Testing Framework: `testify` + table-driven tests

### Integration Tests
1. **Auth Flow** — register (seller) → login → access `/seller/dashboard` → refresh token → logout
2. **Order Lifecycle** — browse products → add to cart → adjust quantity → checkout → order created → seller views order → update status → buyer rates
3. **Saved List** — create list → add items (multiple sellers) → reorder → cart populated → checkout
4. **PostGIS Query** — find nearest warehouse (compare with v1 Haversine result ±0.1 km); query by radius
5. **Cache Invalidation** — update product price → check cache invalidates → re-query returns new price
6. **RBAC** — buyer cannot POST to `/seller/products`; seller cannot access `/admin/dashboard`; all endpoints return 403/401 appropriately

Testing Framework: `testify` + Docker Compose for DB/Redis

### Manual / Smoke Tests
- **Happy Path:** buyer registers → browses products → adds to cart → checks out → receives order confirmation → seller accepts → order shipped → buyer rates
- **Error Scenarios:** invalid pincode (delivery unsupported) → graceful error; product out of stock → cannot add to cart; seller tries to update another seller's product → 403
- **Frontend Build:** `npm run build` with zero TypeScript errors; all pages render without crashes under each role
- **PostGIS Accuracy:** query nearest warehouse, compare distance to v1 Haversine (should match ±0.1 km)

---

## Deployment & DevOps

### Docker & Compose

**docker-compose.yml** additions:
```yaml
db:
  image: postgis/postgis:16-3.4  # upgraded from postgres:16
  environment:
    - POSTGRES_USER=lalaji_user
    - POSTGRES_PASSWORD=<secure>
    - POSTGRES_DB=lalaji

s3-compat:  # optional local minio for dev, or configure real R2
  image: minio/minio:latest
  environment:
    - MINIO_ROOT_USER=minioadmin
    - MINIO_ROOT_PASSWORD=minioadmin

backendv2:
  build: ./cmd/server
  environment:
    - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
    - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
    - S3_BUCKET=${S3_BUCKET}
    - S3_ENDPOINT=${S3_ENDPOINT}  # for R2 or minio
```

### Render / Production

Update `render.yaml` to specify:
- PostgreSQL instance type supporting PostGIS (e.g., `Standard-4`)
- Environment variable placeholders for S3 credentials
- Separate build pack for Next.js frontend (or separate web service)

### GitHub Actions / CI-CD

- Run unit tests on every PR
- Run integration tests (with Docker Compose) on PR
- Build Go binary + frontend
- Push Docker images to registry on merge to main
- Deploy to Render on tag/release

---

## Timeline & Effort Estimate

| Phase | Tasks | Effort | Days |
|---|---|---|---|
| **DB & Auth** | Migrations 3-6, UserRepo, AuthService, AuthHandler, RoleGuard | High | 3-4 |
| **Product Upload** | S3Service, ProductHandler (POST/PUT), StorageConfig | Medium | 2 |
| **Orders & Cart** | OrderService (transaction), Cart APIs, Cart UI, Saved Lists | High | 4-5 |
| **Admin Panel** | AdminHandler, admin UI pages (dashboard, users, orders, config) | High | 3-4 |
| **Error Handling & Caching** | Domain errors, error middleware, cache decorator, 15+ cache keys | Medium | 2-3 |
| **PostGIS** | Migration 3, WarehouseRepo refactor (ST_Distance), map UI | Medium | 2 |
| **Frontend (Creative UI)** | Layout redesign, all new pages, components, animations, Leaflet map | High | 5-6 |
| **Testing** | Unit tests, integration tests, smoke tests, PostGIS accuracy | Medium | 2-3 |
| **Deployment** | docker-compose updates, Render config, CI-CD pipeline | Low | 1 |

**Total Estimated Effort:** ~4-6 weeks (parallel tracks possible: backend auth/cart, sales team for S3 setup, frontend UI design)

---

## Key Success Criteria

- [ ] All APIs gracefully handle invalid/missing parameters with 4xx responses
- [ ] PostGIS distance queries match v1 Haversine ±0.1 km
- [ ] JWT refresh token rotation is stateful (tokens can be revoked)
- [ ] RBAC: buyers cannot see sellers' orders; sellers cannot modify admin config
- [ ] Cache invalidation is eager (not TTL-only) for orders and stock
- [ ] Frontend has zero TypeScript errors on `npm run build`
- [ ] Product image upload to S3/R2 works end-to-end
- [ ] Order placement is atomic (DB transaction; all-or-nothing)
- [ ] PostGIS Leaflet map renders without errors on buyer/admin pages
- [ ] Saved list "Reorder All" adds items correctly, skips out-of-stock gracefully

---

## Notes

- This document is a **template** for future implementation; no code changes made as of March 3, 2026.
- All v1 design patterns and existing code remain intact when v2 is implemented incrementally.
- Payment gateway (Razorpay, etc.) is **out of scope** for v2; `payment_mode` field stores intent but no actual charge integration.
- Notifications (email, SMS, WhatsApp) are **out of scope**; `order_status_events` table provides hook for future async workers.
- Image optimization (thumbnails, webp conversion) can be added as post-processing in S3 upload workflow (e.g., via Lambda or next-image).
- Recommendation engine (e.g., "Buyers like you also ordered...") is **future phase** (v3).

---

**Created by:** Copilot | **Last Updated:** March 3, 2026
