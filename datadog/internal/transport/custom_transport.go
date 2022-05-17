package transport

import (
	"bytes"
	"context"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"
)

var (
	defaultBackOffMultiplier float64 = 2
	defaultBackOffBase       float64 = 2
	defaultHTTPRetryTimeout          = 60 * time.Second
	rateLimitResetHeader             = "X-Ratelimit-Reset"
)

// CustomTransport holds DefaultTransport configuration and is used to for custom http error handling
type CustomTransport struct {
	defaultTransport http.RoundTripper
	httpRetryTimeout time.Duration
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
		ctx, ccancel = context.WithTimeout(ctx, t.httpRetryTimeout)
		defer ccancel()
	}

	retryCount := 0
	for {
		newRequest := t.copyRequest(req)
		resp, respErr := t.defaultTransport.RoundTrip(newRequest)
		// Close the body so connection can be re-used
		if resp != nil {
			localVarBody, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
		}
		if respErr != nil {
			return resp, respErr
		}

		// Check if request should be retried and get retry time
		retryDuration, retry := t.retryRequest(resp, retryCount)
		if !retry {
			return resp, respErr
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

func (t *CustomTransport) retryRequest(response *http.Response, retryCount int) (*time.Duration, bool) {
	if v := response.Header.Get(rateLimitResetHeader); v != "" && response.StatusCode == 429 {
		vInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, true
		}
		retryDuration := time.Duration(vInt) * time.Second
		return &retryDuration, true
	}

	if response.StatusCode >= 500 {
		// Calculate the retry val (base * multiplier^2)
		retryVal := defaultBackOffBase * math.Pow(defaultBackOffMultiplier, float64(retryCount))
		// retry duration shouldn't exceed default timeout period
		retryVal = math.Min(float64(t.httpRetryTimeout/time.Second), retryVal)
		retryDuration := time.Duration(retryVal) * time.Second
		return &retryDuration, true
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
		defaultTransport: t,
	}

	if opt.Timeout != nil {
		ct.httpRetryTimeout = *opt.Timeout
	} else {
		ct.httpRetryTimeout = defaultHTTPRetryTimeout
	}

	return &ct
}
