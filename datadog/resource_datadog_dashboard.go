package datadog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/dashboardmapping"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogDashboard() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog dashboard resource. This can be used to create and manage Datadog dashboards.\n\n!> The `is_read_only` field is deprecated and non-functional. Use `restricted_roles` instead to define which roles are required to edit the dashboard.",
		CreateContext: resourceDatadogDashboardCreate,
		UpdateContext: resourceDatadogDashboardUpdate,
		ReadContext:   resourceDatadogDashboardRead,
		DeleteContext: resourceDatadogDashboardDelete,
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			oldValue, newValue := diff.GetChange("dashboard_lists")
			if !oldValue.(*schema.Set).Equal(newValue.(*schema.Set)) {
				// Only calculate removed when the list change, to no create useless diffs
				removed := oldValue.(*schema.Set).Difference(newValue.(*schema.Set))
				if err := diff.SetNew("dashboard_lists_removed", removed); err != nil {
					return err
				}
			} else {
				if err := diff.Clear("dashboard_lists_removed"); err != nil {
					return err
				}
			}

			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"title": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The title of the dashboard.",
				},
				"widget": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of widgets to display on the dashboard.",
					Elem: &schema.Resource{
						Schema: getWidgetSchema(),
					},
				},
				"layout_type": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					Description:      "The layout type of the dashboard.",
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewDashboardLayoutTypeFromValue),
				},
				"reflow_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Description:      "The reflow type of a new dashboard layout. Set this only when layout type is `ordered`. If set to `fixed`, the dashboard expects all widgets to have a layout, and if it's set to `auto`, widgets should not have layouts.",
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewDashboardReflowTypeFromValue),
				},
				"description": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The description of the dashboard.",
				},
				"url": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "The URL of the dashboard.",
					DiffSuppressFunc: func(_, _, _ string, _ *schema.ResourceData) bool {
						// This value is computed and cannot be updated.
						// To maintain backward compatibility, always suppress diff rather
						// than converting the attribute to `Computed` only
						return true
					},
				},
				"restricted_roles": {
					Type:          schema.TypeSet,
					Optional:      true,
					Elem:          &schema.Schema{Type: schema.TypeString},
					ConflictsWith: []string{"is_read_only"},
					Description:   "UUIDs of roles whose associated users are authorized to edit the dashboard.",
				},
				"template_variable": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of template variables for this dashboard.",
					Elem: &schema.Resource{
						Schema: getTemplateVariableSchema(),
					},
				},
				"template_variable_preset": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The list of selectable template variable presets for this dashboard.",
					Elem: &schema.Resource{
						Schema: getTemplateVariablePresetSchema(),
					},
				},
				"notify_list": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "The list of handles for the users to notify when changes are made to this dashboard.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"dashboard_lists": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "A list of dashboard lists this dashboard belongs to. This attribute should not be set if managing the corresponding dashboard lists using Terraform as it causes inconsistent behavior.",
					Elem:        &schema.Schema{Type: schema.TypeInt},
				},
				"dashboard_lists_removed": {
					Type:        schema.TypeSet,
					Computed:    true,
					Description: "A list of dashboard lists this dashboard should be removed from. Internal only.",
					Elem:        &schema.Schema{Type: schema.TypeInt},
				},
				"is_read_only": {
					Type:          schema.TypeBool,
					Optional:      true,
					Default:       false,
					ConflictsWith: []string{"restricted_roles"},
					Description:   "Whether this dashboard is read-only.",
					Deprecated:    "This field is deprecated and non-functional. Use `restricted_roles` instead to define which roles are required to edit the dashboard.",
				},
				"tags": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    5,
					Description: "A list of tags assigned to the Dashboard. Only team names of the form `team:<name>` are supported.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

// resourceDatadogDashboardCreate, resourceDatadogDashboardRead,
// resourceDatadogDashboardUpdate, and resourceDatadogDashboardDelete implement
// CRUD for the dashboard resource using the dashboardmapping engine.

func resourceDatadogDashboardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	bodyStr, err := dashboardmapping.MarshalDashboardJSON(d)
	if err != nil {
		return diag.FromErr(err)
	}

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", dashboardmapping.DashboardAPIPath, &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	id, ok := respMap["id"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving id from response"))
	}
	d.SetId(fmt.Sprintf("%v", id))

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	var httpResponse *http.Response
	retryErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		_, httpResponse, err = utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+d.Id(), nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("dashboard not created yet"))
			}
			return retry.NonRetryableError(err)
		}
		// We only log the error, as failing to update the list shouldn't fail dashboard creation
		updateDashboardLists(d, providerConf, d.Id(), fmt.Sprintf("%v", layoutType))
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}

	return dashboardmapping.UpdateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardmapping.DashboardAPIPath+"/"+d.Id(), nil)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	return dashboardmapping.UpdateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	bodyStr, err := dashboardmapping.MarshalDashboardJSON(d)
	if err != nil {
		return diag.FromErr(err)
	}

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "PUT", dashboardmapping.DashboardAPIPath+"/"+d.Id(), &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	updateDashboardLists(d, providerConf, d.Id(), fmt.Sprintf("%v", layoutType))

	return dashboardmapping.UpdateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	_, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", dashboardmapping.DashboardAPIPath+"/"+d.Id(), nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard")
	}

	return nil
}

// isWidgetTimeUnparsedObject checks if an unparsed object is a WidgetTime.
// WidgetTime can be unparsed due to oneOf ambiguity between legacy and new formats.
func isWidgetTimeUnparsedObject(obj interface{}) bool {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return false
	}

	if typeVal, hasType := m["type"]; hasType {
		switch typeVal {
		case "live":
			_, hasUnit := m["unit"]
			_, hasValue := m["value"]
			return hasUnit && hasValue
		case "fixed":
			_, hasFrom := m["from"]
			_, hasTo := m["to"]
			return hasFrom && hasTo
		}
	}

	// Check for WidgetLegacyLiveSpan signature: live_span (and optionally hide_incomplete_cost_data)
	// This is less specific, so only match if it has live_span and no other unexpected fields
	if _, hasLiveSpan := m["live_span"]; hasLiveSpan {
		// Should only have live_span and optionally hide_incomplete_cost_data
		for key := range m {
			if key != "live_span" && key != "hide_incomplete_cost_data" {
				return false // Has other fields, not a legacy live span
			}
		}
		return true
	}

	return false
}

func updateDashboardLists(d *schema.ResourceData, providerConf *ProviderConfiguration, dashboardID string, layoutType string) {
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

//
// Template Variable helpers
//

func getPpkTemplateVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"controlled_externally": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Template variables controlled by the external resource, such as the dashboard this powerpack is on.",
			Elem: &schema.Resource{
				Schema: getPpkTemplateVariableContentSchema(),
			},
		},
		"controlled_by_powerpack": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Template variables controlled at the powerpack level.",
			Elem: &schema.Resource{
				Schema: getPpkTemplateVariableContentSchema(),
			},
		},
	}
}

func getPpkTemplateVariableContentSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the variable.",
		},
		"prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The tag prefix associated with the variable. Only tags with this prefix appear in the variable dropdown.",
		},
		"values": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			Description: "One or many template variable values within the saved view, which will be unioned together using `OR` if more than one is specified.",
		},
	}
}

func getTemplateVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the variable.",
		},
		"prefix": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The tag prefix associated with the variable. Only tags with this prefix appear in the variable dropdown.",
		},
		"default": {
			Type:        schema.TypeString,
			Optional:    true,
			Deprecated:  "Use `defaults` instead.",
			Description: "The default value for the template variable on dashboard load. Cannot be used in conjunction with `defaults`.",
		},
		"defaults": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			Description: "One or many default values for template variables on load. If more than one default is specified, they will be unioned together with `OR`. Cannot be used in conjunction with `default`.",
		},
		"available_values": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "The list of values that the template variable drop-down is be limited to",
		},
	}
}

func buildDatadogPowerpackTVarContents(contents []interface{}) []datadogV1.PowerpackTemplateVariableContents {
	tVarContents := make([]datadogV1.PowerpackTemplateVariableContents, len(contents))
	for ind, tvp := range contents {
		typecastTvp := tvp.(map[string]interface{})
		tvar := datadogV1.NewPowerpackTemplateVariableContentsWithDefaults()
		if name, ok := typecastTvp["name"].(string); ok {
			tvar.SetName(name)
		}
		if v, ok := typecastTvp["values"].([]interface{}); ok && len(v) != 0 {
			var values []string
			for _, s := range v {
				values = append(values, s.(string))
			}
			tvar.SetValues(values)
		}
		if prefix, ok := typecastTvp["prefix"].(string); ok {
			tvar.SetPrefix(prefix)
		}
		tVarContents[ind] = *tvar
	}
	return tVarContents
}

func buildTerraformPowerpackTVarContents(tVarContents []datadogV1.PowerpackTemplateVariableContents) []map[string]interface{} {
	ppkTvarContents := make([]map[string]interface{}, len(tVarContents))
	for i, templateVariable := range tVarContents {
		terraformTemplateVariable := map[string]interface{}{}
		if v, ok := templateVariable.GetNameOk(); ok {
			terraformTemplateVariable["name"] = *v
		}
		if v := templateVariable.GetPrefix(); len(v) > 0 {
			terraformTemplateVariable["prefix"] = v
		}
		if v, ok := templateVariable.GetValuesOk(); ok && len(*v) > 0 {
			var tags []string
			tags = append(tags, *v...)
			terraformTemplateVariable["values"] = tags
		}
		ppkTvarContents[i] = terraformTemplateVariable
	}
	return ppkTvarContents
}

//
// Template Variable Preset Helpers
//

func getTemplateVariablePresetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the preset.",
		},
		"template_variable": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The template variable names and assumed values under the given preset",
			Elem: &schema.Resource{
				Schema: getTemplateVariablePresetValueSchema(),
			},
		},
	}
}

func getTemplateVariablePresetValueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the template variable",
			Optional:    true,
		},
		"value": {
			Type:        schema.TypeString,
			Description: "The value that should be assumed by the template variable in this preset. Cannot be used in conjunction with `values`.",
			Optional:    true,
			Deprecated:  "Use `values` instead.",
		},
		"values": {
			Type:     schema.TypeList,
			Optional: true,
			MinItems: 1,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			Description: "One or many template variable values within the saved view, which will be unioned together using `OR` if more than one is specified. Cannot be used in conjunction with `value`.",
		},
	}
}

//
// Restricted Roles helpers
//

func buildTerraformRestrictedRoles(datadogRestrictedRoles *[]string) *[]string {
	if datadogRestrictedRoles == nil {
		terraformRestrictedRoles := make([]string, 0)
		return &terraformRestrictedRoles
	}
	terraformRestrictedRoles := make([]string, len(*datadogRestrictedRoles))
	for i, roleUUID := range *datadogRestrictedRoles {
		terraformRestrictedRoles[i] = roleUUID
	}
	return &terraformRestrictedRoles
}

//
// Widget helpers
//

// The generic widget schema is a combination of the schema for a non-group widget
// and the schema for a Group Widget (which can contains only non-group widgets)
func getWidgetSchema() map[string]*schema.Schema {
	widgetSchema := getNonGroupWidgetSchema(false)
	// Build the group_definition schema from GroupWidgetSpec and inject the widget sub-schema.
	groupSchema := dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.GroupWidgetSpec)
	groupSchema.Elem.(*schema.Resource).Schema["widget"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "The list of widgets in this group.",
		Elem: &schema.Resource{
			Schema: getNonGroupWidgetSchema(false),
		},
	}
	widgetSchema["group_definition"] = groupSchema
	return widgetSchema
}

func getNonGroupWidgetSchema(isPowerpackSchema bool) map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"widget_layout": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "The layout of the widget on a 'free' dashboard.",
			Elem: &schema.Resource{
				Schema: getWidgetLayoutSchema(),
			},
		},
		"id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the widget.",
		},
		// A widget should implement exactly one of the following definitions
		"alert_graph_definition":              dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.AlertGraphWidgetSpec),
		"alert_value_definition":             dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.AlertValueWidgetSpec),
		"change_definition":                  dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ChangeWidgetSpec),
		"check_status_definition":            dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.CheckStatusWidgetSpec),
		"distribution_definition":            dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.DistributionWidgetSpec),
		"event_stream_definition":            dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.EventStreamWidgetSpec),
		"event_timeline_definition":          dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.EventTimelineWidgetSpec),
		"free_text_definition":               dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.FreeTextWidgetSpec),
		"heatmap_definition":                 dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.HeatmapWidgetSpec),
		"hostmap_definition":                 dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.HostmapWidgetSpec),
		"iframe_definition":                  dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.IFrameWidgetSpec),
		"image_definition":                   dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ImageWidgetSpec),
		"list_stream_definition":             dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ListStreamWidgetSpec),
		"log_stream_definition":              dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.LogStreamWidgetSpec),
		"manage_status_definition":           dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ManageStatusWidgetSpec),
		"note_definition":                    dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.NoteWidgetSpec),
		"query_value_definition":             dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.QueryValueWidgetSpec),
		"query_table_definition":             dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.QueryTableWidgetSpec),
		"scatterplot_definition":             dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ScatterplotWidgetSpec),
		"servicemap_definition":              dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ServiceMapWidgetSpec),
		"service_level_objective_definition": dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.SLOWidgetSpec),
		"slo_list_definition":                dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.SLOListWidgetSpec),
		"sunburst_definition":                dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.SunburstWidgetSpec),
		"timeseries_definition":              dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.TimeseriesWidgetSpec),
		"toplist_definition":                 dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.ToplistWidgetSpec),
		"topology_map_definition":            dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.TopologyMapWidgetSpec),
		"trace_service_definition":           dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.TraceServiceWidgetSpec),
		"treemap_definition":                 dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.TreemapWidgetSpec),
		"geomap_definition":                  dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.GeomapWidgetSpec),
		"run_workflow_definition":            dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.RunWorkflowWidgetSpec),
	}

	// Non powerpack specific widgets
	if !isPowerpackSchema {
		s["powerpack_definition"] = dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.PowerpackWidgetSpec)
		s["split_graph_definition"] = dashboardmapping.WidgetSpecToSchemaBlock(dashboardmapping.SplitGraphWidgetSpec)
	}

	return s
}

