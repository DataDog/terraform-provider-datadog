package fwprovider

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fleetSchedulesAPI interface {
	CreateFleetSchedule(context.Context, datadogV2.FleetScheduleCreateRequest) (datadogV2.FleetScheduleResponse, *http.Response, error)
	DeleteFleetSchedule(context.Context, string) (*http.Response, error)
	GetFleetScheduleV2(context.Context, string) (datadogV2.FleetScheduleV2Response, *http.Response, error)
	ListFleetSchedulesV2(context.Context) (datadogV2.FleetSchedulesV2Response, *http.Response, error)
	UpdateFleetSchedule(context.Context, string, datadogV2.FleetSchedulePatchRequest) (datadogV2.FleetScheduleResponse, *http.Response, error)
}

type fleetScheduleRuleModel struct {
	DaysOfWeek                types.Set    `tfsdk:"days_of_week"`
	MaintenanceWindowDuration types.Int64  `tfsdk:"maintenance_window_duration"`
	StartMaintenanceWindow    types.String `tfsdk:"start_maintenance_window"`
	Timezone                  types.String `tfsdk:"timezone"`
}

type fleetScheduleResourceModel struct {
	ID              types.String            `tfsdk:"id"`
	Name            types.String            `tfsdk:"name"`
	Query           types.String            `tfsdk:"query"`
	Status          types.String            `tfsdk:"status"`
	VersionToLatest types.Int64             `tfsdk:"version_to_latest"`
	Rule            *fleetScheduleRuleModel `tfsdk:"rule"`
}

type fleetScheduleDataSourceRuleModel struct {
	DaysOfWeek                types.Set    `tfsdk:"days_of_week"`
	Interval                  types.Int64  `tfsdk:"interval"`
	MaintenanceWindowDuration types.Int64  `tfsdk:"maintenance_window_duration"`
	StartMaintenanceWindow    types.String `tfsdk:"start_maintenance_window"`
	Timezone                  types.String `tfsdk:"timezone"`
}

type fleetScheduleNotificationRuleModel struct {
	Handles types.Set `tfsdk:"handles"`
	Tags    types.Set `tfsdk:"tags"`
}

type fleetScheduleDataSourceModel struct {
	ID               types.String                        `tfsdk:"id"`
	Name             types.String                        `tfsdk:"name"`
	Query            types.String                        `tfsdk:"query"`
	Status           types.String                        `tfsdk:"status"`
	VersionToLatest  types.Int64                         `tfsdk:"version_to_latest"`
	Rule             *fleetScheduleDataSourceRuleModel   `tfsdk:"rule"`
	CreatedAt        types.String                        `tfsdk:"created_at"`
	CreatedBy        types.String                        `tfsdk:"created_by"`
	UpdatedAt        types.String                        `tfsdk:"updated_at"`
	UpdatedBy        types.String                        `tfsdk:"updated_by"`
	IsDefault        types.Bool                          `tfsdk:"is_default"`
	NextRun          types.String                        `tfsdk:"next_run"`
	NotificationRule *fleetScheduleNotificationRuleModel `tfsdk:"notification_rule"`
}

func normalizeFleetScheduleTime(value string) string {
	if len(value) == 4 && !strings.Contains(value, ":") &&
		value[0] >= '0' && value[0] <= '9' &&
		value[1] >= '0' && value[1] <= '9' &&
		value[2] >= '0' && value[2] <= '9' &&
		value[3] >= '0' && value[3] <= '9' {
		return value[:2] + ":" + value[2:]
	}
	return value
}

func fleetScheduleRuleFromModel(ctx context.Context, model *fleetScheduleRuleModel) (datadogV2.FleetScheduleRecurrenceRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if model == nil {
		diags.AddError("Missing fleet schedule rule", "A fleet schedule rule is required.")
		return datadogV2.FleetScheduleRecurrenceRule{}, diags
	}

	var days []string
	diags.Append(model.DaysOfWeek.ElementsAs(ctx, &days, false)...)
	sort.Strings(days)
	return datadogV2.FleetScheduleRecurrenceRule{
		DaysOfWeek:                days,
		MaintenanceWindowDuration: model.MaintenanceWindowDuration.ValueInt64(),
		StartMaintenanceWindow:    model.StartMaintenanceWindow.ValueString(),
		Timezone:                  model.Timezone.ValueString(),
	}, diags
}

