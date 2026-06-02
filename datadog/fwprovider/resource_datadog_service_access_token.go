package fwprovider

import (
	"context"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	frameworkPlanModifiers "github.com/terraform-providers/terraform-provider-datadog/datadog/internal/planmodifiers"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &serviceAccessTokenResource{}
	_ resource.ResourceWithImportState = &serviceAccessTokenResource{}
)

type serviceAccessTokenResource struct {
	Api  *datadogV2.ServiceAccountsApi
	Auth context.Context
}

type serviceAccessTokenModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	Name             types.String `tfsdk:"name"`
	Scopes           types.Set    `tfsdk:"scopes"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	Key              types.String `tfsdk:"key"`
	CreatedAt        types.String `tfsdk:"created_at"`
	LastUsedAt       types.String `tfsdk:"last_used_at"`
	PublicPortion    types.String `tfsdk:"public_portion"`
}

func NewServiceAccessTokenResource() resource.Resource {
	return &serviceAccessTokenResource{}
}

func (r *serviceAccessTokenResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceAccountsApiV2()
	r.Auth = providerData.Auth
}

func (r *serviceAccessTokenResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "service_access_token"
}

func (r *serviceAccessTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog `service_access_token` resource. This can be used to create and manage Datadog service access tokens (SATs). A SAT is an access token scoped to a service account; this resource is intentionally limited to service-account-owned tokens.",
		Attributes: map[string]schema.Attribute{
			"service_account_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the service account that owns this access token.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the service access token. Must be non-empty.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"scopes": schema.SetAttribute{
				Required:    true,
				Description: "Authorization scopes granted to the service access token. At least one scope is required.",
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Description: "Expiration date of the service access token, in RFC3339 format. Omit for a non-expiring token. The Datadog API caps expirations to within 365 days from creation. This attribute is immutable: it cannot be added, changed, or removed after creation. To rotate the expiration, destroy and re-create the resource.",
				PlanModifiers: []planmodifier.String{
					frameworkPlanModifiers.ImmutableString("expires_at"),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The value of the service access token. This value is only available at creation time and cannot be imported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the access token, in RFC3339 format.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date the access token was last used, in RFC3339 format. Empty if the token has never been used.",
			},
			"public_portion": schema.StringAttribute{
				Computed:    true,
				Description: "The public portion of the access token, used for identification.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *serviceAccessTokenResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.SplitN(request.ID, ":", 2)
	if len(parts) != 2 {
		response.Diagnostics.AddError("error retrieving service_account_id or service_access_token id from given ID", "expected format `<service_account_id>:<token_id>`")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("service_account_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("key"), "")...)
	response.Diagnostics.AddWarning("Importing a service access token will not import the key value.", "")
}

func (r *serviceAccessTokenResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state serviceAccessTokenModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetServiceAccountAccessToken(r.Auth, serviceAccountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ServiceAccessToken"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateStatePartialToken(ctx, &state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccessTokenResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state serviceAccessTokenModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()

	body, diags := r.buildServiceAccessTokenCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateServiceAccountAccessToken(r.Auth, serviceAccountId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating ServiceAccessToken"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateStateFullToken(ctx, &state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccessTokenResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state serviceAccessTokenModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()
	id := state.ID.ValueString()

	body, diags := r.buildServiceAccessTokenUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateServiceAccountAccessToken(r.Auth, serviceAccountId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ServiceAccessToken"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateStatePartialToken(ctx, &state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccessTokenResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state serviceAccessTokenModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()
	id := state.ID.ValueString()

	httpResp, err := r.Api.RevokeServiceAccountAccessToken(r.Auth, serviceAccountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error revoking service_access_token"))
		return
	}
}

func (r *serviceAccessTokenResource) updateStatePartialToken(ctx context.Context, state *serviceAccessTokenModel, resp *datadogV2.ServiceAccessTokenResponse) {
	data, ok := resp.GetDataOk()
	if !ok || data == nil {
		return
	}
	state.ID = types.StringValue(data.GetId())

	attributes, ok := data.GetAttributesOk()
	if !ok || attributes == nil {
		return
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}
	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.Format(time.RFC3339))
	}
	if expiresAt, ok := attributes.GetExpiresAtOk(); ok && expiresAt != nil {
		state.ExpiresAt = types.StringValue(expiresAt.Format(time.RFC3339))
	} else {
		state.ExpiresAt = types.StringNull()
	}
	if lastUsedAt, ok := attributes.GetLastUsedAtOk(); ok && lastUsedAt != nil {
		state.LastUsedAt = types.StringValue(lastUsedAt.Format(time.RFC3339))
	} else {
		state.LastUsedAt = types.StringValue("")
	}
	if publicPortion, ok := attributes.GetPublicPortionOk(); ok {
		state.PublicPortion = types.StringValue(*publicPortion)
	}
	if attributes.HasScopes() {
		state.Scopes, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetScopes())
	}
}

func (r *serviceAccessTokenResource) updateStateFullToken(ctx context.Context, state *serviceAccessTokenModel, resp *datadogV2.ServiceAccessTokenCreateResponse) {
	data, ok := resp.GetDataOk()
	if !ok || data == nil {
		return
	}
	state.ID = types.StringValue(data.GetId())

	attributes, ok := data.GetAttributesOk()
	if !ok || attributes == nil {
		return
	}

	if key, ok := attributes.GetKeyOk(); ok {
		state.Key = types.StringValue(*key)
	}
	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}
	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.Format(time.RFC3339))
	}
	if expiresAt, ok := attributes.GetExpiresAtOk(); ok && expiresAt != nil {
		state.ExpiresAt = types.StringValue(expiresAt.Format(time.RFC3339))
	} else {
		state.ExpiresAt = types.StringNull()
	}
	state.LastUsedAt = types.StringValue("")
	if publicPortion, ok := attributes.GetPublicPortionOk(); ok {
		state.PublicPortion = types.StringValue(*publicPortion)
	}
	if attributes.HasScopes() {
		state.Scopes, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetScopes())
	}
}

func (r *serviceAccessTokenResource) buildServiceAccessTokenCreateRequestBody(_ context.Context, state *serviceAccessTokenModel) (*datadogV2.ServiceAccountAccessTokenCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewServiceAccountAccessTokenCreateAttributes(
		state.Name.ValueString(),
		getScopesFromStateAttribute(state.Scopes),
	)

	if !state.ExpiresAt.IsNull() && !state.ExpiresAt.IsUnknown() && state.ExpiresAt.ValueString() != "" {
		expiresAt, err := time.Parse(time.RFC3339, state.ExpiresAt.ValueString())
		if err != nil {
			diags.AddError("error parsing expires_at", err.Error())
			return nil, diags
		}
		attributes.SetExpiresAt(expiresAt)
	}

	data := datadogV2.NewServiceAccountAccessTokenCreateDataWithDefaults()
	data.SetAttributes(*attributes)

	req := datadogV2.NewServiceAccountAccessTokenCreateRequestWithDefaults()
	req.SetData(*data)

	return req, diags
}

func (r *serviceAccessTokenResource) buildServiceAccessTokenUpdateRequestBody(_ context.Context, state *serviceAccessTokenModel) (*datadogV2.ServiceAccountAccessTokenUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewServiceAccountAccessTokenUpdateAttributesWithDefaults()
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}
	attributes.SetScopes(getScopesFromStateAttribute(state.Scopes))

	data := datadogV2.NewServiceAccountAccessTokenUpdateDataWithDefaults()
	data.SetAttributes(*attributes)
	if !state.ID.IsNull() {
		data.SetId(state.ID.ValueString())
	}

	req := datadogV2.NewServiceAccountAccessTokenUpdateRequestWithDefaults()
	req.SetData(*data)

	return req, diags
}
