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
	_ resource.ResourceWithConfigure   = &statusPageDegradationTemplateResource{}
	_ resource.ResourceWithImportState = &statusPageDegradationTemplateResource{}
)

type statusPageDegradationTemplateResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

type statusPageDegradationTemplateModel struct {
	ID                 types.String                                          `tfsdk:"id"`
	PageID             types.String                                          `tfsdk:"page_id"`
	Name               types.String                                          `tfsdk:"name"`
	DegradationTitle   types.String                                          `tfsdk:"degradation_title"`
	ComponentsAffected []statusPageDegradationTemplateComponentAffectedModel `tfsdk:"components_affected"`
	Updates            []statusPageDegradationTemplateUpdateModel            `tfsdk:"updates"`
	CreatedAt          types.String                                          `tfsdk:"created_at"`
	ModifiedAt         types.String                                          `tfsdk:"modified_at"`
}

type statusPageDegradationTemplateComponentAffectedModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

type statusPageDegradationTemplateUpdateModel struct {
	Message types.String `tfsdk:"message"`
	Status  types.String `tfsdk:"status"`
}

func NewStatusPageDegradationTemplateResource() resource.Resource {
	return &statusPageDegradationTemplateResource{}
}

func (r *statusPageDegradationTemplateResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "status_page_degradation_template"
}

