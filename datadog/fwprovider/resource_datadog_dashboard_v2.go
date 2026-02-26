package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &dashboardV2Resource{}
	_ resource.ResourceWithImportState = &dashboardV2Resource{}
	_ resource.ResourceWithModifyPlan  = &dashboardV2Resource{}
)

// NewDashboardV2Resource returns a new framework resource for datadog_dashboard_v2.
func NewDashboardV2Resource() resource.Resource {
	return &dashboardV2Resource{}
}

type dashboardV2Resource struct {
	ApiInstances *utils.ApiInstances
	Auth         context.Context
}

// dashboardV2ResourceModel holds the Terraform state for datadog_dashboard_v2.
// Simple top-level fields use concrete types; complex nested blocks use types.List/types.Set.
type dashboardV2ResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Title                  types.String `tfsdk:"title"`
	LayoutType             types.String `tfsdk:"layout_type"`
	ReflowType             types.String `tfsdk:"reflow_type"`
	Description            types.String `tfsdk:"description"`
	URL                    types.String `tfsdk:"url"`
	IsReadOnly             types.Bool   `tfsdk:"is_read_only"`
	RestrictedRoles        types.List   `tfsdk:"restricted_roles"` // ListAttribute (was Set - see schema_gen.go UseSet comment)
	NotifyList             types.List   `tfsdk:"notify_list"`      // ListAttribute (was Set - see schema_gen.go UseSet comment)
	Tags                   types.List   `tfsdk:"tags"`
	DashboardLists         types.Set    `tfsdk:"dashboard_lists"`         // TypeIntList+UseSet → still Set
	DashboardListsRemoved  types.Set    `tfsdk:"dashboard_lists_removed"` // TypeIntList+UseSet → still Set
	TemplateVariable       types.List   `tfsdk:"template_variable"`
	TemplateVariablePreset types.List   `tfsdk:"template_variable_preset"`
	Widget                 types.List   `tfsdk:"widget"`
}

func (r *dashboardV2Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	providerData, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		return
	}
	r.ApiInstances = providerData.DatadogApiInstances
	r.Auth = providerData.Auth
}

