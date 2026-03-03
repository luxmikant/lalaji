package transport

// Strategy defines the interface for transport mode selection.
// Each transport mode (MiniVan, Truck, Aeroplane) implements this interface.
type Strategy interface {
	// Name returns the transport mode name (e.g., "minivan", "truck", "aeroplane").
	Name() string

	// RatePerKmPerKg returns the rate charged per kilometre per kilogram.
	RatePerKmPerKg() float64
}

// MiniVanStrategy handles short-distance deliveries (0–100 km).
type MiniVanStrategy struct {
	rate float64
}

func (s *MiniVanStrategy) Name() string            { return "minivan" }
func (s *MiniVanStrategy) RatePerKmPerKg() float64 { return s.rate }

// TruckStrategy handles medium-distance deliveries (100–500 km).
type TruckStrategy struct {
	rate float64
}

func (s *TruckStrategy) Name() string            { return "truck" }
func (s *TruckStrategy) RatePerKmPerKg() float64 { return s.rate }

// AeroplaneStrategy handles long-distance deliveries (500+ km).
type AeroplaneStrategy struct {
	rate float64
}

func (s *AeroplaneStrategy) Name() string            { return "aeroplane" }
func (s *AeroplaneStrategy) RatePerKmPerKg() float64 { return s.rate }
