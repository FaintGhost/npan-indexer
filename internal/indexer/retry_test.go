package indexer

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"

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
