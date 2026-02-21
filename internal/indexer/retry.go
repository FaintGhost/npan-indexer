package indexer

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"net"
	neturl "net/url"
	"strings"
	"syscall"
	"time"

	"github.com/meilisearch/meilisearch-go"

	"npan/internal/models"
	"npan/internal/npan"
)

func isRetriable(err error) bool {
	if err == nil {
		return false
	}

	// 先处理常见网络瞬时错误，避免长任务被偶发断链直接打断。
	if isRetriableNetworkError(err) {
		return true
	}

	var statusErr *npan.StatusError
	if errors.As(err, &statusErr) {
		status := statusErr.Status
		return status == 429 || (status >= 500 && status <= 599)
	}

	var meiliErr *meilisearch.Error
	if errors.As(err, &meiliErr) {
		switch meiliErr.ErrCode {
		case meilisearch.MeilisearchTimeoutError,
			meilisearch.MeilisearchCommunicationError:
			return true
		case meilisearch.MeilisearchApiError,
			meilisearch.MeilisearchApiErrorWithoutMessage:
			return meiliErr.StatusCode == 429 ||
				(meiliErr.StatusCode >= 500 && meiliErr.StatusCode <= 599)
		}
	}

	return false
}

func isRetriableNetworkError(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	if errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}

	var urlErr *neturl.Error
	if errors.As(err, &urlErr) && urlErr != nil {
		if isRetriableNetworkError(urlErr.Err) {
			return true
		}
		if urlErr.Timeout() {
			return true
		}
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr != nil {
		if isRetriableNetworkError(opErr.Err) {
			return true
		}
		if opErr.Timeout() {
			return true
		}
	}

	message := strings.ToLower(err.Error())
	if strings.Contains(message, "connection reset by peer") ||
		strings.Contains(message, "broken pipe") ||
		strings.Contains(message, "use of closed network connection") ||
		strings.Contains(message, "connection refused") ||
		strings.Contains(message, "timeout") {
		return true
	}

	return false
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

func WithRetryVoid(ctx context.Context, operation func() error, opts models.RetryPolicyOptions) error {
	_, err := WithRetry(ctx, func() (struct{}, error) {
		return struct{}{}, operation()
	}, opts)
	return err
}
