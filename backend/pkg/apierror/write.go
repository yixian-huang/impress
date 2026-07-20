package apierror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Write responds with a structured APIError body:
//
//	{"error":{"code":"...","message":"...","details":{...}}}
//
// Prefer this over ad-hoc gin.H{"error": ...} so clients see a stable shape.
func Write(c *gin.Context, err *APIError) {
	if c == nil || err == nil {
		return
	}
	c.JSON(err.HTTPStatus, err)
}

// AbortWrite is Write followed by c.Abort().
func AbortWrite(c *gin.Context, err *APIError) {
	Write(c, err)
	if c != nil {
		c.Abort()
	}
}

// WriteErr maps common sentinel errors to APIError responses.
// Unknown errors become 500 InternalServerError with the message (or a generic
// message when empty). Returns true if a response was written.
func WriteErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		Write(c, apiErr)
		return true
	}
	Write(c, InternalServerError(err.Error()))
	return true
}

// Message is a convenience for handlers that only have a status + human message
// and want the standard nested error envelope with a generic code.
func Message(c *gin.Context, status int, message string) {
	code := "ERROR"
	switch status {
	case http.StatusBadRequest:
		code = "BAD_REQUEST"
	case http.StatusUnauthorized:
		code = "UNAUTHORIZED"
	case http.StatusForbidden:
		code = "FORBIDDEN"
	case http.StatusNotFound:
		code = "NOT_FOUND"
	case http.StatusConflict:
		code = "CONFLICT"
	case http.StatusInternalServerError:
		code = "INTERNAL_ERROR"
	}
	Write(c, New(status, code, message))
}
