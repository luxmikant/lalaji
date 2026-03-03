# Design — B2B Kirana Shipping Charge Estimator

> **Project:** Jambotails — B2B e-Commerce Shipping Module  
> **Tech Stack:** Go (Gin), PostgreSQL, Redis  
> **Date:** March 3, 2026  
> **Status:** Draft v1.0

---

## Table of Contents

1. [Technology Choices](#1-technology-choices)
2. [Entity Design](#2-entity-design)
3. [Database Schema (PostgreSQL)](#3-database-schema)
4. [Data Flow](#4-data-flow)
5. [High-Level Architecture](#5-high-level-architecture)
6. [Project Structure](#6-project-structure)
7. [API Design](#7-api-design)
8. [Core Algorithms](#8-core-algorithms)
9. [Design Patterns Used](#9-design-patterns-used)
10. [Caching Strategy](#10-caching-strategy)

---

## 1. Technology Choices

| Layer | Technology | Reason |
|---|---|---|
| Language | Go 1.22 | Performance, native concurrency, compiled binary |
| Web Framework | Gin | Fast HTTP router, middleware ecosystem |
| Database | PostgreSQL 16 | Relational, PostGIS for geo, ACID transactions |
| Cache | Redis 7 | Sub-ms reads, TTL support, simple key-value |
| Validation | go-playground/validator v10 | Struct tag validation, custom validators |
| Logging | uber-go/zap | Structured JSON logs, high performance |
| Metrics | prometheus/client_golang | Industry standard observability |
| Testing | testify + httptest | Go-native, mock-friendly |
| Migration | golang-migrate | SQL-based, version controlled schema |
| Config | viper | Env-based, 12-factor app |

---

## 2. Entity Design

### 2.1 Customer (Kirana Store Owner)
```
Represents a registered Kirana store that places orders on the platform.
Creative additions: GST number (B2B legal), store type, credit limit, 
preferred delivery slot — all critical for a small retailer context.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | Auto-increment |
| `name` | VARCHAR(255) | Store/shop name |
| `owner_name` | VARCHAR(255) | Owner's personal name |
| `phone` | VARCHAR(15) | Primary contact |
| `email` | VARCHAR(255) | For order notifications |
| `gst_number` | VARCHAR(20) | B2B tax compliance (unique) |
| `store_type` | ENUM | `grocery`, `dairy`, `general`, `medical`, `bakery` |
| `credit_limit` | DECIMAL(12,2) | Max credit allowed for B2B invoices |
| `preferred_delivery_slot` | VARCHAR(50) | e.g. `"morning"`, `"evening"` — UX hint |
| `lat` | DECIMAL(10,7) | Store latitude |
| `lng` | DECIMAL(10,7) | Store longitude |
| `address_line1` | VARCHAR(255) | Street address |
| `city` | VARCHAR(100) | City |
| `state` | VARCHAR(100) | State |
| `pincode` | VARCHAR(10) | PIN code |
| `is_active` | BOOLEAN | Soft delete / deactivation |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

---

### 2.2 Seller
```
A wholesale supplier that lists products and drops them at the nearest warehouse.
Creative additions: GST, business type, rating — mirrors real B2B seller lifecycle.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | Auto-increment |
| `name` | VARCHAR(255) | Business name |
| `owner_name` | VARCHAR(255) | |
| `phone` | VARCHAR(15) | |
| `email` | VARCHAR(255) | |
| `gst_number` | VARCHAR(20) | Unique, legally required in B2B India |
| `business_type` | ENUM | `manufacturer`, `distributor`, `wholesaler` |
| `lat` | DECIMAL(10,7) | Seller's dispatch location latitude |
| `lng` | DECIMAL(10,7) | Seller's dispatch location longitude |
| `address_line1` | VARCHAR(255) | |
| `city` | VARCHAR(100) | |
| `state` | VARCHAR(100) | |
| `pincode` | VARCHAR(10) | |
| `rating` | DECIMAL(3,2) | 0.00 – 5.00, platform-computed |
| `is_verified` | BOOLEAN | KYC verified by platform |
| `is_active` | BOOLEAN | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

---

### 2.3 Product
```
A product listed by a seller with physical attributes used in shipping cost computation.
Key insight: volumetric weight must be stored/computed — airlines and logistics 
companies always bill on max(actual, volumetric). This prevents undercharging for 
bulky-light items like chips packets.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | |
| `seller_id` | BIGINT FK | → sellers.id |
| `name` | VARCHAR(255) | Product name |
| `description` | TEXT | |
| `sku` | VARCHAR(100) | Seller's internal SKU (unique per seller) |
| `category` | ENUM | `rice`, `pulses`, `masala`, `beverages`, `snacks`, `dairy`, `oil`, `flour`, `sugar`, `other` |
| `mrp` | DECIMAL(10,2) | Maximum retail price |
| `selling_price` | DECIMAL(10,2) | Platform selling price |
| `bulk_price` | DECIMAL(10,2) | Price for orders above min_order_qty |
| `actual_weight_kg` | DECIMAL(8,3) | Actual weight |
| `length_cm` | DECIMAL(8,2) | Packaging length |
| `width_cm` | DECIMAL(8,2) | Packaging width |
| `height_cm` | DECIMAL(8,2) | Packaging height |
| `volumetric_weight_kg` | DECIMAL(8,3) | Computed: (L×W×H)/5000 — stored for performance |
| `is_fragile` | BOOLEAN | Affects handling, future surcharge |
| `is_perishable` | BOOLEAN | May force express delivery |
| `stock_quantity` | INT | Current inventory |
| `min_order_quantity` | INT | Minimum B2B order quantity |
| `is_active` | BOOLEAN | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

---

### 2.4 Warehouse
```
Marketplace-owned distribution centers. The backbone of the routing algorithm.
Creative additions: serviceable_pincodes (delivery coverage), operating_hours,
current_load_percent — all needed in a real warehouse management scenario.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | |
| `name` | VARCHAR(255) | e.g. `BLR_Warehouse` |
| `code` | VARCHAR(50) | Unique short code e.g. `BLR`, `MUMB` |
| `lat` | DECIMAL(10,7) | Warehouse latitude |
| `lng` | DECIMAL(10,7) | Warehouse longitude |
| `address_line1` | VARCHAR(255) | |
| `city` | VARCHAR(100) | |
| `state` | VARCHAR(100) | |
| `pincode` | VARCHAR(10) | |
| `contact_person` | VARCHAR(255) | |
| `contact_phone` | VARCHAR(15) | |
| `max_capacity_sqft` | INT | Maximum floor space |
| `current_load_percent` | DECIMAL(5,2) | 0–100%, operational status |
| `serviceable_states` | TEXT[] | States this WH can deliver to |
| `operating_hours` | JSONB | `{"open": "08:00", "close": "20:00"}` |
| `is_active` | BOOLEAN | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

---

### 2.5 Order
```
Ties customer, seller, and product together. Created at checkout.
Creative additions: payment_mode (COD is huge for Kirana), 
estimated_delivery_date, tracking_id — all standard e-commerce must-haves.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | |
| `customer_id` | BIGINT FK | → customers.id |
| `seller_id` | BIGINT FK | → sellers.id |
| `nearest_warehouse_id` | BIGINT FK | → warehouses.id |
| `product_id` | BIGINT FK | → products.id |
| `quantity` | INT | |
| `unit_price` | DECIMAL(10,2) | Price at time of order |
| `total_product_amount` | DECIMAL(12,2) | unit_price × quantity |
| `shipping_charge` | DECIMAL(10,2) | Final shipping amount |
| `total_amount` | DECIMAL(12,2) | product_amount + shipping_charge |
| `delivery_speed` | ENUM | `standard`, `express` |
| `status` | ENUM | `placed`, `warehouse_received`, `in_transit`, `out_for_delivery`, `delivered`, `cancelled` |
| `payment_mode` | ENUM | `prepaid`, `cod`, `credit` |
| `tracking_id` | VARCHAR(100) | Unique tracking identifier |
| `estimated_delivery_date` | DATE | |
| `actual_delivery_date` | DATE | |
| `notes` | TEXT | Special instructions from Kirana owner |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

---

### 2.6 Shipment
```
The logistics leg from warehouse to customer.
Stores the computed distance and transport mode at time of dispatch — 
important because rates may change but historical charge must remain accurate.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | |
| `order_id` | BIGINT FK | → orders.id |
| `source_warehouse_id` | BIGINT FK | → warehouses.id |
| `destination_customer_id` | BIGINT FK | → customers.id |
| `distance_km` | DECIMAL(10,3) | Computed at dispatch |
| `transport_mode` | ENUM | `aeroplane`, `truck`, `minivan` |
| `billable_weight_kg` | DECIMAL(8,3) | max(actual, volumetric) at time of order |
| `rate_per_km_per_kg` | DECIMAL(8,4) | Locked rate at time of dispatch |
| `base_courier_charge` | DECIMAL(8,2) | Rs 10 |
| `distance_charge` | DECIMAL(10,2) | rate × distance × weight |
| `express_charge` | DECIMAL(10,2) | 0 if standard, 1.2×weight if express |
| `total_shipping_charge` | DECIMAL(10,2) | Sum of above |
| `status` | ENUM | `pending`, `dispatched`, `in_transit`, `delivered` |
| `dispatched_at` | TIMESTAMP | |
| `delivered_at` | TIMESTAMP | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

---

### 2.7 TransportRate (Config Entity)
```
Stores transport mode pricing rules in DB — no hardcoding.
Allows rate changes to be deployed via admin API without touching application code.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | |
| `mode` | ENUM | `aeroplane`, `truck`, `minivan` |
| `min_distance_km` | DECIMAL(10,3) | Inclusive lower bound |
| `max_distance_km` | DECIMAL(10,3) | Exclusive upper bound (NULL = infinity) |
| `rate_per_km_per_kg` | DECIMAL(8,4) | Base rate |
| `effective_from` | DATE | Rate valid from |
| `effective_to` | DATE | NULL = currently active |
| `is_active` | BOOLEAN | |
| `created_at` | TIMESTAMP | |

---

### 2.8 DeliverySpeedConfig (Config Entity)
```
Stores delivery speed surcharge rules in DB.
Making this configurable allows new speed tiers (e.g., "Same Day") 
without code changes.
```

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT PK | |
| `speed` | ENUM | `standard`, `express` |
| `base_courier_charge` | DECIMAL(8,2) | Flat Rs 10 for all speeds |
| `extra_charge_per_kg` | DECIMAL(8,4) | Rs 0 for standard, Rs 1.2 for express |
| `is_active` | BOOLEAN | |
| `created_at` | TIMESTAMP | |

---

## 3. Database Schema

```sql
-- Customers (Kirana Stores)
CREATE TABLE customers (
    id                      BIGSERIAL PRIMARY KEY,
    name                    VARCHAR(255)    NOT NULL,
    owner_name              VARCHAR(255)    NOT NULL,
    phone                   VARCHAR(15)     NOT NULL UNIQUE,
    email                   VARCHAR(255)    UNIQUE,
    gst_number              VARCHAR(20)     UNIQUE,
    store_type              VARCHAR(50)     NOT NULL DEFAULT 'general',
    credit_limit            DECIMAL(12,2)   NOT NULL DEFAULT 0,
    preferred_delivery_slot VARCHAR(50),
    lat                     DECIMAL(10,7)   NOT NULL,
    lng                     DECIMAL(10,7)   NOT NULL,
    address_line1           VARCHAR(255),
    city                    VARCHAR(100),
    state                   VARCHAR(100),
    pincode                 VARCHAR(10),
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Sellers
CREATE TABLE sellers (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(255)    NOT NULL,
    owner_name      VARCHAR(255)    NOT NULL,
    phone           VARCHAR(15)     NOT NULL UNIQUE,
    email           VARCHAR(255)    UNIQUE,
    gst_number      VARCHAR(20)     UNIQUE,
    business_type   VARCHAR(50)     NOT NULL DEFAULT 'wholesaler',
    lat             DECIMAL(10,7)   NOT NULL,
    lng             DECIMAL(10,7)   NOT NULL,
    address_line1   VARCHAR(255),
    city            VARCHAR(100),
    state           VARCHAR(100),
    pincode         VARCHAR(10),
    rating          DECIMAL(3,2)    NOT NULL DEFAULT 0.0,
    is_verified     BOOLEAN         NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Products
CREATE TABLE products (
    id                      BIGSERIAL PRIMARY KEY,
    seller_id               BIGINT          NOT NULL REFERENCES sellers(id),
    name                    VARCHAR(255)    NOT NULL,
    description             TEXT,
    sku                     VARCHAR(100)    NOT NULL,
    category                VARCHAR(50)     NOT NULL DEFAULT 'other',
    mrp                     DECIMAL(10,2)   NOT NULL,
    selling_price           DECIMAL(10,2)   NOT NULL,
    bulk_price              DECIMAL(10,2),
    actual_weight_kg        DECIMAL(8,3)    NOT NULL,
    length_cm               DECIMAL(8,2)    NOT NULL,
    width_cm                DECIMAL(8,2)    NOT NULL,
    height_cm               DECIMAL(8,2)    NOT NULL,
    volumetric_weight_kg    DECIMAL(8,3)    GENERATED ALWAYS AS
                                ((length_cm * width_cm * height_cm) / 5000) STORED,
    is_fragile              BOOLEAN         NOT NULL DEFAULT FALSE,
    is_perishable           BOOLEAN         NOT NULL DEFAULT FALSE,
    stock_quantity          INT             NOT NULL DEFAULT 0,
    min_order_quantity      INT             NOT NULL DEFAULT 1,
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    UNIQUE(seller_id, sku)
);

-- Warehouses
CREATE TABLE warehouses (
    id                      BIGSERIAL PRIMARY KEY,
    name                    VARCHAR(255)    NOT NULL UNIQUE,
    code                    VARCHAR(50)     NOT NULL UNIQUE,
    lat                     DECIMAL(10,7)   NOT NULL,
    lng                     DECIMAL(10,7)   NOT NULL,
    address_line1           VARCHAR(255),
    city                    VARCHAR(100),
    state                   VARCHAR(100),
    pincode                 VARCHAR(10),
    contact_person          VARCHAR(255),
    contact_phone           VARCHAR(15),
    max_capacity_sqft       INT,
    current_load_percent    DECIMAL(5,2)    DEFAULT 0,
    serviceable_states      TEXT[],
    operating_hours         JSONB,
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Transport Rate Config
CREATE TABLE transport_rates (
    id                  BIGSERIAL PRIMARY KEY,
    mode                VARCHAR(20)     NOT NULL,   -- aeroplane | truck | minivan
    min_distance_km     DECIMAL(10,3)   NOT NULL,
    max_distance_km     DECIMAL(10,3),              -- NULL = infinity
    rate_per_km_per_kg  DECIMAL(8,4)    NOT NULL,
    effective_from      DATE            NOT NULL DEFAULT CURRENT_DATE,
    effective_to        DATE,
    is_active           BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Delivery Speed Config
CREATE TABLE delivery_speed_configs (
    id                      BIGSERIAL PRIMARY KEY,
    speed                   VARCHAR(20)     NOT NULL UNIQUE,   -- standard | express
    base_courier_charge     DECIMAL(8,2)    NOT NULL DEFAULT 10.00,
    extra_charge_per_kg     DECIMAL(8,4)    NOT NULL DEFAULT 0.00,
    is_active               BOOLEAN         NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Orders
CREATE TABLE orders (
    id                      BIGSERIAL PRIMARY KEY,
    customer_id             BIGINT          NOT NULL REFERENCES customers(id),
    seller_id               BIGINT          NOT NULL REFERENCES sellers(id),
    nearest_warehouse_id    BIGINT          NOT NULL REFERENCES warehouses(id),
    product_id              BIGINT          NOT NULL REFERENCES products(id),
    quantity                INT             NOT NULL DEFAULT 1,
    unit_price              DECIMAL(10,2)   NOT NULL,
    total_product_amount    DECIMAL(12,2)   NOT NULL,
    shipping_charge         DECIMAL(10,2)   NOT NULL,
    total_amount            DECIMAL(12,2)   NOT NULL,
    delivery_speed          VARCHAR(20)     NOT NULL,
    status                  VARCHAR(30)     NOT NULL DEFAULT 'placed',
    payment_mode            VARCHAR(20)     NOT NULL DEFAULT 'prepaid',
    tracking_id             VARCHAR(100)    UNIQUE,
    estimated_delivery_date DATE,
    actual_delivery_date    DATE,
    notes                   TEXT,
    created_at              TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Shipments
CREATE TABLE shipments (
    id                          BIGSERIAL PRIMARY KEY,
    order_id                    BIGINT          NOT NULL REFERENCES orders(id),
    source_warehouse_id         BIGINT          NOT NULL REFERENCES warehouses(id),
    destination_customer_id     BIGINT          NOT NULL REFERENCES customers(id),
    distance_km                 DECIMAL(10,3)   NOT NULL,
    transport_mode              VARCHAR(20)     NOT NULL,
    billable_weight_kg          DECIMAL(8,3)    NOT NULL,
    rate_per_km_per_kg          DECIMAL(8,4)    NOT NULL,
    base_courier_charge         DECIMAL(8,2)    NOT NULL,
    distance_charge             DECIMAL(10,2)   NOT NULL,
    express_charge              DECIMAL(10,2)   NOT NULL DEFAULT 0,
    total_shipping_charge       DECIMAL(10,2)   NOT NULL,
    status                      VARCHAR(20)     NOT NULL DEFAULT 'pending',
    dispatched_at               TIMESTAMP,
    delivered_at                TIMESTAMP,
    created_at                  TIMESTAMP       NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP       NOT NULL DEFAULT NOW()
);

-- Seed: Transport Rates
INSERT INTO transport_rates (mode, min_distance_km, max_distance_km, rate_per_km_per_kg)
VALUES
    ('minivan',    0,   100,  3.0000),
    ('truck',      100, 500,  2.0000),
    ('aeroplane',  500, NULL, 1.0000);

-- Seed: Delivery Speed Configs
INSERT INTO delivery_speed_configs (speed, base_courier_charge, extra_charge_per_kg)
VALUES
    ('standard', 10.00, 0.0000),
    ('express',  10.00, 1.2000);

-- Seed: Warehouses
INSERT INTO warehouses (name, code, lat, lng, city, state, is_active)
VALUES
    ('BLR_Warehouse',  'BLR',  12.99999, 37.923273, 'Bengaluru', 'Karnataka', TRUE),
    ('MUMB_Warehouse', 'MUMB', 11.99999, 27.923273, 'Mumbai',    'Maharashtra', TRUE);
```

---

## 4. Data Flow

### Flow 1: GET Nearest Warehouse

```
Client
  │  GET /api/v1/warehouse/nearest?sellerId=123&productId=456
  ▼
[Auth Middleware] → Validate JWT
  ▼
[Validation Middleware] → sellerId > 0, productId > 0
  ▼
[WarehouseHandler.GetNearest()]
  ▼
[WarehouseService.FindNearest(sellerID, productID)]
  │
  ├── Check Redis: key = "nearest_wh:{sellerID}"
  │       HIT  → return cached response (< 20ms)
  │       MISS → continue
  │
  ├── [SellerRepository.GetByID(sellerID)] → validate seller exists
  ├── [ProductRepository.GetByID(productID)] → validate product & ownership
  ├── [WarehouseRepository.GetAllActive()] → fetch all active warehouses
  │
  ├── [GeoService.FindNearest(seller.lat, seller.lng, warehouses[])]
  │       Loop all warehouses:
  │           dist = Haversine(seller.lat, seller.lng, wh.lat, wh.lng)
  │       Return warehouse with minimum dist
  │
  ├── Set Redis: "nearest_wh:{sellerID}" TTL 10min
  └── Return 200 { warehouseId, warehouseName, warehouseLocation, distanceKm }
```

---

### Flow 2: GET Shipping Charge

```
Client
  │  GET /api/v1/shipping-charge?warehouseId=789&customerId=456
  │                              &productId=123&deliverySpeed=standard
  ▼
[Auth + Validation Middleware]
  ▼
[ShippingHandler.GetCharge()]
  ▼
[ShippingService.CalculateCharge(warehouseID, customerID, productID, speed)]
  │
  ├── Check Redis: key = "shipping:{warehouseID}:{customerID}:{productID}:{speed}"
  │       HIT → return cached response
  │       MISS → continue
  │
  ├── [WarehouseRepository.GetByID(warehouseID)]
  ├── [CustomerRepository.GetByID(customerID)]
  ├── [ProductRepository.GetByID(productID)]
  │
  ├── [GeoService.Distance(wh.lat, wh.lng, customer.lat, customer.lng)]
  │       → distanceKm (Haversine)
  │
  ├── [TransportModeStrategy.Select(distanceKm)]
  │       distanceKm < 100  → Minivan (rate = 3.0)
  │       distanceKm < 500  → Truck   (rate = 2.0)
  │       distanceKm >= 500 → Aeroplane(rate = 1.0)
  │       [rates fetched from DB config, cached in-memory]
  │
  ├── billableWeight = max(product.actual_weight_kg, product.volumetric_weight_kg)
  │
  ├── [DeliverySpeedStrategy.Calculate(speed, distanceKm, billableWeight, rate)]
  │       Standard: 10 + (rate × dist × weight)
  │       Express:  10 + (rate × dist × weight) + (1.2 × weight)
  │
  ├── Set Redis: key TTL 5min
  └── Return 200 { shippingCharge, breakdown: {...} }
```

---

### Flow 3: POST Calculate (Combined)

```
Client
  │  POST /api/v1/shipping-charge/calculate
  │  Body: { sellerId, customerId, productId, deliverySpeed }
  ▼
[Auth + Validation Middleware]
  ▼
[ShippingHandler.Calculate()]
  ▼
[ShippingService.CalculateFull(sellerID, customerID, productID, speed)]
  │
  ├── Step 1: WarehouseService.FindNearest(sellerID, productID)
  │           → nearestWarehouse  [Redis/DB as in Flow 1]
  │
  ├── Step 2: ShippingService.CalculateCharge(
  │               nearestWarehouse.ID, customerID, productID, speed)
  │           → shippingCharge, breakdown  [Redis/DB as in Flow 2]
  │
  └── Return 200 {
          shippingCharge,
          breakdown,
          nearestWarehouse: { warehouseId, warehouseName, warehouseLocation }
      }
```

---

## 5. High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                        Clients                                   │
│              (Mobile App / Web Dashboard / 3rd party)            │
└────────────────────────────┬─────────────────────────────────────┘
                             │  HTTPS / REST
                             ▼
┌──────────────────────────────────────────────────────────────────┐
│               API Gateway / Nginx (Load Balancer)                │
│            [TLS Termination] [Rate Limiting] [Routing]           │
└────────────────────────────┬─────────────────────────────────────┘
                             │
                             ▼
┌──────────────────────────────────────────────────────────────────┐
│                 Go REST API Server (Gin) — Stateless             │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Middleware Chain                      │    │
│  │  [RequestID] → [Logger] → [Auth(JWT)] → [Validator]     │    │
│  │  → [RateLimiter] → [Recover(panic)]                     │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌───────────────┐  ┌───────────────┐  ┌──────────────────┐    │
│  │   Warehouse   │  │   Shipping    │  │  Health / Metrics│    │
│  │   Handler     │  │   Handler     │  │  Handler         │    │
│  └───────┬───────┘  └──────┬────────┘  └──────────────────┘    │
│          │                 │                                     │
│  ┌───────▼─────────────────▼──────────────────────────────┐    │
│  │                    Service Layer                        │    │
│  │                                                         │    │
│  │  ┌─────────────────┐   ┌──────────────────────────┐   │    │
│  │  │ Warehouse       │   │ Shipping Service          │   │    │
│  │  │ Service         │   │ ┌────────────────────┐   │   │    │
│  │  │                 │   │ │ TransportMode      │   │   │    │
│  │  │                 │   │ │ Strategy           │   │   │    │
│  │  │                 │   │ │ (Van|Truck|Plane)  │   │   │    │
│  │  │                 │   │ ├────────────────────┤   │   │    │
│  │  │                 │   │ │ DeliverySpeed      │   │   │    │
│  │  │                 │   │ │ Strategy           │   │   │    │
│  │  │                 │   │ │ (Std|Express)      │   │   │    │
│  │  └────────┬────────┘   └──────────┬───────────┘   │   │    │
│  │           │                       │               │    │    │
│  │  ┌────────▼───────────────────────▼──────────┐   │    │    │
│  │  │              Geo Service                  │   │    │    │
│  │  │         (Haversine Distance Calc)         │   │    │    │
│  │  └────────────────────────────────────────────┘   │    │    │
│  └────────────────────────┬───────────────────────────┘    │    │
│                           │                                 │    │
│  ┌────────────────────────▼───────────────────────────┐    │    │
│  │                  Repository Layer                  │    │    │
│  │  [WarehouseRepo] [CustomerRepo] [ProductRepo]      │    │    │
│  │  [SellerRepo] [ShipmentRepo] [OrderRepo]           │    │    │
│  │  [TransportRateRepo] [DeliverySpeedConfigRepo]     │    │    │
│  └──────────┬─────────────────────┬────────────────────┘    │    │
└─────────────┼─────────────────────┼──────────────────────────┘    │
              │                     │
              ▼                     ▼
┌─────────────────────┐   ┌─────────────────────┐
│   PostgreSQL 16     │   │     Redis 7          │
│   (Primary DB)      │   │     (Cache Layer)    │
│                     │   │                      │
│  - customers        │   │  nearest_wh:{sid}    │
│  - sellers          │   │  TTL: 10 min         │
│  - products         │   │                      │
│  - warehouses       │   │  shipping:{wid}:     │
│  - orders           │   │  {cid}:{pid}:{speed} │
│  - shipments        │   │  TTL: 5 min          │
│  - transport_rates  │   │                      │
│  - speed_configs    │   │  transport_rates     │
│                     │   │  TTL: 60 min         │
└─────────────────────┘   └─────────────────────┘
```

---

## 6. Project Structure

```
jambotails/
│
├── cmd/
│   └── server/
│       └── main.go                  ← App entry point: init config, DB, Redis, router
│
├── config/
│   └── config.go                    ← Viper config loading from env/.env
│
├── internal/
│   ├── handlers/
│   │   ├── warehouse_handler.go     ← GET /api/v1/warehouse/nearest
│   │   ├── shipping_handler.go      ← GET /api/v1/shipping-charge
│   │   │                               POST /api/v1/shipping-charge/calculate
│   │   └── health_handler.go        ← GET /health
│   │
│   ├── services/
│   │   ├── warehouse_service.go     ← Find nearest warehouse logic
│   │   ├── shipping_service.go      ← Orchestrate charge calculation
│   │   ├── geo/
│   │   │   └── haversine.go         ← Haversine distance algorithm
│   │   ├── transport/
│   │   │   ├── strategy.go          ← TransportStrategy interface
│   │   │   ├── minivan.go           ← Mini Van implementation
│   │   │   ├── truck.go             ← Truck implementation
│   │   │   ├── aeroplane.go         ← Aeroplane implementation
│   │   │   └── factory.go           ← Select strategy by distance
│   │   └── pricing/
│   │       ├── strategy.go          ← PricingStrategy interface
│   │       ├── standard.go          ← Standard delivery pricing
│   │       ├── express.go           ← Express delivery pricing
│   │       └── factory.go           ← Select strategy by speed string
│   │
│   ├── repositories/
│   │   ├── interfaces.go            ← All repo interfaces (for mocking)
│   │   ├── warehouse_repo.go        ← PostgreSQL implementation
│   │   ├── customer_repo.go
│   │   ├── seller_repo.go
│   │   ├── product_repo.go
│   │   ├── order_repo.go
│   │   ├── shipment_repo.go
│   │   └── rate_config_repo.go      ← TransportRate + SpeedConfig
│   │
│   ├── models/
│   │   ├── customer.go
│   │   ├── seller.go
│   │   ├── product.go
│   │   ├── warehouse.go
│   │   ├── order.go
│   │   ├── shipment.go
│   │   └── config.go                ← TransportRate, DeliverySpeedConfig
│   │
│   ├── middleware/
│   │   ├── auth.go                  ← JWT validation
│   │   ├── logger.go                ← Request logging with zap
│   │   ├── rate_limiter.go          ← Token bucket rate limiter
│   │   ├── request_id.go            ← Inject X-Request-ID
│   │   └── recovery.go              ← Panic recovery → 500 response
│   │
│   └── cache/
│       ├── redis_client.go          ← Redis connection setup
│       └── cache_service.go         ← Get/Set/Delete with TTL abstraction
│
├── pkg/
│   ├── validator/
│   │   └── validator.go             ← Custom validators: lat range, lng range, enum
│   ├── errors/
│   │   └── errors.go               ← AppError struct, error codes
│   └── response/
│       └── response.go              ← Standardized JSON response wrapper
│
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   └── 000002_seed_data.up.sql
│
├── tests/
│   ├── unit/
│   │   ├── geo_test.go              ← Haversine formula tests
│   │   ├── transport_test.go        ← Strategy selection tests
│   │   ├── pricing_test.go          ← Charge calculation tests
│   │   └── warehouse_service_test.go
│   └── integration/
│       ├── warehouse_api_test.go
│       ├── shipping_api_test.go
│       └── testhelpers.go           ← Test DB setup, mock injection
│
├── docker-compose.yml               ← PostgreSQL + Redis + App
├── Dockerfile
├── go.mod
├── go.sum
├── .env.example
└── Makefile                         ← make run, make test, make migrate
```

---

## 7. API Design

### Standard Response Envelope

**Success:**
```json
{
  "success": true,
  "data": { ... },
  "requestId": "uuid-v4"
}
```

**Error:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "sellerId is required",
    "fields": { "sellerId": "required" }
  },
  "requestId": "uuid-v4"
}
```

---

### API 1: GET /api/v1/warehouse/nearest

**Query Params:** `sellerId` (int, required), `productId` (int, required)

**Success 200:**
```json
{
  "success": true,
  "data": {
    "warehouseId": 1,
    "warehouseName": "BLR_Warehouse",
    "warehouseLocation": { "lat": 12.99999, "lng": 37.923273 },
    "distanceKm": 48.35
  },
  "requestId": "a1b2c3d4-..."
}
```

**Error Cases:** 400 (missing/invalid params), 404 (seller/product not found), 503 (no active warehouses)

---

### API 2: GET /api/v1/shipping-charge

**Query Params:** `warehouseId`, `customerId`, `productId`, `deliverySpeed`

**Success 200:**
```json
{
  "success": true,
  "data": {
    "shippingCharge": 641.00,
    "breakdown": {
      "distanceKm": 320.5,
      "transportMode": "truck",
      "ratePerKmPerKg": 2.0,
      "billableWeightKg": 10.0,
      "baseCourierCharge": 10.00,
      "distanceCharge": 641.00,
      "expressCharge": 0.00
    }
  },
  "requestId": "..."
}
```

---

### API 3: POST /api/v1/shipping-charge/calculate

**Body:**
```json
{
  "sellerId": 123,
  "customerId": 456,
  "productId": 789,
  "deliverySpeed": "express"
}
```

**Success 200:**
```json
{
  "success": true,
  "data": {
    "shippingCharge": 653.00,
    "breakdown": {
      "distanceKm": 320.5,
      "transportMode": "truck",
      "ratePerKmPerKg": 2.0,
      "billableWeightKg": 10.0,
      "baseCourierCharge": 10.00,
      "distanceCharge": 641.00,
      "expressCharge": 12.00
    },
    "nearestWarehouse": {
      "warehouseId": 1,
      "warehouseName": "BLR_Warehouse",
      "warehouseLocation": { "lat": 12.99999, "lng": 37.923273 },
      "distanceKm": 48.35
    }
  },
  "requestId": "..."
}
```

---

## 8. Core Algorithms

### 8.1 Haversine Formula (Geo Distance)

Used to compute straight-line distance between two lat/lng coordinates in km.

```
R = 6371  (Earth's radius in km)

φ1, φ2 = lat1, lat2 in radians
Δφ = (lat2 - lat1) in radians
Δλ = (lng2 - lng1) in radians

a = sin²(Δφ/2) + cos(φ1) × cos(φ2) × sin²(Δλ/2)
c = 2 × atan2(√a, √(1−a))
d = R × c
```

### 8.2 Billable Weight

```
volumetric_weight = (length_cm × width_cm × height_cm) / 5000
billable_weight   = max(actual_weight_kg, volumetric_weight)
```

### 8.3 Shipping Charge Calculation

```
transport_rate = lookup(distance_km)        // via TransportModeStrategy
speed_config   = lookup(delivery_speed)     // via DeliverySpeedStrategy

distance_charge  = transport_rate × distance_km × billable_weight
express_charge   = speed_config.extra_per_kg × billable_weight
total_charge     = speed_config.base_courier + distance_charge + express_charge
```

---

## 9. Design Patterns Used

### Strategy Pattern — Transport Mode

```go
// TransportStrategy interface
type TransportStrategy interface {
    Name() string
    RatePerKmPerKg() float64
}

// Implementations: MiniVanStrategy, TruckStrategy, AeroplaneStrategy

// Factory selects strategy based on distance
func NewTransportStrategy(distanceKm float64, rates []TransportRate) TransportStrategy {
    for _, rate := range rates {
        if distanceKm >= rate.MinDistanceKm &&
           (rate.MaxDistanceKm == nil || distanceKm < *rate.MaxDistanceKm) {
            return strategyFor(rate.Mode, rate.RatePerKmPerKg)
        }
    }
    return DefaultAeroplaneStrategy{}
}
```

**Why:** Adding a new transport mode (e.g., Drone, Ship) = add a new struct + DB row. Zero changes to existing code.

---

### Strategy Pattern — Delivery Speed Pricing

```go
// PricingStrategy interface
type PricingStrategy interface {
    Calculate(distanceKm, billableWeightKg, ratePerKmPerKg float64) PricingBreakdown
}

// StandardPricing: 10 + (rate × dist × weight)
// ExpressPricing:  10 + (rate × dist × weight) + (1.2 × weight)
```

**Why:** Adding "Same Day" delivery = new struct. The handler/service never changes.

---

### Repository Pattern

```go
// Interface in repositories/interfaces.go
type WarehouseRepository interface {
    GetByID(ctx context.Context, id int64) (*models.Warehouse, error)
    GetAllActive(ctx context.Context) ([]models.Warehouse, error)
}

// PostgreSQL implementation in warehouse_repo.go
// Mock implementation in tests/
```

**Why:** Service layer depends on interface, not DB. Tests inject mocks without spinning up PostgreSQL.

---

## 10. Caching Strategy

| Cache Key | TTL | Invalidation |
|---|---|---|
| `nearest_wh:{sellerID}` | 10 min | Manual flush if seller location changes |
| `shipping:{whID}:{custID}:{prodID}:{speed}` | 5 min | TTL expiry only |
| `transport_rates` (in-memory) | 60 min | Server restart or admin flush endpoint |
| `speed_configs` (in-memory) | 60 min | Server restart or admin flush endpoint |

**Fallback behavior:** If Redis is unavailable, all operations fall through to PostgreSQL. A circuit breaker logs the Redis failure but does **not** return an error to the client — graceful degradation only.
