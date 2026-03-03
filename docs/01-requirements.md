# Requirements — B2B Kirana Shipping Charge Estimator

> **Project:** Jambotails — B2B e-Commerce Shipping Module  
> **Tech Stack:** Go (Gin), PostgreSQL, Redis  
> **Author:** Backend Engineering  
> **Date:** March 3, 2026  
> **Status:** Draft v1.0

---

## Table of Contents

1. [Problem Statement](#1-problem-statement)
2. [User Context](#2-user-context)
3. [Actors & Roles](#3-actors--roles)
4. [Functional Requirements](#4-functional-requirements)
5. [Non-Functional Requirements](#5-non-functional-requirements)
6. [Business Rules](#6-business-rules)
7. [Input Validation Rules](#7-input-validation-rules)
8. [Error Handling Requirements](#8-error-handling-requirements)
9. [Out of Scope](#9-out-of-scope)

---

## 1. Problem Statement

Build a backend API system that calculates the **shipping charge for delivering a product** in a
B2B e-commerce marketplace focused on Kirana (small retail) stores.

The system must:
- Identify the **nearest warehouse** for a seller to drop their products
- Calculate **shipping charges** from warehouse to customer based on distance and transport mode
- Factor in **delivery speed preferences** (Standard vs Express)
- Be **scalable, validated, and extensible**

---

## 2. User Context

```
B2B e-Commerce Platform
       │
       ▼
User (Kirana Store Owner)
       │
       ▼
Registers their Shop (Customer entity created)
       │
       ▼
Browses Product Catalog → Adds to Cart
       │
       ├──────────────────────────────────────────────────┐
       │                                                  ▼
       │                               [Delivery Charge Calculation Engine]
       │                               - Find nearest warehouse for seller
       │                               - Calculate distance: warehouse → customer
       │                               - Apply transport mode rates
       │                               - Apply delivery speed surcharge
       │                                                  │
       ▼                                                  │
Checkout Gateway ◀─────────────────────────────────────────┘
       │
       ▼
Order + Profile Data Stored
```

### Who is a Kirana Store Owner?
- Runs a small general/grocery store in a locality
- Orders in **bulk quantities** (B2B, not individual consumer)
- Has a **GST number** for business transactions
- Needs **reliable, predictable delivery charges** to plan their margins
- Cares about **delivery speed** — perishable items need express; bulk rice can be standard
- Wants to know **which warehouse** their order is coming from

---

## 3. Actors & Roles

| Actor | Description | System Role |
|---|---|---|
| **Kirana Store Owner (Customer)** | Registered shop owner who orders products | Places orders, triggers shipping charge calc |
| **Seller** | Wholesale supplier (e.g., Nestle, Rice Seller) | Drops product at nearest warehouse |
| **Warehouse** | Marketplace-owned distribution center | Intermediate storage + dispatch point |
| **Platform (System)** | Our B2B marketplace | Orchestrates warehouse routing + pricing |
| **Admin** | Internal team | Manages transport rates, warehouse config |

---

## 4. Functional Requirements

### FR-01: Customer Registration
**ID:** FR-01  
**Priority:** Must Have  
**Description:** A Kirana store owner must be able to register their shop on the platform.

**Derived entities needed:** `Customer`, including location, GST, store type, contact, address.

---

### FR-02: Seller & Product Management
**ID:** FR-02  
**Priority:** Must Have  
**Description:** The system must store seller information and their products, including weight and dimensional attributes used in pricing.

**Derived calculation:** VolumetricWeight = `(length × width × height) / 5000`  
**Used weight in pricing:** `max(actual_weight, volumetric_weight)` — standard logistics practice.

---

### FR-03: Get Nearest Warehouse for a Seller
**ID:** FR-03  
**Priority:** Must Have  
**API:** `GET /api/v1/warehouse/nearest?sellerId={id}&productId={id}`

**Description:** Given a seller's location, return the warehouse with the shortest straight-line (Haversine) distance from the seller.

**Constraints:**
- Must return the single nearest warehouse
- If no warehouses exist or are active → return `404` with a clear error
- Seller must exist in the system — else `404`
- Product must belong to the seller — else `400`

**Sample Response:**
```json
{
  "warehouseId": 789,
  "warehouseName": "BLR_Warehouse",
  "warehouseLocation": { "lat": 12.99999, "lng": 37.923273 },
  "distanceKm": 48.3
}
```

---

### FR-04: Get Shipping Charge from Warehouse to Customer
**ID:** FR-04  
**Priority:** Must Have  
**API:** `GET /api/v1/shipping-charge?warehouseId={id}&customerId={id}&deliverySpeed={speed}&productId={id}`

**Description:** Calculate the shipping charge for moving a product from a warehouse to a customer's location, based on:
1. Distance (Haversine: warehouse → customer)
2. Product's billable weight
3. Applicable transport mode (distance-driven)
4. Delivery speed surcharge

**Transport Mode Selection:**

| Distance | Mode | Rate |
|---|---|---|
| 0 – 99.99 km | Mini Van | Rs 3 / km / kg |
| 100 – 499.99 km | Truck | Rs 2 / km / kg |
| 500 km+ | Aeroplane | Rs 1 / km / kg |

**Delivery Speed Pricing:**

| Speed | Formula |
|---|---|
| Standard | Rs 10 (base) + (rate × distance × weight) |
| Express | Rs 10 (base) + (rate × distance × weight) + (Rs 1.2 × weight) |

**Constraints:**
- `deliverySpeed` must be one of: `standard`, `express` (case-insensitive)
- Warehouse must exist → else `404`
- Customer must exist → else `404`
- Product must be provided (needed for weight) → else `400`

---

### FR-05: End-to-End Shipping Charge Calculation (Seller → Customer)
**ID:** FR-05  
**Priority:** Must Have  
**API:** `POST /api/v1/shipping-charge/calculate`

**Description:** Calculate the full shipping charge by internally chaining FR-03 and FR-04:
- Step 1: Find nearest warehouse for seller
- Step 2: Calculate shipping from that warehouse to customer
- Return combined result

**Request Body:**
```json
{
  "sellerId": 123,
  "customerId": 456,
  "productId": 789,
  "deliverySpeed": "express"
}
```

**Response:**
```json
{
  "shippingCharge": 180.00,
  "breakdown": {
    "distanceKm": 320.5,
    "transportMode": "truck",
    "billableWeightKg": 10,
    "baseCharge": 10.00,
    "distanceCharge": 641.00,
    "expressCharge": 12.00
  },
  "nearestWarehouse": {
    "warehouseId": 789,
    "warehouseName": "BLR_Warehouse",
    "warehouseLocation": { "lat": 12.99999, "lng": 37.923273 }
  }
}
```

---

### FR-06: Rate Configuration Management (Admin)
**ID:** FR-06  
**Priority:** Good to Have  
**Description:** Transport rates and delivery speed charges should be stored in the DB and manageable without code changes. Admin APIs to CRUD these configs.

---

### FR-07: Response Caching
**ID:** FR-07  
**Priority:** Good to Have  
**Description:**
- Cache result of FR-03 (nearest warehouse for seller) — key: `nearest_wh:{sellerId}`, TTL: 10 minutes
- Cache result of FR-04 (shipping charge) — key: `shipping:{warehouseId}:{customerId}:{speed}:{productId}`, TTL: 5 minutes

---

## 5. Non-Functional Requirements

### NFR-01: Performance
- Nearest warehouse query must complete in < **100ms** (DB + Haversine)
- Shipping charge calc must complete in < **150ms** end-to-end
- Cache hit response must be < **20ms**

### NFR-02: Scalability
- API server must be **stateless** — scale horizontally behind a load balancer
- DB must use **connection pooling** (max 25 connections per pod)
- Redis cache must reduce DB load by **> 60%** for repeated queries

### NFR-03: Availability
- Target **99.9% uptime** for shipping charge APIs (used at checkout — critical path)
- Graceful degradation: if Redis is down, fall through to DB without crashing

### NFR-04: Security
- All endpoints must require **JWT authentication**
- Input must be **sanitized** against SQL injection and XSS
- Rate limiting: max **100 requests / minute per IP**

### NFR-05: Maintainability
- Code must follow **clean architecture**: handler → service → repository
- All business rules (transport rates, speed surcharges) must be **configurable via DB**, not hardcoded
- New transport modes or delivery speeds must be addable **without modifying existing code** (Open/Closed Principle via Strategy Pattern)

### NFR-06: Testability
- Service layer must be **interface-driven** to allow mock repositories in tests
- Minimum **70% unit test coverage** on service + geo + pricing packages
- Integration tests for all three main API endpoints

### NFR-07: Observability
- Structured JSON logs using `zap` on every request and error
- Prometheus metrics: request count, latency histograms, cache hit ratio
- Request ID propagated through all log entries for traceability

### NFR-08: Data Integrity
- All order and shipment writes wrapped in **DB transactions**
- Location coordinates validated as real lat/lng ranges:
  - Latitude: -90 to 90
  - Longitude: -180 to 180

---

## 6. Business Rules

| Rule ID | Rule |
|---|---|
| BR-01 | Billable weight = `max(actual_weight_kg, volumetric_weight_kg)` where `volumetric = (L×W×H)/5000` |
| BR-02 | Transport mode is determined purely by distance (warehouse → customer), not seller proximity |
| BR-03 | Express surcharge is applied **per kg** of billable weight |
| BR-04 | Standard courier base charge (Rs 10) applies to **all** delivery speeds |
| BR-05 | If a seller has no nearby warehouse within India, return the globally nearest one |
| BR-06 | Distance calculation uses the **Haversine formula** (great-circle distance) |
| BR-07 | Transport rate boundaries are: Van [0, 100), Truck [100, 500), Aeroplane [500, ∞) |
| BR-08 | Rate configs are read from DB; cached in memory with 1-hour TTL on server startup |

---

## 7. Input Validation Rules

### GET /api/v1/warehouse/nearest
| Field | Type | Validation |
|---|---|---|
| `sellerId` | int | Required, > 0, must exist in DB |
| `productId` | int | Required, > 0, must belong to sellerId |

### GET /api/v1/shipping-charge
| Field | Type | Validation |
|---|---|---|
| `warehouseId` | int | Required, > 0, must exist in DB |
| `customerId` | int | Required, > 0, must exist in DB |
| `productId` | int | Required, > 0, must exist in DB |
| `deliverySpeed` | string | Required, enum: `standard` \| `express` |

### POST /api/v1/shipping-charge/calculate
| Field | Type | Validation |
|---|---|---|
| `sellerId` | int | Required, > 0, must exist in DB |
| `customerId` | int | Required, > 0, must exist in DB |
| `productId` | int | Required, > 0, must belong to sellerId |
| `deliverySpeed` | string | Required, enum: `standard` \| `express` |

---

## 8. Error Handling Requirements

| Scenario | HTTP Code | Response |
|---|---|---|
| Missing required param | 400 | `{ "error": "sellerId is required" }` |
| Invalid param type (string instead of int) | 400 | `{ "error": "sellerId must be a valid integer" }` |
| Invalid enum value for deliverySpeed | 400 | `{ "error": "deliverySpeed must be 'standard' or 'express'" }` |
| Seller not found | 404 | `{ "error": "seller not found" }` |
| Customer not found | 404 | `{ "error": "customer not found" }` |
| Warehouse not found | 404 | `{ "error": "warehouse not found" }` |
| Product not found or wrong seller | 404 / 403 | `{ "error": "product not found" }` |
| No active warehouses in system | 503 | `{ "error": "no active warehouses available" }` |
| DB error | 500 | `{ "error": "internal server error", "requestId": "xxx" }` |
| Unauthenticated | 401 | `{ "error": "unauthorized" }` |
| Rate limit exceeded | 429 | `{ "error": "too many requests" }` |

All error responses include a `requestId` field for traceability.

---

## 9. Out of Scope (v1.0)

- Real-time shipment tracking
- Payment processing
- Seller onboarding workflows
- Returns and refunds
- Multi-item cart shipping optimization
- International shipping
- Dynamic surge pricing
- SMS/email notifications
