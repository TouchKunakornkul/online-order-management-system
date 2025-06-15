package retryutil

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ConnectionError represents database connection related errors
type ConnectionError struct {
	Err error
}

func (e ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}

func (e ConnectionError) Unwrap() error {
	return e.Err
}

// IsConnectionError checks if the error is related to database connection limits
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "too many clients already") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "server closed the connection unexpectedly") ||
		strings.Contains(errStr, "no connection to the server")
}

// RetryConfig contains configuration for retry logic
type RetryConfig struct {
	MaxRetries     int
	BaseDelay      time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	RetryCondition func(error) bool
}

// DefaultRetryConfig returns default retry configuration for database operations
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		BaseDelay:      10 * time.Millisecond,
		MaxDelay:       500 * time.Millisecond,
		BackoffFactor:  2.0,
		RetryCondition: IsConnectionError,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with configurable factor
			backoff := time.Duration(float64(config.BaseDelay) *
				(config.BackoffFactor * float64(attempt)))

			if backoff > config.MaxDelay {
				backoff = config.MaxDelay
			}

			select {
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check retry condition
		if config.RetryCondition != nil && !config.RetryCondition(err) {
			return fmt.Errorf("retry condition not met: %w", err)
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}
