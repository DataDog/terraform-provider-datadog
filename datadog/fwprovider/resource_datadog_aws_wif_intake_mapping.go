package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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
	awsWifIntakeMappingCreateTimeout     = 2 * time.Minute
	awsWifIntakeMappingVisibilityTimeout = 30 * time.Second
)

var (
	_ resource.ResourceWithConfigure   = &awsWifIntakeMappingResource{}
	_ resource.ResourceWithImportState = &awsWifIntakeMappingResource{}
)

type awsWifIntakeMappingResource struct {
	Api  *datadogV2.CloudAuthenticationApi
	Auth context.Context
}

type awsWifIntakeMappingModel struct {
	ID         types.String `tfsdk:"id"`
	ArnPattern types.String `tfsdk:"arn_pattern"`
}

func NewAwsWifIntakeMappingResource() resource.Resource {
	return &awsWifIntakeMappingResource{}
}

func (r *awsWifIntakeMappingResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_wif_intake_mapping"
}

func (r *awsWifIntakeMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides an AWS Workload Identity Federation (WIF) intake mapping. The mapping allows Datadog Agents with an AWS identity matching `arn_pattern` to authenticate for telemetry submission using a Datadog-managed, automatically rotated API key. The AWS account in the ARN must already be integrated with Datadog. Creating the mapping requires the Workload Identity Federation write permission. Configure the Agent separately to use WIF with the target organization UUID. The API rejects ARN patterns that conflict with existing intake or identity mappings. Mapping changes may take several minutes to affect authentication. This resource uses a public beta API and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"arn_pattern": schema.StringAttribute{
				Description: "The AWS caller ARN pattern allowed to authenticate. Currently, only the `aws` partition is supported. For role-based authentication, use the STS assumed-role ARN returned by `aws sts get-caller-identity`, not the IAM role ARN shown in the AWS console. A pattern may contain one wildcard only, as a trailing `/*` after a specific resource, for example `arn:aws:sts::123456789012:assumed-role/DatadogAgentRole/*`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (r *awsWifIntakeMappingResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *awsWifIntakeMappingResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *awsWifIntakeMappingResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state awsWifIntakeMappingModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.NewAWSCloudAuthIntakeMappingCreateAttributes(
		state.ArnPattern.ValueString(),
	)
	data := datadogV2.NewAWSCloudAuthIntakeMappingCreateData(
		datadogV2.AWSCLOUDAUTHINTAKEMAPPINGTYPE_AWS_CLOUD_AUTH_INTAKE_MAPPING,
		*attributes,
	)
	body := datadogV2.NewAWSCloudAuthIntakeMappingCreateRequest(*data)

	var apiResponse datadogV2.AWSCloudAuthIntakeMappingResponse
	var httpResponse *http.Response
	err := retry.RetryContext(ctx, awsWifIntakeMappingCreateTimeout, func() *retry.RetryError {
		// Keep the request error local: cancellation can return before this callback.
		var createErr error
		apiResponse, httpResponse, createErr = r.Api.CreateAWSCloudAuthIntakeMapping(r.Auth, *body)
		if createErr == nil {
			return nil
		}

		translatedError := utils.TranslateClientError(createErr, httpResponse, "error creating AWS WIF intake mapping")
		if isAwsIntegrationPropagationError(createErr, httpResponse) {
			return retry.RetryableError(translatedError)
		}
		return retry.NonRetryableError(translatedError)
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error creating AWS WIF intake mapping"))
		return
	}
	unparsedErr := utils.CheckForUnparsed(apiResponse)
	createdData := apiResponse.GetData()
	mappingID := createdData.GetId()
	if unparsedErr == nil {
		r.updateState(&state, &apiResponse)
	} else if mappingID != "" {
		// Preserve enough state to track and destroy a mapping when the SDK can
		// still expose its ID from an otherwise unparsed public-beta response.
		state.ID = types.StringValue(mappingID)
	}
	if mappingID != "" {
		response.Diagnostics.Append(response.State.Set(ctx, &state)...)
	}
	if response.Diagnostics.HasError() {
		return
	}
	if unparsedErr != nil {
		response.Diagnostics.AddError("response contains unparsed object", unparsedErr.Error())
		return
	}
	if mappingID == "" {
		response.Diagnostics.AddError("response contains no mapping ID", "The API created an AWS WIF intake mapping but returned an empty ID, so Terraform cannot track it.")
		return
	}

	apiResponse, httpResponse, err = r.readWithRetry(ctx, mappingID)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error waiting for AWS WIF intake mapping to become visible"), "",
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

func (r *awsWifIntakeMappingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state awsWifIntakeMappingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiResponse, httpResponse, err := r.readWithRetry(ctx, state.ID.ValueString())
	if err != nil {
		if ctx.Err() == nil && httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error retrieving AWS WIF intake mapping"), "",
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

// A newly created mapping can intermittently return 404 even after a successful
// GET. Retry before treating it as deleted so a refresh cannot orphan it.
func (r *awsWifIntakeMappingResource) readWithRetry(ctx context.Context, id string) (datadogV2.AWSCloudAuthIntakeMappingResponse, *http.Response, error) {
	var result datadogV2.AWSCloudAuthIntakeMappingResponse
	var httpResponse *http.Response
	// RetryContext can return on cancellation while its callback is still running.
	var resultMu sync.Mutex
	err := retry.RetryContext(ctx, awsWifIntakeMappingVisibilityTimeout, func() *retry.RetryError {
		readResponse, readHTTPResponse, err := r.Api.GetAWSCloudAuthIntakeMapping(r.Auth, id)
		resultMu.Lock()
		result, httpResponse = readResponse, readHTTPResponse
		resultMu.Unlock()
		if err == nil {
			return nil
		}
		if readHTTPResponse != nil && readHTTPResponse.StatusCode == http.StatusNotFound {
			return retry.RetryableError(err)
		}
		return retry.NonRetryableError(err)
	})
	resultMu.Lock()
	defer resultMu.Unlock()
	return result, httpResponse, err
}

func (r *awsWifIntakeMappingResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "AWS WIF intake mappings cannot be updated; changing the ARN pattern replaces the mapping.")
}

func (r *awsWifIntakeMappingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state awsWifIntakeMappingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResponse, err := r.Api.DeleteAWSCloudAuthIntakeMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error deleting AWS WIF intake mapping"), "",
		))
	}
}

func (r *awsWifIntakeMappingResource) updateState(state *awsWifIntakeMappingModel, apiResponse *datadogV2.AWSCloudAuthIntakeMappingResponse) {
	data := apiResponse.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.ArnPattern = types.StringValue(attributes.GetArnPattern())
}
