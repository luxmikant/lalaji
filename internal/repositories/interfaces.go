package repositories

import (
	"context"

	"github.com/jambotails/shipping-service/internal/models"
)

// WarehouseRepository defines data access for warehouses.
type WarehouseRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Warehouse, error)
	GetAllActive(ctx context.Context) ([]models.Warehouse, error)
}

// CustomerRepository defines data access for customers (Kirana stores).
type CustomerRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Customer, error)
}

// SellerRepository defines data access for sellers.
type SellerRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Seller, error)
}

// ProductRepository defines data access for products.
type ProductRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	GetByIDAndSellerID(ctx context.Context, productID, sellerID int64) (*models.Product, error)
}

// TransportRateRepository defines data access for transport rate configs.
type TransportRateRepository interface {
	GetAllActive(ctx context.Context) ([]models.TransportRate, error)
}

// DeliverySpeedConfigRepository defines data access for delivery speed configs.
type DeliverySpeedConfigRepository interface {
	GetBySpeed(ctx context.Context, speed string) (*models.DeliverySpeedConfig, error)
}

// OrderRepository defines data access for orders.
type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
}

// ShipmentRepository defines data access for shipments.
type ShipmentRepository interface {
	Create(ctx context.Context, shipment *models.Shipment) (*models.Shipment, error)
}
