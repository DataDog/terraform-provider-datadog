package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &incidentHandleResource{}
	_ resource.ResourceWithImportState = &incidentHandleResource{}
)

type incidentHandleResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentHandleModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	IncidentTypeID  types.String `tfsdk:"incident_type_id"`
	Severity        types.List   `tfsdk:"severity"`
	CommanderUserID types.String `tfsdk:"commander_user_id"`
}

func NewIncidentHandleResource() resource.Resource {
	return &incidentHandleResource{}
}

func (r *incidentHandleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	r.Auth = providerData.Auth
}

func (r *incidentHandleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "incident_handle"
}

func (r *incidentHandleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog global incident handle resource. An incident handle (for example `@incident-sev-1`) can be mentioned in monitor notifications to declare an incident with the configured incident type and fields. Note: the Datadog API has no endpoint to delete a single handle, so destroying this resource only removes it from Terraform state; delete the handle in the Datadog UI.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident handle.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The handle name, including the leading `@`, for example `@incident-sev-1`.",
				Required:    true,
			},
			"incident_type_id": schema.StringAttribute{
				Description: "The ID of the incident type for incidents declared through this handle.",
				Required:    true,
			},
			"severity": schema.ListAttribute{
				Description: "Severities set on incidents declared through this handle, for example `[\"SEV-1\"]`.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"commander_user_id": schema.StringAttribute{
				Description: "The ID of the user assigned as incident commander for incidents declared through this handle.",
				Optional:    true,
			},
		},
	}
}

func (r *incidentHandleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state incidentHandleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := buildIncidentHandleRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.CreateGlobalIncidentHandle(r.Auth, *body)
	if err != nil {
		errorMsg := "Could not create incident handle, unexpected error: " + err.Error()
		if httpResp != nil {
			errorMsg += fmt.Sprintf(" (Status: %d)", httpResp.StatusCode)
		}
		response.Diagnostics.AddError("Error creating incident handle", errorMsg)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident handle",
			fmt.Sprintf("Could not create incident handle, status code: %d", httpResp.StatusCode),
		)
		return
	}

	response.Diagnostics.Append(updateIncidentHandleState(ctx, &state, &resp.Data)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentHandleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentHandleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// The API has no get-by-ID endpoint, so read through the list.
	resp, _, err := r.Api.ListGlobalIncidentHandles(r.Auth)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading incident handle",
			"Could not list incident handles, unexpected error: "+err.Error(),
		)
		return
	}

	for i := range resp.Data {
		if resp.Data[i].GetId() == state.ID.ValueString() {
			response.Diagnostics.Append(updateIncidentHandleState(ctx, &state, &resp.Data[i])...)
			response.Diagnostics.Append(response.State.Set(ctx, &state)...)
			return
		}
	}
	response.State.RemoveResource(ctx)
}

func (r *incidentHandleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state incidentHandleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := buildIncidentHandleRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	body.Data.SetId(state.ID.ValueString())

	resp, httpResp, err := r.Api.UpdateGlobalIncidentHandle(r.Auth, *body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident handle",
			"Could not update incident handle, unexpected error: "+err.Error(),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident handle",
			fmt.Sprintf("Could not update incident handle, status code: %d", httpResp.StatusCode),
		)
		return
	}

	response.Diagnostics.Append(updateIncidentHandleState(ctx, &state, &resp.Data)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Delete only drops the handle from state. DeleteGlobalIncidentHandle takes no handle ID
// (DELETE on the collection path with no parameters or body), so calling it cannot target
// this handle and might affect every handle in the org.
func (r *incidentHandleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentHandleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.AddWarning(
		"Incident handle not deleted in Datadog",
		fmt.Sprintf("The Datadog API does not support deleting a single incident handle. Handle %s (%s) was removed from Terraform state only; delete it in the Datadog UI.", state.Name.ValueString(), state.ID.ValueString()),
	)
}

func (r *incidentHandleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func buildIncidentHandleRequest(ctx context.Context, state *incidentHandleModel) (*datadogV2.IncidentHandleRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := datadogV2.NewIncidentHandleAttributesRequest(state.Name.ValueString())
	if !state.Severity.IsNull() && !state.Severity.IsUnknown() {
		var severity []string
		diags.Append(state.Severity.ElementsAs(ctx, &severity, false)...)
		attributes.SetFields(datadogV2.IncidentHandleAttributesFields{Severity: severity})
	}

	relationships := datadogV2.NewIncidentHandleRelationshipsRequest(*datadogV2.NewIncidentHandleRelationship(
		*datadogV2.NewIncidentHandleRelationshipData(state.IncidentTypeID.ValueString(), "incident_types"),
	))
	if !state.CommanderUserID.IsNull() && !state.CommanderUserID.IsUnknown() {
		relationships.SetCommanderUser(*datadogV2.NewIncidentHandleRelationship(
			*datadogV2.NewIncidentHandleRelationshipData(state.CommanderUserID.ValueString(), "users"),
		))
	}

	data := datadogV2.NewIncidentHandleDataRequest(*attributes, datadogV2.INCIDENTHANDLETYPE_INCIDENTS_HANDLES)
	data.SetRelationships(*relationships)
	return datadogV2.NewIncidentHandleRequest(*data), diags
}

func updateIncidentHandleState(ctx context.Context, state *incidentHandleModel, data *datadogV2.IncidentHandleDataResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	state.ID = types.StringValue(data.GetId())
	state.Name = types.StringValue(data.Attributes.GetName())

	severity := data.Attributes.Fields.GetSeverity()
	if len(severity) > 0 || !state.Severity.IsNull() {
		state.Severity, diags = types.ListValueFrom(ctx, types.StringType, severity)
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		state.IncidentTypeID = types.StringValue(relationships.IncidentType.Data.GetId())
		if commander, ok := relationships.GetCommanderUserOk(); ok {
			state.CommanderUserID = types.StringValue(commander.Data.GetId())
		} else {
			state.CommanderUserID = types.StringNull()
		}
	}
	return diags
}
