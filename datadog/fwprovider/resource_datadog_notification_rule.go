package fwprovider

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &NotificationRuleResource{}
	_ resource.ResourceWithImportState = &NotificationRuleResource{}
)

type NotificationRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

var writeMutex = sync.Mutex{}

type notificationRuleModel struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Selectors        selectorsModel `tfsdk:"selectors"`
	Targets          types.List     `tfsdk:"targets"`
	TimeAggregation  types.Int64    `tfsdk:"time_aggregation"`
	Version          types.Int64    `tfsdk:"version"`
	Enabled          types.Bool     `tfsdk:"enabled"`
	CreatedAt        types.Int64    `tfsdk:"created_at"`
	ModifiedAt       types.Int64    `tfsdk:"modified_at"`
	CreatedByName    types.String   `tfsdk:"created_by_name"`
	CreatedByHandle  types.String   `tfsdk:"created_by_handle"`
	ModifiedByName   types.String   `tfsdk:"modified_by_name"`
	ModifiedByHandle types.String   `tfsdk:"modified_by_handle"`
}

type selectorsModel struct {
	TriggerSource types.String `tfsdk:"trigger_source"`
	RuleTypes     types.List   `tfsdk:"rule_types"`
	Severities    types.List   `tfsdk:"severities"`
	Query         types.String `tfsdk:"query"`
}

func (r *NotificationRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *NotificationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func NewNotificationRuleResource() resource.Resource {
	return &NotificationRuleResource{}
}

func (r *NotificationRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "notification_rule"
}

func (r *NotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Notification Rule API resource. It can be used to create and manage Datadog security notification rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the notification rule.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the rule (must be unique).",
				Required:    true,
			},
			"targets": schema.ListAttribute{
				Description: "List of handle targets for the notifications.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.AtLeastOneOf(),
				},
			},
			"time_aggregation": schema.Int64Attribute{
				Description: "Time period (in seconds) used to aggregate the notification.",
				Optional:    true,
			},
			"version": schema.Int64Attribute{
				Description: "Rule version (incremented at each update).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the rule is enabled.",
				Optional:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "When this rule was created.",
				Computed:    true,
			},
			"modified_at": schema.Int64Attribute{
				Description: "When this rule was last modified.",
				Computed:    true,
			},
			"created_by_name": schema.StringAttribute{
				Description: "Name of the rule creator.",
				Computed:    true,
			},
			"modified_by_name": schema.StringAttribute{
				Description: "Name of the rule last modifier.",
				Computed:    true,
			},
			"created_by_handle": schema.StringAttribute{
				Description: "Handle of the rule creator.",
				Computed:    true,
			},
			"modified_by_handle": schema.StringAttribute{
				Description: "Handle of the rule last modifier.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"selectors": schema.SingleNestedBlock{
				Description: "Selectors used to filter security issues for which notifications are generated.",
				Attributes: map[string]schema.Attribute{
					"trigger_source": schema.StringAttribute{
						Description: "The type of security issues on which the rule applies. Rules based on security signals must use the trigger source security_signals, while rules based on vulnerabilities must use security_findings.",
						Required:    true,
					},
					"rule_types": schema.ListAttribute{
						Description: "Security rule types used to filter signals and vulnerabilities generating notifications.",
						Required:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.AtLeastOneOf(),
						},
					},
					"severities": schema.ListAttribute{
						Description: "The security rules severities to consider.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"query": schema.StringAttribute{
						Description: "The query is composed of one or several key:value pairs, which can be used to filter security issues on tags and attributes.",
						Optional:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
		},
	}
}

