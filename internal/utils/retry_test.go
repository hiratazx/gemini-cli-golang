// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRetryWithBackoff_SuccessFirstTry(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:       3,
		InitialDelay:     10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:           false,
	}

	attempts := 0
	err := RetryWithBackoff(context.Background(), cfg, func(attempt int) error {
		attempts++
		return nil // succeed immediately
	})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetryWithBackoff_SuccessOnRetry(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:       5,
		InitialDelay:     10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:           false,
	}

	attempts := 0
	err := RetryWithBackoff(context.Background(), cfg, func(attempt int) error {
		attempts++
		if attempt < 3 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff_AllFail(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:       3,
		InitialDelay:     10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:           false,
	}

	attempts := 0
	err := RetryWithBackoff(context.Background(), cfg, func(attempt int) error {
		attempts++
		return errors.New("always fail")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if !strings.Contains(err.Error(), "all 3 attempts failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRetryWithBackoff_ContextCancelled(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:       10,
		InitialDelay:     100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:           false,
	}

	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	err := RetryWithBackoff(ctx, cfg, func(attempt int) error {
		attempts++
		if attempt == 2 {
			cancel() // Cancel after second attempt
		}
		return errors.New("fail")
	})

	if err == nil {
		t.Error("expected error on cancel")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected cancellation error, got: %v", err)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != DefaultMaxRetryAttempts {
		t.Errorf("expected %d max attempts, got %d", DefaultMaxRetryAttempts, cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("expected 1s initial delay, got %v", cfg.InitialDelay)
	}
	if !cfg.Jitter {
		t.Error("expected jitter to be enabled by default")
	}
}
