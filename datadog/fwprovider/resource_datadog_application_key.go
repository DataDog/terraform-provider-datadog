package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &applicationKeyResource{}
	_ resource.ResourceWithImportState = &applicationKeyResource{}
)

func NewApplicationKeyResource() resource.Resource {
	return &applicationKeyResource{}
}

type applicationKeyResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Key    types.String `tfsdk:"key"`
	Scopes types.Set    `tfsdk:"scopes"`
}

type applicationKeyResource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func (r *applicationKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.AddWarning(
		"Deprecated",
		"The import functionality for datadog_application_key resources is deprecated and will be removed in a future release with prior notice. Securely store your application keys using a secret management system or use the datadog_application_key resource to create and manage new application keys.",
	)
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *applicationKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Application Key resource. This can be used to create and manage Datadog Application Keys. Import functionality for this resource is deprecated and will be removed in a future release with prior notice. Securely store your application keys using a secret management system or use this resource to create and manage new application keys.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for Application Key.",
				Required:    true,
			},
			"key": schema.StringAttribute{
				Description: "The value of the Application Key.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"scopes": schema.SetAttribute{
				Description: "Authorization scopes for the Application Key. Application Keys configured with no scopes have full access.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *applicationKeyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *applicationKeyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "application_key"
}

func (r *applicationKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state applicationKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateCurrentUserApplicationKey(r.Auth, *r.buildDatadogApplicationKeyCreateV2Struct(&state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating application key"))
		return
	}
	applicationKeyData := resp.GetData()
	state.ID = types.StringValue(applicationKeyData.GetId())
	r.updateState(ctx, &state, &applicationKeyData)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)

}

func (r *applicationKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state applicationKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetCurrentUserApplicationKey(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Application Key"))
		return
	}

	applicationKeyData := resp.GetData()
	r.updateState(ctx, &state, &applicationKeyData)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *applicationKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state applicationKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateCurrentUserApplicationKey(r.Auth, state.ID.ValueString(), *r.buildDatadogApplicationKeyUpdateV2Struct(&state))

	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating application key"))
		return
	}

	applicationKeyData := resp.GetData()
	state.ID = types.StringValue(applicationKeyData.GetId())
	r.updateState(ctx, &state, &applicationKeyData)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *applicationKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state applicationKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := r.Api.DeleteCurrentUserApplicationKey(r.Auth, state.ID.ValueString()); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting application key"))
	}
}

func (r *applicationKeyResource) buildDatadogApplicationKeyCreateV2Struct(state *applicationKeyResourceModel) *datadogV2.ApplicationKeyCreateRequest {
	applicationKeyAttributes := datadogV2.NewApplicationKeyCreateAttributes(state.Name.ValueString())
	applicationKeyAttributes.SetScopes(getScopesFromStateAttribute(state.Scopes))
	applicationKeyData := datadogV2.NewApplicationKeyCreateData(*applicationKeyAttributes, datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS)
	applicationKeyRequest := datadogV2.NewApplicationKeyCreateRequest(*applicationKeyData)

	return applicationKeyRequest
}

func (r *applicationKeyResource) buildDatadogApplicationKeyUpdateV2Struct(state *applicationKeyResourceModel) *datadogV2.ApplicationKeyUpdateRequest {
	applicationKeyAttributes := datadogV2.NewApplicationKeyUpdateAttributes()
	applicationKeyAttributes.SetName(state.Name.ValueString())
	applicationKeyAttributes.SetScopes(getScopesFromStateAttribute(state.Scopes))
	applicationKeyData := datadogV2.NewApplicationKeyUpdateData(*applicationKeyAttributes, state.ID.ValueString(), datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS)
	applicationKeyRequest := datadogV2.NewApplicationKeyUpdateRequest(*applicationKeyData)

	return applicationKeyRequest
}

func (r *applicationKeyResource) updateState(ctx context.Context, state *applicationKeyResourceModel, applicationKeyData *datadogV2.FullApplicationKey) {
	applicationKeyAttributes := applicationKeyData.GetAttributes()
	state.Name = types.StringValue(applicationKeyAttributes.GetName())
	if applicationKeyAttributes.HasKey() {
		state.Key = types.StringValue(applicationKeyAttributes.GetKey())
	}
	if applicationKeyAttributes.HasScopes() {
		state.Scopes, _ = types.SetValueFrom(ctx, types.StringType, applicationKeyAttributes.GetScopes())
	}
}

func getScopesFromStateAttribute(scopes types.Set) []string {
	scopesList := []string{}

	for _, scope := range scopes.Elements() {
		scopesList = append(scopesList, scope.(types.String).ValueString())
	}
	return scopesList
}
