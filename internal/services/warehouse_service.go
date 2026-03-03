package services

import (
	"context"
	"fmt"
	"math"

	"github.com/jambotails/shipping-service/internal/models"
	"github.com/jambotails/shipping-service/internal/repositories"
	"github.com/jambotails/shipping-service/internal/services/geo"
)

// NearestWarehouseResult holds the response data for nearest warehouse lookup.
type NearestWarehouseResult struct {
	WarehouseID       int64   `json:"warehouseId"`
	WarehouseName     string  `json:"warehouseName"`
	WarehouseLocation LatLng  `json:"warehouseLocation"`
	DistanceKm        float64 `json:"distanceKm"`
}

// LatLng represents a geographic coordinate pair.
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// WarehouseService handles warehouse-related business logic.
type WarehouseService struct {
	warehouseRepo repositories.WarehouseRepository
	sellerRepo    repositories.SellerRepository
	productRepo   repositories.ProductRepository
}

// NewWarehouseService creates a new WarehouseService.
func NewWarehouseService(
	warehouseRepo repositories.WarehouseRepository,
	sellerRepo repositories.SellerRepository,
	productRepo repositories.ProductRepository,
) *WarehouseService {
	return &WarehouseService{
		warehouseRepo: warehouseRepo,
		sellerRepo:    sellerRepo,
		productRepo:   productRepo,
	}
}

// FindNearest finds the nearest active warehouse to a given seller.
// Validates that the seller and product exist and that the product belongs to the seller.
func (s *WarehouseService) FindNearest(ctx context.Context, sellerID, productID int64) (*NearestWarehouseResult, error) {
	// Validate seller exists
	seller, err := s.sellerRepo.GetByID(ctx, sellerID)
	if err != nil {
		return nil, fmt.Errorf("seller not found: %w", err)
	}

	// Validate product exists and belongs to seller
	_, err = s.productRepo.GetByIDAndSellerID(ctx, productID, sellerID)
	if err != nil {
		return nil, fmt.Errorf("product not found for seller: %w", err)
	}

	// Fetch all active warehouses
	warehouses, err := s.warehouseRepo.GetAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch warehouses: %w", err)
	}
	if len(warehouses) == 0 {
		return nil, fmt.Errorf("no active warehouses available")
	}

	// Find the nearest warehouse using Haversine distance
	var nearest *models.Warehouse
	minDist := math.MaxFloat64

	for i := range warehouses {
		dist := geo.Distance(seller.Lat, seller.Lng, warehouses[i].Lat, warehouses[i].Lng)
		if dist < minDist {
			minDist = dist
			nearest = &warehouses[i]
		}
	}

	return &NearestWarehouseResult{
		WarehouseID:   nearest.ID,
		WarehouseName: nearest.Name,
		WarehouseLocation: LatLng{
			Lat: nearest.Lat,
			Lng: nearest.Lng,
		},
		DistanceKm: math.Round(minDist*100) / 100,
	}, nil
}
