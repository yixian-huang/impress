package apierror

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a Gin middleware that catches APIError instances
// from c.Errors and converts them to proper JSON responses.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last().Err

			// Check if it's an APIError
			if apiErr, ok := err.(*APIError); ok {
				// Respond with structured error
				c.JSON(apiErr.HTTPStatus, gin.H{
					"error": apiErr.ErrorResponse,
				})
				return
			}

			// For non-APIError, return generic internal server error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": ErrorResponse{
					Code:    "INTERNAL_ERROR",
					Message: "An unexpected error occurred",
				},
			})
		}
	}
}
