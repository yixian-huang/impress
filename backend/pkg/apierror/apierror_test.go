package apierror

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	err := New(http.StatusBadRequest, "TEST_ERROR", "Test error message")
	assert.Equal(t, "Test error message", err.Error())
}

func TestNew(t *testing.T) {
	err := New(http.StatusNotFound, "NOT_FOUND", "Resource not found")

	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
	assert.Equal(t, "NOT_FOUND", err.ErrorResponse.Code)
	assert.Equal(t, "Resource not found", err.ErrorResponse.Message)
	assert.Nil(t, err.ErrorResponse.Details)
}

func TestWithDetails(t *testing.T) {
	err := New(http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed").
		WithDetails(map[string]any{
			"field":  "email",
			"reason": "invalid format",
		})

	assert.NotNil(t, err.ErrorResponse.Details)
	assert.Equal(t, "email", err.ErrorResponse.Details["field"])
	assert.Equal(t, "invalid format", err.ErrorResponse.Details["reason"])
}

func TestUnauthorized(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := Unauthorized("Invalid token")
		assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
		assert.Equal(t, "UNAUTHORIZED", err.ErrorResponse.Code)
		assert.Equal(t, "Invalid token", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := Unauthorized("")
		assert.Equal(t, "Authentication required", err.ErrorResponse.Message)
	})
}

func TestForbidden(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := Forbidden("Insufficient permissions")
		assert.Equal(t, http.StatusForbidden, err.HTTPStatus)
		assert.Equal(t, "FORBIDDEN", err.ErrorResponse.Code)
		assert.Equal(t, "Insufficient permissions", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := Forbidden("")
		assert.Equal(t, "Access forbidden", err.ErrorResponse.Message)
	})
}

func TestNotFound(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := NotFound("User not found")
		assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
		assert.Equal(t, "NOT_FOUND", err.ErrorResponse.Code)
		assert.Equal(t, "User not found", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := NotFound("")
		assert.Equal(t, "Resource not found", err.ErrorResponse.Message)
	})
}

func TestValidationFailed(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := ValidationFailed("Email is required")
		assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
		assert.Equal(t, "VALIDATION_FAILED", err.ErrorResponse.Code)
		assert.Equal(t, "Email is required", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := ValidationFailed("")
		assert.Equal(t, "Validation failed", err.ErrorResponse.Message)
	})
}

func TestConflict(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := Conflict("Version mismatch")
		assert.Equal(t, http.StatusConflict, err.HTTPStatus)
		assert.Equal(t, "CONFLICT", err.ErrorResponse.Code)
		assert.Equal(t, "Version mismatch", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := Conflict("")
		assert.Equal(t, "Resource conflict", err.ErrorResponse.Message)
	})
}

func TestInternalServerError(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := InternalServerError("Database connection failed")
		assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
		assert.Equal(t, "INTERNAL_ERROR", err.ErrorResponse.Code)
		assert.Equal(t, "Database connection failed", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := InternalServerError("")
		assert.Equal(t, "Internal server error", err.ErrorResponse.Message)
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := BadRequest("Invalid JSON")
		assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
		assert.Equal(t, "BAD_REQUEST", err.ErrorResponse.Code)
		assert.Equal(t, "Invalid JSON", err.ErrorResponse.Message)
	})

	t.Run("with default message", func(t *testing.T) {
		err := BadRequest("")
		assert.Equal(t, "Bad request", err.ErrorResponse.Message)
	})
}
