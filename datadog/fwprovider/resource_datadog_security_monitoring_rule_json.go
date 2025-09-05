package fwprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func canonicalizeJSONText(s string) (string, error) {
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil { // parse the JSON TEXT
		return "", err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil { // encode the OBJECT back to text
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func canonicalizeInterface(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringRuleJSONResource{}
	_ resource.ResourceWithImportState = &securityMonitoringRuleJSONResource{}
)

type securityMonitoringRuleJSONResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityMonitoringRuleJSONModel struct {
	ID   types.String `tfsdk:"id"`
	JSON types.String `tfsdk:"json"`
}

func NewSecurityMonitoringRuleJSONResource() resource.Resource {
	return &securityMonitoringRuleJSONResource{}
}

func (r *securityMonitoringRuleJSONResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityMonitoringRuleJSONResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_rule_json"
}

func (r *securityMonitoringRuleJSONResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule JSON resource. This can be used to create and manage Datadog security monitoring rules using raw JSON.",
		Attributes: map[string]schema.Attribute{
			"json": schema.StringAttribute{
				Required:    true,
				Description: "The JSON definition of the Security Monitoring Rule.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *securityMonitoringRuleJSONResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// Helper to recursively filter API response to only user-supplied fields
func filterToUserFields(user interface{}, api interface{}) interface{} {
	switch userVal := user.(type) {
	case map[string]interface{}:
		apiMap, ok := api.(map[string]interface{})
		if !ok {
			return user
		}
		filtered := make(map[string]interface{})
		for k, v := range userVal {
			if apiV, ok := apiMap[k]; ok {
				filtered[k] = filterToUserFields(v, apiV)
			}
		}
		return filtered
	case []interface{}:
		apiArr, ok := api.([]interface{})
		if !ok {
			return user
		}
		filteredArr := make([]interface{}, len(userVal))
		for i := range userVal {
			if i < len(apiArr) {
				filteredArr[i] = filterToUserFields(userVal[i], apiArr[i])
			} else {
				filteredArr[i] = userVal[i]
			}
		}
		return filteredArr
	default:
		return api
	}
}

func (r *securityMonitoringRuleJSONResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityMonitoringRuleJSONModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Also read the original config (what the user actually typed)
	var cfg securityMonitoringRuleJSONModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use cfg.JSON to build the payload (so we truly use what user sent)
	var userRule map[string]interface{}
	if err := json.Unmarshal([]byte(cfg.JSON.ValueString()), &userRule); err != nil {
		resp.Diagnostics.AddError("Failed to parse JSON", err.Error())
		return
	}

	// Build payload from userRule (unchanged)
	payload := datadogV2.SecurityMonitoringRuleCreatePayload{}
	b, _ := json.Marshal(userRule)
	if err := json.Unmarshal(b, &payload); err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal to payload", err.Error())
		return
	}

	res, httpResp, err := r.Api.CreateSecurityMonitoringRule(r.Auth, payload)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error creating security monitoring rule"), ""))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		resp.Diagnostics.AddError("Failed to parse response", err.Error())
		return
	}

	var state securityMonitoringRuleJSONModel
	// IMPORTANT: write back EXACTLY what the user had in config
	state.JSON = types.StringValue(cfg.JSON.ValueString())

	if res.SecurityMonitoringStandardRuleResponse != nil {
		state.ID = types.StringValue(res.SecurityMonitoringStandardRuleResponse.GetId())
	} else if res.SecurityMonitoringSignalRuleResponse != nil {
		state.ID = types.StringValue(res.SecurityMonitoringSignalRuleResponse.GetId())
	} else {
		resp.Diagnostics.AddError("Invalid response", "Response did not contain an ID")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// func (r *securityMonitoringRuleJSONResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
// 	var state securityMonitoringRuleJSONModel
// 	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
// 	if response.Diagnostics.HasError() {
// 		return
// 	}

// 	// Parse user JSON into a map
// 	var userRule map[string]interface{}
// 	if err := json.Unmarshal([]byte(state.JSON.ValueString()), &userRule); err != nil {
// 		response.Diagnostics.AddError("Failed to parse JSON", err.Error())
// 		return
// 	}

// 	tflog.Warn(ctx, "MARTIN Create: incoming JSON", map[string]any{
// 		"json": state.JSON.ValueString(),
// 	})

// 	// Convert the map to SecurityMonitoringRuleCreatePayload
// 	payload := datadogV2.SecurityMonitoringRuleCreatePayload{}
// 	jsonBytes, err := json.Marshal(userRule)
// 	if err != nil {
// 		response.Diagnostics.AddError("Failed to marshal rule", err.Error())
// 		return
// 	}
// 	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
// 		response.Diagnostics.AddError("Failed to unmarshal to payload", err.Error())
// 		return
// 	}

// 	res, httpResp, err := r.Api.CreateSecurityMonitoringRule(r.Auth, payload)
// 	if err != nil {
// 		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error creating security monitoring rule"), ""))
// 		return
// 	}
// 	if err := utils.CheckForUnparsed(res); err != nil {
// 		response.Diagnostics.AddError("Failed to parse response", err.Error())
// 		return
// 	}

// 	var apiRule map[string]interface{}
// 	if res.SecurityMonitoringStandardRuleResponse != nil {
// 		jsonBytes, err = json.Marshal(res.SecurityMonitoringStandardRuleResponse)
// 	} else if res.SecurityMonitoringSignalRuleResponse != nil {
// 		jsonBytes, err = json.Marshal(res.SecurityMonitoringSignalRuleResponse)
// 	} else {
// 		response.Diagnostics.AddError("Invalid response", "Response did not contain a rule")
// 		return
// 	}
// 	if err != nil {
// 		response.Diagnostics.AddError("Failed to marshal response", err.Error())
// 		return
// 	}

// 	if err := json.Unmarshal(jsonBytes, &apiRule); err != nil {
// 		response.Diagnostics.AddError("Failed to parse response", err.Error())
// 		return
// 	}

// 	// Filter API response to only user-supplied fields
// 	filtered := filterToUserFields(userRule, apiRule)
// 	// jsonBytes, err = json.Marshal(filtered)
// 	// if err != nil {
// 	// 	response.Diagnostics.AddError("Failed to marshal filtered response", err.Error())
// 	// 	return
// 	// }

// 	// debug
// 	norm, err := canonicalizeInterface(filtered)
// 	if err != nil {
// 		response.Diagnostics.AddError("Failed to normalize filtered response", err.Error())
// 		return
// 	}
// 	state.JSON = types.StringValue(norm)

// 	tflog.Warn(ctx, "MARTIN response", map[string]any{
// 		"json": norm,
// 	})
// 	// debug

// 	if res.SecurityMonitoringStandardRuleResponse != nil {
// 		state.ID = types.StringValue(res.SecurityMonitoringStandardRuleResponse.GetId())
// 	} else if res.SecurityMonitoringSignalRuleResponse != nil {
// 		state.ID = types.StringValue(res.SecurityMonitoringSignalRuleResponse.GetId())
// 	} else {
// 		response.Diagnostics.AddError("Invalid response", "Response did not contain an ID")
// 		return
// 	}
// 	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
// }

func (r *securityMonitoringRuleJSONResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityMonitoringRuleJSONModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, httpResp, err := r.Api.GetSecurityMonitoringRule(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error reading security monitoring rule"), ""))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		resp.Diagnostics.AddError("Failed to parse response", err.Error())
		return
	}

	// Do NOT rewrite state.JSON here. Keep user’s exact JSON.
	// Just ensure ID stays (it already does).
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// func (r *securityMonitoringRuleJSONResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
// 	var state securityMonitoringRuleJSONModel
// 	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
// 	if response.Diagnostics.HasError() {
// 		return
// 	}

// 	res, httpResp, err := r.Api.GetSecurityMonitoringRule(r.Auth, state.ID.ValueString())
// 	if err != nil {
// 		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
// 			response.State.RemoveResource(ctx)
// 			return
// 		}
// 		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error reading security monitoring rule"), ""))
// 		return
// 	}
// 	if err := utils.CheckForUnparsed(res); err != nil {
// 		response.Diagnostics.AddError("Failed to parse response", err.Error())
// 		return
// 	}

// 	var userRule map[string]interface{}
// 	if err := json.Unmarshal([]byte(state.JSON.ValueString()), &userRule); err != nil {
// 		response.Diagnostics.AddError("Failed to parse state JSON", err.Error())
// 		return
// 	}

// 	var apiRule map[string]interface{}
// 	var jsonBytes []byte
// 	if res.SecurityMonitoringStandardRuleResponse != nil {
// 		jsonBytes, err = json.Marshal(res.SecurityMonitoringStandardRuleResponse)
// 	} else if res.SecurityMonitoringSignalRuleResponse != nil {
// 		jsonBytes, err = json.Marshal(res.SecurityMonitoringSignalRuleResponse)
// 	} else {
// 		response.Diagnostics.AddError("Invalid response", "Response did not contain a rule")
// 		return
// 	}
// 	if err != nil {
// 		response.Diagnostics.AddError("Failed to marshal response", err.Error())
// 		return
// 	}

// 	if err := json.Unmarshal(jsonBytes, &apiRule); err != nil {
// 		response.Diagnostics.AddError("Failed to parse response", err.Error())
// 		return
// 	}

// 	filtered := filterToUserFields(userRule, apiRule)
// 	jsonBytes, err = json.Marshal(filtered)
// 	if err != nil {
// 		response.Diagnostics.AddError("Failed to marshal filtered response", err.Error())
// 		return
// 	}
// 	state.JSON = types.StringValue(string(jsonBytes))
// 	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
// }

func (r *securityMonitoringRuleJSONResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityMonitoringRuleJSONModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var userRule map[string]interface{}
	if err := json.Unmarshal([]byte(state.JSON.ValueString()), &userRule); err != nil {
		response.Diagnostics.AddError("Failed to parse JSON", err.Error())
		return
	}

	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}
	jsonBytes, err := json.Marshal(userRule)
	if err != nil {
		response.Diagnostics.AddError("Failed to marshal rule", err.Error())
		return
	}
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		response.Diagnostics.AddError("Failed to unmarshal to payload", err.Error())
		return
	}

	res, httpResp, err := r.Api.UpdateSecurityMonitoringRule(r.Auth, state.ID.ValueString(), payload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error updating security monitoring rule"), ""))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.AddError("Failed to parse response", err.Error())
		return
	}

	var apiRule map[string]interface{}
	if res.SecurityMonitoringStandardRuleResponse != nil {
		jsonBytes, err = json.Marshal(res.SecurityMonitoringStandardRuleResponse)
	} else if res.SecurityMonitoringSignalRuleResponse != nil {
		jsonBytes, err = json.Marshal(res.SecurityMonitoringSignalRuleResponse)
	} else {
		response.Diagnostics.AddError("Invalid response", "Response did not contain a rule")
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Failed to marshal response", err.Error())
		return
	}

	if err := json.Unmarshal(jsonBytes, &apiRule); err != nil {
		response.Diagnostics.AddError("Failed to parse response", err.Error())
		return
	}

	filtered := filterToUserFields(userRule, apiRule)
	jsonBytes, err = json.Marshal(filtered)
	if err != nil {
		response.Diagnostics.AddError("Failed to marshal filtered response", err.Error())
		return
	}
	state.JSON = types.StringValue(string(jsonBytes))
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringRuleJSONResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityMonitoringRuleJSONModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteSecurityMonitoringRule(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error deleting security monitoring rule"), ""))
		return
	}
}