func buildDatadogSourceWidgetDefinition(terraformWidget map[string]interface{}) (*datadogV1.SplitGraphSourceWidgetDefinition, error) {
	// Build widget Definition
	var definition datadogV1.SplitGraphSourceWidgetDefinition
	sourceWidgetCount := 0
	if def, ok := terraformWidget["change_definition"].([]interface{}); ok && len(def) > 0 {
		if changeDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ChangeWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogChangeDefinition(changeDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}

	if def, ok := terraformWidget["query_value_definition"].([]interface{}); ok && len(def) > 0 {
		if queryValueDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.QueryValueWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogQueryValueDefinition(queryValueDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["query_table_definition"].([]interface{}); ok && len(def) > 0 {

		if queryTableDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TableWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogQueryTableDefinition(queryTableDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["scatterplot_definition"].([]interface{}); ok && len(def) > 0 {
		if scatterplotDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ScatterPlotWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogScatterplotDefinition(scatterplotDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["sunburst_definition"].([]interface{}); ok && len(def) > 0 {
		if sunburstDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.SunburstWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogSunburstDefinition(sunburstDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["timeseries_definition"].([]interface{}); ok && len(def) > 0 {

		if timeseriesDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TimeseriesWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogTimeseriesDefinition(timeseriesDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["toplist_definition"].([]interface{}); ok && len(def) > 0 {

		if toplistDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ToplistWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogToplistDefinition(toplistDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["treemap_definition"].([]interface{}); ok && len(def) > 0 {

		if treemapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TreeMapWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogTreemapDefinition(treemapDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if def, ok := terraformWidget["geomap_definition"].([]interface{}); ok && len(def) > 0 {

		if geomapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.GeomapWidgetDefinitionAsSplitGraphSourceWidgetDefinition(buildDatadogGeomapDefinition(geomapDefinition))
			sourceWidgetCount += 1
		}
		if sourceWidgetCount > 1 {
			return nil, fmt.Errorf("source widget definition must contain exactly one value")
		}
	}
	if sourceWidgetCount == 0 {
		return nil, fmt.Errorf("failed to find valid definition in widget configuration")
	}

	return &definition, nil
}

// Helper to build a source widget definition for terraform
func buildTerraformSourceWidgetDefinition(datadogSourceWidgetDefinition *datadogV1.SplitGraphSourceWidgetDefinition) (map[string]interface{}, error) {
	terraformWidgetDefinition := map[string]interface{}{}

	// Build definition
	if datadogSourceWidgetDefinition.ChangeWidgetDefinition != nil {
		terraformDefinition := buildTerraformChangeDefinition(datadogSourceWidgetDefinition.ChangeWidgetDefinition)
		terraformWidgetDefinition["change_definition"] = terraformDefinition
	} else if datadogSourceWidgetDefinition.QueryValueWidgetDefinition != nil {
		terraformDefinition := buildTerraformQueryValueDefinition(datadogSourceWidgetDefinition.QueryValueWidgetDefinition)
		terraformWidgetDefinition["query_value_definition"] = []map[string]interface{}{terraformDefinition}
	} else if datadogSourceWidgetDefinition.TableWidgetDefinition != nil {
		terraformDefinition := buildTerraformQueryTableDefinition(datadogSourceWidgetDefinition.TableWidgetDefinition)
		terraformWidgetDefinition["query_table_definition"] = []map[string]interface{}{terraformDefinition}
	} else if datadogSourceWidgetDefinition.SunburstWidgetDefinition != nil {
		terraformDefinition := buildTerraformSunburstDefinition(datadogSourceWidgetDefinition.SunburstWidgetDefinition)
		terraformWidgetDefinition["sunburst_definition"] = []map[string]interface{}{terraformDefinition}
	} else if datadogSourceWidgetDefinition.TimeseriesWidgetDefinition != nil {
		terraformDefinition := buildTerraformTimeseriesDefinition(datadogSourceWidgetDefinition.TimeseriesWidgetDefinition)
		terraformWidgetDefinition["timeseries_definition"] = []map[string]interface{}{terraformDefinition}
	} else if datadogSourceWidgetDefinition.ToplistWidgetDefinition != nil {
		terraformDefinition := buildTerraformToplistDefinition(datadogSourceWidgetDefinition.ToplistWidgetDefinition)
		terraformWidgetDefinition["toplist_definition"] = []map[string]interface{}{terraformDefinition}
	} else if datadogSourceWidgetDefinition.TreeMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformTreemapDefinition(datadogSourceWidgetDefinition.TreeMapWidgetDefinition)
		terraformWidgetDefinition["treemap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if datadogSourceWidgetDefinition.GeomapWidgetDefinition != nil {
		terraformDefinition := buildTerraformGeomapDefinition(datadogSourceWidgetDefinition.GeomapWidgetDefinition)
		terraformWidgetDefinition["geomap_definition"] = []map[string]interface{}{terraformDefinition}
	} else {
		return nil, fmt.Errorf("unsupported widget type used as split graph source widget: %s", datadogSourceWidgetDefinition.GetActualInstance())
	}
	return terraformWidgetDefinition, nil
}

//
// Widget Layout helpers
//

func getWidgetLayoutSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"x": {
			Description: "The position of the widget on the x (horizontal) axis. Must be greater than or equal to 0.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"y": {
			Description: "The position of the widget on the y (vertical) axis. Must be greater than or equal to 0.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"width": {
			Description: "The width of the widget.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"height": {
			Description: "The height of the widget.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"is_column_break": {
			Description: "Whether the widget should be the first one on the second column in high density or not. Only one widget in the dashboard should have this property set to `true`.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}
}

func buildDatadogWidgetLayout(terraformLayout map[string]interface{}) *datadogV1.WidgetLayout {
	datadogLayout := datadogV1.NewWidgetLayoutWithDefaults()
	datadogLayout.SetX(int64(terraformLayout["x"].(int)))
	datadogLayout.SetY(int64(terraformLayout["y"].(int)))
	datadogLayout.SetHeight(int64(terraformLayout["height"].(int)))
	datadogLayout.SetWidth(int64(terraformLayout["width"].(int)))
	if value, ok := terraformLayout["is_column_break"].(bool); ok && value {
		datadogLayout.SetIsColumnBreak(value)
	}
	return datadogLayout
}

func buildDatadogChangeDefinition(terraformDefinition map[string]interface{}) *datadogV1.ChangeWidgetDefinition {
	datadogDefinition := datadogV1.NewChangeWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogChangeRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildTerraformChangeDefinition(datadogDefinition *datadogV1.ChangeWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformChangeRequests(&datadogDefinition.Requests)

	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}
func buildDatadogChangeRequests(terraformRequests *[]interface{}) *[]datadogV1.ChangeWidgetRequest {
	datadogRequests := make([]datadogV1.ChangeWidgetRequest, len(*terraformRequests))
	for i, request := range *terraformRequests {
		if request == nil {
			continue
		}
		terraformRequest := request.(map[string]interface{})
		// Build ChangeRequest
		datadogChangeRequest := datadogV1.NewChangeWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogChangeRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogChangeRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogChangeRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogChangeRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogChangeRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogChangeRequest.SetQueries(queries)
			// Change request for formulas and functions always have a response format of "scalar"
			datadogChangeRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}

		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogChangeRequest.SetFormulas(formulas)
		}

		if v, ok := terraformRequest["change_type"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetChangeType(datadogV1.WidgetChangeType(v))
		}
		if v, ok := terraformRequest["compare_to"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetCompareTo(datadogV1.WidgetCompareTo(v))
		}
		if v, ok := terraformRequest["increase_good"].(bool); ok {
			datadogChangeRequest.SetIncreaseGood(v)
		}
		if v, ok := terraformRequest["order_by"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetOrderBy(datadogV1.WidgetOrderBy(v))
		}
		if v, ok := terraformRequest["order_dir"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetOrderDir(datadogV1.WidgetSort(v))
		}
		if v, ok := terraformRequest["show_present"].(bool); ok {
			datadogChangeRequest.SetShowPresent(v)
		}

		datadogRequests[i] = *datadogChangeRequest
	}
	return &datadogRequests
}
func buildTerraformChangeRequests(datadogChangeRequests *[]datadogV1.ChangeWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogChangeRequests))
	for i, datadogRequest := range *datadogChangeRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetSecurityQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}

		if v, ok := datadogRequest.GetChangeTypeOk(); ok {
			terraformRequest["change_type"] = *v
		}
		if v, ok := datadogRequest.GetCompareToOk(); ok {
			terraformRequest["compare_to"] = *v
		}
		if v, ok := datadogRequest.GetIncreaseGoodOk(); ok {
			terraformRequest["increase_good"] = *v
		}
		if v, ok := datadogRequest.GetOrderByOk(); ok {
			terraformRequest["order_by"] = *v
		}
		if v, ok := datadogRequest.GetOrderDirOk(); ok {
			terraformRequest["order_dir"] = *v
		}
		if v, ok := datadogRequest.GetShowPresentOk(); ok {
			terraformRequest["show_present"] = *v
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildDatadogQueryValueDefinition(terraformDefinition map[string]interface{}) *datadogV1.QueryValueWidgetDefinition {
	datadogDefinition := datadogV1.NewQueryValueWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogQueryValueRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["autoscale"].(bool); ok {
		datadogDefinition.SetAutoscale(v)
	}
	if v, ok := terraformDefinition["custom_unit"].(string); ok && len(v) != 0 {
		datadogDefinition.SetCustomUnit(v)
	}
	if v, ok := terraformDefinition["precision"].(int); ok {
		datadogDefinition.SetPrecision(int64(v))
	}
	if timeseriesBackground, ok := terraformDefinition["timeseries_background"].([]interface{}); ok && len(timeseriesBackground) > 0 {
		if v, ok := timeseriesBackground[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.SetTimeseriesBackground(*buildDatadogTimeseriesBackground(v))
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}

func buildTerraformTimeseriesBackground(datadogTimeseriesBackground *datadogV1.TimeseriesBackground) []map[string]interface{} {
	terraformTimeseriesBackground := map[string]interface{}{}
	if v, ok := datadogTimeseriesBackground.GetTypeOk(); ok {
		terraformTimeseriesBackground["type"] = v
	}

	if v, ok := datadogTimeseriesBackground.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformTimeseriesBackground["yaxis"] = []map[string]interface{}{axis}
	}

	terraformTimeseriesBackgroundArray := []map[string]interface{}{terraformTimeseriesBackground}
	return terraformTimeseriesBackgroundArray

}

func buildTerraformQueryValueDefinition(datadogDefinition *datadogV1.QueryValueWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformQueryValueRequests(&datadogDefinition.Requests)

	// Optional params
	if v, ok := datadogDefinition.GetAutoscaleOk(); ok {
		terraformDefinition["autoscale"] = *v
	}
	if v, ok := datadogDefinition.GetCustomUnitOk(); ok {
		terraformDefinition["custom_unit"] = *v
	}
	if v, ok := datadogDefinition.GetPrecisionOk(); ok {
		terraformDefinition["precision"] = *v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeseriesBackgroundOk(); ok {
		terraformDefinition["timeseries_background"] = buildTerraformTimeseriesBackground(v)
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}
func buildDatadogQueryValueRequests(terraformRequests *[]interface{}) *[]datadogV1.QueryValueWidgetRequest {
	datadogRequests := make([]datadogV1.QueryValueWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build QueryValueRequest
		datadogQueryValueRequest := datadogV1.NewQueryValueWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogQueryValueRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["audit_query"].([]interface{}); ok && len(v) > 0 {
			auditQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.AuditQuery = buildDatadogApmOrLogQuery(auditQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogQueryValueRequest.SetQueries(queries)
			// Query Value requests for formulas and functions always has a response format of "scalar"
			datadogQueryValueRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogQueryValueRequest.SetFormulas(formulas)
		}

		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogQueryValueRequest.ConditionalFormats = *buildDatadogWidgetConditionalFormat(&v)
		}
		if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
			datadogQueryValueRequest.SetAggregator(datadogV1.WidgetAggregator(v))
		}

		datadogRequests[i] = *datadogQueryValueRequest
	}
	return &datadogRequests
}
func buildTerraformQueryValueRequests(datadogQueryValueRequests *[]datadogV1.QueryValueWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogQueryValueRequests))
	for i, datadogRequest := range *datadogQueryValueRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = *v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetSecurityQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetAuditQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["audit_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}

		if datadogRequest.ConditionalFormats != nil {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(&datadogRequest.ConditionalFormats)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}

		if v, ok := datadogRequest.GetAggregatorOk(); ok {
			terraformRequest["aggregator"] = *v
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

// Query Table Widget Definition helpers
func buildDatadogQueryTableDefinition(terraformDefinition map[string]interface{}) *datadogV1.TableWidgetDefinition {
	datadogDefinition := datadogV1.NewTableWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogQueryTableRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	if v, ok := terraformDefinition["has_search_bar"].(string); ok && len(v) != 0 {
		datadogDefinition.SetHasSearchBar(datadogV1.TableWidgetHasSearchBar(v))
	}
	return datadogDefinition
}
func buildTerraformQueryTableDefinition(datadogDefinition *datadogV1.TableWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformQueryTableRequests(&datadogDefinition.Requests)

	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	if v, ok := datadogDefinition.GetHasSearchBarOk(); ok {
		terraformDefinition["has_search_bar"] = *v
	}
	return terraformDefinition
}
func buildDatadogQueryTableRequests(terraformRequests *[]interface{}) *[]datadogV1.TableWidgetRequest {
	datadogRequests := make([]datadogV1.TableWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build QueryTableRequest
		datadogQueryTableRequest := datadogV1.NewTableWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["apm_stats_query"].([]interface{}); ok && len(v) > 0 {
			apmStatsQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.ApmStatsQuery = buildDatadogApmStatsQuery(apmStatsQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["apm_dependency_stats_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionAPMDependencyStatsQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["apm_resource_stats_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionAPMResourceStatsQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogQueryTableRequest.SetQueries(queries)
			// Query Table request for formulas and functions always have a response format of "scalar"
			datadogQueryTableRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}

		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogQueryTableRequest.SetFormulas(formulas)
		}

		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogQueryTableRequest.ConditionalFormats = *buildDatadogWidgetConditionalFormat(&v)
		}
		if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetAggregator(datadogV1.WidgetAggregator(v))
		}
		if v, ok := terraformRequest["alias"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetAlias(v)

		}
		if v, ok := terraformRequest["limit"].(int); ok && v != 0 {
			datadogQueryTableRequest.SetLimit(int64(v))
		}
		if v, ok := terraformRequest["order"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetOrder(datadogV1.WidgetSort(v))
		}
		if v, ok := terraformRequest["cell_display_mode"].([]interface{}); ok && len(v) != 0 {
			datadogCellDisplayMode := make([]datadogV1.TableWidgetCellDisplayMode, len(v))
			for i, cellDisplayMode := range v {
				datadogCellDisplayMode[i] = datadogV1.TableWidgetCellDisplayMode(cellDisplayMode.(string))
			}
			datadogQueryTableRequest.CellDisplayMode = datadogCellDisplayMode
		}
		if v, ok := terraformRequest["text_formats"].([]interface{}); ok && len(v) != 0 {
			datadogQueryTableRequest.TextFormats = make([][]datadogV1.TableWidgetTextFormatRule, len(v))
			for i, w := range v {
				if c, ok := w.(map[string]interface{}); ok {
					if textFormat, ok := c["text_format"].([]interface{}); ok && len(textFormat) > 0 {
						datadogQueryTableRequest.TextFormats[i] = *buildDatadogQueryTableTextFormat(&textFormat)
					}
				} else {
					datadogQueryTableRequest.TextFormats[i] = []datadogV1.TableWidgetTextFormatRule{}
				}
			}
		}

		datadogRequests[i] = *datadogQueryTableRequest
	}
	return &datadogRequests
}
func buildTerraformQueryTableRequests(datadogQueryTableRequests *[]datadogV1.TableWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogQueryTableRequests))
	for i, datadogRequest := range *datadogQueryTableRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetSecurityQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetApmStatsQueryOk(); ok {
			terraformQuery := buildTerraformApmStatsQuery(*v)
			terraformRequest["apm_stats_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, true)
		}

		if v, ok := datadogRequest.GetConditionalFormatsOk(); ok {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(v)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}

		if v, ok := datadogRequest.GetAggregatorOk(); ok {
			terraformRequest["aggregator"] = *v
		}
		if v, ok := datadogRequest.GetAliasOk(); ok {
			terraformRequest["alias"] = *v
		}
		if v, ok := datadogRequest.GetLimitOk(); ok {
			terraformRequest["limit"] = *v
		}
		if v, ok := datadogRequest.GetOrderOk(); ok {
			terraformRequest["order"] = *v
		}
		if v, ok := datadogRequest.GetCellDisplayModeOk(); ok {
			terraformCellDisplayMode := make([]datadogV1.TableWidgetCellDisplayMode, len(*v))
			for i, cellDisplayMode := range *v {
				terraformCellDisplayMode[i] = cellDisplayMode
			}
			terraformRequest["cell_display_mode"] = terraformCellDisplayMode
		}
		if v, ok := datadogRequest.GetTextFormatsOk(); ok {
			terraformTextFormats := make([]map[string][]map[string]interface{}, len(*v))
			for i, textFormat := range *v {
				test := buildTerraformQueryTableTextFormat(&textFormat)
				terraformTextFormats[i] = make(map[string][]map[string]interface{})
				terraformTextFormats[i]["text_format"] = *test
			}
			terraformRequest["text_formats"] = terraformTextFormats
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

// Query Table Widget Text Format Helpers
func buildDatadogQueryTableTextFormat(terraformQueryTableTextFormat *[]interface{}) *[]datadogV1.TableWidgetTextFormatRule {
	datadogQueryTableTextFormat := make([]datadogV1.TableWidgetTextFormatRule, len(*terraformQueryTableTextFormat))
	for j, textFormatRule := range *terraformQueryTableTextFormat {
		terraformTextFormatRule := textFormatRule.(map[string]interface{})

		match, _ := terraformTextFormatRule["match"].([]interface{})
		terraformTextFormatMatch := match[0].(map[string]interface{})
		datadogMatch := datadogV1.NewTableWidgetTextFormatMatch(datadogV1.TableWidgetTextFormatMatchType(terraformTextFormatMatch["type"].(string)), terraformTextFormatMatch["value"].(string))
		datadogTextFormatRule := datadogV1.NewTableWidgetTextFormatRule(*datadogMatch)
		// Optional

		if v, ok := terraformTextFormatRule["replace"].([]interface{}); ok && len(v) > 0 {
			if replace, ok := v[0].(map[string]interface{}); ok {
				if w, ok := replace["type"].(string); ok && len(w) != 0 {
					switch w {
					case "all":
						datadogReplace := datadogV1.NewTableWidgetTextFormatReplaceAll(datadogV1.TABLEWIDGETTEXTFORMATREPLACEALLTYPE_ALL, replace["with"].(string))
						datadogTextFormatRule.SetReplace(datadogV1.TableWidgetTextFormatReplaceAllAsTableWidgetTextFormatReplace(datadogReplace))
					case "substring":
						datadogReplace := datadogV1.NewTableWidgetTextFormatReplaceSubstring(replace["substring"].(string), datadogV1.TABLEWIDGETTEXTFORMATREPLACESUBSTRINGTYPE_SUBSTRING, replace["with"].(string))
						datadogTextFormatRule.SetReplace(datadogV1.TableWidgetTextFormatReplaceSubstringAsTableWidgetTextFormatReplace(datadogReplace))
					}
				}
			}
		}
		if v, ok := terraformTextFormatRule["palette"].(string); ok && len(v) != 0 {
			datadogTextFormatRule.SetPalette(datadogV1.TableWidgetTextFormatPalette(v))
		} else {
			datadogTextFormatRule.Palette = nil
		}
		if v, ok := terraformTextFormatRule["custom_bg_color"].(string); ok && len(v) != 0 {
			datadogTextFormatRule.SetCustomBgColor(v)
		}
		if v, ok := terraformTextFormatRule["custom_fg_color"].(string); ok && len(v) != 0 {
			datadogTextFormatRule.SetCustomFgColor(v)
		}
		datadogQueryTableTextFormat[j] = *datadogTextFormatRule
	}
	return &datadogQueryTableTextFormat
}
func buildTerraformQueryTableTextFormat(datadogQueryTableTextFormat *[]datadogV1.TableWidgetTextFormatRule) *[]map[string]interface{} {
	terraformQueryTableTextFormat := make([]map[string]interface{}, len(*datadogQueryTableTextFormat))
	for i, datadogQueryTableTextFormatRule := range *datadogQueryTableTextFormat {
		terraformQueryTableTextFormatRule := map[string]interface{}{}
		// Required params
		match := make(map[string]interface{})
		match["type"] = datadogQueryTableTextFormatRule.GetMatch().Type
		match["value"] = datadogQueryTableTextFormatRule.GetMatch().Value
		terraformQueryTableTextFormatRule["match"] = []interface{}{match}
		// Optional params
		if v, ok := datadogQueryTableTextFormatRule.GetReplaceOk(); ok {
			if v.TableWidgetTextFormatReplaceAll != nil {
				replace := make(map[string]interface{})
				replace["type"] = v.TableWidgetTextFormatReplaceAll.Type
				replace["with"] = v.TableWidgetTextFormatReplaceAll.With
				terraformQueryTableTextFormatRule["replace"] = []interface{}{replace}
			}
			if v.TableWidgetTextFormatReplaceSubstring != nil {
				replace := make(map[string]interface{})
				replace["type"] = v.TableWidgetTextFormatReplaceSubstring.Type
				replace["with"] = v.TableWidgetTextFormatReplaceSubstring.With
				replace["substring"] = v.TableWidgetTextFormatReplaceSubstring.Substring
				terraformQueryTableTextFormatRule["replace"] = []interface{}{replace}
			}
		}
		if v, ok := datadogQueryTableTextFormatRule.GetPaletteOk(); ok {
			terraformQueryTableTextFormatRule["palette"] = v
		}
		if v, ok := datadogQueryTableTextFormatRule.GetCustomBgColorOk(); ok {
			terraformQueryTableTextFormatRule["custom_bg_color"] = v
		}
		if v, ok := datadogQueryTableTextFormatRule.GetCustomFgColorOk(); ok {
			terraformQueryTableTextFormatRule["custom_fg_color"] = v
		}
		terraformQueryTableTextFormat[i] = terraformQueryTableTextFormatRule
	}
	return &terraformQueryTableTextFormat
}

func buildDatadogScatterplotDefinition(terraformDefinition map[string]interface{}) *datadogV1.ScatterPlotWidgetDefinition {
	datadogDefinition := datadogV1.NewScatterPlotWidgetDefinitionWithDefaults()
	// Required params
	if v, ok := terraformDefinition["request"].([]interface{}); ok && len(v) > 0 {
		terraformRequests := v[0].(map[string]interface{})
		datadogRequests := datadogV1.NewScatterPlotWidgetDefinitionRequestsWithDefaults()
		if terraformXArray, ok := terraformRequests["x"].([]interface{}); ok && len(terraformXArray) > 0 {
			terraformX := terraformXArray[0].(map[string]interface{})
			datadogRequests.SetX(*buildDatadogScatterplotRequest(terraformX))
		}
		if terraformYArray, ok := terraformRequests["y"].([]interface{}); ok && len(terraformYArray) > 0 {
			terraformY := terraformYArray[0].(map[string]interface{})
			datadogRequests.SetY(*buildDatadogScatterplotRequest(terraformY))
		}

		if terraformScatterplotTableRequests, ok := terraformRequests["scatterplot_table"].([]interface{}); ok && len(terraformScatterplotTableRequests) > 0 {
			terraformScatterplotTableRequest := terraformScatterplotTableRequests[0].(map[string]interface{})
			datadogRequests.SetTable(*buildDatadogScatterplotTableRequest(terraformScatterplotTableRequest))
		}

		datadogDefinition.SetRequests(*datadogRequests)
	}

	// Optional params
	if axis, ok := terraformDefinition["xaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Xaxis = buildDatadogWidgetAxis(v)
		}
	}
	if axis, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(v)
		}
	}
	if terraformColorByGroups, ok := terraformDefinition["color_by_groups"].([]interface{}); ok && len(terraformColorByGroups) > 0 {
		datadogColorByGroups := make([]string, len(terraformColorByGroups))
		for i, colorByGroup := range terraformColorByGroups {
			datadogColorByGroups[i] = colorByGroup.(string)
		}
		datadogDefinition.ColorByGroups = datadogColorByGroups
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}

func buildDatadogScatterplotTableRequest(terraformRequest map[string]interface{}) *datadogV1.ScatterplotTableRequest {
	datadogScatterplotTableRequest := datadogV1.NewScatterplotTableRequest()

	if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
		queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
		for i, q := range v {
			query := q.(map[string]interface{})
			if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
				queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
			} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
				queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
			} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
				queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
			} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
				queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
			} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
				queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
			}
		}
		datadogScatterplotTableRequest.SetQueries(queries)
		datadogScatterplotTableRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
	}

	if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
		formulas := make([]datadogV1.ScatterplotWidgetFormula, len(v))
		for i, formula := range v {
			formulas[i] = *buildDatadogScatterplotFormula(formula.(map[string]interface{}))
		}
		datadogScatterplotTableRequest.SetFormulas(formulas)
	}

	return datadogScatterplotTableRequest
}

func buildDatadogScatterplotRequest(terraformRequest map[string]interface{}) *datadogV1.ScatterPlotRequest {

	datadogScatterplotRequest := datadogV1.NewScatterPlotRequest()
	if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
		datadogScatterplotRequest.SetQ(v)
	} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
		apmQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
	} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
		logQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
	} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
		processQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
	} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
		rumQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
	} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
		securityQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
	}

	if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
		datadogScatterplotRequest.SetAggregator(datadogV1.ScatterplotWidgetAggregator(v))
	}

	return datadogScatterplotRequest
}

