package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResourceHeaderTransport_SetsHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("DD-Terraform-Resource")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: WrapTransportWithResourceHeader(http.DefaultTransport),
	}
	ctx := WithTerraformResource(context.Background(), "datadog_dashboard")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotHeader != "datadog_dashboard" {
		t.Errorf("expected DD-Terraform-Resource = %q, got %q", "datadog_dashboard", gotHeader)
	}
}

func TestResourceHeaderTransport_NoHeaderWithoutContext(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("DD-Terraform-Resource")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: WrapTransportWithResourceHeader(http.DefaultTransport),
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if gotHeader != "" {
		t.Errorf("expected no DD-Terraform-Resource header, got %q", gotHeader)
	}
}

func TestResourceHeaderTransport_DoesNotMutateOriginalRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: WrapTransportWithResourceHeader(http.DefaultTransport),
	}
	ctx := WithTerraformResource(context.Background(), "datadog_monitor")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)

	origHeader := req.Header.Get("DD-Terraform-Resource")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	afterHeader := req.Header.Get("DD-Terraform-Resource")
	if origHeader != "" || afterHeader != "" {
		t.Errorf("original request was mutated: before=%q after=%q", origHeader, afterHeader)
	}
}
