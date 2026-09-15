package fwprovider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFleetScheduleResourceSchemaConstraints(t *testing.T) {
	t.Parallel()

	var response resource.SchemaResponse
	NewFleetScheduleResource().Schema(context.Background(), resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", response.Diagnostics)
	}

	name := response.Schema.Attributes["name"].(resourceschema.StringAttribute)
	if !name.Required || len(name.Validators) == 0 {
		t.Fatal("name must be required and validated")
	}
	var nameValidation validator.StringResponse
	name.Validators[0].ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue("")}, &nameValidation)
	if !nameValidation.Diagnostics.HasError() {
		t.Fatal("empty name must be rejected")
	}

	status := response.Schema.Attributes["status"].(resourceschema.StringAttribute)
	if !status.Optional || !status.Computed {
		t.Fatal("status must be optional and computed")
	}
	var statusValidation validator.StringResponse
	status.Validators[0].ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue("paused")}, &statusValidation)
	if !statusValidation.Diagnostics.HasError() {
		t.Fatal("invalid status must be rejected")
	}

	version := response.Schema.Attributes["version_to_latest"].(resourceschema.Int64Attribute)
	var versionValidation validator.Int64Response
	version.Validators[0].ValidateInt64(context.Background(), validator.Int64Request{ConfigValue: types.Int64Value(3)}, &versionValidation)
	if !versionValidation.Diagnostics.HasError() {
		t.Fatal("version_to_latest greater than 2 must be rejected")
	}

	rule := response.Schema.Attributes["rule"].(resourceschema.SingleNestedAttribute)
	if !rule.Required {
		t.Fatal("rule must be required")
	}
	duration := rule.Attributes["maintenance_window_duration"].(resourceschema.Int64Attribute)
	var durationValidation validator.Int64Response
	duration.Validators[0].ValidateInt64(context.Background(), validator.Int64Request{ConfigValue: types.Int64Value(0)}, &durationValidation)
	if !durationValidation.Diagnostics.HasError() {
		t.Fatal("zero maintenance window duration must be rejected")
	}
	start := rule.Attributes["start_maintenance_window"].(resourceschema.StringAttribute)
	var startValidation validator.StringResponse
	start.Validators[0].ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue("9:00")}, &startValidation)
	if !startValidation.Diagnostics.HasError() {
		t.Fatal("non-canonical maintenance window time must be rejected")
	}
}