func (r *dashboardV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

func (r *dashboardV2Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "dashboard_v2"
}

func (r *dashboardV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	topAttrs, topBlocks := dashboardmapping.FieldSpecsToFWSchema(dashboardmapping.DashboardTopLevelFields)

	// Build the widget block (with nested group_definition support)
	widgetAttrs, widgetBlocks := dashboardmapping.AllWidgetFWBlocks(false)
	topBlocks["widget"] = schema.ListNestedBlock{
		Description: "The list of widgets to display on the dashboard.",
		NestedObject: schema.NestedBlockObject{
			Attributes: widgetAttrs,
			Blocks:     widgetBlocks,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Provides a Datadog dashboard resource (v2, FieldSpec engine). This can be used to create and manage Datadog dashboards.\n\n!> The `is_read_only` field is deprecated and non-functional. Use `restricted_roles` instead to define which roles are required to edit the dashboard.",
		Attributes:  topAttrs,
		Blocks:      topBlocks,
	}
	// Prepend "datadog_" to the type name in provider metadata
	resp.Schema.Attributes["id"] = utils.ResourceIDAttribute()
}

func (r *dashboardV2Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Replicate the SDKv2 CustomizeDiff for dashboard_lists_removed:
	// Whenever dashboard_lists changes, compute the removed list.
	if req.Plan.Raw.IsNull() {
		// Destroy plan - nothing to do
		return
	}
	var plan dashboardV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dashboardV2ResourceModel
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Compute dashboard_lists_removed = old lists - new lists
	oldLists := setToInt64Slice(state.DashboardLists)
	newLists := setToInt64Slice(plan.DashboardLists)
	newSet := make(map[int64]bool, len(newLists))
	for _, id := range newLists {
		newSet[id] = true
	}
	removed := make([]int64, 0)
	for _, id := range oldLists {
		if !newSet[id] {
			removed = append(removed, id)
		}
	}

	if len(removed) > 0 {
		removedVals := make([]types.Int64, len(removed))
		for i, id := range removed {
			removedVals[i] = types.Int64Value(id)
		}
		// Update dashboard_lists_removed in the plan
		removedAttrVals := make([]interface{}, len(removed))
		for i, id := range removed {
			removedAttrVals[i] = id
		}
		removedSet, diags := types.SetValueFrom(ctx, types.Int64Type, removedAttrVals)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			plan.DashboardListsRemoved = removedSet
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
	}
}

func (r *dashboardV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	topLevelAttrs := modelToTopLevelAttrs(&plan)
	bodyStr, err := dashboardmapping.MarshalDashboardJSON(topLevelAttrs, "")
	if err != nil {
		resp.Diagnostics.AddError("Error building dashboard JSON", err.Error())
		return
	}

	respByte, httpresp, err := utils.SendRequest(r.Auth, r.ApiInstances.HttpClient, "POST", dashboardmapping.DashboardAPIPath, &bodyStr)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error creating dashboard"))
		return
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing dashboard response", err.Error())
		return
	}

	id, ok := respMap["id"]
	if !ok {
		resp.Diagnostics.AddError("Error retrieving dashboard ID", "id not found in response")
		return
	}
	dashboardID := fmt.Sprintf("%v", id)
	plan.ID = types.StringValue(dashboardID)

	layoutType, _ := respMap["layout_type"].(string)

	// Retry GET until the dashboard is available
	var httpResponse *http.Response
	retryErr := retryContextDashboard(ctx, func() error {
		_, httpResponse, err = utils.SendRequest(r.Auth, r.ApiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+dashboardID, nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return fmt.Errorf("dashboard not created yet")
			}
			return fmt.Errorf("non-retryable error: %w", err)
		}
		return nil
	})
	if retryErr != nil {
		resp.Diagnostics.AddError("Error waiting for dashboard", retryErr.Error())
		return
	}

	// Update dashboard lists (side-effect)
	r.updateDashboardLists(plan.DashboardLists, plan.DashboardListsRemoved, dashboardID, layoutType)

	// Set state from response
	diags := r.setStateFromResp(ctx, respMap, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := state.ID.ValueString()
	respByte, httpresp, err := utils.SendRequest(r.Auth, r.ApiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+dashboardID, nil)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error getting dashboard"))
		return
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing dashboard response", err.Error())
		return
	}

	diags := r.setStateFromResp(ctx, respMap, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dashboardV2ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := plan.ID.ValueString()
	topLevelAttrs := modelToTopLevelAttrs(&plan)
	bodyStr, err := dashboardmapping.MarshalDashboardJSON(topLevelAttrs, dashboardID)
	if err != nil {
		resp.Diagnostics.AddError("Error building dashboard JSON", err.Error())
		return
	}

	respByte, httpresp, err := utils.SendRequest(r.Auth, r.ApiInstances.HttpClient, "PUT", dashboardmapping.DashboardAPIPath+"/"+dashboardID, &bodyStr)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error updating dashboard"))
		return
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		resp.Diagnostics.AddError("Error parsing dashboard response", err.Error())
		return
	}

	layoutType, _ := respMap["layout_type"].(string)
	r.updateDashboardLists(plan.DashboardLists, plan.DashboardListsRemoved, dashboardID, layoutType)

	diags := r.setStateFromResp(ctx, respMap, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardV2ResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := state.ID.ValueString()
	_, httpresp, err := utils.SendRequest(r.Auth, r.ApiInstances.HttpClient, "DELETE", dashboardmapping.DashboardAPIPath+"/"+dashboardID, nil)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error deleting dashboard"))
		return
	}
}

// ============================================================
// Helpers
// ============================================================