func (r *statusPageDegradationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog status page degradation template resource. This can be used to create and manage pre-filled templates for degradations on a status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the degradation template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the status page this degradation template belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the degradation template.",
				Required:    true,
			},
			"degradation_title": schema.StringAttribute{
				Description: "The title used for a degradation created from this template.",
				Optional:    true,
			},
			"components_affected": schema.ListNestedAttribute{
				Description: "The components affected by a degradation created from this template.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the affected component.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the affected component.",
							Optional:    true,
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The pre-filled status for this component. Valid values are: operational, degraded, partial_outage, major_outage.",
							Required:    true,
						},
					},
				},
			},
			"updates": schema.ListNestedAttribute{
				Description: "The pre-filled updates for a degradation created from this template.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"message": schema.StringAttribute{
							Description: "The pre-filled message for this update.",
							Optional:    true,
						},
						"status": schema.StringAttribute{
							Description: "The pre-filled degradation status for this update. Valid values are: investigating, identified, monitoring, resolved.",
							Required:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the degradation template was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp of when the degradation template was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *statusPageDegradationTemplateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func buildDegradationTemplateComponentsAffected(components []statusPageDegradationTemplateComponentAffectedModel) []datadogV2.CreateDegradationTemplateRequestDataAttributesComponentsAffectedItems {
	if components == nil {
		return nil
	}
	result := make([]datadogV2.CreateDegradationTemplateRequestDataAttributesComponentsAffectedItems, len(components))
	for i, component := range components {
		item := datadogV2.CreateDegradationTemplateRequestDataAttributesComponentsAffectedItems{
			Id:     component.ID.ValueString(),
			Status: datadogV2.PatchDegradationTemplateRequestDataAttributesComponentsAffectedItemsStatus(component.Status.ValueString()),
		}
		if !component.Name.IsNull() && !component.Name.IsUnknown() {
			item.Name = component.Name.ValueStringPointer()
		}
		result[i] = item
	}
	return result
}

func buildDegradationTemplateUpdates(updates []statusPageDegradationTemplateUpdateModel) []datadogV2.CreateDegradationTemplateRequestDataAttributesUpdatesItems {
	if updates == nil {
		return nil
	}
	result := make([]datadogV2.CreateDegradationTemplateRequestDataAttributesUpdatesItems, len(updates))
	for i, update := range updates {
		item := datadogV2.CreateDegradationTemplateRequestDataAttributesUpdatesItems{
			Status: datadogV2.CreateDegradationRequestDataAttributesStatus(update.Status.ValueString()),
		}
		if !update.Message.IsNull() && !update.Message.IsUnknown() {
			item.Message = update.Message.ValueStringPointer()
		}
		result[i] = item
	}
	return result
}

func (r *statusPageDegradationTemplateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan statusPageDegradationTemplateModel
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

	attributes := &datadogV2.CreateDegradationTemplateRequestDataAttributes{
		Name:               plan.Name.ValueString(),
		ComponentsAffected: buildDegradationTemplateComponentsAffected(plan.ComponentsAffected),
		Updates:            buildDegradationTemplateUpdates(plan.Updates),
	}
	if !plan.DegradationTitle.IsNull() && !plan.DegradationTitle.IsUnknown() {
		attributes.DegradationTitle = plan.DegradationTitle.ValueStringPointer()
	}

	body := datadogV2.CreateDegradationTemplateRequest{
		Data: &datadogV2.CreateDegradationTemplateRequestData{
			Type:       datadogV2.PATCHDEGRADATIONTEMPLATEREQUESTDATATYPE_DEGRADATION_TEMPLATES,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.CreateDegradationTemplate(r.Auth, pageID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating degradation template",
			fmt.Sprintf("Could not create degradation template, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating degradation template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state statusPageDegradationTemplateModel
	state.PageID = plan.PageID
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageDegradationTemplateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageDegradationTemplateModel
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
			"Could not parse degradation template ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetDegradationTemplate(r.Auth, pageID, templateID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading degradation template",
			"Could not read degradation template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageDegradationTemplateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan statusPageDegradationTemplateModel
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
			"Could not parse degradation template ID: "+err.Error(),
		)
		return
	}

	name := plan.Name.ValueString()
	attributes := &datadogV2.PatchDegradationTemplateRequestDataAttributes{
		Name:               &name,
		ComponentsAffected: buildDegradationTemplatePatchComponentsAffected(plan.ComponentsAffected),
		Updates:            buildDegradationTemplatePatchUpdates(plan.Updates),
	}
	if !plan.DegradationTitle.IsNull() && !plan.DegradationTitle.IsUnknown() {
		attributes.DegradationTitle = plan.DegradationTitle.ValueStringPointer()
	}

	body := datadogV2.PatchDegradationTemplateRequest{
		Data: &datadogV2.PatchDegradationTemplateRequestData{
			Id:         templateID.String(),
			Type:       datadogV2.PATCHDEGRADATIONTEMPLATEREQUESTDATATYPE_DEGRADATION_TEMPLATES,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.UpdateDegradationTemplate(r.Auth, templateID, pageID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating degradation template",
			fmt.Sprintf("Could not update degradation template ID %s, unexpected error: %s. HTTP Response: %v", plan.ID.ValueString(), err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating degradation template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func buildDegradationTemplatePatchComponentsAffected(components []statusPageDegradationTemplateComponentAffectedModel) []datadogV2.PatchDegradationTemplateRequestDataAttributesComponentsAffectedItems {
	if components == nil {
		return nil
	}
	result := make([]datadogV2.PatchDegradationTemplateRequestDataAttributesComponentsAffectedItems, len(components))
	for i, component := range components {
		item := datadogV2.PatchDegradationTemplateRequestDataAttributesComponentsAffectedItems{
			Id:     component.ID.ValueString(),
			Status: datadogV2.PatchDegradationTemplateRequestDataAttributesComponentsAffectedItemsStatus(component.Status.ValueString()),
		}
		if !component.Name.IsNull() && !component.Name.IsUnknown() {
			item.Name = component.Name.ValueStringPointer()
		}
		result[i] = item
	}
	return result
}

func buildDegradationTemplatePatchUpdates(updates []statusPageDegradationTemplateUpdateModel) []datadogV2.PatchDegradationTemplateRequestDataAttributesUpdatesItems {
	if updates == nil {
		return nil
	}
	result := make([]datadogV2.PatchDegradationTemplateRequestDataAttributesUpdatesItems, len(updates))
	for i, update := range updates {
		item := datadogV2.PatchDegradationTemplateRequestDataAttributesUpdatesItems{
			Status: datadogV2.CreateDegradationRequestDataAttributesStatus(update.Status.ValueString()),
		}
		if !update.Message.IsNull() && !update.Message.IsUnknown() {
			item.Message = update.Message.ValueStringPointer()
		}
		result[i] = item
	}
	return result
}

func (r *statusPageDegradationTemplateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageDegradationTemplateModel
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
			"Could not parse degradation template ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteDegradationTemplate(r.Auth, pageID, templateID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting degradation template",
			"Could not delete degradation template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *statusPageDegradationTemplateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("unexpected import format", "expected '<page_id>:<template_id>'")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("page_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
}

func (r *statusPageDegradationTemplateResource) updateStateFromResponse(state *statusPageDegradationTemplateModel, resp *datadogV2.DegradationTemplate) {
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
		if degradationTitle, ok := attributes.GetDegradationTitleOk(); ok && degradationTitle != nil {
			state.DegradationTitle = types.StringValue(*degradationTitle)
		} else {
			state.DegradationTitle = types.StringNull()
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.String())
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.String())
		}
		if components, ok := attributes.GetComponentsAffectedOk(); ok && components != nil {
			state.ComponentsAffected = make([]statusPageDegradationTemplateComponentAffectedModel, len(*components))
			for i, component := range *components {
				componentModel := statusPageDegradationTemplateComponentAffectedModel{
					ID:     types.StringValue(component.Id),
					Status: types.StringValue(string(component.Status)),
				}
				if component.Name != nil {
					componentModel.Name = types.StringValue(*component.Name)
				}
				state.ComponentsAffected[i] = componentModel
			}
		} else {
			state.ComponentsAffected = nil
		}
		if updates, ok := attributes.GetUpdatesOk(); ok && updates != nil {
			state.Updates = make([]statusPageDegradationTemplateUpdateModel, len(*updates))
			for i, update := range *updates {
				updateModel := statusPageDegradationTemplateUpdateModel{
					Status: types.StringValue(string(update.Status)),
				}
				if update.Message != nil {
					updateModel.Message = types.StringValue(*update.Message)
				}
				state.Updates[i] = updateModel
			}
		} else {
			state.Updates = nil
		}
	}
}
