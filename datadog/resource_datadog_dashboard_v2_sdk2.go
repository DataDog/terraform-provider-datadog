package datadog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// resourceDatadogDashboardV2SDK2 returns the SDKv2 resource for datadog_dashboard_v2_sdk2.
// This is a performance-benchmarking parallel of the framework-based datadog_dashboard_v2.
// It shares all FieldSpec/WidgetSpec declarations via the dashboardmapping package.
func resourceDatadogDashboardV2SDK2() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog dashboard resource (SDKv2 implementation for performance comparison).",
		CreateContext: resourceDatadogDashboardV2SDK2Create,
		ReadContext:   resourceDatadogDashboardV2SDK2Read,
		UpdateContext: resourceDatadogDashboardV2SDK2Update,
		DeleteContext: resourceDatadogDashboardV2SDK2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
			oldValue, newValue := diff.GetChange("dashboard_lists")
			if !oldValue.(*schema.Set).Equal(newValue.(*schema.Set)) {
				removed := oldValue.(*schema.Set).Difference(newValue.(*schema.Set))
				if err := diff.SetNew("dashboard_lists_removed", removed); err != nil {
					return err
				}
			} else {
				if err := diff.Clear("dashboard_lists_removed"); err != nil {
					return err
				}
			}

			// Validate ConflictsWith constraints on widget request fields (e.g., "q" vs "query"/"formula")
			widgetData := map[string]interface{}{"widget": diff.Get("widget")}
			if errs := dashboardmapping.ValidateWidgetConflicts(widgetData); len(errs) > 0 {
				return fmt.Errorf("%s", strings.Join(errs, "\n"))
			}

			return nil
		},
		SchemaFunc: buildDashboardV2SDK2Schema,
	}
}

// buildDashboardV2SDK2Schema builds the schema map for datadog_dashboard_v2_sdk2.
// Derives all fields from shared FieldSpec/WidgetSpec declarations.
func buildDashboardV2SDK2Schema() map[string]*schema.Schema {
	// Generate top-level fields from FieldSpec declarations (excluding SchemaOnly)
	topSchema := dashboardmapping.FieldSpecsToSDKv2Schema(dashboardmapping.DashboardTopLevelFields)

	// Override url to be Computed+Optional with diff suppression (like v1)
	topSchema["url"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "The URL of the dashboard.",
		DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool {
			return true
		},
	}

	// Add widget block with all widget types
	widgetSchema := dashboardmapping.AllWidgetSDKv2Schema(false)
	topSchema["widget"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "The list of widgets to display on the dashboard.",
		Elem: &schema.Resource{
			Schema: widgetSchema,
		},
	}

	return topSchema
}

// resourceDatadogDashboardV2SDK2Create creates a new dashboard via the Datadog API.
func resourceDatadogDashboardV2SDK2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	data := collectDashboardData(d)
	bodyStr, err := dashboardmapping.MarshalDashboardJSONFromMap(data, "")
	if err != nil {
		return diag.Errorf("failed to build dashboard JSON: %s", err)
	}

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", dashboardmapping.DashboardAPIPath, &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.Errorf("error parsing dashboard response: %s", err)
	}

	id, ok := respMap["id"]
	if !ok {
		return diag.Errorf("error retrieving dashboard ID: id not found in response")
	}
	dashboardID := fmt.Sprintf("%v", id)
	d.SetId(dashboardID)

	layoutType, _ := respMap["layout_type"].(string)

	// Retry GET until the dashboard is available
	var httpResponse *http.Response
	retryErr := retryDashboardV2SDK2(ctx, func() error {
		_, httpResponse, err = utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+dashboardID, nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return fmt.Errorf("dashboard not created yet")
			}
			return fmt.Errorf("non-retryable error: %w", err)
		}
		return nil
	})
	if retryErr != nil {
		return diag.Errorf("error waiting for dashboard: %s", retryErr)
	}

	// Update dashboard lists (side-effect)
	updateDashboardListsSDKv2(d, providerConf, dashboardID, layoutType)

	return resourceDatadogDashboardV2SDK2Read(ctx, d, meta)
}

// resourceDatadogDashboardV2SDK2Read reads a dashboard from the Datadog API and sets state.
func resourceDatadogDashboardV2SDK2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	dashboardID := d.Id()
	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+dashboardID, nil)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error reading dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.Errorf("error parsing dashboard response: %s", err)
	}

	return setDashboardStateSDKv2(d, respMap)
}

// resourceDatadogDashboardV2SDK2Update updates an existing dashboard via the Datadog API.
func resourceDatadogDashboardV2SDK2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	dashboardID := d.Id()
	data := collectDashboardData(d)
	bodyStr, err := dashboardmapping.MarshalDashboardJSONFromMap(data, dashboardID)
	if err != nil {
		return diag.Errorf("failed to build dashboard JSON: %s", err)
	}

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "PUT", dashboardmapping.DashboardAPIPath+"/"+dashboardID, &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.Errorf("error parsing dashboard response: %s", err)
	}

	layoutType, _ := respMap["layout_type"].(string)
	updateDashboardListsSDKv2(d, providerConf, dashboardID, layoutType)

	return setDashboardStateSDKv2(d, respMap)
}

