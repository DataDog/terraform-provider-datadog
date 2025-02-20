package fwprovider

import (
	"context"
	"net/http"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &SecurityNotificationRuleResource{}
	_ resource.ResourceWithImportState = &SecurityNotificationRuleResource{}
)

type SecurityNotificationRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

var writeMutex = sync.Mutex{}

type securityNotificationRuleModel struct {
	ID               types.String    `tfsdk:"id"`
	Name             types.String    `tfsdk:"name"`
	Selectors        *selectorsModel `tfsdk:"selectors"`
	Targets          types.Set       `tfsdk:"targets"`
	TimeAggregation  types.Int64     `tfsdk:"time_aggregation"`
	Version          types.Int64     `tfsdk:"version"`
	Enabled          types.Bool      `tfsdk:"enabled"`
	CreatedAt        types.Int64     `tfsdk:"created_at"`
	ModifiedAt       types.Int64     `tfsdk:"modified_at"`
	CreatedByName    types.String    `tfsdk:"created_by_name"`
	CreatedByHandle  types.String    `tfsdk:"created_by_handle"`
	ModifiedByName   types.String    `tfsdk:"modified_by_name"`
	ModifiedByHandle types.String    `tfsdk:"modified_by_handle"`
}

type selectorsModel struct {
	TriggerSource types.String `tfsdk:"trigger_source"`
	RuleTypes     types.Set    `tfsdk:"rule_types"`
	Severities    types.Set    `tfsdk:"severities"`
	Query         types.String `tfsdk:"query"`
}

func (r *SecurityNotificationRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *SecurityNotificationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func NewSecurityNotificationRuleResource() resource.Resource {
	return &SecurityNotificationRuleResource{}
}

func (r *SecurityNotificationRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "security_notification_rule"
}

func (r *SecurityNotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Notification Rule API resource for creating and managing Datadog security notification rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the rule (must be unique).",
				Required:    true,
			},
			"targets": schema.SetAttribute{
				Description: "The list of handle targets for the notifications. A target must be prefixed with an @. It can be an email address (@bob@email.com), or any installed integration. For example, a Slack recipient (@slack-ops), or a Teams recipient (@teams-ops).",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.AtLeastOneOf(),
				},
			},
			"time_aggregation": schema.Int64Attribute{
				Description: "Specifies the time period, in seconds, used to aggregate the notification.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(int64(0)),
			},
			"version": schema.Int64Attribute{
				Description: "The rule version (incremented at each update).",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates whether the rule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"created_at": schema.Int64Attribute{
				Description: "Indicates when this rule was created.",
				Computed:    true,
			},
			"modified_at": schema.Int64Attribute{
				Description: "Indicates when this rule was last modified.",
				Computed:    true,
			},
			"created_by_name": schema.StringAttribute{
				Description: "The name of the rule creator.",
				Computed:    true,
			},
			"modified_by_name": schema.StringAttribute{
				Description: "The name of the rule last modifier.",
				Computed:    true,
			},
			"created_by_handle": schema.StringAttribute{
				Description: "The handle of the rule creator.",
				Computed:    true,
			},
			"modified_by_handle": schema.StringAttribute{
				Description: "The handle of the rule last modifier.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"selectors": schema.SingleNestedBlock{
				Description: "Defines selectors to filter security issues that generate notifications.",
				Attributes: map[string]schema.Attribute{
					"trigger_source": schema.StringAttribute{
						Description: "The type of security issues the rule applies to. Use `security_signals` for rules based on security signals and `security_findings` for those based on vulnerabilities.",
						Required:    true,
					},
					"rule_types": schema.SetAttribute{
						Description: "Specifies security rule types for filtering signals and vulnerabilities that generate notifications.",
						Required:    true,
						ElementType: types.StringType,
						Validators: []validator.Set{
							setvalidator.AtLeastOneOf(),
						},
					},
					"severities": schema.SetAttribute{
						Description: "The security rules severities to consider.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
					},
					"query": schema.StringAttribute{
						Description: "Comprises one or several key:value pairs for filtering security issues based on tags and attributes.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
		},
	}
}

func (r *SecurityNotificationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityNotificationRuleModel
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

func (r *SecurityNotificationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityNotificationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var response datadogV2.NotificationRuleResponse

	// Selectors can be null when terraform import is performed
	if state.Selectors != nil {
		triggerSource, err := datadogV2.NewTriggerSourceFromValue((*state.Selectors).TriggerSource.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid trigger source", err.Error())
			return
		}

		if *triggerSource == datadogV2.TRIGGERSOURCE_SECURITY_SIGNALS {
			response, _, err = r.api.GetSignalNotificationRule(r.auth, state.ID.ValueString())
		} else {
			response, _, err = r.api.GetVulnerabilityNotificationRule(r.auth, state.ID.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError("Error reading notification rule", err.Error())
			return
		}
	} else {
		// This path is reached when terraform import is performed
		// In this case we don't know the trigger source, so we try all of them
		var httpResponse *http.Response
		var err error

		response, httpResponse, err = r.api.GetSignalNotificationRule(r.auth, state.ID.ValueString())
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response, _, err = r.api.GetVulnerabilityNotificationRule(r.auth, state.ID.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError("Error reading notification rule", err.Error())
			return
		}
	}

	r.updateStateFromResponse(ctx, &state, &response)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SecurityNotificationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityNotificationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var previousState securityNotificationRuleModel
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

func (r *SecurityNotificationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityNotificationRuleModel
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

func (r *SecurityNotificationRuleResource) updateStateFromResponse(ctx context.Context, state *securityNotificationRuleModel, response *datadogV2.NotificationRuleResponse) {
	state.ID = types.StringValue(response.Data.Id)

	attributes := response.Data.Attributes

	if state.Name.IsNull() {
		// Empty state mean that we are in the terraform import flow
		// Let's initialize the state accordingly
		emptySet, _ := types.SetValue(types.StringType, []attr.Value{})
		state.Selectors = &selectorsModel{
			TriggerSource: types.StringValue(""),
			RuleTypes:     emptySet,
			Severities:    emptySet,
			Query:         types.StringValue(""),
		}
	}

	state.Name = types.StringValue(attributes.GetName())

	selectors := attributes.Selectors
	state.Selectors.TriggerSource = types.StringValue(string(selectors.GetTriggerSource()))
	if ruleTypes, ok := selectors.GetRuleTypesOk(); ok {
		state.Selectors.RuleTypes, _ = types.SetValueFrom(ctx, types.StringType, ruleTypes)
	} else {
		state.Selectors.RuleTypes = types.SetNull(types.StringType)
	}
	if severities, ok := selectors.GetSeveritiesOk(); ok {
		state.Selectors.Severities, _ = types.SetValueFrom(ctx, types.StringType, severities)
	} else {
		state.Selectors.Severities, _ = types.SetValue(types.StringType, []attr.Value{})
	}
	if query, ok := selectors.GetQueryOk(); ok {
		state.Selectors.Query = types.StringValue(*query)
	} else {
		state.Selectors.Query = types.StringValue("")
	}

	if targets, ok := attributes.GetTargetsOk(); ok {
		state.Targets, _ = types.SetValueFrom(ctx, types.StringType, targets)
	}
	if field, ok := attributes.GetTimeAggregationOk(); ok {
		state.TimeAggregation = types.Int64Value(*field)
	} else {
		state.TimeAggregation = types.Int64Value(int64(0))
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
