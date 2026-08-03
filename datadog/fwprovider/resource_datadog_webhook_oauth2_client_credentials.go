package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &webhookOauth2ClientCredentialsResource{}
	_ resource.ResourceWithImportState = &webhookOauth2ClientCredentialsResource{}
)

type webhookOauth2ClientCredentialsResource struct {
	Api  *datadogV2.WebhooksIntegrationApi
	Auth context.Context
}

type webhookOauth2ClientCredentialsModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AccessTokenURL types.String `tfsdk:"access_token_url"`
	ClientID       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	Audience       types.String `tfsdk:"audience"`
	Scope          types.String `tfsdk:"scope"`
}

func NewWebhookOauth2ClientCredentialsResource() resource.Resource {
	return &webhookOauth2ClientCredentialsResource{}
}

func (r *webhookOauth2ClientCredentialsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetWebhooksIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *webhookOauth2ClientCredentialsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "webhook_oauth2_client_credentials"
}

func (r *webhookOauth2ClientCredentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog webhook OAuth2 client credentials auth method resource. This can be used to create and manage the auth methods available under Integrations -> Webhooks -> Auth Methods.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the auth method.",
				Required:    true,
			},
			"access_token_url": schema.StringAttribute{
				Description: "The URL used to fetch the access token.",
				Required:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "The OAuth2 client ID.",
				Required:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The OAuth2 client secret. This value is not returned by the API, so it cannot be detected as drifted or filled in on import.",
				Required:    true,
				Sensitive:   true,
			},
			"audience": schema.StringAttribute{
				Description: "The audience requested when fetching the access token.",
				Optional:    true,
			},
			"scope": schema.StringAttribute{
				Description: "The scope requested when fetching the access token.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
		},
	}
}

func (r *webhookOauth2ClientCredentialsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *webhookOauth2ClientCredentialsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state webhookOauth2ClientCredentialsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetOAuth2ClientCredentials(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting webhook OAuth2 client credentials"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *webhookOauth2ClientCredentialsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state webhookOauth2ClientCredentialsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateOAuth2ClientCredentials(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating webhook OAuth2 client credentials"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *webhookOauth2ClientCredentialsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state webhookOauth2ClientCredentialsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var prevState webhookOauth2ClientCredentialsModel
	response.Diagnostics.Append(request.State.Get(ctx, &prevState)...)
	if response.Diagnostics.HasError() {
		return
	}
	state.ID = prevState.ID

	body, diags := r.buildUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateOAuth2ClientCredentials(r.Auth, state.ID.ValueString(), *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating webhook OAuth2 client credentials"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *webhookOauth2ClientCredentialsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state webhookOauth2ClientCredentialsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteOAuth2ClientCredentials(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting webhook OAuth2 client credentials"))
		return
	}
}

// updateState copies the API response into state. client_secret is never
// returned by the API, so the value already in state/plan is kept as is.
func (r *webhookOauth2ClientCredentialsResource) updateState(ctx context.Context, state *webhookOauth2ClientCredentialsModel, resp *datadogV2.WebhooksOAuth2ClientCredentialsResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}
	if accessTokenURL, ok := attributes.GetAccessTokenUrlOk(); ok {
		state.AccessTokenURL = types.StringValue(*accessTokenURL)
	}
	if clientID, ok := attributes.GetClientIdOk(); ok {
		state.ClientID = types.StringValue(*clientID)
	}
	if audience, ok := attributes.GetAudienceOk(); ok && audience != nil {
		state.Audience = types.StringValue(*audience)
	} else {
		state.Audience = types.StringNull()
	}
	if scope, ok := attributes.GetScopeOk(); ok && scope != nil {
		state.Scope = types.StringValue(*scope)
	} else {
		state.Scope = types.StringNull()
	}
}

func (r *webhookOauth2ClientCredentialsResource) buildCreateRequestBody(ctx context.Context, state *webhookOauth2ClientCredentialsModel) (*datadogV2.WebhooksOAuth2ClientCredentialsCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewWebhooksOAuth2ClientCredentialsCreateAttributes(
		state.AccessTokenURL.ValueString(),
		state.ClientID.ValueString(),
		state.ClientSecret.ValueString(),
		state.Name.ValueString(),
	)
	if !state.Audience.IsNull() {
		attributes.SetAudience(state.Audience.ValueString())
	}
	if !state.Scope.IsNull() {
		attributes.SetScope(state.Scope.ValueString())
	}

	data := datadogV2.NewWebhooksOAuth2ClientCredentialsCreateData(
		*attributes,
		datadogV2.WEBHOOKSOAUTH2CLIENTCREDENTIALSTYPE_WEBHOOKS_AUTH_METHOD_OAUTH2_CLIENT_CREDENTIALS,
	)

	return datadogV2.NewWebhooksOAuth2ClientCredentialsCreateRequest(*data), diags
}

func (r *webhookOauth2ClientCredentialsResource) buildUpdateRequestBody(ctx context.Context, state *webhookOauth2ClientCredentialsModel) (*datadogV2.WebhooksOAuth2ClientCredentialsUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewWebhooksOAuth2ClientCredentialsUpdateAttributes()
	attributes.SetName(state.Name.ValueString())
	attributes.SetAccessTokenUrl(state.AccessTokenURL.ValueString())
	attributes.SetClientId(state.ClientID.ValueString())
	attributes.SetClientSecret(state.ClientSecret.ValueString())
	// audience and scope are nullable: send an explicit null to clear them
	// when they are removed from the configuration.
	if !state.Audience.IsNull() {
		attributes.SetAudience(state.Audience.ValueString())
	} else {
		attributes.SetAudienceNil()
	}
	if !state.Scope.IsNull() {
		attributes.SetScope(state.Scope.ValueString())
	} else {
		attributes.SetScopeNil()
	}

	data := datadogV2.NewWebhooksOAuth2ClientCredentialsUpdateData(
		*attributes,
		datadogV2.WEBHOOKSOAUTH2CLIENTCREDENTIALSTYPE_WEBHOOKS_AUTH_METHOD_OAUTH2_CLIENT_CREDENTIALS,
	)

	return datadogV2.NewWebhooksOAuth2ClientCredentialsUpdateRequest(*data), diags
}
