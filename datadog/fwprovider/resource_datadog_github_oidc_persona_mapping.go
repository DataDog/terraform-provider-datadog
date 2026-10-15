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

const githubOidcMappingVisibilityTimeout = 30 * time.Second

var (
	_ resource.ResourceWithConfigure   = &githubOidcPersonaMappingResource{}
	_ resource.ResourceWithImportState = &githubOidcPersonaMappingResource{}
)

type githubOidcPersonaMappingResource struct {
	Api  *datadogV2.CloudAuthenticationApi
	Auth context.Context
}

// A passthrough import writes a state with only `id`, leaving claim_matchers
// null; the pointer field lets that null unmarshal into the model before
// Read fills it in.
type githubOidcPersonaMappingModel struct {
	ID                types.String                  `tfsdk:"id"`
	AccountIdentifier types.String                  `tfsdk:"account_identifier"`
	AccountUUID       types.String                  `tfsdk:"account_uuid"`
	ClaimMatchers     *githubOIDCClaimMatchersModel `tfsdk:"claim_matchers"`
}

func NewGithubOidcPersonaMappingResource() resource.Resource {
	return &githubOidcPersonaMappingResource{}
}

func (r *githubOidcPersonaMappingResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "github_oidc_persona_mapping"
}

func (r *githubOidcPersonaMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a GitHub Actions OIDC persona mapping for Datadog workload identity federation. The mapping allows a GitHub Actions workload whose OIDC token matches the `claim_matchers` patterns to authenticate as the Datadog user or service account identified by `account_identifier`. The identity creating the mapping must have every permission assigned to the target identity. Creating the initial mapping requires API and application credentials with the Workload Identity Federation write permission; a provider already using federation cannot bootstrap its own mapping. Mappings cannot be updated; changing any attribute replaces the mapping. This resource uses a public beta API and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"account_identifier": schema.StringAttribute{
				Description: "The email or handle of the Datadog user or service account that the GitHub workload authenticates as. For a Terraform-managed service account, prefer the stable UUID exported by `datadog_service_account.id`; Datadog accepts it as the service account identifier. Datadog normalizes an email to the account's handle, so the handle form is the only value that survives `terraform import` unchanged — importing a mapping configured by email produces a diff on this attribute, which forces replacement.",
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
			"claim_matchers": githubOIDCClaimMatchersAttribute(),
		},
	}
}

func (r *githubOidcPersonaMappingResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *githubOidcPersonaMappingResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *githubOidcPersonaMappingResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state githubOidcPersonaMappingModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.NewGitHubCloudAuthPersonaMappingCreateAttributes(
		state.AccountIdentifier.ValueString(),
		state.ClaimMatchers.toSDK(),
	)
	data := datadogV2.NewGitHubCloudAuthPersonaMappingCreateData(
		*attributes,
		datadogV2.GITHUBCLOUDAUTHPERSONAMAPPINGTYPE_GITHUB_OIDC_AUTH_CONFIG,
	)
	body := datadogV2.NewGitHubCloudAuthPersonaMappingCreateRequest(*data)

	apiResponse, httpResponse, err := r.Api.CreateGitHubCloudAuthPersonaMapping(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error creating GitHub OIDC persona mapping"),
			"Error creating GitHub OIDC persona mapping",
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
		createdAttributes := createdData.GetAttributes()
		if accountUUID := createdAttributes.GetAccountUuid(); accountUUID != "" {
			state.AccountUUID = types.StringValue(accountUUID)
		}
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
		response.Diagnostics.AddError("response contains no mapping ID", "The API created a GitHub OIDC persona mapping but returned an empty ID, so Terraform cannot track it.")
		return
	}

	apiResponse, httpResponse, err = r.readWithRetry(ctx, mappingID)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error waiting for GitHub OIDC persona mapping to become visible"), "",
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

func (r *githubOidcPersonaMappingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state githubOidcPersonaMappingModel
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
			utils.TranslateClientError(err, httpResponse, "error retrieving GitHub OIDC persona mapping"), "",
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
func (r *githubOidcPersonaMappingResource) readWithRetry(ctx context.Context, id string) (datadogV2.GitHubCloudAuthPersonaMappingResponse, *http.Response, error) {
	var result datadogV2.GitHubCloudAuthPersonaMappingResponse
	var httpResponse *http.Response
	// RetryContext can return on cancellation while its callback is still running.
	var resultMu sync.Mutex
	err := retry.RetryContext(ctx, githubOidcMappingVisibilityTimeout, func() *retry.RetryError {
		readResponse, readHTTPResponse, err := r.Api.GetGitHubCloudAuthPersonaMapping(r.Auth, id)
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

func (r *githubOidcPersonaMappingResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported", "GitHub OIDC persona mappings cannot be updated; changing any configured attribute replaces the mapping.")
}

func (r *githubOidcPersonaMappingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state githubOidcPersonaMappingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResponse, err := r.Api.DeleteGitHubCloudAuthPersonaMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(
			utils.TranslateClientError(err, httpResponse, "error deleting GitHub OIDC persona mapping"), "",
		))
	}
}

func (r *githubOidcPersonaMappingResource) updateState(state *githubOidcPersonaMappingModel, apiResponse *datadogV2.GitHubCloudAuthPersonaMappingResponse) {
	data := apiResponse.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId())
	state.AccountUUID = types.StringValue(attributes.GetAccountUuid())
	claimMatchers := githubOIDCClaimMatchersFromSDK(attributes.GetClaimMatchers())
	state.ClaimMatchers = &claimMatchers

	// The API normalizes an email identifier to the resolved handle. Preserve the
	// configured value to prevent a perpetual replacement; imports start with no value.
	if state.AccountIdentifier.IsNull() || state.AccountIdentifier.IsUnknown() {
		state.AccountIdentifier = types.StringValue(attributes.GetAccountIdentifier())
	}
}
