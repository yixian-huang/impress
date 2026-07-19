package apierror

import (
	"net/http"
)

// APIError represents a standardized API error response.
// It implements the error interface and provides structured error information.
type APIError struct {
	// HTTPStatus is the HTTP status code to return
	HTTPStatus int `json:"-"`
	// ErrorResponse is the JSON error payload
	ErrorResponse ErrorResponse `json:"error"`
}

// ErrorResponse represents the JSON structure of error responses
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return e.ErrorResponse.Message
}

// New creates a new APIError with the given HTTP status, code, and message
func New(status int, code string, message string) *APIError {
	return &APIError{
		HTTPStatus: status,
		ErrorResponse: ErrorResponse{
			Code:    code,
			Message: message,
		},
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details map[string]any) *APIError {
	e.ErrorResponse.Details = details
	return e
}

// Factory functions for common errors

// Unauthorized creates a 401 Unauthorized error
func Unauthorized(message string) *APIError {
	if message == "" {
		message = "Authentication required"
	}
	return New(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden creates a 403 Forbidden error
func Forbidden(message string) *APIError {
	if message == "" {
		message = "Access forbidden"
	}
	return New(http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound creates a 404 Not Found error
func NotFound(message string) *APIError {
	if message == "" {
		message = "Resource not found"
	}
	return New(http.StatusNotFound, "NOT_FOUND", message)
}

// ValidationFailed creates a 400 Bad Request error for validation failures
func ValidationFailed(message string) *APIError {
	if message == "" {
		message = "Validation failed"
	}
	return New(http.StatusBadRequest, "VALIDATION_FAILED", message)
}

// Conflict creates a 409 Conflict error
func Conflict(message string) *APIError {
	if message == "" {
		message = "Resource conflict"
	}
	return New(http.StatusConflict, "CONFLICT", message)
}

// InternalServerError creates a 500 Internal Server Error
func InternalServerError(message string) *APIError {
	if message == "" {
		message = "Internal server error"
	}
	return New(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// BadRequest creates a 400 Bad Request error
func BadRequest(message string) *APIError {
	if message == "" {
		message = "Bad request"
	}
	return New(http.StatusBadRequest, "BAD_REQUEST", message)
}
