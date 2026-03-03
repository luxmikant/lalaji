package transport_test

import (
	"testing"

	"github.com/jambotails/shipping-service/internal/models"
	"github.com/jambotails/shipping-service/internal/services/transport"
)

// testRates returns a standard set of transport rates for testing.
func testRates() []models.TransportRate {
	maxDist100 := 100.0
	maxDist500 := 500.0
	return []models.TransportRate{
		{Mode: "minivan", MinDistanceKm: 0, MaxDistanceKm: &maxDist100, RatePerKmPerKg: 3.0},
		{Mode: "truck", MinDistanceKm: 100, MaxDistanceKm: &maxDist500, RatePerKmPerKg: 2.0},
		{Mode: "aeroplane", MinDistanceKm: 500, MaxDistanceKm: nil, RatePerKmPerKg: 1.0},
	}
}

func TestNewStrategy_MiniVan(t *testing.T) {
	s, err := transport.NewStrategy(50, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "minivan" {
		t.Errorf("expected minivan, got %s", s.Name())
	}
	if s.RatePerKmPerKg() != 3.0 {
		t.Errorf("expected rate 3.0, got %f", s.RatePerKmPerKg())
	}
}

func TestNewStrategy_MiniVan_Zero(t *testing.T) {
	s, err := transport.NewStrategy(0, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "minivan" {
		t.Errorf("expected minivan for 0 km, got %s", s.Name())
	}
}

func TestNewStrategy_MiniVan_BoundaryBelow100(t *testing.T) {
	s, err := transport.NewStrategy(99.99, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "minivan" {
		t.Errorf("expected minivan for 99.99 km, got %s", s.Name())
	}
}

func TestNewStrategy_Truck(t *testing.T) {
	s, err := transport.NewStrategy(250, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "truck" {
		t.Errorf("expected truck, got %s", s.Name())
	}
	if s.RatePerKmPerKg() != 2.0 {
		t.Errorf("expected rate 2.0, got %f", s.RatePerKmPerKg())
	}
}

func TestNewStrategy_Truck_LowerBoundary(t *testing.T) {
	s, err := transport.NewStrategy(100, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "truck" {
		t.Errorf("expected truck for exactly 100 km, got %s", s.Name())
	}
}

func TestNewStrategy_Truck_UpperBoundary(t *testing.T) {
	s, err := transport.NewStrategy(499.99, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "truck" {
		t.Errorf("expected truck for 499.99 km, got %s", s.Name())
	}
}

func TestNewStrategy_Aeroplane(t *testing.T) {
	s, err := transport.NewStrategy(600, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "aeroplane" {
		t.Errorf("expected aeroplane, got %s", s.Name())
	}
	if s.RatePerKmPerKg() != 1.0 {
		t.Errorf("expected rate 1.0, got %f", s.RatePerKmPerKg())
	}
}

func TestNewStrategy_Aeroplane_Boundary(t *testing.T) {
	s, err := transport.NewStrategy(500, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "aeroplane" {
		t.Errorf("expected aeroplane for exactly 500 km, got %s", s.Name())
	}
}

func TestNewStrategy_Aeroplane_LargeDistance(t *testing.T) {
	s, err := transport.NewStrategy(5000, testRates())
	if err != nil {
		t.Fatal(err)
	}
	if s.Name() != "aeroplane" {
		t.Errorf("expected aeroplane for 5000 km, got %s", s.Name())
	}
}

func TestNewStrategy_EmptyRates(t *testing.T) {
	_, err := transport.NewStrategy(50, []models.TransportRate{})
	if err == nil {
		t.Error("expected error for empty rates")
	}
}

func TestNewStrategy_NegativeDistance(t *testing.T) {
	// Edge case: negative distance should not match any range
	_, err := transport.NewStrategy(-10, testRates())
	if err == nil {
		t.Error("expected error for negative distance")
	}
}
