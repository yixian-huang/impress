package metrics

import (
	"sync"
	"time"
)

// Metrics provides observability counters and histograms
// In production, this would integrate with Prometheus or similar
type Metrics struct {
	mu sync.RWMutex

	// Publish metrics
	publishTotal   int64
	publishSuccess int64
	publishFailure int64

	// Validation metrics
	validationTotal        int64
	validationFailureCount int64

	// Rollback metrics
	rollbackTotal        int64
	rollbackSuccess      int64
	rollbackFailure      int64
	rollbackLatencies    []time.Duration
	rollbackLatenciesP95 time.Duration

	// Public endpoint metrics
	publicGetTotal     int64
	publicGetSuccess   int64
	publicGetFailure   int64
	publicGetLatencies []time.Duration
	publicGetP95       time.Duration

	// HTTP request metrics (all routes)
	httpTotal        int64
	http2xx          int64
	http4xx          int64
	http5xx          int64
	httpSlow         int64
	httpLatencies    []time.Duration
	httpLatenciesP95 time.Duration
}

// Global singleton instance
var globalMetrics = &Metrics{
	rollbackLatencies:  make([]time.Duration, 0, 1000),
	publicGetLatencies: make([]time.Duration, 0, 1000),
	httpLatencies:      make([]time.Duration, 0, 1000),
}

// Global returns the global metrics instance
func Global() *Metrics {
	return globalMetrics
}

// RecordPublishAttempt increments publish total counter
func (m *Metrics) RecordPublishAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishTotal++
}

// RecordPublishSuccess increments publish success counter
func (m *Metrics) RecordPublishSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishSuccess++
}

// RecordPublishFailure increments publish failure counter
func (m *Metrics) RecordPublishFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishFailure++
}

// RecordValidationAttempt increments validation total counter
func (m *Metrics) RecordValidationAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validationTotal++
}

// RecordValidationFailure increments validation failure counter (invalid result)
func (m *Metrics) RecordValidationFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validationFailureCount++
}

// RecordRollbackAttempt increments rollback total counter
func (m *Metrics) RecordRollbackAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollbackTotal++
}

// RecordRollbackSuccess increments rollback success counter
func (m *Metrics) RecordRollbackSuccess(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollbackSuccess++
	m.rollbackLatencies = append(m.rollbackLatencies, latency)
	m.updateRollbackP95()
}

// RecordRollbackFailure increments rollback failure counter
func (m *Metrics) RecordRollbackFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollbackFailure++
}

// RecordPublicGetAttempt increments public endpoint total counter
func (m *Metrics) RecordPublicGetAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publicGetTotal++
}

// RecordPublicGetSuccess increments public endpoint success counter with latency
func (m *Metrics) RecordPublicGetSuccess(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publicGetSuccess++
	m.publicGetLatencies = append(m.publicGetLatencies, latency)
	m.updatePublicGetP95()
}

// RecordPublicGetFailure increments public endpoint failure counter
func (m *Metrics) RecordPublicGetFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publicGetFailure++
}

// GetPublishMetrics returns current publish metrics snapshot
func (m *Metrics) GetPublishMetrics() (total, success, failure int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publishTotal, m.publishSuccess, m.publishFailure
}

// GetValidationMetrics returns current validation metrics snapshot
func (m *Metrics) GetValidationMetrics() (total, failures int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.validationTotal, m.validationFailureCount
}

// GetRollbackMetrics returns current rollback metrics snapshot
func (m *Metrics) GetRollbackMetrics() (total, success, failure int64, p95 time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rollbackTotal, m.rollbackSuccess, m.rollbackFailure, m.rollbackLatenciesP95
}

// GetPublicGetMetrics returns current public endpoint metrics snapshot
func (m *Metrics) GetPublicGetMetrics() (total, success, failure int64, p95 time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.publicGetTotal, m.publicGetSuccess, m.publicGetFailure, m.publicGetP95
}

// RecordHTTPRequest records a completed HTTP request for ops dashboards.
// Slow threshold mirrors middleware.DefaultSlowRequestThreshold (500ms).
func (m *Metrics) RecordHTTPRequest(status int, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.httpTotal++
	switch {
	case status >= 500:
		m.http5xx++
	case status >= 400:
		m.http4xx++
	case status >= 200 && status < 300:
		m.http2xx++
	}
	if latency >= 500*time.Millisecond {
		m.httpSlow++
	}
	m.httpLatencies = append(m.httpLatencies, latency)
	m.updateHTTPP95()
}

// GetHTTPMetrics returns aggregate HTTP request metrics.
func (m *Metrics) GetHTTPMetrics() (total, s2xx, s4xx, s5xx, slow int64, p95 time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.httpTotal, m.http2xx, m.http4xx, m.http5xx, m.httpSlow, m.httpLatenciesP95
}

func (m *Metrics) updateHTTPP95() {
	if len(m.httpLatencies) == 0 {
		m.httpLatenciesP95 = 0
		return
	}
	if len(m.httpLatencies) > 1000 {
		m.httpLatencies = m.httpLatencies[len(m.httpLatencies)-1000:]
	}
	sorted := make([]time.Duration, len(m.httpLatencies))
	copy(sorted, m.httpLatencies)
	sortDurations(sorted)
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	m.httpLatenciesP95 = sorted[p95Index]
}

// updateRollbackP95 calculates p95 from latency samples (must hold lock)
func (m *Metrics) updateRollbackP95() {
	if len(m.rollbackLatencies) == 0 {
		m.rollbackLatenciesP95 = 0
		return
	}
	// Keep last 1000 samples to avoid unbounded memory growth
	if len(m.rollbackLatencies) > 1000 {
		m.rollbackLatencies = m.rollbackLatencies[len(m.rollbackLatencies)-1000:]
	}
	// Simple percentile: sort and take 95th index
	sorted := make([]time.Duration, len(m.rollbackLatencies))
	copy(sorted, m.rollbackLatencies)
	sortDurations(sorted)
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	m.rollbackLatenciesP95 = sorted[p95Index]
}

// updatePublicGetP95 calculates p95 from latency samples (must hold lock)
func (m *Metrics) updatePublicGetP95() {
	if len(m.publicGetLatencies) == 0 {
		m.publicGetP95 = 0
		return
	}
	// Keep last 1000 samples to avoid unbounded memory growth
	if len(m.publicGetLatencies) > 1000 {
		m.publicGetLatencies = m.publicGetLatencies[len(m.publicGetLatencies)-1000:]
	}
	// Simple percentile: sort and take 95th index
	sorted := make([]time.Duration, len(m.publicGetLatencies))
	copy(sorted, m.publicGetLatencies)
	sortDurations(sorted)
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	m.publicGetP95 = sorted[p95Index]
}

// sortDurations is a simple bubble sort for time.Duration slices
func sortDurations(arr []time.Duration) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}