func TestBuildFleetScheduleCreateRequest(t *testing.T) {
	t.Parallel()
	model := testFleetScheduleResourceModel(t)
	body, diags := buildFleetScheduleCreateRequest(context.Background(), &model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	attributes := body.Data.Attributes
	if attributes.Name != "test schedule" || attributes.Query != "env:nonprod" {
		t.Fatalf("unexpected core attributes: %#v", attributes)
	}
	if got, want := attributes.Rule.DaysOfWeek, []string{"Mon", "Wed"}; !equalStrings(got, want) {
		t.Fatalf("days = %v, want %v", got, want)
	}
	if !attributes.HasStatus() || attributes.GetStatus() != datadogV2.FLEETSCHEDULESTATUS_INACTIVE {
		t.Fatalf("status = %q", attributes.GetStatus())
	}
	if !attributes.HasVersionToLatest() || attributes.GetVersionToLatest() != 1 {
		t.Fatalf("version_to_latest = %d", attributes.GetVersionToLatest())
	}
}

func TestBuildFleetScheduleCreateReconciliation(t *testing.T) {
	t.Parallel()
	model := testFleetScheduleResourceModel(t)

	body, changed := buildFleetScheduleCreateReconciliation(&model)
	if !changed {
		t.Fatal("active create response must be reconciled with configured inactive status")
	}
	attributes := body.Data.GetAttributes()
	if !attributes.HasStatus() || attributes.GetStatus() != datadogV2.FLEETSCHEDULESTATUS_INACTIVE {
		t.Fatalf("status reconciliation = %#v", attributes)
	}
	if !attributes.HasVersionToLatest() || attributes.GetVersionToLatest() != 1 {
		t.Fatal("explicit version_to_latest must be reconciled")
	}
}

func TestBuildFleetSchedulePatchRequestIsPartial(t *testing.T) {
	t.Parallel()
	state := testFleetScheduleResourceModel(t)
	plan := state
	plan.Name = types.StringValue("renamed")

	body, changed, diags := buildFleetSchedulePatchRequest(context.Background(), &state, &plan)
	if diags.HasError() || !changed {
		t.Fatalf("changed = %v, diagnostics = %v", changed, diags)
	}
	attributes := body.Data.GetAttributes()
	if !attributes.HasName() || attributes.GetName() != "renamed" {
		t.Fatalf("name was not included: %#v", attributes)
	}
	if attributes.HasQuery() || attributes.HasRule() || attributes.HasStatus() || attributes.HasVersionToLatest() {
		t.Fatalf("unchanged fields were included: %#v", attributes)
	}
}

func TestBuildFleetSchedulePatchRequestIncludesFullRuleWhenRuleChanges(t *testing.T) {
	t.Parallel()
	state := testFleetScheduleResourceModel(t)
	plan := testFleetScheduleResourceModel(t)
	plan.Rule.MaintenanceWindowDuration = types.Int64Value(90)

	body, changed, diags := buildFleetSchedulePatchRequest(context.Background(), &state, &plan)
	if diags.HasError() || !changed {
		t.Fatalf("changed = %v, diagnostics = %v", changed, diags)
	}
	attributes := body.Data.GetAttributes()
	if !attributes.HasRule() {
		t.Fatal("changed rule was not included")
	}
	rule := attributes.GetRule()
	if rule.MaintenanceWindowDuration != 90 || rule.StartMaintenanceWindow != "03:15" || rule.Timezone != "UTC" {
		t.Fatalf("partial recurrence rule sent: %#v", rule)
	}
}

func TestFleetScheduleResponseMapping(t *testing.T) {
	t.Parallel()
	nextRun := "2026-09-16T03:15:00Z"
	schedule := completeFleetScheduleV2("schedule-b")
	schedule.Attributes.NextRun = &nextRun

	model, diags := fleetScheduleDataSourceModelFromAPI(context.Background(), schedule)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if model.Rule == nil || model.Rule.StartMaintenanceWindow.ValueString() != "03:15" {
		t.Fatalf("start time was not normalized: %#v", model.Rule)
	}
	if model.Rule.Interval.ValueInt64() != 2 {
		t.Fatalf("interval = %d", model.Rule.Interval.ValueInt64())
	}
	if model.NextRun.ValueString() != nextRun {
		t.Fatalf("next_run = %q", model.NextRun.ValueString())
	}
	if model.NotificationRule == nil || model.NotificationRule.Handles.IsNull() || model.NotificationRule.Tags.IsNull() {
		t.Fatalf("notification rule was not mapped: %#v", model.NotificationRule)
	}
}

func TestFleetScheduleResponseMappingNullableValues(t *testing.T) {
	t.Parallel()
	schedule := completeFleetScheduleV2("schedule-a")
	schedule.Attributes.NextRun = nil
	schedule.Attributes.NotificationRule = nil

	model, diags := fleetScheduleDataSourceModelFromAPI(context.Background(), schedule)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !model.NextRun.IsNull() {
		t.Fatalf("next_run = %v, want null", model.NextRun)
	}
	if model.NotificationRule != nil {
		t.Fatalf("notification_rule = %#v, want nil", model.NotificationRule)
	}
}

func TestFleetSchedulesResponseMappingSortsAndAllowsEmpty(t *testing.T) {
	t.Parallel()

	response := datadogV2.FleetSchedulesV2Response{Data: []datadogV2.FleetScheduleV2{
		completeFleetScheduleV2("schedule-z"),
		completeFleetScheduleV2("schedule-a"),
	}}
	model, diags := fleetSchedulesDataSourceModelFromAPI(context.Background(), response)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := []string{model.Schedules[0].ID.ValueString(), model.Schedules[1].ID.ValueString()}; !equalStrings(got, []string{"schedule-a", "schedule-z"}) {
		t.Fatalf("IDs = %v", got)
	}
	if model.TotalCount.ValueInt64() != 2 {
		t.Fatalf("total_count = %d", model.TotalCount.ValueInt64())
	}

	empty, diags := fleetSchedulesDataSourceModelFromAPI(context.Background(), datadogV2.FleetSchedulesV2Response{Data: []datadogV2.FleetScheduleV2{}})
	if diags.HasError() || len(empty.Schedules) != 0 || empty.TotalCount.ValueInt64() != 0 {
		t.Fatalf("empty response mapping = %#v, diagnostics = %v", empty, diags)
	}
}

func TestNormalizeFleetScheduleTime(t *testing.T) {
	t.Parallel()
	for input, want := range map[string]string{"0315": "03:15", "03:15": "03:15", "abcd": "abcd", "bad": "bad"} {
		if got := normalizeFleetScheduleTime(input); got != want {
			t.Errorf("normalizeFleetScheduleTime(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFleetScheduleResourceLifecycleUsesStableReads(t *testing.T) {
	const scheduleID = "schedule-test-id"
	ctx := context.Background()
	updated := false
	requests := make([]string, 0, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		requests = append(requests, request.Method+" "+request.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/unstable/fleet/schedules":
			fmt.Fprintf(w, `{"data":{"id":%q,"type":"schedule","attributes":{"name":"test schedule","query":"env:nonprod","status":"inactive","version_to_latest":1,"rule":{"days_of_week":["Mon","Wed"],"maintenance_window_duration":60,"start_maintenance_window":"03:15","timezone":"UTC"}}}}`, scheduleID)
		case request.Method == http.MethodPatch && request.URL.Path == "/api/unstable/fleet/schedules/"+scheduleID:
			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("reading PATCH body: %v", err)
			}
			if strings.Contains(string(body), `"name"`) {
				for _, expected := range []string{`"name":"updated schedule"`, `"query":"env:test"`, `"version_to_latest":2`, `"maintenance_window_duration":90`} {
					if !strings.Contains(string(body), expected) {
						t.Errorf("PATCH body %s does not contain %s", body, expected)
					}
				}
				updated = true
			} else if !strings.Contains(string(body), `"status":"inactive"`) || !strings.Contains(string(body), `"version_to_latest":1`) {
				t.Errorf("create reconciliation PATCH has unexpected body %s", body)
			}
			fmt.Fprintf(w, `{"data":{"id":%q,"type":"schedule","attributes":{}}}`, scheduleID)
		case request.Method == http.MethodGet && request.URL.Path == "/api/v2/fleet/schedules/"+scheduleID:
			name, query, duration, start, version := "test schedule", "env:nonprod", 60, "0315", 1
			if updated {
				name, query, duration, start, version = "updated schedule", "env:test", 90, "0445", 2
			}
			fmt.Fprintf(w, `{"data":{"id":%q,"type":"schedule","attributes":{"name":%q,"query":%q,"status":"inactive","version_to_latest":%d,"rule":{"days_of_week":["Mon","Wed"],"interval":1,"maintenance_window_duration":%d,"start_maintenance_window":%q,"timezone":"UTC"}}}}`, scheduleID, name, query, version, duration, start)
		case request.Method == http.MethodDelete && request.URL.Path == "/api/unstable/fleet/schedules/"+scheduleID:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	clientConfig := datadog.NewConfiguration()
	clientConfig.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	clientConfig.OperationServers = nil
	clientConfig.HTTPClient = server.Client()
	for _, operation := range []string{"v2.CreateFleetSchedule", "v2.UpdateFleetSchedule", "v2.DeleteFleetSchedule"} {
		clientConfig.SetUnstableOperationEnabled(operation, true)
	}
	r := &fleetScheduleResource{Api: datadogV2.NewFleetAutomationApi(datadog.NewAPIClient(clientConfig)), Auth: ctx}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	createModel := testFleetScheduleResourceModel(t)
	createModel.ID = types.StringUnknown()
	createPlan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := createPlan.Set(ctx, &createModel); diags.HasError() {
		t.Fatalf("building create plan: %v", diags)
	}
	createResponse := resource.CreateResponse{State: tfsdk.State{Raw: createPlan.Raw, Schema: schemaResponse.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: createPlan}, &createResponse)
	if createResponse.Diagnostics.HasError() {
		t.Fatalf("creating schedule: %v", createResponse.Diagnostics)
	}
	var created fleetScheduleResourceModel
	if diags := createResponse.State.Get(ctx, &created); diags.HasError() {
		t.Fatalf("reading create state: %v", diags)
	}
	if created.ID.ValueString() != scheduleID || created.Rule.StartMaintenanceWindow.ValueString() != "03:15" {
		t.Fatalf("unexpected create state: %#v", created)
	}

	updateModel := testFleetScheduleResourceModel(t)
	updateModel.ID = types.StringValue(scheduleID)
	updateModel.Name = types.StringValue("updated schedule")
	updateModel.Query = types.StringValue("env:test")
	updateModel.VersionToLatest = types.Int64Value(2)
	updateModel.Rule.MaintenanceWindowDuration = types.Int64Value(90)
	updateModel.Rule.StartMaintenanceWindow = types.StringValue("04:45")
	updatePlan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := updatePlan.Set(ctx, &updateModel); diags.HasError() {
		t.Fatalf("building update plan: %v", diags)
	}
	updateResponse := resource.UpdateResponse{State: tfsdk.State{Raw: updatePlan.Raw, Schema: schemaResponse.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: updatePlan, State: createResponse.State}, &updateResponse)
	if updateResponse.Diagnostics.HasError() {
		t.Fatalf("updating schedule: %v", updateResponse.Diagnostics)
	}
	var updatedState fleetScheduleResourceModel
	if diags := updateResponse.State.Get(ctx, &updatedState); diags.HasError() {
		t.Fatalf("reading update state: %v", diags)
	}
	if updatedState.ID.ValueString() != scheduleID || updatedState.Name.ValueString() != "updated schedule" || updatedState.Rule.StartMaintenanceWindow.ValueString() != "04:45" {
		t.Fatalf("unexpected update state: %#v", updatedState)
	}

	deleteResponse := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: updateResponse.State}, &deleteResponse)
	if deleteResponse.Diagnostics.HasError() {
		t.Fatalf("deleting schedule: %v", deleteResponse.Diagnostics)
	}
	wantRequests := []string{
		"POST /api/unstable/fleet/schedules",
		"PATCH /api/unstable/fleet/schedules/" + scheduleID,
		"GET /api/v2/fleet/schedules/" + scheduleID,
		"PATCH /api/unstable/fleet/schedules/" + scheduleID,
		"GET /api/v2/fleet/schedules/" + scheduleID,
		"DELETE /api/unstable/fleet/schedules/" + scheduleID,
	}
	if !equalStrings(requests, wantRequests) {
		t.Fatalf("requests = %v, want %v", requests, wantRequests)
	}
}

func TestFleetScheduleResourceNotFoundHandling(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()
	clientConfig := datadog.NewConfiguration()
	clientConfig.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	clientConfig.OperationServers = nil
	clientConfig.HTTPClient = server.Client()
	clientConfig.SetUnstableOperationEnabled("v2.DeleteFleetSchedule", true)
	r := &fleetScheduleResource{Api: datadogV2.NewFleetAutomationApi(datadog.NewAPIClient(clientConfig)), Auth: ctx}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	model := testFleetScheduleResourceModel(t)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("building state: %v", diags)
	}

	readResponse := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResponse)
	if readResponse.Diagnostics.HasError() || !readResponse.State.Raw.IsNull() {
		t.Fatalf("read 404 should remove state; state=%v diagnostics=%v", readResponse.State.Raw, readResponse.Diagnostics)
	}
	deleteResponse := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: state}, &deleteResponse)
	if deleteResponse.Diagnostics.HasError() {
		t.Fatalf("delete 404 should succeed: %v", deleteResponse.Diagnostics)
	}
}

