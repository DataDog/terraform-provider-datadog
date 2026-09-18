package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &statusPageMaintenanceTemplateResource{}
	_ resource.ResourceWithImportState = &statusPageMaintenanceTemplateResource{}
)

type statusPageMaintenanceTemplateResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

type statusPageMaintenanceTemplateModel struct {
	ID                    types.String   `tfsdk:"id"`
	PageID                types.String   `tfsdk:"page_id"`
	Name                  types.String   `tfsdk:"name"`
	MaintenanceTitle      types.String   `tfsdk:"maintenance_title"`
	ComponentIds          []types.String `tfsdk:"component_ids"`
	ScheduledDescription  types.String   `tfsdk:"scheduled_description"`
	InProgressDescription types.String   `tfsdk:"in_progress_description"`
	CompletedDescription  types.String   `tfsdk:"completed_description"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	ModifiedAt            types.String   `tfsdk:"modified_at"`
}

func NewStatusPageMaintenanceTemplateResource() resource.Resource {
	return &statusPageMaintenanceTemplateResource{}
}

func (r *statusPageMaintenanceTemplateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "status_page_maintenance_template"
}

func (r *statusPageMaintenanceTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog status page maintenance template resource. This can be used to create and manage pre-filled templates for scheduled maintenances on a status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the maintenance template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the status page this maintenance template belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the maintenance template.",
				Required:    true,
			},
			"maintenance_title": schema.StringAttribute{
				Description: "The title used for a maintenance created from this template.",
				Optional:    true,
			},
			"component_ids": schema.ListAttribute{
				Description: "The IDs of the components affected by a maintenance created from this template.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"scheduled_description": schema.StringAttribute{
				Description: "The pre-filled description shown while the maintenance is scheduled.",
				Optional:    true,
			},
			"in_progress_description": schema.StringAttribute{
				Description: "The pre-filled description shown while the maintenance is in progress.",
				Optional:    true,
			},
			"completed_description": schema.StringAttribute{
				Description: "The pre-filled description shown once the maintenance is completed.",
				Optional:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the maintenance template was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp of when the maintenance template was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *statusPageMaintenanceTemplateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	r.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	r.Auth = providerData.Auth
}

func buildMaintenanceTemplateComponentIds(componentIDs []types.String) []string {
	if componentIDs == nil {
		return nil
	}
	result := make([]string, len(componentIDs))
	for i, id := range componentIDs {
		result[i] = id.ValueString()
	}
	return result
}

func (r *statusPageMaintenanceTemplateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan statusPageMaintenanceTemplateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(plan.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	attributes := &datadogV2.CreateMaintenanceTemplateRequestDataAttributes{
		Name:         plan.Name.ValueString(),
		ComponentIds: buildMaintenanceTemplateComponentIds(plan.ComponentIds),
	}
	if !plan.MaintenanceTitle.IsNull() && !plan.MaintenanceTitle.IsUnknown() {
		attributes.MaintenanceTitle = plan.MaintenanceTitle.ValueStringPointer()
	}
	if !plan.ScheduledDescription.IsNull() && !plan.ScheduledDescription.IsUnknown() {
		attributes.ScheduledDescription = plan.ScheduledDescription.ValueStringPointer()
	}
	if !plan.InProgressDescription.IsNull() && !plan.InProgressDescription.IsUnknown() {
		attributes.InProgressDescription = plan.InProgressDescription.ValueStringPointer()
	}
	if !plan.CompletedDescription.IsNull() && !plan.CompletedDescription.IsUnknown() {
		attributes.CompletedDescription = plan.CompletedDescription.ValueStringPointer()
	}

	body := datadogV2.CreateMaintenanceTemplateRequest{
		Data: &datadogV2.CreateMaintenanceTemplateRequestData{
			Type:       datadogV2.PATCHMAINTENANCETEMPLATEREQUESTDATATYPE_MAINTENANCE_TEMPLATES,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.CreateMaintenanceTemplate(r.Auth, pageID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating maintenance template",
			fmt.Sprintf("Could not create maintenance template, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating maintenance template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state statusPageMaintenanceTemplateModel
	state.PageID = plan.PageID
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageMaintenanceTemplateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageMaintenanceTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	templateID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse maintenance template ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetMaintenanceTemplate(r.Auth, pageID, templateID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading maintenance template",
			"Could not read maintenance template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageMaintenanceTemplateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan statusPageMaintenanceTemplateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(plan.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	templateID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse maintenance template ID: "+err.Error(),
		)
		return
	}

	name := plan.Name.ValueString()
	attributes := &datadogV2.PatchMaintenanceTemplateRequestDataAttributes{
		Name:         &name,
		ComponentIds: buildMaintenanceTemplateComponentIds(plan.ComponentIds),
	}
	if !plan.MaintenanceTitle.IsNull() && !plan.MaintenanceTitle.IsUnknown() {
		attributes.MaintenanceTitle = plan.MaintenanceTitle.ValueStringPointer()
	}
	if !plan.ScheduledDescription.IsNull() && !plan.ScheduledDescription.IsUnknown() {
		attributes.ScheduledDescription = plan.ScheduledDescription.ValueStringPointer()
	}
	if !plan.InProgressDescription.IsNull() && !plan.InProgressDescription.IsUnknown() {
		attributes.InProgressDescription = plan.InProgressDescription.ValueStringPointer()
	}
	if !plan.CompletedDescription.IsNull() && !plan.CompletedDescription.IsUnknown() {
		attributes.CompletedDescription = plan.CompletedDescription.ValueStringPointer()
	}

	body := datadogV2.PatchMaintenanceTemplateRequest{
		Data: &datadogV2.PatchMaintenanceTemplateRequestData{
			Id:         templateID.String(),
			Type:       datadogV2.PATCHMAINTENANCETEMPLATEREQUESTDATATYPE_MAINTENANCE_TEMPLATES,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.UpdateMaintenanceTemplate(r.Auth, pageID, templateID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating maintenance template",
			fmt.Sprintf("Could not update maintenance template ID %s, unexpected error: %s. HTTP Response: %v", plan.ID.ValueString(), err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating maintenance template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *statusPageMaintenanceTemplateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageMaintenanceTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	templateID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse maintenance template ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteMaintenanceTemplate(r.Auth, pageID, templateID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting maintenance template",
			"Could not delete maintenance template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *statusPageMaintenanceTemplateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("unexpected import format", "expected '<page_id>:<template_id>'")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("page_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
}

func (r *statusPageMaintenanceTemplateResource) updateStateFromResponse(state *statusPageMaintenanceTemplateModel, resp *datadogV2.MaintenanceTemplate) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId())

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if statusPage, ok := relationships.GetStatusPageOk(); ok && statusPage != nil {
			pageData := statusPage.GetData()
			state.PageID = types.StringValue(pageData.GetId())
		}
	}

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		if name, ok := attributes.GetNameOk(); ok && name != nil {
			state.Name = types.StringValue(*name)
		}
		if maintenanceTitle, ok := attributes.GetMaintenanceTitleOk(); ok && maintenanceTitle != nil {
			state.MaintenanceTitle = types.StringValue(*maintenanceTitle)
		} else {
			state.MaintenanceTitle = types.StringNull()
		}
		if componentIDs, ok := attributes.GetComponentIdsOk(); ok && componentIDs != nil {
			state.ComponentIds = make([]types.String, len(*componentIDs))
			for i, id := range *componentIDs {
				state.ComponentIds[i] = types.StringValue(id)
			}
		} else {
			state.ComponentIds = nil
		}
		if scheduledDescription, ok := attributes.GetScheduledDescriptionOk(); ok && scheduledDescription != nil {
			state.ScheduledDescription = types.StringValue(*scheduledDescription)
		} else {
			state.ScheduledDescription = types.StringNull()
		}
		if inProgressDescription, ok := attributes.GetInProgressDescriptionOk(); ok && inProgressDescription != nil {
			state.InProgressDescription = types.StringValue(*inProgressDescription)
		} else {
			state.InProgressDescription = types.StringNull()
		}
		if completedDescription, ok := attributes.GetCompletedDescriptionOk(); ok && completedDescription != nil {
			state.CompletedDescription = types.StringValue(*completedDescription)
		} else {
			state.CompletedDescription = types.StringNull()
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.String())
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.String())
		}
	}
}
