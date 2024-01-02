package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &rumApplicationResource{}
	_ resource.ResourceWithImportState = &rumApplicationResource{}
)

type rumApplicationResource struct {
	Api  *datadogV2.RUMApi
	Auth context.Context
}

type rumApplicationModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	ClientToken types.String `tfsdk:"client_token"`
}

func NewRumApplicationResource() resource.Resource {
	return &rumApplicationResource{}
}

func (r *rumApplicationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRumApiV2()
	r.Auth = providerData.Auth
}

func (r *rumApplicationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_application"
}

func (r *rumApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RUM application resource. This can be used to create and manage Datadog RUM applications.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the RUM application.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("browser"),
				Description: "Type of the RUM application. Supported values are `browser`, `ios`, `android`, `react-native`, `flutter`.",
			},
			"client_token": schema.StringAttribute{
				Computed:    true,
				Description: "The client token.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *rumApplicationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *rumApplicationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumApplicationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetRUMApplication(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumApplication"))
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

func (r *rumApplicationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumApplicationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildRumApplicationRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateRUMApplication(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumApplication"))
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

func (r *rumApplicationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumApplicationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildRumApplicationUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateRUMApplication(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Rum Application"))
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

func (r *rumApplicationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumApplicationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteRUMApplication(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting rum_application"))
		return
	}
}

func (r *rumApplicationResource) updateState(ctx context.Context, state *rumApplicationModel, resp *datadogV2.RUMApplicationResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if clientToken, ok := attributes.GetClientTokenOk(); ok {
		state.ClientToken = types.StringValue(*clientToken)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if typeVar, ok := attributes.GetTypeOk(); ok {
		state.Type = types.StringValue(*typeVar)
	}
}

func (r *rumApplicationResource) buildRumApplicationRequestBody(ctx context.Context, state *rumApplicationModel) (*datadogV2.RUMApplicationCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRUMApplicationCreateAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	if !state.Type.IsNull() {
		attributes.SetType(state.Type.ValueString())
	}

	req := datadogV2.NewRUMApplicationCreateRequestWithDefaults()
	req.Data = *datadogV2.NewRUMApplicationCreateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *rumApplicationResource) buildRumApplicationUpdateRequestBody(ctx context.Context, state *rumApplicationModel) (*datadogV2.RUMApplicationUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRUMApplicationUpdateAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	if !state.Type.IsNull() {
		attributes.SetType(state.Type.ValueString())
	}

	req := datadogV2.NewRUMApplicationUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewRUMApplicationUpdateWithDefaults()
	req.Data.SetId(state.ID.ValueString())
	req.Data.SetAttributes(*attributes)
	return req, diags
}
