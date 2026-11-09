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

// testGithubOIDCClaimMatchers builds a claim matchers model with every claim
// null except the required `sub` and, when nonempty, `repository`. Manually
// constructed models must fill every attribute: a zero-value types.String is
// invalid and cannot be marshaled into a plan or state.
func testGithubOIDCClaimMatchers(sub, repository string) *githubOIDCClaimMatchersModel {
	null := types.StringNull()
	m := githubOIDCClaimMatchersModel{
		Actor:                null,
		ActorId:              null,
		Enterprise:           null,
		EnterpriseId:         null,
		Environment:          null,
		EventName:            null,
		JobWorkflowRef:       null,
		Ref:                  null,
		RefType:              null,
		Repository:           null,
		RepositoryId:         null,
		RepositoryOwner:      null,
		RepositoryOwnerId:    null,
		RepositoryVisibility: null,
		RunnerEnvironment:    null,
		Sub:                  types.StringValue(sub),
		Workflow:             null,
		WorkflowRef:          null,
	}
	if repository != "" {
		m.Repository = types.StringValue(repository)
	}
	return &m
}

func TestGithubOidcIdentityMappingCreate(t *testing.T) {
	const (
		mappingID         = "7c405332-7033-40d2-a046-a27a075a22cd"
		accountIdentifier = "service-account-id"
		accountUUID       = "0bd557d2-6ed0-423e-9654-c0e1cd376abc"
		subClaim          = "repo:example-org/example-repo:ref:refs/heads/main"
		repositoryClaim   = "^example-org/example-repo$"
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
			responseType: "github_oidc_auth_config",
			wantGet:      true,
		},
		{
			name:         "POST response contains an unparsed type",
			responseType: "future_github_oidc_auth_config",
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
				case request.Method == http.MethodPost && request.URL.Path == "/api/v2/wif/github/persona_mapping":
					if test.errorBody != "" {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, test.errorBody)
						return
					}
					w.WriteHeader(http.StatusCreated)
					fmt.Fprintf(w, `{"data":{"id":%q,"type":%q,"attributes":{"account_identifier":%q,"account_uuid":%q,"claim_matchers":{"sub":%q,"repository":%q}}}}`,
						mappingID, test.responseType, accountIdentifier, accountUUID, subClaim, repositoryClaim)
				case request.Method == http.MethodGet && request.URL.Path == "/api/v2/wif/github/persona_mapping/"+mappingID:
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
			config.SetUnstableOperationEnabled("v2.CreateGitHubCloudAuthPersonaMapping", true)
			config.SetUnstableOperationEnabled("v2.GetGitHubCloudAuthPersonaMapping", true)
			r := &githubOidcIdentityMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}

			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			planned := githubOidcIdentityMappingModel{
				ID:                types.StringUnknown(),
				AccountIdentifier: types.StringValue(accountIdentifier),
				AccountUUID:       types.StringUnknown(),
				ClaimMatchers:     testGithubOIDCClaimMatchers(subClaim, repositoryClaim),
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
				if got := diags[0].Summary(); got != "Error creating GitHub OIDC identity mapping" {
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

			var state githubOidcIdentityMappingModel
			if diags := response.State.Get(ctx, &state); diags.HasError() {
				t.Fatalf("reading response state: %v", diags.Errors())
			}
			if got := state.ID.ValueString(); got != mappingID {
				t.Fatalf("id = %q, want %q", got, mappingID)
			}
			if got := state.AccountIdentifier.ValueString(); got != accountIdentifier {
				t.Fatalf("account_identifier = %q, want %q", got, accountIdentifier)
			}
			if got := state.AccountUUID.ValueString(); got != accountUUID {
				t.Fatalf("account_uuid = %q, want %q", got, accountUUID)
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

func TestGithubOidcIdentityMappingUpdateState(t *testing.T) {
	apiResponse := newGithubOidcIdentityMappingResponse(
		"mapping-id",
		"service-account-handle",
		"account-uuid",
		"repo:example-org/example-repo:ref:refs/heads/main",
		"^example-org/example-repo$",
	)
	resource := &githubOidcIdentityMappingResource{}

	t.Run("preserves configured email when API returns handle", func(t *testing.T) {
		state := githubOidcIdentityMappingModel{
			AccountIdentifier: types.StringValue("terraform-service-account@example.com"),
		}

		resource.updateState(&state, apiResponse)

		if got := state.AccountIdentifier.ValueString(); got != "terraform-service-account@example.com" {
			t.Fatalf("account_identifier = %q, want configured email", got)
		}
		if got := state.AccountUUID.ValueString(); got != "account-uuid" {
			t.Fatalf("account_uuid = %q, want account-uuid", got)
		}
	})

	t.Run("uses API handle during import", func(t *testing.T) {
		state := githubOidcIdentityMappingModel{AccountIdentifier: types.StringNull()}

		resource.updateState(&state, apiResponse)

		if got := state.AccountIdentifier.ValueString(); got != "service-account-handle" {
			t.Fatalf("account_identifier = %q, want API handle", got)
		}
	})

	t.Run("maps claim matchers", func(t *testing.T) {
		state := githubOidcIdentityMappingModel{}

		resource.updateState(&state, apiResponse)

		if got := state.ClaimMatchers.Sub.ValueString(); got != "repo:example-org/example-repo:ref:refs/heads/main" {
			t.Fatalf("claim_matchers.sub = %q, want the API sub pattern", got)
		}
		if got := state.ClaimMatchers.Repository.ValueString(); got != "^example-org/example-repo$" {
			t.Fatalf("claim_matchers.repository = %q, want the API repository pattern", got)
		}
		if !state.ClaimMatchers.Actor.IsNull() {
			t.Fatalf("claim_matchers.actor = %q, want null for an unset claim", state.ClaimMatchers.Actor)
		}
	})
}

func newGithubOidcIdentityMappingResponse(id, accountIdentifier, accountUUID, sub, repository string) *datadogV2.GitHubCloudAuthPersonaMappingResponse {
	claimMatchers := datadogV2.NewGitHubOIDCClaimPatterns(sub)
	claimMatchers.SetRepository(repository)
	attributes := datadogV2.NewGitHubCloudAuthPersonaMappingAttributesResponse(accountIdentifier, accountUUID, *claimMatchers)
	data := datadogV2.NewGitHubCloudAuthPersonaMappingDataResponse(
		*attributes,
		id,
		datadogV2.GITHUBCLOUDAUTHPERSONAMAPPINGTYPE_GITHUB_OIDC_AUTH_CONFIG,
	)
	return datadogV2.NewGitHubCloudAuthPersonaMappingResponse(*data)
}

func TestGithubOidcIdentityMappingRead(t *testing.T) {
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
				if request.Method != http.MethodGet || request.URL.Path != "/api/v2/wif/github/persona_mapping/mapping-id" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				index := min(int(requests.Add(1))-1, len(tc.statuses)-1)
				status := tc.statuses[index]
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status == http.StatusOK {
					if err := json.NewEncoder(w).Encode(newGithubOidcIdentityMappingResponse(
						"mapping-id",
						"handle",
						"account-uuid",
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
			config.SetUnstableOperationEnabled("v2.GetGitHubCloudAuthPersonaMapping", true)
			r := &githubOidcIdentityMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			prior := githubOidcIdentityMappingModel{
				ID:                types.StringValue("mapping-id"),
				AccountIdentifier: types.StringValue("configured@example.com"),
				AccountUUID:       types.StringValue("account-uuid"),
				ClaimMatchers:     testGithubOIDCClaimMatchers("repo:example-org/example-repo:ref:refs/heads/main", "^example-org/example-repo$"),
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
				t.Fatal("read did not preserve the existing mapping and configured identifier")
			}
			if tc.wantError && requests.Load() != 1 {
				t.Fatalf("unexpected retries: got %d requests", requests.Load())
			}
		})
	}
}
