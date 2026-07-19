package apierror

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandler_APIError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test handler that returns an APIError
	handler := func(c *gin.Context) {
		err := NotFound("User not found")
		_ = c.Error(err)
	}

	// Apply error handler middleware
	errorMiddleware := ErrorHandler()

	// Execute handlers
	handler(c)
	errorMiddleware(c)

	// Verify response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", response["error"].Code)
	assert.Equal(t, "User not found", response["error"].Message)
}

func TestErrorHandler_APIErrorWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test handler that returns an APIError with details
	handler := func(c *gin.Context) {
		err := ValidationFailed("Validation failed").WithDetails(map[string]any{
			"field":  "email",
			"reason": "invalid format",
		})
		_ = c.Error(err)
	}

	errorMiddleware := ErrorHandler()

	handler(c)
	errorMiddleware(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "VALIDATION_FAILED", response["error"].Code)
	assert.NotNil(t, response["error"].Details)
}

func TestErrorHandler_GenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test handler that returns a generic error
	handler := func(c *gin.Context) {
		err := errors.New("unexpected error")
		_ = c.Error(err)
	}

	errorMiddleware := ErrorHandler()

	handler(c)
	errorMiddleware(c)

	// Should return generic internal server error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "INTERNAL_ERROR", response["error"].Code)
	assert.Equal(t, "An unexpected error occurred", response["error"].Message)
}

func TestErrorHandler_NoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test handler that succeeds
	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	errorMiddleware := ErrorHandler()

	handler(c)
	errorMiddleware(c)

	// Should return the success response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["message"])
}

func TestErrorHandler_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a test handler that adds multiple errors
	handler := func(c *gin.Context) {
		_ = c.Error(errors.New("first error"))
		_ = c.Error(NotFound("Resource not found"))
	}

	errorMiddleware := ErrorHandler()

	handler(c)
	errorMiddleware(c)

	// Should return the last error (NotFound)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", response["error"].Code)
	assert.Equal(t, "Resource not found", response["error"].Message)
}
