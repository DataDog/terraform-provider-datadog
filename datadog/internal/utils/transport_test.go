package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResourceUserAgentTransport(t *testing.T) {
	tests := []struct {
		name         string
		resourceName string // empty means no context value set
		inputUA      string
		wantUA       string
	}{
		{
			name:         "appends terraform_resource comment",
			resourceName: "datadog_dashboard",
			inputUA:      "terraform-provider-datadog/3.39.0 (terraform 1.5.7; os linux; arch amd64)",
			wantUA:       "terraform-provider-datadog/3.39.0 (terraform 1.5.7; os linux; arch amd64; terraform_resource datadog_dashboard)",
		},
		{
			name:    "no modification without context",
			inputUA: "terraform-provider-datadog/3.39.0 (terraform 1.5.7; os linux; arch amd64)",
			wantUA:  "terraform-provider-datadog/3.39.0 (terraform 1.5.7; os linux; arch amd64)",
		},
		{
			name:         "no modification when UA has no comment section",
			resourceName: "datadog_monitor",
			inputUA:      "terraform-provider-datadog/3.39.0",
			wantUA:       "terraform-provider-datadog/3.39.0",
		},
		{
			name:         "does not mutate original request",
			resourceName: "datadog_monitor",
			inputUA:      "terraform-provider-datadog/3.39.0 (terraform 1.5.7)",
			wantUA:       "terraform-provider-datadog/3.39.0 (terraform 1.5.7; terraform_resource datadog_monitor)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUA string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUA = r.Header.Get("User-Agent")
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			client := &http.Client{
				Transport: WrapTransportWithResourceUserAgent(http.DefaultTransport),
			}

			ctx := context.Background()
			if tt.resourceName != "" {
				ctx = WithTerraformResource(ctx, tt.resourceName)
			}
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
			req.Header.Set("User-Agent", tt.inputUA)

			origUA := req.Header.Get("User-Agent")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			resp.Body.Close()

			if gotUA != tt.wantUA {
				t.Errorf("expected User-Agent = %q, got %q", tt.wantUA, gotUA)
			}

			// Verify original request was not mutated
			if tt.resourceName != "" {
				afterUA := req.Header.Get("User-Agent")
				if origUA != afterUA {
					t.Errorf("original request was mutated: before=%q after=%q", origUA, afterUA)
				}
			}
		})
	}
}
