package fwprovider

import (
	"context"
	"encoding/json"
	"errors"
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

func TestIsAwsIntegrationPropagationError(t *testing.T) {
	propagationError := datadog.GenericOpenAPIError{
		ErrorBody:    []byte(`{"errors":[{"detail":"AWS Account Id is not integrated with this Datadog account"}]}`),
		ErrorMessage: "400 Bad Request",
	}

	tests := []struct {
		name         string
		err          error
		httpResponse *http.Response
		want         bool
	}{
		{
			name:         "matching API error",
			err:          propagationError,
			httpResponse: &http.Response{StatusCode: http.StatusBadRequest},
			want:         true,
		},
		{
			name:         "wrapped matching API error",
			err:          fmt.Errorf("create failed: %w", propagationError),
			httpResponse: &http.Response{StatusCode: http.StatusBadRequest},
			want:         true,
		},
		{
			name:         "different bad request",
			err:          datadog.GenericOpenAPIError{ErrorBody: []byte(`{"errors":[{"detail":"invalid ARN"}]}`)},
			httpResponse: &http.Response{StatusCode: http.StatusBadRequest},
			want:         false,
		},
		{
			name:         "matching body with non-bad-request status",
			err:          propagationError,
			httpResponse: &http.Response{StatusCode: http.StatusForbidden},
			want:         false,
		},
		{
			name:         "non-API error",
			err:          errors.New("AWS Account Id is not integrated with this Datadog account"),
			httpResponse: &http.Response{StatusCode: http.StatusBadRequest},
			want:         false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isAwsIntegrationPropagationError(test.err, test.httpResponse); got != test.want {
				t.Fatalf("isAwsIntegrationPropagationError() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestAwsWifIdentityMappingCreate(t *testing.T) {
	const (
		mappingID         = "7c405332-7033-40d2-a046-a27a075a22cd"
		accountIdentifier = "service-account-id"
		accountUUID       = "0bd557d2-6ed0-423e-9654-c0e1cd376abc"
		arnPattern        = "arn:aws:sts::123456789012:assumed-role/terraform-runner/*"
	)

	tests := []struct {
		name         string
		errorBody    string
		wantDetail   string
		responseType string
		cancel       bool
		wantGet      bool
	}{
		{
			name:       "backend validation error with string errors",
			errorBody:  `{"errors":["Invalid ARN pattern"]}`,
			wantDetail: "Invalid ARN pattern",
		},
		{
			name:       "backend validation error with structured errors",
			errorBody:  `{"errors":[{"status":"400","title":"Bad Request","detail":"Invalid ARN pattern","source":{"pointer":"/data/attributes/arn_pattern"}}]}`,
			wantDetail: "Invalid ARN pattern",
		},
		{
			name:         "visibility GET fails",
			responseType: "aws_cloud_auth_config",
			wantGet:      true,
		},
		{
			name:         "POST response contains an unparsed type",
			responseType: "future_aws_cloud_auth_config",
			wantGet:      false,
		},
		{
			name:   "cancellation during POST returns an error",
			cancel: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sawGet := false
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				w.Header().Set("Content-Type", "application/json")
				switch {
				case request.Method == http.MethodPost && request.URL.Path == "/api/v2/cloud_auth/aws/persona_mapping":
					if test.errorBody != "" {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, test.errorBody)
						return
					}
					if test.cancel {
						cancel()
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, `{"errors":["AWS Account Id is not integrated with this Datadog account"]}`)
						return
					}
					w.WriteHeader(http.StatusCreated)
					fmt.Fprintf(w, `{"data":{"id":%q,"type":%q,"attributes":{"account_identifier":%q,"account_uuid":%q,"arn_pattern":%q}}}`,
						mappingID, test.responseType, accountIdentifier, accountUUID, arnPattern)
				case request.Method == http.MethodGet && request.URL.Path == "/api/v2/cloud_auth/aws/persona_mapping/"+mappingID:
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
			config.SetUnstableOperationEnabled("v2.CreateAWSCloudAuthPersonaMapping", true)
			config.SetUnstableOperationEnabled("v2.GetAWSCloudAuthPersonaMapping", true)
			r := &awsWifIdentityMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}

			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			planned := awsWifIdentityMappingModel{
				ID:                types.StringUnknown(),
				AccountIdentifier: types.StringValue(accountIdentifier),
				AccountUUID:       types.StringUnknown(),
				ArnPattern:        types.StringValue(arnPattern),
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
				if got := diags[0].Summary(); got != "Error creating AWS WIF identity mapping" {
					t.Fatalf("diagnostic summary = %q", got)
				}
				for _, detail := range []string{test.wantDetail, "400 Bad Request", test.errorBody} {
					if !strings.Contains(diags[0].Detail(), detail) {
						t.Errorf("diagnostic detail = %q, want to contain %q", diags[0].Detail(), detail)
					}
				}
				if got := requests.Load(); got != 1 {
					t.Errorf("got %d requests, want 1 (validation failures must not be retried)", got)
				}
				if !response.State.Raw.Equal(plan.Raw) {
					t.Error("failed create changed state")
				}
				return
			}

			if test.cancel {
				return
			}

			var state awsWifIdentityMappingModel
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
			if got := state.ArnPattern.ValueString(); got != arnPattern {
				t.Fatalf("arn_pattern = %q, want %q", got, arnPattern)
			}
		})
	}
}

func TestAwsWifIdentityMappingUpdateState(t *testing.T) {
	apiResponse := newAwsWifIdentityMappingResponse(
		"mapping-id",
		"service-account-handle",
		"account-uuid",
		"arn:aws:sts::123456789012:assumed-role/terraform-runner/session-name",
	)
	resource := &awsWifIdentityMappingResource{}

	t.Run("preserves configured email when API returns handle", func(t *testing.T) {
		state := awsWifIdentityMappingModel{
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
		state := awsWifIdentityMappingModel{AccountIdentifier: types.StringNull()}

		resource.updateState(&state, apiResponse)

		if got := state.AccountIdentifier.ValueString(); got != "service-account-handle" {
			t.Fatalf("account_identifier = %q, want API handle", got)
		}
	})
}

func newAwsWifIdentityMappingResponse(id, accountIdentifier, accountUUID, arnPattern string) *datadogV2.AWSCloudAuthPersonaMappingResponse {
	attributes := datadogV2.NewAWSCloudAuthPersonaMappingAttributesResponse(accountIdentifier, accountUUID, arnPattern)
	data := datadogV2.NewAWSCloudAuthPersonaMappingDataResponse(
		*attributes,
		id,
		datadogV2.AWSCLOUDAUTHPERSONAMAPPINGTYPE_AWS_CLOUD_AUTH_CONFIG,
	)
	return datadogV2.NewAWSCloudAuthPersonaMappingResponse(*data)
}

func TestAwsWifIdentityMappingRead(t *testing.T) {
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
				if request.Method != http.MethodGet || request.URL.Path != "/api/v2/cloud_auth/aws/persona_mapping/mapping-id" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				index := min(int(requests.Add(1))-1, len(tc.statuses)-1)
				status := tc.statuses[index]
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status == http.StatusOK {
					if err := json.NewEncoder(w).Encode(newAwsWifIdentityMappingResponse("mapping-id", "handle", "account-uuid", "arn:aws:sts::123456789012:assumed-role/terraform-runner/*")); err != nil {
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
			config.SetUnstableOperationEnabled("v2.GetAWSCloudAuthPersonaMapping", true)
			r := &awsWifIdentityMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			prior := awsWifIdentityMappingModel{
				ID:                types.StringValue("mapping-id"),
				AccountIdentifier: types.StringValue("configured@example.com"),
				AccountUUID:       types.StringValue("account-uuid"),
				ArnPattern:        types.StringValue("arn:aws:sts::123456789012:assumed-role/terraform-runner/*"),
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
