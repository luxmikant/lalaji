package pricing

import (
	"fmt"
	"strings"

	"github.com/jambotails/shipping-service/internal/models"
)

// NewStrategy creates the appropriate pricing strategy based on the delivery speed config.
func NewStrategy(speedConfig *models.DeliverySpeedConfig) (Strategy, error) {
	switch strings.ToLower(speedConfig.Speed) {
	case "standard":
		return &StandardStrategy{}, nil
	case "express":
		return &ExpressStrategy{ExtraChargePerKg: speedConfig.ExtraChargePerKg}, nil
	default:
		return nil, fmt.Errorf("unsupported delivery speed: %s", speedConfig.Speed)
	}
}
