package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// warehouseRepo is the PostgreSQL implementation of WarehouseRepository.
type warehouseRepo struct {
	db *sql.DB
}

// NewWarehouseRepository creates a new warehouse repository.
func NewWarehouseRepository(db *sql.DB) WarehouseRepository {
	return &warehouseRepo{db: db}
}

// GetByID fetches a single warehouse by its ID.
func (r *warehouseRepo) GetByID(ctx context.Context, id int64) (*models.Warehouse, error) {
	query := `
		SELECT id, name, code, lat, lng, address_line1, city, state, pincode,
		       contact_person, contact_phone, max_capacity_sqft, current_load_percent,
		       is_active, created_at, updated_at
		FROM warehouses
		WHERE id = $1`

	w := &models.Warehouse{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.Name, &w.Code, &w.Lat, &w.Lng,
		&w.AddressLine1, &w.City, &w.State, &w.Pincode,
		&w.ContactPerson, &w.ContactPhone, &w.MaxCapacitySqft,
		&w.CurrentLoadPercent, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("warehouse with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouse: %w", err)
	}
	return w, nil
}

// GetAllActive fetches all warehouses where is_active = true.
func (r *warehouseRepo) GetAllActive(ctx context.Context) ([]models.Warehouse, error) {
	query := `
		SELECT id, name, code, lat, lng, address_line1, city, state, pincode,
		       contact_person, contact_phone, max_capacity_sqft, current_load_percent,
		       is_active, created_at, updated_at
		FROM warehouses
		WHERE is_active = TRUE
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []models.Warehouse
	for rows.Next() {
		var w models.Warehouse
		if err := rows.Scan(
			&w.ID, &w.Name, &w.Code, &w.Lat, &w.Lng,
			&w.AddressLine1, &w.City, &w.State, &w.Pincode,
			&w.ContactPerson, &w.ContactPhone, &w.MaxCapacitySqft,
			&w.CurrentLoadPercent, &w.IsActive, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan warehouse row: %w", err)
		}
		warehouses = append(warehouses, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating warehouse rows: %w", err)
	}

	return warehouses, nil
}