func testFleetScheduleResourceModel(t *testing.T) fleetScheduleResourceModel {
	t.Helper()
	days, diags := types.SetValueFrom(context.Background(), types.StringType, []string{"Wed", "Mon"})
	if diags.HasError() {
		t.Fatalf("creating days set: %v", diags)
	}
	return fleetScheduleResourceModel{
		ID:              types.StringValue("schedule-id"),
		Name:            types.StringValue("test schedule"),
		Query:           types.StringValue("env:nonprod"),
		Status:          types.StringValue("inactive"),
		VersionToLatest: types.Int64Value(1),
		Rule: &fleetScheduleRuleModel{
			DaysOfWeek:                days,
			MaintenanceWindowDuration: types.Int64Value(60),
			StartMaintenanceWindow:    types.StringValue("03:15"),
			Timezone:                  types.StringValue("UTC"),
		},
	}
}

func completeFleetScheduleV2(id string) datadogV2.FleetScheduleV2 {
	name := "test"
	query := "env:nonprod"
	status := datadogV2.FLEETSCHEDULESTATUS_INACTIVE
	version := int64(1)
	interval := int64(2)
	duration := int64(60)
	start := "0315"
	timezone := "UTC"
	createdAt := "2026-09-15T00:00:00Z"
	createdBy := "creator@example.com"
	updatedAt := "2026-09-15T01:00:00Z"
	updatedBy := "updater@example.com"
	isDefault := false
	return datadogV2.FleetScheduleV2{
		Id: id,
		Attributes: datadogV2.FleetScheduleV2Attributes{
			Name:            &name,
			Query:           &query,
			Status:          &status,
			VersionToLatest: &version,
			CreatedAt:       &createdAt,
			CreatedBy:       &createdBy,
			UpdatedAt:       &updatedAt,
			UpdatedBy:       &updatedBy,
			IsDefault:       &isDefault,
			Rule: &datadogV2.FleetScheduleV2RecurrenceRule{
				DaysOfWeek:                []string{"Wed", "Mon"},
				Interval:                  &interval,
				MaintenanceWindowDuration: &duration,
				StartMaintenanceWindow:    &start,
				Timezone:                  &timezone,
			},
			NotificationRule: &datadogV2.FleetScheduleV2NotificationRule{
				Handles: []string{"@slack-example"},
				Tags:    []string{"team:fleet"},
			},
		},
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