// modelToTopLevelAttrs converts the model struct to a flat map[string]attr.Value
// suitable for passing to dashboardmapping.MarshalDashboardJSON.
func modelToTopLevelAttrs(m *dashboardV2ResourceModel) map[string]attr.Value {
	return map[string]attr.Value{
		"title":                    m.Title,
		"layout_type":              m.LayoutType,
		"reflow_type":              m.ReflowType,
		"description":              m.Description,
		"is_read_only":             m.IsReadOnly,
		"restricted_roles":         m.RestrictedRoles,
		"notify_list":              m.NotifyList,
		"tags":                     m.Tags,
		"template_variable":        m.TemplateVariable,
		"template_variable_preset": m.TemplateVariablePreset,
		"widget":                   m.Widget,
		// SchemaOnly fields - included for completeness but not serialized to JSON
		"url":                     m.URL,
		"dashboard_lists":         m.DashboardLists,
		"dashboard_lists_removed": m.DashboardListsRemoved,
	}
}

// setStringFromResp sets a types.String field from a response map, defaulting to "" if absent.
// This ensures no unknown values remain after apply (framework requires all values to be known).
func setStringFromResp(resp map[string]interface{}, key string) types.String {
	if v, ok := resp[key]; ok && v != nil {
		return types.StringValue(fmt.Sprintf("%v", v))
	}
	return types.StringValue("")
}

// setStateFromResp populates the model from the dashboard API response map.
func (r *dashboardV2Resource) setStateFromResp(ctx context.Context, resp map[string]interface{}, m *dashboardV2ResourceModel) (diags diag.Diagnostics) {
	// Simple string fields (always set to known value - "" if API omits them)
	m.Title = setStringFromResp(resp, "title")
	m.LayoutType = setStringFromResp(resp, "layout_type")
	m.ReflowType = setStringFromResp(resp, "reflow_type")
	m.Description = setStringFromResp(resp, "description")
	if v, ok := resp["url"].(string); ok {
		m.URL = types.StringValue(v)
	} else {
		m.URL = types.StringValue("")
	}

	// is_read_only / restricted_roles
	if restrictedRoles, ok := resp["restricted_roles"].([]interface{}); ok {
		roles := make([]string, len(restrictedRoles))
		for i, r := range restrictedRoles {
			roles[i] = fmt.Sprintf("%v", r)
		}
		rolesList, d := types.ListValueFrom(ctx, types.StringType, roles)
		diags.Append(d...)
		m.RestrictedRoles = rolesList
		m.IsReadOnly = types.BoolValue(false)
	} else {
		isReadOnly := false
		if v, ok := resp["is_read_only"].(bool); ok {
			isReadOnly = v
		}
		m.IsReadOnly = types.BoolValue(isReadOnly)
		// Always set restricted_roles to known value (empty if not in response)
		emptyRoles, d := types.ListValue(types.StringType, []attr.Value{})
		diags.Append(d...)
		m.RestrictedRoles = emptyRoles
	}

	// notify_list
	if v, ok := resp["notify_list"].([]interface{}); ok {
		notifyList := make([]string, len(v))
		for i, n := range v {
			notifyList[i] = fmt.Sprintf("%v", n)
		}
		notifyListVal, d := types.ListValueFrom(ctx, types.StringType, notifyList)
		diags.Append(d...)
		m.NotifyList = notifyListVal
	} else {
		emptyList, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		m.NotifyList = emptyList
	}

	// tags (always set to known value)
	if v, ok := resp["tags"].([]interface{}); ok {
		tags := make([]string, len(v))
		for i, t := range v {
			tags[i] = fmt.Sprintf("%v", t)
		}
		tagsList, d := types.ListValueFrom(ctx, types.StringType, tags)
		diags.Append(d...)
		m.Tags = tagsList
	} else {
		emptyTags, d := types.ListValue(types.StringType, []attr.Value{})
		diags.Append(d...)
		m.Tags = emptyTags
	}

	// template_variable
	var tvAttrTypes map[string]attr.Type
	for _, f := range dashboardmapping.DashboardTopLevelFields {
		if f.HCLKey == "template_variable" {
			tvAttrTypes = dashboardmapping.FieldSpecsToAttrTypes(f.Children)
			break
		}
	}
	if v, ok := resp["template_variables"].([]interface{}); ok {
		flattened := dashboardmapping.FlattenTemplateVariables(v)
		tvList, err := dashboardmapping.FlattenListToFW(flattened, tvAttrTypes)
		if err != nil {
			diags.AddError("Error flattening template variables", err.Error())
		} else {
			m.TemplateVariable = tvList
		}
	}

	// template_variable_preset
	var tvpAttrTypes map[string]attr.Type
	for _, f := range dashboardmapping.DashboardTopLevelFields {
		if f.HCLKey == "template_variable_preset" {
			tvpAttrTypes = dashboardmapping.FieldSpecsToAttrTypes(f.Children)
			break
		}
	}
	if v, ok := resp["template_variable_presets"].([]interface{}); ok {
		flattened := dashboardmapping.FlattenTemplateVariablePresets(v)
		tvpList, err := dashboardmapping.FlattenListToFW(flattened, tvpAttrTypes)
		if err != nil {
			diags.AddError("Error flattening template variable presets", err.Error())
		} else {
			m.TemplateVariablePreset = tvpList
		}
	}

	// widgets
	if v, ok := resp["widgets"].([]interface{}); ok {
		widgetList, err := dashboardmapping.FlattenWidgetsToFW(v, false)
		if err != nil {
			diags.AddError("Error flattening widgets", err.Error())
		} else {
			m.Widget = widgetList
		}
	}

	// dashboard_lists_removed: clear after apply (lists have been removed, no longer pending)
	emptySet, d := types.SetValue(types.Int64Type, []attr.Value{})
	diags.Append(d...)
	m.DashboardListsRemoved = emptySet

	return diags
}

