package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// newTestClientAndCtx returns an APIClient pointed at srv, plus a base context
// configured to route requests there via ContextServerIndex=1 / ContextServerVariables.
func newTestClientAndCtx(srv *httptest.Server) (*datadog.APIClient, context.Context) {
	cfg := datadog.NewConfiguration()
	cfg.HTTPClient = srv.Client()
	client := datadog.NewAPIClient(cfg)

	ctx := context.Background()
	ctx = context.WithValue(ctx, datadog.ContextServerIndex, 1)
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{
		"name":     srv.Listener.Addr().String(),
		"protocol": "http",
	})
	return client, ctx
}

func TestBuildRequest_APIKeyAuth(t *testing.T) {
	var gotAPIKey, gotAppKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("DD-API-KEY")
		gotAppKey = r.Header.Get("DD-APPLICATION-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, ctx := newTestClientAndCtx(srv)
	ctx = context.WithValue(ctx, datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: "test-api-key"},
		"appKeyAuth": {Key: "test-app-key"},
	})

	req, err := buildRequest(ctx, client, "GET", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}
	if _, err := client.CallAPI(req); err != nil {
		t.Fatalf("CallAPI error: %v", err)
	}

	if gotAPIKey != "test-api-key" {
		t.Errorf("DD-API-KEY: want %q, got %q", "test-api-key", gotAPIKey)
	}
	if gotAppKey != "test-app-key" {
		t.Errorf("DD-APPLICATION-KEY: want %q, got %q", "test-app-key", gotAppKey)
	}
}

// TestBuildRequest_DelegatedTokenAuth is the regression test for the cloud-auth bug:
// buildRequest was not calling UseDelegatedTokenAuth, so POST /api/v1/dashboard
// (and other SendRequest callers) went out with no Authorization header when
// cloud_provider_type = "aws" was set.
func TestBuildRequest_DelegatedTokenAuth(t *testing.T) {
	var gotAuthorization string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, ctx := newTestClientAndCtx(srv)

	// Pre-seed a valid, non-expired delegated token so UseDelegatedTokenAuth
	// doesn't try to reach out to the external-token-servicer.
	creds := &datadog.DelegatedTokenCredentials{
		DelegatedToken: "test-delegated-bearer-token",
		Expiration:     time.Now().Add(15 * time.Minute),
	}
	ctx = context.WithValue(ctx, datadog.ContextDelegatedToken, creds)

	req, err := buildRequest(ctx, client, "POST", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}
	if _, err := client.CallAPI(req); err != nil {
		t.Fatalf("CallAPI error: %v", err)
	}

	want := "Bearer test-delegated-bearer-token"
	if gotAuthorization != want {
		t.Errorf("Authorization: want %q, got %q", want, gotAuthorization)
	}
}

// TestBuildRequest_DelegatedTokenAuth_NoAPIKeyLeakage ensures that when cloud
// auth is active, no DD-API-KEY or DD-APPLICATION-KEY headers are emitted.
func TestBuildRequest_DelegatedTokenAuth_NoAPIKeyLeakage(t *testing.T) {
	var gotAPIKey, gotAppKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("DD-API-KEY")
		gotAppKey = r.Header.Get("DD-APPLICATION-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, ctx := newTestClientAndCtx(srv)
	creds := &datadog.DelegatedTokenCredentials{
		DelegatedToken: "test-delegated-bearer-token",
		Expiration:     time.Now().Add(15 * time.Minute),
	}
	ctx = context.WithValue(ctx, datadog.ContextDelegatedToken, creds)

	req, err := buildRequest(ctx, client, "POST", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}
	if _, err := client.CallAPI(req); err != nil {
		t.Fatalf("CallAPI error: %v", err)
	}

	if gotAPIKey != "" {
		t.Errorf("DD-API-KEY should be absent with cloud auth, got %q", gotAPIKey)
	}
	if gotAppKey != "" {
		t.Errorf("DD-APPLICATION-KEY should be absent with cloud auth, got %q", gotAppKey)
	}
}

func TestBuildRequest_NoAuth(t *testing.T) {
	var gotAuthorization, gotAPIKey, gotAppKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		gotAPIKey = r.Header.Get("DD-API-KEY")
		gotAppKey = r.Header.Get("DD-APPLICATION-KEY")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, ctx := newTestClientAndCtx(srv)
	// No auth values in context.

	req, err := buildRequest(ctx, client, "GET", "/api/v1/dashboard", nil)
	if err != nil {
		t.Fatalf("buildRequest error: %v", err)
	}
	if _, err := client.CallAPI(req); err != nil {
		t.Fatalf("CallAPI error: %v", err)
	}

	if gotAuthorization != "" {
		t.Errorf("Authorization should be absent, got %q", gotAuthorization)
	}
	if gotAPIKey != "" {
		t.Errorf("DD-API-KEY should be absent, got %q", gotAPIKey)
	}
	if gotAppKey != "" {
		t.Errorf("DD-APPLICATION-KEY should be absent, got %q", gotAppKey)
	}
}
