package models

import "time"

// TransportRate stores transport mode pricing rules — configurable via DB.
type TransportRate struct {
	ID             int64     `json:"id" db:"id"`
	Mode           string    `json:"mode" db:"mode" validate:"required,oneof=aeroplane truck minivan"`
	MinDistanceKm  float64   `json:"minDistanceKm" db:"min_distance_km" validate:"gte=0"`
	MaxDistanceKm  *float64  `json:"maxDistanceKm,omitempty" db:"max_distance_km"` // nil = infinity
	RatePerKmPerKg float64   `json:"ratePerKmPerKg" db:"rate_per_km_per_kg" validate:"required,gt=0"`
	EffectiveFrom  string    `json:"effectiveFrom" db:"effective_from"`
	EffectiveTo    *string   `json:"effectiveTo,omitempty" db:"effective_to"`
	IsActive       bool      `json:"isActive" db:"is_active"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
}

// DeliverySpeedConfig stores delivery speed surcharge rules — configurable via DB.
type DeliverySpeedConfig struct {
	ID                int64     `json:"id" db:"id"`
	Speed             string    `json:"speed" db:"speed" validate:"required,oneof=standard express"`
	BaseCourierCharge float64   `json:"baseCourierCharge" db:"base_courier_charge"`
	ExtraChargePerKg  float64   `json:"extraChargePerKg" db:"extra_charge_per_kg"`
	IsActive          bool      `json:"isActive" db:"is_active"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
}
