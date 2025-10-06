package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &incidentNotificationRuleResource{}
	_ resource.ResourceWithImportState = &incidentNotificationRuleResource{}
)

type incidentNotificationRuleResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentNotificationRuleModel struct {
	ID                   types.String                             `tfsdk:"id"`
	Conditions           []incidentNotificationRuleConditionModel `tfsdk:"conditions"`
	Enabled              types.Bool                               `tfsdk:"enabled"`
	Handles              []types.String                           `tfsdk:"handles"`
	RenotifyOn           []types.String                           `tfsdk:"renotify_on"`
	Trigger              types.String                             `tfsdk:"trigger"`
	Visibility           types.String                             `tfsdk:"visibility"`
	IncidentType         types.String                             `tfsdk:"incident_type"`
	NotificationTemplate types.String                             `tfsdk:"notification_template"`
	Created              types.String                             `tfsdk:"created"`
	Modified             types.String                             `tfsdk:"modified"`
}

type incidentNotificationRuleConditionModel struct {
	Field  types.String   `tfsdk:"field"`
	Values []types.String `tfsdk:"values"`
}

func NewIncidentNotificationRuleResource() resource.Resource {
	return &incidentNotificationRuleResource{}
}

func (r *incidentNotificationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "incident_notification_rule"
}

