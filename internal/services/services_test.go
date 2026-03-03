package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jambotails/shipping-service/internal/models"
	"github.com/jambotails/shipping-service/internal/services"
)

// ── Mock Repositories ────────────────────────────────────────

type mockWarehouseRepo struct {
	warehouses []models.Warehouse
	byID       map[int64]*models.Warehouse
	err        error
}

func (m *mockWarehouseRepo) GetByID(_ context.Context, id int64) (*models.Warehouse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if w, ok := m.byID[id]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("warehouse %d not found", id)
}

func (m *mockWarehouseRepo) GetAllActive(_ context.Context) ([]models.Warehouse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.warehouses, nil
}

type mockSellerRepo struct {
	sellers map[int64]*models.Seller
	err     error
}

func (m *mockSellerRepo) GetByID(_ context.Context, id int64) (*models.Seller, error) {
	if m.err != nil {
		return nil, m.err
	}
	if s, ok := m.sellers[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("seller %d not found", id)
}

type mockProductRepo struct {
	products       map[int64]*models.Product
	sellerProducts map[string]*models.Product // key = "productID-sellerID"
	err            error
}

func (m *mockProductRepo) GetByID(_ context.Context, id int64) (*models.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	if p, ok := m.products[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("product %d not found", id)
}

func (m *mockProductRepo) GetByIDAndSellerID(_ context.Context, productID, sellerID int64) (*models.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := fmt.Sprintf("%d-%d", productID, sellerID)
	if p, ok := m.sellerProducts[key]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("product %d not found for seller %d", productID, sellerID)
}

type mockCustomerRepo struct {
	customers map[int64]*models.Customer
	err       error
}

func (m *mockCustomerRepo) GetByID(_ context.Context, id int64) (*models.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	if c, ok := m.customers[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("customer %d not found", id)
}

type mockTransportRateRepo struct {
	rates []models.TransportRate
	err   error
}

func (m *mockTransportRateRepo) GetAllActive(_ context.Context) ([]models.TransportRate, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rates, nil
}

type mockSpeedConfigRepo struct {
	configs map[string]*models.DeliverySpeedConfig
	err     error
}

func (m *mockSpeedConfigRepo) GetBySpeed(_ context.Context, speed string) (*models.DeliverySpeedConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	if c, ok := m.configs[speed]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("speed config '%s' not found", speed)
}

// ── Test Data ────────────────────────────────────────────────

func testSeller() *models.Seller {
	return &models.Seller{
		ID:   1,
		Name: "Test Seller",
		Lat:  12.9716, // Bangalore
		Lng:  77.5946,
	}
}

func testProduct() *models.Product {
	return &models.Product{
		ID:             1,
		SellerID:       1,
		Name:           "Test Rice",
		ActualWeightKg: 10,
		LengthCm:       30,
		WidthCm:        20,
		HeightCm:       15,
		// Volumetric = 30*20*15/5000 = 1.8 kg → billable = 10 kg
	}
}

func testWarehouses() []models.Warehouse {
	return []models.Warehouse{
		{ID: 1, Name: "Bangalore WH", Lat: 12.9716, Lng: 77.5946, IsActive: true}, // 0 km from seller
		{ID: 2, Name: "Mumbai WH", Lat: 19.0760, Lng: 72.8777, IsActive: true},    // ~842 km
		{ID: 3, Name: "Delhi WH", Lat: 28.7041, Lng: 77.1025, IsActive: true},     // ~1745 km
	}
}

func testCustomer() *models.Customer {
	return &models.Customer{
		ID:   1,
		Name: "Test Kirana",
		Lat:  13.0827, // Chennai-ish
		Lng:  80.2707,
	}
}

func testTransportRates() []models.TransportRate {
	max100 := 100.0
	max500 := 500.0
	return []models.TransportRate{
		{Mode: "minivan", MinDistanceKm: 0, MaxDistanceKm: &max100, RatePerKmPerKg: 3.0},
		{Mode: "truck", MinDistanceKm: 100, MaxDistanceKm: &max500, RatePerKmPerKg: 2.0},
		{Mode: "aeroplane", MinDistanceKm: 500, MaxDistanceKm: nil, RatePerKmPerKg: 1.0},
	}
}

func testSpeedConfigs() map[string]*models.DeliverySpeedConfig {
	return map[string]*models.DeliverySpeedConfig{
		"standard": {Speed: "standard", BaseCourierCharge: 10, ExtraChargePerKg: 0},
		"express":  {Speed: "express", BaseCourierCharge: 10, ExtraChargePerKg: 1.2},
	}
}

// ── WarehouseService Tests ───────────────────────────────────

func TestFindNearest_ReturnsClosestWarehouse(t *testing.T) {
	whs := testWarehouses()
	whRepo := &mockWarehouseRepo{
		warehouses: whs,
		byID:       map[int64]*models.Warehouse{1: &whs[0], 2: &whs[1], 3: &whs[2]},
	}
	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	product := testProduct()
	productRepo := &mockProductRepo{
		products:       map[int64]*models.Product{1: product},
		sellerProducts: map[string]*models.Product{"1-1": product},
	}

	svc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	result, err := svc.FindNearest(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bangalore WH (ID=1) is 0 km from seller in Bangalore
	if result.WarehouseID != 1 {
		t.Errorf("expected warehouse 1, got %d", result.WarehouseID)
	}
	if result.DistanceKm != 0 {
		t.Errorf("expected 0 km distance, got %.2f", result.DistanceKm)
	}
	if result.WarehouseName != "Bangalore WH" {
		t.Errorf("expected 'Bangalore WH', got '%s'", result.WarehouseName)
	}
}

func TestFindNearest_SellerNotFound(t *testing.T) {
	whRepo := &mockWarehouseRepo{warehouses: testWarehouses()}
	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{}} // empty
	productRepo := &mockProductRepo{products: map[int64]*models.Product{1: testProduct()}, sellerProducts: map[string]*models.Product{}}

	svc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	_, err := svc.FindNearest(context.Background(), 999, 1)
	if err == nil {
		t.Error("expected error for missing seller")
	}
}

func TestFindNearest_ProductNotBelongToSeller(t *testing.T) {
	whs := testWarehouses()
	whRepo := &mockWarehouseRepo{warehouses: whs}
	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	productRepo := &mockProductRepo{
		products:       map[int64]*models.Product{1: testProduct()},
		sellerProducts: map[string]*models.Product{}, // product 1 doesn't belong to seller 1
	}

	svc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	_, err := svc.FindNearest(context.Background(), 1, 1)
	if err == nil {
		t.Error("expected error when product doesn't belong to seller")
	}
}

func TestFindNearest_NoActiveWarehouses(t *testing.T) {
	whRepo := &mockWarehouseRepo{warehouses: []models.Warehouse{}} // empty
	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	product := testProduct()
	productRepo := &mockProductRepo{
		products:       map[int64]*models.Product{1: product},
		sellerProducts: map[string]*models.Product{"1-1": product},
	}

	svc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	_, err := svc.FindNearest(context.Background(), 1, 1)
	if err == nil {
		t.Error("expected error when no warehouses available")
	}
}

// ── ShippingService Tests ────────────────────────────────────

func TestCalculateCharge_Standard_ShortDistance(t *testing.T) {
	whs := testWarehouses()
	whRepo := &mockWarehouseRepo{
		warehouses: whs,
		byID:       map[int64]*models.Warehouse{1: &whs[0], 2: &whs[1], 3: &whs[2]},
	}
	// Customer in same city as Bangalore WH → very short distance
	customer := &models.Customer{ID: 1, Name: "Local Kirana", Lat: 12.98, Lng: 77.60}
	customerRepo := &mockCustomerRepo{customers: map[int64]*models.Customer{1: customer}}
	product := testProduct()
	productRepo := &mockProductRepo{products: map[int64]*models.Product{1: product}}
	rateRepo := &mockTransportRateRepo{rates: testTransportRates()}
	speedRepo := &mockSpeedConfigRepo{configs: testSpeedConfigs()}

	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	whSvc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	svc := services.NewShippingService(whRepo, customerRepo, productRepo, rateRepo, speedRepo, whSvc)

	result, err := svc.CalculateCharge(context.Background(), 1, 1, 1, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ShippingCharge <= 0 {
		t.Error("expected positive shipping charge")
	}
	if result.Breakdown.TransportMode != "minivan" {
		t.Errorf("expected minivan for short distance, got %s", result.Breakdown.TransportMode)
	}
	if result.Breakdown.ExpressCharge != 0 {
		t.Errorf("expected 0 express charge for standard, got %.2f", result.Breakdown.ExpressCharge)
	}
	if result.Breakdown.BaseCourierCharge != 10 {
		t.Errorf("expected base 10, got %.2f", result.Breakdown.BaseCourierCharge)
	}
}

func TestCalculateCharge_Express_AddsExpressCharge(t *testing.T) {
	whs := testWarehouses()
	whRepo := &mockWarehouseRepo{
		warehouses: whs,
		byID:       map[int64]*models.Warehouse{1: &whs[0]},
	}
	customer := &models.Customer{ID: 1, Name: "Local Kirana", Lat: 12.98, Lng: 77.60}
	customerRepo := &mockCustomerRepo{customers: map[int64]*models.Customer{1: customer}}
	product := testProduct()
	productRepo := &mockProductRepo{products: map[int64]*models.Product{1: product}}
	rateRepo := &mockTransportRateRepo{rates: testTransportRates()}
	speedRepo := &mockSpeedConfigRepo{configs: testSpeedConfigs()}

	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	whSvc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	svc := services.NewShippingService(whRepo, customerRepo, productRepo, rateRepo, speedRepo, whSvc)

	result, err := svc.CalculateCharge(context.Background(), 1, 1, 1, "express")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Express should have positive express charge = 1.2 * billableWeight
	billable := product.BillableWeightKg()
	expectedExpress := 1.2 * billable
	if result.Breakdown.ExpressCharge != expectedExpress {
		t.Errorf("expected express charge %.2f, got %.2f", expectedExpress, result.Breakdown.ExpressCharge)
	}
}

func TestCalculateCharge_InvalidSpeed(t *testing.T) {
	whs := testWarehouses()
	whRepo := &mockWarehouseRepo{
		warehouses: whs,
		byID:       map[int64]*models.Warehouse{1: &whs[0]},
	}
	customerRepo := &mockCustomerRepo{customers: map[int64]*models.Customer{1: testCustomer()}}
	productRepo := &mockProductRepo{products: map[int64]*models.Product{1: testProduct()}}
	rateRepo := &mockTransportRateRepo{rates: testTransportRates()}
	speedRepo := &mockSpeedConfigRepo{configs: testSpeedConfigs()}

	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	whSvc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	svc := services.NewShippingService(whRepo, customerRepo, productRepo, rateRepo, speedRepo, whSvc)

	_, err := svc.CalculateCharge(context.Background(), 1, 1, 1, "overnight")
	if err == nil {
		t.Error("expected error for invalid delivery speed")
	}
}

func TestCalculateCharge_WarehouseNotFound(t *testing.T) {
	whRepo := &mockWarehouseRepo{byID: map[int64]*models.Warehouse{}} // empty
	customerRepo := &mockCustomerRepo{customers: map[int64]*models.Customer{1: testCustomer()}}
	productRepo := &mockProductRepo{products: map[int64]*models.Product{1: testProduct()}}
	rateRepo := &mockTransportRateRepo{rates: testTransportRates()}
	speedRepo := &mockSpeedConfigRepo{configs: testSpeedConfigs()}

	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: testSeller()}}
	whSvc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	svc := services.NewShippingService(whRepo, customerRepo, productRepo, rateRepo, speedRepo, whSvc)

	_, err := svc.CalculateCharge(context.Background(), 999, 1, 1, "standard")
	if err == nil {
		t.Error("expected error for missing warehouse")
	}
}

func TestCalculateFull_EndToEnd(t *testing.T) {
	whs := testWarehouses()
	whRepo := &mockWarehouseRepo{
		warehouses: whs,
		byID:       map[int64]*models.Warehouse{1: &whs[0], 2: &whs[1], 3: &whs[2]},
	}
	seller := testSeller()
	sellerRepo := &mockSellerRepo{sellers: map[int64]*models.Seller{1: seller}}
	product := testProduct()
	productRepo := &mockProductRepo{
		products:       map[int64]*models.Product{1: product},
		sellerProducts: map[string]*models.Product{"1-1": product},
	}
	customer := testCustomer()
	customerRepo := &mockCustomerRepo{customers: map[int64]*models.Customer{1: customer}}
	rateRepo := &mockTransportRateRepo{rates: testTransportRates()}
	speedRepo := &mockSpeedConfigRepo{configs: testSpeedConfigs()}

	whSvc := services.NewWarehouseService(whRepo, sellerRepo, productRepo)
	svc := services.NewShippingService(whRepo, customerRepo, productRepo, rateRepo, speedRepo, whSvc)

	result, err := svc.CalculateFull(context.Background(), 1, 1, 1, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should pick Bangalore WH (closest to seller in Bangalore)
	if result.NearestWarehouse.WarehouseID != 1 {
		t.Errorf("expected nearest warehouse 1, got %d", result.NearestWarehouse.WarehouseID)
	}

	// Charge should be positive
	if result.ShippingCharge <= 0 {
		t.Error("expected positive shipping charge")
	}

	// Should contain valid breakdown
	if result.Breakdown.TransportMode == "" {
		t.Error("expected transport mode in breakdown")
	}
	if result.Breakdown.BillableWeightKg <= 0 {
		t.Error("expected positive billable weight")
	}
}
