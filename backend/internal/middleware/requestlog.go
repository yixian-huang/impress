package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yixian-huang/inkless/backend/pkg/logger"
	"github.com/yixian-huang/inkless/backend/pkg/metrics"
)

const (
	// RequestIDHeader is the HTTP header used to propagate request IDs.
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the gin context key for the request ID.
	RequestIDContextKey = "request_id"
	// DefaultSlowRequestThreshold logs at Warn when a request exceeds this duration.
	DefaultSlowRequestThreshold = 500 * time.Millisecond
)

// RequestLoggerOptions configures request logging middleware.
type RequestLoggerOptions struct {
	// SlowThreshold marks requests at/above this duration as slow (Warn).
	// Zero uses DefaultSlowRequestThreshold.
	SlowThreshold time.Duration
	// QuietPaths are logged at Debug (e.g. health checks).
	QuietPaths map[string]struct{}
}

// RequestLogger logs method/path/status/duration_ms/request_id and records HTTP metrics.
func RequestLogger(log *logger.Logger, opts RequestLoggerOptions) gin.HandlerFunc {
	if opts.SlowThreshold <= 0 {
		opts.SlowThreshold = DefaultSlowRequestThreshold
	}
	if opts.QuietPaths == nil {
		opts.QuietPaths = map[string]struct{}{
			"/health":  {},
			"/metrics": {},
			"/version": {},
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Set(RequestIDContextKey, requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		durationMs := duration.Milliseconds()

		metrics.Global().RecordHTTPRequest(status, duration)

		fields := []interface{}{
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", durationMs,
			"ip", c.ClientIP(),
		}
		if q := c.Request.URL.RawQuery; q != "" && len(q) < 200 {
			fields = append(fields, "query", q)
		}

		if _, quiet := opts.QuietPaths[path]; quiet {
			log.Debug("Request", fields...)
			return
		}
		if duration >= opts.SlowThreshold || status >= 500 {
			log.Warn("Slow or failed request", fields...)
			return
		}
		log.Info("Request", fields...)
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000000")))
	}
	return hex.EncodeToString(b)
}
