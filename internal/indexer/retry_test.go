package indexer

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"

	meilisearch "github.com/meilisearch/meilisearch-go"

	"npan/internal/models"
)

func TestWithRetry_RetryOnNetworkResetByPeer(t *testing.T) {
	t.Parallel()

	attempts := 0
	_, err := WithRetry(context.Background(), func() (int, error) {
		attempts++
		if attempts == 1 {
			return 0, &net.OpError{
				Op:  "read",
				Net: "tcp",
				Err: syscall.ECONNRESET,
			}
		}
		return 1, nil
	}, models.RetryPolicyOptions{
		MaxRetries:  2,
		BaseDelayMS: 1,
		MaxDelayMS:  1,
	})
	if err != nil {
		t.Fatalf("expected retry success on network reset, got err: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestWithRetry_NoRetryOnGenericError(t *testing.T) {
	t.Parallel()

	attempts := 0
	_, err := WithRetry(context.Background(), func() (int, error) {
		attempts++
		return 0, errors.New("validation failed")
	}, models.RetryPolicyOptions{
		MaxRetries:  3,
		BaseDelayMS: 1,
		MaxDelayMS:  1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected no retry, got attempts=%d", attempts)
	}
}

func TestIsRetriable_MeiliSearchTimeout(t *testing.T) {
	t.Parallel()

	err := &meilisearch.Error{ErrCode: meilisearch.MeilisearchTimeoutError}
	if !isRetriable(err) {
		t.Fatal("expected MeilisearchTimeoutError to be retriable")
	}
}

func TestIsRetriable_MeiliSearch429(t *testing.T) {
	t.Parallel()

	err := &meilisearch.Error{ErrCode: meilisearch.MeilisearchApiError, StatusCode: 429}
	if !isRetriable(err) {
		t.Fatal("expected MeilisearchApiError with status 429 to be retriable")
	}
}

func TestIsRetriable_MeiliSearch503(t *testing.T) {
	t.Parallel()

	err := &meilisearch.Error{ErrCode: meilisearch.MeilisearchApiError, StatusCode: 503}
	if !isRetriable(err) {
		t.Fatal("expected MeilisearchApiError with status 503 to be retriable")
	}
}

func TestIsRetriable_MeiliSearch400(t *testing.T) {
	t.Parallel()

	err := &meilisearch.Error{ErrCode: meilisearch.MeilisearchApiError, StatusCode: 400}
	if isRetriable(err) {
		t.Fatal("expected MeilisearchApiError with status 400 to NOT be retriable")
	}
}

func TestIsRetriable_MeiliSearchCommunicationError(t *testing.T) {
	t.Parallel()

	err := &meilisearch.Error{ErrCode: meilisearch.MeilisearchCommunicationError}
	if !isRetriable(err) {
		t.Fatal("expected MeilisearchCommunicationError to be retriable")
	}
}

func TestWithRetryVoid_Success(t *testing.T) {
	t.Parallel()

	err := WithRetryVoid(context.Background(), func() error {
		return nil
	}, models.RetryPolicyOptions{
		MaxRetries:  2,
		BaseDelayMS: 1,
		MaxDelayMS:  1,
	})
	if err != nil {
		t.Fatalf("expected nil error on first-try success, got: %v", err)
	}
}

func TestWithRetryVoid_RetryThenSuccess(t *testing.T) {
	t.Parallel()

	attempts := 0
	err := WithRetryVoid(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return &net.OpError{
				Op:  "read",
				Net: "tcp",
				Err: syscall.ECONNRESET,
			}
		}
		return nil
	}, models.RetryPolicyOptions{
		MaxRetries:  2,
		BaseDelayMS: 1,
		MaxDelayMS:  1,
	})
	if err != nil {
		t.Fatalf("expected nil error after retry, got: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestWithRetryVoid_PermanentFailure(t *testing.T) {
	t.Parallel()

	attempts := 0
	err := WithRetryVoid(context.Background(), func() error {
		attempts++
		return errors.New("bad request")
	}, models.RetryPolicyOptions{
		MaxRetries:  3,
		BaseDelayMS: 1,
		MaxDelayMS:  1,
	})
	if err == nil {
		t.Fatal("expected non-nil error for permanent failure")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt (no retry on permanent error), got %d", attempts)
	}
}
