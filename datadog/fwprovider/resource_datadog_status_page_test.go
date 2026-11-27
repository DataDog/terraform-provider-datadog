package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"
)

const statusPageTestID = "00000000-0000-4000-8000-000000000001"

func TestStatusPageSettingsCreateUpdateRead(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		layout types.String
		rss    types.Bool
	}{
		{"grid with RSS", types.StringValue("grid"), types.BoolValue(true)},
		{"list without RSS", types.StringValue("list"), types.BoolValue(false)},
		{"omitted", types.StringNull(), types.BoolNull()},
		{"unknown", types.StringUnknown(), types.BoolUnknown()},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			layout, rss := "list", false
			if !test.layout.IsNull() && !test.layout.IsUnknown() {
				layout = test.layout.ValueString()
			}
			if !test.rss.IsNull() && !test.rss.IsUnknown() {
				rss = test.rss.ValueBool()
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPost {
					require.Equal(t, "/api/v2/statuspages", request.URL.Path)
				} else {
					require.Equal(t, "/api/v2/statuspages/"+statusPageTestID, request.URL.Path)
				}
				if request.Method != http.MethodGet {
					require.Contains(t, []string{http.MethodPost, http.MethodPatch}, request.Method)
					var body struct {
						Data struct {
							Attributes map[string]interface{} `json:"attributes"`
						} `json:"data"`
					}
					require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
					if test.layout.IsNull() || test.layout.IsUnknown() {
						require.NotContains(t, body.Data.Attributes, "visualization_layout")
					} else {
						require.Equal(t, layout, body.Data.Attributes["visualization_layout"])
					}
					if test.rss.IsNull() || test.rss.IsUnknown() {
						require.NotContains(t, body.Data.Attributes, "rss_feed_enabled")
					} else {
						require.Equal(t, rss, body.Data.Attributes["rss_feed_enabled"])
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if request.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
				}
				fmt.Fprintf(w, `{"data":%s}`, statusPageTestData(layout, rss))
			}))
			defer server.Close()
			r := &statusPageResource{Api: statusPageTestAPI(server), Auth: ctx}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			model := statusPageTestModel()
			model.VisualizationLayout, model.RssFeedEnabled = test.layout, test.rss
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, model).HasError())
			created := resource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
			r.Create(ctx, resource.CreateRequest{Plan: plan}, &created)
			require.False(t, created.Diagnostics.HasError(), "%v", created.Diagnostics)
			var got statusPageModel
			require.False(t, created.State.Get(ctx, &got).HasError())
			require.Equal(t, types.StringValue(layout), got.VisualizationLayout)
			require.Equal(t, types.BoolValue(rss), got.RssFeedEnabled)

			model.ID = types.StringValue(statusPageTestID)
			require.False(t, plan.Set(ctx, model).HasError())
			updated := resource.UpdateResponse{State: created.State}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: created.State}, &updated)
			require.False(t, updated.Diagnostics.HasError(), "%v", updated.Diagnostics)
			require.False(t, updated.State.Get(ctx, &got).HasError())
			require.Equal(t, types.StringValue(layout), got.VisualizationLayout)
			require.Equal(t, types.BoolValue(rss), got.RssFeedEnabled)

			read := resource.ReadResponse{State: updated.State}
			r.Read(ctx, resource.ReadRequest{State: updated.State}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			require.False(t, read.State.Get(ctx, &got).HasError())
			require.Equal(t, types.StringValue(layout), got.VisualizationLayout)
			require.Equal(t, types.BoolValue(rss), got.RssFeedEnabled)
		})
	}
}

func TestStatusPageLayoutValidation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		layout        types.String
		visualization types.String
		wantError     bool
	}{
		{types.StringValue("list"), types.StringValue("bars_only"), false},
		{types.StringValue("grid"), types.StringValue("component_name_only"), false},
		{types.StringValue("grid"), types.StringValue("bars_only"), true},
		{types.StringValue("grid"), types.StringValue("bars_and_uptime_percentage"), true},
		{types.StringValue("invalid"), types.StringValue("component_name_only"), true},
		{types.StringNull(), types.StringValue("bars_only"), false},
		{types.StringUnknown(), types.StringValue("bars_only"), false},
		{types.StringValue("grid"), types.StringUnknown(), false},
	} {
		t.Run(fmt.Sprintf("%s/%s", test.layout, test.visualization), func(t *testing.T) {
			model := statusPageTestModel()
			model.VisualizationLayout, model.VisualizationType = test.layout, test.visualization
			server := providerserver.NewProtocol6(New())()
			response, err := server.ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{
				TypeName: "datadog_status_page",
				Config:   statusPageTestDynamicValue(t, model),
			})
			require.NoError(t, err)
			hasError := false
			for _, diagnostic := range response.Diagnostics {
				hasError = hasError || diagnostic.Severity == tfprotov6.DiagnosticSeverityError
			}
			require.Equal(t, test.wantError, hasError, "%v", response.Diagnostics)
		})
	}
}

