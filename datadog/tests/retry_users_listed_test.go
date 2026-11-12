package test

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryUntilUsersListed(t *testing.T) {
	t.Parallel()
	emptyErr := errors.New("user is not disabled")

	t.Run("retries until listed", func(t *testing.T) {
		calls := 0
		err := retryUntilUsersListed(context.Background(), 30*time.Second, func() (int, error) {
			calls++
			if calls == 1 {
				return 0, nil
			}
			return 1, nil
		}, emptyErr)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if calls != 2 {
			t.Fatalf("expected 2 calls, got %d", calls)
		}
	})

	t.Run("times out when never listed", func(t *testing.T) {
		err := retryUntilUsersListed(context.Background(), time.Second, func() (int, error) {
			return 0, nil
		}, emptyErr)
		if err == nil || !errors.Is(err, emptyErr) {
			t.Fatalf("expected %v, got %v", emptyErr, err)
		}
	})

	t.Run("stops on list error", func(t *testing.T) {
		listErr := errors.New("boom")
		calls := 0
		err := retryUntilUsersListed(context.Background(), 30*time.Second, func() (int, error) {
			calls++
			return 0, listErr
		}, emptyErr)
		if !errors.Is(err, listErr) || calls != 1 {
			t.Fatalf("expected one call and %v, got %d calls and %v", listErr, calls, err)
		}
	})
}
