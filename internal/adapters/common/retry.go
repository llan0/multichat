package common

import (
	"context"
	"time"
)

type RetryConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// exponential backoff retry
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn func() error) error {
	delay := cfg.InitialDelay

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		// wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// exponential backoff
		delay = time.Duration(float64(delay) * cfg.Multiplier)
		delay = min(delay, cfg.MaxDelay)
	}
}