func TestStatusPageLayoutRecomputedWhenVisualizationChanges(t *testing.T) {
	t.Parallel()
	prior := statusPageTestModel()
	prior.ID = types.StringValue(statusPageTestID)
	prior.VisualizationLayout = types.StringValue("grid")
	prior.RssFeedEnabled = types.BoolValue(true)
	config := statusPageTestModel()
	config.VisualizationType = types.StringValue("bars_only")
	proposed := config
	proposed.ID = prior.ID
	proposed.VisualizationLayout = prior.VisualizationLayout
	proposed.RssFeedEnabled = prior.RssFeedEnabled
	server := providerserver.NewProtocol6(New())()
	response, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{
		TypeName:         "datadog_status_page",
		PriorState:       statusPageTestDynamicValue(t, prior),
		Config:           statusPageTestDynamicValue(t, config),
		ProposedNewState: statusPageTestDynamicValue(t, proposed),
	})
	require.NoError(t, err)
	for _, diagnostic := range response.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s", diagnostic.Detail)
	}
	var schemaResponse resource.SchemaResponse
	NewStatusPageResource().Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	planned := tfsdk.Plan{Schema: schemaResponse.Schema}
	planned.Raw, err = response.PlannedState.Unmarshal(schemaResponse.Schema.Type().TerraformType(context.Background()))
	require.NoError(t, err)
	var got statusPageModel
	require.False(t, planned.Get(context.Background(), &got).HasError())
	require.True(t, got.VisualizationLayout.IsUnknown(), "the API can reset grid to list when visualization_type changes")
	require.Equal(t, prior.RssFeedEnabled, got.RssFeedEnabled)
}

func TestStatusPageURLPlanModifier(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name        string
		priorType   string
		plannedType string
		prefix      string
		wantUnknown bool
	}{
		{"public to private", "public", "private", "example-status", true},
		{"private to public", "private", "public", "example-status", true},
		{"internal to public", "internal", "public", "example-status", true},
		{"public to internal", "public", "internal", "example-status", true},
		{"internal to private", "internal", "private", "example-status", true},
		{"private to internal", "private", "internal", "example-status", true},
		{"prefix changed", "public", "public", "updated-status", true},
		{"public unchanged", "public", "public", "example-status", false},
		{"private unchanged", "private", "private", "example-status", false},
		{"internal unchanged", "internal", "internal", "example-status", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			prior := statusPageTestModel()
			prior.ID = types.StringValue(statusPageTestID)
			prior.Type = types.StringValue(test.priorType)
			switch test.priorType {
			case "private":
				prior.PageURL = types.StringValue("https://example-status.private.statuspage.datadoghq.com")
			case "internal":
				prior.PageURL = types.StringValue("https://app.datadoghq.com/status-pages/" + statusPageTestID + "/view")
			default:
				prior.PageURL = types.StringValue("https://example-status.statuspage.datadoghq.com")
			}
			config := statusPageTestModel()
			config.Type = types.StringValue(test.plannedType)
			config.DomainPrefix = types.StringValue(test.prefix)
			proposed := config
			proposed.ID = prior.ID
			proposed.PageURL = prior.PageURL
			server := providerserver.NewProtocol6(New())()
			response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName:         "datadog_status_page",
				PriorState:       statusPageTestDynamicValue(t, prior),
				Config:           statusPageTestDynamicValue(t, config),
				ProposedNewState: statusPageTestDynamicValue(t, proposed),
			})
			require.NoError(t, err)
			for _, diagnostic := range response.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s", diagnostic.Detail)
			}
			var schemaResponse resource.SchemaResponse
			NewStatusPageResource().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			planned := tfsdk.Plan{Schema: schemaResponse.Schema}
			planned.Raw, err = response.PlannedState.Unmarshal(schemaResponse.Schema.Type().TerraformType(ctx))
			require.NoError(t, err)
			var got statusPageModel
			require.False(t, planned.Get(ctx, &got).HasError())
			if test.wantUnknown {
				require.True(t, got.PageURL.IsUnknown())
			} else {
				require.Equal(t, prior.PageURL, got.PageURL)
			}
		})
	}
}

