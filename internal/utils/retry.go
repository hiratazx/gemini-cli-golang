// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// DefaultMaxRetryAttempts is the default maximum number of retry attempts.
const DefaultMaxRetryAttempts = 5

// RetryConfig configures retry behavior.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including the first).
	MaxAttempts int
	// InitialDelay is the delay before the first retry.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// BackoffMultiplier is the multiplier for exponential backoff.
	BackoffMultiplier float64
	// Jitter adds randomization to delays to prevent thundering herd.
	Jitter bool
}

// DefaultRetryConfig returns a default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       DefaultMaxRetryAttempts,
		InitialDelay:     1 * time.Second,
		MaxDelay:          60 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:           true,
	}
}

// RetryWithBackoff retries a function with exponential backoff.
// The function should return nil on success, or an error to trigger a retry.
// Returns the error from the last attempt if all attempts fail.
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn func(attempt int) error) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		lastErr = fn(attempt)
		if lastErr == nil {
			return nil
		}

		// Don't sleep after the last attempt
		if attempt == cfg.MaxAttempts {
			break
		}

		// Calculate delay with exponential backoff
		delay := float64(cfg.InitialDelay) * math.Pow(cfg.BackoffMultiplier, float64(attempt-1))
		if delay > float64(cfg.MaxDelay) {
			delay = float64(cfg.MaxDelay)
		}

		// Add jitter
		if cfg.Jitter {
			jitter := delay * 0.1 * (rand.Float64()*2 - 1) // ±10%
			delay += jitter
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(time.Duration(delay)):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("all %d attempts failed, last error: %w", cfg.MaxAttempts, lastErr)
}
