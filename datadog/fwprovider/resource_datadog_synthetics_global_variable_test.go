package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
)

const persistentEmailAddress = "7a03c7471a3e4a68e60a.0f7d3f56-17a8-4e90-a342-9b6ae2a709bc@synthetics.dtdg.co"

func syntheticsGlobalVariableTestSchema(t *testing.T) schema.Schema {
	t.Helper()
	var response resource.SchemaResponse
	(&syntheticsGlobalVariableResource{}).Schema(context.Background(), resource.SchemaRequest{}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
	return response.Schema
}

func syntheticsGlobalVariableTestModel(t *testing.T, s schema.Schema) syntheticsGlobalVariableModel {
	t.Helper()
	ctx := context.Background()
	objectType := s.Type().TerraformType(ctx).(tftypes.Object)
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(objectType, attributes)}
	var model syntheticsGlobalVariableModel
	diags := state.Get(ctx, &model)
	require.False(t, diags.HasError(), "%v", diags)
	model.Name = types.StringValue("PERSISTENT_EMAIL")
	model.Description = types.StringValue("")
	model.Tags = types.ListValueMust(types.StringType, nil)
	model.Secure = types.BoolValue(false)
	model.IsEmail = types.BoolValue(true)
	model.IsFido = types.BoolValue(false)
	model.IsTotp = types.BoolValue(false)
	return model
}

func syntheticsGlobalVariableTestState(t *testing.T, s schema.Schema, model syntheticsGlobalVariableModel) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: s}
	diags := state.Set(context.Background(), &model)
	require.False(t, diags.HasError(), "%v", diags)
	return state
}

func TestSyntheticsGlobalVariableEmailValidateConfig(t *testing.T) {
	s := syntheticsGlobalVariableTestSchema(t)
	tests := []struct {
		name    string
		edit    func(*syntheticsGlobalVariableModel)
		invalid bool
	}{
		{"generated email", func(m *syntheticsGlobalVariableModel) {}, false},
		{"plain value still required", func(m *syntheticsGlobalVariableModel) { m.IsEmail = types.BoolValue(false) }, true},
		{"unknown kind", func(m *syntheticsGlobalVariableModel) { m.IsEmail = types.BoolUnknown() }, false},
		{"configured address", func(m *syntheticsGlobalVariableModel) { m.Value = types.StringValue(persistentEmailAddress) }, true},
		{"write only address", func(m *syntheticsGlobalVariableModel) { m.ValueWo = types.StringValue("address") }, true},
		{"write only version", func(m *syntheticsGlobalVariableModel) { m.ValueWoVersion = types.StringValue("1") }, true},
		{"secure", func(m *syntheticsGlobalVariableModel) { m.Secure = types.BoolValue(true) }, true},
		{"fido", func(m *syntheticsGlobalVariableModel) { m.IsFido = types.BoolValue(true) }, true},
		{"totp", func(m *syntheticsGlobalVariableModel) { m.IsTotp = types.BoolValue(true) }, true},
		{"parse test", func(m *syntheticsGlobalVariableModel) { m.ParseTestId = types.StringValue("abc-def-ghi") }, true},
		{"ordinary text", func(m *syntheticsGlobalVariableModel) {
			m.IsEmail = types.BoolValue(false)
			m.Value = types.StringValue("text")
		}, false},
		{"ordinary fido", func(m *syntheticsGlobalVariableModel) {
			m.IsEmail = types.BoolValue(false)
			m.IsFido = types.BoolValue(true)
			m.Secure = types.BoolValue(true)
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := syntheticsGlobalVariableTestModel(t, s)
			tt.edit(&model)
			state := syntheticsGlobalVariableTestState(t, s, model)
			var response resource.ValidateConfigResponse
			(&syntheticsGlobalVariableResource{}).ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: state.Raw}}, &response)
			require.Equal(t, tt.invalid, response.Diagnostics.HasError(), "%v", response.Diagnostics)
		})
	}
}

func TestSyntheticsGlobalVariableEmailRequestOmitsValue(t *testing.T) {
	r := &syntheticsGlobalVariableResource{}
	model := syntheticsGlobalVariableTestModel(t, syntheticsGlobalVariableTestSchema(t))
	// Even after refresh, the computed address must never become an API input.
	model.Value = types.StringValue(persistentEmailAddress)
	body, diags := r.buildSyntheticsGlobalVariableRequestBody(context.Background(), &model, fwutils.SecretResult{ShouldSetValue: true, Value: persistentEmailAddress})
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, body.GetIsEmail())
	require.False(t, body.HasValue())
	model.IsEmail = types.BoolValue(false)
	body, diags = r.buildSyntheticsGlobalVariableRequestBody(context.Background(), &model, fwutils.SecretResult{ShouldSetValue: true, Value: "normal value"})
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, body.HasIsEmail(), "do not change existing non-email request payloads")
	value := body.GetValue()
	require.Equal(t, "normal value", value.GetValue())
}

