package models

import "time"

// Seller represents a wholesale supplier who lists products on the platform.
type Seller struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	OwnerName    string    `json:"ownerName" db:"owner_name" validate:"required,min=2,max=255"`
	Phone        string    `json:"phone" db:"phone" validate:"required,min=10,max=15"`
	Email        string    `json:"email" db:"email" validate:"omitempty,email"`
	GSTNumber    string    `json:"gstNumber" db:"gst_number" validate:"omitempty,max=20"`
	BusinessType string    `json:"businessType" db:"business_type" validate:"required,oneof=manufacturer distributor wholesaler"`
	Lat          float64   `json:"lat" db:"lat" validate:"required,latitude"`
	Lng          float64   `json:"lng" db:"lng" validate:"required,longitude"`
	AddressLine1 string    `json:"addressLine1" db:"address_line1"`
	City         string    `json:"city" db:"city"`
	State        string    `json:"state" db:"state"`
	Pincode      string    `json:"pincode" db:"pincode"`
	Rating       float64   `json:"rating" db:"rating"`
	IsVerified   bool      `json:"isVerified" db:"is_verified"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}