func buildFleetScheduleCreateRequest(ctx context.Context, model *fleetScheduleResourceModel) (datadogV2.FleetScheduleCreateRequest, diag.Diagnostics) {
	rule, diags := fleetScheduleRuleFromModel(ctx, model.Rule)
	attributes := datadogV2.FleetScheduleCreateAttributes{
		Name:  model.Name.ValueString(),
		Query: model.Query.ValueString(),
		Rule:  rule,
	}
	if !model.Status.IsNull() && !model.Status.IsUnknown() {
		attributes.SetStatus(datadogV2.FleetScheduleStatus(model.Status.ValueString()))
	}
	if !model.VersionToLatest.IsNull() && !model.VersionToLatest.IsUnknown() {
		attributes.SetVersionToLatest(model.VersionToLatest.ValueInt64())
	}

	return datadogV2.FleetScheduleCreateRequest{
		Data: datadogV2.FleetScheduleCreate{
			Type:       datadogV2.FLEETSCHEDULERESOURCETYPE_SCHEDULE,
			Attributes: attributes,
		},
	}, diags
}

func buildFleetScheduleCreateReconciliation(config *fleetScheduleResourceModel) (datadogV2.FleetSchedulePatchRequest, bool) {
	attributes := datadogV2.FleetSchedulePatchAttributes{}
	changed := false

	if !config.Status.IsNull() && !config.Status.IsUnknown() {
		attributes.SetStatus(datadogV2.FleetScheduleStatus(config.Status.ValueString()))
		changed = true
	}
	if !config.VersionToLatest.IsNull() && !config.VersionToLatest.IsUnknown() {
		attributes.SetVersionToLatest(config.VersionToLatest.ValueInt64())
		changed = true
	}

	return datadogV2.FleetSchedulePatchRequest{
		Data: datadogV2.FleetSchedulePatch{
			Type:       datadogV2.FLEETSCHEDULERESOURCETYPE_SCHEDULE,
			Attributes: &attributes,
		},
	}, changed
}

func fleetScheduleRuleModelsEqual(a, b *fleetScheduleRuleModel) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.DaysOfWeek.Equal(b.DaysOfWeek) &&
		a.MaintenanceWindowDuration.Equal(b.MaintenanceWindowDuration) &&
		a.StartMaintenanceWindow.Equal(b.StartMaintenanceWindow) &&
		a.Timezone.Equal(b.Timezone)
}

func buildFleetSchedulePatchRequest(ctx context.Context, state, plan *fleetScheduleResourceModel) (datadogV2.FleetSchedulePatchRequest, bool, diag.Diagnostics) {
	attributes := datadogV2.FleetSchedulePatchAttributes{}
	var diags diag.Diagnostics
	changed := false

	if !state.Name.Equal(plan.Name) {
		attributes.SetName(plan.Name.ValueString())
		changed = true
	}
	if !state.Query.Equal(plan.Query) {
		attributes.SetQuery(plan.Query.ValueString())
		changed = true
	}
	if !state.Status.Equal(plan.Status) && !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		attributes.SetStatus(datadogV2.FleetScheduleStatus(plan.Status.ValueString()))
		changed = true
	}
	if !state.VersionToLatest.Equal(plan.VersionToLatest) && !plan.VersionToLatest.IsNull() && !plan.VersionToLatest.IsUnknown() {
		attributes.SetVersionToLatest(plan.VersionToLatest.ValueInt64())
		changed = true
	}
	if !fleetScheduleRuleModelsEqual(state.Rule, plan.Rule) {
		rule, ruleDiags := fleetScheduleRuleFromModel(ctx, plan.Rule)
		diags.Append(ruleDiags...)
		attributes.SetRule(rule)
		changed = true
	}

	return datadogV2.FleetSchedulePatchRequest{
		Data: datadogV2.FleetSchedulePatch{
			Type:       datadogV2.FLEETSCHEDULERESOURCETYPE_SCHEDULE,
			Attributes: &attributes,
		},
	}, changed, diags
}

