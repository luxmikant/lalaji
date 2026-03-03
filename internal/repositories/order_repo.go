package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// orderRepo is the PostgreSQL implementation of OrderRepository.
type orderRepo struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository.
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepo{db: db}
}

// Create inserts a new order and returns it with the generated ID.
func (r *orderRepo) Create(ctx context.Context, order *models.Order) (*models.Order, error) {
	query := `
		INSERT INTO orders (customer_id, seller_id, nearest_warehouse_id, product_id,
		                    quantity, unit_price, total_product_amount, shipping_charge,
		                    total_amount, delivery_speed, status, payment_mode,
		                    tracking_id, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		order.CustomerID, order.SellerID, order.NearestWarehouseID, order.ProductID,
		order.Quantity, order.UnitPrice, order.TotalProductAmount, order.ShippingCharge,
		order.TotalAmount, order.DeliverySpeed, order.Status, order.PaymentMode,
		order.TrackingID, order.Notes,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	return order, nil
}

// shipmentRepo is the PostgreSQL implementation of ShipmentRepository.
type shipmentRepo struct {
	db *sql.DB
}

// NewShipmentRepository creates a new shipment repository.
func NewShipmentRepository(db *sql.DB) ShipmentRepository {
	return &shipmentRepo{db: db}
}

// Create inserts a new shipment and returns it with the generated ID.
func (r *shipmentRepo) Create(ctx context.Context, shipment *models.Shipment) (*models.Shipment, error) {
	query := `
		INSERT INTO shipments (order_id, source_warehouse_id, destination_customer_id,
		                       distance_km, transport_mode, billable_weight_kg,
		                       rate_per_km_per_kg, base_courier_charge, distance_charge,
		                       express_charge, total_shipping_charge, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		shipment.OrderID, shipment.SourceWarehouseID, shipment.DestinationCustomerID,
		shipment.DistanceKm, shipment.TransportMode, shipment.BillableWeightKg,
		shipment.RatePerKmPerKg, shipment.BaseCourierCharge, shipment.DistanceCharge,
		shipment.ExpressCharge, shipment.TotalShippingCharge, shipment.Status,
	).Scan(&shipment.ID, &shipment.CreatedAt, &shipment.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}
	return shipment, nil
}