func TestSyntheticsGlobalVariableEmailModifyPlan(t *testing.T) {
	ctx := context.Background()
	s := syntheticsGlobalVariableTestSchema(t)
	tests := []struct {
		name       string
		priorEmail bool
		hasPrior   bool
		email      bool
		writeOnly  bool
		want       types.String
	}{
		{name: "create", email: true, want: types.StringUnknown()},
		{name: "metadata update", priorEmail: true, hasPrior: true, email: true, want: types.StringValue(persistentEmailAddress)},
		{name: "replace text with email", hasPrior: true, email: true, want: types.StringUnknown()},
		{name: "replace email with fido", priorEmail: true, hasPrior: true, want: types.StringNull()},
		{name: "switch to write only", hasPrior: true, writeOnly: true, want: types.StringNull()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configModel := syntheticsGlobalVariableTestModel(t, s)
			configModel.IsEmail = types.BoolValue(tt.email)
			configModel.IsFido = types.BoolValue(!tt.email && !tt.writeOnly)
			if tt.writeOnly {
				configModel.ValueWo = types.StringValue("secret")
				configModel.ValueWoVersion = types.StringValue("1")
			}
			config := syntheticsGlobalVariableTestState(t, s, configModel)
			planModel := configModel
			planModel.Value = types.StringUnknown()
			planState := syntheticsGlobalVariableTestState(t, s, planModel)
			plan := tfsdk.Plan{Schema: s, Raw: planState.Raw}
			prior := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
			if tt.hasPrior {
				priorModel := configModel
				priorModel.IsEmail = types.BoolValue(tt.priorEmail)
				priorModel.Value = types.StringValue(persistentEmailAddress)
				prior = syntheticsGlobalVariableTestState(t, s, priorModel)
			}
			response := resource.ModifyPlanResponse{Plan: plan}
			(&syntheticsGlobalVariableResource{}).ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, State: prior, Config: tfsdk.Config{Schema: s, Raw: config.Raw}}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			var value types.String
			require.False(t, response.Plan.GetAttribute(ctx, path.Root("value"), &value).HasError())
			require.Equal(t, tt.want, value)
		})
	}
}

func TestSyntheticsGlobalVariableEmailLifecycle(t *testing.T) {
	ctx := context.Background()
	const id = "00000000-0000-4000-8000-000000000001"
	description := ""
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		if req.Method == http.MethodPost || req.Method == http.MethodPut {
			var body map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
				http.Error(w, "bad request", 400)
				return
			}
			if body["is_email"] != true {
				t.Error("missing is_email in write")
			}
			if _, ok := body["value"]; ok {
				t.Error("email writes must omit value")
			}
			description, _ = body["description"].(string)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": id, "name": "PERSISTENT_EMAIL", "description": description, "tags": []string{}, "is_email": true,
			"value": map[string]interface{}{"value": persistentEmailAddress, "secure": false},
		})
	}))
	defer server.Close()
	apiConfig := datadog.NewConfiguration()
	apiConfig.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	apiConfig.OperationServers = nil
	apiConfig.HTTPClient = server.Client()
	r := &syntheticsGlobalVariableResource{Api: datadogV1.NewSyntheticsApi(datadog.NewAPIClient(apiConfig)), Auth: ctx}
	s := syntheticsGlobalVariableTestSchema(t)
	configModel := syntheticsGlobalVariableTestModel(t, s)
	configState := syntheticsGlobalVariableTestState(t, s, configModel)
	config := tfsdk.Config{Schema: s, Raw: configState.Raw}
	model := configModel
	model.Value = types.StringUnknown()
	model.Id = types.StringUnknown()
	planState := syntheticsGlobalVariableTestState(t, s, model)
	create := resource.CreateResponse{State: planState}
	r.Create(ctx, resource.CreateRequest{Config: config, Plan: tfsdk.Plan{Schema: s, Raw: planState.Raw}}, &create)
	require.False(t, create.Diagnostics.HasError(), "%v", create.Diagnostics)
	assertEmail := func(state tfsdk.State) {
		t.Helper()
		var got syntheticsGlobalVariableModel
		require.False(t, state.Get(ctx, &got).HasError())
		require.Equal(t, persistentEmailAddress, got.Value.ValueString())
		require.True(t, got.IsEmail.ValueBool())
		require.False(t, got.Secure.ValueBool())
	}
	assertEmail(create.State)
	read := resource.ReadResponse{State: create.State}
	r.Read(ctx, resource.ReadRequest{State: create.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	assertEmail(read.State)
	require.False(t, read.State.Get(ctx, &model).HasError())
	model.Description = types.StringValue("Updated metadata")
	configModel.Description = model.Description
	configState = syntheticsGlobalVariableTestState(t, s, configModel)
	planState = syntheticsGlobalVariableTestState(t, s, model)
	update := resource.UpdateResponse{State: planState}
	r.Update(ctx, resource.UpdateRequest{Config: tfsdk.Config{Schema: s, Raw: configState.Raw}, Plan: tfsdk.Plan{Schema: s, Raw: planState.Raw}, State: read.State}, &update)
	require.False(t, update.Diagnostics.HasError(), "%v", update.Diagnostics)
	assertEmail(update.State)
	require.Equal(t, "Updated metadata", description)
	// Import starts with only an ID; Read must restore the generated value and kind.
	imported := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	importResponse := resource.ImportStateResponse{State: imported}
	r.ImportState(ctx, resource.ImportStateRequest{ID: id}, &importResponse)
	require.False(t, importResponse.Diagnostics.HasError(), "%v", importResponse.Diagnostics)
	importedRead := resource.ReadResponse{State: importResponse.State}
	r.Read(ctx, resource.ReadRequest{State: importResponse.State}, &importedRead)
	require.False(t, importedRead.Diagnostics.HasError(), "%v", importedRead.Diagnostics)
	assertEmail(importedRead.State)
	require.Equal(t, 4, requests, fmt.Sprintf("unexpected number of API requests: %d", requests))
}
