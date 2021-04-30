package transport

import (
	"context"
	"net/http"
	"time"
)

var retryTime = 10 * time.Second
var defaultTimeout = 60 * time.Second

// CustomTransport holds DefaultTransport configuration and is used to for custom http error handling
type CustomTransport struct {
	defaultTransport http.RoundTripper
}

// RoundTrip method used to retry http errors
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	retryCount := 0

	var ccancel context.CancelFunc
	ctx := req.Context()
	if _, set := ctx.Deadline(); !set {
		ctx, ccancel = context.WithTimeout(ctx, defaultTimeout)
		defer ccancel()
	}

	for {
		newRequest := t.copyRequest(req)
		resp, err := t.defaultTransport.RoundTrip(newRequest)
		if err != nil {
			return resp, err
		}
		if !t.retryRequest(resp) {
			return resp, err
		}
		retryCount++

		select {
		case <-ctx.Done():
			return resp, err
		case <-time.After(retryTime):
			continue
		}
	}
}

func (t *CustomTransport) copyRequest(r *http.Request) *http.Request {
	newRequest := *r
	bd, _ := r.GetBody()

	if r.Body == nil || r.Body == http.NoBody {

		newRequest.Body = bd
	}

	return &newRequest
}

func (t *CustomTransport) retryRequest(response *http.Response) bool {
	if response.StatusCode == 429 {
		return true
	}

	return false
}

// NewCustomTransport returns new CustomTransport struct from existing http.Client
func NewCustomTransport() *CustomTransport {

	return &CustomTransport{
		defaultTransport: http.DefaultTransport,
	}
}