// resourceDatadogDashboardV2SDK2Delete deletes a dashboard from the Datadog API.
func resourceDatadogDashboardV2SDK2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	dashboardID := d.Id()
	_, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", dashboardmapping.DashboardAPIPath+"/"+dashboardID, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard")
	}
	return nil
}

// ============================================================
// Helpers
// ============================================================

// collectDashboardData collects all relevant fields from ResourceData into a plain map
// suitable for passing to MarshalDashboardJSONFromMap.
func collectDashboardData(d *schema.ResourceData) map[string]interface{} {
	data := make(map[string]interface{})
	for _, f := range dashboardmapping.DashboardTopLevelFields {
		if v, ok := d.GetOk(f.HCLKey); ok {
			data[f.HCLKey] = v
		} else {
			// Include zero values for required fields
			data[f.HCLKey] = d.Get(f.HCLKey)
		}
	}
	data["widget"] = d.Get("widget")
	return data
}

// setDashboardStateSDKv2 populates ResourceData from the dashboard API response map.
func setDashboardStateSDKv2(d *schema.ResourceData, resp map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Simple string fields
	for _, key := range []string{"title", "layout_type", "reflow_type", "description", "url"} {
		if v, ok := resp[key]; ok && v != nil {
			if err := d.Set(key, fmt.Sprintf("%v", v)); err != nil {
				diags = append(diags, diag.FromErr(err)...)
			}
		}
	}

	// is_read_only / restricted_roles
	if restrictedRoles, ok := resp["restricted_roles"].([]interface{}); ok {
		roles := make([]string, len(restrictedRoles))
		for i, r := range restrictedRoles {
			roles[i] = fmt.Sprintf("%v", r)
		}
		if err := d.Set("restricted_roles", roles); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
		if err := d.Set("is_read_only", false); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		isReadOnly := false
		if v, ok := resp["is_read_only"].(bool); ok {
			isReadOnly = v
		}
		if err := d.Set("is_read_only", isReadOnly); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// notify_list
	if v, ok := resp["notify_list"].([]interface{}); ok {
		notifyList := make([]string, len(v))
		for i, n := range v {
			notifyList[i] = fmt.Sprintf("%v", n)
		}
		if err := d.Set("notify_list", notifyList); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// tags
	if v, ok := resp["tags"].([]interface{}); ok {
		tags := make([]string, len(v))
		for i, t := range v {
			tags[i] = fmt.Sprintf("%v", t)
		}
		if err := d.Set("tags", tags); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// template_variable
	if v, ok := resp["template_variables"].([]interface{}); ok {
		flattened := dashboardmapping.FlattenTemplateVariables(v)
		if err := d.Set("template_variable", flattened); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// template_variable_preset
	if v, ok := resp["template_variable_presets"].([]interface{}); ok {
		flattened := dashboardmapping.FlattenTemplateVariablePresets(v)
		if err := d.Set("template_variable_preset", flattened); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// widgets
	if v, ok := resp["widgets"].([]interface{}); ok {
		flatWidgets := dashboardmapping.FlattenWidgetsForSDKv2(v)
		if err := d.Set("widget", flatWidgets); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	// dashboard_lists_removed: clear after apply
	if err := d.Set("dashboard_lists_removed", []int{}); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

// updateDashboardListsSDKv2 manages dashboard list membership after create/update.
// Mirrors updateDashboardLists from resource_datadog_dashboard.go.
func updateDashboardListsSDKv2(d *schema.ResourceData, providerConf *ProviderConfiguration, dashboardID, layoutType string) {
	dashTypeString := "custom_screenboard"
	if layoutType == "ordered" {
		dashTypeString = "custom_timeboard"
	}
	dashType := datadogV2.DashboardType(dashTypeString)
	itemsRequest := []datadogV2.DashboardListItemRequest{*datadogV2.NewDashboardListItemRequest(dashboardID, dashType)}
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if v, ok := d.GetOk("dashboard_lists"); ok && v.(*schema.Set).Len() > 0 {
		items := datadogV2.NewDashboardListAddItemsRequest()
		items.SetDashboards(itemsRequest)
		for _, id := range v.(*schema.Set).List() {
			_, _, err := apiInstances.GetDashboardListsApiV2().CreateDashboardListItems(auth, int64(id.(int)), *items)
			if err != nil {
				log.Printf("[DEBUG] Got error adding to dashboard list %d: %v", id.(int), err)
			}
		}
	}

	if v, ok := d.GetOk("dashboard_lists_removed"); ok && v.(*schema.Set).Len() > 0 {
		items := datadogV2.NewDashboardListDeleteItemsRequest()
		items.SetDashboards(itemsRequest)
		for _, id := range v.(*schema.Set).List() {
			_, _, err := apiInstances.GetDashboardListsApiV2().DeleteDashboardListItems(auth, int64(id.(int)), *items)
			if err != nil {
				log.Printf("[DEBUG] Got error removing from dashboard list %d: %v", id.(int), err)
			}
		}
	}
}

// retryDashboardV2SDK2 retries fn up to 3 times with a simple loop, returning the last error.
func retryDashboardV2SDK2(ctx context.Context, fn func() error) error {
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
