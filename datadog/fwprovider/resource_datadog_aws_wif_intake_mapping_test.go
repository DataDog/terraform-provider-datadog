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

func TestAwsWifIntakeMappingCreate(t *testing.T) {
	const (
		mappingID  = "7c405332-7033-40d2-a046-a27a075a22cd"
		arnPattern = "arn:aws:sts::123456789012:assumed-role/terraform-runner/*"
	)

	tests := []struct {
		name         string
		errorBody    string
		wantDetail   string
		responseType string
		cancel       bool
		wantGet      bool
		success      bool
		propagation  bool
		missingID    bool
	}{
		{name: "success", responseType: "aws_cloud_auth_intake_mapping", wantGet: true, success: true},
		{name: "integration propagation retries", responseType: "aws_cloud_auth_intake_mapping", wantGet: true, success: true, propagation: true},
		{name: "missing ID rejected", responseType: "aws_cloud_auth_intake_mapping", missingID: true},
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
			responseType: "aws_cloud_auth_intake_mapping",
			wantGet:      true,
		},
		{
			name:         "POST response contains an unparsed type",
			responseType: "future_aws_cloud_auth_intake_mapping",
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
				case request.Method == http.MethodPost && request.URL.Path == "/api/v2/cloud_auth/aws/intake_mapping":
					var body struct {
						Data struct {
							Type       string            `json:"type"`
							Attributes map[string]string `json:"attributes"`
						} `json:"data"`
					}
					if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body.Data.Type != "aws_cloud_auth_intake_mapping" || len(body.Data.Attributes) != 1 || body.Data.Attributes["arn_pattern"] != arnPattern {
						t.Errorf("unexpected create payload: %+v", body)
					}
					if test.propagation && requests.Load() == 1 {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, `{"errors":["AWS Account Id is not integrated with this Datadog account"]}`)
						return
					}

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
					responseID := mappingID
					if test.missingID {
						responseID = ""
					}
					w.WriteHeader(http.StatusCreated)
					fmt.Fprintf(w, `{"data":{"id":%q,"type":%q,"attributes":{"arn_pattern":%q}}}`,
						responseID, test.responseType, arnPattern)
				case request.Method == http.MethodGet && request.URL.Path == "/api/v2/cloud_auth/aws/intake_mapping/"+mappingID:
					sawGet = true
					if test.success {
						fmt.Fprintf(w, `{"data":{"id":%q,"type":"aws_cloud_auth_intake_mapping","attributes":{"arn_pattern":%q}}}`, mappingID, arnPattern)
						return
					}
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
			config.SetUnstableOperationEnabled("v2.CreateAWSCloudAuthIntakeMapping", true)
			config.SetUnstableOperationEnabled("v2.GetAWSCloudAuthIntakeMapping", true)
			r := &awsWifIntakeMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}

			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			planned := awsWifIntakeMappingModel{
				ID:         types.StringUnknown(),
				ArnPattern: types.StringValue(arnPattern),
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			if diags := plan.Set(ctx, &planned); diags.HasError() {
				t.Fatalf("building plan: %v", diags.Errors())
			}

			response := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
			if response.Diagnostics.HasError() == test.success {
				t.Fatalf("unexpected create diagnostics: %v", response.Diagnostics)
			}
			if sawGet != test.wantGet {
				t.Fatalf("visibility GET called = %t, want %t", sawGet, test.wantGet)
			}

			if test.errorBody != "" {
				diags := response.Diagnostics.Errors()
				if len(diags) != 1 {
					t.Fatalf("got %d error diagnostics, want 1: %v", len(diags), diags)
				}
				if got := diags[0].Summary(); got != "Error creating AWS WIF intake mapping" {
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

			if test.cancel || test.missingID {
				return
			}

			if test.propagation && requests.Load() != 3 {
				t.Fatalf("got %d requests, want retry and visibility read", requests.Load())
			}
			var state awsWifIntakeMappingModel
			if diags := response.State.Get(ctx, &state); diags.HasError() {
				t.Fatalf("reading response state: %v", diags.Errors())
			}
			if got := state.ID.ValueString(); got != mappingID {
				t.Fatalf("id = %q, want %q", got, mappingID)
			}
			if got := state.ArnPattern.ValueString(); got != arnPattern {
				t.Fatalf("arn_pattern = %q, want %q", got, arnPattern)
			}
		})
	}
}

func TestAwsWifIntakeMappingRead(t *testing.T) {
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
				if request.Method != http.MethodGet || request.URL.Path != "/api/v2/cloud_auth/aws/intake_mapping/mapping-id" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				index := min(int(requests.Add(1))-1, len(tc.statuses)-1)
				status := tc.statuses[index]
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status == http.StatusOK {
					fmt.Fprint(w, `{"data":{"id":"mapping-id","type":"aws_cloud_auth_intake_mapping","attributes":{"arn_pattern":"arn:aws:sts::123456789012:assumed-role/terraform-runner/*"}}}`)
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
			config.SetUnstableOperationEnabled("v2.GetAWSCloudAuthIntakeMapping", true)
			r := &awsWifIntakeMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: context.Background(),
			}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			prior := awsWifIntakeMappingModel{
				ID:         types.StringValue("mapping-id"),
				ArnPattern: types.StringValue("arn:aws:sts::123456789012:assumed-role/terraform-runner/*"),
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

func TestAwsWifIntakeMappingDelete(t *testing.T) {
	for _, status := range []int{http.StatusNoContent, http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != http.MethodDelete || req.URL.Path != "/api/v2/cloud_auth/aws/intake_mapping/mapping-id" {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status != http.StatusNoContent {
					fmt.Fprint(w, `{"errors":["delete failed"]}`)
				}
			}))
			defer server.Close()
			config := datadog.NewConfiguration()
			config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
			config.OperationServers = nil
			config.HTTPClient = server.Client()
			config.SetUnstableOperationEnabled("v2.DeleteAWSCloudAuthIntakeMapping", true)
			r := &awsWifIntakeMappingResource{Api: datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)), Auth: context.Background()}
			var schemaResponse resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
			state := tfsdk.State{Schema: schemaResponse.Schema}
			prior := awsWifIntakeMappingModel{ID: types.StringValue("mapping-id"), ArnPattern: types.StringValue("arn:aws:sts::123456789012:assumed-role/DatadogAgentRole/*")}
			if diags := state.Set(context.Background(), &prior); diags.HasError() {
				t.Fatal(diags)
			}
			response := resource.DeleteResponse{State: state}
			r.Delete(context.Background(), resource.DeleteRequest{State: state}, &response)
			wantError := status != http.StatusNoContent && status != http.StatusNotFound
			if response.Diagnostics.HasError() != wantError {
				t.Fatalf("diagnostics = %v, want error %t", response.Diagnostics, wantError)
			}
		})
	}
}