// Geomap Widget Definition helpers
func buildDatadogGeomapDefinition(terraformDefinition map[string]interface{}) *datadogV1.GeomapWidgetDefinition {
	datadogDefinition := datadogV1.NewGeomapWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogGeomapRequests(&terraformRequests)

	if style, ok := terraformDefinition["style"].([]interface{}); ok && len(style) > 0 {
		if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Style = *buildDatadogGeomapRequestStyle(v)
		}
	}

	if view, ok := terraformDefinition["view"].([]interface{}); ok && len(view) > 0 {
		if v, ok := view[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.View = *buildDatadogGeomapRequestView(v)
		}
	}

	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}

	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}

	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}

	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}

	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}

	return datadogDefinition
}

func buildDatadogGeomapRequests(terraformRequests *[]interface{}) *[]datadogV1.GeomapWidgetRequest {
	datadogRequests := make([]datadogV1.GeomapWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build Geomap Request
		datadogGeomapRequest := datadogV1.NewGeomapWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogGeomapRequest.SetQ(v)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogGeomapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogGeomapRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				}
			}
			datadogGeomapRequest.SetQueries(queries)
			// Geomap requests for formulas and functions always has a response format of "scalar"
			datadogGeomapRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogGeomapRequest.SetFormulas(formulas)
		}

		datadogRequests[i] = *datadogGeomapRequest
	}
	return &datadogRequests
}

func buildTerraformGeomapRequests(datadogGeomapRequests *[]datadogV1.GeomapWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogGeomapRequests))
	for i, datadogRequest := range *datadogGeomapRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}

		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildDatadogTimeseriesDefinition(terraformDefinition map[string]interface{}) *datadogV1.TimeseriesWidgetDefinition {
	datadogDefinition := datadogV1.NewTimeseriesWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogTimeseriesRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["marker"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Markers = *buildDatadogWidgetMarkers(&v)
	}
	if v, ok := terraformDefinition["event"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Events = *buildDatadogWidgetEvents(&v)
	}
	if v, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(v) > 0 {
		if axis, ok := v[0].(map[string]interface{}); ok && len(axis) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(axis)
		}
	}
	if v, ok := terraformDefinition["right_yaxis"].([]interface{}); ok && len(v) > 0 {
		if axis, ok := v[0].(map[string]interface{}); ok && len(axis) > 0 {
			datadogDefinition.RightYaxis = buildDatadogWidgetAxis(axis)
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.SetShowLegend(v)
	}
	if v, ok := terraformDefinition["legend_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendSize(v)
	}
	if v, ok := terraformDefinition["legend_layout"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendLayout(datadogV1.TimeseriesWidgetLegendLayout(v))
	}
	if v, ok := terraformDefinition["legend_columns"]; ok && v.(*schema.Set).Len() != 0 {
		datadogLegendColumns := make([]datadogV1.TimeseriesWidgetLegendColumn, v.(*schema.Set).Len())
		for i, legendColumn := range v.(*schema.Set).List() {
			datadogLegendColumns[i] = datadogV1.TimeseriesWidgetLegendColumn(legendColumn.(string))
		}
		datadogDefinition.SetLegendColumns(datadogLegendColumns)
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}

func buildTerraformTimeseriesDefinition(datadogDefinition *datadogV1.TimeseriesWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformTimeseriesRequests(&datadogDefinition.Requests)

	// Optional params
	if v, ok := datadogDefinition.GetMarkersOk(); ok {
		terraformDefinition["marker"] = buildTerraformWidgetMarkers(v)
	}
	if v, ok := datadogDefinition.GetEventsOk(); ok {
		terraformDefinition["event"] = buildTerraformWidgetEvents(v)
	}
	if v, ok := datadogDefinition.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["yaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetRightYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["right_yaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}
	if v, ok := datadogDefinition.GetShowLegendOk(); ok {
		terraformDefinition["show_legend"] = *v
	}
	if v, ok := datadogDefinition.GetLegendSizeOk(); ok {
		terraformDefinition["legend_size"] = *v
	}
	if v, ok := datadogDefinition.GetLegendLayoutOk(); ok {
		terraformDefinition["legend_layout"] = *v
	}
	if v, ok := datadogDefinition.GetLegendColumnsOk(); ok {
		terraformLegendColumns := make([]string, len(*v))
		for i, legendColumn := range *v {
			terraformLegendColumns[i] = string(legendColumn)
		}
		terraformDefinition["legend_columns"] = terraformLegendColumns
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func buildDatadogSunburstLegendInline(terraformSunburstLegendInline map[string]interface{}) *datadogV1.SunburstWidgetLegend {
	datadogSunburstLegendInline := &datadogV1.SunburstWidgetLegendInlineAutomatic{}
	if v, ok := terraformSunburstLegendInline["type"].(string); ok && len(v) != 0 {
		legendType := datadogV1.SunburstWidgetLegendInlineAutomaticType(terraformSunburstLegendInline["type"].(string))
		datadogSunburstLegendInline.SetType(legendType)
	}

	if v, ok := terraformSunburstLegendInline["hide_value"].(bool); ok {
		datadogSunburstLegendInline.SetHideValue(v)
	}

	if v, ok := terraformSunburstLegendInline["hide_percent"].(bool); ok {
		datadogSunburstLegendInline.SetHidePercent(v)
	}

	datadogSunburstLegend := datadogV1.SunburstWidgetLegend{}
	datadogSunburstLegend.SunburstWidgetLegendInlineAutomatic = datadogSunburstLegendInline

	return &datadogSunburstLegend
}

func buildDatadogSunburstLegendTable(terraformSunburstLegendTable map[string]interface{}) *datadogV1.SunburstWidgetLegend {
	datadogSunburstLegendTable := &datadogV1.SunburstWidgetLegendTable{}
	if v, ok := terraformSunburstLegendTable["type"].(string); ok && len(v) != 0 {
		legendType := datadogV1.SunburstWidgetLegendTableType(terraformSunburstLegendTable["type"].(string))
		datadogSunburstLegendTable.SetType(legendType)
	}

	datadogSunburstLegend := datadogV1.SunburstWidgetLegend{}
	datadogSunburstLegend.SunburstWidgetLegendTable = datadogSunburstLegendTable

	return &datadogSunburstLegend
}

func buildDatadogSunburstDefinition(terraformDefinition map[string]interface{}) *datadogV1.SunburstWidgetDefinition {
	datadogDefinition := datadogV1.NewSunburstWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogSunburstRequests(&terraformRequests)

	// Optional params
	if legendInline, ok := terraformDefinition["legend_inline"].([]interface{}); ok && len(legendInline) > 0 {
		if v, ok := legendInline[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.SetLegend(*buildDatadogSunburstLegendInline(v))
		}
	}

	if legendTable, ok := terraformDefinition["legend_table"].([]interface{}); ok && len(legendTable) > 0 {
		if v, ok := legendTable[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.SetLegend(*buildDatadogSunburstLegendTable(v))
		}
	}

	if hideTotal, ok := terraformDefinition["hide_total"].(bool); ok && hideTotal {
		datadogDefinition.SetHideTotal(hideTotal)
	}

	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}

	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}

	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}

	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}

	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}

	return datadogDefinition
}

func buildDatadogSunburstRequests(terraformRequests *[]interface{}) *[]datadogV1.SunburstWidgetRequest {
	datadogRequests := make([]datadogV1.SunburstWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build Sunburst request
		datadogSunburstRequest := datadogV1.NewSunburstWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogSunburstRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["network_query"].([]interface{}); ok && len(v) > 0 {
			networkQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.NetworkQuery = buildDatadogApmOrLogQuery(networkQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["audit_query"].([]interface{}); ok && len(v) > 0 {
			auditQuery := v[0].(map[string]interface{})
			datadogSunburstRequest.AuditQuery = buildDatadogApmOrLogQuery(auditQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogSunburstRequest.SetQueries(queries)
			datadogSunburstRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogSunburstRequest.SetFormulas(formulas)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogSunburstRequest.Style = buildDatadogWidgetStyle(v)
			}
		}
		datadogRequests[i] = *datadogSunburstRequest
	}
	return &datadogRequests
}

func buildTerraformSunburstRequests(datadogSunburstRequests *[]datadogV1.SunburstWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogSunburstRequests))
	for i, datadogRequest := range *datadogSunburstRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetNetworkQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["network_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetSecurityQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetAuditQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["audit_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildTerraformSunburstLegendInline(datadogSunburstLegend *datadogV1.SunburstWidgetLegend) []map[string]interface{} {
	terraformSunburstLegend := map[string]interface{}{}
	terraformSunburstLegendInline := datadogSunburstLegend.SunburstWidgetLegendInlineAutomatic
	if terraformSunburstLegendInline != nil {
		if v, ok := terraformSunburstLegendInline.GetTypeOk(); ok {
			terraformSunburstLegend["type"] = v
		}

		if v, ok := terraformSunburstLegendInline.GetHideValueOk(); ok {
			terraformSunburstLegend["hide_value"] = v
		}

		if v, ok := terraformSunburstLegendInline.GetHidePercentOk(); ok {
			terraformSunburstLegend["hide_percent"] = v
		}
	}

	terraformSunburstLegendArray := []map[string]interface{}{terraformSunburstLegend}
	return terraformSunburstLegendArray
}

func buildTerraformSunburstLegendTable(datadogSunburstLegend *datadogV1.SunburstWidgetLegend) []map[string]interface{} {
	terraformSunburstLegend := map[string]interface{}{}
	terraformSunburstLegendTable := datadogSunburstLegend.SunburstWidgetLegendTable
	if terraformSunburstLegendTable != nil {
		if v, ok := terraformSunburstLegendTable.GetTypeOk(); ok {
			terraformSunburstLegend["type"] = v
		}
	}

	terraformSunburstLegendArray := []map[string]interface{}{terraformSunburstLegend}
	return terraformSunburstLegendArray
}

func buildTerraformSunburstDefinition(datadogDefinition *datadogV1.SunburstWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformSunburstRequests(&datadogDefinition.Requests)

	if v, ok := datadogDefinition.GetLegendOk(); ok {
		// Use `hide_value` as a discriminant to determine which type of legend we are serializing
		if _, ok := v.SunburstWidgetLegendInlineAutomatic.GetHideValueOk(); ok {
			terraformDefinition["legend_inline"] = buildTerraformSunburstLegendInline(v)
		} else {
			terraformDefinition["legend_table"] = buildTerraformSunburstLegendTable(v)
		}
	}

	if v, ok := datadogDefinition.GetHideTotalOk(); ok {
		terraformDefinition["hide_total"] = v
	}

	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}

	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}

	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}

	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}

	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}

	return terraformDefinition
}

// getQueryTableFormulaSchema returns the formula schema for query_table widgets.
// cell_display_mode_options is now included in getFormulaSchema() for all formula-capable
// widgets, so this function is an alias.
func buildDatadogScatterplotFormula(data map[string]interface{}) *datadogV1.ScatterplotWidgetFormula {
	formula := datadogV1.ScatterplotWidgetFormula{}
	if formulaExpression, ok := data["formula_expression"].(string); ok && len(formulaExpression) != 0 {
		formula.SetFormula(formulaExpression)
	}
	if alias, ok := data["alias"].(string); ok && len(alias) != 0 {
		formula.SetAlias(alias)
	}
	if dimension, ok := data["dimension"].(string); ok && len(dimension) != 0 {
		formula.SetDimension(datadogV1.ScatterplotDimension(dimension))
	}
	return &formula
}

func buildDatadogFormula(data map[string]interface{}) *datadogV1.WidgetFormula {
	formula := datadogV1.WidgetFormula{}
	if formulaExpression, ok := data["formula_expression"].(string); ok && len(formulaExpression) != 0 {
		formula.SetFormula(formulaExpression)
	}
	if alias, ok := data["alias"].(string); ok && len(alias) != 0 {
		formula.SetAlias(alias)
	}
	if limits, ok := data["limit"].([]interface{}); ok && len(limits) != 0 {
		datadogLimit := datadogV1.NewWidgetFormulaLimit()
		limit := limits[0].(map[string]interface{})
		if count, ok := limit["count"].(int); ok && count != 0 {
			datadogLimit.SetCount(int64(count))
		}
		if order, ok := limit["order"].(string); ok && len(order) > 0 {
			datadogLimit.SetOrder(datadogV1.QuerySortOrder(order))
		}
		formula.SetLimit(*datadogLimit)
	}
	if value, ok := data["cell_display_mode"].(string); ok && len(value) != 0 {
		formula.SetCellDisplayMode(datadogV1.TableWidgetCellDisplayMode(value))
	}
	if value, ok := data["cell_display_mode_options"].([]interface{}); ok && len(value) != 0 {
		if options, ok := value[0].(map[string]interface{}); ok {
			o := datadogV1.NewWidgetFormulaCellDisplayModeOptions()
			if v, ok := options["trend_type"].(string); ok {
				o.SetTrendType(datadogV1.WidgetFormulaCellDisplayModeOptionsTrendType(v))
			}
			if v, ok := options["y_scale"].(string); ok {
				o.SetYScale(datadogV1.WidgetFormulaCellDisplayModeOptionsYScale(v))
			}
			formula.SetCellDisplayModeOptions(*o)
		}
	}

	if v, ok := data["conditional_formats"].([]interface{}); ok && len(v) != 0 {
		formula.ConditionalFormats = *buildDatadogWidgetConditionalFormat(&v)
	}

	if style, ok := data["style"].([]interface{}); ok && len(style) != 0 {
		datadogFormulaStyle := datadogV1.NewWidgetFormulaStyle()
		style_attr := style[0].(map[string]interface{})
		if palette, ok := style_attr["palette"].(string); ok {
			datadogFormulaStyle.SetPalette(palette)
		}
		if palette_index, ok := style_attr["palette_index"].(int); ok {
			datadogFormulaStyle.SetPaletteIndex(int64(palette_index))
		}
		formula.SetStyle(*datadogFormulaStyle)
	}
	if number, ok := data["number_format"].([]interface{}); ok && len(number) != 0 {
		datadogNumberFormat := buildDatadogNumberFormatFormulaSchema(number[0].(map[string]interface{}))
		formula.SetNumberFormat(*datadogNumberFormat)
	}

	return &formula
}

func buildDatadogEventQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionEventsDataSource(data["data_source"].(string))
	computeList := data["compute"].([]interface{})
	computeMap := computeList[0].(map[string]interface{})
	aggregation := datadogV1.FormulaAndFunctionEventAggregation(computeMap["aggregation"].(string))
	compute := datadogV1.NewFormulaAndFunctionEventQueryDefinitionCompute(aggregation)
	if interval, ok := computeMap["interval"].(int); ok && interval != 0 {
		compute.SetInterval(int64(interval))
	}
	if metric, ok := computeMap["metric"].(string); ok && len(metric) > 0 {
		compute.SetMetric(metric)
	}
	eventQuery := datadogV1.NewFormulaAndFunctionEventQueryDefinition(*compute, dataSource, data["name"].(string))
	if storage, ok := data["storage"].(string); ok && storage != "" {
		eventQuery.SetStorage(storage)
	}
	eventQueryIndexes := data["indexes"].([]interface{})
	indexes := make([]string, len(eventQueryIndexes))
	for i, index := range eventQueryIndexes {
		indexes[i] = index.(string)
	}
	eventQuery.SetIndexes(indexes)

	if terraformSearches, ok := data["search"].([]interface{}); ok && len(terraformSearches) > 0 {
		terraformSearch := terraformSearches[0].(map[string]interface{})
		eventQuery.Search = datadogV1.NewFormulaAndFunctionEventQueryDefinitionSearch(terraformSearch["query"].(string))
	}

	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			eventQuery.CrossOrgUuids = []string{c}
		}
	}

	// GroupBy
	if terraformGroupBys, ok := data["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]datadogV1.FormulaAndFunctionEventQueryGroupBy, len(terraformGroupBys))
		for i, g := range terraformGroupBys {
			groupBy := g.(map[string]interface{})

			// Facet
			datadogGroupBy := datadogV1.NewFormulaAndFunctionEventQueryGroupBy(groupBy["facet"].(string))

			// Limit
			if v, ok := groupBy["limit"].(int); ok && v != 0 {
				datadogGroupBy.SetLimit(int64(v))
			}

			// Sort
			if v, ok := groupBy["sort"].([]interface{}); ok && len(v) > 0 {
				if v, ok := v[0].(map[string]interface{}); ok && len(v) > 0 {
					sortMap := &datadogV1.FormulaAndFunctionEventQueryGroupBySort{}
					if aggr, ok := v["aggregation"].(string); ok && len(aggr) > 0 {
						aggregation := datadogV1.FormulaAndFunctionEventAggregation(v["aggregation"].(string))
						sortMap.SetAggregation(aggregation)
					}
					if order, ok := v["order"].(string); ok && len(order) > 0 {
						eventSort := datadogV1.QuerySortOrder(order)
						sortMap.SetOrder(eventSort)
					}
					if metric, ok := v["metric"].(string); ok && len(metric) > 0 {
						sortMap.SetMetric(metric)
					}
					datadogGroupBy.SetSort(*sortMap)
				}
			}

			datadogGroupBys[i] = *datadogGroupBy
		}
		eventQuery.SetGroupBy(datadogGroupBys)
	}

	definition := datadogV1.FormulaAndFunctionEventQueryDefinitionAsFormulaAndFunctionQueryDefinition(eventQuery)
	return &definition
}

func buildDatadogMetricQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionMetricDataSource("metrics")
	metricQuery := datadogV1.NewFormulaAndFunctionMetricQueryDefinition(dataSource, data["name"].(string), data["query"].(string))
	if v, ok := data["aggregator"].(string); ok && len(v) != 0 {
		aggregator := datadogV1.FormulaAndFunctionMetricAggregation(data["aggregator"].(string))
		metricQuery.SetAggregator(aggregator)
	}

	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			metricQuery.CrossOrgUuids = []string{c}
		}
	}

	if v, ok := data["semantic_mode"].(string); ok && len(v) != 0 {
		semanticMode := datadogV1.FormulaAndFunctionMetricSemanticMode(v)
		metricQuery.SetSemanticMode(semanticMode)
	}

	definition := datadogV1.FormulaAndFunctionMetricQueryDefinitionAsFormulaAndFunctionQueryDefinition(metricQuery)
	return &definition
}

func buildDatadogFormulaAndFunctionAPMResourceStatsQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionApmResourceStatsDataSource(data["data_source"].(string))
	stat := datadogV1.FormulaAndFunctionApmResourceStatName(data["stat"].(string))
	apmResourceStatsQuery := datadogV1.NewFormulaAndFunctionApmResourceStatsQueryDefinition(dataSource, data["env"].(string), data["name"].(string), data["service"].(string), stat)

	// cross_org_uuids
	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			apmResourceStatsQuery.CrossOrgUuids = []string{c}
		}
	}

	// operation_name
	if v, ok := data["operation_name"].(string); ok && len(v) != 0 {
		apmResourceStatsQuery.SetOperationName(v)
	}

	// resource_name
	if v, ok := data["resource_name"].(string); ok && len(v) != 0 {
		apmResourceStatsQuery.SetResourceName(v)
	}

	// primary_tag_name
	if v, ok := data["primary_tag_name"].(string); ok && len(v) != 0 {
		apmResourceStatsQuery.SetPrimaryTagName(v)
	}

	// primary_tag_value
	if v, ok := data["primary_tag_value"].(string); ok && len(v) != 0 {
		apmResourceStatsQuery.SetPrimaryTagValue(v)
	}

	// group_by
	if terraformGroupBys, ok := data["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]string, len(terraformGroupBys))
		for i, groupBy := range terraformGroupBys {
			datadogGroupBys[i] = groupBy.(string)
		}
		apmResourceStatsQuery.SetGroupBy(datadogGroupBys)
	}

	definition := datadogV1.FormulaAndFunctionApmResourceStatsQueryDefinitionAsFormulaAndFunctionQueryDefinition(apmResourceStatsQuery)
	return &definition
}

func buildDatadogFormulaAndFunctionAPMDependencyStatsQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionApmDependencyStatsDataSource(data["data_source"].(string))
	stat := datadogV1.FormulaAndFunctionApmDependencyStatName(data["stat"].(string))
	apmDependencyStatsQuery := datadogV1.NewFormulaAndFunctionApmDependencyStatsQueryDefinition(dataSource, data["env"].(string), data["name"].(string), data["operation_name"].(string), data["resource_name"].(string), data["service"].(string), stat)

	// cross_org_uuids
	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			apmDependencyStatsQuery.CrossOrgUuids = []string{c}
		}
	}

	// primary_tag_name
	if v, ok := data["primary_tag_name"].(string); ok && len(v) != 0 {
		apmDependencyStatsQuery.SetPrimaryTagName(v)
	}

	// primary_tag_value
	if v, ok := data["primary_tag_value"].(string); ok && len(v) != 0 {
		apmDependencyStatsQuery.SetPrimaryTagValue(v)
	}

	// is_upstream
	if v, ok := data["is_upstream"].(bool); ok {
		apmDependencyStatsQuery.SetIsUpstream(v)
	}

	definition := datadogV1.FormulaAndFunctionApmDependencyStatsQueryDefinitionAsFormulaAndFunctionQueryDefinition(apmDependencyStatsQuery)
	return &definition
}

func buildDatadogFormulaAndFunctionProcessQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionProcessQueryDataSource(data["data_source"].(string))
	processQuery := datadogV1.NewFormulaAndFunctionProcessQueryDefinition(dataSource, data["metric"].(string), data["name"].(string))

	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			processQuery.CrossOrgUuids = []string{c}
		}
	}

	// Text Filter
	if v, ok := data["text_filter"].(string); ok && len(v) != 0 {
		processQuery.SetTextFilter(v)
	}

	terraformFilters := data["tag_filters"].([]interface{})
	datadogFilters := make([]string, len(terraformFilters))
	for i, filter := range terraformFilters {
		datadogFilters[i] = filter.(string)
	}
	processQuery.SetTagFilters(datadogFilters)

	// Limit
	if v, ok := data["limit"].(int); ok && v != 0 {
		processQuery.SetLimit(int64(v))
	}

	// Aggregator
	if v, ok := data["aggregator"].(string); ok && len(v) != 0 {
		aggregator := datadogV1.FormulaAndFunctionMetricAggregation(data["aggregator"].(string))
		processQuery.SetAggregator(aggregator)
	}

	// is_normalized_cpu
	if v, ok := data["is_normalized_cpu"].(bool); ok {
		processQuery.SetIsNormalizedCpu(v)
	}

	// Sort
	if v, ok := data["sort"].(string); ok && len(v) != 0 {
		sort := datadogV1.QuerySortOrder(v)
		processQuery.SetSort(sort)
	}

	definition := datadogV1.FormulaAndFunctionProcessQueryDefinitionAsFormulaAndFunctionQueryDefinition(processQuery)
	return &definition
}

func buildDatadogFormulaAndFunctionSLOQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionSLODataSource(data["data_source"].(string))
	measure := datadogV1.FormulaAndFunctionSLOMeasure(data["measure"].(string))

	SloQuery := datadogV1.NewFormulaAndFunctionSLOQueryDefinition(dataSource, measure, data["slo_id"].(string))

	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			SloQuery.CrossOrgUuids = []string{c}
		}
	}

	if v, ok := data["group_mode"].(string); ok && len(v) != 0 {
		SloQuery.SetGroupMode(datadogV1.FormulaAndFunctionSLOGroupMode(v))
	}
	if v, ok := data["slo_query_type"].(string); ok && len(v) != 0 {
		SloQuery.SetSloQueryType(datadogV1.FormulaAndFunctionSLOQueryType(v))
	}
	if v, ok := data["name"].(string); ok && len(v) != 0 {
		SloQuery.SetName(v)
	}
	if v, ok := data["additional_query_filters"].(string); ok && len(v) != 0 {
		SloQuery.SetAdditionalQueryFilters(v)
	}

	definition := datadogV1.FormulaAndFunctionSLOQueryDefinitionAsFormulaAndFunctionQueryDefinition(SloQuery)
	return &definition
}

func buildDatadogFormulaAndFunctionCloudCostQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionCloudCostDataSource(data["data_source"].(string))

	CloudCostQuery := datadogV1.NewFormulaAndFunctionCloudCostQueryDefinition(dataSource, data["name"].(string), data["query"].(string))

	if cross_org_uuids, ok := data["cross_org_uuids"].([]interface{}); ok && len(cross_org_uuids) == 1 {
		if c, ok := cross_org_uuids[0].(string); ok && len(c) != 0 {
			CloudCostQuery.CrossOrgUuids = []string{c}
		}
	}

	if v, ok := data["aggregator"].(string); ok && len(v) != 0 {
		CloudCostQuery.SetAggregator(datadogV1.WidgetAggregator(v))
	}

	definition := datadogV1.FormulaAndFunctionCloudCostQueryDefinitionAsFormulaAndFunctionQueryDefinition(CloudCostQuery)
	return &definition
}

