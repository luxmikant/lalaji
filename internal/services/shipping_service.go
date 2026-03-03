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
	apperrors "github.com/jambotails/shipping-service/pkg/errors"
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
		return nil, apperrors.NewInvalidDeliverySpeedError(deliverySpeed)
	}

	// Fetch warehouse
	warehouse, err := s.warehouseRepo.GetByID(ctx, warehouseID)
	if err != nil {
		return nil, apperrors.NewWarehouseNotFoundError(warehouseID)
	}

	// Fetch customer
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, apperrors.NewCustomerNotFoundError(customerID)
	}

	// Fetch product
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, apperrors.NewProductNotFoundError(productID, 0)
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

	// Select transport strategy based on distance.
	// If no transport mode covers this distance, the delivery location is unsupported.
	transportStrategy, err := transport.NewStrategy(distanceKm, rates)
	if err != nil {
		return nil, apperrors.NewDeliveryUnsupportedError(distanceKm)
	}

	// Get delivery speed config from DB
	speedConfig, err := s.speedConfigRepo.GetBySpeed(ctx, speed)
	if err != nil {
		return nil, apperrors.NewTransportConfigError(fmt.Sprintf("delivery speed config for '%s' not found", speed))
	}

	// Select pricing strategy
	pricingStrategy, err := pricing.NewStrategy(speedConfig)
	if err != nil {
		return nil, apperrors.NewTransportConfigError("pricing strategy initialisation failed: " + err.Error())
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
	// Step 1: Find nearest warehouse for the seller.
	// Errors here are already typed AppErrors (seller/product not found, no active warehouses).
	nearestWh, err := s.warehouseSvc.FindNearest(ctx, sellerID, productID)
	if err != nil {
		return nil, err // propagate typed AppError as-is
	}

	// Step 2: Calculate shipping charge from that warehouse to the customer.
	// Errors here are already typed AppErrors (customer/product not found, unsupported distance).
	chargeResult, err := s.CalculateCharge(ctx, nearestWh.WarehouseID, customerID, productID, deliverySpeed)
	if err != nil {
		return nil, err // propagate typed AppError as-is
	}

	return &FullCalculationResult{
		ShippingCharge:   chargeResult.ShippingCharge,
		Breakdown:        chargeResult.Breakdown,
		NearestWarehouse: nearestWh,
	}, nil
}
