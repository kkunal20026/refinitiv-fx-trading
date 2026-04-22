package errors

import (
	"fmt"
	"net/http"
)

// ErrorType defines the type of error
type ErrorType string

const (
	ValidationError ErrorType = "VALIDATION_ERROR"
	NotFoundError   ErrorType = "NOT_FOUND_ERROR"
	ConflictError   ErrorType = "CONFLICT_ERROR"
	InternalError   ErrorType = "INTERNAL_ERROR"
	UnauthorizedError ErrorType = "UNAUTHORIZED_ERROR"
	ForbiddenError  ErrorType = "FORBIDDEN_ERROR"
	RateLimitError  ErrorType = "RATE_LIMIT_ERROR"
	ServiceError    ErrorType = "SERVICE_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Type       ErrorType   `json:"type"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Details    interface{} `json:"details,omitempty"`
	Err        error       `json:"-"`
}

// Error implements the error interface
func (ae *AppError) Error() string {
	if ae.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", ae.Type, ae.Message, ae.Err)
	}
	return fmt.Sprintf("%s: %s", ae.Type, ae.Message)
}

// Unwrap returns the underlying error
func (ae *AppError) Unwrap() error {
	return ae.Err
}

// NewAppError creates a new application error
func NewAppError(errType ErrorType, message string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       errType,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}, err error) *AppError {
	return &AppError{
		Type:       ValidationError,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    details,
		Err:        err,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		Type:       NotFoundError,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string, err error) *AppError {
	return &AppError{
		Type:       ConflictError,
		Message:    message,
		StatusCode: http.StatusConflict,
		Err:        err,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		Type:       UnauthorizedError,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string, err error) *AppError {
	return &AppError{
		Type:       ForbiddenError,
		Message:    message,
		StatusCode: http.StatusForbidden,
		Err:        err,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string, err error) *AppError {
	return &AppError{
		Type:       RateLimitError,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
		Err:        err,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       InternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewServiceError creates a service error (e.g., from Refinitiv API)
func NewServiceError(message string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       ServiceError,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// IsAppError checks if error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	appErr, ok := err.(*AppError)
	return appErr, ok
}