func buildDatadogTimeseriesRequests(terraformRequests *[]interface{}) *[]datadogV1.TimeseriesWidgetRequest {
	datadogRequests := make([]datadogV1.TimeseriesWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build TimeseriesRequest
		datadogTimeseriesRequest := datadogV1.NewTimeseriesWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogTimeseriesRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["network_query"].([]interface{}); ok && len(v) > 0 {
			networkQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.NetworkQuery = buildDatadogApmOrLogQuery(networkQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["audit_query"].([]interface{}); ok && len(v) > 0 {
			auditQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.AuditQuery = buildDatadogApmOrLogQuery(auditQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogTimeseriesRequest.SetQueries(queries)
			datadogTimeseriesRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("timeseries"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogTimeseriesRequest.SetFormulas(formulas)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogTimeseriesRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}
		// Metadata
		if terraformMetadataList, ok := terraformRequest["metadata"].([]interface{}); ok && len(terraformMetadataList) > 0 {
			datadogMetadataList := make([]datadogV1.TimeseriesWidgetExpressionAlias, len(terraformMetadataList))
			for i, m := range terraformMetadataList {
				metadata, ok := m.(map[string]interface{})
				if !ok {
					continue
				}
				// Expression
				datadogMetadata := datadogV1.NewTimeseriesWidgetExpressionAlias(metadata["expression"].(string))
				// AliasName
				if v, ok := metadata["alias_name"].(string); ok && len(v) != 0 {
					datadogMetadata.SetAliasName(v)
				}
				datadogMetadataList[i] = *datadogMetadata
			}
			datadogTimeseriesRequest.SetMetadata(datadogMetadataList)
		}
		if v, ok := terraformRequest["display_type"].(string); ok && len(v) != 0 {
			datadogTimeseriesRequest.SetDisplayType(datadogV1.WidgetDisplayType(v))
		}
		if v, ok := terraformRequest["on_right_yaxis"].(bool); ok {
			datadogTimeseriesRequest.SetOnRightYaxis(v)
		}
		datadogRequests[i] = *datadogTimeseriesRequest
	}
	return &datadogRequests
}
func buildTerraformTimeseriesRequests(datadogTimeseriesRequests *[]datadogV1.TimeseriesWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogTimeseriesRequests))
	for i, datadogRequest := range *datadogTimeseriesRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetNetworkQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["network_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetSecurityQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetAuditQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["audit_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}

		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetRequestStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
		}
		// Metadata
		if datadogRequest.Metadata != nil {
			terraformMetadataList := make([]map[string]interface{}, len(datadogRequest.GetMetadata()))
			for i, metadata := range datadogRequest.GetMetadata() {
				// Expression
				terraformMetadata := map[string]interface{}{
					"expression": metadata.GetExpression(),
				}
				// AliasName
				if metadata.AliasName != nil {
					terraformMetadata["alias_name"] = metadata.GetAliasName()
				}

				terraformMetadataList[i] = terraformMetadata
			}
			terraformRequest["metadata"] = &terraformMetadataList
		}
		if v, ok := datadogRequest.GetDisplayTypeOk(); ok {
			terraformRequest["display_type"] = v
		}
		if v, ok := datadogRequest.GetOnRightYaxisOk(); ok {
			terraformRequest["on_right_yaxis"] = v
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildDatadogToplistDefinition(terraformDefinition map[string]interface{}) *datadogV1.ToplistWidgetDefinition {
	datadogDefinition := datadogV1.NewToplistWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogToplistRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}

	if style, ok := terraformDefinition["style"].([]interface{}); ok && len(style) > 0 {
		if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.SetStyle(buildDatadogToplistStyle(v))
		}
	}
	return datadogDefinition
}

func buildDatadogToplistStyle(terraformToplistStyle map[string]interface{}) datadogV1.ToplistWidgetStyle {
	datadogToplistStyle := datadogV1.NewToplistWidgetStyleWithDefaults()

	if display, ok := terraformToplistStyle["display"].([]interface{}); ok && len(display) > 0 {
		if v, ok := display[0].(map[string]interface{}); ok && len(v) > 0 {
			if t, ok := v["type"].(string); ok && len(t) != 0 {
				if t == "stacked" {
					toplistWidgetStacked := &datadogV1.ToplistWidgetStacked{
						Legend: datadogV1.TOPLISTWIDGETLEGEND_AUTOMATIC.Ptr(),
						Type:   datadogV1.TOPLISTWIDGETSTACKEDTYPE_STACKED,
					}
					datadogToplistStyle.SetDisplay(datadogV1.ToplistWidgetDisplay{
						ToplistWidgetStacked: toplistWidgetStacked,
					})
				} else if t == "flat" {
					datadogToplistStyle.SetDisplay(datadogV1.ToplistWidgetDisplay{
						ToplistWidgetFlat: datadogV1.NewToplistWidgetFlatWithDefaults(),
					})
				}
			}
		}
	}
	if palette, ok := terraformToplistStyle["palette"].(string); ok && len(palette) != 0 {
		datadogToplistStyle.SetPalette(palette)
	}
	if scaling, ok := terraformToplistStyle["scaling"].(string); ok && len(scaling) != 0 {
		datadogToplistStyle.SetScaling(datadogV1.ToplistWidgetScaling(scaling))
	}
	return *datadogToplistStyle
}

func buildTerraformToplistDefinition(datadogDefinition *datadogV1.ToplistWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformToplistRequests(&datadogDefinition.Requests)

	// Optional params
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	if v, ok := datadogDefinition.GetStyleOk(); ok {
		terraformDefinition["style"] = buildTerraformToplistWidgetStyle(v)
	}
	return terraformDefinition
}
func buildDatadogToplistRequests(terraformRequests *[]interface{}) *[]datadogV1.ToplistWidgetRequest {
	datadogRequests := make([]datadogV1.ToplistWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build ToplistRequest
		datadogToplistRequest := datadogV1.NewToplistWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogToplistRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogToplistRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogToplistRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogToplistRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogToplistRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogToplistRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["audit_query"].([]interface{}); ok && len(v) > 0 {
			auditQuery := v[0].(map[string]interface{})
			datadogToplistRequest.AuditQuery = buildDatadogApmOrLogQuery(auditQuery)
		} else if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogToplistRequest.SetQueries(queries)
			// Toplist requests for formulas and functions always has a response format of "scalar"
			datadogToplistRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogToplistRequest.SetFormulas(formulas)
		}
		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogToplistRequest.ConditionalFormats = *buildDatadogWidgetConditionalFormat(&v)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogToplistRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}
		datadogRequests[i] = *datadogToplistRequest
	}
	return &datadogRequests
}
func buildTerraformToplistRequests(datadogToplistRequests *[]datadogV1.ToplistWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogToplistRequests))
	for i, datadogRequest := range *datadogToplistRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetSecurityQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetAuditQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["audit_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}

		if v, ok := datadogRequest.GetConditionalFormatsOk(); ok {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(v)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetRequestStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildTerraformToplistWidgetStyle(datadogToplistStyle *datadogV1.ToplistWidgetStyle) *[]map[string]interface{} {
	terraformStyles := make([]map[string]interface{}, 1)
	terraformStyle := map[string]interface{}{}
	if display, ok := datadogToplistStyle.GetDisplayOk(); ok {
		terraformDisplays := make([]map[string]interface{}, 1)
		terraformDisplay := map[string]interface{}{}
		if display.ToplistWidgetStacked != nil {
			terraformDisplay["type"] = datadogV1.TOPLISTWIDGETSTACKEDTYPE_STACKED
		} else if display.ToplistWidgetFlat != nil {
			terraformDisplay["type"] = datadogV1.TOPLISTWIDGETFLATTYPE_FLAT
		}
		terraformDisplays[0] = terraformDisplay
		terraformStyle["display"] = terraformDisplays
	}
	if palette, ok := datadogToplistStyle.GetPaletteOk(); ok {
		terraformStyle["palette"] = palette
	}
	if scaling, ok := datadogToplistStyle.GetScalingOk(); ok {
		terraformStyle["scaling"] = scaling
	}
	terraformStyles[0] = terraformStyle
	return &terraformStyles
}

func buildDatadogSplitGraphDefinition(terraformDefinition map[string]interface{}) (*datadogV1.SplitGraphWidgetDefinition, error) {
	datadogDefinition := datadogV1.NewSplitGraphWidgetDefinitionWithDefaults()
	// Required params
	//size,source_widget,split_config, type
	if size, ok := terraformDefinition["size"].(string); ok && size != "" {
		datadogDefinition.SetSize(datadogV1.SplitGraphVizSize(size))
	}

	if terraformSourceWidget, ok := terraformDefinition["source_widget_definition"].([]interface{}); ok && len(terraformSourceWidget) > 0 {
		if v, ok := terraformSourceWidget[0].(map[string]interface{}); ok {
			datadogWidget, err := buildDatadogSourceWidgetDefinition(v)
			if err != nil {
				return nil, err
			}
			datadogDefinition.SetSourceWidgetDefinition(*datadogWidget)
		} else {
			return nil, fmt.Errorf("failed to find valid definition in widget configuration")
		}
	}

	if terraformSplitConfig, ok := terraformDefinition["split_config"].([]interface{}); ok && len(terraformSplitConfig) > 0 {
		datadogDefinition.SetSplitConfig(*buildDatadogSplitConfig(terraformSplitConfig[0].(map[string]interface{})))
	}

	//optional params
	if yAxes, ok := terraformDefinition["has_uniform_y_axes"].(bool); ok {
		datadogDefinition.SetHasUniformYAxes(yAxes)
	}

	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.PtrString(v)
	}
	if widgetTime := buildDatadogWidgetTime(terraformDefinition); widgetTime != nil {
		datadogDefinition.Time = widgetTime
	}

	return datadogDefinition, nil
}

func buildDatadogPowerpackDefinition(terraformDefinition map[string]interface{}) (*datadogV1.PowerpackWidgetDefinition, error) {
	datadogDefinition := datadogV1.NewPowerpackWidgetDefinitionWithDefaults()
	// Required params
	//type, powerpack_id

	powerpack_type, _ := datadogV1.NewPowerpackWidgetDefinitionTypeFromValue("powerpack")
	datadogDefinition.SetType(*powerpack_type)

	if powerpack_id, ok := terraformDefinition["powerpack_id"].(string); ok && powerpack_id != "" {
		datadogDefinition.SetPowerpackId(powerpack_id)
	}

	if background_color, ok := terraformDefinition["background_color"].(string); ok && background_color != "" {
		datadogDefinition.SetBackgroundColor(background_color)
	}

	if banner_img, ok := terraformDefinition["banner_img"].(string); ok && banner_img != "" {
		datadogDefinition.SetBannerImg(banner_img)
	}

	if show_title, ok := terraformDefinition["show_title"].(bool); ok {
		datadogDefinition.SetShowTitle(show_title)
	}

	if template_variables, ok := terraformDefinition["template_variables"].([]interface{}); ok {
		ppkTVars := datadogV1.NewPowerpackTemplateVariablesWithDefaults()
		tvars := template_variables[0].(map[string]interface{})
		if tfControlledByPowerpack, ok := tvars["controlled_by_powerpack"].([]interface{}); ok {
			ppkTVars.SetControlledByPowerpack(buildDatadogPowerpackTVarContents(tfControlledByPowerpack))
		}
		if tfControlledExternally, ok := tvars["controlled_externally"].([]interface{}); ok {
			ppkTVars.SetControlledExternally(buildDatadogPowerpackTVarContents(tfControlledExternally))
		}
		datadogDefinition.SetTemplateVariables(*ppkTVars)
	}

	if title, ok := terraformDefinition["title"].(string); ok && title != "" {
		datadogDefinition.SetTitle(title)
	}

	return datadogDefinition, nil
}

func buildDatadogSplitConfig(terraformSplitConfig map[string]interface{}) *datadogV1.SplitConfig {
	datadogSplitConfig := datadogV1.NewSplitConfigWithDefaults()

	if limit, ok := terraformSplitConfig["limit"].(int); ok {
		datadogSplitConfig.SetLimit(int64(limit))
	}

	if sort, ok := terraformSplitConfig["sort"].([]interface{}); ok && len(sort) > 0 {
		datadogSplitConfig.SetSort(*buildDatadogSplitSort(sort[0].(map[string]interface{})))
	}

	if splitDimensions, ok := terraformSplitConfig["split_dimensions"].([]interface{}); ok && len(splitDimensions) > 0 {
		terraformSplitDimension := splitDimensions[0].(map[string]interface{})
		if v, ok := terraformSplitDimension["one_graph_per"].(string); ok && len(v) > 0 {
			datadogSplitDimensions := make([]datadogV1.SplitDimension, 1)
			datadogSplitDimensions[0] = *datadogV1.NewSplitDimension(v)
			datadogSplitConfig.SetSplitDimensions(datadogSplitDimensions)
		}
	}

	if v, ok := terraformSplitConfig["static_splits"].([]interface{}); ok && len(v) > 0 {
		datadogStaticSplits := buildDatadogStaticSplits(v)
		datadogSplitConfig.SetStaticSplits(*datadogStaticSplits)
	}
	return datadogSplitConfig
}

func buildDatadogSplitSort(terraformSplitSort map[string]interface{}) *datadogV1.SplitSort {
	datadogSplitSort := datadogV1.SplitSort{}

	if order, ok := terraformSplitSort["order"].(string); ok && len(order) > 0 {
		datadogSplitSort.SetOrder(datadogV1.WidgetSort(order))
	}

	if compute, ok := terraformSplitSort["compute"].([]interface{}); ok && len(compute) > 0 {
		sortCompute := compute[0].(map[string]interface{})
		var datadogSortAggregation string
		var datadogSortMetric string
		if aggregation, ok := sortCompute["aggregation"].(string); ok && len(aggregation) > 0 {
			datadogSortAggregation = aggregation
		}
		if metric, ok := sortCompute["metric"].(string); ok && len(metric) > 0 {
			datadogSortMetric = metric
		}
		datadogSplitSort.SetCompute(*datadogV1.NewSplitConfigSortCompute(datadogSortAggregation, datadogSortMetric))
	}
	return &datadogSplitSort
}

// Build static splits for backend  format from static splits
func buildDatadogStaticSplits(terraformStaticSplits []interface{}) *[][]datadogV1.SplitVectorEntryItem {
	datadogStaticSplits := make([][]datadogV1.SplitVectorEntryItem, len(terraformStaticSplits))
	//going over each static split
	for i, terraformStaticSplit := range terraformStaticSplits {
		terraformStaticSplitMap := terraformStaticSplit.(map[string]interface{})
		//building inner array for static split from terraform split vector list.
		for _, splitVector := range terraformStaticSplitMap["split_vector"].([]interface{}) {
			datadogSplitVectorMap := splitVector.(map[string]interface{})

			datadogSplitVector := datadogV1.SplitVectorEntryItem{}
			if v, ok := datadogSplitVectorMap["tag_key"].(string); ok {
				datadogSplitVector.SetTagKey(v)
			}
			if tagValuesList, ok := datadogSplitVectorMap["tag_values"].([]interface{}); ok {
				datadogTagValues := make([]string, len(tagValuesList))
				for k, tagValues := range tagValuesList {
					datadogTagValues[k] = tagValues.(string)
				}
				datadogSplitVector.SetTagValues(datadogTagValues)
			}
			datadogStaticSplits[i] = append(datadogStaticSplits[i], datadogSplitVector)
		}
	}
	return &datadogStaticSplits
}

func buildTerraformSplitGraphDefinition(datadogDefinition *datadogV1.SplitGraphWidgetDefinition) (map[string]interface{}, error) {
	terraformDefinition := map[string]interface{}{}
	// Required params
	if v, ok := datadogDefinition.GetSourceWidgetDefinitionOk(); ok {
		terraformSourceWidgetDefinition, err := buildTerraformSourceWidgetDefinition(v)
		if err != nil {
			return nil, err
		}
		terraformDefinition["source_widget_definition"] = []map[string]interface{}{terraformSourceWidgetDefinition}
	}
	if v, ok := datadogDefinition.GetSizeOk(); ok {
		terraformDefinition["size"] = v
	}
	if v, ok := datadogDefinition.GetSplitConfigOk(); ok {
		terraformDefinition["split_config"] = []map[string]interface{}{*buildTerraformSplitConfig(v)}
	}
	// Optional params
	if v, ok := datadogDefinition.GetHasUniformYAxesOk(); ok {
		terraformDefinition["has_uniform_y_axes"] = *v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}

	return terraformDefinition, nil
}

func buildTerraformPowerpackDefinition(datadogDefinition *datadogV1.PowerpackWidgetDefinition) (map[string]interface{}, error) {
	terraformDefinition := map[string]interface{}{}
	// Required params: powerpack_id
	if v, ok := datadogDefinition.GetPowerpackIdOk(); ok {
		terraformDefinition["powerpack_id"] = v
	}
	if v, ok := datadogDefinition.GetBackgroundColorOk(); ok {
		terraformDefinition["background_color"] = v
	}
	if v, ok := datadogDefinition.GetBannerImgOk(); ok {
		terraformDefinition["banner_img"] = v
	}
	if v, ok := datadogDefinition.GetShowTitleOk(); ok {
		terraformDefinition["show_title"] = v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = v
	}
	if templateVariables, ok := datadogDefinition.GetTemplateVariablesOk(); ok {
		terraformTemplateVariables := make([]map[string]interface{}, 1)
		terraformTemplateVariable := map[string]interface{}{}

		if ddControlledByPowerpack, ok := templateVariables.GetControlledByPowerpackOk(); ok {
			controlledByPowerpackTVars := buildTerraformPowerpackTVarContents(*ddControlledByPowerpack)
			terraformTemplateVariable["controlled_by_powerpack"] = controlledByPowerpackTVars
		}
		if ddControlledExternally, ok := templateVariables.GetControlledExternallyOk(); ok {
			controlledExternallyTVars := buildTerraformPowerpackTVarContents(*ddControlledExternally)
			terraformTemplateVariable["controlled_externally"] = controlledExternallyTVars
		}
		terraformTemplateVariables[0] = terraformTemplateVariable

		terraformDefinition["template_variables"] = terraformTemplateVariables
	}

	return terraformDefinition, nil
}

func buildTerraformSplitConfig(datadogSplitConfig *datadogV1.SplitConfig) *map[string]interface{} {
	terraformSplitConfig := map[string]interface{}{}

	if v, ok := datadogSplitConfig.GetSplitDimensionsOk(); ok {
		datadogSplitDimensions := *v
		terraformOneGraphPer := map[string]interface{}{}

		terraformOneGraphPer["one_graph_per"] = datadogSplitDimensions[0].OneGraphPer
		terraformSplitConfig["split_dimensions"] = []map[string]interface{}{terraformOneGraphPer}
	}

	if v, ok := datadogSplitConfig.GetLimitOk(); ok {
		terraformSplitConfig["limit"] = *v
	}

	if datadogSort, ok := datadogSplitConfig.GetSortOk(); ok {
		terraformSortList := []map[string]interface{}{
			{
				"order": datadogSort.Order,
			},
		}

		if datadogSortCompute, datadogSortComputeOk := datadogSort.GetComputeOk(); datadogSortComputeOk {
			terraformSortList[0]["compute"] = []map[string]interface{}{
				{
					"aggregation": datadogSortCompute.Aggregation,
					"metric":      datadogSortCompute.Metric,
				},
			}
		}
		terraformSplitConfig["sort"] = terraformSortList
	}
	if v, ok := datadogSplitConfig.GetStaticSplitsOk(); ok {
		terraformSplitConfig["static_splits"] = buildTerraformStaticSplits(v)
	}
	return &terraformSplitConfig
}

// Build static splits for terraform from backend format
func buildTerraformStaticSplits(datadogStaticSplits *[][]datadogV1.SplitVectorEntryItem) *[]interface{} {
	terraformStaticSplits := make([]interface{}, len(*datadogStaticSplits))
	for i, staticSplit := range *datadogStaticSplits {
		terraformSplitVectors := make([]map[string]interface{}, len(staticSplit))
		for j, splitVector := range staticSplit {
			terraformSplitVectors[j] = map[string]interface{}{
				"tag_key":    splitVector.GetTagKey(),
				"tag_values": splitVector.GetTagValues(),
			}
		}
		terraformStaticSplits[i] = map[string]interface{}{
			"split_vector": terraformSplitVectors,
		}
	}
	return &terraformStaticSplits
}

func buildDatadogTreemapDefinition(terraformDefinition map[string]interface{}) *datadogV1.TreeMapWidgetDefinition {
	datadogDefinition := datadogV1.NewTreeMapWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogTreemapRequests(&terraformRequests)

	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}

	if v, ok := terraformDefinition["custom_links"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}

	return datadogDefinition
}

func buildDatadogTreemapRequests(terraformRequests *[]interface{}) *[]datadogV1.TreeMapWidgetRequest {
	datadogRequests := make([]datadogV1.TreeMapWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build Treemap request
		datadogTreemapRequest := datadogV1.NewTreeMapWidgetRequest()
		if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			queries := make([]datadogV1.FormulaAndFunctionQueryDefinition, len(v))
			for i, q := range v {
				query := q.(map[string]interface{})
				if w, ok := query["event_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogEventQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["metric_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogMetricQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["process_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionProcessQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["slo_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionSLOQuery(w[0].(map[string]interface{}))
				} else if w, ok := query["cloud_cost_query"].([]interface{}); ok && len(w) > 0 {
					queries[i] = *buildDatadogFormulaAndFunctionCloudCostQuery(w[0].(map[string]interface{}))
				}
			}
			datadogTreemapRequest.SetQueries(queries)
			datadogTreemapRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("scalar"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogTreemapRequest.SetFormulas(formulas)
		}
		datadogRequests[i] = *datadogTreemapRequest
	}
	return &datadogRequests
}

func buildTerraformTreemapRequests(datadogTreemapRequests *[]datadogV1.TreeMapWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogTreemapRequests))
	for i, datadogRequest := range *datadogTreemapRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQueriesOk(); ok {
			terraformRequest["query"] = buildTerraformQuery(v)
		}

		if v, ok := datadogRequest.GetFormulasOk(); ok {
			terraformRequest["formula"] = buildTerraformFormula(v, false)
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildTerraformTreemapDefinition(datadogDefinition *datadogV1.TreeMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformTreemapRequests(&datadogDefinition.Requests)

	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}

	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_links"] = buildTerraformWidgetCustomLinks(v)
	}

	return terraformDefinition
}

func buildTerraformGeomapDefinition(datadogDefinition *datadogV1.GeomapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformGeomapRequests(&datadogDefinition.Requests)

	if v, ok := datadogDefinition.GetStyleOk(); ok {
		style := buildTerraformGeomapRequestStyle(*v)
		terraformDefinition["style"] = []map[string]interface{}{style}
	}

	if v, ok := datadogDefinition.GetViewOk(); ok {
		view := buildTerraformGeomapRequestView(*v)
		terraformDefinition["view"] = []map[string]interface{}{view}
	}

	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}

	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}

	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}

	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}

	if v, ok := datadogDefinition.GetTimeOk(); ok {
		buildTerraformWidgetTime(v, terraformDefinition)
	}

	return terraformDefinition
}

