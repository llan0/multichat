package common

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	if cfg.InitialDelay != 1*time.Second {
		t.Errorf("InitialDelay = %v, want 1s", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay = %v, want 30s", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("Multiplier = %v, want 2.0", cfg.Multiplier)
	}
}

func TestRetryWithBackoff_ImmediateSuccess(t *testing.T) {
	cfg := RetryConfig{InitialDelay: 10 * time.Millisecond, MaxDelay: 100 * time.Millisecond, Multiplier: 2.0}
	calls := 0

	err := RetryWithBackoff(context.Background(), cfg, func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestRetryWithBackoff_ContextCancelled(t *testing.T) {
	cfg := RetryConfig{InitialDelay: 10 * time.Millisecond, MaxDelay: 100 * time.Millisecond, Multiplier: 2.0}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RetryWithBackoff(ctx, cfg, func() error {
		return errors.New("should not reach here")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestRetryWithBackoff_RetriesUntilSuccess(t *testing.T) {
	cfg := RetryConfig{InitialDelay: 5 * time.Millisecond, MaxDelay: 50 * time.Millisecond, Multiplier: 2.0}
	calls := 0

	err := RetryWithBackoff(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return errors.New("fail")
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}
