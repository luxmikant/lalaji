package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// customerRepo is the PostgreSQL implementation of CustomerRepository.
type customerRepo struct {
	db *sql.DB
}

// NewCustomerRepository creates a new customer repository.
func NewCustomerRepository(db *sql.DB) CustomerRepository {
	return &customerRepo{db: db}
}

// GetByID fetches a single customer (Kirana store) by ID.
func (r *customerRepo) GetByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `
		SELECT id, name, owner_name, phone, email, gst_number, store_type,
		       credit_limit, preferred_delivery_slot, lat, lng,
		       address_line1, city, state, pincode, is_active,
		       created_at, updated_at
		FROM customers
		WHERE id = $1`

	c := &models.Customer{}
	var email, gst, slot, addr, city, state, pincode sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.OwnerName, &c.Phone, &email, &gst,
		&c.StoreType, &c.CreditLimit, &slot, &c.Lat, &c.Lng,
		&addr, &city, &state, &pincode,
		&c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("customer with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if email.Valid {
		c.Email = email.String
	}
	if gst.Valid {
		c.GSTNumber = gst.String
	}
	if slot.Valid {
		c.PreferredDeliverySlot = slot.String
	}
	if addr.Valid {
		c.AddressLine1 = addr.String
	}
	if city.Valid {
		c.City = city.String
	}
	if state.Valid {
		c.State = state.String
	}
	if pincode.Valid {
		c.Pincode = pincode.String
	}

	return c, nil
}
