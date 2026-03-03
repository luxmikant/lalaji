package models

import "time"

// Shipment represents the logistics leg from warehouse to customer.
type Shipment struct {
	ID                    int64      `json:"id" db:"id"`
	OrderID               int64      `json:"orderId" db:"order_id"`
	SourceWarehouseID     int64      `json:"sourceWarehouseId" db:"source_warehouse_id"`
	DestinationCustomerID int64      `json:"destinationCustomerId" db:"destination_customer_id"`
	DistanceKm            float64    `json:"distanceKm" db:"distance_km"`
	TransportMode         string     `json:"transportMode" db:"transport_mode"`
	BillableWeightKg      float64    `json:"billableWeightKg" db:"billable_weight_kg"`
	RatePerKmPerKg        float64    `json:"ratePerKmPerKg" db:"rate_per_km_per_kg"`
	BaseCourierCharge     float64    `json:"baseCourierCharge" db:"base_courier_charge"`
	DistanceCharge        float64    `json:"distanceCharge" db:"distance_charge"`
	ExpressCharge         float64    `json:"expressCharge" db:"express_charge"`
	TotalShippingCharge   float64    `json:"totalShippingCharge" db:"total_shipping_charge"`
	Status                string     `json:"status" db:"status"`
	DispatchedAt          *time.Time `json:"dispatchedAt,omitempty" db:"dispatched_at"`
	DeliveredAt           *time.Time `json:"deliveredAt,omitempty" db:"delivered_at"`
	CreatedAt             time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time  `json:"updatedAt" db:"updated_at"`
}
