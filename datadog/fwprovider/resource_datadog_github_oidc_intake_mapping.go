package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &githubOidcIntakeMappingResource{}
	_ resource.ResourceWithImportState = &githubOidcIntakeMappingResource{}
)

type githubOidcIntakeMappingResource struct {
	Api  *datadogV2.CloudAuthenticationApi
	Auth context.Context
}

// A passthrough import writes a state with only `id`, leaving claim_matchers
// null; the pointer field lets that null unmarshal into the model before
// Read fills it in.
type githubOidcIntakeMappingModel struct {
	ID            types.String                  `tfsdk:"id"`
	ClaimMatchers *githubOIDCClaimMatchersModel `tfsdk:"claim_matchers"`
}

func NewGithubOidcIntakeMappingResource() resource.Resource {
	return &githubOidcIntakeMappingResource{}
}

func (r *githubOidcIntakeMappingResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "github_oidc_intake_mapping"
}

func (r *githubOidcIntakeMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a GitHub Actions OIDC intake mapping for Datadog workload identity federation. The mapping allows GitHub Actions workloads whose OIDC tokens match the `claim_matchers` patterns to send telemetry to Datadog intake endpoints. Mappings cannot be updated; changing any attribute replaces the mapping. This resource uses a public beta API and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id":             utils.ResourceIDAttribute(),
			"claim_matchers": githubOIDCClaimMatchersAttribute(),
		},
	}
}

func (r *githubOidcIntakeMappingResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *githubOidcIntakeMappingResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *githubOidcIntakeMappingResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state githubOidcIntakeMappingModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.NewGitHubCloudAuthIntakeMappingCreateAttributes(
		state.ClaimMatchers.toSDK(),
	)
	data := datadogV2.NewGitHubCloudAuthIntakeMappingCreateData(
		*attributes,
		datadogV2.GITHUBCLOUDAUTHINTAKEMAPPINGTYPE_GITHUB_OIDC_AUTH_INTAKE_MAPPING,
	)
	body := datadogV2.NewGitHubCloudAuthIntakeMappingCreateRequest(*data)

	apiResponse, httpResponse, err := r.Api.CreateGitHubCloudAuthIntakeMapping(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error creating GitHub OIDC intake mapping"),
			"Error creating GitHub OIDC intake mapping",
		))
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
		response.Diagnostics.AddError("response contains no mapping ID", "The API created a GitHub OIDC intake mapping but returned an empty ID, so Terraform cannot track it.")
		return
	}

	apiResponse, httpResponse, err = r.readWithRetry(ctx, mappingID)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error waiting for GitHub OIDC intake mapping to become visible"), "",
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

func (r *githubOidcIntakeMappingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state githubOidcIntakeMappingModel
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
			utils.TranslateClientError(err, httpResponse, "error retrieving GitHub OIDC intake mapping"), "",
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
func (r *githubOidcIntakeMappingResource) readWithRetry(ctx context.Context, id string) (datadogV2.GitHubCloudAuthIntakeMappingResponse, *http.Response, error) {
	var result datadogV2.GitHubCloudAuthIntakeMappingResponse
	var httpResponse *http.Response
	// RetryContext can return on cancellation while its callback is still running.
	var resultMu sync.Mutex
	err := retry.RetryContext(ctx, githubOidcMappingVisibilityTimeout, func() *retry.RetryError {
		readResponse, readHTTPResponse, err := r.Api.GetGitHubCloudAuthIntakeMapping(r.Auth, id)
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

func (r *githubOidcIntakeMappingResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "GitHub OIDC intake mappings cannot be updated; changing any configured attribute replaces the mapping.")
}

func (r *githubOidcIntakeMappingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state githubOidcIntakeMappingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResponse, err := r.Api.DeleteGitHubCloudAuthIntakeMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error deleting GitHub OIDC intake mapping"), "",
		))
	}
}

func (r *githubOidcIntakeMappingResource) updateState(state *githubOidcIntakeMappingModel, apiResponse *datadogV2.GitHubCloudAuthIntakeMappingResponse) {
	data := apiResponse.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	claimMatchers := githubOIDCClaimMatchersFromSDK(attributes.GetClaimMatchers())
	state.ClaimMatchers = &claimMatchers
}
