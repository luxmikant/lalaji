package errors

import (
	"errors"
	"fmt"
)

// Error codes used across the application.
const (
	CodeValidation          = "VALIDATION_ERROR"
	CodeNotFound            = "NOT_FOUND"
	CodeInternal            = "INTERNAL_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeRateLimited         = "RATE_LIMITED"
	CodeServiceUnavail      = "SERVICE_UNAVAILABLE"
	CodeWarehouseNotFound   = "WAREHOUSE_NOT_FOUND"
	CodeNoWarehouseActive   = "NO_ACTIVE_WAREHOUSE"
	CodeSellerNotFound      = "SELLER_NOT_FOUND"
	CodeProductNotFound     = "PRODUCT_NOT_FOUND"
	CodeCustomerNotFound    = "CUSTOMER_NOT_FOUND"
	CodeDeliveryUnsupported = "DELIVERY_UNSUPPORTED"
	CodeInvalidSpeed        = "INVALID_DELIVERY_SPEED"
	CodeTransportConfig     = "TRANSPORT_CONFIG_ERROR"
)

// AsAppError attempts to unwrap err into an *AppError.
// Returns (appErr, true) if successful, (nil, false) otherwise.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// AppError is a structured error used throughout the application.
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	HTTPStatus int               `json:"-"`
	Fields     map[string]string `json:"fields,omitempty"`
}

// Error satisfies the error interface.
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewValidationError returns a 400 validation error.
func NewValidationError(message string, fields map[string]string) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		HTTPStatus: 400,
		Fields:     fields,
	}
}

// NewNotFoundError returns a 404 not found error.
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    message,
		HTTPStatus: 404,
	}
}

// NewInternalError returns a 500 internal server error.
func NewInternalError(message string) *AppError {
	return &AppError{
		Code:       CodeInternal,
		Message:    message,
		HTTPStatus: 500,
	}
}

// NewUnauthorizedError returns a 401 unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: 401,
	}
}

// NewRateLimitedError returns a 429 error.
func NewRateLimitedError() *AppError {
	return &AppError{
		Code:       CodeRateLimited,
		Message:    "too many requests",
		HTTPStatus: 429,
	}
}

// NewServiceUnavailableError returns a 503 error.
func NewServiceUnavailableError(message string) *AppError {
	return &AppError{
		Code:       CodeServiceUnavail,
		Message:    message,
		HTTPStatus: 503,
	}
}

// ─── Domain-specific constructors ────────────────────────────────────────────

// NewWarehouseNotFoundError returns a 404 when a specific warehouse ID is not found.
func NewWarehouseNotFoundError(id int64) *AppError {
	return &AppError{
		Code:       CodeWarehouseNotFound,
		Message:    fmt.Sprintf("warehouse with id %d not found", id),
		HTTPStatus: 404,
	}
}

// NewNoActiveWarehouseError returns a 503 when there are zero active warehouses.
// This is a platform-level problem, not a client error.
func NewNoActiveWarehouseError() *AppError {
	return &AppError{
		Code:       CodeNoWarehouseActive,
		Message:    "no active warehouses are currently available; please try again later",
		HTTPStatus: 503,
	}
}

// NewSellerNotFoundError returns a 404 when a seller is not found or inactive.
func NewSellerNotFoundError(id int64) *AppError {
	return &AppError{
		Code:       CodeSellerNotFound,
		Message:    fmt.Sprintf("seller with id %d not found or is not active", id),
		HTTPStatus: 404,
	}
}

// NewProductNotFoundError returns a 404 when a product is not found or does not belong to the seller.
func NewProductNotFoundError(productID, sellerID int64) *AppError {
	return &AppError{
		Code:       CodeProductNotFound,
		Message:    fmt.Sprintf("product with id %d not found for seller %d", productID, sellerID),
		HTTPStatus: 404,
	}
}

// NewCustomerNotFoundError returns a 404 when a customer is not found.
func NewCustomerNotFoundError(id int64) *AppError {
	return &AppError{
		Code:       CodeCustomerNotFound,
		Message:    fmt.Sprintf("customer with id %d not found", id),
		HTTPStatus: 404,
	}
}

// NewDeliveryUnsupportedError returns a 422 when no transport mode covers the distance.
// This means the delivery location is outside the range of all configured transport modes.
func NewDeliveryUnsupportedError(distanceKm float64) *AppError {
	return &AppError{
		Code:       CodeDeliveryUnsupported,
		Message:    fmt.Sprintf("no transport mode is configured for a distance of %.2f km; delivery to this location is not supported", distanceKm),
		HTTPStatus: 422,
	}
}

// NewInvalidDeliverySpeedError returns a 400 for unrecognised delivery speed values.
func NewInvalidDeliverySpeedError(got string) *AppError {
	return &AppError{
		Code:       CodeInvalidSpeed,
		Message:    fmt.Sprintf("invalid deliverySpeed '%s': must be 'standard' or 'express'", got),
		HTTPStatus: 400,
	}
}

// NewTransportConfigError returns a 500 when transport/pricing strategy cannot be initialised.
func NewTransportConfigError(detail string) *AppError {
	return &AppError{
		Code:       CodeTransportConfig,
		Message:    "failed to initialise transport configuration: " + detail,
		HTTPStatus: 500,
	}
}
