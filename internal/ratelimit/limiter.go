package ratelimit

import (
	"sync"
	"time"

	"github.com/leofvo/bridgr/pkg/logger"
)

// Limiter implements a token bucket rate limiter
type Limiter struct {
	rate       float64
	bucketSize float64
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewLimiter creates a new rate limiter with the specified requests per second
func NewLimiter(requestsPerSecond float64) *Limiter {
	// Set bucket size to allow for some burst (1.5x the rate)
	bucketSize := requestsPerSecond * 1.5
	logger.Info("Initializing rate limiter: rate=%.2f req/s, bucket_size=%.2f", requestsPerSecond, bucketSize)
	return &Limiter{
		rate:       requestsPerSecond,
		bucketSize: bucketSize,
		tokens:     bucketSize, // Start with a full bucket
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.lastRefill = now

	// Add new tokens based on elapsed time
	l.tokens = min(l.bucketSize, l.tokens+elapsed*l.rate)

	// Check if we have enough tokens
	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		logger.Debug("Rate limit: request allowed, tokens remaining: %.2f, rate: %.2f req/s", l.tokens, l.rate)
		return true
	}

	logger.Info("Rate limit: request denied, tokens available: %.2f, rate: %.2f req/s", l.tokens, l.rate)
	return false
}

// Wait blocks until a request is allowed
func (l *Limiter) Wait() {
	start := time.Now()
	attempts := 0

	for !l.Allow() {
		attempts++
		// Calculate sleep duration based on rate
		// Sleep for a fraction of the time between requests
		sleepDuration := time.Duration(float64(time.Second) / l.rate / 5)
		time.Sleep(sleepDuration)

		// Log if we've been waiting for a while
		if attempts%5 == 0 {
			logger.Info("Rate limit: waiting for token, elapsed: %v, attempts: %d, rate: %.2f req/s", 
				time.Since(start), attempts, l.rate)
		}
	}

	if attempts > 0 {
		logger.Info("Rate limit: request allowed after %v and %d attempts, rate: %.2f req/s", 
			time.Since(start), attempts, l.rate)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
} 