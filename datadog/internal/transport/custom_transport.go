package transport

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

var (
	defaultRetryDuration = 5 * time.Second
	defaultTimeout       = 3600 * time.Second
	rateLimitResetHeader = "X-Ratelimit-Reset"
)

// CustomTransport holds DefaultTransport configuration and is used to for custom http error handling
type CustomTransport struct {
	defaultTransport http.RoundTripper
}

// RoundTrip method used to retry http errors
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var ccancel context.CancelFunc
	ctx := req.Context()
	if _, set := ctx.Deadline(); !set {
		ctx, ccancel = context.WithTimeout(ctx, defaultTimeout)
		defer ccancel()
	}

	retryCount := 0
	for {
		newRequest := t.copyRequest(req)
		resp, respErr := t.defaultTransport.RoundTrip(newRequest)
		if respErr != nil {
			return resp, respErr
		}

		// Check if request should be retried and long
		retryDuration, retry := t.retryRequest(resp)
		if !retry {
			return resp, respErr
		}

		if retryDuration == nil {
			newVal := time.Duration(retryCount) * defaultRetryDuration
			retryDuration = &newVal
		}

		select {
		case <-ctx.Done():
			return resp, respErr
		case <-time.After(*retryDuration):
			retryCount++
			continue
		}
	}
}

func (t *CustomTransport) copyRequest(r *http.Request) *http.Request {
	newRequest := *r

	if r.Body == nil || r.Body == http.NoBody {
		return &newRequest
	}

	body, _ := r.GetBody()
	newRequest.Body = body

	return &newRequest
}

func (t *CustomTransport) retryRequest(response *http.Response) (*time.Duration, bool) {
	if v := response.Header.Get(rateLimitResetHeader); v != "" && response.StatusCode == 429 {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, true
		}
		retryDuration := time.Duration(vInt)
		return &retryDuration, true
	}

	if response.StatusCode >= 500 {
		return nil, true
	}

	return nil, false
}

// NewCustomTransport returns new CustomTransport struct from existing http.Client
func NewCustomTransport() *CustomTransport {

	return &CustomTransport{
		defaultTransport: http.DefaultTransport,
	}
}
