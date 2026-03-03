package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// sellerRepo is the PostgreSQL implementation of SellerRepository.
type sellerRepo struct {
	db *sql.DB
}

// NewSellerRepository creates a new seller repository.
func NewSellerRepository(db *sql.DB) SellerRepository {
	return &sellerRepo{db: db}
}

// GetByID fetches a single seller by ID.
func (r *sellerRepo) GetByID(ctx context.Context, id int64) (*models.Seller, error) {
	query := `
		SELECT id, name, owner_name, phone, email, gst_number, business_type,
		       lat, lng, address_line1, city, state, pincode,
		       rating, is_verified, is_active, created_at, updated_at
		FROM sellers
		WHERE id = $1`

	s := &models.Seller{}
	var email, gst, addr sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.OwnerName, &s.Phone, &email, &gst,
		&s.BusinessType, &s.Lat, &s.Lng, &addr,
		&s.City, &s.State, &s.Pincode,
		&s.Rating, &s.IsVerified, &s.IsActive,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("seller with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get seller: %w", err)
	}

	if email.Valid {
		s.Email = email.String
	}
	if gst.Valid {
		s.GSTNumber = gst.String
	}
	if addr.Valid {
		s.AddressLine1 = addr.String
	}

	return s, nil
}
