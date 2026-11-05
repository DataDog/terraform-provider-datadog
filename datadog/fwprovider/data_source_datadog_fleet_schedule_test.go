package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFleetScheduleDataSourcesReadStableAPI(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/api/v2/fleet/schedules/schedule-b":
			fmt.Fprint(w, fleetScheduleDataSourceResponseJSON("schedule-b"))
		case "/api/v2/fleet/schedules":
			fmt.Fprintf(w, `{"data":[%s,%s],"meta":{"page":{"total_count":2}}}`, fleetScheduleDataJSON("schedule-z"), fleetScheduleDataJSON("schedule-a"))
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	clientConfig := datadog.NewConfiguration()
	clientConfig.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	clientConfig.OperationServers = nil
	clientConfig.HTTPClient = server.Client()
	api := datadogV2.NewFleetAutomationApi(datadog.NewAPIClient(clientConfig))

	singular := &fleetScheduleDataSource{Api: api, Auth: ctx}
	var singularSchema datasource.SchemaResponse
	singular.Schema(ctx, datasource.SchemaRequest{}, &singularSchema)
	configPlan := tfsdk.Plan{Schema: singularSchema.Schema}
	if diags := configPlan.Set(ctx, &fleetScheduleDataSourceModel{ID: types.StringValue("schedule-b")}); diags.HasError() {
		t.Fatalf("building singular config: %v", diags)
	}
	config := tfsdk.Config{Raw: configPlan.Raw, Schema: singularSchema.Schema}
	singularResponse := datasource.ReadResponse{State: tfsdk.State{Schema: singularSchema.Schema}}
	singular.Read(ctx, datasource.ReadRequest{Config: config}, &singularResponse)
	if singularResponse.Diagnostics.HasError() {
		t.Fatalf("reading singular data source: %v", singularResponse.Diagnostics)
	}
	var singularState fleetScheduleDataSourceModel
	if diags := singularResponse.State.Get(ctx, &singularState); diags.HasError() {
		t.Fatalf("decoding singular state: %v", diags)
	}
	if singularState.ID.ValueString() != "schedule-b" || singularState.Rule == nil || singularState.Rule.StartMaintenanceWindow.ValueString() != "03:15" {
		t.Fatalf("unexpected singular state: %#v", singularState)
	}

	plural := &fleetSchedulesDataSource{Api: api, Auth: ctx}
	var pluralSchema datasource.SchemaResponse
	plural.Schema(ctx, datasource.SchemaRequest{}, &pluralSchema)
	pluralResponse := datasource.ReadResponse{State: tfsdk.State{Schema: pluralSchema.Schema}}
	plural.Read(ctx, datasource.ReadRequest{}, &pluralResponse)
	if pluralResponse.Diagnostics.HasError() {
		t.Fatalf("reading plural data source: %v", pluralResponse.Diagnostics)
	}
	var pluralState fleetSchedulesDataSourceModel
	if diags := pluralResponse.State.Get(ctx, &pluralState); diags.HasError() {
		t.Fatalf("decoding plural state: %v", diags)
	}
	if len(pluralState.Schedules) != 2 || pluralState.Schedules[0].ID.ValueString() != "schedule-a" || pluralState.Schedules[1].ID.ValueString() != "schedule-z" {
		t.Fatalf("unexpected plural state: %#v", pluralState)
	}
}

func fleetScheduleDataSourceResponseJSON(id string) string {
	return fmt.Sprintf(`{"data":%s}`, fleetScheduleDataJSON(id))
}

func fleetScheduleDataJSON(id string) string {
	return fmt.Sprintf(`{"id":%q,"type":"schedule","attributes":{"name":"test","query":"env:nonprod","status":"inactive","version_to_latest":1,"created_at":"2026-09-15T00:00:00Z","created_by":"creator@example.com","updated_at":"2026-09-15T01:00:00Z","updated_by":"updater@example.com","is_default":false,"next_run":null,"rule":{"days_of_week":["Mon"],"interval":1,"maintenance_window_duration":60,"start_maintenance_window":"0315","timezone":"UTC"},"notification_rule":{"handles":[],"tags":["team:fleet"]}}}`, id)
}
