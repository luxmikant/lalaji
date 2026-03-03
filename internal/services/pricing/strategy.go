package pricing

import "math"

// Breakdown holds the detailed pricing components of a shipping charge calculation.
type Breakdown struct {
	DistanceKm        float64 `json:"distanceKm"`
	TransportMode     string  `json:"transportMode"`
	RatePerKmPerKg    float64 `json:"ratePerKmPerKg"`
	BillableWeightKg  float64 `json:"billableWeightKg"`
	BaseCourierCharge float64 `json:"baseCourierCharge"`
	DistanceCharge    float64 `json:"distanceCharge"`
	ExpressCharge     float64 `json:"expressCharge"`
	TotalCharge       float64 `json:"totalCharge"`
}

// Strategy defines the interface for delivery speed pricing.
type Strategy interface {
	// Calculate computes the full shipping charge based on distance, weight, and rate.
	Calculate(distanceKm, billableWeightKg, ratePerKmPerKg, baseCourierCharge float64) Breakdown
}

// StandardStrategy computes shipping for standard delivery.
// Formula: baseCourier + (rate × distance × weight)
type StandardStrategy struct{}

func (s *StandardStrategy) Calculate(distanceKm, billableWeightKg, ratePerKmPerKg, baseCourierCharge float64) Breakdown {
	distanceCharge := roundTo2(ratePerKmPerKg * distanceKm * billableWeightKg)
	total := roundTo2(baseCourierCharge + distanceCharge)
	return Breakdown{
		BaseCourierCharge: baseCourierCharge,
		DistanceCharge:    distanceCharge,
		ExpressCharge:     0,
		TotalCharge:       total,
	}
}

// ExpressStrategy computes shipping for express delivery.
// Formula: baseCourier + (rate × distance × weight) + (extraPerKg × weight)
type ExpressStrategy struct {
	ExtraChargePerKg float64
}

func (s *ExpressStrategy) Calculate(distanceKm, billableWeightKg, ratePerKmPerKg, baseCourierCharge float64) Breakdown {
	distanceCharge := roundTo2(ratePerKmPerKg * distanceKm * billableWeightKg)
	expressCharge := roundTo2(s.ExtraChargePerKg * billableWeightKg)
	total := roundTo2(baseCourierCharge + distanceCharge + expressCharge)
	return Breakdown{
		BaseCourierCharge: baseCourierCharge,
		DistanceCharge:    distanceCharge,
		ExpressCharge:     expressCharge,
		TotalCharge:       total,
	}
}

// roundTo2 rounds a float to 2 decimal places.
func roundTo2(val float64) float64 {
	return math.Round(val*100) / 100
}
