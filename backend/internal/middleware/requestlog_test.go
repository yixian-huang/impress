package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/pkg/logger"
)

func TestRequestLoggerSetsRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("test", nil)

	r := gin.New()
	r.Use(RequestLogger(log, RequestLoggerOptions{}))
	r.GET("/ping", func(c *gin.Context) {
		id, _ := c.Get(RequestIDContextKey)
		c.String(http.StatusOK, "%v", id)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotEmpty(t, w.Header().Get(RequestIDHeader))
	require.Equal(t, w.Header().Get(RequestIDHeader), w.Body.String())
}

func TestRequestLoggerPropagatesIncomingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New("test", nil)

	r := gin.New()
	r.Use(RequestLogger(log, RequestLoggerOptions{}))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, "abc-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, "abc-123", w.Header().Get(RequestIDHeader))
}
