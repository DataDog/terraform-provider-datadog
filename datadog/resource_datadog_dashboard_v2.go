package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const dashboardWidgetValidationPath = "/api/v1/dashboard/widgets/validate"

// resourceDatadogDashboardV2 returns the SDKv2 resource for datadog_dashboard_v2.
// It shares all FieldSpec/WidgetSpec declarations via the dashboardmapping package.
func resourceDatadogDashboardV2() *schema.Resource {
	return &schema.Resource{
		Description:   "[BETA] Provides an updated version of the Datadog dashboard resource which improves compliance with Datadog's dashboard API spec. This version is currently experimental and prone to changes.",
		CreateContext: resourceDatadogDashboardV2Create,
		ReadContext:   resourceDatadogDashboardV2Read,
		UpdateContext: resourceDatadogDashboardV2Update,
		DeleteContext: resourceDatadogDashboardV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

			return resourceDatadogDashboardV2ValidateWidgets(ctx, diff, meta)
		},
		SchemaFunc: buildDashboardV2Schema,
	}
}

// buildDashboardV2Schema builds the schema map for datadog_dashboard_v2_sdk2.
// Derives all fields from shared FieldSpec/WidgetSpec declarations.
func buildDashboardV2Schema() map[string]*schema.Schema {
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
	topSchema["validate"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Whether to send widgets to the Datadog API to validate widget configuration and query values during `terraform plan`. Defaults to `true`. Setting this to `false` skips only the Datadog API validation; local Terraform schema and conflicting-field checks still run.",
		DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool {
			// This provider-only setting is never sent to the backend.
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

type dashboardWidget struct {
	terraformPath string
	widget        map[string]interface{}
}

type dashboardWidgetValidationResult struct {
	IsValid      bool    `json:"is_valid"`
	WidgetType   *string `json:"widget_type"`
	ErrorMessage *string `json:"error_message"`
	ErrorPath    *string `json:"error_path"`
}

// resourceDatadogDashboardV2ValidateWidgets validates widget definitions during
// planning. Like monitor validation, callers can explicitly opt out.
func resourceDatadogDashboardV2ValidateWidgets(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if validate, ok := diff.GetOkExists("validate"); ok && !validate.(bool) {
		return nil
	}
	widgets, ok := diff.GetOk("widget")
	if !ok {
		return nil
	}
	validationWidgets := flattenDashboardWidgets(widgets)
	if len(validationWidgets) == 0 {
		return nil
	}
	layoutType, _ := diff.Get("layout_type").(string)
	reflowType, _ := diff.Get("reflow_type").(string)

	providerConf := meta.(*ProviderConfiguration)
	return validateDashboardWidgets(
		ctx,
		providerConf.Auth,
		providerConf.DatadogApiInstances.HttpClient,
		validationWidgets,
		layoutType,
		reflowType,
	)
}

// flattenDashboardWidgets converts the SDKv2 widget representation to API
// widgets. Group children are validated separately so errors point to the
// specific nested widget.
func flattenDashboardWidgets(widgets interface{}) []dashboardWidget {
	widgetList, ok := widgets.([]interface{})
	if !ok {
		return nil
	}
	validationWidgets := make([]dashboardWidget, 0, len(widgetList))
	for i, rawWidget := range widgetList {
		widget, ok := rawWidget.(map[string]interface{})
		if !ok {
			continue
		}
		widgetJSON := dashboardmapping.BuildWidgetEngineJSONFromMap(widget)
		definition, ok := widgetJSON["definition"].(map[string]interface{})
		if !ok {
			continue
		}
		widgetType, _ := definition["type"].(string)
		definitionKey := dashboardmapping.WidgetDefinitionHCLKey(widgetType)
		terraformPath := fmt.Sprintf("widget.%d.%s.0", i, definitionKey)
		appendDashboardWidget(&validationWidgets, terraformPath, widgetJSON)
	}
	return validationWidgets
}

func appendDashboardWidget(widgets *[]dashboardWidget, terraformPath string, widget map[string]interface{}) {
	definition, ok := widget["definition"].(map[string]interface{})
	if !ok {
		return
	}
	if definition["type"] != "group" {
		*widgets = append(*widgets, dashboardWidget{terraformPath: terraformPath, widget: widget})
		return
	}

	groupWidget := make(map[string]interface{}, len(widget))
	for key, value := range widget {
		groupWidget[key] = value
	}
	groupDefinition := make(map[string]interface{}, len(definition))
	for key, value := range definition {
		groupDefinition[key] = value
	}
	groupDefinition["widgets"] = []interface{}{}
	groupWidget["definition"] = groupDefinition
	*widgets = append(*widgets, dashboardWidget{terraformPath: terraformPath, widget: groupWidget})

	children, ok := definition["widgets"].([]interface{})
	if !ok {
		return
	}
	for i, rawChild := range children {
		child, ok := rawChild.(map[string]interface{})
		if !ok {
			continue
		}
		childDefinition, ok := child["definition"].(map[string]interface{})
		if !ok {
			continue
		}
		childType, _ := childDefinition["type"].(string)
		childDefinitionKey := dashboardmapping.WidgetDefinitionHCLKey(childType)
		childTerraformPath := fmt.Sprintf("%s.widget.%d.%s.0", terraformPath, i, childDefinitionKey)
		appendDashboardWidget(widgets, childTerraformPath, child)
	}
}

func validateDashboardWidgets(
	ctx context.Context,
	auth context.Context,
	client *datadog.APIClient,
	validationWidgets []dashboardWidget,
	layoutType string,
	reflowType string,
) error {
	widgets := make([]map[string]interface{}, len(validationWidgets))
	for i, validationWidget := range validationWidgets {
		widgets[i] = validationWidget.widget
	}
	body := map[string]interface{}{
		"widgets":     widgets,
		"layout_type": layoutType,
	}
	if reflowType != "" {
		body["reflow_type"] = reflowType
	}

	return retry.RetryContext(ctx, retryTimeout, func() *retry.RetryError {
		responseBody, httpResponse, err := utils.SendRequest(auth, client, http.MethodPost, dashboardWidgetValidationPath, &body)
		if err != nil {
			translatedErr := utils.TranslateClientError(err, httpResponse, "error validating dashboard widgets")
			if httpResponse != nil && (httpResponse.StatusCode == http.StatusBadGateway || httpResponse.StatusCode == http.StatusGatewayTimeout) {
				return retry.RetryableError(translatedErr)
			}
			return retry.NonRetryableError(translatedErr)
		}

		var response struct {
			Results []dashboardWidgetValidationResult `json:"results"`
		}
		if err := json.Unmarshal(responseBody, &response); err != nil {
			return retry.NonRetryableError(fmt.Errorf("error parsing dashboard widget validation response: %w", err))
		}
		if len(response.Results) != len(validationWidgets) {
			return retry.NonRetryableError(fmt.Errorf(
				"dashboard widget validation response contained %d results for %d widgets",
				len(response.Results),
				len(validationWidgets),
			))
		}

		failures := make([]string, 0)
		for i, result := range response.Results {
			if result.IsValid {
				continue
			}
			message := "invalid widget definition"
			if result.ErrorMessage != nil && *result.ErrorMessage != "" {
				message = *result.ErrorMessage
			}
			widgetType := "unknown"
			if result.WidgetType != nil && *result.WidgetType != "" {
				widgetType = *result.WidgetType
			}
			terraformPath := validationWidgets[i].terraformPath
			if result.ErrorPath != nil && *result.ErrorPath != "" {
				mappedPath := dashboardmapping.WidgetErrorPathToHCL(widgetType, *result.ErrorPath)
				if mappedPath != "" {
					terraformPath += "." + mappedPath
				}
			}
			failures = append(failures, fmt.Sprintf("%s (%s): %s", terraformPath, widgetType, message))
		}
		if len(failures) > 0 {
			return retry.NonRetryableError(fmt.Errorf("dashboard widget validation failed:\n%s", strings.Join(failures, "\n")))
		}

		return nil
	})
}

// resourceDatadogDashboardV2Create creates a new dashboard via the Datadog API.
func resourceDatadogDashboardV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Update dashboard lists (side-effect)
	updateDashboardListsSDKv2(d, providerConf, dashboardID, layoutType)

	return resourceDatadogDashboardV2Read(ctx, d, meta)
}

// resourceDatadogDashboardV2Read reads a dashboard from the Datadog API and sets state.
func resourceDatadogDashboardV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

// resourceDatadogDashboardV2Update updates an existing dashboard via the Datadog API.
func resourceDatadogDashboardV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

// resourceDatadogDashboardV2Delete deletes a dashboard from the Datadog API.
func resourceDatadogDashboardV2Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	var apiWidgets []interface{}
	if v, ok := resp["widgets"].([]interface{}); ok {
		apiWidgets = v
		flatWidgets, dropped := dashboardmapping.FlattenWidgetsForSDKv2(v)
		if err := d.Set("widget", flatWidgets); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
		if len(dropped) > 0 {
			for _, p := range dropped {
				log.Printf("[WARN] datadog_dashboard_v2 %s: dropped unmapped API field at %s", d.Id(), p)
			}
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Dashboard contains %d field(s) returned by the Datadog API with no schema mapping", len(dropped)),
				Detail: fmt.Sprintf(
					"The following paths were stripped from state because the schema does not declare them; their values will not be preserved on apply.\n  %s\nIf any of these are fields you need, file an issue against the provider.",
					strings.Join(dropped, "\n  "),
				),
			})
		}
	}

	// tabs — flatten after widgets so we can reverse-map widget IDs to @N references
	if v, ok := resp["tabs"].([]interface{}); ok && len(v) > 0 {
		flatTabs := dashboardmapping.FlattenTabs(v, apiWidgets)
		if err := d.Set("tab", flatTabs); err != nil {
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
