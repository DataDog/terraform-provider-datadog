package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const (
	awsWifPersonaMappingCreateTimeout     = 2 * time.Minute
	awsWifPersonaMappingVisibilityTimeout = 30 * time.Second
)

var (
	_ resource.ResourceWithConfigure   = &awsWifPersonaMappingResource{}
	_ resource.ResourceWithImportState = &awsWifPersonaMappingResource{}

	// Keep the accepted partition and resource characters aligned with the Cloud
	// Authentication API, while limiting resource types to identities that AWS
	// STS GetCallerIdentity can return.
	awsWifArnPattern = regexp.MustCompile(`^arn:aws:(?:sts::[0-9]{12}:(?:assumed-role/(?:[A-Za-z0-9_.\-:@]+/)+(?:[A-Za-z0-9_.\-:@]+|\*)|federated-user/[A-Za-z0-9_.\-:@]+)|iam::[0-9]{12}:user/(?:[A-Za-z0-9_.\-:@]+/)*[A-Za-z0-9_.\-:@]+(?:/\*)?)$`)
)

type awsWifPersonaMappingResource struct {
	Api  *datadogV2.CloudAuthenticationApi
	Auth context.Context
}

type awsWifPersonaMappingModel struct {
	ID                types.String `tfsdk:"id"`
	AccountIdentifier types.String `tfsdk:"account_identifier"`
	AccountUUID       types.String `tfsdk:"account_uuid"`
	ArnPattern        types.String `tfsdk:"arn_pattern"`
}

func NewAwsWifPersonaMappingResource() resource.Resource {
	return &awsWifPersonaMappingResource{}
}

func (r *awsWifPersonaMappingResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_wif_persona_mapping"
}

func (r *awsWifPersonaMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides an AWS Workload Identity Federation (WIF) persona mapping. The mapping allows an AWS IAM principal matching `arn_pattern` to authenticate as the Datadog user or service account identified by `account_identifier`. The AWS account in the ARN must already be integrated with Datadog. The identity creating the mapping must have every permission assigned to the target identity. Creating the initial mapping requires API and application credentials with the Workload Identity Federation write permission; a provider already using WIF cannot bootstrap its own mapping. This resource uses a public beta API and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"account_identifier": schema.StringAttribute{
				Description: "The email or handle of the Datadog user or service account that the AWS principal authenticates as. Prefer the `id` exported by `datadog_service_account`, which is also the service account handle.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"account_uuid": schema.StringAttribute{
				Description: "The UUID of the Datadog user or service account resolved from `account_identifier`.",
				Computed:    true,
			},
			"arn_pattern": schema.StringAttribute{
				Description: "The AWS caller ARN pattern allowed to authenticate. Currently, only the `aws` partition is supported. For role-based authentication, use the STS assumed-role ARN returned by `aws sts get-caller-identity`, not the IAM role ARN shown in the AWS console. A pattern may contain one wildcard only, as a trailing `/*` after a specific resource, for example `arn:aws:sts::123456789012:assumed-role/terraform-runner/*`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(awsWifArnPattern, "must be an AWS GetCallerIdentity ARN supported by Datadog for an IAM user, assumed role, or federated user; a wildcard is allowed only as one trailing /* after a specific resource"),
				},
			},
		},
	}
}

func (r *awsWifPersonaMappingResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetCloudAuthenticationApiV2()
	r.Auth = providerData.Auth
}

func (r *awsWifPersonaMappingResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *awsWifPersonaMappingResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state awsWifPersonaMappingModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.NewAWSCloudAuthPersonaMappingCreateAttributes(
		state.AccountIdentifier.ValueString(),
		state.ArnPattern.ValueString(),
	)
	data := datadogV2.NewAWSCloudAuthPersonaMappingCreateData(
		*attributes,
		datadogV2.AWSCLOUDAUTHPERSONAMAPPINGTYPE_AWS_CLOUD_AUTH_CONFIG,
	)
	body := datadogV2.NewAWSCloudAuthPersonaMappingCreateRequest(*data)

	var apiResponse datadogV2.AWSCloudAuthPersonaMappingResponse
	var httpResponse *http.Response
	var err error
	err = retry.RetryContext(ctx, awsWifPersonaMappingCreateTimeout, func() *retry.RetryError {
		apiResponse, httpResponse, err = r.Api.CreateAWSCloudAuthPersonaMapping(r.Auth, *body)
		if err == nil {
			return nil
		}

		translatedError := utils.TranslateClientError(err, httpResponse, "error creating AWS WIF persona mapping")
		if isAwsIntegrationPropagationError(err, httpResponse) {
			return retry.RetryableError(translatedError)
		}
		return retry.NonRetryableError(translatedError)
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
		return
	}
	if err := utils.CheckForUnparsed(apiResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	createdData := apiResponse.GetData()
	mappingID := createdData.GetId()
	r.updateState(&state, &apiResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	err = retry.RetryContext(ctx, awsWifPersonaMappingVisibilityTimeout, func() *retry.RetryError {
		var readResponse datadogV2.AWSCloudAuthPersonaMappingResponse
		readResponse, httpResponse, err = r.Api.GetAWSCloudAuthPersonaMapping(r.Auth, mappingID)
		if err == nil {
			if unparsedErr := utils.CheckForUnparsed(readResponse); unparsedErr != nil {
				return retry.NonRetryableError(fmt.Errorf("response contains unparsed object: %w", unparsedErr))
			}
			apiResponse = readResponse
			return nil
		}

		translatedError := utils.TranslateClientError(err, httpResponse, "error waiting for AWS WIF persona mapping to become visible")
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return retry.RetryableError(translatedError)
		}
		return retry.NonRetryableError(translatedError)
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
		return
	}

	r.updateState(&state, &apiResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func isAwsIntegrationPropagationError(err error, httpResponse *http.Response) bool {
	if httpResponse == nil || httpResponse.StatusCode != http.StatusBadRequest {
		return false
	}

	var apiError datadog.GenericOpenAPIError
	if !errors.As(err, &apiError) {
		return false
	}
	return strings.Contains(string(apiError.Body()), "AWS Account Id is not integrated with this Datadog account")
}

func (r *awsWifPersonaMappingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state awsWifPersonaMappingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiResponse, httpResponse, err := r.Api.GetAWSCloudAuthPersonaMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error retrieving AWS WIF persona mapping"), "",
		))
		return
	}
	if err := utils.CheckForUnparsed(apiResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	r.updateState(&state, &apiResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *awsWifPersonaMappingResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "AWS WIF persona mappings cannot be updated; changing either configured attribute replaces the mapping.")
}

func (r *awsWifPersonaMappingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state awsWifPersonaMappingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResponse, err := r.Api.DeleteAWSCloudAuthPersonaMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error deleting AWS WIF persona mapping"), "",
		))
	}
}

func (r *awsWifPersonaMappingResource) updateState(state *awsWifPersonaMappingModel, apiResponse *datadogV2.AWSCloudAuthPersonaMappingResponse) {
	data := apiResponse.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.AccountUUID = types.StringValue(attributes.GetAccountUuid())
	state.ArnPattern = types.StringValue(attributes.GetArnPattern())

	// The API normalizes an email identifier to the resolved handle. Preserve the
	// configured value to prevent a perpetual replacement; imports start with no value.
	if state.AccountIdentifier.IsNull() || state.AccountIdentifier.IsUnknown() {
		state.AccountIdentifier = types.StringValue(attributes.GetAccountIdentifier())
	}
}
