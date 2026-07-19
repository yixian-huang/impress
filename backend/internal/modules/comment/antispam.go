package comment

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/provider"
)

// AntiSpamService provides rate-limiting and keyword-based spam checks for comments.
type AntiSpamService struct {
	captcha    provider.CaptchaProvider
	keywords   []string
	mu         sync.RWMutex
	ipTracker  map[string][]time.Time
	rateLimit  int
	rateWindow time.Duration
	done       chan struct{}
}

func newAntiSpamService(captcha provider.CaptchaProvider) *AntiSpamService {
	svc := &AntiSpamService{
		captcha:    captcha,
		keywords:   []string{},
		ipTracker:  make(map[string][]time.Time),
		rateLimit:  5,
		rateWindow: 10 * time.Minute,
		done:       make(chan struct{}),
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				svc.cleanupTracker()
			case <-svc.done:
				return
			}
		}
	}()
	return svc
}

func (s *AntiSpamService) Stop() {
	close(s.done)
}

func (s *AntiSpamService) Check(ctx context.Context, ip, content, captchaToken string) error {
	if !s.checkRateLimit(ip) {
		return &SpamError{Reason: "rate_limit", Message: "Too many submissions, please try again later"}
	}
	if s.containsBannedKeyword(content) {
		return &SpamError{Reason: "keyword", Message: "Content contains blocked keywords"}
	}
	if captchaToken != "" {
		if err := s.captcha.Verify(ctx, captchaToken, ip); err != nil {
			return &SpamError{Reason: "captcha", Message: "Captcha verification failed"}
		}
	}
	s.recordSubmission(ip)
	return nil
}

func (s *AntiSpamService) checkRateLimit(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	timestamps := s.ipTracker[ip]
	cutoff := time.Now().Add(-s.rateWindow)
	count := 0
	for _, t := range timestamps {
		if t.After(cutoff) {
			count++
		}
	}
	return count < s.rateLimit
}

func (s *AntiSpamService) recordSubmission(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ipTracker[ip] = append(s.ipTracker[ip], time.Now())
}

func (s *AntiSpamService) containsBannedKeyword(content string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lower := strings.ToLower(content)
	for _, kw := range s.keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func (s *AntiSpamService) cleanupTracker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := time.Now().Add(-s.rateWindow)
	for ip, timestamps := range s.ipTracker {
		var active []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				active = append(active, t)
			}
		}
		if len(active) == 0 {
			delete(s.ipTracker, ip)
		} else {
			s.ipTracker[ip] = active
		}
	}
}

// SpamError is returned when a submission is detected as spam.
type SpamError struct {
	Reason  string
	Message string
}

func (e *SpamError) Error() string { return e.Message }