// updateDashboardLists manages dashboard list membership after create/update.
func (r *dashboardV2Resource) updateDashboardLists(dashboardLists, dashboardListsRemoved types.Set, dashboardID, layoutType string) {
	dashTypeString := "custom_screenboard"
	if layoutType == "ordered" {
		dashTypeString = "custom_timeboard"
	}
	dashType := datadogV2.DashboardType(dashTypeString)
	itemsRequest := []datadogV2.DashboardListItemRequest{
		*datadogV2.NewDashboardListItemRequest(dashboardID, dashType),
	}

	if !dashboardLists.IsNull() && !dashboardLists.IsUnknown() {
		for _, elem := range dashboardLists.Elements() {
			if idVal, ok := elem.(types.Int64); ok {
				items := datadogV2.NewDashboardListAddItemsRequest()
				items.SetDashboards(itemsRequest)
				_, _, err := r.ApiInstances.GetDashboardListsApiV2().CreateDashboardListItems(r.Auth, idVal.ValueInt64(), *items)
				if err != nil {
					log.Printf("[DEBUG] Got error adding to dashboard list %d: %v", idVal.ValueInt64(), err)
				}
			}
		}
	}

	if !dashboardListsRemoved.IsNull() && !dashboardListsRemoved.IsUnknown() {
		for _, elem := range dashboardListsRemoved.Elements() {
			if idVal, ok := elem.(types.Int64); ok {
				items := datadogV2.NewDashboardListDeleteItemsRequest()
				items.SetDashboards(itemsRequest)
				_, _, err := r.ApiInstances.GetDashboardListsApiV2().DeleteDashboardListItems(r.Auth, idVal.ValueInt64(), *items)
				if err != nil {
					log.Printf("[DEBUG] Got error removing from dashboard list %d: %v", idVal.ValueInt64(), err)
				}
			}
		}
	}
}

// setToInt64Slice converts a types.Set to []int64.
func setToInt64Slice(s types.Set) []int64 {
	if s.IsNull() || s.IsUnknown() {
		return nil
	}
	result := make([]int64, 0, len(s.Elements()))
	for _, elem := range s.Elements() {
		if iv, ok := elem.(types.Int64); ok {
			result = append(result, iv.ValueInt64())
		}
	}
	return result
}

// retryContextDashboard retries fn up to 3 times with a backoff, returning the last error.
func retryContextDashboard(ctx context.Context, fn func() error) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		if len(err.Error()) > 14 && err.Error()[:14] == "non-retryable:" {
			return err
		}
		lastErr = err
	}
	return lastErr
}
