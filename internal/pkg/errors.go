package pkg

import (
	"fmt"
	"net/http"
)

// AppError is the standardized application error used across handlers.
// It carries a machine-readable Code, a human Message, the HTTP status,
// optional internal error detail (never exposed to clients), and optional
// field-level Details for validation failures.
type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Internal   error
	Details    map[string][]string
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap allows errors.Is / errors.As to reach the internal error.
func (e *AppError) Unwrap() error {
	return e.Internal
}

// NewAppError builds an AppError with an optional internal cause.
func NewAppError(code, message string, httpStatus int, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Internal:   internal,
	}
}

// NewValidationError builds an AppError with field-level details.
func NewValidationError(message string, details map[string][]string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
		Details:    details,
	}
}

// BadRequest creates a 400 AppError.
func BadRequest(message string) *AppError {
	return NewAppError("BAD_REQUEST", message, http.StatusBadRequest, nil)
}

// Unauthorized creates a 401 AppError.
func Unauthorized(message string) *AppError {
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized, nil)
}

// Forbidden creates a 403 AppError.
func Forbidden(message string) *AppError {
	return NewAppError("FORBIDDEN", message, http.StatusForbidden, nil)
}

// NotFound creates a 404 AppError.
func NotFound(message string) *AppError {
	return NewAppError("NOT_FOUND", message, http.StatusNotFound, nil)
}

// Conflict creates a 409 AppError.
func Conflict(message string) *AppError {
	return NewAppError("CONFLICT", message, http.StatusConflict, nil)
}

// Internal creates a 500 AppError that never leaks its internal cause.
func Internal(message string, internal error) *AppError {
	return NewAppError("INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError, internal)
}

// RateLimited creates a 429 AppError.
func RateLimited(message string) *AppError {
	return NewAppError("RATE_LIMITED", message, http.StatusTooManyRequests, nil)
}

// ServiceUnavailable creates a 503 AppError (used by throttling).
func ServiceUnavailable(message string) *AppError {
	return NewAppError("SERVICE_UNAVAILABLE", message, http.StatusServiceUnavailable, nil)
}

// Sentinel errors for use with errors.Is. These are plain sentinel vars.
var (
	ErrNotFound     = NotFound("not found")
	ErrUnauthorized = Unauthorized("unauthorized")
	ErrForbidden    = Forbidden("forbidden")
	ErrConflict     = Conflict("conflict")
	ErrInternal     = Internal("internal server error", nil)
	ErrRateLimited  = RateLimited("too many requests")
	ErrUnavailable  = ServiceUnavailable("service unavailable")
)
