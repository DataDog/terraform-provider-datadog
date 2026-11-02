package fwprovider

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadCloudAuthMappingWithRetry(t *testing.T) {
	for _, tc := range []struct {
		name     string
		statuses []int
		wantErr  bool
	}{
		{name: "success", statuses: []int{http.StatusOK}},
		{name: "transient not found", statuses: []int{http.StatusNotFound, http.StatusOK}},
		{name: "forbidden", statuses: []int{http.StatusForbidden}, wantErr: true},
		{name: "server error", statuses: []int{http.StatusInternalServerError}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			readErr := errors.New("read failed")
			result, response, err := readCloudAuthMappingWithRetry(context.Background(), time.Second, func() (int, *http.Response, error) {
				status := tc.statuses[min(calls, len(tc.statuses)-1)]
				calls++
				response := &http.Response{StatusCode: status}
				if status != http.StatusOK {
					return calls, response, readErr
				}
				return calls, response, nil
			})
			if (err != nil) != tc.wantErr || (tc.wantErr && !errors.Is(err, readErr)) {
				t.Fatalf("error = %v, want error %t", err, tc.wantErr)
			}
			if calls != len(tc.statuses) || result != calls || response.StatusCode != tc.statuses[len(tc.statuses)-1] {
				t.Fatalf("calls = %d, result = %d, response = %v", calls, result, response)
			}
		})
	}
}

func TestReadCloudAuthMappingWithRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	_, _, err := readCloudAuthMappingWithRetry(ctx, time.Second, func() (int, *http.Response, error) {
		calls.Add(1)
		cancel()
		return 0, &http.Response{StatusCode: http.StatusNotFound}, errors.New("not found")
	})
	// RetryContext may return the last API error when cancellation interrupts a
	// retry. Callers inspect ctx.Err() before treating a 404 as deletion.
	if err == nil || ctx.Err() != context.Canceled || calls.Load() != 1 {
		t.Fatalf("error = %v, context error = %v, calls = %d", err, ctx.Err(), calls.Load())
	}
}
