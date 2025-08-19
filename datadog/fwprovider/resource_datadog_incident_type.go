package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &incidentTypeResource{}
	_ resource.ResourceWithImportState = &incidentTypeResource{}
)

type incidentTypeResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentTypeModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func NewIncidentTypeResource() resource.Resource {
	return &incidentTypeResource{}
}

func (r *incidentTypeResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	r.Auth = providerData.Auth
}

func (r *incidentTypeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "incident_type"
}

func (r *incidentTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog incident type resource. This can be used to create and manage Datadog incident types.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident type.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the incident type. Must be between 1 and 50 characters.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the incident type. The description can have a maximum of 512 characters.",
				Optional:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this incident type is the default type.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *incidentTypeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.IncidentTypeCreateRequest{
		Data: datadogV2.IncidentTypeCreateData{
			Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			Attributes: datadogV2.IncidentTypeAttributes{
				Name: state.Name.ValueString(),
			},
		},
	}

	if !state.Description.IsNull() {
		body.Data.Attributes.SetDescription(state.Description.ValueString())
	}

	if !state.IsDefault.IsNull() {
		body.Data.Attributes.SetIsDefault(state.IsDefault.ValueBool())
	}

	resp, httpResp, err := r.Api.CreateIncidentType(r.Auth, body)
	if err != nil {
		errorMsg := "Could not create incident type, unexpected error: " + err.Error()
		if httpResp != nil {
			errorMsg += fmt.Sprintf(" (Status: %d)", httpResp.StatusCode)
		}
		response.Diagnostics.AddError(
			"Error creating incident type",
			errorMsg,
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident type",
			fmt.Sprintf("Could not create incident type, status code: %d", httpResp.StatusCode),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetIncidentType(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident type",
			"Could not read incident type, unexpected error: "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.IncidentTypePatchRequest{
		Data: datadogV2.IncidentTypePatchData{
			Type:       datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			Id:         state.ID.ValueString(),
			Attributes: datadogV2.IncidentTypeUpdateAttributes{},
		},
	}

	if !state.Name.IsNull() {
		body.Data.Attributes.SetName(state.Name.ValueString())
	}

	if !state.Description.IsNull() {
		body.Data.Attributes.SetDescription(state.Description.ValueString())
	}

	if !state.IsDefault.IsNull() {
		body.Data.Attributes.SetIsDefault(state.IsDefault.ValueBool())
	}

	resp, httpResp, err := r.Api.UpdateIncidentType(r.Auth, state.ID.ValueString(), body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident type",
			"Could not update incident type, unexpected error: "+err.Error(),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident type",
			fmt.Sprintf("Could not update incident type, status code: %d", httpResp.StatusCode),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteIncidentType(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting incident type",
			fmt.Sprintf("Could not delete incident type, unexpected error: %s (Status: %d)", err.Error(), httpResp.StatusCode),
		)
		return
	}
}

func (r *incidentTypeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *incidentTypeResource) updateStateFromResponse(state *incidentTypeModel, resp *datadogV2.IncidentTypeResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	if attributes, ok := resp.Data.GetAttributesOk(); ok {
		state.Name = types.StringValue(attributes.GetName())
		state.Description = types.StringValue(attributes.GetDescription())
		state.IsDefault = types.BoolValue(attributes.GetIsDefault())
	}
}
