package errors

import "fmt"

// Error codes used across the application.
const (
	CodeValidation     = "VALIDATION_ERROR"
	CodeNotFound       = "NOT_FOUND"
	CodeInternal       = "INTERNAL_ERROR"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeRateLimited    = "RATE_LIMITED"
	CodeServiceUnavail = "SERVICE_UNAVAILABLE"
)

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
