package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jambotails/shipping-service/internal/models"
)

// rateConfigRepo is the PostgreSQL implementation of transport rate and delivery speed config repos.
type rateConfigRepo struct {
	db *sql.DB
}

// NewTransportRateRepository creates a new transport rate config repository.
func NewTransportRateRepository(db *sql.DB) TransportRateRepository {
	return &rateConfigRepo{db: db}
}

// NewDeliverySpeedConfigRepository creates a new delivery speed config repository.
func NewDeliverySpeedConfigRepository(db *sql.DB) DeliverySpeedConfigRepository {
	return &rateConfigRepo{db: db}
}

// GetAllActive returns all active transport rate configs.
func (r *rateConfigRepo) GetAllActive(ctx context.Context) ([]models.TransportRate, error) {
	query := `
		SELECT id, mode, min_distance_km, max_distance_km, rate_per_km_per_kg,
		       effective_from, effective_to, is_active, created_at
		FROM transport_rates
		WHERE is_active = TRUE
		  AND effective_from <= CURRENT_DATE
		  AND (effective_to IS NULL OR effective_to >= CURRENT_DATE)
		ORDER BY min_distance_km ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query transport rates: %w", err)
	}
	defer rows.Close()

	var rates []models.TransportRate
	for rows.Next() {
		var tr models.TransportRate
		if err := rows.Scan(
			&tr.ID, &tr.Mode, &tr.MinDistanceKm, &tr.MaxDistanceKm,
			&tr.RatePerKmPerKg, &tr.EffectiveFrom, &tr.EffectiveTo,
			&tr.IsActive, &tr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transport rate: %w", err)
		}
		rates = append(rates, tr)
	}

	return rates, rows.Err()
}

// GetBySpeed returns the active delivery speed config for the given speed.
func (r *rateConfigRepo) GetBySpeed(ctx context.Context, speed string) (*models.DeliverySpeedConfig, error) {
	query := `
		SELECT id, speed, base_courier_charge, extra_charge_per_kg, is_active, created_at
		FROM delivery_speed_configs
		WHERE speed = $1 AND is_active = TRUE`

	cfg := &models.DeliverySpeedConfig{}
	err := r.db.QueryRowContext(ctx, query, speed).Scan(
		&cfg.ID, &cfg.Speed, &cfg.BaseCourierCharge,
		&cfg.ExtraChargePerKg, &cfg.IsActive, &cfg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("delivery speed config '%s' not found", speed)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get speed config: %w", err)
	}

	return cfg, nil
}
