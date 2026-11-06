package datadog

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/cloudauth"
)

func TestProviderGitHubConfiguration(t *testing.T) {
	const orgUUID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"

	for _, mode := range []string{"lazy", "validate", "missing-request", "invalid-provider"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "")
			t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", "")

			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				switch r.URL.Path {
				case "/token":
					if r.URL.Query().Get("audience") != "datadog/"+orgUUID {
						t.Errorf("audience = %q", r.URL.Query().Get("audience"))
					}
					if r.Header.Get("Authorization") != "Bearer actions-token" {
						t.Errorf("OIDC authorization = %q", r.Header.Get("Authorization"))
					}
					fmt.Fprint(w, `{"value":"header.payload.signature"}`)
				case "/api/v2/delegated-token":
					if r.Header.Get("Authorization") != "Delegated header.payload.signature" {
						t.Errorf("delegated authorization = %q", r.Header.Get("Authorization"))
					}
					fmt.Fprint(w, `{"data":{"attributes":{"access_token":"test-token","expires":"4102444800"}}}`)
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			values := map[string]interface{}{
				"cloud_provider_type": "github",
				"org_uuid":            orgUUID,
				"api_url":             server.URL,
				"validate":            "false",
			}
			switch mode {
			case "lazy", "validate":
				t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "actions-token")
				t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", server.URL+"/token")
				if mode == "validate" {
					values["validate"] = "true"
				}
			case "missing-request":
				values["validate"] = "true"
			case "invalid-provider":
				values["cloud_provider_type"] = "azure"
			}

			data := schema.TestResourceDataRaw(t, Provider().Schema, values)
			result, diags := providerConfigure(context.Background(), data)
			switch mode {
			case "missing-request", "invalid-provider":
				if !diags.HasError() {
					t.Fatal("expected configuration error")
				}
				if calls.Load() != 0 {
					t.Fatal("failed configuration contacted an endpoint")
				}
				return
			}
			if diags.HasError() {
				t.Fatalf("configure: %v", diags)
			}
			config := result.(*ProviderConfiguration)
			client := config.DatadogApiInstances.HttpClient
			delegated := client.GetConfig().DelegatedTokenConfig
			if delegated == nil || delegated.Provider != cloudauth.GitHubProvider || delegated.OrgUUID != orgUUID {
				t.Fatalf("delegated config = %#v", delegated)
			}
			if _, ok := config.Auth.Value(datadog.ContextDelegatedToken).(*datadog.DelegatedTokenCredentials); !ok {
				t.Fatal("delegated token context missing")
			}
			if config.Auth.Value(datadog.ContextAWSVariables) != nil {
				t.Fatal("AWS credentials context set for GitHub auth")
			}

			wantCalls := int32(0)
			if mode == "validate" {
				wantCalls = 2
			}
			if calls.Load() != wantCalls {
				t.Fatalf("calls after configure = %d, want %d", calls.Load(), wantCalls)
			}
			if mode == "validate" {
				return
			}

			token, err := client.GetDelegatedToken(config.Auth)
			if err != nil {
				t.Fatal(err)
			}
			if token.DelegatedToken != "test-token" || calls.Load() != 2 {
				t.Fatalf("token = %q, calls = %d", token.DelegatedToken, calls.Load())
			}
			if strings.Contains(token.DelegatedProof, "actions-token") {
				t.Fatal("delegated proof contains the GitHub request token")
			}
		})
	}
}
