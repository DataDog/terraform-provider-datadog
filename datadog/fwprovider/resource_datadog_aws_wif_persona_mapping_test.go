package fwprovider

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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
			value: "arn:aws:sts::123456789012:assumed-role/team.blue_prod@terraform/session-name:ci",
			valid: true,
		},
		{
			name:  "AWS name characters not supported by the API",
			value: "arn:aws:sts::123456789012:assumed-role/team+blue=prod,ops@terraform/session+name=ci,1",
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

func TestAwsWifPersonaMappingUpdateState(t *testing.T) {
	apiResponse := newAwsWifPersonaMappingResponse(
		"mapping-id",
		"service-account-handle",
		"account-uuid",
		"arn:aws:iam::123456789012:role/terraform-runner",
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