// Widget Conditional Format helpers
func buildDatadogWidgetConditionalFormat(terraformWidgetConditionalFormat *[]interface{}) *[]datadogV1.WidgetConditionalFormat {
	datadogWidgetConditionalFormat := make([]datadogV1.WidgetConditionalFormat, len(*terraformWidgetConditionalFormat))
	for i, conditionalFormat := range *terraformWidgetConditionalFormat {
		terraformConditionalFormat := conditionalFormat.(map[string]interface{})
		datadogConditionalFormat := datadogV1.NewWidgetConditionalFormat(
			datadogV1.WidgetComparator(terraformConditionalFormat["comparator"].(string)),
			datadogV1.WidgetPalette(terraformConditionalFormat["palette"].(string)),
			terraformConditionalFormat["value"].(float64))
		// Optional
		if v, ok := terraformConditionalFormat["custom_bg_color"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetCustomBgColor(v)
		}
		if v, ok := terraformConditionalFormat["custom_fg_color"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetCustomFgColor(v)
		}
		if v, ok := terraformConditionalFormat["image_url"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetImageUrl(v)
		}
		if v, ok := terraformConditionalFormat["hide_value"].(bool); ok {
			datadogConditionalFormat.SetHideValue(v)
		}
		if v, ok := terraformConditionalFormat["timeframe"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetTimeframe(v)
		}
		if v, ok := terraformConditionalFormat["metric"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetMetric(v)
		}
		datadogWidgetConditionalFormat[i] = *datadogConditionalFormat
	}
	return &datadogWidgetConditionalFormat
}
func buildTerraformWidgetConditionalFormat(datadogWidgetConditionalFormat *[]datadogV1.WidgetConditionalFormat) *[]map[string]interface{} {
	terraformWidgetConditionalFormat := make([]map[string]interface{}, len(*datadogWidgetConditionalFormat))
	for i, datadogConditionalFormat := range *datadogWidgetConditionalFormat {
		terraformConditionalFormat := map[string]interface{}{}
		// Required params
		terraformConditionalFormat["comparator"] = datadogConditionalFormat.GetComparator()
		terraformConditionalFormat["value"] = datadogConditionalFormat.GetValue()
		terraformConditionalFormat["palette"] = datadogConditionalFormat.GetPalette()
		// Optional params
		if datadogConditionalFormat.CustomBgColor != nil {
			terraformConditionalFormat["custom_bg_color"] = datadogConditionalFormat.GetCustomBgColor()
		}
		if v, ok := datadogConditionalFormat.GetCustomFgColorOk(); ok {
			terraformConditionalFormat["custom_fg_color"] = v
		}
		if v, ok := datadogConditionalFormat.GetImageUrlOk(); ok {
			terraformConditionalFormat["image_url"] = v
		}
		if v, ok := datadogConditionalFormat.GetHideValueOk(); ok {
			terraformConditionalFormat["hide_value"] = v
		}
		if v, ok := datadogConditionalFormat.GetTimeframeOk(); ok {
			terraformConditionalFormat["timeframe"] = v
		}
		if v, ok := datadogConditionalFormat.GetMetricOk(); ok {
			terraformConditionalFormat["metric"] = v
		}
		terraformWidgetConditionalFormat[i] = terraformConditionalFormat
	}
	return &terraformWidgetConditionalFormat
}

// Widget Custom Link helpers

// Toplist Widget Style helpers

// Widget Display helper

func buildDatadogTimeseriesBackground(terraformTimeseriesBackground map[string]interface{}) *datadogV1.TimeseriesBackground {
	datadogTimeseriesBackground := &datadogV1.TimeseriesBackground{}
	if v, ok := terraformTimeseriesBackground["type"].(string); ok && len(v) != 0 {
		timeseriesBackgroundType := datadogV1.TimeseriesBackgroundType(terraformTimeseriesBackground["type"].(string))
		datadogTimeseriesBackground.SetType(timeseriesBackgroundType)
	}

	// Optional params
	if axis, ok := terraformTimeseriesBackground["yaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogTimeseriesBackground.Yaxis = buildDatadogWidgetAxis(v)
		}
	}

	return datadogTimeseriesBackground
}

func buildDatadogWidgetCustomLinks(terraformWidgetCustomLinks *[]interface{}) *[]datadogV1.WidgetCustomLink {
	datadogWidgetCustomLinks := make([]datadogV1.WidgetCustomLink, len(*terraformWidgetCustomLinks))
	for i, customLink := range *terraformWidgetCustomLinks {
		terraformCustomLink := customLink.(map[string]interface{})
		datadogWidgetCustomLink := datadogV1.WidgetCustomLink{}
		if v, ok := terraformCustomLink["override_label"].(string); ok && len(v) > 0 {
			datadogWidgetCustomLink.SetOverrideLabel(v)
		}
		// if override_label is provided, the label field will be omitted.
		if v, ok := terraformCustomLink["label"].(string); ok && len(v) > 0 && !datadogWidgetCustomLink.HasOverrideLabel() {
			datadogWidgetCustomLink.SetLabel(v)
		}
		if v, ok := terraformCustomLink["is_hidden"].(bool); ok && v && datadogWidgetCustomLink.HasOverrideLabel() {
			datadogWidgetCustomLink.SetIsHidden(v)
		}
		if v, ok := terraformCustomLink["link"].(string); ok && len(v) > 0 {
			datadogWidgetCustomLink.SetLink(v)
		}
		datadogWidgetCustomLinks[i] = datadogWidgetCustomLink
	}
	return &datadogWidgetCustomLinks
}
func buildTerraformWidgetCustomLinks(datadogWidgetCustomLinks *[]datadogV1.WidgetCustomLink) *[]map[string]interface{} {
	terraformWidgetCustomLinks := make([]map[string]interface{}, len(*datadogWidgetCustomLinks))
	for i, customLink := range *datadogWidgetCustomLinks {
		terraformWidgetCustomLink := map[string]interface{}{}
		// Optional params
		if v, ok := customLink.GetLabelOk(); ok {
			terraformWidgetCustomLink["label"] = *v
		}
		if v, ok := customLink.GetLinkOk(); ok {
			terraformWidgetCustomLink["link"] = *v
		}
		if v, ok := customLink.GetOverrideLabelOk(); ok {
			terraformWidgetCustomLink["override_label"] = *v
		}
		if v, ok := customLink.GetIsHiddenOk(); ok {
			terraformWidgetCustomLink["is_hidden"] = *v
		}
		terraformWidgetCustomLinks[i] = terraformWidgetCustomLink
	}
	return &terraformWidgetCustomLinks
}

// Widget Event helpers
func buildDatadogWidgetEvents(terraformWidgetEvents *[]interface{}) *[]datadogV1.WidgetEvent {
	datadogWidgetEvents := make([]datadogV1.WidgetEvent, len(*terraformWidgetEvents))
	for i, event := range *terraformWidgetEvents {
		terraformEvent := event.(map[string]interface{})
		datadogWidgetEvent := datadogV1.NewWidgetEvent(terraformEvent["q"].(string))
		if v, ok := terraformEvent["tags_execution"].(string); ok && len(v) > 0 {
			datadogWidgetEvent.SetTagsExecution(v)
		}
		datadogWidgetEvents[i] = *datadogWidgetEvent
	}

	return &datadogWidgetEvents
}
func buildTerraformWidgetEvents(datadogWidgetEvents *[]datadogV1.WidgetEvent) *[]map[string]string {
	terraformWidgetEvents := make([]map[string]string, len(*datadogWidgetEvents))
	for i, datadogWidget := range *datadogWidgetEvents {
		terraformWidget := map[string]string{}
		// Required params
		terraformWidget["q"] = datadogWidget.GetQ()
		// Optional params
		if v, ok := datadogWidget.GetTagsExecutionOk(); ok {
			terraformWidget["tags_execution"] = *v
		}

		terraformWidgetEvents[i] = terraformWidget
	}
	return &terraformWidgetEvents
}

// Widget Time helpers

// widgetLiveSpanUnitToAbbrev converts WidgetLiveSpanUnit enum to legacy format abbreviation
func widgetLiveSpanUnitToAbbrev(unit datadogV1.WidgetLiveSpanUnit) string {
	switch unit {
	case "minute":
		return "m"
	case "hour":
		return "h"
	case "day":
		return "d"
	case "week":
		return "w"
	case "month":
		return "mo"
	case "year":
		return "y"
	default:
		return string(unit)
	}
}

// buildDatadogWidgetTime creates a WidgetTime from Terraform definition.
// Always uses legacy format - API will handle conversion if needed.
func buildDatadogWidgetTime(terraformDefinition map[string]interface{}) *datadogV1.WidgetTime {
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		widgetLegacyLiveSpan := &datadogV1.WidgetLegacyLiveSpan{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
		// Only set hide_incomplete_cost_data if explicitly present in config
		if hic, ok := terraformDefinition["hide_incomplete_cost_data"].(bool); ok && hic {
			widgetLegacyLiveSpan.SetHideIncompleteCostData(hic)
		}
		return &datadogV1.WidgetTime{
			WidgetLegacyLiveSpan: widgetLegacyLiveSpan,
		}
	}
	return nil
}

// buildTerraformWidgetTime extracts time settings from a WidgetTime union into a Terraform definition map.
// Handles WidgetLegacyLiveSpan, WidgetNewLiveSpan, and UnparsedObject fallback.
func buildTerraformWidgetTime(widgetTime *datadogV1.WidgetTime, terraformDefinition map[string]interface{}) {
	if widgetTime == nil {
		return
	}

	if widgetTime.WidgetLegacyLiveSpan != nil {
		// Legacy format: {"live_span": "1h"}
		terraformDefinition["live_span"] = widgetTime.WidgetLegacyLiveSpan.GetLiveSpan()
		// Only set hide_incomplete_cost_data if true (API defaults to false)
		if hic, ok := widgetTime.WidgetLegacyLiveSpan.GetHideIncompleteCostDataOk(); ok && *hic {
			terraformDefinition["hide_incomplete_cost_data"] = true
		}
	} else if widgetTime.WidgetNewLiveSpan != nil {
		// New format: {"type": "live", "unit": "hour", "value": 1}
		// Convert back to legacy format for Terraform
		unit := widgetTime.WidgetNewLiveSpan.GetUnit()
		value := widgetTime.WidgetNewLiveSpan.GetValue()
		unitAbbrev := widgetLiveSpanUnitToAbbrev(unit)
		terraformDefinition["live_span"] = fmt.Sprintf("%d%s", value, unitAbbrev)
		// Only set hide_incomplete_cost_data if true (API defaults to false)
		if hic, ok := widgetTime.WidgetNewLiveSpan.GetHideIncompleteCostDataOk(); ok && *hic {
			terraformDefinition["hide_incomplete_cost_data"] = true
		}
	} else if widgetTime.UnparsedObject != nil {
		// Handle unparsed WidgetTime (due to oneOf ambiguity)
		// Use the API client's own UnmarshalJSON methods to parse
		unparsedBytes, err := json.Marshal(widgetTime.UnparsedObject)
		if err == nil {
			// Try new format first (more specific - has required fields)
			var newLiveSpan datadogV1.WidgetNewLiveSpan
			if err := json.Unmarshal(unparsedBytes, &newLiveSpan); err == nil && newLiveSpan.UnparsedObject == nil {
				// Successfully parsed as new format
				unit := newLiveSpan.GetUnit()
				value := newLiveSpan.GetValue()
				unitAbbrev := widgetLiveSpanUnitToAbbrev(unit)
				terraformDefinition["live_span"] = fmt.Sprintf("%d%s", value, unitAbbrev)
				// Only set hide_incomplete_cost_data if true (API defaults to false)
				if hic, ok := newLiveSpan.GetHideIncompleteCostDataOk(); ok && *hic {
					terraformDefinition["hide_incomplete_cost_data"] = true
				}
			} else {
				// Try legacy format
				var legacyLiveSpan datadogV1.WidgetLegacyLiveSpan
				if err := json.Unmarshal(unparsedBytes, &legacyLiveSpan); err == nil && legacyLiveSpan.UnparsedObject == nil {
					terraformDefinition["live_span"] = legacyLiveSpan.GetLiveSpan()
					// Only set hide_incomplete_cost_data if true (API defaults to false)
					if hic, ok := legacyLiveSpan.GetHideIncompleteCostDataOk(); ok && *hic {
						terraformDefinition["hide_incomplete_cost_data"] = true
					}
				}
			}
		}
	}
}

// Widget Marker helpers
func buildDatadogWidgetMarkers(terraformWidgetMarkers *[]interface{}) *[]datadogV1.WidgetMarker {
	datadogWidgetMarkers := make([]datadogV1.WidgetMarker, len(*terraformWidgetMarkers))
	for i, marker := range *terraformWidgetMarkers {
		terraformMarker := marker.(map[string]interface{})
		// Required
		datadogMarker := datadogV1.NewWidgetMarker(terraformMarker["value"].(string))
		// Optional
		if v, ok := terraformMarker["display_type"].(string); ok && len(v) != 0 {
			datadogMarker.SetDisplayType(v)
		}
		if v, ok := terraformMarker["label"].(string); ok && len(v) != 0 {
			datadogMarker.SetLabel(v)
		}
		datadogWidgetMarkers[i] = *datadogMarker
	}
	return &datadogWidgetMarkers
}
func buildTerraformWidgetMarkers(datadogWidgetMarkers *[]datadogV1.WidgetMarker) *[]map[string]string {
	terraformWidgetMarkers := make([]map[string]string, len(*datadogWidgetMarkers))
	for i, datadogMarker := range *datadogWidgetMarkers {
		terraformMarker := map[string]string{}
		// Required params
		terraformMarker["value"] = datadogMarker.Value
		// Optional params
		if v, ok := datadogMarker.GetDisplayTypeOk(); ok {
			terraformMarker["display_type"] = *v
		}
		if v, ok := datadogMarker.GetLabelOk(); ok {
			terraformMarker["label"] = *v
		}
		terraformWidgetMarkers[i] = terraformMarker
	}
	return &terraformWidgetMarkers
}

//
// Widget Query helpers
//

// Metric Query
// APM, Log, Network, RUM or Audit Query
func buildDatadogQueryCompute(terraformCompute map[string]interface{}) *datadogV1.LogsQueryCompute {
	datadogCompute := datadogV1.NewLogsQueryComputeWithDefaults()
	if aggr, ok := terraformCompute["aggregation"].(string); ok && len(aggr) != 0 {
		datadogCompute.SetAggregation(aggr)
		if facet, ok := terraformCompute["facet"].(string); ok && len(facet) != 0 {
			datadogCompute.SetFacet(facet)
		}
		if interval, ok := terraformCompute["interval"].(int); ok && interval != 0 {
			datadogCompute.SetInterval(int64(interval))
		}
	}
	return datadogCompute
}

func buildDatadogApmOrLogQuery(terraformQuery map[string]interface{}) *datadogV1.LogQueryDefinition {
	// Index
	datadogQuery := datadogV1.NewLogQueryDefinition()
	datadogQuery.SetIndex(terraformQuery["index"].(string))

	// Compute
	if terraformComputeList, ok := terraformQuery["compute_query"].([]interface{}); ok && len(terraformComputeList) != 0 {
		if terraformCompute, ok := terraformComputeList[0].(map[string]interface{}); ok {
			datadogQuery.SetCompute(*buildDatadogQueryCompute(terraformCompute))
		}
	}
	// Multi-compute
	terraformMultiCompute := terraformQuery["multi_compute"].([]interface{})
	if len(terraformMultiCompute) > 0 {
		// TODO: raise an error if compute is already set
		datadogComputeList := make([]datadogV1.LogsQueryCompute, len(terraformMultiCompute))
		for i, terraformCompute := range terraformMultiCompute {
			terraformComputeMap := terraformCompute.(map[string]interface{})
			datadogCompute := datadogV1.NewLogsQueryComputeWithDefaults()
			if aggr, ok := terraformComputeMap["aggregation"].(string); ok && len(aggr) != 0 {
				datadogCompute.SetAggregation(aggr)
			}
			if facet, ok := terraformComputeMap["facet"].(string); ok && len(facet) != 0 {
				datadogCompute.SetFacet(facet)
			}
			if interval, ok := terraformComputeMap["interval"].(int); ok && interval != 0 {
				datadogCompute.SetInterval(int64(interval))
			}
			datadogComputeList[i] = *datadogCompute
		}
		datadogQuery.SetMultiCompute(datadogComputeList)
	}
	// Search
	if terraformSearchQuery, ok := terraformQuery["search_query"].(string); ok {
		datadogQuery.Search = &datadogV1.LogQueryDefinitionSearch{
			Query: terraformSearchQuery,
		}
	}
	// GroupBy
	if terraformGroupBys, ok := terraformQuery["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]datadogV1.LogQueryDefinitionGroupBy, len(terraformGroupBys))
		for i, g := range terraformGroupBys {
			if groupBy, ok := g.(map[string]interface{}); ok {
				// Facet
				datadogGroupBy := datadogV1.NewLogQueryDefinitionGroupBy(groupBy["facet"].(string))
				// Limit
				if v, ok := groupBy["limit"].(int); ok && v != 0 {
					datadogGroupBy.SetLimit(int64(v))
				}
				// Sort
				if sortList, ok := groupBy["sort_query"].([]interface{}); ok && len(sortList) > 0 {
					if sort, ok := sortList[0].(map[string]interface{}); ok && len(sort) > 0 {
						datadogGroupBy.Sort = buildDatadogGroupBySort(sort)
					}
				}
				datadogGroupBys[i] = *datadogGroupBy
			}
		}
		datadogQuery.SetGroupBy(datadogGroupBys)
	}
	return datadogQuery
}

func buildDatadogGroupBySort(sort map[string]interface{}) *datadogV1.LogQueryDefinitionGroupBySort {
	ddSort := &datadogV1.LogQueryDefinitionGroupBySort{}
	if aggr, ok := sort["aggregation"].(string); ok && len(aggr) > 0 {
		ddSort.SetAggregation(aggr)
	}
	if order, ok := sort["order"].(string); ok && len(order) > 0 {
		ddSort.SetOrder(datadogV1.WidgetSort(order))
	}
	if facet, ok := sort["facet"].(string); ok && len(facet) > 0 {
		ddSort.SetFacet(facet)
	}
	return ddSort
}

