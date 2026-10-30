package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestDerivedMetricReplacementPlan(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		metricName  string
		formula     string
		query       string
		wantReplace bool
	}{
		{"unchanged", "custom.cpu.original", "a", "avg:system.cpu.user{*}", false},
		{"rename", "custom.cpu.renamed", "a", "avg:system.cpu.user{*}", true},
		{"formula update", "custom.cpu.original", "a * 2", "avg:system.cpu.user{*}", false},
		{"query update", "custom.cpu.original", "a", "avg:system.cpu.system{*}", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			server := providerserver.NewProtocol6(New())()
			prior := derivedMetricTestState(t)
			config := prior
			config.ID = types.StringNull()
			config.Name = types.StringValue(test.metricName)
			config.Formula = types.StringValue(test.formula)
			config.Queries = append([]derivedMetricQueryModel(nil), prior.Queries...)
			config.Queries[0].Query = types.StringValue(test.query)
			proposed := config
			proposed.ID = prior.ID

			response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName:         "datadog_derived_metric",
				PriorState:       derivedMetricTestDynamicValue(t, prior),
				Config:           derivedMetricTestDynamicValue(t, config),
				ProposedNewState: derivedMetricTestDynamicValue(t, proposed),
			})
			require.NoError(t, err)
			for _, diagnostic := range response.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
			}
			if test.wantReplace {
				require.Len(t, response.RequiresReplace, 1)
				require.True(t, response.RequiresReplace[0].Equal(tftypes.NewAttributePath().WithAttributeName("name")))
			} else {
				require.Empty(t, response.RequiresReplace)
			}
		})
	}
}

func TestDerivedMetricReadMissing(t *testing.T) {
	t.Parallel()

	for _, status := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodGet || request.URL.Path != "/api/v2/metrics/derived-metrics/custom.cpu.original" {
					t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				fmt.Fprintf(w, `{"errors":[%q]}`, http.StatusText(status))
			}))
			defer server.Close()

			config := datadog.NewConfiguration()
			config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
			config.OperationServers = nil
			config.HTTPClient = server.Client()
			r := &derivedMetricResource{
				Api:  datadogV2.NewMetricsApi(datadog.NewAPIClient(config)),
				Auth: ctx,
			}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			state := tfsdk.State{Schema: schemaResponse.Schema}
			require.False(t, state.Set(ctx, derivedMetricTestState(t)).HasError())
			response := resource.ReadResponse{State: state}

			r.Read(ctx, resource.ReadRequest{State: state}, &response)

			if status == http.StatusNotFound {
				require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
				require.True(t, response.State.Raw.IsNull(), "an externally deleted metric must be removed from Terraform state")
			} else {
				require.True(t, response.Diagnostics.HasError(), "API errors must be reported")
				require.True(t, response.State.Raw.Equal(state.Raw), "API errors must preserve Terraform state")
			}
		})
	}
}

func derivedMetricTestState(t *testing.T) derivedMetricModel {
	t.Helper()
	return derivedMetricModel{
		ID:      types.StringValue("custom.cpu.original"),
		Name:    types.StringValue("custom.cpu.original"),
		Formula: types.StringValue("a"),
		Queries: []derivedMetricQueryModel{{
			Name:       types.StringValue("a"),
			DataSource: types.StringValue("metrics"),
			Query:      types.StringValue("avg:system.cpu.user{*}"),
		}},
	}
}

func derivedMetricTestDynamicValue(t *testing.T, model derivedMetricModel) *tfprotov6.DynamicValue {
	t.Helper()
	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	NewDerivedMetricResource().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	state := tfsdk.State{Schema: schemaResponse.Schema}
	require.False(t, state.Set(ctx, model).HasError())
	value, err := tfprotov6.NewDynamicValue(state.Raw.Type(), state.Raw)
	require.NoError(t, err)
	return &value
}
