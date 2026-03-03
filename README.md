# 📦 Jambotails — B2B Shipping Charge Estimator

<div align="center">

[![Live Demo](https://img.shields.io/badge/Live%20Demo-Vercel-blue?style=for-the-badge&logo=vercel)](https://lalaji-eight.vercel.app/)
[![API Docs](https://img.shields.io/badge/API%20Docs-Swagger-green?style=for-the-badge&logo=swagger)](https://lalaji-eight.vercel.app/docs)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-14-000000?style=for-the-badge&logo=next.js)](https://nextjs.org)

**A modern B2B Kirana shipping charge estimator with real-time warehouse routing, smart pricing strategies, and a beautiful animated frontend.**

</div>

A Go REST API that estimates shipping charges for Kirana (small retail) stores in a B2B e-commerce context. Given a seller, product, customer, and delivery speed, it finds the nearest warehouse and calculates the total shipping charge using the **Haversine formula**, a **Strategy Pattern** for transport modes and pricing, and **DB-configurable** rates.

---

## 🎯 Quick Start

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

## 🚀 Project Walkthrough

### **Step 1: Select a Seller**
Choose from 3 B2B sellers (distributors) across India. Each seller has a geographic location and product catalog.

```
✅ Nestle India Distributor (Bengaluru) — FMCG products
✅ Premium Rice Traders (Mumbai)         — Staple foods
✅ Gujarat Sugar Mills (Ahmedabad)       — Commodities
```

**Under the hood:**
- API validates `sellerId` exists and is active
- Retrieves seller coordinates (latitude/longitude) from DB
- Used to find the nearest warehouse in the network

---

### **Step 2: Pick a Product**
Select a product from the chosen seller's inventory. Each product has:
- **Price & weight** — for billable weight calculation
- **Fragility/Perishability** — future fulfillment flags
- **Volumetric dimensions** — L×W×H for dimensional weight

```json
{
  "id": 1,
  "name": "Maggi 500g Packet",
  "price": ₹10 (₹14 MRP),
  "weight": 0.5 kg,
  "dimensions": "20cm × 15cm × 5cm"
}
```

---

### **Step 3: Choose a Kirana Store (Delivery Destination)**
Select where the shipment should be delivered. The app shows 5 stores across major Indian metros.

```
🏬 Shree Kirana Store (Bengaluru)      → PIN 560034
🏬 Andheri Mini Mart (Mumbai)          → PIN 400053
🏬 Dilli Grocery Hub (New Delhi)       → PIN 110001
🏬 Hyderabad Fresh Mart (Hyderabad)    → PIN 500034
🏬 Chennai Bazaar (Chennai)            → PIN 600017
```

---

### **Step 4: Choose Delivery Speed & Get Estimate**

#### 🐢 **Standard Delivery** (slower, cheaper)
- Base charge: ₹10
- Rate: ₹{distanceKm} × {billableWeight}
- Examples: 200 km × 10 kg @ ₹2/km/kg = ₹4,000 + ₹10 base = **₹4,010**

#### ⚡ **Express Delivery** (faster, includes surcharge)
- Base charge: ₹10
- Rate: ₹{distanceKm} × {billableWeight}
- Express surcharge: ₹1.2 per kg
- Example: Same route + weight = ₹4,010 + (10 kg × ₹1.2) = **₹4,022**

**What happens behind the scenes:**

1. **Nearest Warehouse Lookup** — Using Haversine formula, finds the closest warehouse to the seller
   - Distance from Seller → Warehouse (e.g., 4.26 km)
   
2. **Transport Mode Selection** — Automatically selects based on final distance:
   - 🚐 **MiniVan** (0–100 km) — local delivery
   - 🚚 **Truck** (100–500 km) — regional distribution
   - ✈️ **Aeroplane** (500+ km) — cross-country express

3. **Billable Weight Calculation**
   ```
   billableWeight = max(
     actualWeight,                    // 10 kg
     volumetricWeight                 // L×W×H÷5000
   )
   ```

4. **Total Charge Calculation**
   ```
   totalCharge = baseCourier + distanceCharge + expressCharge
   ```

5. **Response Chain**
   - Full route visualization (Seller → Warehouse → Customer)
   - Charge breakdown table
   - Nearest warehouse details + Google Maps link

---

## 📊 API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | 🏥 DB + Redis health check |
| `GET` | `/api/v1/warehouse/nearest` | 🏭 Find nearest warehouse to seller |
| `GET` | `/api/v1/shipping-charge` | 💰 Compute shipping charge |
| `POST` | `/api/v1/shipping-charge/calculate` | 📦 End-to-end: nearest warehouse + charge |

**📖 Interactive Docs:** [Swagger UI at `{API_URL}/docs`](https://lalaji-eight.vercel.app/docs)

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

## 🧪 Testing

### 🌐 Web UI
Visit the live demo: **[https://lalaji-eight.vercel.app/](https://lalaji-eight.vercel.app/)**

### 📖 Interactive API Testing
1. Open **[Swagger UI at /docs](https://lalaji-eight.vercel.app/docs)**
2. Click "Try it out" on any endpoint
3. Test with sample data (auto-populated)
4. View full request/response including headers

### 💻 Command-line Testing

**Check health:**
```bash
curl http://localhost:8080/health
```

**Find nearest warehouse:**
```bash
curl "http://localhost:8080/api/v1/warehouse/nearest?sellerId=1&lat=12.9716&lng=77.5946"
```

**Calculate shipping charge (POST):**
```bash
curl -X POST http://localhost:8080/api/v1/shipping-charge/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "sellerId": 1,
    "productId": 1,
    "customerId": 1,
    "deliverySpeed": "standard"
  }'
```

### ⚡ Load Testing

**Using `hey`:**
```bash
hey -n 1000 -c 50 -m POST \
  -H "Content-Type: application/json" \
  -d '{"sellerId":1,"productId":1,"customerId":1,"deliverySpeed":"standard"}' \
  http://localhost:8080/api/v1/shipping-charge/calculate
```

**Using `k6`:**
```js
// save as load_test.js
import http from 'k6/http';
export const options = { vus: 50, duration: '30s' };
export default function () {
  http.post('http://localhost:8080/api/v1/shipping-charge/calculate',
    JSON.stringify({ sellerId: 1, productId: 1, customerId: 1, deliverySpeed: 'standard' }),
    { headers: { 'Content-Type': 'application/json' } });
}
```

Run: `k6 run load_test.js`

---

## 🏗️ Project Structure

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

## ⚙️ Architecture Highlights

- **🎯 Strategy Pattern** — transport mode (MiniVan/Truck/Aeroplane) selected by distance range; pricing (Standard/Express) selected by speed
- **🗂️ Repository Pattern** — all DB access via interfaces for easy mocking in tests
- **📍 Haversine Formula** — great-circle distance between lat/lng coordinates
- **⚖️ Billable Weight** — `max(actualWeight, volumetricWeight)` where volumetric = `L × W × H / 5000`
- **⚙️ DB-configurable Rates** — transport rates, delivery speed configs, base charges all stored in DB
- **🛡️ Graceful Degradation** — if Redis is down, API continues using DB directly
- **🧪 Error Handling** — Typed domain errors with proper HTTP status codes (400/404/422/503)
- **📡 Request Tracing** — Every request gets a unique ID for distributed logging

---

## 🧬 Running Tests

```bash
make test                 # all tests, verbose
make test-cover           # with HTML coverage report
```

✅ **39 unit tests** covering:
- Haversine distance calculation
- Transport strategy selection (boundary conditions)
- Pricing calculations (standard vs express)
- Billable weight calculations
- Full service-level tests with mock repositories

---

## 📚 Documentation

| # | 📄 Document | 📝 Purpose |
|---|----------|---------|
| 1 | [Requirements Mapping](docs/01-requirements.md) | FRs, NFRs, business rules, validation matrix |
| 2 | [Design Document](docs/02-design.md) | Entity design, architecture, data flow, API contracts |
| 3 | [Task Breakdown](docs/03-tasks.md) | Phased deliverables, effort estimates, dependency graph |
| 4 | [V2 Roadmap](docs/04-v2-roadmap.md) | PostGIS mapping, login, seller catalog, inventory tracking |

---

## 🛠️ Make Targets

| Target | Description |
|--------|-------------|
| `make build` | 🔨 Compile binary to `bin/` |
| `make run` | ▶️ Build + run server |
| `make test` | ✅ Run all unit tests |
| `make test-cover` | 📊 Tests + HTML coverage |
| `make lint` | 🔍 golangci-lint |
| `make migrate-up` | ⬆️ Apply DB migrations |
| `make migrate-down` | ⬇️ Rollback migrations |
| `make docker-up` | 🐳 Start full stack via Docker Compose |
| `make docker-down` | 🛑 Tear down Docker stack |

---

## 📦 Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| **Backend** | Go | 1.25 |
| **API Framework** | Gin | latest |
| **Database** | PostgreSQL | 16 |
| **Cache** | Redis | 7 |
| **Frontend** | Next.js | 14 |
| **Styling** | Tailwind CSS | 3 |
| **Type Safety** | TypeScript | latest |
| **API Docs** | OpenAPI 3.0 / Swagger UI | 5.x |

---

## 🎨 Frontend Features

- **📱 Responsive Design** — Mobile-first, works on all devices
- **✨ Smooth Animations** — Keyframe animations for cards, loading states, and transitions
- **🌈 Modern Palette** — Violet/teal/emerald/amber color scheme
- **⚡ Real-time Feedback** — Instant validation and error messages
- **🗺️ Interactive Route Map** — Visual shipping route with distances
- **💾 Session Memory** — Remembers your selections

---

## 📋 License

MIT License — see LICENSE file for details

---

<div align="center">

**Built with ❤️ for B2B Kirana e-commerce**

[🌐 Live Demo](https://lalaji-eight.vercel.app/) · [📖 API Docs](https://lalaji-eight.vercel.app/docs) · [💻 GitHub](https://github.com/luxmikant/lalaji)

</div>
