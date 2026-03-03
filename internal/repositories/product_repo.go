package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// productRepo is the PostgreSQL implementation of ProductRepository.
type productRepo struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository.
func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepo{db: db}
}

// GetByID fetches a single product by its ID.
func (r *productRepo) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
		SELECT id, seller_id, name, description, sku, category,
		       mrp, selling_price, bulk_price, actual_weight_kg,
		       length_cm, width_cm, height_cm, volumetric_weight_kg,
		       is_fragile, is_perishable, stock_quantity, min_order_quantity,
		       is_active, created_at, updated_at
		FROM products
		WHERE id = $1`

	p := &models.Product{}
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.SellerID, &p.Name, &description, &p.SKU, &p.Category,
		&p.MRP, &p.SellingPrice, &p.BulkPrice, &p.ActualWeightKg,
		&p.LengthCm, &p.WidthCm, &p.HeightCm, &p.VolumetricWeightKg,
		&p.IsFragile, &p.IsPerishable, &p.StockQuantity, &p.MinOrderQuantity,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if description.Valid {
		p.Description = description.String
	}

	return p, nil
}

// GetByIDAndSellerID fetches a product validating seller ownership.
func (r *productRepo) GetByIDAndSellerID(ctx context.Context, productID, sellerID int64) (*models.Product, error) {
	query := `
		SELECT id, seller_id, name, description, sku, category,
		       mrp, selling_price, bulk_price, actual_weight_kg,
		       length_cm, width_cm, height_cm, volumetric_weight_kg,
		       is_fragile, is_perishable, stock_quantity, min_order_quantity,
		       is_active, created_at, updated_at
		FROM products
		WHERE id = $1 AND seller_id = $2`

	p := &models.Product{}
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, productID, sellerID).Scan(
		&p.ID, &p.SellerID, &p.Name, &description, &p.SKU, &p.Category,
		&p.MRP, &p.SellingPrice, &p.BulkPrice, &p.ActualWeightKg,
		&p.LengthCm, &p.WidthCm, &p.HeightCm, &p.VolumetricWeightKg,
		&p.IsFragile, &p.IsPerishable, &p.StockQuantity, &p.MinOrderQuantity,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product %d not found for seller %d", productID, sellerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if description.Valid {
		p.Description = description.String
	}

	return p, nil
}
