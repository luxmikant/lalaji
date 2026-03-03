package models

import "time"

// Warehouse represents a marketplace-owned distribution center.
type Warehouse struct {
	ID                 int64     `json:"id" db:"id"`
	Name               string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	Code               string    `json:"code" db:"code" validate:"required,max=50"`
	Lat                float64   `json:"lat" db:"lat" validate:"required,latitude"`
	Lng                float64   `json:"lng" db:"lng" validate:"required,longitude"`
	AddressLine1       string    `json:"addressLine1" db:"address_line1"`
	City               string    `json:"city" db:"city"`
	State              string    `json:"state" db:"state"`
	Pincode            string    `json:"pincode" db:"pincode"`
	ContactPerson      string    `json:"contactPerson" db:"contact_person"`
	ContactPhone       string    `json:"contactPhone" db:"contact_phone"`
	MaxCapacitySqft    *int      `json:"maxCapacitySqft,omitempty" db:"max_capacity_sqft"`
	CurrentLoadPercent float64   `json:"currentLoadPercent" db:"current_load_percent"`
	IsActive           bool      `json:"isActive" db:"is_active"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}
