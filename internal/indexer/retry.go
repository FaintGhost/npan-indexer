package indexer

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"npan/internal/models"
	"npan/internal/npan"
)

func isRetriable(err error) bool {
	if err == nil {
		return false
	}

	var statusErr *npan.StatusError
	if !errors.As(err, &statusErr) {
		return false
	}

	status := statusErr.Status
	return status == 429 || (status >= 500 && status <= 599)
}

func computeDelay(attempt int, opts models.RetryPolicyOptions) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}

	base := opts.BaseDelayMS
	if base <= 0 {
		base = 1
	}
	maxDelay := opts.MaxDelayMS
	if maxDelay <= 0 {
		maxDelay = base
	}

	raw := base << (attempt - 1)
	if raw > maxDelay {
		raw = maxDelay
	}

	jitter := 0
	if opts.JitterMS > 0 {
		jitter = rand.Intn(opts.JitterMS + 1)
	}

	return time.Duration(raw+jitter) * time.Millisecond
}

func WithRetry[T any](ctx context.Context, operation func() (T, error), opts models.RetryPolicyOptions) (T, error) {
	var zero T
	attempt := 0

	for {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		attempt++
		if !isRetriable(err) || attempt > opts.MaxRetries {
			return zero, err
		}

		delay := computeDelay(attempt, opts)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}
}
