package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// AppValidator wraps go-playground/validator with custom rules.
var AppValidator *validator.Validate

func init() {
	AppValidator = validator.New()
}

// ValidateDeliverySpeed checks if the speed is a valid enum value.
func ValidateDeliverySpeed(speed string) error {
	s := strings.ToLower(strings.TrimSpace(speed))
	if s != "standard" && s != "express" {
		return fmt.Errorf("deliverySpeed must be 'standard' or 'express'")
	}
	return nil
}

// FormatValidationErrors converts validator.ValidationErrors into a user-friendly map.
func FormatValidationErrors(err error) map[string]string {
	fields := make(map[string]string)
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fields[fe.Field()] = formatFieldError(fe)
		}
	}
	return fields
}

// formatFieldError generates a human-readable message for a single field validation error.
func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email", fe.Field())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude (-90 to 90)", fe.Field())
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude (-180 to 180)", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
