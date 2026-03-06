package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResourceHeaderTransport(t *testing.T) {
	tests := []struct {
		name           string
		resourceName   string // empty means no context value set
		wantHeader     string
		checkNoMutate  bool
	}{
		{
			name:         "sets header from context",
			resourceName: "datadog_dashboard",
			wantHeader:   "datadog_dashboard",
		},
		{
			name:       "no header without context",
			wantHeader: "",
		},
		{
			name:          "does not mutate original request",
			resourceName:  "datadog_monitor",
			wantHeader:    "datadog_monitor",
			checkNoMutate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotHeader string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeader = r.Header.Get("DD-Terraform-Resource")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			client := &http.Client{
				Transport: WrapTransportWithResourceHeader(http.DefaultTransport),
			}

			ctx := context.Background()
			if tt.resourceName != "" {
				ctx = WithTerraformResource(ctx, tt.resourceName)
			}
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)

			origHeader := req.Header.Get("DD-Terraform-Resource")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			resp.Body.Close()

			if gotHeader != tt.wantHeader {
				t.Errorf("expected DD-Terraform-Resource = %q, got %q", tt.wantHeader, gotHeader)
			}

			if tt.checkNoMutate {
				afterHeader := req.Header.Get("DD-Terraform-Resource")
				if origHeader != "" || afterHeader != "" {
					t.Errorf("original request was mutated: before=%q after=%q", origHeader, afterHeader)
				}
			}
		})
	}
}