func (r *NotificationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	triggerSource, err := datadogV2.NewTriggerSourceFromValue(plan.Selectors.TriggerSource.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid trigger source", err.Error())
		return
	}
	selectors := datadogV2.NewSelectors(*triggerSource)
	if !plan.Selectors.RuleTypes.IsNull() {
		ruleTypes := make([]datadogV2.RuleTypesItems, len(plan.Selectors.RuleTypes.Elements()))
		resp.Diagnostics.Append(plan.Selectors.RuleTypes.ElementsAs(ctx, &ruleTypes, false)...)
		selectors.SetRuleTypes(ruleTypes)
	}
	if !plan.Selectors.Severities.IsNull() {
		severities := make([]datadogV2.RuleSeverity, len(plan.Selectors.Severities.Elements()))
		resp.Diagnostics.Append(plan.Selectors.Severities.ElementsAs(ctx, &severities, false)...)
		selectors.SetSeverities(severities)
	}
	if !plan.Selectors.Query.IsNull() {
		selectors.SetQuery(plan.Selectors.Query.ValueString())
	}

	targets := make([]string, len(plan.Targets.Elements()))
	resp.Diagnostics.Append(plan.Targets.ElementsAs(ctx, &targets, false)...)

	ruleAttributes := datadogV2.NewCreateNotificationRuleParametersDataAttributesWithDefaults()
	ruleAttributes.SetName(plan.Name.ValueString())
	ruleAttributes.SetSelectors(*selectors)
	ruleAttributes.SetTargets(targets)
	if !plan.Enabled.IsNull() {
		ruleAttributes.SetEnabled(plan.Enabled.ValueBool())
	}
	if !plan.TimeAggregation.IsNull() {
		ruleAttributes.SetTimeAggregation(plan.TimeAggregation.ValueInt64())
	}

	body := datadogV2.NewCreateNotificationRuleParametersWithDefaults()
	body.Data = datadogV2.NewCreateNotificationRuleParametersData(*ruleAttributes, datadogV2.NOTIFICATIONRULESTYPE_NOTIFICATION_RULES)

	writeMutex.Lock()
	defer writeMutex.Unlock()

	var response datadogV2.NotificationRuleResponse
	if *triggerSource == datadogV2.TRIGGERSOURCE_SECURITY_SIGNALS {
		response, _, err = r.api.CreateSignalNotificationRule(r.auth, *body)
	} else {
		response, _, err = r.api.CreateVulnerabilityNotificationRule(r.auth, *body)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification rule", err.Error())
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		resp.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	r.updateStateFromResponse(ctx, &plan, &response)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NotificationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	triggerSource, err := datadogV2.NewTriggerSourceFromValue(state.Selectors.TriggerSource.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid trigger source", err.Error())
		return
	}

	var response datadogV2.NotificationRuleResponse
	if *triggerSource == datadogV2.TRIGGERSOURCE_SECURITY_SIGNALS {
		response, _, err = r.api.GetSignalNotificationRule(r.auth, state.ID.ValueString())
	} else {
		response, _, err = r.api.GetVulnerabilityNotificationRule(r.auth, state.ID.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading notification rule", err.Error())
		return
	}

	r.updateStateFromResponse(ctx, &state, &response)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *NotificationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var previousState notificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &previousState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := previousState.ID.ValueString()
	version := previousState.Version.ValueInt64()

	triggerSource, err := datadogV2.NewTriggerSourceFromValue(plan.Selectors.TriggerSource.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid trigger source", err.Error())
		return
	}
	selectors := datadogV2.NewSelectors(*triggerSource)
	if !plan.Selectors.RuleTypes.IsNull() {
		ruleTypes := make([]datadogV2.RuleTypesItems, len(plan.Selectors.RuleTypes.Elements()))
		resp.Diagnostics.Append(plan.Selectors.RuleTypes.ElementsAs(ctx, &ruleTypes, false)...)
		selectors.SetRuleTypes(ruleTypes)
	}
	if !plan.Selectors.Severities.IsNull() {
		severities := make([]datadogV2.RuleSeverity, len(plan.Selectors.Severities.Elements()))
		resp.Diagnostics.Append(plan.Selectors.Severities.ElementsAs(ctx, &severities, false)...)
		selectors.SetSeverities(severities)
	}
	if !plan.Selectors.Query.IsNull() {
		selectors.SetQuery(plan.Selectors.Query.ValueString())
	}

	targets := make([]string, len(plan.Targets.Elements()))
	resp.Diagnostics.Append(plan.Targets.ElementsAs(ctx, &targets, false)...)

	ruleAttributes := datadogV2.NewPatchNotificationRuleParametersDataAttributes()
	ruleAttributes.SetName(plan.Name.ValueString())
	ruleAttributes.SetVersion(version)
	ruleAttributes.SetSelectors(*selectors)
	ruleAttributes.SetTargets(targets)
	if !plan.Enabled.IsNull() {
		ruleAttributes.SetEnabled(plan.Enabled.ValueBool())
	}
	if !plan.TimeAggregation.IsNull() {
		ruleAttributes.SetTimeAggregation(plan.TimeAggregation.ValueInt64())
	}

	body := datadogV2.NewPatchNotificationRuleParameters()
	body.Data = datadogV2.NewPatchNotificationRuleParametersData(*ruleAttributes, id, datadogV2.NOTIFICATIONRULESTYPE_NOTIFICATION_RULES)

	writeMutex.Lock()
	defer writeMutex.Unlock()

	var response datadogV2.NotificationRuleResponse
	if *triggerSource == datadogV2.TRIGGERSOURCE_SECURITY_SIGNALS {
		response, _, err = r.api.PatchSignalNotificationRule(r.auth, id, *body)
	} else {
		response, _, err = r.api.PatchVulnerabilityNotificationRule(r.auth, id, *body)
	}

	if err != nil {
		resp.Diagnostics.AddError("Error updating notification rule", err.Error())
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		resp.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	r.updateStateFromResponse(ctx, &plan, &response)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *NotificationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	triggerSource, err := datadogV2.NewTriggerSourceFromValue(state.Selectors.TriggerSource.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid trigger source", err.Error())
		return
	}

	if *triggerSource == datadogV2.TRIGGERSOURCE_SECURITY_SIGNALS {
		_, err = r.api.DeleteSignalNotificationRule(r.auth, state.ID.ValueString())
	} else {
		_, err = r.api.DeleteVulnerabilityNotificationRule(r.auth, state.ID.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Error deleting notification rule", err.Error())
		return
	}
}

func (r *NotificationRuleResource) updateStateFromResponse(ctx context.Context, state *notificationRuleModel, response *datadogV2.NotificationRuleResponse) {
	state.ID = types.StringValue(response.Data.Id)

	// Only update the state if the description is not empty, or if it's not null in the plan
	// If the description is null in the TF config, it is omitted from the API call
	// The API returns an empty string, which, if put in the state, would result in a mismatch between state and config
	attributes := response.Data.Attributes
	state.Name = types.StringValue(attributes.GetName())

	selectors := attributes.Selectors
	state.Selectors.TriggerSource = types.StringValue(string(selectors.GetTriggerSource()))
	if ruleTypes, ok := selectors.GetRuleTypesOk(); ok || !state.Selectors.RuleTypes.IsNull() {
		state.Selectors.RuleTypes.ElementsAs(ctx, &ruleTypes, false)
	} else {
		state.Selectors.RuleTypes = types.ListNull(types.StringType)
	}
	if severities, ok := selectors.GetSeveritiesOk(); ok || !state.Selectors.Severities.IsNull() {
		state.Selectors.Severities.ElementsAs(ctx, &severities, false)
	} else {
		state.Selectors.Severities = types.ListNull(types.StringType)
	}
	if query, ok := selectors.GetQueryOk(); ok || !state.Selectors.Query.IsNull() {
		state.Selectors.Query = types.StringValue(*query)
	} else {
		state.Selectors.Query = types.StringNull()
	}

	if targets, ok := attributes.GetTargetsOk(); ok || !state.Targets.IsNull() {
		state.Targets.ElementsAs(ctx, &targets, false)
	}
	if field, ok := attributes.GetTimeAggregationOk(); ok {
		state.TimeAggregation = types.Int64Value(*field)
	} else {
		state.TimeAggregation = types.Int64Null()
	}

	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Version = types.Int64Value(attributes.GetVersion())
	state.CreatedAt = types.Int64Value(attributes.CreatedAt)
	state.CreatedByName = types.StringValue(*attributes.CreatedBy.Name)
	state.CreatedByHandle = types.StringValue(*attributes.CreatedBy.Handle)
	state.ModifiedAt = types.Int64Value(attributes.ModifiedAt)
	state.ModifiedByName = types.StringValue(*attributes.ModifiedBy.Name)
	state.ModifiedByHandle = types.StringValue(*attributes.ModifiedBy.Handle)
}