func TestStatusPageSettingsDataSources(t *testing.T) {
	t.Parallel()
	for _, rss := range []bool{true, false} {
		t.Run(fmt.Sprint(rss), func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				require.Equal(t, http.MethodGet, request.Method)
				w.Header().Set("Content-Type", "application/json")
				switch request.URL.Path {
				case "/api/v2/statuspages/" + statusPageTestID:
					fmt.Fprintf(w, `{"data":%s}`, statusPageTestData("grid", rss))
				case "/api/v2/statuspages":
					fmt.Fprintf(w, `{"data":[%s]}`, statusPageTestData("grid", rss))
				default:
					http.NotFound(w, request)
				}
			}))
			defer server.Close()
			api := statusPageTestAPI(server)
			singular := &statusPageDataSource{Api: api, Auth: ctx}
			var singularSchema datasource.SchemaResponse
			singular.Schema(ctx, datasource.SchemaRequest{}, &singularSchema)
			config := tfsdk.Plan{Schema: singularSchema.Schema}
			require.False(t, config.Set(ctx, statusPageDataSourceModel{ID: types.StringValue(statusPageTestID)}).HasError())
			singularResponse := datasource.ReadResponse{State: tfsdk.State{Schema: singularSchema.Schema}}
			singular.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Raw: config.Raw, Schema: singularSchema.Schema}}, &singularResponse)
			require.False(t, singularResponse.Diagnostics.HasError(), "%v", singularResponse.Diagnostics)
			var singularState statusPageDataSourceModel
			require.False(t, singularResponse.State.Get(ctx, &singularState).HasError())
			require.Equal(t, types.StringValue("grid"), singularState.VisualizationLayout)
			require.Equal(t, types.BoolValue(rss), singularState.RssFeedEnabled)

			plural := &statusPagesDataSource{Api: api, Auth: ctx}
			var pluralSchema datasource.SchemaResponse
			plural.Schema(ctx, datasource.SchemaRequest{}, &pluralSchema)
			pluralConfig := tfsdk.Plan{Schema: pluralSchema.Schema}
			require.False(t, pluralConfig.Set(ctx, statusPagesDataSourceModel{}).HasError())
			pluralResponse := datasource.ReadResponse{State: tfsdk.State{Schema: pluralSchema.Schema}}
			plural.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Raw: pluralConfig.Raw, Schema: pluralSchema.Schema}}, &pluralResponse)
			require.False(t, pluralResponse.Diagnostics.HasError(), "%v", pluralResponse.Diagnostics)
			var pluralState statusPagesDataSourceModel
			require.False(t, pluralResponse.State.Get(ctx, &pluralState).HasError())
			require.Len(t, pluralState.StatusPages, 1)
			require.Equal(t, singularState.VisualizationLayout, pluralState.StatusPages[0].VisualizationLayout)
			require.Equal(t, singularState.RssFeedEnabled, pluralState.StatusPages[0].RssFeedEnabled)
		})
	}
}

func statusPageTestModel() statusPageModel {
	return statusPageModel{
		Name:              types.StringValue("Example Status"),
		Type:              types.StringValue("public"),
		DomainPrefix:      types.StringValue("example-status"),
		VisualizationType: types.StringValue("component_name_only"),
	}
}

func statusPageTestAPI(server *httptest.Server) *datadogV2.StatusPagesApi {
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	return datadogV2.NewStatusPagesApi(datadog.NewAPIClient(config))
}

func statusPageTestData(layout string, rss bool) string {
	return fmt.Sprintf(`{"id":%q,"type":"status_pages","attributes":{"name":"Example Status","type":"public","domain_prefix":"example-status","visualization_type":"component_name_only","visualization_layout":%q,"rss_feed_enabled":%t}}`, statusPageTestID, layout, rss)
}

func statusPageTestDynamicValue(t *testing.T, model statusPageModel) *tfprotov6.DynamicValue {
	t.Helper()
	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	NewStatusPageResource().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	require.False(t, state.Set(ctx, model).HasError())
	value, err := tfprotov6.NewDynamicValue(state.Raw.Type(), state.Raw)
	require.NoError(t, err)
	return &value
}
