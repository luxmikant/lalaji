package transport

import (
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// NewStrategy selects the appropriate transport strategy based on distance
// and the transport rate configuration from the DB.
//
// Rates are expected to have non-overlapping distance ranges:
//   - minivan:    [0, 100)
//   - truck:      [100, 500)
//   - aeroplane:  [500, ∞)
func NewStrategy(distanceKm float64, rates []models.TransportRate) (Strategy, error) {
	for _, r := range rates {
		minOk := distanceKm >= r.MinDistanceKm
		maxOk := r.MaxDistanceKm == nil || distanceKm < *r.MaxDistanceKm
		if minOk && maxOk {
			return strategyFor(r.Mode, r.RatePerKmPerKg)
		}
	}
	return nil, fmt.Errorf("no transport rate found for distance %.2f km", distanceKm)
}

// strategyFor creates a Strategy instance based on the mode string.
func strategyFor(mode string, rate float64) (Strategy, error) {
	switch mode {
	case "minivan":
		return &MiniVanStrategy{rate: rate}, nil
	case "truck":
		return &TruckStrategy{rate: rate}, nil
	case "aeroplane":
		return &AeroplaneStrategy{rate: rate}, nil
	default:
		return nil, fmt.Errorf("unknown transport mode: %s", mode)
	}
}
