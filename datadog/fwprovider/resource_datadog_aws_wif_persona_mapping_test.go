package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestAwsWifArnPattern(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "exact IAM user",
			value: "arn:aws:iam::123456789012:user/terraform-runner",
			valid: true,
		},
		{
			name:  "exact STS assumed role session",
			value: "arn:aws:sts::123456789012:assumed-role/terraform-runner/session-name",
			valid: true,
		},
		{
			name:  "assumed role with trailing wildcard",
			value: "arn:aws:sts::123456789012:assumed-role/terraform-runner/*",
			valid: true,
		},
		{
			name:  "special characters supported by the API",
			value: "arn:aws:sts::123456789012:assumed-role/team.blue_prod@terraform/session-name_ci",
			valid: true,
		},
		{
			name:  "AWS name characters not supported by the API",
			value: "arn:aws:sts::123456789012:assumed-role/team+blue=prod,ops@terraform/session+name=ci,1",
			valid: false,
		},
		{
			name:  "colon is not valid in an STS name",
			value: "arn:aws:sts::123456789012:assumed-role/terraform-runner/session:name",
			valid: false,
		},
		{
			name:  "federated user",
			value: "arn:aws:sts::123456789012:federated-user/terraform-runner",
			valid: true,
		},
		{
			name:  "IAM role is not a caller ARN",
			value: "arn:aws:iam::123456789012:role/terraform-runner",
			valid: false,
		},
		{
			name:  "mismatched STS resource type",
			value: "arn:aws:sts::123456789012:group/terraform-runners",
			valid: false,
		},
		{
			name:  "mismatched IAM resource type",
			value: "arn:aws:iam::123456789012:assumed-role/terraform-runner/session-name",
			valid: false,
		},
		{
			name:  "account ID must contain 12 digits",
			value: "arn:aws:sts::12345:assumed-role/terraform-runner/session-name",
			valid: false,
		},
		{
			name:  "wildcard without a specific resource",
			value: "arn:aws:sts::123456789012:assumed-role/*",
			valid: false,
		},
		{
			name:  "wildcard in the middle",
			value: "arn:aws:sts::123456789012:assumed-role/*/session-name",
			valid: false,
		},
		{
			name:  "multiple wildcards",
			value: "arn:aws:sts::123456789012:assumed-role/terraform-runner/*/*",
			valid: false,
		},
		{
			name:  "question mark wildcard",
			value: "arn:aws:iam::123456789012:role/terraform-runner?",
			valid: false,
		},
		{
			name:  "China partition is not supported by the API",
			value: "arn:aws-cn:sts::123456789012:assumed-role/terraform-runner/session-name",
			valid: false,
		},
		{
			name:  "GovCloud partition is not supported by the API",
			value: "arn:aws-us-gov:sts::123456789012:assumed-role/terraform-runner/session-name",
			valid: false,
		},
		{
			name:  "ISO partition is not supported by the API",
			value: "arn:aws-iso:sts::123456789012:assumed-role/terraform-runner/session-name",
			valid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := awsWifArnPattern.MatchString(test.value); got != test.valid {
				t.Fatalf("awsWifArnPattern.MatchString(%q) = %t, want %t", test.value, got, test.valid)
			}
		})
	}
}

func TestAwsWifPersonaMappingCreatePreservesStateOnPostCreateFailure(t *testing.T) {
	const (
		mappingID         = "7c405332-7033-40d2-a046-a27a075a22cd"
		accountIdentifier = "service-account-id"
		accountUUID       = "0bd557d2-6ed0-423e-9654-c0e1cd376abc"
		arnPattern        = "arn:aws:sts::123456789012:assumed-role/terraform-runner/*"
	)

	tests := []struct {
		name         string
		responseType string
		wantGet      bool
	}{
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			sawGet := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case request.Method == http.MethodPost && request.URL.Path == "/api/v2/cloud_auth/aws/persona_mapping":
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
			r := &awsWifPersonaMappingResource{
				Api:  datadogV2.NewCloudAuthenticationApi(datadog.NewAPIClient(config)),
				Auth: ctx,
			}

			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			planned := awsWifPersonaMappingModel{
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

			var state awsWifPersonaMappingModel
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

func TestAwsWifPersonaMappingUpdateState(t *testing.T) {
	apiResponse := newAwsWifPersonaMappingResponse(
		"mapping-id",
		"service-account-handle",
		"account-uuid",
		"arn:aws:sts::123456789012:assumed-role/terraform-runner/session-name",
	)
	resource := &awsWifPersonaMappingResource{}

	t.Run("preserves configured email when API returns handle", func(t *testing.T) {
		state := awsWifPersonaMappingModel{
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
		state := awsWifPersonaMappingModel{AccountIdentifier: types.StringNull()}

		resource.updateState(&state, apiResponse)

		if got := state.AccountIdentifier.ValueString(); got != "service-account-handle" {
			t.Fatalf("account_identifier = %q, want API handle", got)
		}
	})
}

func newAwsWifPersonaMappingResponse(id, accountIdentifier, accountUUID, arnPattern string) *datadogV2.AWSCloudAuthPersonaMappingResponse {
	attributes := datadogV2.NewAWSCloudAuthPersonaMappingAttributesResponse(accountIdentifier, accountUUID, arnPattern)
	data := datadogV2.NewAWSCloudAuthPersonaMappingDataResponse(
		*attributes,
		id,
		datadogV2.AWSCLOUDAUTHPERSONAMAPPINGTYPE_AWS_CLOUD_AUTH_CONFIG,
	)
	return datadogV2.NewAWSCloudAuthPersonaMappingResponse(*data)
}
