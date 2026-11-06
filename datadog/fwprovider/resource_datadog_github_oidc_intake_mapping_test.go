package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGithubOidcIntakeMappingCreate(t *testing.T) {
	const (
		mappingID       = "7c405332-7033-40d2-a046-a27a075a22cd"
		subClaim        = "repo:example-org/example-repo:ref:refs/heads/main"
		repositoryClaim = "^example-org/example-repo$"
	)

	tests := []struct {
		name         string
		errorBody    string
		wantDetail   string
		responseType string
		wantGet      bool
	}{
		{
			name:       "backend validation error with string errors",
			errorBody:  `{"errors":["Invalid claim pattern"]}`,
			wantDetail: "Invalid claim pattern",
		},
		{
			name:       "backend validation error with structured errors",
			errorBody:  `{"errors":[{"status":"400","title":"Bad Request","detail":"Invalid sub claim","source":{"pointer":"/data/attributes/claim_matchers/sub"}}]}`,
			wantDetail: "Invalid sub claim",
		},
		{
			name:         "visibility GET fails",
			responseType: "github_oidc_auth_intake_mapping",
			wantGet:      true,
		},
		{
			name:         "POST response contains an unparsed type",
			responseType: "future_github_oidc_auth_intake_mapping",
			wantGet:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sawGet := false
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				w.Header().Set("Content-Type", "application/json")
				switch {
				case request.Method == http.MethodPost && request.URL.Path == "/api/v2/wif/github/intake_mapping":
					if test.errorBody != "" {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, test.errorBody)
						return
					}
					w.WriteHeader(http.StatusCreated)
					fmt.Fprintf(w, `{"data":{"id":%q,"type":%q,"attributes":{"claim_matchers":{"sub":%q,"repository":%q}}}}`,
						mappingID, test.responseType, subClaim, repositoryClaim)
				case request.Method == http.MethodGet && request.URL.Path == "/api/v2/wif/github/intake_mapping/"+mappingID:
					sawGet = true
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprint(w, `{"errors":[{"detail":"visibility failed"}]}`)
				default:
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			config := datadog.NewConfiguration()
			config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
			config.OperationServers = nil
			config.HTTPClient = server.Client()
			config.SetUnstableOperationEnabled("v2.CreateGitHubCloudAuthIntakeMapping", true)
			config.SetUnstableOperationEnabled("v2.GetGitHubCloudAuthIntakeMapping", true)
			r := &githubOidcIntakeMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}

			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			planned := githubOidcIntakeMappingModel{
				ID:            types.StringUnknown(),
				ClaimMatchers: testGithubOIDCClaimMatchers(subClaim, repositoryClaim),
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			if diags := plan.Set(ctx, &planned); diags.HasError() {
				t.Fatalf("building plan: %v", diags.Errors())
			}

			response := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
			if !response.Diagnostics.HasError() {
				t.Fatal("Create returned no error diagnostic")
			}
			if sawGet != test.wantGet {
				t.Fatalf("visibility GET called = %t, want %t", sawGet, test.wantGet)
			}

			if test.errorBody != "" {
				diags := response.Diagnostics.Errors()
				if len(diags) != 1 {
					t.Fatalf("got %d error diagnostics, want 1: %v", len(diags), diags)
				}
				if got := diags[0].Summary(); got != "Error creating GitHub OIDC intake mapping" {
					t.Fatalf("diagnostic summary = %q", got)
				}
				for _, detail := range []string{test.wantDetail, "400 Bad Request", test.errorBody} {
					if !strings.Contains(diags[0].Detail(), detail) {
						t.Errorf("diagnostic detail = %q, want to contain %q", diags[0].Detail(), detail)
					}
				}
				if got := requests.Load(); got != 1 {
					t.Errorf("got %d requests, want 1", got)
				}
				if !response.State.Raw.Equal(plan.Raw) {
					t.Error("failed create changed state")
				}
				return
			}

			var state githubOidcIntakeMappingModel
			if diags := response.State.Get(ctx, &state); diags.HasError() {
				t.Fatalf("reading response state: %v", diags.Errors())
			}
			if got := state.ID.ValueString(); got != mappingID {
				t.Fatalf("id = %q, want %q", got, mappingID)
			}
			if got := state.ClaimMatchers.Sub.ValueString(); got != subClaim {
				t.Fatalf("claim_matchers.sub = %q, want %q", got, subClaim)
			}
			if got := state.ClaimMatchers.Repository.ValueString(); got != repositoryClaim {
				t.Fatalf("claim_matchers.repository = %q, want %q", got, repositoryClaim)
			}
		})
	}
}

