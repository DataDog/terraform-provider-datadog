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

func normalizeContent(content string) string {
	return strings.TrimRight(content, "\n")
}

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &incidentNotificationTemplateResource{}
	_ resource.ResourceWithImportState = &incidentNotificationTemplateResource{}
)

type incidentNotificationTemplateResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentNotificationTemplateModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Subject      types.String `tfsdk:"subject"`
	Content      types.String `tfsdk:"content"`
	Category     types.String `tfsdk:"category"`
	IncidentType types.String `tfsdk:"incident_type"`
	Created      types.String `tfsdk:"created"`
	Modified     types.String `tfsdk:"modified"`
}

func NewIncidentNotificationTemplateResource() resource.Resource {
	return &incidentNotificationTemplateResource{}
}

func (r *incidentNotificationTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "incident_notification_template"
}

func (r *incidentNotificationTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog incident notification template resource. This can be used to create and manage Datadog incident notification templates.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident notification template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification template.",
				Required:    true,
			},
			"subject": schema.StringAttribute{
				Description: "The subject line of the notification template.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content body of the notification template.",
				Required:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category of the notification template. Valid values are `alert`, `incident`, `recovery`.",
				Required:    true,
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this notification template is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the notification template was created.",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the notification template was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *incidentNotificationTemplateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	r.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	r.Auth = providerData.Auth
}

func (r *incidentNotificationTemplateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan incidentNotificationTemplateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.CreateIncidentNotificationTemplateRequest{
		Data: datadogV2.IncidentNotificationTemplateCreateData{
			Type: datadogV2.INCIDENTNOTIFICATIONTEMPLATETYPE_NOTIFICATION_TEMPLATES,
			Attributes: datadogV2.IncidentNotificationTemplateCreateAttributes{
				Name:     plan.Name.ValueString(),
				Subject:  plan.Subject.ValueString(),
				Content:  normalizeContent(plan.Content.ValueString()),
				Category: plan.Category.ValueString(),
			},
			Relationships: &datadogV2.IncidentNotificationTemplateCreateDataRelationships{
				IncidentType: &datadogV2.RelationshipToIncidentType{
					Data: datadogV2.RelationshipToIncidentTypeData{
						Id:   plan.IncidentType.ValueString(),
						Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
					},
				},
			},
		},
	}

	resp, httpResp, err := r.Api.CreateIncidentNotificationTemplate(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating incident notification template",
			fmt.Sprintf("Could not create incident notification template, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident notification template",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state incidentNotificationTemplateModel
	r.updateStateFromResponse(&state, &resp)
	state.Content = plan.Content

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentNotificationTemplateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentNotificationTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification template ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetIncidentNotificationTemplate(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident notification template",
			"Could not read incident notification template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with response data
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentNotificationTemplateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan incidentNotificationTemplateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state incidentNotificationTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification template ID: "+err.Error(),
		)
		return
	}

	if !plan.IncidentType.Equal(state.IncidentType) {
		response.Diagnostics.AddError(
			"Incident type cannot be updated",
			"The incident_type field cannot be updated. To change the incident type, the resource must be recreated.",
		)
		return
	}

	name := plan.Name.ValueString()
	subject := plan.Subject.ValueString()
	content := normalizeContent(plan.Content.ValueString())
	category := plan.Category.ValueString()

	body := datadogV2.PatchIncidentNotificationTemplateRequest{
		Data: datadogV2.IncidentNotificationTemplateUpdateData{
			Id:   id,
			Type: datadogV2.INCIDENTNOTIFICATIONTEMPLATETYPE_NOTIFICATION_TEMPLATES,
			Attributes: &datadogV2.IncidentNotificationTemplateUpdateAttributes{
				Name:     &name,
				Subject:  &subject,
				Content:  &content,
				Category: &category,
			},
		},
	}

	resp, httpResp, err := r.Api.UpdateIncidentNotificationTemplate(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident notification template",
			"Could not update incident notification template ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident notification template",
			fmt.Sprintf("Received HTTP status %d", httpResp.StatusCode),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *incidentNotificationTemplateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentNotificationTemplateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification template ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteIncidentNotificationTemplate(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting incident notification template",
			"Could not delete incident notification template ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *incidentNotificationTemplateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *incidentNotificationTemplateResource) updateStateFromResponse(state *incidentNotificationTemplateModel, resp *datadogV2.IncidentNotificationTemplate) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		state.Name = types.StringValue(attributes.GetName())
		state.Subject = types.StringValue(attributes.GetSubject())

		apiContent := attributes.GetContent()
		if state.Content.IsNull() || normalizeContent(state.Content.ValueString()) != apiContent {
			state.Content = types.StringValue(apiContent)
		}

		state.Category = types.StringValue(attributes.GetCategory())
		state.Created = types.StringValue(attributes.GetCreated().Format("2006-01-02T15:04:05Z"))
		state.Modified = types.StringValue(attributes.GetModified().Format("2006-01-02T15:04:05Z"))
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId())
			}
		}
	}
}
