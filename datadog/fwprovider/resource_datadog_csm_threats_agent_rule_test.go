package fwprovider

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestRetryOnCSMThreatsLocked(t *testing.T) {
	errAPI := errors.New("api error")

	tests := []struct {
		name      string
		responses []int
		wantCalls int
		wantErr   bool
	}{
		{name: "success first try", responses: []int{http.StatusOK}, wantCalls: 1},
		{name: "locked then success", responses: []int{http.StatusLocked, http.StatusLocked, http.StatusOK}, wantCalls: 3},
		{name: "non-locked error is not retried", responses: []int{http.StatusBadRequest}, wantCalls: 1, wantErr: true},
		{name: "nil response is not retried", responses: []int{0}, wantCalls: 1, wantErr: true},
		{name: "locked until timeout", responses: []int{http.StatusLocked}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			err := retryOnCSMThreatsLocked(context.Background(), 2*time.Second, func() (*http.Response, error) {
				status := tt.responses[min(calls, len(tt.responses)-1)]
				calls++
				switch status {
				case http.StatusOK:
					return &http.Response{StatusCode: status}, nil
				case 0:
					return nil, errAPI
				default:
					return &http.Response{StatusCode: status}, errAPI
				}
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantCalls > 0 && calls != tt.wantCalls {
				t.Fatalf("got %d calls, want %d", calls, tt.wantCalls)
			}
			if tt.wantCalls == 0 && calls < 2 {
				t.Fatalf("got %d calls, want retries until timeout", calls)
			}
		})
	}
}