func newGithubOidcIntakeMappingResponse(id, sub, repository string) *datadogV2.GitHubCloudAuthIntakeMappingResponse {
	claimMatchers := datadogV2.NewGitHubOIDCClaimPatterns(sub)
	claimMatchers.SetRepository(repository)
	attributes := datadogV2.NewGitHubCloudAuthIntakeMappingAttributesResponse(*claimMatchers)
	data := datadogV2.NewGitHubCloudAuthIntakeMappingDataResponse(
		*attributes,
		id,
		datadogV2.GITHUBCLOUDAUTHINTAKEMAPPINGTYPE_GITHUB_OIDC_AUTH_INTAKE_MAPPING,
	)
	return datadogV2.NewGitHubCloudAuthIntakeMappingResponse(*data)
}

func TestGithubOidcIntakeMappingRead(t *testing.T) {
	for _, tc := range []struct {
		name        string
		statuses    []int
		cancel      bool
		wantError   bool
		wantRemoved bool
	}{
		{name: "transient 404 preserves mapping", statuses: []int{http.StatusNotFound, http.StatusOK}},
		{name: "persistent 404 removes mapping", statuses: []int{http.StatusNotFound}, wantRemoved: true},
		{name: "forbidden preserves mapping", statuses: []int{http.StatusForbidden}, wantError: true},
		{name: "server error preserves mapping", statuses: []int{http.StatusInternalServerError}, wantError: true},
		{name: "cancellation after 404 preserves mapping", statuses: []int{http.StatusNotFound}, cancel: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodGet || request.URL.Path != "/api/v2/wif/github/intake_mapping/mapping-id" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				index := min(int(requests.Add(1))-1, len(tc.statuses)-1)
				status := tc.statuses[index]
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status == http.StatusOK {
					if err := json.NewEncoder(w).Encode(newGithubOidcIntakeMappingResponse(
						"mapping-id",
						"repo:example-org/example-repo:ref:refs/heads/main",
						"^example-org/example-repo$",
					)); err != nil {
						t.Error(err)
					}
				} else {
					fmt.Fprint(w, `{"errors":[{"detail":"read failed"}]}`)
				}
				if tc.cancel {
					cancel()
				}
			}))
			defer server.Close()
			config := datadog.NewConfiguration()
			config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
			config.OperationServers = nil
			config.HTTPClient = server.Client()
			config.SetUnstableOperationEnabled("v2.GetGitHubCloudAuthIntakeMapping", true)
			r := &githubOidcIntakeMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			prior := githubOidcIntakeMappingModel{
				ID:            types.StringValue("mapping-id"),
				ClaimMatchers: testGithubOIDCClaimMatchers("repo:example-org/example-repo:ref:refs/heads/main", "^example-org/example-repo$"),
			}
			state := tfsdk.State{Schema: schemaResponse.Schema}
			if diags := state.Set(ctx, &prior); diags.HasError() {
				t.Fatalf("building state: %v", diags.Errors())
			}
			response := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &response)
			if response.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("error diagnostics = %v, want error %t", response.Diagnostics.Errors(), tc.wantError)
			}
			if response.State.Raw.IsNull() != tc.wantRemoved {
				t.Fatalf("state removed = %t, want %t", response.State.Raw.IsNull(), tc.wantRemoved)
			}
			if !tc.wantRemoved && !response.State.Raw.Equal(state.Raw) {
				t.Fatal("read did not preserve the existing mapping")
			}
			if tc.wantError && requests.Load() != 1 {
				t.Fatalf("unexpected retries: got %d requests", requests.Load())
			}
		})
	}
}
