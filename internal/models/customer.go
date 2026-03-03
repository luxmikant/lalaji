package models

import "time"

// Customer represents a Kirana store owner registered on the platform.
type Customer struct {
	ID                    int64     `json:"id" db:"id"`
	Name                  string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	OwnerName             string    `json:"ownerName" db:"owner_name" validate:"required,min=2,max=255"`
	Phone                 string    `json:"phone" db:"phone" validate:"required,min=10,max=15"`
	Email                 string    `json:"email" db:"email" validate:"omitempty,email"`
	GSTNumber             string    `json:"gstNumber" db:"gst_number" validate:"omitempty,max=20"`
	StoreType             string    `json:"storeType" db:"store_type" validate:"required,oneof=grocery dairy general medical bakery"`
	CreditLimit           float64   `json:"creditLimit" db:"credit_limit"`
	PreferredDeliverySlot string    `json:"preferredDeliverySlot" db:"preferred_delivery_slot"`
	Lat                   float64   `json:"lat" db:"lat" validate:"required,latitude"`
	Lng                   float64   `json:"lng" db:"lng" validate:"required,longitude"`
	AddressLine1          string    `json:"addressLine1" db:"address_line1"`
	City                  string    `json:"city" db:"city"`
	State                 string    `json:"state" db:"state"`
	Pincode               string    `json:"pincode" db:"pincode"`
	IsActive              bool      `json:"isActive" db:"is_active"`
	CreatedAt             time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" db:"updated_at"`
}
