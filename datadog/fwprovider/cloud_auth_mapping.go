package fwprovider

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// A newly created mapping can intermittently return 404 even after a successful
// GET. Retry before treating it as deleted so a refresh cannot orphan it.
func readCloudAuthMappingWithRetry[T any](ctx context.Context, timeout time.Duration, read func() (T, *http.Response, error)) (T, *http.Response, error) {
	var result T
	var httpResponse *http.Response
	// RetryContext can return on cancellation while its callback is still running.
	var resultMu sync.Mutex
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		readResponse, readHTTPResponse, err := read()
		resultMu.Lock()
		result, httpResponse = readResponse, readHTTPResponse
		resultMu.Unlock()
		if err == nil {
			return nil
		}
		if readHTTPResponse != nil && readHTTPResponse.StatusCode == http.StatusNotFound {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	})
	resultMu.Lock()
	defer resultMu.Unlock()
	return result, httpResponse, err
}