func (r *incidentNotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog incident notification rule resource. This can be used to create and manage Datadog incident notification rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident notification rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the notification rule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"handles": schema.ListAttribute{
				Description: "The notification handles (targets) for this rule. Examples: @team-email@company.com, @slack-channel.",
				Required:    true,
				ElementType: types.StringType,
			},
			"renotify_on": schema.ListAttribute{
				Description: "List of incident fields that trigger re-notification when changed. Valid values are: status, severity, customer_impact, title, description, detected, root_cause, services, state.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"trigger": schema.StringAttribute{
				Description: "The trigger event for this notification rule. Valid values are: incident_created_trigger, incident_saved_trigger.",
				Required:    true,
			},
			"visibility": schema.StringAttribute{
				Description: "The visibility of the notification rule. Valid values are: all, organization, private. Defaults to organization.",
				Optional:    true,
				Computed:    true,
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this notification rule is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"notification_template": schema.StringAttribute{
				Description: "The ID of the notification template to use for this rule.",
				Optional:    true,
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the notification rule was created.",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the notification rule was last modified.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"conditions": schema.ListNestedBlock{
				Description: "The conditions that trigger this notification rule. At least one condition is required.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "The incident field to evaluate. Common values include: state, severity, services, teams. Custom fields are also supported.",
							Required:    true,
						},
						"values": schema.ListAttribute{
							Description: "The value(s) to compare against. Multiple values are ORed together.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *incidentNotificationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *incidentNotificationRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan incidentNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Build conditions
	conditions := make([]datadogV2.IncidentNotificationRuleConditionsItems, len(plan.Conditions))
	for i, condition := range plan.Conditions {
		values := make([]string, len(condition.Values))
		for j, value := range condition.Values {
			values[j] = value.ValueString()
		}
		conditions[i] = datadogV2.IncidentNotificationRuleConditionsItems{
			Field:  condition.Field.ValueString(),
			Values: values,
		}
	}

	// Build handles
	handles := make([]string, len(plan.Handles))
	for i, handle := range plan.Handles {
		handles[i] = handle.ValueString()
	}

	// Build renotify_on
	var renotifyOn []string
	if len(plan.RenotifyOn) > 0 {
		renotifyOn = make([]string, len(plan.RenotifyOn))
		for i, item := range plan.RenotifyOn {
			renotifyOn[i] = item.ValueString()
		}
	}

	// Build attributes
	attributes := datadogV2.IncidentNotificationRuleCreateAttributes{
		Conditions: conditions,
		Handles:    handles,
		Trigger:    plan.Trigger.ValueString(),
	}

	if !plan.Enabled.IsNull() {
		enabled := plan.Enabled.ValueBool()
		attributes.Enabled = &enabled
	}

	if len(renotifyOn) > 0 {
		attributes.RenotifyOn = renotifyOn
	}

	if !plan.Visibility.IsNull() && !plan.Visibility.IsUnknown() {
		visibility := datadogV2.IncidentNotificationRuleCreateAttributesVisibility(plan.Visibility.ValueString())
		attributes.Visibility = &visibility
	}

	// Build relationships
	relationships := &datadogV2.IncidentNotificationRuleCreateDataRelationships{
		IncidentType: &datadogV2.RelationshipToIncidentType{
			Data: datadogV2.RelationshipToIncidentTypeData{
				Id:   plan.IncidentType.ValueString(),
				Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			},
		},
	}

	if !plan.NotificationTemplate.IsNull() && !plan.NotificationTemplate.IsUnknown() {
		templateId, err := uuid.Parse(plan.NotificationTemplate.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				"Error parsing notification template ID",
				"Could not parse notification template ID: "+err.Error(),
			)
			return
		}
		relationships.NotificationTemplate = &datadogV2.RelationshipToIncidentNotificationTemplate{
			Data: datadogV2.RelationshipToIncidentNotificationTemplateData{
				Id:   templateId,
				Type: datadogV2.INCIDENTNOTIFICATIONTEMPLATETYPE_NOTIFICATION_TEMPLATES,
			},
		}
	}

	body := datadogV2.CreateIncidentNotificationRuleRequest{
		Data: datadogV2.IncidentNotificationRuleCreateData{
			Type:          datadogV2.INCIDENTNOTIFICATIONRULETYPE_INCIDENT_NOTIFICATION_RULES,
			Attributes:    attributes,
			Relationships: relationships,
		},
	}

	resp, httpResp, err := r.Api.CreateIncidentNotificationRule(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating incident notification rule",
			fmt.Sprintf("Could not create incident notification rule, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident notification rule",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state incidentNotificationRuleModel
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentNotificationRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification rule ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetIncidentNotificationRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident notification rule",
			"Could not read incident notification rule ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with response data
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentNotificationRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan incidentNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification rule ID: "+err.Error(),
		)
		return
	}

	// Build conditions
	conditions := make([]datadogV2.IncidentNotificationRuleConditionsItems, len(plan.Conditions))
	for i, condition := range plan.Conditions {
		values := make([]string, len(condition.Values))
		for j, value := range condition.Values {
			values[j] = value.ValueString()
		}
		conditions[i] = datadogV2.IncidentNotificationRuleConditionsItems{
			Field:  condition.Field.ValueString(),
			Values: values,
		}
	}

	// Build handles
	handles := make([]string, len(plan.Handles))
	for i, handle := range plan.Handles {
		handles[i] = handle.ValueString()
	}

	// Build renotify_on
	var renotifyOn []string
	if len(plan.RenotifyOn) > 0 {
		renotifyOn = make([]string, len(plan.RenotifyOn))
		for i, item := range plan.RenotifyOn {
			renotifyOn[i] = item.ValueString()
		}
	}

	// Build update attributes
	enabled := plan.Enabled.ValueBool()
	trigger := plan.Trigger.ValueString()

	attributes := &datadogV2.IncidentNotificationRuleCreateAttributes{
		Conditions: conditions,
		Enabled:    &enabled,
		Handles:    handles,
		RenotifyOn: renotifyOn,
		Trigger:    trigger,
	}

	if !plan.Visibility.IsNull() && !plan.Visibility.IsUnknown() {
		visibility := datadogV2.IncidentNotificationRuleCreateAttributesVisibility(plan.Visibility.ValueString())
		attributes.Visibility = &visibility
	}

	// Build relationships for update
	relationships := &datadogV2.IncidentNotificationRuleCreateDataRelationships{
		IncidentType: &datadogV2.RelationshipToIncidentType{
			Data: datadogV2.RelationshipToIncidentTypeData{
				Id:   plan.IncidentType.ValueString(),
				Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			},
		},
	}

	if !plan.NotificationTemplate.IsNull() && !plan.NotificationTemplate.IsUnknown() {
		templateId, err := uuid.Parse(plan.NotificationTemplate.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				"Error parsing notification template ID",
				"Could not parse notification template ID: "+err.Error(),
			)
			return
		}
		relationships.NotificationTemplate = &datadogV2.RelationshipToIncidentNotificationTemplate{
			Data: datadogV2.RelationshipToIncidentNotificationTemplateData{
				Id:   templateId,
				Type: datadogV2.INCIDENTNOTIFICATIONTEMPLATETYPE_NOTIFICATION_TEMPLATES,
			},
		}
	}

	body := datadogV2.PutIncidentNotificationRuleRequest{
		Data: datadogV2.IncidentNotificationRuleUpdateData{
			Id:            id,
			Type:          datadogV2.INCIDENTNOTIFICATIONRULETYPE_INCIDENT_NOTIFICATION_RULES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	}

	resp, httpResp, err := r.Api.UpdateIncidentNotificationRule(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident notification rule",
			fmt.Sprintf("Could not update incident notification rule ID %s, unexpected error: %s. HTTP Response: %v", plan.ID.ValueString(), err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident notification rule",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *incidentNotificationRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification rule ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteIncidentNotificationRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting incident notification rule",
			"Could not delete incident notification rule ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *incidentNotificationRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *incidentNotificationRuleResource) updateStateFromResponse(state *incidentNotificationRuleModel, resp *datadogV2.IncidentNotificationRule) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		// Convert conditions
		if conditions, conditionsOk := attributes.GetConditionsOk(); conditionsOk && conditions != nil {
			state.Conditions = make([]incidentNotificationRuleConditionModel, len(*conditions))
			for i, condition := range *conditions {
				values := make([]types.String, len(condition.Values))
				for j, value := range condition.Values {
					values[j] = types.StringValue(value)
				}
				state.Conditions[i] = incidentNotificationRuleConditionModel{
					Field:  types.StringValue(condition.Field),
					Values: values,
				}
			}
		}

		if enabled, enabledOk := attributes.GetEnabledOk(); enabledOk && enabled != nil {
			state.Enabled = types.BoolValue(*enabled)
		}

		if handles, handlesOk := attributes.GetHandlesOk(); handlesOk && handles != nil {
			state.Handles = make([]types.String, len(*handles))
			for i, handle := range *handles {
				state.Handles[i] = types.StringValue(handle)
			}
		}

		if renotifyOn, renotifyOnOk := attributes.GetRenotifyOnOk(); renotifyOnOk && renotifyOn != nil {
			state.RenotifyOn = make([]types.String, len(*renotifyOn))
			for i, item := range *renotifyOn {
				state.RenotifyOn[i] = types.StringValue(item)
			}
		}

		state.Trigger = types.StringValue(attributes.GetTrigger())

		if visibility, visibilityOk := attributes.GetVisibilityOk(); visibilityOk && visibility != nil {
			state.Visibility = types.StringValue(string(*visibility))
		}

		if created, createdOk := attributes.GetCreatedOk(); createdOk && created != nil {
			state.Created = types.StringValue(created.Format("2006-01-02T15:04:05Z"))
		}

		if modified, modifiedOk := attributes.GetModifiedOk(); modifiedOk && modified != nil {
			state.Modified = types.StringValue(modified.Format("2006-01-02T15:04:05Z"))
		}
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId())
			}
		}

		if notificationTemplate, ok := relationships.GetNotificationTemplateOk(); ok && notificationTemplate != nil {
			if templateData, ok := notificationTemplate.GetDataOk(); ok && templateData != nil {
				state.NotificationTemplate = types.StringValue(templateData.GetId().String())
			}
		}
	}
}
