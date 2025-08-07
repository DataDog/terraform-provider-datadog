package fwprovider

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &monitorResource{}
	_ resource.ResourceWithImportState = &monitorResource{}
	_ resource.ResourceWithModifyPlan  = &monitorResource{}
)

var stringFloatValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`\d*(\.\d*)?`), "value must be a float")

type monitorResourceModel struct {
	ID                types.String                     `tfsdk:"id"`
	Name              types.String                     `tfsdk:"name"`
	Message           customtypes.TrimSpaceStringValue `tfsdk:"message"`
	EscalationMessage customtypes.TrimSpaceStringValue `tfsdk:"escalation_message"`
	Type              customtypes.MonitorTypeValue     `tfsdk:"type"`
	Query             customtypes.TrimSpaceStringValue `tfsdk:"query"`
	Priority          types.String                     `tfsdk:"priority"`
	Tags              types.Set                        `tfsdk:"tags"`
	MonitorThresholds []MonitorThreshold               `tfsdk:"monitor_thresholds"`
}

type MonitorThreshold struct {
	Ok               types.String `tfsdk:"ok"`
	Unknown          types.String `tfsdk:"unknown"`
	Warning          types.String `tfsdk:"warning"`
	WarningRecovery  types.String `tfsdk:"warning_recovery"`
	Critical         types.String `tfsdk:"critical"`
	CriticalRecovery types.String `tfsdk:"critical_recovery"`
}

type monitorResource struct {
	Api  *datadogV1.MonitorsApi
	Auth context.Context
}

func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

func (r *monitorResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMonitorsApiV1()
	r.Auth = providerData.Auth
}

