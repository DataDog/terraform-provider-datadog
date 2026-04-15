package utils

import (
	"context"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// newTestClient returns an APIClient with the default configuration.
// Tests use this to call buildRequest without making real HTTP calls;
// they inspect the returned *http.Request headers directly.
func newTestClient() *datadog.APIClient {
	return datadog.NewAPIClient(datadog.NewConfiguration())
}

func TestBuildRequest_APIKeyAuth(t *testing.T) {
	ctx := context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: "test-api-key"},
		"appKeyAuth": {Key: "test-app-key"},
	})

	req, err := buildRequest(ctx, newTestClient(), "GET", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}

	if got := req.Header.Get("DD-API-KEY"); got != "test-api-key" {
		t.Errorf("DD-API-KEY: want %q, got %q", "test-api-key", got)
	}
	if got := req.Header.Get("DD-APPLICATION-KEY"); got != "test-app-key" {
		t.Errorf("DD-APPLICATION-KEY: want %q, got %q", "test-app-key", got)
	}
}

// TestBuildRequest_DelegatedTokenAuth is the regression test for the cloud-auth bug:
// buildRequest was not calling UseDelegatedTokenAuth, so POST /api/v1/dashboard
// (and other SendRequest callers) went out with no Authorization header when
// cloud_provider_type = "aws" was set.
//
// Also verifies that no DD-API-KEY / DD-APPLICATION-KEY headers are emitted
// under cloud auth (no key leakage).
func TestBuildRequest_DelegatedTokenAuth(t *testing.T) {
	// Pre-seed a valid, non-expired delegated token so UseDelegatedTokenAuth
	// doesn't try to reach out to the external-token-servicer.
	creds := &datadog.DelegatedTokenCredentials{
		DelegatedToken: "test-delegated-bearer-token",
		Expiration:     time.Now().Add(15 * time.Minute),
	}
	ctx := context.WithValue(context.Background(), datadog.ContextDelegatedToken, creds)

	req, err := buildRequest(ctx, newTestClient(), "POST", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}

	want := "Bearer test-delegated-bearer-token"
	if got := req.Header.Get("Authorization"); got != want {
		t.Errorf("Authorization: want %q, got %q", want, got)
	}
	if got := req.Header.Get("DD-API-KEY"); got != "" {
		t.Errorf("DD-API-KEY should be absent with cloud auth, got %q", got)
	}
	if got := req.Header.Get("DD-APPLICATION-KEY"); got != "" {
		t.Errorf("DD-APPLICATION-KEY should be absent with cloud auth, got %q", got)
	}
}

func TestBuildRequest_NoAuth(t *testing.T) {
	req, err := buildRequest(context.Background(), newTestClient(), "GET", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}

	if got := req.Header.Get("Authorization"); got != "" {
		t.Errorf("Authorization should be absent, got %q", got)
	}
	if got := req.Header.Get("DD-API-KEY"); got != "" {
		t.Errorf("DD-API-KEY should be absent, got %q", got)
	}
	if got := req.Header.Get("DD-APPLICATION-KEY"); got != "" {
		t.Errorf("DD-APPLICATION-KEY should be absent, got %q", got)
	}
}