func buildTerraformQuery(datadogQueries *[]datadogV1.FormulaAndFunctionQueryDefinition) []map[string]interface{} {
	queries := make([]map[string]interface{}, len(*datadogQueries))
	for i, query := range *datadogQueries {
		terraformQuery := map[string]interface{}{}
		terraformEventQueryDefinition := query.FormulaAndFunctionEventQueryDefinition
		if terraformEventQueryDefinition != nil {
			if dataSource, ok := terraformEventQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if crossOrgUuids, ok := terraformEventQueryDefinition.GetCrossOrgUuidsOk(); ok {
				terraformQuery["cross_org_uuids"] = crossOrgUuids
			}
			if name, ok := terraformEventQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if indexes, ok := terraformEventQueryDefinition.GetIndexesOk(); ok {
				terraformQuery["indexes"] = indexes
			}
			if storage, ok := terraformEventQueryDefinition.GetStorageOk(); ok {
				terraformQuery["storage"] = storage
			}
			if search, ok := terraformEventQueryDefinition.GetSearchOk(); ok {
				if len(search.GetQuery()) > 0 {
					terraformSearch := map[string]interface{}{}
					terraformSearch["query"] = search.GetQuery()
					terraformSearchList := []map[string]interface{}{terraformSearch}
					terraformQuery["search"] = terraformSearchList
				}
			}
			if compute, ok := terraformEventQueryDefinition.GetComputeOk(); ok {
				terraformCompute := map[string]interface{}{}
				if aggregation, ok := compute.GetAggregationOk(); ok {
					terraformCompute["aggregation"] = aggregation
				}
				if interval, ok := compute.GetIntervalOk(); ok {
					terraformCompute["interval"] = interval
				}
				if metric, ok := compute.GetMetricOk(); ok {
					terraformCompute["metric"] = metric
				}
				terraformComputeList := []map[string]interface{}{terraformCompute}
				terraformQuery["compute"] = terraformComputeList
			}
			if terraformEventQuery, ok := terraformEventQueryDefinition.GetGroupByOk(); ok {
				terraformGroupBys := make([]map[string]interface{}, len(*terraformEventQuery))
				for i, groupBy := range *terraformEventQuery {
					// Facet
					terraformGroupBy := map[string]interface{}{
						"facet": groupBy.GetFacet(),
					}
					// Limit
					if v, ok := groupBy.GetLimitOk(); ok {
						terraformGroupBy["limit"] = *v
					}
					// Sort
					if v, ok := groupBy.GetSortOk(); ok {
						terraformSort := map[string]interface{}{}
						if metric, ok := v.GetMetricOk(); ok {
							terraformSort["metric"] = metric
						}
						if order, ok := v.GetOrderOk(); ok {
							terraformSort["order"] = order
						}
						if aggregation, ok := v.GetAggregationOk(); ok {
							terraformSort["aggregation"] = aggregation
						}
						terraformGroupBy["sort"] = []map[string]interface{}{terraformSort}
					}
					terraformGroupBys[i] = terraformGroupBy
				}
				terraformQuery["group_by"] = &terraformGroupBys
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformEventQuery := map[string]interface{}{}
			terraformEventQuery["event_query"] = terraformQueries
			queries[i] = terraformEventQuery
		}
		terraformMetricQueryDefinition := query.FormulaAndFunctionMetricQueryDefinition
		if terraformMetricQueryDefinition != nil {
			if dataSource, ok := terraformMetricQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if crossOrgUuids, ok := terraformMetricQueryDefinition.GetCrossOrgUuidsOk(); ok {
				terraformQuery["cross_org_uuids"] = crossOrgUuids
			}
			if metricQuery, ok := terraformMetricQueryDefinition.GetQueryOk(); ok {
				terraformQuery["query"] = metricQuery
			}
			if aggregator, ok := terraformMetricQueryDefinition.GetAggregatorOk(); ok {
				terraformQuery["aggregator"] = aggregator
			}
			if name, ok := terraformMetricQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if semanticMode, ok := terraformMetricQueryDefinition.GetSemanticModeOk(); ok {
				terraformQuery["semantic_mode"] = semanticMode
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformMetricQuery := map[string]interface{}{}
			terraformMetricQuery["metric_query"] = terraformQueries
			queries[i] = terraformMetricQuery
		}
		terraformApmDependencyStatsQueryDefinition := query.FormulaAndFunctionApmDependencyStatsQueryDefinition
		if terraformApmDependencyStatsQueryDefinition != nil {
			if dataSource, ok := terraformApmDependencyStatsQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if crossOrgUuids, ok := terraformEventQueryDefinition.GetCrossOrgUuidsOk(); ok {
				terraformQuery["cross_org_uuids"] = crossOrgUuids
			}
			if env, ok := terraformApmDependencyStatsQueryDefinition.GetEnvOk(); ok {
				terraformQuery["env"] = env
			}
			if stat, ok := terraformApmDependencyStatsQueryDefinition.GetStatOk(); ok {
				terraformQuery["stat"] = stat
			}
			if operationName, ok := terraformApmDependencyStatsQueryDefinition.GetOperationNameOk(); ok {
				terraformQuery["operation_name"] = operationName
			}
			if resourceName, ok := terraformApmDependencyStatsQueryDefinition.GetResourceNameOk(); ok {
				terraformQuery["resource_name"] = resourceName
			}
			if service, ok := terraformApmDependencyStatsQueryDefinition.GetServiceOk(); ok {
				terraformQuery["service"] = service
			}
			if primaryTagName, ok := terraformApmDependencyStatsQueryDefinition.GetPrimaryTagNameOk(); ok {
				terraformQuery["primary_tag_name"] = primaryTagName
			}
			if primaryTagValue, ok := terraformApmDependencyStatsQueryDefinition.GetPrimaryTagValueOk(); ok {
				terraformQuery["primary_tag_value"] = primaryTagValue
			}
			if isUpstream, ok := terraformApmDependencyStatsQueryDefinition.GetIsUpstreamOk(); ok {
				terraformQuery["is_upstream"] = isUpstream
			}
			if name, ok := terraformApmDependencyStatsQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformApmDependencyStatQuery := map[string]interface{}{}
			terraformApmDependencyStatQuery["apm_dependency_stats_query"] = terraformQueries
			queries[i] = terraformApmDependencyStatQuery
		}
		terraformApmResourceStatsQueryDefinition := query.FormulaAndFunctionApmResourceStatsQueryDefinition
		if terraformApmResourceStatsQueryDefinition != nil {
			if dataSource, ok := terraformApmResourceStatsQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if env, ok := terraformApmResourceStatsQueryDefinition.GetEnvOk(); ok {
				terraformQuery["env"] = env
			}
			if stat, ok := terraformApmResourceStatsQueryDefinition.GetStatOk(); ok {
				terraformQuery["stat"] = stat
			}
			if operationName, ok := terraformApmResourceStatsQueryDefinition.GetOperationNameOk(); ok {
				terraformQuery["operation_name"] = operationName
			}
			if resourceName, ok := terraformApmResourceStatsQueryDefinition.GetResourceNameOk(); ok {
				terraformQuery["resource_name"] = resourceName
			}
			if service, ok := terraformApmResourceStatsQueryDefinition.GetServiceOk(); ok {
				terraformQuery["service"] = service
			}
			if primaryTagName, ok := terraformApmResourceStatsQueryDefinition.GetPrimaryTagNameOk(); ok {
				terraformQuery["primary_tag_name"] = primaryTagName
			}
			if primaryTagValue, ok := terraformApmResourceStatsQueryDefinition.GetPrimaryTagValueOk(); ok {
				terraformQuery["primary_tag_value"] = primaryTagValue
			}
			if groupBy, ok := terraformApmResourceStatsQueryDefinition.GetGroupByOk(); ok {
				terraformQuery["group_by"] = groupBy
			}
			if name, ok := terraformApmResourceStatsQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformApmResourceStatQuery := map[string]interface{}{}
			terraformApmResourceStatQuery["apm_resource_stats_query"] = terraformQueries
			queries[i] = terraformApmResourceStatQuery
		}
		terraformProcessqueryDefinition := query.FormulaAndFunctionProcessQueryDefinition
		if terraformProcessqueryDefinition != nil {
			if dataSource, ok := terraformProcessqueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if crossOrgUuids, ok := terraformProcessqueryDefinition.GetCrossOrgUuidsOk(); ok {
				terraformQuery["cross_org_uuids"] = crossOrgUuids
			}
			if metric, ok := terraformProcessqueryDefinition.GetMetricOk(); ok {
				terraformQuery["metric"] = metric
			}
			if textFilter, ok := terraformProcessqueryDefinition.GetTextFilterOk(); ok {
				terraformQuery["text_filter"] = textFilter
			}
			if tagFilters, ok := terraformProcessqueryDefinition.GetTagFiltersOk(); ok {
				terraformQuery["tag_filters"] = tagFilters
			}
			if limit, ok := terraformProcessqueryDefinition.GetLimitOk(); ok {
				terraformQuery["limit"] = limit
			}
			if sort, ok := terraformProcessqueryDefinition.GetSortOk(); ok {
				terraformQuery["sort"] = sort
			}
			if isNormalizedCPU, ok := terraformProcessqueryDefinition.GetIsNormalizedCpuOk(); ok {
				terraformQuery["is_normalized_cpu"] = isNormalizedCPU
			}
			if aggregator, ok := terraformProcessqueryDefinition.GetAggregatorOk(); ok {
				terraformQuery["aggregator"] = aggregator
			}
			if name, ok := terraformProcessqueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformProcessQuery := map[string]interface{}{}
			terraformProcessQuery["process_query"] = terraformQueries
			queries[i] = terraformProcessQuery
		}
		terraformSLOQueryDefinition := query.FormulaAndFunctionSLOQueryDefinition
		if terraformSLOQueryDefinition != nil {
			if dataSource, ok := terraformSLOQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if crossOrgUuids, ok := terraformSLOQueryDefinition.GetCrossOrgUuidsOk(); ok {
				terraformQuery["cross_org_uuids"] = crossOrgUuids
			}
			if measure, ok := terraformSLOQueryDefinition.GetMeasureOk(); ok {
				terraformQuery["measure"] = measure
			}
			if sloID, ok := terraformSLOQueryDefinition.GetSloIdOk(); ok {
				terraformQuery["slo_id"] = sloID
			}
			if groupMode, ok := terraformSLOQueryDefinition.GetGroupModeOk(); ok {
				terraformQuery["group_mode"] = groupMode
			}
			if sloQueryType, ok := terraformSLOQueryDefinition.GetSloQueryTypeOk(); ok {
				terraformQuery["slo_query_type"] = sloQueryType
			}
			if name, ok := terraformSLOQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if additionalQueryFilters, ok := terraformSLOQueryDefinition.GetAdditionalQueryFiltersOk(); ok {
				terraformQuery["additional_query_filters"] = additionalQueryFilters
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformSLOQuery := map[string]interface{}{}
			terraformSLOQuery["slo_query"] = terraformQueries
			queries[i] = terraformSLOQuery
		}
		terraformCloudCostQueryDefinition := query.FormulaAndFunctionCloudCostQueryDefinition
		if terraformCloudCostQueryDefinition != nil {
			if dataSource, ok := terraformCloudCostQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if crossOrgUuids, ok := terraformCloudCostQueryDefinition.GetCrossOrgUuidsOk(); ok {
				terraformQuery["cross_org_uuids"] = crossOrgUuids
			}
			if aggregator, ok := terraformCloudCostQueryDefinition.GetAggregatorOk(); ok {
				terraformQuery["aggregator"] = aggregator
			}
			if name, ok := terraformCloudCostQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if query, ok := terraformCloudCostQueryDefinition.GetQueryOk(); ok {
				terraformQuery["query"] = query
			}
			terraformQueries := []map[string]interface{}{terraformQuery}
			terraformCloudCostQuery := map[string]interface{}{}
			terraformCloudCostQuery["cloud_cost_query"] = terraformQueries
			queries[i] = terraformCloudCostQuery
		}
	}
	return queries
}

func buildTerraformScatterplotFormula(datadogFormulas *[]datadogV1.ScatterplotWidgetFormula) []map[string]interface{} {
	formulas := make([]map[string]interface{}, len(*datadogFormulas))
	for i, formula := range *datadogFormulas {
		terraformFormula := map[string]interface{}{}
		terraformFormula["formula_expression"] = formula.GetFormula()
		terraformFormula["dimension"] = formula.GetDimension()
		if alias, ok := formula.GetAliasOk(); ok {
			terraformFormula["alias"] = alias
		}
		formulas[i] = terraformFormula
	}
	return formulas
}

func buildTerraformFormula(datadogFormulas *[]datadogV1.WidgetFormula, cellDisplayOptionFlag bool) []map[string]interface{} {
	formulas := make([]map[string]interface{}, len(*datadogFormulas))
	for i, formula := range *datadogFormulas {
		terraformFormula := map[string]interface{}{}
		terraformFormula["formula_expression"] = formula.GetFormula()
		if alias, ok := formula.GetAliasOk(); ok {
			terraformFormula["alias"] = alias
		}
		if limit, ok := formula.GetLimitOk(); ok {
			terraFormLimit := make(map[string]interface{})
			if count, ok := limit.GetCountOk(); ok {
				terraFormLimit["count"] = count
			}
			if order, ok := limit.GetOrderOk(); ok {
				terraFormLimit["order"] = string(*order)
			}
			terraformFormula["limit"] = []map[string]interface{}{terraFormLimit}
		}
		if v, ok := formula.GetConditionalFormatsOk(); ok {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(v)
			terraformFormula["conditional_formats"] = terraformConditionalFormats
		}
		if cellDisplayMode, cellDisplayModeOk := formula.GetCellDisplayModeOk(); cellDisplayModeOk {
			terraformFormula["cell_display_mode"] = cellDisplayMode
		}
		if cellDisplayOptionFlag {
			if cellDisplayOptions, cellDisplayOptionsOk := formula.GetCellDisplayModeOptionsOk(); cellDisplayOptionsOk {
				if cellDisplayOptions != nil {
					terraformFormula["cell_display_mode_options"] = []interface{}{
						map[string]interface{}{
							"trend_type": cellDisplayOptions.TrendType,
							"y_scale":    cellDisplayOptions.YScale,
						},
					}
				}
			}
		}
		if style, ok := formula.GetStyleOk(); ok {
			terraFormstyle := make(map[string]interface{})
			if palette, ok := style.GetPaletteOk(); ok {
				terraFormstyle["palette"] = palette
			}
			if palette_index, ok := style.GetPaletteIndexOk(); ok {
				terraFormstyle["palette_index"] = palette_index
			}
			terraformFormula["style"] = []map[string]interface{}{terraFormstyle}
		}
		if numberFormat, ok := formula.GetNumberFormatOk(); ok {
			terraformFormula["number_format"] = buildTerraformNumberFormatFormulaSchema(*numberFormat)
		}
		formulas[i] = terraformFormula
	}
	return formulas
}

func buildTerraformApmOrLogQueryCompute(compute *datadogV1.LogsQueryCompute) map[string]interface{} {
	terraformCompute := map[string]interface{}{
		"aggregation": compute.GetAggregation(),
	}
	if v, ok := compute.GetFacetOk(); ok {
		terraformCompute["facet"] = *v
	}
	if v, ok := compute.GetIntervalOk(); ok {
		terraformCompute["interval"] = *v
	}

	return terraformCompute
}

func buildTerraformApmOrLogQuery(datadogQuery datadogV1.LogQueryDefinition) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	// Index
	terraformQuery["index"] = datadogQuery.GetIndex()
	// Compute
	if compute, ok := datadogQuery.GetComputeOk(); ok {
		terraformQuery["compute_query"] = []map[string]interface{}{buildTerraformApmOrLogQueryCompute(compute)}
	}
	// Multi-compute
	if multiCompute, ok := datadogQuery.GetMultiComputeOk(); ok {
		terraformComputeList := make([]map[string]interface{}, len(*multiCompute))
		for i, compute := range *multiCompute {
			terraformCompute := map[string]interface{}{
				"aggregation": compute.GetAggregation(),
			}
			if v, ok := compute.GetFacetOk(); ok {
				terraformCompute["facet"] = *v
			}
			if compute.Interval != nil {
				terraformCompute["interval"] = *compute.Interval
			}
			terraformComputeList[i] = terraformCompute
		}
		terraformQuery["multi_compute"] = terraformComputeList
	}
	// Search
	if datadogQuery.Search != nil {
		terraformQuery["search_query"] = datadogQuery.Search.Query
	}
	// GroupBy
	if v, ok := datadogQuery.GetGroupByOk(); ok {
		terraformGroupBys := make([]map[string]interface{}, len(datadogQuery.GetGroupBy()))
		for i, groupBy := range *v {
			// Facet
			terraformGroupBy := map[string]interface{}{
				"facet": groupBy.GetFacet(),
			}
			// Limit
			if v, ok := groupBy.GetLimitOk(); ok {
				terraformGroupBy["limit"] = *v
			}
			// Sort
			if v, ok := groupBy.GetSortOk(); ok {
				sort := map[string]string{
					"aggregation": v.GetAggregation(),
					"order":       string(v.GetOrder()),
				}
				if groupBy.Sort.Facet != nil {
					sort["facet"] = *groupBy.Sort.Facet
				}
				terraformGroupBy["sort_query"] = []map[string]string{sort}
			}

			terraformGroupBys[i] = terraformGroupBy
		}
		terraformQuery["group_by"] = &terraformGroupBys
	}
	return terraformQuery
}

// Process Query
func buildDatadogProcessQuery(terraformQuery map[string]interface{}) *datadogV1.ProcessQueryDefinition {
	datadogQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
	if v, ok := terraformQuery["metric"].(string); ok && len(v) != 0 {
		datadogQuery.SetMetric(v)
	}
	if v, ok := terraformQuery["search_by"].(string); ok && len(v) != 0 {
		datadogQuery.SetSearchBy(v)
	}

	if terraformFilterBys, ok := terraformQuery["filter_by"].([]interface{}); ok && len(terraformFilterBys) > 0 {
		datadogFilterbys := make([]string, len(terraformFilterBys))
		for i, filterBy := range terraformFilterBys {
			datadogFilterbys[i] = filterBy.(string)
		}
		datadogQuery.SetFilterBy(datadogFilterbys)
	}

	if v, ok := terraformQuery["limit"].(int); ok && v != 0 {
		datadogQuery.SetLimit(int64(v))
	}

	return datadogQuery
}

func buildTerraformProcessQuery(datadogQuery datadogV1.ProcessQueryDefinition) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	if v, ok := datadogQuery.GetMetricOk(); ok {
		terraformQuery["metric"] = v
	}
	if v, ok := datadogQuery.GetSearchByOk(); ok {
		terraformQuery["search_by"] = v
	}
	if v, ok := datadogQuery.GetFilterByOk(); ok {
		terraformFilterBys := make([]string, len(*v))
		for i, datadogFilterBy := range *v {
			terraformFilterBys[i] = datadogFilterBy
		}
		terraformQuery["filter_by"] = terraformFilterBys
	}
	if v, ok := datadogQuery.GetLimitOk(); ok {
		terraformQuery["limit"] = v
	}

	return terraformQuery
}

// APM Resources Query
func buildDatadogApmStatsQueryColumn(terraformColumn map[string]interface{}) *datadogV1.ApmStatsQueryColumnType {

	datadogColumn := datadogV1.NewApmStatsQueryColumnTypeWithDefaults()

	if value, ok := terraformColumn["name"].(string); ok && len(value) != 0 {
		datadogColumn.SetName(value)
	}
	if value, ok := terraformColumn["alias"].(string); ok && len(value) != 0 {
		datadogColumn.SetAlias(value)
	}
	// avoid creating unnecessary diff with default value
	datadogColumn.CellDisplayMode = nil
	if value, ok := terraformColumn["cell_display_mode"].(string); ok && len(value) != 0 {
		datadogColumn.SetCellDisplayMode(datadogV1.TableWidgetCellDisplayMode(value))
	}
	if value, ok := terraformColumn["order"].(string); ok && len(value) != 0 {
		datadogColumn.SetOrder(datadogV1.WidgetSort(value))
	}

	return datadogColumn
}

func buildDatadogApmStatsQuery(terraformQuery map[string]interface{}) *datadogV1.ApmStatsQueryDefinition {
	datadogQuery := datadogV1.NewApmStatsQueryDefinitionWithDefaults()
	if v, ok := terraformQuery["service"].(string); ok && len(v) != 0 {
		datadogQuery.SetService(v)
	}
	if v, ok := terraformQuery["name"].(string); ok && len(v) != 0 {
		datadogQuery.SetName(v)
	}
	if v, ok := terraformQuery["env"].(string); ok && len(v) != 0 {
		datadogQuery.SetEnv(v)
	}
	if v, ok := terraformQuery["primary_tag"].(string); ok && len(v) != 0 {
		datadogQuery.SetPrimaryTag(v)
	}
	if v, ok := terraformQuery["row_type"].(string); ok && len(v) != 0 {
		datadogQuery.SetRowType(datadogV1.ApmStatsQueryRowType(v))
	}
	if v, ok := terraformQuery["resource"].(string); ok && len(v) != 0 {
		datadogQuery.SetResource(v)
	}

	if terraformColumns, ok := terraformQuery["columns"].([]interface{}); ok && len(terraformColumns) > 0 {
		datadogColumns := make([]datadogV1.ApmStatsQueryColumnType, len(terraformColumns))
		for i, column := range terraformColumns {
			datadogColumns[i] = *buildDatadogApmStatsQueryColumn(column.(map[string]interface{}))
		}
		datadogQuery.SetColumns(datadogColumns)
	}

	return datadogQuery
}

func buildTerraformApmStatsQuery(datadogQuery datadogV1.ApmStatsQueryDefinition) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	if v, ok := datadogQuery.GetServiceOk(); ok {
		terraformQuery["service"] = v
	}
	if v, ok := datadogQuery.GetNameOk(); ok {
		terraformQuery["name"] = v
	}
	if v, ok := datadogQuery.GetEnvOk(); ok {
		terraformQuery["env"] = v
	}
	if v, ok := datadogQuery.GetPrimaryTagOk(); ok {
		terraformQuery["primary_tag"] = v
	}
	if v, ok := datadogQuery.GetRowTypeOk(); ok {
		terraformQuery["row_type"] = v
	}
	if v, ok := datadogQuery.GetResourceOk(); ok {
		terraformQuery["resource"] = v
	}
	if v, ok := datadogQuery.GetColumnsOk(); ok {
		terraformColumns := make([]interface{}, len(*v))
		for i, datadogColumn := range *v {
			terraformColumn := map[string]interface{}{}
			if name, nameOk := datadogColumn.GetNameOk(); nameOk {
				terraformColumn["name"] = name
			}
			if alias, aliasOk := datadogColumn.GetAliasOk(); aliasOk {
				terraformColumn["alias"] = alias
			}
			if cellDisplayMode, cellDisplayModeOk := datadogColumn.GetCellDisplayModeOk(); cellDisplayModeOk {
				terraformColumn["cell_display_mode"] = cellDisplayMode
			}
			if order, orderOk := datadogColumn.GetOrderOk(); orderOk {
				terraformColumn["order"] = order
			}
			terraformColumns[i] = terraformColumn
		}
		terraformQuery["columns"] = terraformColumns
	}

	return terraformQuery
}