func (r *monitorResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of Datadog monitor.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "A message to include with notifications for this monitor.\n\nEmail notifications can be sent to specific users by using the same `@username` notation as events.",
				Required:    true,
				CustomType:  customtypes.TrimSpaceStringType{},
			},
			"query": schema.StringAttribute{
				Description: "The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.\n\n**Note:** APM latency data is now available as Distribution Metrics. Existing monitors have been migrated automatically but all terraformed monitors can still use the existing metrics. We strongly recommend updating monitor definitions to query the new metrics. To learn more, or to see examples of how to update your terraform definitions to utilize the new distribution metrics, see the [detailed doc](https://docs.datadoghq.com/tracing/guide/ddsketch_trace_metrics/).",
				Required:    true,
				CustomType:  customtypes.TrimSpaceStringType{},
			},
			"type": schema.StringAttribute{
				Description: "The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(r.getAllowTypes()...),
				},
				// Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
				CustomType: customtypes.MonitorTypeType{},
			},
			"escalation_message": schema.StringAttribute{
				Description: "A message to include with a re-notification. Supports the `@username` notification allowed elsewhere.",
				Optional:    true,
				CustomType:  customtypes.TrimSpaceStringType{},
			},
			"priority": schema.StringAttribute{
				Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("0", "1", "2", "3", "4", "5"),
				},
			},
			"tags": schema.SetAttribute{
				Description: "A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"monitor_thresholds": schema.ListNestedBlock{
				Description: "Alert thresholds of the monitor.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ok": schema.StringAttribute{
							Description: "The monitor `OK` threshold. Only supported in monitor type `service check`. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional: true,
						},
						"warning": schema.StringAttribute{
							Description: "The monitor `WARNING` threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional: true,
						},
						"critical": schema.StringAttribute{
							Description: "The monitor `CRITICAL` threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional: true,
						},
						"unknown": schema.StringAttribute{
							Description: "The monitor `UNKNOWN` threshold. Only supported in monitor type `service check`. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional: true,
						},
						"warning_recovery": schema.StringAttribute{
							Description: "The monitor `WARNING` recovery threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional: true,
						},
						"critical_recovery": schema.StringAttribute{
							Description: "The monitor `CRITICAL` recovery threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// To be implemented
	// resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

func (r *monitorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, idErr := r.getMonitorId(&state, response.Diagnostics)
	if idErr != nil {
		return
	}
	resp, httpResp, err := r.Api.GetMonitor(r.Auth, *id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	monitorBody, _, diags := r.buildMonitorStruct(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateMonitor(r.Auth, *monitorBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, idErr := r.getMonitorId(&state, response.Diagnostics)
	if idErr != nil {
		return
	}
	_, updateRequestBody, diags := r.buildMonitorStruct(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateMonitor(r.Auth, *id, *updateRequestBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, idErr := r.getMonitorId(&state, response.Diagnostics)
	if idErr != nil {
		return
	}
	_, httpResp, err := r.Api.DeleteMonitor(r.Auth, *id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting monitor"))
		return
	}
}

func (r *monitorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the plan is null (resource is being destroyed) or no state exists yet, return early
	// as there's nothing to modify
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan, state monitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	m, _, diags := r.buildMonitorStruct(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isCreation := state.ID.IsNull()
	var httpresp *http.Response
	var err error
	log.Printf("[DEBUG] monitor/validate m=%#v", m)
	if !isCreation {
		id, idErr := r.getMonitorId(&state, resp.Diagnostics)
		if idErr != nil {
			return
		}
		_, httpresp, err = r.Api.ValidateExistingMonitor(r.Auth, *id, *m)
	} else {
		_, httpresp, err = r.Api.ValidateMonitor(r.Auth, *m)
	}
	if err != nil {
		if httpresp != nil && (httpresp.StatusCode == 502 || httpresp.StatusCode == 504) {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor, retrying"))
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor"))
	}
}

func (r *monitorResource) buildMonitorStruct(ctx context.Context, state *monitorResourceModel) (*datadogV1.Monitor, *datadogV1.MonitorUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	message := strings.TrimSpace(state.Message.ValueString())
	query := strings.TrimSpace(state.Query.ValueString())

	monitorType := datadogV1.MonitorType(state.Type.ValueString())
	m := datadogV1.NewMonitor(query, monitorType)
	m.SetName(state.Name.ValueString())
	m.SetMessage(message)

	u := datadogV1.NewMonitorUpdateRequest()
	u.SetType(monitorType)
	u.SetQuery(query)
	u.SetName(state.Name.ValueString())
	u.SetMessage(message)

	if v := r.parseInt(state.Priority); v != nil {
		m.SetPriority(*v)
		u.SetPriority(*v)
	} else {
		m.SetPriorityNil()
		u.SetPriorityNil()
	}

	var tags []string
	if !state.Tags.IsNull() && !state.Tags.IsUnknown() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		sort.Strings(tags)
	}
	m.SetTags(tags)
	u.SetTags(tags)

	monitorOptions := datadogV1.MonitorOptions{}
	if !state.EscalationMessage.IsNull() {
		escalationMessage := strings.TrimSpace(state.EscalationMessage.ValueString())
		monitorOptions.SetEscalationMessage(escalationMessage)
	}

	if state.MonitorThresholds != nil {
		thresholdObj := state.MonitorThresholds[0]
		var thresholds = datadogV1.MonitorThresholds{}
		if v := r.parseFloat(thresholdObj.Ok); v != nil {
			thresholds.SetOk(*v)
		}
		if v := r.parseFloat(thresholdObj.Unknown); v != nil {
			thresholds.SetUnknown(*v)
		}
		if v := r.parseFloat(thresholdObj.Critical); v != nil {
			thresholds.SetCritical(*v)
		}
		if v := r.parseFloat(thresholdObj.CriticalRecovery); v != nil {
			thresholds.SetCriticalRecovery(*v)
		}
		if v := r.parseFloat(thresholdObj.Warning); v != nil {
			thresholds.SetWarning(*v)
		}
		if v := r.parseFloat(thresholdObj.WarningRecovery); v != nil {
			thresholds.SetWarningRecovery(*v)
		}
		monitorOptions.SetThresholds(thresholds)
	}
	m.SetOptions(monitorOptions)
	u.SetOptions(monitorOptions)

	return m, u, diags
}

func (r *monitorResource) updateState(ctx context.Context, state *monitorResourceModel, m *datadogV1.Monitor) {
	if id, ok := m.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(strconv.FormatInt(*id, 10))
	}

	if name, ok := m.GetNameOk(); ok && name != nil {
		state.Name = types.StringValue(*name)
	}

	if message, ok := m.GetMessageOk(); ok && message != nil {
		state.Message = customtypes.TrimSpaceStringValue{
			StringValue: types.StringValue(*message),
		}
	}

	if mType, ok := m.GetTypeOk(); ok && mType != nil {
		state.Type = customtypes.MonitorTypeValue{
			StringValue: types.StringValue(string(*mType)),
		}
	}

	if query, ok := m.GetQueryOk(); ok && query != nil {
		state.Query = customtypes.TrimSpaceStringValue{
			StringValue: types.StringValue(*query),
		}
	}

	if priority, ok := m.GetPriorityOk(); ok && priority != nil {
		state.Priority = types.StringValue(strconv.FormatInt(*priority, 10))
	}
	if tags, ok := m.GetTagsOk(); ok && tags != nil {
		state.Tags, _ = types.SetValueFrom(ctx, types.StringType, tags)
	}

	if escalationMessage, ok := m.Options.GetEscalationMessageOk(); ok && escalationMessage != nil {
		state.EscalationMessage = customtypes.TrimSpaceStringValue{
			StringValue: types.StringValue(*escalationMessage),
		}
	}
	if monitorThresholds, ok := m.Options.GetThresholdsOk(); ok && monitorThresholds != nil {
		state.MonitorThresholds = []MonitorThreshold{{
			Ok:               r.strOrNull(monitorThresholds.Ok.Get()),
			Unknown:          r.strOrNull(monitorThresholds.Unknown.Get()),
			Warning:          r.strOrNull(monitorThresholds.Warning.Get()),
			WarningRecovery:  r.strOrNull(monitorThresholds.WarningRecovery.Get()),
			Critical:         r.strOrNull(monitorThresholds.Critical),
			CriticalRecovery: r.strOrNull(monitorThresholds.CriticalRecovery.Get()),
		}}
	}
}

func (r *monitorResource) getMonitorId(state *monitorResourceModel, diags diag.Diagnostics) (*int64, error) {
	stateId := state.ID.ValueString()
	id, err := strconv.ParseInt(stateId, 10, 64)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error on monitor id"))
		return nil, err
	}
	return &id, nil
}

func (r *monitorResource) getAllowTypes() []string {
	allowed := (*datadogV1.MonitorType)(nil).GetAllowedValues()
	strVals := make([]string, len(allowed))
	for i, v := range allowed {
		strVals[i] = string(v)
	}
	return strVals
}

func (r *monitorResource) parseInt(v types.String) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	result, err := strconv.ParseInt(v.ValueString(), 10, 64)
	if err != nil {
		return nil
	}
	return &result
}

func (r *monitorResource) parseFloat(v types.String) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	result, err := json.Number(v.ValueString()).Float64()
	if err != nil {
		return nil
	}
	return &result
}

func (r *monitorResource) strOrNull(v any) types.String {
	switch t := v.(type) {
	case *float64:
		if t == nil {
			return types.StringNull()
		}
		return types.StringValue(strconv.FormatFloat(*t, 'f', -1, 64))
	case *datadog.NullableFloat64:
		if !t.IsSet() || t.Get() == nil {
			return types.StringNull()
		}
		return r.strOrNull(t.Get())
	}
	return types.StringNull()
}