func updateFleetScheduleResourceModel(ctx context.Context, model *fleetScheduleResourceModel, schedule datadogV2.FleetScheduleV2) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := schedule.GetAttributes()
	model.ID = types.StringValue(schedule.GetId())
	model.Name = stringValueOrNull(attributes.Name)
	model.Query = stringValueOrNull(attributes.Query)
	if attributes.Status != nil {
		model.Status = types.StringValue(string(*attributes.Status))
	} else {
		model.Status = types.StringNull()
	}
	model.VersionToLatest = int64ValueOrNull(attributes.VersionToLatest)

	if attributes.Rule == nil {
		model.Rule = nil
		return diags
	}
	rule := attributes.Rule
	days, setDiags := types.SetValueFrom(ctx, types.StringType, rule.DaysOfWeek)
	diags.Append(setDiags...)
	model.Rule = &fleetScheduleRuleModel{
		DaysOfWeek:                days,
		MaintenanceWindowDuration: int64ValueOrNull(rule.MaintenanceWindowDuration),
		StartMaintenanceWindow:    normalizedStringValueOrNull(rule.StartMaintenanceWindow),
		Timezone:                  stringValueOrNull(rule.Timezone),
	}
	return diags
}

func fleetScheduleDataSourceModelFromAPI(ctx context.Context, schedule datadogV2.FleetScheduleV2) (fleetScheduleDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := schedule.GetAttributes()
	model := fleetScheduleDataSourceModel{
		ID:              types.StringValue(schedule.GetId()),
		Name:            stringValueOrNull(attributes.Name),
		Query:           stringValueOrNull(attributes.Query),
		VersionToLatest: int64ValueOrNull(attributes.VersionToLatest),
		CreatedAt:       stringValueOrNull(attributes.CreatedAt),
		CreatedBy:       stringValueOrNull(attributes.CreatedBy),
		UpdatedAt:       stringValueOrNull(attributes.UpdatedAt),
		UpdatedBy:       stringValueOrNull(attributes.UpdatedBy),
		IsDefault:       boolValueOrNull(attributes.IsDefault),
		NextRun:         stringValueOrNull(attributes.NextRun),
	}
	if attributes.Status != nil {
		model.Status = types.StringValue(string(*attributes.Status))
	} else {
		model.Status = types.StringNull()
	}

	if attributes.Rule != nil {
		rule := attributes.Rule
		days, setDiags := types.SetValueFrom(ctx, types.StringType, rule.DaysOfWeek)
		diags.Append(setDiags...)
		model.Rule = &fleetScheduleDataSourceRuleModel{
			DaysOfWeek:                days,
			Interval:                  int64ValueOrNull(rule.Interval),
			MaintenanceWindowDuration: int64ValueOrNull(rule.MaintenanceWindowDuration),
			StartMaintenanceWindow:    normalizedStringValueOrNull(rule.StartMaintenanceWindow),
			Timezone:                  stringValueOrNull(rule.Timezone),
		}
	}
	if attributes.NotificationRule != nil {
		handles, handlesDiags := types.SetValueFrom(ctx, types.StringType, attributes.NotificationRule.Handles)
		tags, tagsDiags := types.SetValueFrom(ctx, types.StringType, attributes.NotificationRule.Tags)
		diags.Append(handlesDiags...)
		diags.Append(tagsDiags...)
		model.NotificationRule = &fleetScheduleNotificationRuleModel{Handles: handles, Tags: tags}
	}
	return model, diags
}

func stringValueOrNull(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func normalizedStringValueOrNull(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(normalizeFleetScheduleTime(*value))
}

func int64ValueOrNull(value *int64) types.Int64 {
	if value == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*value)
}

func boolValueOrNull(value *bool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*value)
}
