package models

import "time"

// Product represents an item listed by a seller with physical attributes for shipping.
type Product struct {
	ID                 int64     `json:"id" db:"id"`
	SellerID           int64     `json:"sellerId" db:"seller_id" validate:"required,gt=0"`
	Name               string    `json:"name" db:"name" validate:"required,min=2,max=255"`
	Description        string    `json:"description" db:"description"`
	SKU                string    `json:"sku" db:"sku" validate:"required,max=100"`
	Category           string    `json:"category" db:"category" validate:"required,oneof=rice pulses masala beverages snacks dairy oil flour sugar other"`
	MRP                float64   `json:"mrp" db:"mrp" validate:"required,gt=0"`
	SellingPrice       float64   `json:"sellingPrice" db:"selling_price" validate:"required,gt=0"`
	BulkPrice          *float64  `json:"bulkPrice,omitempty" db:"bulk_price"`
	ActualWeightKg     float64   `json:"actualWeightKg" db:"actual_weight_kg" validate:"required,gt=0"`
	LengthCm           float64   `json:"lengthCm" db:"length_cm" validate:"required,gt=0"`
	WidthCm            float64   `json:"widthCm" db:"width_cm" validate:"required,gt=0"`
	HeightCm           float64   `json:"heightCm" db:"height_cm" validate:"required,gt=0"`
	VolumetricWeightKg float64   `json:"volumetricWeightKg" db:"volumetric_weight_kg"`
	IsFragile          bool      `json:"isFragile" db:"is_fragile"`
	IsPerishable       bool      `json:"isPerishable" db:"is_perishable"`
	StockQuantity      int       `json:"stockQuantity" db:"stock_quantity"`
	MinOrderQuantity   int       `json:"minOrderQuantity" db:"min_order_quantity"`
	IsActive           bool      `json:"isActive" db:"is_active"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}

// BillableWeightKg returns max(actual weight, volumetric weight) — standard logistics practice.
func (p *Product) BillableWeightKg() float64 {
	vol := (p.LengthCm * p.WidthCm * p.HeightCm) / 5000.0
	if vol > p.ActualWeightKg {
		return vol
	}
	return p.ActualWeightKg
}