// Widget Axis helpers

func buildDatadogWidgetAxis(terraformWidgetAxis map[string]interface{}) *datadogV1.WidgetAxis {
	datadogWidgetAxis := &datadogV1.WidgetAxis{}
	if v, ok := terraformWidgetAxis["label"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetLabel(v)
	}
	if v, ok := terraformWidgetAxis["scale"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetScale(v)
	}
	if v, ok := terraformWidgetAxis["min"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetMin(v)
	}
	if v, ok := terraformWidgetAxis["max"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetMax(v)
	}
	if v, ok := terraformWidgetAxis["include_zero"].(bool); ok {
		datadogWidgetAxis.SetIncludeZero(v)
	}
	return datadogWidgetAxis
}

func buildTerraformWidgetAxis(datadogWidgetAxis datadogV1.WidgetAxis) map[string]interface{} {
	terraformWidgetAxis := map[string]interface{}{}
	if v, ok := datadogWidgetAxis.GetLabelOk(); ok {
		terraformWidgetAxis["label"] = v
	}
	if v, ok := datadogWidgetAxis.GetScaleOk(); ok {
		terraformWidgetAxis["scale"] = v
	}
	if v, ok := datadogWidgetAxis.GetMinOk(); ok {
		terraformWidgetAxis["min"] = v
	}
	if v, ok := datadogWidgetAxis.GetMaxOk(); ok {
		terraformWidgetAxis["max"] = v
	}
	if v, ok := datadogWidgetAxis.GetIncludeZeroOk(); ok {
		terraformWidgetAxis["include_zero"] = v
	}
	return terraformWidgetAxis
}

// Distribution Widget XAxis helpers

func buildDatadogDistributionWidgetXAxis(terraformDistributionWidgetXAxis map[string]interface{}) *datadogV1.DistributionWidgetXAxis {
	datadogDistributionWidgetXAxis := &datadogV1.DistributionWidgetXAxis{}
	if v, ok := terraformDistributionWidgetXAxis["scale"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetXAxis.SetScale(v)
	}
	if v, ok := terraformDistributionWidgetXAxis["min"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetXAxis.SetMin(v)
	}
	if v, ok := terraformDistributionWidgetXAxis["max"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetXAxis.SetMax(v)
	}
	if v, ok := terraformDistributionWidgetXAxis["include_zero"].(bool); ok {
		datadogDistributionWidgetXAxis.SetIncludeZero(v)
	}
	return datadogDistributionWidgetXAxis
}

func buildTerraformDistributionWidgetXAxis(datadogDistributionWidgetXAxis datadogV1.DistributionWidgetXAxis) map[string]interface{} {
	terraformDistributionWidgetXAxis := map[string]interface{}{}
	if v, ok := datadogDistributionWidgetXAxis.GetScaleOk(); ok {
		terraformDistributionWidgetXAxis["scale"] = v
	}
	if v, ok := datadogDistributionWidgetXAxis.GetMinOk(); ok {
		terraformDistributionWidgetXAxis["min"] = v
	}
	if v, ok := datadogDistributionWidgetXAxis.GetMaxOk(); ok {
		terraformDistributionWidgetXAxis["max"] = v
	}
	if v, ok := datadogDistributionWidgetXAxis.GetIncludeZeroOk(); ok {
		terraformDistributionWidgetXAxis["include_zero"] = v
	}
	return terraformDistributionWidgetXAxis
}

// Distribution Widget YAxis helpers

func buildDatadogDistributionWidgetYAxis(terraformDistributionWidgetYAxis map[string]interface{}) *datadogV1.DistributionWidgetYAxis {
	datadogDistributionWidgetYAxis := &datadogV1.DistributionWidgetYAxis{}
	if v, ok := terraformDistributionWidgetYAxis["scale"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetYAxis.SetScale(v)
	}
	if v, ok := terraformDistributionWidgetYAxis["min"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetYAxis.SetMin(v)
	}
	if v, ok := terraformDistributionWidgetYAxis["max"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetYAxis.SetMax(v)
	}
	if v, ok := terraformDistributionWidgetYAxis["include_zero"].(bool); ok {
		datadogDistributionWidgetYAxis.SetIncludeZero(v)
	}
	if v, ok := terraformDistributionWidgetYAxis["label"].(string); ok && len(v) != 0 {
		datadogDistributionWidgetYAxis.SetLabel(v)
	}
	return datadogDistributionWidgetYAxis
}

func buildTerraformDistributionWidgetYAxis(datadogDistributionWidgetYAxis datadogV1.DistributionWidgetYAxis) map[string]interface{} {
	terraformDistributionWidgetYAxis := map[string]interface{}{}
	if v, ok := datadogDistributionWidgetYAxis.GetScaleOk(); ok {
		terraformDistributionWidgetYAxis["scale"] = v
	}
	if v, ok := datadogDistributionWidgetYAxis.GetMinOk(); ok {
		terraformDistributionWidgetYAxis["min"] = v
	}
	if v, ok := datadogDistributionWidgetYAxis.GetMaxOk(); ok {
		terraformDistributionWidgetYAxis["max"] = v
	}
	if v, ok := datadogDistributionWidgetYAxis.GetIncludeZeroOk(); ok {
		terraformDistributionWidgetYAxis["include_zero"] = v
	}
	if v, ok := datadogDistributionWidgetYAxis.GetLabelOk(); ok {
		terraformDistributionWidgetYAxis["label"] = v
	}
	return terraformDistributionWidgetYAxis
}

// Widget Style helpers
func buildDatadogWidgetStyle(terraformStyle map[string]interface{}) *datadogV1.WidgetStyle {
	datadogStyle := &datadogV1.WidgetStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}

	return datadogStyle
}
func buildTerraformWidgetStyle(datadogStyle datadogV1.WidgetStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if v, ok := datadogStyle.GetPaletteOk(); ok {
		terraformStyle["palette"] = v
	}
	return terraformStyle
}

// Timeseriest Style helpers

func buildDatadogWidgetRequestStyle(terraformStyle map[string]interface{}) *datadogV1.WidgetRequestStyle {
	datadogStyle := &datadogV1.WidgetRequestStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}
	if v, ok := terraformStyle["line_type"].(string); ok && len(v) != 0 {
		datadogStyle.SetLineType(datadogV1.WidgetLineType(v))
	}
	if v, ok := terraformStyle["line_width"].(string); ok && len(v) != 0 {
		datadogStyle.SetLineWidth(datadogV1.WidgetLineWidth(v))
	}
	if v, ok := terraformStyle["order_by"].(string); ok && len(v) != 0 {
		datadogStyle.SetOrderBy(datadogV1.WidgetStyleOrderBy(v))
	}

	return datadogStyle
}
func buildTerraformWidgetRequestStyle(datadogStyle datadogV1.WidgetRequestStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if v, ok := datadogStyle.GetPaletteOk(); ok {
		terraformStyle["palette"] = v
	}
	if v, ok := datadogStyle.GetLineTypeOk(); ok {
		terraformStyle["line_type"] = v
	}
	if v, ok := datadogStyle.GetLineWidthOk(); ok {
		terraformStyle["line_width"] = v
	}
	if v, ok := datadogStyle.GetOrderByOk(); ok {
		terraformStyle["order_by"] = v
	}
	return terraformStyle
}

func buildDatadogGeomapRequestStyle(terraformStyle map[string]interface{}) *datadogV1.GeomapWidgetDefinitionStyle {
	datadogStyle := &datadogV1.GeomapWidgetDefinitionStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}
	if v, ok := terraformStyle["palette_flip"].(bool); ok {
		datadogStyle.SetPaletteFlip(v)
	}

	return datadogStyle
}

func buildTerraformGeomapRequestStyle(datadogStyle datadogV1.GeomapWidgetDefinitionStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if v, ok := datadogStyle.GetPaletteOk(); ok {
		terraformStyle["palette"] = v
	}
	if v, ok := datadogStyle.GetPaletteFlipOk(); ok {
		terraformStyle["palette_flip"] = v
	}
	return terraformStyle
}

func buildDatadogGeomapRequestView(terraformStyle map[string]interface{}) *datadogV1.GeomapWidgetDefinitionView {
	datadogView := &datadogV1.GeomapWidgetDefinitionView{}
	if v, ok := terraformStyle["focus"].(string); ok && len(v) != 0 {
		datadogView.SetFocus(v)
	}

	return datadogView
}

func buildTerraformGeomapRequestView(datadogView datadogV1.GeomapWidgetDefinitionView) map[string]interface{} {
	terraformView := map[string]interface{}{}
	if v, ok := datadogView.GetFocusOk(); ok {
		terraformView["focus"] = v
	}

	return terraformView
}

// Hostmap Style helpers

// Schema validation
func validateTimeseriesWidgetLegendSize(val interface{}, key string) (warns []string, errs []error) {
	value := val.(string)
	switch value {
	case "0", "2", "4", "8", "16", "auto":
		break
	default:
		errs = append(errs, fmt.Errorf(
			"%q contains an invalid value %q. Valid values are `0`, `2`, `4`, `8`, `16`, or `auto`", key, value))
	}
	return
}

// Number format Formula
func buildDatadogNumberFormatFormulaSchema(terraformStyle map[string]interface{}) *datadogV1.WidgetNumberFormat {
	if terraformStyle == nil || len(terraformStyle) == 0 {
		return nil
	}
	var datadogNumber datadogV1.WidgetNumberFormat
	if v, ok := terraformStyle["unit"].([]interface{}); ok && len(v) > 0 {
		unit := v[0].(map[string]interface{})
		datadogNumber.Unit = &datadogV1.NumberFormatUnit{}
		if v, ok := unit["canonical"].([]interface{}); ok && len(v) > 0 {
			canonical := v[0].(map[string]interface{})
			datadogNumber.Unit.NumberFormatUnitCanonical = &datadogV1.NumberFormatUnitCanonical{
				Type: Ptr(datadogV1.NUMBERFORMATUNITSCALETYPE_CANONICAL_UNIT),
			}
			if v, ok := canonical["per_unit_name"].(string); ok && len(v) > 0 {
				datadogNumber.Unit.NumberFormatUnitCanonical.PerUnitName = Ptr(v)
			}
			if v, ok := canonical["unit_name"].(string); ok && len(v) > 0 {
				datadogNumber.Unit.NumberFormatUnitCanonical.UnitName = Ptr(v)
			}
		}
		if v, ok := unit["custom"].([]interface{}); ok && len(v) > 0 {
			custom := v[0].(map[string]interface{})
			datadogNumber.Unit.NumberFormatUnitCustom = &datadogV1.NumberFormatUnitCustom{
				Type: Ptr(datadogV1.NUMBERFORMATUNITCUSTOMTYPE_CUSTOM_UNIT_LABEL),
			}
			if v, ok := custom["label"].(string); ok && len(v) > 0 {
				datadogNumber.Unit.NumberFormatUnitCustom.Label = Ptr(v)
			}
		}
	}
	if v, ok := terraformStyle["unit_scale"].([]interface{}); ok && len(v) > 0 {
		unitScale := v[0].(map[string]interface{})
		if unitName, ok := unitScale["unit_name"].(string); ok && len(unitName) > 0 {
			datadogNumber.UnitScale = *datadogV1.NewNullableNumberFormatUnitScale(&datadogV1.NumberFormatUnitScale{
				Type:     datadogV1.NUMBERFORMATUNITSCALETYPE_CANONICAL_UNIT.Ptr(),
				UnitName: datadog.PtrString(unitName),
			})
		}
	}
	return &datadogNumber
}

func buildTerraformNumberFormatFormulaSchema(datadogStyle datadogV1.WidgetNumberFormat) []map[string]interface{} {
	m := map[string]interface{}{}
	if v, ok := datadogStyle.GetUnitOk(); ok {
		unit := map[string]interface{}{}
		if v.NumberFormatUnitCustom != nil {
			unit["custom"] = []map[string]interface{}{{"label": v.NumberFormatUnitCustom.Label}}
		}
		if v.NumberFormatUnitCanonical != nil {
			unit["canonical"] = []map[string]interface{}{
				{"per_unit_name": v.NumberFormatUnitCanonical.PerUnitName,
					"unit_name": v.NumberFormatUnitCanonical.UnitName},
			}
		}
		m["unit"] = []map[string]interface{}{unit}
	}
	if v, ok := datadogStyle.GetUnitScaleOk(); ok {
		unitScale := map[string]interface{}{"unit_name": v.UnitName}
		m["unit_scale"] = []map[string]interface{}{unitScale}
	}
	return []map[string]interface{}{m}
}
