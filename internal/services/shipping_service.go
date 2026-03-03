package services

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/jambotails/shipping-service/internal/repositories"
	"github.com/jambotails/shipping-service/internal/services/geo"
	"github.com/jambotails/shipping-service/internal/services/pricing"
	"github.com/jambotails/shipping-service/internal/services/transport"
)

// ShippingChargeResult holds the result of a shipping charge calculation.
type ShippingChargeResult struct {
	ShippingCharge float64           `json:"shippingCharge"`
	Breakdown      pricing.Breakdown `json:"breakdown"`
}

// FullCalculationResult holds the combined result: shipping charge + nearest warehouse.
type FullCalculationResult struct {
	ShippingCharge   float64                 `json:"shippingCharge"`
	Breakdown        pricing.Breakdown       `json:"breakdown"`
	NearestWarehouse *NearestWarehouseResult `json:"nearestWarehouse"`
}

// ShippingService handles shipping charge business logic.
type ShippingService struct {
	warehouseRepo     repositories.WarehouseRepository
	customerRepo      repositories.CustomerRepository
	productRepo       repositories.ProductRepository
	transportRateRepo repositories.TransportRateRepository
	speedConfigRepo   repositories.DeliverySpeedConfigRepository
	warehouseSvc      *WarehouseService
}

// NewShippingService creates a new ShippingService.
func NewShippingService(
	warehouseRepo repositories.WarehouseRepository,
	customerRepo repositories.CustomerRepository,
	productRepo repositories.ProductRepository,
	transportRateRepo repositories.TransportRateRepository,
	speedConfigRepo repositories.DeliverySpeedConfigRepository,
	warehouseSvc *WarehouseService,
) *ShippingService {
	return &ShippingService{
		warehouseRepo:     warehouseRepo,
		customerRepo:      customerRepo,
		productRepo:       productRepo,
		transportRateRepo: transportRateRepo,
		speedConfigRepo:   speedConfigRepo,
		warehouseSvc:      warehouseSvc,
	}
}

// CalculateCharge computes the shipping charge from a warehouse to a customer for a product.
func (s *ShippingService) CalculateCharge(
	ctx context.Context,
	warehouseID, customerID, productID int64,
	deliverySpeed string,
) (*ShippingChargeResult, error) {
	// Validate delivery speed
	speed := strings.ToLower(strings.TrimSpace(deliverySpeed))
	if speed != "standard" && speed != "express" {
		return nil, fmt.Errorf("deliverySpeed must be 'standard' or 'express', got '%s'", deliverySpeed)
	}

	// Fetch warehouse
	warehouse, err := s.warehouseRepo.GetByID(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found: %w", err)
	}

	// Fetch customer
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Fetch product
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Calculate distance: warehouse → customer
	distanceKm := geo.Distance(warehouse.Lat, warehouse.Lng, customer.Lat, customer.Lng)
	distanceKm = math.Round(distanceKm*100) / 100

	// Get billable weight
	billableWeight := product.BillableWeightKg()

	// Get transport rates from DB
	rates, err := s.transportRateRepo.GetAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport rates: %w", err)
	}

	// Select transport strategy based on distance
	transportStrategy, err := transport.NewStrategy(distanceKm, rates)
	if err != nil {
		return nil, fmt.Errorf("failed to select transport mode: %w", err)
	}

	// Get delivery speed config from DB
	speedConfig, err := s.speedConfigRepo.GetBySpeed(ctx, speed)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery speed config: %w", err)
	}

	// Select pricing strategy
	pricingStrategy, err := pricing.NewStrategy(speedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing strategy: %w", err)
	}

	// Calculate charge
	breakdown := pricingStrategy.Calculate(
		distanceKm,
		billableWeight,
		transportStrategy.RatePerKmPerKg(),
		speedConfig.BaseCourierCharge,
	)

	// Fill in remaining breakdown fields
	breakdown.DistanceKm = distanceKm
	breakdown.TransportMode = transportStrategy.Name()
	breakdown.RatePerKmPerKg = transportStrategy.RatePerKmPerKg()
	breakdown.BillableWeightKg = billableWeight

	return &ShippingChargeResult{
		ShippingCharge: breakdown.TotalCharge,
		Breakdown:      breakdown,
	}, nil
}

// CalculateFull performs end-to-end calculation: find nearest warehouse + compute shipping charge.
func (s *ShippingService) CalculateFull(
	ctx context.Context,
	sellerID, customerID, productID int64,
	deliverySpeed string,
) (*FullCalculationResult, error) {
	// Step 1: Find nearest warehouse for the seller
	nearestWh, err := s.warehouseSvc.FindNearest(ctx, sellerID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest warehouse: %w", err)
	}

	// Step 2: Calculate shipping charge from that warehouse to the customer
	chargeResult, err := s.CalculateCharge(ctx, nearestWh.WarehouseID, customerID, productID, deliverySpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate shipping charge: %w", err)
	}

	return &FullCalculationResult{
		ShippingCharge:   chargeResult.ShippingCharge,
		Breakdown:        chargeResult.Breakdown,
		NearestWarehouse: nearestWh,
	}, nil
}
