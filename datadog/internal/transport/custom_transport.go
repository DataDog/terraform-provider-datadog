package transport

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	defaultHTTPRetryDuration = 5 * time.Second
	defaultHTTPRetryTimeout  = 60 * time.Second
	rateLimitResetHeader     = "X-Ratelimit-Reset"
)

// CustomTransport holds DefaultTransport configuration and is used to for custom http error handling
type CustomTransport struct {
	defaultTransport  http.RoundTripper
	hTTPRetryDuration time.Duration
	hTTPRetryTimeout  time.Duration
}

// CustomTransportOptions Set options for CustomTransport
type CustomTransportOptions struct {
	Timeout *time.Duration
}

// RoundTrip method used to retry http errors
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var ccancel context.CancelFunc
	ctx := req.Context()
	if _, set := ctx.Deadline(); !set {
		ctx, ccancel = context.WithTimeout(ctx, t.hTTPRetryTimeout)
		defer ccancel()
	}

	retryCount := 0
	for {
		newRequest := t.copyRequest(req)
		resp, respErr := t.defaultTransport.RoundTrip(newRequest)
		// Close the body so connection can be re-used
		localVarBody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
		if respErr != nil {
			return resp, respErr
		}

		// Check if request should be retried and get retry time
		retryDuration, retry := t.retryRequest(resp)
		if !retry {
			return resp, respErr
		}

		// Calculate retryDuration if nil
		if retryDuration == nil {
			newRetryDurationVal := time.Duration(retryCount) * t.hTTPRetryDuration
			retryDuration = &newRetryDurationVal
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
		retryDuration := time.Duration(vInt) * time.Second
		return &retryDuration, true
	}

	if response.StatusCode >= 500 {
		return nil, true
	}

	return nil, false
}

// NewCustomTransport returns new CustomTransport struct
func NewCustomTransport(t http.RoundTripper, opt CustomTransportOptions) *CustomTransport {
	// Use default transport if one provided is nil
	if t == nil {
		t = http.DefaultTransport
	}

	ct := CustomTransport{
		defaultTransport:  t,
		hTTPRetryDuration: defaultHTTPRetryDuration,
	}

	if opt.Timeout != nil {
		ct.hTTPRetryTimeout = *opt.Timeout
	} else {
		ct.hTTPRetryTimeout = defaultHTTPRetryTimeout
	}

	return &ct
}
