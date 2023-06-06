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
	_ resource.ResourceWithConfigure   = &ServiceAccountApplicationKeyResource{}
	_ resource.ResourceWithImportState = &ServiceAccountApplicationKeyResource{}
)

type ServiceAccountApplicationKeyResource struct {
	Api  *datadogV2.ServiceAccountsApi
	Auth context.Context
}

type ServiceAccountApplicationKeyModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceAccountId types.String `tfsdk:"service_account_id"`
	Name             types.String `tfsdk:"name"`
	Scopes           types.List   `tfsdk:"scopes"`
}

func NewServiceAccountApplicationKeyResource() resource.Resource {
	return &ServiceAccountApplicationKeyResource{}
}

func (r *ServiceAccountApplicationKeyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceAccountsApiV2()
	r.Auth = providerData.Auth
}

func (r *ServiceAccountApplicationKeyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "service_account_application_key"
}

func (r *ServiceAccountApplicationKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog ServiceAccountApplicationKey resource. This can be used to create and manage Datadog service_account_application_key.",
		Attributes: map[string]schema.Attribute{
			"service_account_id": schema.StringAttribute{
				Optional:    true,
				Description: "UPDATE ME",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the application key.",
			},
			"scopes": schema.ListAttribute{
				Optional:    true,
				Description: "Array of scopes to grant the application key. This feature is in private beta, please contact Datadog support to enable scopes for your application keys.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *ServiceAccountApplicationKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *ServiceAccountApplicationKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ServiceAccountApplicationKeyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	serviceAccountId := state.ServiceAccountId.ValueString()

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetServiceAccountApplicationKey(r.Auth, serviceAccountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ServiceAccountApplicationKey"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ServiceAccountApplicationKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state ServiceAccountApplicationKeyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()

	body, diags := r.buildServiceAccountApplicationKeyRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateServiceAccountApplicationKey(r.Auth, serviceAccountId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ServiceAccountApplicationKey"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ServiceAccountApplicationKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state ServiceAccountApplicationKeyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serviceAccountId := state.ServiceAccountId.ValueString()

	id := state.ID.ValueString()

	body, diags := r.buildServiceAccountApplicationKeyUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateServiceAccountApplicationKey(r.Auth, serviceAccountId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ServiceAccountApplicationKey"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ServiceAccountApplicationKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ServiceAccountApplicationKeyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	serviceAccountId := state.ServiceAccountId.ValueString()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteServiceAccountApplicationKey(r.Auth, serviceAccountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting service_account_application_key"))
		return
	}
}

func (r *ServiceAccountApplicationKeyResource) updateState(ctx context.Context, state *ServiceAccountApplicationKeyModel, resp *datadogV2.PartialApplicationKeyResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(*createdAt)
	}

	if last4, ok := attributes.GetLast4Ok(); ok {
		state.Last4 = types.StringValue(*last4)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if scopes, ok := attributes.GetScopesOk(); ok && len(*scopes) > 0 {
		state.Scopes, _ = types.ListValueFrom(ctx, types.StringType, *scopes)
	}
}

func (r *ServiceAccountApplicationKeyResource) buildServiceAccountApplicationKeyRequestBody(ctx context.Context, state *ServiceAccountApplicationKeyModel) (*datadogV2.ApplicationKeyCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationKeyCreateAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())

	if !state.Scopes.IsNull() {
		var scopes []string
		diags.Append(state.Scopes.ElementsAs(ctx, &scopes, false)...)
		attributes.SetScopes(scopes)
	}

	req := datadogV2.NewApplicationKeyCreateRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationKeyCreateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *ServiceAccountApplicationKeyResource) buildServiceAccountApplicationKeyUpdateRequestBody(ctx context.Context, state *ServiceAccountApplicationKeyModel) (*datadogV2.ApplicationKeyUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationKeyUpdateAttributesWithDefaults()

	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}

	if !state.Scopes.IsNull() {
		var scopes []string
		diags.Append(state.Scopes.ElementsAs(ctx, &scopes, false)...)
		attributes.SetScopes(scopes)
	}

	req := datadogV2.NewApplicationKeyUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationKeyUpdateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
