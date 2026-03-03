package geo_test

import (
	"math"
	"testing"

	"github.com/jambotails/shipping-service/internal/services/geo"
)

func TestDistance_SamePoint(t *testing.T) {
	d := geo.Distance(12.9716, 77.5946, 12.9716, 77.5946) // Bangalore to itself
	if d != 0 {
		t.Errorf("expected 0, got %f", d)
	}
}

func TestDistance_BangaloreToMumbai(t *testing.T) {
	// Bangalore (12.9716, 77.5946) → Mumbai (19.0760, 72.8777) ≈ 842 km
	d := geo.Distance(12.9716, 77.5946, 19.0760, 72.8777)
	if d < 830 || d > 860 {
		t.Errorf("expected ~842 km, got %.2f", d)
	}
}

func TestDistance_BangaloreToDelhi(t *testing.T) {
	// Bangalore (12.9716, 77.5946) → Delhi (28.7041, 77.1025) ≈ 1745 km
	d := geo.Distance(12.9716, 77.5946, 28.7041, 77.1025)
	if d < 1730 || d > 1760 {
		t.Errorf("expected ~1745 km, got %.2f", d)
	}
}

func TestDistance_ShortDistance(t *testing.T) {
	// Two points ~50 km apart (approx Bangalore to Tumkur)
	d := geo.Distance(12.9716, 77.5946, 13.3379, 77.1025)
	if d < 50 || d > 70 {
		t.Errorf("expected ~60 km, got %.2f", d)
	}
}

func TestDistance_Symmetry(t *testing.T) {
	d1 := geo.Distance(12.9716, 77.5946, 19.0760, 72.8777)
	d2 := geo.Distance(19.0760, 72.8777, 12.9716, 77.5946)
	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("distance not symmetric: %.6f vs %.6f", d1, d2)
	}
}

func TestDistance_Antipodal(t *testing.T) {
	// North Pole to South Pole ≈ 20015 km (half Earth circumference)
	d := geo.Distance(90, 0, -90, 0)
	if d < 20000 || d > 20100 {
		t.Errorf("expected ~20015 km, got %.2f", d)
	}
}

func BenchmarkDistance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		geo.Distance(12.9716, 77.5946, 19.0760, 72.8777)
	}
}
