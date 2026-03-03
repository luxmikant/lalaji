package models

import "time"

// Order ties together a customer, seller, product, and shipping detail.
type Order struct {
	ID                    int64     `json:"id" db:"id"`
	CustomerID            int64     `json:"customerId" db:"customer_id"`
	SellerID              int64     `json:"sellerId" db:"seller_id"`
	NearestWarehouseID    int64     `json:"nearestWarehouseId" db:"nearest_warehouse_id"`
	ProductID             int64     `json:"productId" db:"product_id"`
	Quantity              int       `json:"quantity" db:"quantity"`
	UnitPrice             float64   `json:"unitPrice" db:"unit_price"`
	TotalProductAmount    float64   `json:"totalProductAmount" db:"total_product_amount"`
	ShippingCharge        float64   `json:"shippingCharge" db:"shipping_charge"`
	TotalAmount           float64   `json:"totalAmount" db:"total_amount"`
	DeliverySpeed         string    `json:"deliverySpeed" db:"delivery_speed"`
	Status                string    `json:"status" db:"status"`
	PaymentMode           string    `json:"paymentMode" db:"payment_mode"`
	TrackingID            string    `json:"trackingId" db:"tracking_id"`
	EstimatedDeliveryDate *string   `json:"estimatedDeliveryDate,omitempty" db:"estimated_delivery_date"`
	ActualDeliveryDate    *string   `json:"actualDeliveryDate,omitempty" db:"actual_delivery_date"`
	Notes                 string    `json:"notes" db:"notes"`
	CreatedAt             time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt             time.Time `json:"updatedAt" db:"updated_at"`
}
