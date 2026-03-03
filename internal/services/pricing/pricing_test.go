package pricing_test

import (
	"math"
	"testing"

	"github.com/jambotails/shipping-service/internal/models"
	"github.com/jambotails/shipping-service/internal/services/pricing"
)

func TestStandardStrategy_Calculate(t *testing.T) {
	s := &pricing.StandardStrategy{}
	b := s.Calculate(50, 10, 3.0, 10.0) // 50km, 10kg, Rs3/km/kg, Rs10 base

	expectedDist := 50 * 10 * 3.0        // 1500
	expectedTotal := 10.0 + expectedDist // 1510

	if b.DistanceCharge != expectedDist {
		t.Errorf("distanceCharge: expected %.2f, got %.2f", expectedDist, b.DistanceCharge)
	}
	if b.ExpressCharge != 0 {
		t.Errorf("expressCharge: expected 0, got %.2f", b.ExpressCharge)
	}
	if b.TotalCharge != expectedTotal {
		t.Errorf("totalCharge: expected %.2f, got %.2f", expectedTotal, b.TotalCharge)
	}
	if b.BaseCourierCharge != 10.0 {
		t.Errorf("baseCourierCharge: expected 10, got %.2f", b.BaseCourierCharge)
	}
}

func TestStandardStrategy_ZeroDistance(t *testing.T) {
	s := &pricing.StandardStrategy{}
	b := s.Calculate(0, 10, 3.0, 10.0)

	if b.DistanceCharge != 0 {
		t.Errorf("expected 0 distance charge, got %.2f", b.DistanceCharge)
	}
	if b.TotalCharge != 10.0 {
		t.Errorf("expected total = base 10, got %.2f", b.TotalCharge)
	}
}

func TestExpressStrategy_Calculate(t *testing.T) {
	s := &pricing.ExpressStrategy{ExtraChargePerKg: 1.2}
	b := s.Calculate(50, 10, 3.0, 10.0) // 50km, 10kg, Rs3/km/kg, Rs10 base, Rs1.2/kg extra

	expectedDist := 50 * 10 * 3.0         // 1500
	expectedExpress := 1.2 * 10.0         // 12
	expectedTotal := 10.0 + 1500.0 + 12.0 // 1522

	if b.DistanceCharge != expectedDist {
		t.Errorf("distanceCharge: expected %.2f, got %.2f", expectedDist, b.DistanceCharge)
	}
	if b.ExpressCharge != expectedExpress {
		t.Errorf("expressCharge: expected %.2f, got %.2f", expectedExpress, b.ExpressCharge)
	}
	if b.TotalCharge != expectedTotal {
		t.Errorf("totalCharge: expected %.2f, got %.2f", expectedTotal, b.TotalCharge)
	}
}

func TestExpressStrategy_ZeroExtraPerKg(t *testing.T) {
	s := &pricing.ExpressStrategy{ExtraChargePerKg: 0}
	b := s.Calculate(100, 5, 2.0, 10.0) // same as standard since extra = 0

	expected := 10.0 + 100*5*2.0 // 1010
	if b.TotalCharge != expected {
		t.Errorf("expected %.2f, got %.2f", expected, b.TotalCharge)
	}
	if b.ExpressCharge != 0 {
		t.Errorf("expected 0 express, got %.2f", b.ExpressCharge)
	}
}

func TestNewStrategy_Standard(t *testing.T) {
	cfg := &models.DeliverySpeedConfig{Speed: "standard", BaseCourierCharge: 10, ExtraChargePerKg: 0}
	s, err := pricing.NewStrategy(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b := s.Calculate(100, 5, 2.0, cfg.BaseCourierCharge)
	if b.ExpressCharge != 0 {
		t.Errorf("standard should have 0 express charge, got %.2f", b.ExpressCharge)
	}
}

func TestNewStrategy_Express(t *testing.T) {
	cfg := &models.DeliverySpeedConfig{Speed: "express", BaseCourierCharge: 10, ExtraChargePerKg: 1.2}
	s, err := pricing.NewStrategy(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b := s.Calculate(100, 5, 2.0, cfg.BaseCourierCharge)
	expectedExpress := 1.2 * 5.0
	if b.ExpressCharge != expectedExpress {
		t.Errorf("expected express charge %.2f, got %.2f", expectedExpress, b.ExpressCharge)
	}
}

func TestNewStrategy_InvalidSpeed(t *testing.T) {
	cfg := &models.DeliverySpeedConfig{Speed: "overnight"}
	_, err := pricing.NewStrategy(cfg)
	if err == nil {
		t.Error("expected error for unsupported speed")
	}
}

func TestPricing_RoundingToTwoDecimals(t *testing.T) {
	s := &pricing.StandardStrategy{}
	// 33.33km * 7.77kg * 2.13/km/kg = ~551.33...
	b := s.Calculate(33.33, 7.77, 2.13, 10.0)
	// Verify result is rounded to 2 decimals
	rounded := math.Round(b.TotalCharge*100) / 100
	if b.TotalCharge != rounded {
		t.Errorf("total not rounded to 2 decimals: %f", b.TotalCharge)
	}
}
