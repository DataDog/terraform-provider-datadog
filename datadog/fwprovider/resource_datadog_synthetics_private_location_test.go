package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestSyntheticsPrivateLocationUpdatePreservesConfig(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/api/v1/synthetics/private-locations/pl:test" {
			t.Errorf("unexpected request: %s %s", request.Method, request.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"pl:test","name":"updated","description":"updated description","tags":[]}`)
	}))
	defer server.Close()

	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	r := &syntheticsPrivateLocationResource{
		Api:  datadogV1.NewSyntheticsApi(datadog.NewAPIClient(config)),
		Auth: ctx,
	}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	for _, test := range []struct {
		name    string
		planned types.String
		prior   types.String
		want    types.String
	}{
		{"created", types.StringValue(`{"accessKey":"test"}`), types.StringValue(`{"accessKey":"test"}`), types.StringValue(`{"accessKey":"test"}`)},
		{"imported", types.StringNull(), types.StringNull(), types.StringNull()},
		{"unknown with prior config", types.StringUnknown(), types.StringValue(`{"accessKey":"test"}`), types.StringValue(`{"accessKey":"test"}`)},
		{"unknown after import", types.StringUnknown(), types.StringNull(), types.StringNull()},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := syntheticsPrivateLocationModel{
				Id:          types.StringValue("pl:test"),
				Name:        types.StringValue("updated"),
				Description: types.StringValue("updated description"),
				Tags:        types.ListValueMust(types.StringType, nil),
				Config:      test.planned,
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			require.False(t, plan.Set(ctx, &model).HasError())
			model.Name = types.StringValue("original")
			model.Config = test.prior
			prior := tfsdk.State{Schema: schemaResponse.Schema}
			require.False(t, prior.Set(ctx, &model).HasError())

			response := resource.UpdateResponse{State: tfsdk.State{Schema: schemaResponse.Schema, Raw: plan.Raw}}
			r.Update(ctx, resource.UpdateRequest{Plan: plan, State: prior}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			var got syntheticsPrivateLocationModel
			require.False(t, response.State.Get(ctx, &got).HasError())
			require.Equal(t, test.want, got.Config)
			require.Equal(t, "updated", got.Name.ValueString())
		})
	}
}
