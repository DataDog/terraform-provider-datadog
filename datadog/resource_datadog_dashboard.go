package datadog

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
		Description:   "Provides a Datadog dashboard resource. This can be used to create and manage Datadog dashboards.",
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
					Deprecated:    "Prefer using `restricted_roles` to define which roles are required to edit the dashboard.",
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

func resourceDatadogDashboardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	dashboardPayload, err := buildDatadogDashboard(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dashboard, httpresp, err := apiInstances.GetDashboardsApiV1().CreateDashboard(auth, *dashboardPayload)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard")
	}
	if err := utils.CheckForUnparsed(dashboard); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*dashboard.Id)

	var getDashboard datadogV1.Dashboard
	var httpResponse *http.Response
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		getDashboard, httpResponse, err = apiInstances.GetDashboardsApiV1().GetDashboard(auth, *dashboard.Id)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("dashboard not created yet"))
			}

			return retry.NonRetryableError(err)
		}
		if err := utils.CheckForUnparsed(getDashboard); err != nil {
			return retry.NonRetryableError(err)
		}

		// We only log the error, as failing to update the list shouldn't fail dashboard creation
		updateDashboardLists(d, providerConf, *dashboard.Id, d.Get("layout_type").(string))

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return updateDashboardState(d, &getDashboard)
}

func resourceDatadogDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	dashboard, err := buildDatadogDashboard(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	updatedDashboard, httpresp, err := apiInstances.GetDashboardsApiV1().UpdateDashboard(auth, id, *dashboard)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard")
	}
	if err := utils.CheckForUnparsed(updatedDashboard); err != nil {
		return diag.FromErr(err)
	}

	updateDashboardLists(d, providerConf, *dashboard.Id, d.Get("layout_type").(string))

	return updateDashboardState(d, &updatedDashboard)
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

func updateDashboardState(d *schema.ResourceData, dashboard *datadogV1.Dashboard) diag.Diagnostics {
	if err := d.Set("title", dashboard.GetTitle()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("layout_type", dashboard.GetLayoutType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("reflow_type", dashboard.GetReflowType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", dashboard.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", dashboard.GetUrl()); err != nil {
		return diag.FromErr(err)
	}

	// Set RBAC role settings
	if err := d.Set("is_read_only", dashboard.GetIsReadOnly()); err != nil {
		return diag.FromErr(err)
	}
	restrictedRoles := buildTerraformRestrictedRoles(&dashboard.RestrictedRoles)
	if err := d.Set("restricted_roles", restrictedRoles); err != nil {
		return diag.FromErr(err)
	}

	// Set widgets
	terraformWidgets, err := buildTerraformWidgets(&dashboard.Widgets, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("widget", terraformWidgets); err != nil {
		return diag.FromErr(err)
	}

	// Set template variables
	templateVariables := buildTerraformTemplateVariables(&dashboard.TemplateVariables)
	if err := d.Set("template_variable", templateVariables); err != nil {
		return diag.FromErr(err)
	}

	// Set template variable presets
	templateVariablePresets := buildTerraformTemplateVariablePresets(&dashboard.TemplateVariablePresets)
	if err := d.Set("template_variable_preset", templateVariablePresets); err != nil {
		return diag.FromErr(err)
	}

	// Set notify list
	notifyList := buildTerraformNotifyList(dashboard.NotifyList.Get())
	if err := d.Set("notify_list", notifyList); err != nil {
		return diag.FromErr(err)
	}

	// Set tags
	tags := dashboard.GetTags()
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	dashboard, httpresp, err := apiInstances.GetDashboardsApiV1().GetDashboard(auth, id)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard")
	}
	if err := utils.CheckForUnparsed(dashboard); err != nil {
		return diag.FromErr(err)
	}

	return updateDashboardState(d, &dashboard)
}

func resourceDatadogDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()
	if _, httpresp, err := apiInstances.GetDashboardsApiV1().DeleteDashboard(auth, id); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard")
	}
	return nil
}

func buildDatadogDashboard(d *schema.ResourceData) (*datadogV1.Dashboard, error) {
	var dashboard datadogV1.Dashboard

	dashboard.SetId(d.Id())

	if v, ok := d.GetOk("title"); ok {
		dashboard.SetTitle(v.(string))
	}
	if v, ok := d.GetOk("layout_type"); ok {
		dashboard.SetLayoutType(datadogV1.DashboardLayoutType(v.(string)))
	}
	if v, ok := d.GetOk("reflow_type"); ok {
		dashboard.SetReflowType(datadogV1.DashboardReflowType(v.(string)))
	}
	if v, ok := d.GetOk("description"); ok {
		dashboard.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("is_read_only"); ok {
		dashboard.SetIsReadOnly(v.(bool))
	}
	if v, ok := d.GetOk("restricted_roles"); ok && !dashboard.GetIsReadOnly() {
		// do not set when 'is_read_only = true' because this takes priority on requests
		dashboard.RestrictedRoles = *buildDatadogRestrictedRoles(v.(*schema.Set))
	}

	// Build Widgets
	terraformWidgets := d.Get("widget").([]interface{})
	datadogWidgets, err := buildDatadogWidgets(&terraformWidgets)
	if err != nil {
		return nil, err
	}
	dashboard.SetWidgets(*datadogWidgets)

	// Build NotifyList
	notifyList := d.Get("notify_list").(*schema.Set)
	dashboard.SetNotifyList(*buildDatadogNotifyList(notifyList))

	// Build Tags
	tags := utils.GetStringSlice(d, "tags")
	dashboard.SetTags(tags)

	// Build TemplateVariables
	templateVariables := d.Get("template_variable").([]interface{})
	dashboard.TemplateVariables = *buildDatadogTemplateVariables(&templateVariables)

	// Build TemplateVariablePresets
	templateVariablePresets := d.Get("template_variable_preset").([]interface{})
	dashboard.TemplateVariablePresets = *buildDatadogTemplateVariablePresets(&templateVariablePresets)

	return &dashboard, nil
}

//
// Template Variable helpers
//

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

func buildDatadogTemplateVariables(terraformTemplateVariables *[]interface{}) *[]datadogV1.DashboardTemplateVariable {
	datadogTemplateVariables := make([]datadogV1.DashboardTemplateVariable, len(*terraformTemplateVariables))
	for i, ttv := range *terraformTemplateVariables {
		if ttv == nil {
			continue
		}
		terraformTemplateVariable := ttv.(map[string]interface{})
		var datadogTemplateVariable datadogV1.DashboardTemplateVariable
		if v, ok := terraformTemplateVariable["name"].(string); ok && len(v) != 0 {
			datadogTemplateVariable.SetName(v)
		}
		if v, ok := terraformTemplateVariable["prefix"].(string); ok && len(v) != 0 {
			datadogTemplateVariable.SetPrefix(v)
		}
		if v, ok := terraformTemplateVariable["defaults"].([]interface{}); ok && len(v) != 0 {
			var defaults []string
			for _, s := range v {
				defaults = append(defaults, s.(string))
			}
			datadogTemplateVariable.SetDefaults(defaults)
		} else if v, ok := terraformTemplateVariable["default"].(string); ok && len(v) != 0 {
			datadogTemplateVariable.SetDefault(v)
		}
		if v, ok := terraformTemplateVariable["available_values"].([]interface{}); ok && len(v) > 0 {
			availableValues := make([]string, len(v))
			for i, availableValue := range v {
				availableValues[i] = availableValue.(string)
			}
			datadogTemplateVariable.SetAvailableValues(availableValues)
		}
		datadogTemplateVariables[i] = datadogTemplateVariable
	}
	return &datadogTemplateVariables
}

func buildTerraformTemplateVariables(datadogTemplateVariables *[]datadogV1.DashboardTemplateVariable) *[]map[string]interface{} {
	terraformTemplateVariables := make([]map[string]interface{}, len(*datadogTemplateVariables))
	for i, templateVariable := range *datadogTemplateVariables {
		terraformTemplateVariable := map[string]interface{}{}
		if v, ok := templateVariable.GetNameOk(); ok {
			terraformTemplateVariable["name"] = *v
		}
		if v := templateVariable.GetPrefix(); len(v) > 0 {
			terraformTemplateVariable["prefix"] = v
		}
		if v, ok := templateVariable.GetDefaultsOk(); ok && len(*v) > 0 {
			var tags []string
			tags = append(tags, *v...)
			terraformTemplateVariable["defaults"] = tags
		} else if v, ok := templateVariable.GetDefaultOk(); ok {
			terraformTemplateVariable["default"] = *v
		}
		if v, ok := templateVariable.GetAvailableValuesOk(); ok {
			availableValues := make([]string, len(*v))
			for i, availableValue := range *v {
				availableValues[i] = availableValue
			}
			terraformTemplateVariable["available_values"] = availableValues
		}
		terraformTemplateVariables[i] = terraformTemplateVariable
	}
	return &terraformTemplateVariables
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

func buildDatadogTemplateVariablePresets(terraformTemplateVariablePresets *[]interface{}) *[]datadogV1.DashboardTemplateVariablePreset {
	datadogTemplateVariablePresets := make([]datadogV1.DashboardTemplateVariablePreset, len(*terraformTemplateVariablePresets))

	for i, tvp := range *terraformTemplateVariablePresets {
		if tvp == nil {
			continue
		}
		templateVariablePreset := tvp.(map[string]interface{})
		var datadogTemplateVariablePreset datadogV1.DashboardTemplateVariablePreset

		if v, ok := templateVariablePreset["name"].(string); ok && len(v) != 0 {
			datadogTemplateVariablePreset.SetName(v)
		}

		if templateVariablePresetValues, ok := templateVariablePreset["template_variable"].([]interface{}); ok {
			datadogTemplateVariablePresetValues := make([]datadogV1.DashboardTemplateVariablePresetValue, len(templateVariablePresetValues))

			for j, tvp := range templateVariablePresetValues {
				if tvp == nil {
					continue
				}
				templateVariablePresetValue := tvp.(map[string]interface{})
				var datadogTemplateVariablePresetValue datadogV1.DashboardTemplateVariablePresetValue

				if w, ok := templateVariablePresetValue["name"].(string); ok && len(w) != 0 {
					datadogTemplateVariablePresetValue.SetName(w)
				}

				if w, ok := templateVariablePresetValue["values"].([]interface{}); ok && len(w) != 0 {
					var presets []string
					for _, s := range w {
						presets = append(presets, s.(string))
					}
					datadogTemplateVariablePresetValue.SetValues(presets)
				} else if w, ok := templateVariablePresetValue["value"].(string); ok && len(w) != 0 {
					datadogTemplateVariablePresetValue.SetValue(w)
				}

				datadogTemplateVariablePresetValues[j] = datadogTemplateVariablePresetValue
			}

			datadogTemplateVariablePreset.SetTemplateVariables(datadogTemplateVariablePresetValues)
		}

		datadogTemplateVariablePresets[i] = datadogTemplateVariablePreset
	}

	return &datadogTemplateVariablePresets
}

func buildTerraformTemplateVariablePresets(datadogTemplateVariablePresets *[]datadogV1.DashboardTemplateVariablePreset) *[]map[string]interface{} {
	// Allocate final resting place for tf/hash version
	terraformTemplateVariablePresets := make([]map[string]interface{}, len(*datadogTemplateVariablePresets))

	//iterate over preset objects
	for i, templateVariablePreset := range *datadogTemplateVariablePresets {
		// Allocate for this preset group, a map of string key to obj (string for name, array for preset values
		terraformTemplateVariablePreset := make(map[string]interface{})
		if v, ok := templateVariablePreset.GetNameOk(); ok {
			terraformTemplateVariablePreset["name"] = v
		}

		// allocate for array of preset values (names = name,value, values = name, template variable)

		terraformTemplateVariablePresetValues := make([]map[string]interface{}, len(templateVariablePreset.GetTemplateVariables()))
		for j, templateVariablePresetValue := range templateVariablePreset.GetTemplateVariables() {
			// allocate map for name => name value => value
			terraformTemplateVariablePresetValue := make(map[string]interface{})
			if v, ok := templateVariablePresetValue.GetNameOk(); ok {
				terraformTemplateVariablePresetValue["name"] = *v
			}
			if v, ok := templateVariablePresetValue.GetValuesOk(); ok && len(*v) > 0 {
				var tags []string
				tags = append(tags, *v...)
				terraformTemplateVariablePresetValue["values"] = tags
			} else if v, ok := templateVariablePresetValue.GetValueOk(); ok {
				terraformTemplateVariablePresetValue["value"] = *v
			}

			terraformTemplateVariablePresetValues[j] = terraformTemplateVariablePresetValue
		}

		// Set template_variable to the array of values we just created
		terraformTemplateVariablePreset["template_variable"] = terraformTemplateVariablePresetValues

		// put the preset group into the output var
		terraformTemplateVariablePresets[i] = terraformTemplateVariablePreset
	}

	return &terraformTemplateVariablePresets
}

//
// Restricted Roles helpers
//

func buildDatadogRestrictedRoles(terraformRestrictedRoles *schema.Set) *[]string {
	roles := make([]string, 0)
	for _, r := range terraformRestrictedRoles.List() {
		roles = append(roles, r.(string))
	}
	return &roles
}

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
// Notify List helpers
//

func buildDatadogNotifyList(terraformNotifyList *schema.Set) *[]string {
	datadogNotifyList := make([]string, len(terraformNotifyList.List()))
	for i, authorHandle := range terraformNotifyList.List() {
		datadogNotifyList[i] = authorHandle.(string)
	}
	return &datadogNotifyList
}

func buildTerraformNotifyList(datadogNotifyList *[]string) *[]string {
	if datadogNotifyList == nil {
		terraformNotifyList := make([]string, 0)
		return &terraformNotifyList
	}
	terraformNotifyList := make([]string, len(*datadogNotifyList))
	for i, authorHandle := range *datadogNotifyList {
		terraformNotifyList[i] = authorHandle
	}
	return &terraformNotifyList
}

//
// Widget helpers
//

// The generic widget schema is a combination of the schema for a non-group widget
// and the schema for a Group Widget (which can contains only non-group widgets)
func getWidgetSchema() map[string]*schema.Schema {
	widgetSchema := getNonGroupWidgetSchema()
	widgetSchema["group_definition"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The definition for a Group widget.",
		Elem: &schema.Resource{
			Schema: getGroupDefinitionSchema(),
		},
	}
	return widgetSchema
}

func getNonGroupWidgetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"alert_graph_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Alert Graph widget.",
			Elem: &schema.Resource{
				Schema: getAlertGraphDefinitionSchema(),
			},
		},
		"alert_value_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Alert Value widget.",
			Elem: &schema.Resource{
				Schema: getAlertValueDefinitionSchema(),
			},
		},
		"change_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Change widget.",
			Elem: &schema.Resource{
				Schema: getChangeDefinitionSchema(),
			},
		},
		"check_status_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Check Status widget.",
			Elem: &schema.Resource{
				Schema: getCheckStatusDefinitionSchema(),
			},
		},
		"distribution_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Distribution widget.",
			Elem: &schema.Resource{
				Schema: getDistributionDefinitionSchema(),
			},
		},
		"event_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Event Stream widget.",
			Elem: &schema.Resource{
				Schema: getEventStreamDefinitionSchema(),
			},
		},
		"event_timeline_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Event Timeline widget.",
			Elem: &schema.Resource{
				Schema: getEventTimelineDefinitionSchema(),
			},
		},
		"free_text_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Free Text widget.",
			Elem: &schema.Resource{
				Schema: getFreeTextDefinitionSchema(),
			},
		},
		"heatmap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Heatmap widget.",
			Elem: &schema.Resource{
				Schema: getHeatmapDefinitionSchema(),
			},
		},
		"hostmap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Hostmap widget.",
			Elem: &schema.Resource{
				Schema: getHostmapDefinitionSchema(),
			},
		},
		"iframe_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Iframe widget.",
			Elem: &schema.Resource{
				Schema: getIframeDefinitionSchema(),
			},
		},
		"image_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Image widget",
			Elem: &schema.Resource{
				Schema: getImageDefinitionSchema(),
			},
		},
		"list_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a List Stream widget.",
			Elem: &schema.Resource{
				Schema: getListStreamDefinitionSchema(),
			},
		},
		"log_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Log Stream widget.",
			Elem: &schema.Resource{
				Schema: getLogStreamDefinitionSchema(),
			},
		},
		"manage_status_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Manage Status widget.",
			Elem: &schema.Resource{
				Schema: getManageStatusDefinitionSchema(),
			},
		},
		"note_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Note widget.",
			Elem: &schema.Resource{
				Schema: getNoteDefinitionSchema(),
			},
		},
		"query_value_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Query Value widget.",
			Elem: &schema.Resource{
				Schema: getQueryValueDefinitionSchema(),
			},
		},
		"query_table_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Query Table widget.",
			Elem: &schema.Resource{
				Schema: getQueryTableDefinitionSchema(),
			},
		},
		"scatterplot_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Scatterplot widget.",
			Elem: &schema.Resource{
				Schema: getScatterplotDefinitionSchema(),
			},
		},
		"servicemap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Service Map widget.",
			Elem: &schema.Resource{
				Schema: getServiceMapDefinitionSchema(),
			},
		},
		"service_level_objective_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Service Level Objective widget.",
			Elem: &schema.Resource{
				Schema: getServiceLevelObjectiveDefinitionSchema(),
			},
		},
		"slo_list_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an SLO (Service Level Objective) List widget.",
			Elem: &schema.Resource{
				Schema: getSloListDefinitionSchema(),
			},
		},
		"sunburst_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Sunburst widget.",
			Elem: &schema.Resource{
				Schema: getSunburstDefinitionschema(),
			},
		},
		"timeseries_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Timeseries widget.",
			Elem: &schema.Resource{
				Schema: getTimeseriesDefinitionSchema(),
			},
		},
		"toplist_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Toplist widget.",
			Elem: &schema.Resource{
				Schema: getToplistDefinitionSchema(),
			},
		},
		"topology_map_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Topology Map widget.",
			Elem: &schema.Resource{
				Schema: getTopologyMapDefinitionSchema(),
			},
		},
		"trace_service_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Trace Service widget.",
			Elem: &schema.Resource{
				Schema: getTraceServiceDefinitionSchema(),
			},
		},
		"treemap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Treemap widget.",
			Elem: &schema.Resource{
				Schema: getTreemapDefinitionSchema(),
			},
		},
		"geomap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Geomap widget.",
			Elem: &schema.Resource{
				Schema: getGeomapDefinitionSchema(),
			},
		},
		"run_workflow_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Run Workflow widget.",
			Elem: &schema.Resource{
				Schema: getRunWorkflowDefinitionSchema(),
			},
		},
	}
}

func buildDatadogWidgets(terraformWidgets *[]interface{}) (*[]datadogV1.Widget, error) {
	datadogWidgets := make([]datadogV1.Widget, len(*terraformWidgets))
	for i, terraformWidget := range *terraformWidgets {
		if terraformWidget == nil {
			continue
		}
		datadogWidget, err := buildDatadogWidget(terraformWidget.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		datadogWidgets[i] = *datadogWidget
	}
	return &datadogWidgets, nil
}

// Helper to build a Datadog widget from a Terraform widget
func buildDatadogWidget(terraformWidget map[string]interface{}) (*datadogV1.Widget, error) {
	// Build widget Definition
	var definition datadogV1.WidgetDefinition
	if def, ok := terraformWidget["group_definition"].([]interface{}); ok && len(def) > 0 {
		if groupDefinition, ok := def[0].(map[string]interface{}); ok {
			datadogDefinition, err := buildDatadogGroupDefinition(groupDefinition)
			if err != nil {
				return nil, err
			}
			definition = datadogV1.GroupWidgetDefinitionAsWidgetDefinition(datadogDefinition)
		}
	} else if def, ok := terraformWidget["alert_graph_definition"].([]interface{}); ok && len(def) > 0 {
		if alertGraphDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.AlertGraphWidgetDefinitionAsWidgetDefinition(buildDatadogAlertGraphDefinition(alertGraphDefinition))
		}
	} else if def, ok := terraformWidget["alert_value_definition"].([]interface{}); ok && len(def) > 0 {
		if alertValueDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.AlertValueWidgetDefinitionAsWidgetDefinition(buildDatadogAlertValueDefinition(alertValueDefinition))
		}
	} else if def, ok := terraformWidget["change_definition"].([]interface{}); ok && len(def) > 0 {
		if changeDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ChangeWidgetDefinitionAsWidgetDefinition(buildDatadogChangeDefinition(changeDefinition))
		}
	} else if def, ok := terraformWidget["check_status_definition"].([]interface{}); ok && len(def) > 0 {
		if checkStatusDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.CheckStatusWidgetDefinitionAsWidgetDefinition(buildDatadogCheckStatusDefinition(checkStatusDefinition))
		}
	} else if def, ok := terraformWidget["distribution_definition"].([]interface{}); ok && len(def) > 0 {
		if distributionDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.DistributionWidgetDefinitionAsWidgetDefinition(buildDatadogDistributionDefinition(distributionDefinition))
		}
	} else if def, ok := terraformWidget["event_stream_definition"].([]interface{}); ok && len(def) > 0 {
		if eventStreamDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.EventStreamWidgetDefinitionAsWidgetDefinition(buildDatadogEventStreamDefinition(eventStreamDefinition))
		}
	} else if def, ok := terraformWidget["event_timeline_definition"].([]interface{}); ok && len(def) > 0 {
		if eventTimelineDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.EventTimelineWidgetDefinitionAsWidgetDefinition(buildDatadogEventTimelineDefinition(eventTimelineDefinition))
		}
	} else if def, ok := terraformWidget["free_text_definition"].([]interface{}); ok && len(def) > 0 {
		if freeTextDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.FreeTextWidgetDefinitionAsWidgetDefinition(buildDatadogFreeTextDefinition(freeTextDefinition))
		}
	} else if def, ok := terraformWidget["heatmap_definition"].([]interface{}); ok && len(def) > 0 {
		if heatmapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.HeatMapWidgetDefinitionAsWidgetDefinition(buildDatadogHeatmapDefinition(heatmapDefinition))
		}
	} else if def, ok := terraformWidget["hostmap_definition"].([]interface{}); ok && len(def) > 0 {
		if hostDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.HostMapWidgetDefinitionAsWidgetDefinition(buildDatadogHostmapDefinition(hostDefinition))
		}
	} else if def, ok := terraformWidget["iframe_definition"].([]interface{}); ok && len(def) > 0 {
		if iframeDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.IFrameWidgetDefinitionAsWidgetDefinition(buildDatadogIframeDefinition(iframeDefinition))
		}
	} else if def, ok := terraformWidget["image_definition"].([]interface{}); ok && len(def) > 0 {
		if imageDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ImageWidgetDefinitionAsWidgetDefinition(buildDatadogImageDefinition(imageDefinition))
		}
	} else if def, ok := terraformWidget["log_stream_definition"].([]interface{}); ok && len(def) > 0 {
		if logStreamDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.LogStreamWidgetDefinitionAsWidgetDefinition(buildDatadogLogStreamDefinition(logStreamDefinition))
		}
	} else if def, ok := terraformWidget["manage_status_definition"].([]interface{}); ok && len(def) > 0 {
		if manageStatusDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.MonitorSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogManageStatusDefinition(manageStatusDefinition))
		}
	} else if def, ok := terraformWidget["note_definition"].([]interface{}); ok && len(def) > 0 {
		if noteDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.NoteWidgetDefinitionAsWidgetDefinition(buildDatadogNoteDefinition(noteDefinition))
		}
	} else if def, ok := terraformWidget["query_value_definition"].([]interface{}); ok && len(def) > 0 {
		if queryValueDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.QueryValueWidgetDefinitionAsWidgetDefinition(buildDatadogQueryValueDefinition(queryValueDefinition))
		}
	} else if def, ok := terraformWidget["query_table_definition"].([]interface{}); ok && len(def) > 0 {
		if queryTableDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TableWidgetDefinitionAsWidgetDefinition(buildDatadogQueryTableDefinition(queryTableDefinition))
		}
	} else if def, ok := terraformWidget["scatterplot_definition"].([]interface{}); ok && len(def) > 0 {
		if scatterplotDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ScatterPlotWidgetDefinitionAsWidgetDefinition(buildDatadogScatterplotDefinition(scatterplotDefinition))
		}
	} else if def, ok := terraformWidget["servicemap_definition"].([]interface{}); ok && len(def) > 0 {
		if serviceMapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ServiceMapWidgetDefinitionAsWidgetDefinition(buildDatadogServiceMapDefinition(serviceMapDefinition))
		}
	} else if def, ok := terraformWidget["service_level_objective_definition"].([]interface{}); ok && len(def) > 0 {
		if serviceLevelObjectiveDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.SLOWidgetDefinitionAsWidgetDefinition(buildDatadogServiceLevelObjectiveDefinition(serviceLevelObjectiveDefinition))
		}
	} else if def, ok := terraformWidget["slo_list_definition"].([]interface{}); ok && len(def) > 0 {
		if sloListDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.SLOListWidgetDefinitionAsWidgetDefinition(buildDatadogSloListDefinition(sloListDefinition))
		}
	} else if def, ok := terraformWidget["sunburst_definition"].([]interface{}); ok && len(def) > 0 {
		if sunburstDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.SunburstWidgetDefinitionAsWidgetDefinition(buildDatadogSunburstDefinition(sunburstDefinition))
		}
	} else if def, ok := terraformWidget["timeseries_definition"].([]interface{}); ok && len(def) > 0 {
		if timeseriesDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TimeseriesWidgetDefinitionAsWidgetDefinition(buildDatadogTimeseriesDefinition(timeseriesDefinition))
		}
	} else if def, ok := terraformWidget["toplist_definition"].([]interface{}); ok && len(def) > 0 {
		if toplistDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ToplistWidgetDefinitionAsWidgetDefinition(buildDatadogToplistDefinition(toplistDefinition))
		}
	} else if def, ok := terraformWidget["topology_map_definition"].([]interface{}); ok && len(def) > 0 {
		if topologyMapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TopologyMapWidgetDefinitionAsWidgetDefinition(buildDatadogTopologyMapDefinition(topologyMapDefinition))
		}
	} else if def, ok := terraformWidget["trace_service_definition"].([]interface{}); ok && len(def) > 0 {
		if traceServiceDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ServiceSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogTraceServiceDefinition(traceServiceDefinition))
		}
	} else if def, ok := terraformWidget["treemap_definition"].([]interface{}); ok && len(def) > 0 {
		if treemapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TreeMapWidgetDefinitionAsWidgetDefinition(buildDatadogTreemapDefinition(treemapDefinition))
		}
	} else if def, ok := terraformWidget["geomap_definition"].([]interface{}); ok && len(def) > 0 {
		if geomapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.GeomapWidgetDefinitionAsWidgetDefinition(buildDatadogGeomapDefinition(geomapDefinition))
		}
	} else if def, ok := terraformWidget["list_stream_definition"].([]interface{}); ok && len(def) > 0 {
		if listStreamDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ListStreamWidgetDefinitionAsWidgetDefinition(buildDatadogListStreamDefinition(listStreamDefinition))
		}
	} else if def, ok := terraformWidget["run_workflow_definition"].([]interface{}); ok && len(def) > 0 {
		if runWorkflowDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.RunWorkflowWidgetDefinitionAsWidgetDefinition(buildDatadogRunWorkflowDefinition(runWorkflowDefinition))
		}
	} else {
		return nil, fmt.Errorf("failed to find valid definition in widget configuration")
	}

	datadogWidget := datadogV1.NewWidget(definition)

	// Build widget layout
	if wl, ok := terraformWidget["widget_layout"].([]interface{}); ok && len(wl) != 0 {
		if v, ok := wl[0].(map[string]interface{}); ok && len(v) != 0 {
			datadogWidget.SetLayout(*buildDatadogWidgetLayout(v))
		}
	}

	return datadogWidget, nil
}

// Helper to build a list of Terraform widgets from a list of Datadog widgets
func buildTerraformWidgets(datadogWidgets *[]datadogV1.Widget, d *schema.ResourceData) (*[]map[string]interface{}, error) {

	terraformWidgets := make([]map[string]interface{}, len(*datadogWidgets))
	for i, datadogWidget := range *datadogWidgets {
		terraformWidget, err := buildTerraformWidget(&datadogWidget)
		if err != nil {
			return nil, err
		}
		terraformWidgets[i] = terraformWidget
	}
	return &terraformWidgets, nil
}

// Helper to build a Terraform widget from a Datadog widget
func buildTerraformWidget(datadogWidget *datadogV1.Widget) (map[string]interface{}, error) {
	terraformWidget := map[string]interface{}{}

	// Build layout
	if v, ok := datadogWidget.GetLayoutOk(); ok {
		widgetLayout := map[string]interface{}{
			"x":      (*v).GetX(),
			"y":      (*v).GetY(),
			"height": (*v).GetHeight(),
			"width":  (*v).GetWidth(),
		}
		if value, ok := (*v).GetIsColumnBreakOk(); ok {
			widgetLayout["is_column_break"] = value
		}
		terraformWidget["widget_layout"] = [](map[string]interface{}){
			widgetLayout,
		}
	}
	terraformWidget["id"] = datadogWidget.GetId()

	// Build definition
	widgetDefinition := datadogWidget.GetDefinition()
	if widgetDefinition.GroupWidgetDefinition != nil {
		terraformDefinition := buildTerraformGroupDefinition(widgetDefinition.GroupWidgetDefinition)
		terraformWidget["group_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.AlertGraphWidgetDefinition != nil {
		terraformDefinition := buildTerraformAlertGraphDefinition(widgetDefinition.AlertGraphWidgetDefinition)
		terraformWidget["alert_graph_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.AlertValueWidgetDefinition != nil {
		terraformDefinition := buildTerraformAlertValueDefinition(widgetDefinition.AlertValueWidgetDefinition)
		terraformWidget["alert_value_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ChangeWidgetDefinition != nil {
		terraformDefinition := buildTerraformChangeDefinition(widgetDefinition.ChangeWidgetDefinition)
		terraformWidget["change_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.CheckStatusWidgetDefinition != nil {
		terraformDefinition := buildTerraformCheckStatusDefinition(widgetDefinition.CheckStatusWidgetDefinition)
		terraformWidget["check_status_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.DistributionWidgetDefinition != nil {
		terraformDefinition := buildTerraformDistributionDefinition(widgetDefinition.DistributionWidgetDefinition)
		terraformWidget["distribution_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.EventStreamWidgetDefinition != nil {
		terraformDefinition := buildTerraformEventStreamDefinition(widgetDefinition.EventStreamWidgetDefinition)
		terraformWidget["event_stream_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.EventTimelineWidgetDefinition != nil {
		terraformDefinition := buildTerraformEventTimelineDefinition(widgetDefinition.EventTimelineWidgetDefinition)
		terraformWidget["event_timeline_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.FreeTextWidgetDefinition != nil {
		terraformDefinition := buildTerraformFreeTextDefinition(widgetDefinition.FreeTextWidgetDefinition)
		terraformWidget["free_text_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.HeatMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformHeatmapDefinition(widgetDefinition.HeatMapWidgetDefinition)
		terraformWidget["heatmap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.HostMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformHostmapDefinition(widgetDefinition.HostMapWidgetDefinition)
		terraformWidget["hostmap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.IFrameWidgetDefinition != nil {
		terraformDefinition := buildTerraformIframeDefinition(widgetDefinition.IFrameWidgetDefinition)
		terraformWidget["iframe_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ImageWidgetDefinition != nil {
		terraformDefinition := buildTerraformImageDefinition(widgetDefinition.ImageWidgetDefinition)
		terraformWidget["image_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.LogStreamWidgetDefinition != nil {
		terraformDefinition := buildTerraformLogStreamDefinition(widgetDefinition.LogStreamWidgetDefinition)
		terraformWidget["log_stream_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ListStreamWidgetDefinition != nil {
		terraformDefinition := buildTerraformListStreamDefinition(widgetDefinition.ListStreamWidgetDefinition)
		terraformWidget["list_stream_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.MonitorSummaryWidgetDefinition != nil {
		terraformDefinition := buildTerraformManageStatusDefinition(widgetDefinition.MonitorSummaryWidgetDefinition)
		terraformWidget["manage_status_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.NoteWidgetDefinition != nil {
		terraformDefinition := buildTerraformNoteDefinition(widgetDefinition.NoteWidgetDefinition)
		terraformWidget["note_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.QueryValueWidgetDefinition != nil {
		terraformDefinition := buildTerraformQueryValueDefinition(widgetDefinition.QueryValueWidgetDefinition)
		terraformWidget["query_value_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.TableWidgetDefinition != nil {
		terraformDefinition := buildTerraformQueryTableDefinition(widgetDefinition.TableWidgetDefinition)
		terraformWidget["query_table_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ScatterPlotWidgetDefinition != nil {
		terraformDefinition := buildTerraformScatterplotDefinition(widgetDefinition.ScatterPlotWidgetDefinition)
		terraformWidget["scatterplot_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ServiceMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformServiceMapDefinition(widgetDefinition.ServiceMapWidgetDefinition)
		terraformWidget["servicemap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.SLOWidgetDefinition != nil {
		terraformDefinition := buildTerraformServiceLevelObjectiveDefinition(widgetDefinition.SLOWidgetDefinition)
		terraformWidget["service_level_objective_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.SLOListWidgetDefinition != nil {
		terraformDefinition := buildTerraformSloListDefinition(widgetDefinition.SLOListWidgetDefinition)
		terraformWidget["slo_list_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.SunburstWidgetDefinition != nil {
		terraformDefinition := buildTerraformSunburstDefinition(widgetDefinition.SunburstWidgetDefinition)
		terraformWidget["sunburst_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.TimeseriesWidgetDefinition != nil {
		terraformDefinition := buildTerraformTimeseriesDefinition(widgetDefinition.TimeseriesWidgetDefinition)
		terraformWidget["timeseries_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ToplistWidgetDefinition != nil {
		terraformDefinition := buildTerraformToplistDefinition(widgetDefinition.ToplistWidgetDefinition)
		terraformWidget["toplist_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.TopologyMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformTopologyMapDefinition(widgetDefinition.TopologyMapWidgetDefinition)
		terraformWidget["topology_map_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ServiceSummaryWidgetDefinition != nil {
		terraformDefinition := buildTerraformTraceServiceDefinition(widgetDefinition.ServiceSummaryWidgetDefinition)
		terraformWidget["trace_service_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.TreeMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformTreemapDefinition(widgetDefinition.TreeMapWidgetDefinition)
		terraformWidget["treemap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.GeomapWidgetDefinition != nil {
		terraformDefinition := buildTerraformGeomapDefinition(widgetDefinition.GeomapWidgetDefinition)
		terraformWidget["geomap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.RunWorkflowWidgetDefinition != nil {
		terraformDefinition := buildTerraformRunWorkflowDefinition(widgetDefinition.RunWorkflowWidgetDefinition)
		terraformWidget["run_workflow_definition"] = []map[string]interface{}{terraformDefinition}
	} else {
		return nil, fmt.Errorf("unsupported widget type: %s", widgetDefinition.GetActualInstance())
	}
	return terraformWidget, nil
}

//
// Widget Layout helpers
//

func getWidgetLayoutSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"x": {
			Description: "The position of the widget on the x (horizontal) axis. Should be greater than or equal to 0.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"y": {
			Description: "The position of the widget on the y (vertical) axis. Should be greater than or equal to 0.",
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
			Description: "Whether the widget should be the first one on the second column in high density or not. Only for the new dashboard layout and only one widget in the dashboard should have this property set to `true`.",
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

//
// Group Widget helpers
//

func getGroupDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"widget": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of widgets in this group.",
			Elem: &schema.Resource{
				Schema: getNonGroupWidgetSchema(),
			},
		},
		"layout_type": {
			Type:             schema.TypeString,
			Required:         true,
			Description:      "The layout type of the group.",
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLayoutTypeFromValue),
		},
		"title": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The title of the group.",
		},
		"background_color": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The background color of the group title, options: `vivid_blue`, `vivid_purple`, `vivid_pink`, `vivid_orange`, `vivid_yellow`, `vivid_green`, `blue`, `purple`, `pink`, `orange`, `yellow`, `green`, `gray` or `white`",
		},
		"banner_img": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The image URL to display as a banner for the group.",
		},
		"show_title": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to show the title or not.",
			Default:     true,
		},
	}
}

func buildDatadogGroupDefinition(terraformGroupDefinition map[string]interface{}) (*datadogV1.GroupWidgetDefinition, error) {
	datadogGroupDefinition := datadogV1.NewGroupWidgetDefinitionWithDefaults()

	if v, ok := terraformGroupDefinition["widget"].([]interface{}); ok && len(v) != 0 {
		datadogWidgets, err := buildDatadogWidgets(&v)
		if err != nil {
			return nil, err
		}
		datadogGroupDefinition.SetWidgets(*datadogWidgets)
	} else {
		datadogGroupDefinition.SetWidgets([]datadogV1.Widget{})
	}
	if v, ok := terraformGroupDefinition["layout_type"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetLayoutType(datadogV1.WidgetLayoutType(v))
	}
	if v, ok := terraformGroupDefinition["title"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetTitle(v)
	}
	if v, ok := terraformGroupDefinition["background_color"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetBackgroundColor(v)
	}
	if v, ok := terraformGroupDefinition["banner_img"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetBannerImg(v)
	}
	if v, ok := terraformGroupDefinition["show_title"].(bool); ok {
		datadogGroupDefinition.SetShowTitle(v)
	}

	return datadogGroupDefinition, nil
}

func buildTerraformGroupDefinition(datadogGroupDefinition *datadogV1.GroupWidgetDefinition) map[string]interface{} {
	terraformGroupDefinition := map[string]interface{}{}

	var groupWidgets []map[string]interface{}
	for _, datadogGroupWidgets := range datadogGroupDefinition.Widgets {
		newGroupWidget, _ := buildTerraformWidget(&datadogGroupWidgets)

		groupWidgets = append(groupWidgets, newGroupWidget)
	}
	terraformGroupDefinition["widget"] = groupWidgets

	if v, ok := datadogGroupDefinition.GetLayoutTypeOk(); ok {
		terraformGroupDefinition["layout_type"] = v
	}
	if v, ok := datadogGroupDefinition.GetTitleOk(); ok {
		terraformGroupDefinition["title"] = v
	}
	if v, ok := datadogGroupDefinition.GetBackgroundColorOk(); ok {
		terraformGroupDefinition["background_color"] = v
	}
	if v, ok := datadogGroupDefinition.GetBannerImgOk(); ok {
		terraformGroupDefinition["banner_img"] = v
	}
	if v, ok := datadogGroupDefinition.GetShowTitleOk(); ok {
		terraformGroupDefinition["show_title"] = v
	}

	return terraformGroupDefinition
}

//
// Alert Graph Widget Definition helpers
//

func getAlertGraphDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alert_id": {
			Description: "The ID of the monitor used by the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"viz_type": {
			Description:      "Type of visualization to use when displaying the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetVizTypeFromValue),
			Required:         true,
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
	}
}

func buildDatadogAlertGraphDefinition(terraformDefinition map[string]interface{}) *datadogV1.AlertGraphWidgetDefinition {
	datadogDefinition := datadogV1.NewAlertGraphWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.AlertId = terraformDefinition["alert_id"].(string)
	datadogDefinition.VizType = datadogV1.WidgetVizType(terraformDefinition["viz_type"].(string))
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.PtrString(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.PtrString(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	return datadogDefinition
}

func buildTerraformAlertGraphDefinition(datadogDefinition *datadogV1.AlertGraphWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["alert_id"] = datadogDefinition.AlertId
	terraformDefinition["viz_type"] = datadogDefinition.VizType
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	return terraformDefinition
}

//
// Alert Value Widget Definition helpers
//

func getAlertValueDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alert_id": {
			Description: "The ID of the monitor used by the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"precision": {
			Description: "The precision to use when displaying the value. Use `*` for maximum precision.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"unit": {
			Description: "The unit for the value displayed in the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"text_align": {
			Description:      "The alignment of the text in the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
	}
}

func buildDatadogAlertValueDefinition(terraformDefinition map[string]interface{}) *datadogV1.AlertValueWidgetDefinition {
	datadogDefinition := datadogV1.NewAlertValueWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.AlertId = terraformDefinition["alert_id"].(string)
	// Optional params
	if v, ok := terraformDefinition["precision"].(int); ok && v != 0 {
		datadogDefinition.SetPrecision(int64(v))
	}
	if v, ok := terraformDefinition["unit"].(string); ok && len(v) != 0 {
		datadogDefinition.SetUnit(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
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
	return datadogDefinition
}

func buildTerraformAlertValueDefinition(datadogDefinition *datadogV1.AlertValueWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["alert_id"] = datadogDefinition.GetAlertId()
	// Optional params
	if v, ok := datadogDefinition.GetPrecisionOk(); ok {
		terraformDefinition["precision"] = *v
	}
	if v, ok := datadogDefinition.GetUnitOk(); ok {
		terraformDefinition["unit"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
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
	return terraformDefinition
}

//
// Change Widget Definition helpers
//

func getChangeDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getChangeRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func getChangeRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to Change requests
		"change_type": {
			Description:      "Whether to show absolute or relative change.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetChangeTypeFromValue),
			Optional:         true,
		},
		"compare_to": {
			Description:      "Choose from when to compare current data to.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetCompareToFromValue),
			Optional:         true,
		},
		"increase_good": {
			Description: "A Boolean indicating whether an increase in the value is good (displayed in green) or not (displayed in red).",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"order_by": {
			Description:      "What to order by.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetOrderByFromValue),
			Optional:         true,
		},
		"order_dir": {
			Description:      "Widget sorting method.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
			Optional:         true,
		},
		"show_present": {
			Description: "If set to `true`, displays the current value.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}
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
			terraformRequest["formula"] = buildTerraformFormula(v)
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

//
// Distribution Widget Definition helpers
//

func getDistributionDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getDistributionRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"legend_size": {
			Description:  "The size of the legend displayed in the widget.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateTimeseriesWidgetLegendSize,
		},
		"show_legend": {
			Description: "Whether or not to show the legend on this widget.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"live_span": getWidgetLiveSpanSchema(),
	}
}
func buildDatadogDistributionDefinition(terraformDefinition map[string]interface{}) *datadogV1.DistributionWidgetDefinition {
	datadogDefinition := datadogV1.NewDistributionWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogDistributionRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.SetShowLegend(v)
	}
	if v, ok := terraformDefinition["legend_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendSize(v)
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	return datadogDefinition
}
func buildTerraformDistributionDefinition(datadogDefinition *datadogV1.DistributionWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformDistributionRequests(&datadogDefinition.Requests)

	// Optional params
	if v, ok := datadogDefinition.GetShowLegendOk(); ok {
		terraformDefinition["show_legend"] = *v
	}
	if v, ok := datadogDefinition.GetLegendSizeOk(); ok {
		terraformDefinition["legend_size"] = *v
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

		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	return terraformDefinition
}

func getDistributionRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":               getMetricQuerySchema(),
		"apm_query":       getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":       getApmLogNetworkRumSecurityAuditQuerySchema(),
		"rum_query":       getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query":  getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":   getProcessQuerySchema(),
		"apm_stats_query": getApmStatsQuerySchema(),
		// Settings specific to Distribution requests
		"style": {
			Description: "The style of the widget graph. One nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
}
func buildDatadogDistributionRequests(terraformRequests *[]interface{}) *[]datadogV1.DistributionWidgetRequest {
	datadogRequests := make([]datadogV1.DistributionWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build DistributionRequest
		datadogDistributionRequest := datadogV1.NewDistributionWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogDistributionRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["apm_stats_query"].([]interface{}); ok && len(v) > 0 {
			apmStatsQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ApmStatsQuery = buildDatadogApmStatsQuery(apmStatsQuery)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogDistributionRequest.Style = buildDatadogWidgetStyle(v)
			}
		}

		datadogRequests[i] = *datadogDistributionRequest
	}
	return &datadogRequests
}
func buildTerraformDistributionRequests(datadogDistributionRequests *[]datadogV1.DistributionWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogDistributionRequests))
	for i, datadogRequest := range *datadogDistributionRequests {
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
		}
		if datadogRequest.Style != nil {
			style := buildTerraformWidgetStyle(*datadogRequest.Style)
			terraformRequest["style"] = []map[string]interface{}{style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Event Stream Widget Definition helpers
//

func getEventStreamDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"query": {
			Description: "The query to use in the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"event_size": {
			Description:      "The size to use to display an event.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetEventSizeFromValue),
			Optional:         true,
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"tags_execution": {
			Description: "The execution method for multi-value filters, options: `and` or `or`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

func buildDatadogEventStreamDefinition(terraformDefinition map[string]interface{}) *datadogV1.EventStreamWidgetDefinition {
	datadogDefinition := datadogV1.NewEventStreamWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetQuery(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["event_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetEventSize(datadogV1.WidgetEventSize(v))
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	if v, ok := terraformDefinition["tags_execution"].(string); ok && len(v) > 0 {
		datadogDefinition.SetTagsExecution(v)
	}
	return datadogDefinition
}

func buildTerraformEventStreamDefinition(datadogDefinition *datadogV1.EventStreamWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = datadogDefinition.Query
	// Optional params
	if datadogDefinition.EventSize != nil {
		terraformDefinition["event_size"] = *datadogDefinition.EventSize
	}
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if datadogDefinition.TagsExecution != nil {
		terraformDefinition["tags_execution"] = *datadogDefinition.TagsExecution
	}
	return terraformDefinition
}

//
// Event Timeline Widget Definition helpers
//

func getEventTimelineDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"query": {
			Description: "The query to use in the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"tags_execution": {
			Description: "The execution method for multi-value filters, options: `and` or `or`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

func buildDatadogEventTimelineDefinition(terraformDefinition map[string]interface{}) *datadogV1.EventTimelineWidgetDefinition {
	datadogDefinition := datadogV1.NewEventTimelineWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetQuery(terraformDefinition["query"].(string))
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	if v, ok := terraformDefinition["tags_execution"].(string); ok && len(v) > 0 {
		datadogDefinition.SetTagsExecution(v)
	}
	return datadogDefinition
}

func buildTerraformEventTimelineDefinition(datadogDefinition *datadogV1.EventTimelineWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = datadogDefinition.GetQuery()
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

		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetTagsExecutionOk(); ok {
		terraformDefinition["tags_execution"] = *v
	}
	return terraformDefinition
}

//
// Check Status Widget Definition helpers
//

func getCheckStatusDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"check": {
			Description: "The check to use in the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"grouping": {
			Description:      "The kind of grouping to use.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetGroupingFromValue),
			Required:         true,
		},
		"group": {
			Description: "The check group to use in the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"group_by": {
			Description: "When `grouping = \"cluster\"`, indicates a list of tags to use for grouping.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"tags": {
			Description: "A list of tags to use in the widget.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
	}
}

func buildDatadogCheckStatusDefinition(terraformDefinition map[string]interface{}) *datadogV1.CheckStatusWidgetDefinition {
	datadogDefinition := datadogV1.NewCheckStatusWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetCheck(terraformDefinition["check"].(string))
	datadogDefinition.SetGrouping(datadogV1.WidgetGrouping(terraformDefinition["grouping"].(string)))
	// Optional params
	if v, ok := terraformDefinition["group"].(string); ok && len(v) != 0 {
		datadogDefinition.SetGroup(v)
	}
	if terraformGroupBys, ok := terraformDefinition["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]string, len(terraformGroupBys))
		for i, groupBy := range terraformGroupBys {
			datadogGroupBys[i] = groupBy.(string)
		}
		datadogDefinition.SetGroupBy(datadogGroupBys)
	}
	if terraformTags, ok := terraformDefinition["tags"].([]interface{}); ok && len(terraformTags) > 0 {
		datadogTags := make([]string, len(terraformTags))
		for i, tag := range terraformTags {
			datadogTags[i] = tag.(string)
		}
		datadogDefinition.SetTags(datadogTags)
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	return datadogDefinition
}

func buildTerraformCheckStatusDefinition(datadogDefinition *datadogV1.CheckStatusWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["check"] = datadogDefinition.GetCheck()
	terraformDefinition["grouping"] = datadogDefinition.GetGrouping()
	// Optional params
	if v, ok := datadogDefinition.GetGroupOk(); ok {
		terraformDefinition["group"] = *v
	}
	if v, ok := datadogDefinition.GetGroupByOk(); ok {
		terraformGroupBys := make([]string, len(*v))
		for i, datadogGroupBy := range *v {
			terraformGroupBys[i] = datadogGroupBy
		}
		terraformDefinition["group_by"] = terraformGroupBys
	}
	if v, ok := datadogDefinition.GetTagsOk(); ok {
		terraformTags := make([]string, len(*v))
		for i, datadogTag := range *v {
			terraformTags[i] = datadogTag
		}
		terraformDefinition["tags"] = terraformTags
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	return terraformDefinition
}

//
// Free Text Definition helpers
//

func getFreeTextDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"text": {
			Description: "The text to display in the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"color": {
			Description: "The color of the text in the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"font_size": {
			Description: "The size of the text in the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"text_align": {
			Description:      "The alignment of the text in the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
	}
}

func buildDatadogFreeTextDefinition(terraformDefinition map[string]interface{}) *datadogV1.FreeTextWidgetDefinition {
	datadogDefinition := datadogV1.NewFreeTextWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetText(terraformDefinition["text"].(string))
	// Optional params
	if v, ok := terraformDefinition["color"].(string); ok && len(v) != 0 {
		datadogDefinition.SetColor(v)
	}
	if v, ok := terraformDefinition["font_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetFontSize(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	return datadogDefinition
}

func buildTerraformFreeTextDefinition(datadogDefinition *datadogV1.FreeTextWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["text"] = datadogDefinition.GetText()
	// Optional params
	if v, ok := datadogDefinition.GetColorOk(); ok {
		terraformDefinition["color"] = *v
	}
	if v, ok := datadogDefinition.GetFontSizeOk(); ok {
		terraformDefinition["font_size"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
	}
	return terraformDefinition
}

//
// Heatmap Widget Definition helpers
//

func getHeatmapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getHeatmapRequestSchema(),
			},
		},
		"yaxis": {
			Description: "A nested block describing the Y-Axis Controls. The structure of this block is described below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"event": {
			Description: "The definition of the event to overlay on the graph. Multiple `event` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetEventSchema(),
			},
		},
		"show_legend": {
			Description: "Whether or not to show the legend on this widget.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"legend_size": {
			Description:  "The size of the legend displayed in the widget.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateTimeseriesWidgetLegendSize,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}
func buildDatadogHeatmapDefinition(terraformDefinition map[string]interface{}) *datadogV1.HeatMapWidgetDefinition {
	datadogDefinition := datadogV1.NewHeatMapWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogHeatmapRequests(&terraformRequests)
	// Optional params
	if axis, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(v)
		}
	}
	if v, ok := terraformDefinition["event"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Events = *buildDatadogWidgetEvents(&v)
	}
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.SetShowLegend(v)
	}
	if v, ok := terraformDefinition["legend_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendSize(v)
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildTerraformHeatmapDefinition(datadogDefinition *datadogV1.HeatMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformHeatmapRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["yaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetEventsOk(); ok {
		terraformDefinition["event"] = buildTerraformWidgetEvents(v)
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
	if v, ok := datadogDefinition.GetShowLegendOk(); ok {
		terraformDefinition["show_legend"] = *v
	}
	if v, ok := datadogDefinition.GetLegendSizeOk(); ok {
		terraformDefinition["legend_size"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["live_span"] = v.GetLiveSpan()

	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func getHeatmapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to Heatmap requests
		"style": {
			Description: "The style of the widget graph. One nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
}
func buildDatadogHeatmapRequests(terraformRequests *[]interface{}) *[]datadogV1.HeatMapWidgetRequest {
	datadogRequests := make([]datadogV1.HeatMapWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		if r == nil {
			continue
		}
		terraformRequest := r.(map[string]interface{})
		// Build HeatmapRequest
		datadogHeatmapRequest := datadogV1.NewHeatMapWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogHeatmapRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
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
			datadogHeatmapRequest.SetQueries(queries)
			datadogHeatmapRequest.SetResponseFormat(datadogV1.FormulaAndFunctionResponseFormat("timeseries"))
		}
		if v, ok := terraformRequest["formula"].([]interface{}); ok && len(v) > 0 {
			formulas := make([]datadogV1.WidgetFormula, len(v))
			for i, formula := range v {
				if formula == nil {
					continue
				}
				formulas[i] = *buildDatadogFormula(formula.(map[string]interface{}))
			}
			datadogHeatmapRequest.SetFormulas(formulas)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogHeatmapRequest.Style = buildDatadogWidgetStyle(v)
			}
		}
		datadogRequests[i] = *datadogHeatmapRequest
	}
	return &datadogRequests
}
func buildTerraformHeatmapRequests(datadogHeatmapRequests *[]datadogV1.HeatMapWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogHeatmapRequests))
	for i, datadogRequest := range *datadogHeatmapRequests {
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
			terraformRequest["formula"] = buildTerraformFormula(v)
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Hostmap Widget Definition helpers
//

func getHostmapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"fill": {
						Description: "The query used to fill the map. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getHostmapRequestSchema(),
						},
					},
					"size": {
						Description: "The query used to size the map. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getHostmapRequestSchema(),
						},
					},
				},
			},
		},
		"node_type": {
			Description:      "The type of node used.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetNodeTypeFromValue),
			Optional:         true,
		},
		"no_metric_hosts": {
			Description: "A Boolean indicating whether to show nodes with no metrics.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"no_group_hosts": {
			Description: "A Boolean indicating whether to show ungrouped nodes.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"group": {
			Description: "The list of tags to group nodes by.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"scope": {
			Description: "The list of tags to filter nodes by.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"style": {
			Description: "The style of the widget graph. One nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"palette": {
						Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"palette_flip": {
						Description: "A Boolean indicating whether to flip the palette tones.",
						Type:        schema.TypeBool,
						Optional:    true,
					},
					"fill_min": {
						Description: "The min value to use to color the map.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"fill_max": {
						Description: "The max value to use to color the map.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}
func buildDatadogHostmapDefinition(terraformDefinition map[string]interface{}) *datadogV1.HostMapWidgetDefinition {

	// Required params
	datadogDefinition := datadogV1.NewHostMapWidgetDefinitionWithDefaults()
	if v, ok := terraformDefinition["request"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		terraformRequests := v[0].(map[string]interface{})
		datadogRequests := datadogV1.NewHostMapWidgetDefinitionRequests()
		if terraformFillArray, ok := terraformRequests["fill"].([]interface{}); ok && len(terraformFillArray) > 0 {
			terraformFill := terraformFillArray[0].(map[string]interface{})
			datadogRequests.Fill = buildDatadogHostmapRequest(terraformFill)
		}
		if terraformSizeArray, ok := terraformRequests["size"].([]interface{}); ok && len(terraformSizeArray) > 0 {
			terraformSize := terraformSizeArray[0].(map[string]interface{})
			datadogRequests.Size = buildDatadogHostmapRequest(terraformSize)
		}
		datadogDefinition.SetRequests(*datadogRequests)
	}

	// Optional params
	if v, ok := terraformDefinition["node_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetNodeType(datadogV1.WidgetNodeType(v))
	}
	if v, ok := terraformDefinition["no_metric_hosts"].(bool); ok {
		datadogDefinition.SetNoMetricHosts(v)
	}
	if v, ok := terraformDefinition["no_group_hosts"].(bool); ok {
		datadogDefinition.SetNoGroupHosts(v)
	}
	if terraformGroups, ok := terraformDefinition["group"].([]interface{}); ok && len(terraformGroups) > 0 {
		datadogGroups := make([]string, len(terraformGroups))
		for i, group := range terraformGroups {
			datadogGroups[i] = group.(string)
		}
		datadogDefinition.Group = datadogGroups
	}
	if terraformScopes, ok := terraformDefinition["scope"].([]interface{}); ok && len(terraformScopes) > 0 {
		datadogScopes := make([]string, len(terraformScopes))
		for i, Scope := range terraformScopes {
			datadogScopes[i] = Scope.(string)
		}
		datadogDefinition.SetScope(datadogScopes)
	}
	if style, ok := terraformDefinition["style"].([]interface{}); ok && len(style) > 0 {
		if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Style = buildDatadogHostmapRequestStyle(v)
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
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildTerraformHostmapDefinition(datadogDefinition *datadogV1.HostMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformRequests := map[string]interface{}{}
	if v, ok := datadogDefinition.Requests.GetSizeOk(); ok {
		terraformSize := buildTerraformHostmapRequest(v)

		terraformRequests["size"] = []map[string]interface{}{*terraformSize}
	}
	if v, ok := datadogDefinition.Requests.GetFillOk(); ok {
		terraformFill := buildTerraformHostmapRequest(v)

		terraformRequests["fill"] = []map[string]interface{}{*terraformFill}
	}
	terraformDefinition["request"] = []map[string]interface{}{terraformRequests}
	// Optional params
	if v, ok := datadogDefinition.GetNodeTypeOk(); ok {
		terraformDefinition["node_type"] = *v
	}
	if v, ok := datadogDefinition.GetNoMetricHostsOk(); ok {
		terraformDefinition["no_metric_hosts"] = *v
	}
	if v, ok := datadogDefinition.GetNoGroupHostsOk(); ok {
		terraformDefinition["no_group_hosts"] = *v
	}
	if v, ok := datadogDefinition.GetGroupOk(); ok {
		terraformGroups := make([]string, len(*v))
		for i, datadogGroup := range *v {
			terraformGroups[i] = datadogGroup
		}
		terraformDefinition["group"] = terraformGroups
	}
	if v, ok := datadogDefinition.GetScopeOk(); ok {
		terraformScopes := make([]string, len(*v))
		for i, datadogScope := range *v {
			terraformScopes[i] = datadogScope
		}
		terraformDefinition["scope"] = terraformScopes
	}
	if v, ok := datadogDefinition.GetStyleOk(); ok {
		style := buildTerraformHostmapRequestStyle(v)
		terraformDefinition["style"] = []map[string]interface{}{style}
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
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func getHostmapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement at least one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
	}
}
func buildDatadogHostmapRequest(terraformRequest map[string]interface{}) *datadogV1.HostMapRequest {

	datadogHostmapRequest := &datadogV1.HostMapRequest{}
	if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
		datadogHostmapRequest.SetQ(v)
	} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
		apmQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
	} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
		logQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
	} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
		processQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
	} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
		rumQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
	} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
		securityQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
	}

	return datadogHostmapRequest
}
func buildTerraformHostmapRequest(datadogHostmapRequest *datadogV1.HostMapRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}
	if v, ok := datadogHostmapRequest.GetQOk(); ok {
		terraformRequest["q"] = v
	} else if v, ok := datadogHostmapRequest.GetApmQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogHostmapRequest.GetLogQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogHostmapRequest.GetProcessQueryOk(); ok {
		terraformQuery := buildTerraformProcessQuery(*v)
		terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogHostmapRequest.GetRumQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogHostmapRequest.GetSecurityQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
	}
	return &terraformRequest
}

//
// Iframe Definition helpers
//

func getIframeDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Description: "The URL to use as a data source for the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func buildDatadogIframeDefinition(terraformDefinition map[string]interface{}) *datadogV1.IFrameWidgetDefinition {
	datadogDefinition := datadogV1.NewIFrameWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetUrl(terraformDefinition["url"].(string))
	return datadogDefinition
}

func buildTerraformIframeDefinition(datadogDefinition *datadogV1.IFrameWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["url"] = datadogDefinition.GetUrl()
	return terraformDefinition
}

//
// Image Widget Definition helpers
//

func getImageDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Description: "The URL to use as a data source for the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"url_dark_theme": {
			Description: "The URL in dark mode to use as a data source for the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"sizing": {
			Description:      "The preferred method to adapt the dimensions of the image. The values are based on the image `object-fit` CSS properties. Note: `zoom`, `fit` and `center` values are deprecated.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetImageSizingFromValue),
			Optional:         true,
		},
		"margin": {
			Description:      "The margins to use around the image. Note: `small` and `large` values are deprecated.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetMarginFromValue),
			Optional:         true,
		},
		"has_background": {
			Description: "Whether to display a background or not.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"has_border": {
			Description: "Whether to display a border or not.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"horizontal_align": {
			Description:      "The horizontal alignment for the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetHorizontalAlignFromValue),
			Optional:         true,
		},
		"vertical_align": {
			Description:      "The vertical alignment for the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetVerticalAlignFromValue),
			Optional:         true,
		},
	}
}

func buildDatadogImageDefinition(terraformDefinition map[string]interface{}) *datadogV1.ImageWidgetDefinition {
	datadogDefinition := datadogV1.NewImageWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetUrl(terraformDefinition["url"].(string))
	// Optional params
	if v, ok := terraformDefinition["url_dark_theme"].(string); ok && len(v) != 0 {
		datadogDefinition.SetUrlDarkTheme(v)
	}
	if v, ok := terraformDefinition["sizing"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSizing(datadogV1.WidgetImageSizing(v))
	}
	if v, ok := terraformDefinition["margin"].(string); ok && len(v) != 0 {
		datadogDefinition.SetMargin(datadogV1.WidgetMargin(v))
	}
	if v, ok := terraformDefinition["has_background"].(bool); ok {
		datadogDefinition.SetHasBackground(v)
	}
	if v, ok := terraformDefinition["has_border"].(bool); ok {
		datadogDefinition.SetHasBorder(v)
	}
	if v, ok := terraformDefinition["horizontal_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetHorizontalAlign(datadogV1.WidgetHorizontalAlign(v))
	}
	if v, ok := terraformDefinition["vertical_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetVerticalAlign(datadogV1.WidgetVerticalAlign(v))
	}
	return datadogDefinition
}

func buildTerraformImageDefinition(datadogDefinition *datadogV1.ImageWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["url"] = datadogDefinition.GetUrl()
	// Optional params
	if v, ok := datadogDefinition.GetUrlDarkThemeOk(); ok {
		terraformDefinition["url_dark_theme"] = *v
	}
	if v, ok := datadogDefinition.GetSizingOk(); ok {
		terraformDefinition["sizing"] = *v
	}
	if v, ok := datadogDefinition.GetMarginOk(); ok {
		terraformDefinition["margin"] = *v
	}
	if v, ok := datadogDefinition.GetHasBackgroundOk(); ok {
		terraformDefinition["has_background"] = *v
	}
	if v, ok := datadogDefinition.GetHasBorderOk(); ok {
		terraformDefinition["has_border"] = *v
	}
	if v, ok := datadogDefinition.GetHorizontalAlignOk(); ok {
		terraformDefinition["horizontal_align"] = *v
	}
	if v, ok := datadogDefinition.GetVerticalAlignOk(); ok {
		terraformDefinition["vertical_align"] = *v
	}
	return terraformDefinition
}

//
// Log Stream Widget Definition helpers
//

func getLogStreamDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"indexes": {
			Description: "An array of index names to query in the stream.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"query": {
			Description: "The query to use in the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"columns": {
			Description: "Stringified list of columns to use, for example: `[\"column1\",\"column2\",\"column3\"]`.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"show_date_column": {
			Description: "If the date column should be displayed.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_message_column": {
			Description: "If the message column should be displayed.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"message_display": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The number of log lines to display.",
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetMessageDisplayFromValue),
		},
		"sort": {
			Description: "The facet and order to sort the data, for example: `{\"column\": \"time\", \"order\": \"desc\"}`.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetFieldSortSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
	}
}

func getWidgetFieldSortSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"column": {
			Description: "The facet path for the column.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"order": {
			Description:      "Widget sorting methods.",
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
		},
	}
}

func buildDatadogLogStreamDefinition(terraformDefinition map[string]interface{}) *datadogV1.LogStreamWidgetDefinition {
	datadogDefinition := datadogV1.NewLogStreamWidgetDefinitionWithDefaults()
	// Required params
	terraformIndexes := terraformDefinition["indexes"].([]interface{})
	datadogIndexes := make([]string, len(terraformIndexes))
	for i, index := range terraformIndexes {
		datadogIndexes[i] = index.(string)
	}
	datadogDefinition.SetIndexes(datadogIndexes)
	terraformColumns := terraformDefinition["columns"].([]interface{})
	datadogColumns := make([]string, len(terraformColumns))
	for i, column := range terraformColumns {
		datadogColumns[i] = column.(string)
	}
	datadogDefinition.SetColumns(datadogColumns)
	// Optional params
	if v, ok := terraformDefinition["query"].(string); ok && len(v) != 0 {
		datadogDefinition.SetQuery(v)
	}
	if v, ok := terraformDefinition["show_date_column"].(bool); ok {
		datadogDefinition.SetShowDateColumn(v)
	}
	if v, ok := terraformDefinition["show_message_column"].(bool); ok {
		datadogDefinition.SetShowMessageColumn(v)
	}
	if v, ok := terraformDefinition["message_display"].(string); ok && len(v) != 0 {
		datadogDefinition.SetMessageDisplay(datadogV1.WidgetMessageDisplay(v))
	}
	if v, ok := terraformDefinition["sort"].([]interface{}); ok && len(v) > 0 {
		if v, ok := v[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Sort = buildDatadogWidgetFieldSort(v)
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	return datadogDefinition
}

func buildDatadogWidgetFieldSort(terraformWidgetFieldSort map[string]interface{}) *datadogV1.WidgetFieldSort {
	datadogWidgetFieldSort := &datadogV1.WidgetFieldSort{}
	if v, ok := terraformWidgetFieldSort["column"].(string); ok && len(v) != 0 {
		datadogWidgetFieldSort.SetColumn(v)
	}
	if v, ok := terraformWidgetFieldSort["order"].(string); ok && len(v) != 0 {
		datadogWidgetFieldSort.SetOrder(datadogV1.WidgetSort(v))
	}
	return datadogWidgetFieldSort
}

func buildTerraformLogStreamDefinition(datadogDefinition *datadogV1.LogStreamWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Optional params

	if v, ok := datadogDefinition.GetIndexesOk(); ok {
		terraformDefinition["indexes"] = *v
	}
	if v, ok := datadogDefinition.GetQueryOk(); ok {
		terraformDefinition["query"] = *v
	}
	if v, ok := datadogDefinition.GetColumnsOk(); ok {
		terraformColumns := make([]string, len(*v))
		for i, datadogColumn := range *v {
			terraformColumns[i] = datadogColumn
		}
		terraformDefinition["columns"] = terraformColumns
	}
	if v, ok := datadogDefinition.GetShowDateColumnOk(); ok {
		terraformDefinition["show_date_column"] = *v
	}
	if v, ok := datadogDefinition.GetShowMessageColumnOk(); ok {
		terraformDefinition["show_message_column"] = *v
	}
	if v, ok := datadogDefinition.GetMessageDisplayOk(); ok {
		terraformDefinition["message_display"] = *v
	}
	if v, ok := datadogDefinition.GetSortOk(); ok {
		sort := buildTerraformWidgetFieldSort(v)
		terraformDefinition["sort"] = []map[string]interface{}{sort}
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	return terraformDefinition
}

func buildTerraformWidgetFieldSort(datadogWidgetFieldSort *datadogV1.WidgetFieldSort) map[string]interface{} {
	terraformWidgetFieldSort := map[string]interface{}{}
	if v, ok := datadogWidgetFieldSort.GetColumnOk(); ok {
		terraformWidgetFieldSort["column"] = *v
	}
	if v, ok := datadogWidgetFieldSort.GetOrderOk(); ok {
		terraformWidgetFieldSort["order"] = string(*v)
	}
	return terraformWidgetFieldSort
}

//
// Manage Status Widget Definition helpers
//

func getManageStatusDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"query": {
			Description: "The query to use in the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"summary_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "The summary type to use.",
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSummaryTypeFromValue),
		},
		"sort": {
			Description:      "The method to sort the monitors.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetMonitorSummarySortFromValue),
			Optional:         true,
		},
		"display_format": {
			Description:      "The display setting to use.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetMonitorSummaryDisplayFormatFromValue),
			Optional:         true,
		},
		"color_preference": {
			Description:      "Whether to colorize text or background.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetColorPreferenceFromValue),
			Optional:         true,
		},
		"hide_zero_counts": {
			Description: "A Boolean indicating whether to hide empty categories.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_last_triggered": {
			Description: "A Boolean indicating whether to show when monitors/groups last triggered.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_priority": {
			Description: "Whether to show the priorities column.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
	}
}

func buildDatadogManageStatusDefinition(terraformDefinition map[string]interface{}) *datadogV1.MonitorSummaryWidgetDefinition {
	datadogDefinition := datadogV1.NewMonitorSummaryWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetQuery(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["summary_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSummaryType(datadogV1.WidgetSummaryType(v))
	}
	if v, ok := terraformDefinition["sort"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSort(datadogV1.WidgetMonitorSummarySort(v))
	}
	if v, ok := terraformDefinition["display_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetDisplayFormat(datadogV1.WidgetMonitorSummaryDisplayFormat(v))
	}
	if v, ok := terraformDefinition["color_preference"].(string); ok && len(v) != 0 {
		datadogDefinition.SetColorPreference(datadogV1.WidgetColorPreference(v))
	}
	if v, ok := terraformDefinition["hide_zero_counts"].(bool); ok {
		datadogDefinition.SetHideZeroCounts(v)
	}
	if v, ok := terraformDefinition["show_last_triggered"].(bool); ok {
		datadogDefinition.SetShowLastTriggered(v)
	}
	if v, ok := terraformDefinition["show_priority"].(bool); ok {
		datadogDefinition.SetShowPriority(v)
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
	return datadogDefinition
}

func buildTerraformManageStatusDefinition(datadogDefinition *datadogV1.MonitorSummaryWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = datadogDefinition.GetQuery()
	// Optional params
	if v, ok := datadogDefinition.GetSummaryTypeOk(); ok {
		terraformDefinition["summary_type"] = *v
	}
	if v, ok := datadogDefinition.GetSortOk(); ok {
		terraformDefinition["sort"] = *v
	}
	if v, ok := datadogDefinition.GetDisplayFormatOk(); ok {
		terraformDefinition["display_format"] = *v
	}
	if v, ok := datadogDefinition.GetColorPreferenceOk(); ok {
		terraformDefinition["color_preference"] = *v
	}
	if v, ok := datadogDefinition.GetHideZeroCountsOk(); ok {
		terraformDefinition["hide_zero_counts"] = *v
	}
	if v, ok := datadogDefinition.GetShowLastTriggeredOk(); ok {
		terraformDefinition["show_last_triggered"] = *v
	}
	if v, ok := datadogDefinition.GetShowPriorityOk(); ok {
		terraformDefinition["show_priority"] = *v
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
	return terraformDefinition
}

//
// Note Widget Definition helpers
//

func getNoteDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"content": {
			Description:  "The content of the note.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"background_color": {
			Description: "The background color of the note.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"font_size": {
			Description: "The size of the text.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"text_align": {
			Description:      "The alignment of the widget's text.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"vertical_align": {
			Description:      "The vertical alignment for the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetVerticalAlignFromValue),
			Optional:         true,
		},
		"has_padding": {
			Description: "Whether to add padding or not.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"show_tick": {
			Description: "Whether to show a tick or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"tick_pos": {
			Description: "When `tick = true`, a string with a percent sign indicating the position of the tick, for example: `tick_pos = \"50%\"` is centered alignment.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"tick_edge": {
			Description:      "When `tick = true`, a string indicating on which side of the widget the tick should be displayed.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTickEdgeFromValue),
			Optional:         true,
		},
	}
}

func buildDatadogNoteDefinition(terraformDefinition map[string]interface{}) *datadogV1.NoteWidgetDefinition {
	datadogDefinition := datadogV1.NewNoteWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetContent(terraformDefinition["content"].(string))
	// Optional params
	if v, ok := terraformDefinition["background_color"].(string); ok && len(v) != 0 {
		datadogDefinition.SetBackgroundColor(v)
	}
	if v, ok := terraformDefinition["font_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetFontSize(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["vertical_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetVerticalAlign(datadogV1.WidgetVerticalAlign(v))
	}
	if v, ok := terraformDefinition["has_padding"].(bool); ok {
		datadogDefinition.SetHasPadding(v)
	}
	if v, ok := terraformDefinition["show_tick"]; ok {
		datadogDefinition.SetShowTick(v.(bool))
	}
	if v, ok := terraformDefinition["tick_pos"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTickPos(v)
	}
	if v, ok := terraformDefinition["tick_edge"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTickEdge(datadogV1.WidgetTickEdge(v))
	}
	return datadogDefinition
}

func buildTerraformNoteDefinition(datadogDefinition *datadogV1.NoteWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["content"] = datadogDefinition.GetContent()
	// Optional params
	if v, ok := datadogDefinition.GetBackgroundColorOk(); ok {
		terraformDefinition["background_color"] = *v
	}
	if v, ok := datadogDefinition.GetFontSizeOk(); ok {
		terraformDefinition["font_size"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
	}
	if v, ok := datadogDefinition.GetVerticalAlignOk(); ok {
		terraformDefinition["vertical_align"] = *v
	}
	if v, ok := datadogDefinition.GetHasPaddingOk(); ok {
		terraformDefinition["has_padding"] = *v
	}
	if v, ok := datadogDefinition.GetShowTickOk(); ok {
		terraformDefinition["show_tick"] = *v
	}
	if v, ok := datadogDefinition.GetTickPosOk(); ok {
		terraformDefinition["tick_pos"] = *v
	}
	if v, ok := datadogDefinition.GetTickEdgeOk(); ok {
		terraformDefinition["tick_edge"] = *v
	}
	return terraformDefinition
}

//
// Query Value Widget Definition helpers
//

func getQueryValueDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getQueryValueRequestSchema(),
			},
		},
		"autoscale": {
			Description: "A Boolean indicating whether to automatically scale the tile.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"custom_unit": {
			Description: "The unit for the value displayed in the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"precision": {
			Description: "The precision to use when displaying the tile.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"text_align": {
			Description:      "The alignment of the widget's text.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"timeseries_background": {
			Description: "Set a timeseries on the widget background.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: getQueryValueTimeseriesBackgroundSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func getQueryValueTimeseriesBackgroundSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"type": {
			Description:      "Whether the Timeseries is made using an area or bars.",
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTimeseriesBackgroundTypeFromValue),
		},
		"yaxis": {
			Description: "A nested block describing the Y-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
	}
}

func getQueryValueRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"audit_query":    getApmLogNetworkRumSecurityAuditQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to QueryValue requests
		"conditional_formats": {
			Description: "Conditional formats allow you to set the color of your widget content or background depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetConditionalFormatSchema(),
			},
		},
		"aggregator": {
			Description:      "The aggregator to use for time aggregation.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetAggregatorFromValue),
			Optional:         true,
		},
	}
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
			terraformRequest["formula"] = buildTerraformFormula(v)
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
func getQueryTableDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the `request` block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getQueryTableRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
		"has_search_bar": {
			Description:      "Controls the display of the search bar.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTableWidgetHasSearchBarFromValue),
			Optional:         true,
		},
	}
}
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	if v, ok := datadogDefinition.GetHasSearchBarOk(); ok {
		terraformDefinition["has_search_bar"] = *v
	}
	return terraformDefinition
}

func getQueryTableRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":               getMetricQuerySchema(),
		"apm_query":       getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":       getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":   getProcessQuerySchema(),
		"rum_query":       getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query":  getApmLogNetworkRumSecurityAuditQuerySchema(),
		"apm_stats_query": getApmStatsQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to QueryTable requests
		"conditional_formats": {
			Description: "Conditional formats allow you to set the color of your widget content or background, depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetConditionalFormatSchema(),
			},
		},
		"alias": {
			Description: "The alias for the column name (defaults to metric name).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"aggregator": {
			Description:      "The aggregator to use for time aggregation.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetAggregatorFromValue),
			Optional:         true,
		},
		"limit": {
			Description: "The number of lines to show in the table.",
			Type:        schema.TypeInt,
			Optional:    true,
		},
		"order": {
			Description:      "The sort order for the rows.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
			Optional:         true,
		},
		"cell_display_mode": {
			Description: "A list of display modes for each table cell. List items one of `number`, `bar`.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTableWidgetCellDisplayModeFromValue),
			},
		},
	}
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
			terraformRequest["formula"] = buildTerraformFormula(v)
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
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Scatterplot Widget Definition helpers
//

func getScatterplotDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"x": {
						Description: "The query used for the X-Axis. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block).",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getScatterplotRequestSchema(),
						},
					},
					"y": {
						Description: "The query used for the Y-Axis. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block).",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getScatterplotRequestSchema(),
						},
					},
					"scatterplot_table": {
						Description: "Scatterplot request containing formulas and functions.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getScatterplotTableRequestSchema(),
						},
					},
				},
			},
		},
		"xaxis": {
			Description: "A nested block describing the X-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"yaxis": {
			Description: "A nested block describing the Y-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"color_by_groups": {
			Description: "List of groups used for colors.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildTerraformScatterplotDefinition(datadogDefinition *datadogV1.ScatterPlotWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformRequests := map[string]interface{}{}
	if v, ok := datadogDefinition.Requests.GetXOk(); ok {
		terraformX := buildTerraformScatterplotRequest(v)

		terraformRequests["x"] = []map[string]interface{}{*terraformX}
	}
	if v, ok := datadogDefinition.Requests.GetYOk(); ok {
		terraformY := buildTerraformScatterplotRequest(v)

		terraformRequests["y"] = []map[string]interface{}{*terraformY}
	}
	if v, ok := datadogDefinition.Requests.GetTableOk(); ok {
		teraformScatterplotTableRequest := buildTerraformScatterplotTableRequest(v)
		terraformRequests["scatterplot_table"] = []map[string]interface{}{*teraformScatterplotTableRequest}
	}
	terraformDefinition["request"] = []map[string]interface{}{terraformRequests}

	// Optional params
	if v, ok := datadogDefinition.GetXaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["xaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["yaxis"] = []map[string]interface{}{axis}
	}

	if v, ok := datadogDefinition.GetColorByGroupsOk(); ok {
		terraformColorByGroups := make([]string, len(*v))
		for i, datadogColorByGroup := range *v {
			terraformColorByGroups[i] = datadogColorByGroup
		}
		terraformDefinition["color_by_groups"] = terraformColorByGroups
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func getScatterplotTableRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getScatterplotFormulaSchema(),
	}
}

func getScatterplotRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		// Settings specific to Scatterplot requests
		"aggregator": {
			Description:      "Aggregator used for the request.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetAggregatorFromValue),
			Optional:         true,
		},
	}
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

func buildTerraformScatterplotTableRequest(datadogScatterplotTableRequest *datadogV1.ScatterplotTableRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}

	if v, ok := datadogScatterplotTableRequest.GetQueriesOk(); ok {
		terraformRequest["query"] = buildTerraformQuery(v)
	}
	if v, ok := datadogScatterplotTableRequest.GetFormulasOk(); ok {
		terraformRequest["formula"] = buildTerraformScatterplotFormula(v)
	}

	return &terraformRequest
}

func buildTerraformScatterplotRequest(datadogScatterplotRequest *datadogV1.ScatterPlotRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}
	if v, ok := datadogScatterplotRequest.GetQOk(); ok {
		terraformRequest["q"] = *v
	} else if v, ok := datadogScatterplotRequest.GetApmQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogScatterplotRequest.GetLogQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogScatterplotRequest.GetProcessQueryOk(); ok {
		terraformQuery := buildTerraformProcessQuery(*v)
		terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogScatterplotRequest.GetRumQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogScatterplotRequest.GetSecurityQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
	}

	if datadogScatterplotRequest.Aggregator != nil {
		terraformRequest["aggregator"] = *datadogScatterplotRequest.Aggregator
	}
	return &terraformRequest
}

//
// ServiceMap Widget Definition helpers
//

func getServiceMapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service": {
			Description: "The ID of the service to map.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"filters": {
			Description: "Your environment and primary tag (or `*` if enabled for your account).",
			Type:        schema.TypeList,
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}
func buildDatadogServiceMapDefinition(terraformDefinition map[string]interface{}) *datadogV1.ServiceMapWidgetDefinition {
	datadogDefinition := datadogV1.NewServiceMapWidgetDefinitionWithDefaults()

	// Required params
	datadogDefinition.SetService(terraformDefinition["service"].(string))
	terraformFilters := terraformDefinition["filters"].([]interface{})
	datadogFilters := make([]string, len(terraformFilters))
	for i, terraformFilter := range terraformFilters {
		datadogFilters[i] = terraformFilter.(string)
	}
	datadogDefinition.SetFilters(datadogFilters)

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
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildTerraformServiceMapDefinition(datadogDefinition *datadogV1.ServiceMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}

	// Required params
	terraformDefinition["service"] = datadogDefinition.GetService()
	terraformDefinition["filters"] = datadogDefinition.GetFilters()

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
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

//
// TopologyMap Widget Definition helpers
//

func getTopologyMapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (`query` and `request_type` are required within the request).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getTopologyRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}

func getTopologyRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request_type": {
			Description:      "The request type for the Topology request ('topology').",
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTopologyRequestTypeFromValue),
		},
		"query": getTopologyQuerySchema(),
	}
}

func getTopologyQuerySchema() *schema.Schema {
	return &schema.Schema{
		Description: "The query for a Topology request.",
		Type:        schema.TypeList,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_source": {
					Description:      "The data source for the Topology request ('service_map' or 'data_streams').",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTopologyQueryDataSourceFromValue),
				},
				"service": {
					Description: "The ID of the service to map.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"filters": {
					Description: "Your environment and primary tag (or `*` if enabled for your account).",
					Type:        schema.TypeList,
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func buildDatadogTopologyMapDefinition(terraformDefinition map[string]interface{}) *datadogV1.TopologyMapWidgetDefinition {
	datadogDefinition := datadogV1.NewTopologyMapWidgetDefinitionWithDefaults()

	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogTopologyRequests(&terraformRequests)
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
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildDatadogTopologyRequests(terraformRequests *[]interface{}) *[]datadogV1.TopologyRequest {
	datadogRequests := make([]datadogV1.TopologyRequest, len(*terraformRequests))
	for i, request := range *terraformRequests {
		if request == nil {
			continue
		}
		terraformRequest := request.(map[string]interface{})
		// Build TopologyRequest
		datadogTopologyRequest := datadogV1.NewTopologyRequest()
		if v, ok := terraformRequest["request_type"].(string); ok && len(v) != 0 {
			datadogTopologyRequest.SetRequestType(datadogV1.TopologyRequestType(v))
		}
		if v, ok := terraformRequest["query"].([]interface{}); ok && len(v) > 0 {
			topologyQuery := v[0].(map[string]interface{})
			datadogTopologyRequest.Query = buildDatadogTopologyQuery(topologyQuery)
		}

		datadogRequests[i] = *datadogTopologyRequest
	}
	return &datadogRequests
}
func buildDatadogTopologyQuery(terraformQuery map[string]interface{}) *datadogV1.TopologyQuery {
	datadogQuery := datadogV1.NewTopologyQuery()
	// Required params
	datadogQuery.SetDataSource(datadogV1.TopologyQueryDataSource(terraformQuery["data_source"].(string)))
	datadogQuery.SetService(terraformQuery["service"].(string))
	terraformFilters := terraformQuery["filters"].([]interface{})
	datadogFilters := make([]string, len(terraformFilters))
	for i, terraformFilter := range terraformFilters {
		datadogFilters[i] = terraformFilter.(string)
	}
	datadogQuery.SetFilters(datadogFilters)

	return datadogQuery
}

func buildTerraformTopologyMapDefinition(datadogDefinition *datadogV1.TopologyMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}

	// Required params
	terraformDefinition["request"] = buildTerraformTopologyRequests(&datadogDefinition.Requests)

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
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}
func buildTerraformTopologyRequests(datadogTopologyRequests *[]datadogV1.TopologyRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogTopologyRequests))
	for i, datadogRequest := range *datadogTopologyRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetRequestTypeOk(); ok {
			terraformRequest["request_type"] = *v
		}
		if v, ok := datadogRequest.GetQueryOk(); ok {
			terraformQuery := buildTerraformTopologyQuery(v)
			terraformRequest["query"] = []map[string]interface{}{terraformQuery}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}
func buildTerraformTopologyQuery(datadogQuery *datadogV1.TopologyQuery) map[string]interface{} {
	terraformQuery := map[string]interface{}{}

	// Required params
	terraformQuery["data_source"] = datadogQuery.GetDataSource()
	terraformQuery["service"] = datadogQuery.GetService()
	terraformQuery["filters"] = datadogQuery.GetFilters()

	return terraformQuery
}

//
// Service Level Objective Widget Definition helpers
//

func getServiceLevelObjectiveDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"view_type": {
			Description: "The type of view to use when displaying the widget. Only `detail` is supported.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"slo_id": {
			Description: "The ID of the service level objective used by the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"show_error_budget": {
			Description: "Whether to show the error budget or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"view_mode": {
			Description:      "The view mode for the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetViewModeFromValue),
			Required:         true,
		},
		"time_windows": {
			Description: "A list of time windows to display in the widget.",
			Type:        schema.TypeList,
			Required:    true,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTimeWindowsFromValue),
			},
		},
		"global_time_target": {
			Description: "The global time target of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"additional_query_filters": {
			Description: "Additional filters applied to the SLO query.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

func buildDatadogServiceLevelObjectiveDefinition(terraformDefinition map[string]interface{}) *datadogV1.SLOWidgetDefinition {
	datadogDefinition := datadogV1.NewSLOWidgetDefinitionWithDefaults()
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
	if v, ok := terraformDefinition["view_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetViewType(v)
	}
	if v, ok := terraformDefinition["slo_id"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSloId(v)
	}
	if v, ok := terraformDefinition["show_error_budget"].(bool); ok {
		datadogDefinition.SetShowErrorBudget(v)
	}
	if v, ok := terraformDefinition["view_mode"].(string); ok && len(v) != 0 {
		datadogDefinition.SetViewMode(datadogV1.WidgetViewMode(v))
	}
	if terraformTimeWindows, ok := terraformDefinition["time_windows"].([]interface{}); ok && len(terraformTimeWindows) > 0 {
		datadogTimeWindows := make([]datadogV1.WidgetTimeWindows, len(terraformTimeWindows))
		for i, timeWindows := range terraformTimeWindows {
			datadogTimeWindows[i] = datadogV1.WidgetTimeWindows(timeWindows.(string))
		}
		datadogDefinition.TimeWindows = datadogTimeWindows
	}
	if v, ok := terraformDefinition["global_time_target"].(string); ok && len(v) != 0 {
		datadogDefinition.SetGlobalTimeTarget(v)
	}
	if v, ok := terraformDefinition["additional_query_filters"].(string); ok && len(v) != 0 {
		datadogDefinition.SetAdditionalQueryFilters(v)
	}
	return datadogDefinition
}

func buildTerraformServiceLevelObjectiveDefinition(datadogDefinition *datadogV1.SLOWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	// Optional params
	if title, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = title
	}
	if titleSize, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = titleSize
	}
	if titleAlign, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = titleAlign
	}
	if viewType, ok := datadogDefinition.GetViewTypeOk(); ok {
		terraformDefinition["view_type"] = viewType
	}
	if datadogDefinition.SloId != nil {
		terraformDefinition["slo_id"] = datadogDefinition.GetSloId()
	}
	if showErrorBudget, ok := datadogDefinition.GetShowErrorBudgetOk(); ok {
		terraformDefinition["show_error_budget"] = showErrorBudget
	}
	if viewMode, ok := datadogDefinition.GetViewModeOk(); ok {
		terraformDefinition["view_mode"] = viewMode
	}
	if datadogDefinition.TimeWindows != nil {
		terraformTimeWindows := make([]string, len(datadogDefinition.GetTimeWindows()))
		for i, datadogTimeWindow := range datadogDefinition.GetTimeWindows() {
			terraformTimeWindows[i] = string(datadogTimeWindow)
		}
		terraformDefinition["time_windows"] = terraformTimeWindows
	}
	if globalTimeTarget, ok := datadogDefinition.GetGlobalTimeTargetOk(); ok {
		terraformDefinition["global_time_target"] = globalTimeTarget
	}

	if AdditionalQueryFilters, ok := datadogDefinition.GetAdditionalQueryFiltersOk(); ok {
		terraformDefinition["additional_query_filters"] = AdditionalQueryFilters
	}
	return terraformDefinition
}

//
// SLO (Service Level Objective) List Widget Definition helpers
//

func getSloListDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: getSloListRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
	}
}

func getSloListRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request_type": {
			Description:      "The request type for the SLO List request.",
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOListWidgetRequestTypeFromValue),
		},
		"query": {
			Description: "Updated SLO List widget.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query_string": {
						Description: "Widget query.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"limit": {
						Description: "Maximum number of results to display in the table.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     100,
					},
					"sort": {
						Description: "The facet and order to sort the data, for example: `{\"column\": \"status.sli\", \"order\": \"desc\"}`.",
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getWidgetFieldSortSchema(),
						},
					},
				},
			},
		},
	}
}

func buildDatadogSloListDefinition(terraformDefinition map[string]interface{}) *datadogV1.SLOListWidgetDefinition {
	datadogDefinition := datadogV1.NewSLOListWidgetDefinitionWithDefaults()
	// Required params
	terraformRequest := terraformDefinition["request"].([]interface{})
	datadogDefinition.SetRequests(*buildDatadogSloListRequests(&terraformRequest))
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
	return datadogDefinition
}

func buildDatadogSloListRequests(terraformRequests *[]interface{}) *[]datadogV1.SLOListWidgetRequest {
	datadogRequests := make([]datadogV1.SLOListWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		datadogSloListRequest := datadogV1.NewSLOListWidgetRequestWithDefaults()

		if v, ok := terraformRequest["request_type"].(string); ok && len(v) != 0 {
			datadogSloListRequest.SetRequestType(datadogV1.SLOListWidgetRequestType(v))
		}
		if terraformQuery, ok := terraformRequest["query"].([]interface{}); ok && len(terraformQuery) > 0 {
			q := terraformQuery[0].(map[string]interface{})
			datadogQuery := datadogV1.NewSLOListWidgetQueryWithDefaults()

			if v, ok := q["query_string"].(string); ok {
				datadogQuery.SetQueryString(v)
			}
			if v, ok := q["limit"].(int); ok {
				datadogQuery.SetLimit(int64(v))
			}

			terraformSort := q["sort"].([]interface{})
			datadogQuery.SetSort(*buildDatadogSloListSort(&terraformSort))

			datadogSloListRequest.SetQuery(*datadogQuery)
		}
		datadogRequests[i] = *datadogSloListRequest
	}
	return &datadogRequests
}

func buildDatadogSloListSort(terraformSorts *[]interface{}) *[]datadogV1.WidgetFieldSort {
	datadogSorts := make([]datadogV1.WidgetFieldSort, len(*terraformSorts))
	for i, s := range *terraformSorts {
		terraformSort := s.(map[string]interface{})

		datadogSort := datadogV1.NewWidgetFieldSortWithDefaults()
		if v, ok := terraformSort["column"].(string); ok {
			datadogSort.SetColumn(v)
		}
		if v, ok := terraformSort["order"].(string); ok {
			order, _ := datadogV1.NewWidgetSortFromValue(v)
			datadogSort.SetOrder(*order)
		}

		datadogSorts[i] = *datadogSort
	}
	return &datadogSorts
}

func buildTerraformSloListDefinition(datadogDefinition *datadogV1.SLOListWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformSloListRequests(&datadogDefinition.Requests)
	// Optional params
	if title, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = title
	}
	if titleSize, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = titleSize
	}
	if titleAlign, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = titleAlign
	}
	return terraformDefinition
}

func buildTerraformSloListRequests(datadogSloListRequests *[]datadogV1.SLOListWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogSloListRequests))
	for i, datadogRequest := range *datadogSloListRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetRequestTypeOk(); ok {
			terraformRequest["request_type"] = *v
		}
		if datadogQuery, ok := datadogRequest.GetQueryOk(); ok {
			terraformQuery := map[string]interface{}{}
			if v, ok := datadogQuery.GetQueryStringOk(); ok {
				terraformQuery["query_string"] = v
			}
			if v, ok := datadogQuery.GetLimitOk(); ok {
				terraformQuery["limit"] = v
			}
			if v, ok := datadogQuery.GetSortOk(); ok {
				terraformQuery["sort"] = buildTerraformSloListSort(v)
			}
			terraformRequest["query"] = []map[string]interface{}{terraformQuery}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildTerraformSloListSort(datadogSorts *[]datadogV1.WidgetFieldSort) *[]map[string]interface{} {
	terraformSorts := make([]map[string]interface{}, len(*datadogSorts))
	for i, datadogSort := range *datadogSorts {
		terraformSort := map[string]interface{}{}
		if v, ok := datadogSort.GetColumnOk(); ok {
			terraformSort["column"] = v
		}
		if v, ok := datadogSort.GetOrderOk(); ok {
			terraformSort["order"] = v
		}
		terraformSorts[i] = terraformSort
	}
	return &terraformSorts
}

//
// List Stream Widget Definition helpers
//

func getListStreamDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "Nested block describing the requests to use when displaying the widget. Multiple `request` blocks are allowed with the structure below.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: getListStreamRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title. Default is 16.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
	}
}

func getListStreamRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"columns": {
			Description: "Widget columns.",
			Type:        schema.TypeList,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"width": {
						Description:      "Widget column width.",
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewListStreamColumnWidthFromValue),
						Required:         true,
					},
					"field": {
						Description: "Widget column field.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"response_format": {
			Description:      "Widget response format.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewListStreamResponseFormatFromValue),
			Required:         true,
		},
		"query": {
			Description: "Updated list stream widget.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"data_source": {
						Description:      "Source from which to query items to display in the stream.",
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewListStreamSourceFromValue),
						Required:         true,
					},
					"query_string": {
						Description: "Widget query.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"event_size": {
						Description:      "Size of events displayed in widget. Required if `data_source` is `event_stream`.",
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetEventSizeFromValue),
					},
					"indexes": {
						Description: "List of indexes.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"storage": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Storage location (private beta).",
					},
					"sort": {
						Description: "The facet and order to sort the data, for example: `{\"column\": \"time\", \"order\": \"desc\"}`.",
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: getWidgetFieldSortSchema(),
						},
					},
				},
			},
		},
	}
}

func buildDatadogListStreamDefinition(terraformDefinition map[string]interface{}) *datadogV1.ListStreamWidgetDefinition {
	datadogDefinition := datadogV1.NewListStreamWidgetDefinitionWithDefaults()
	// Required params
	terraformRequest := terraformDefinition["request"].([]interface{})
	datadogDefinition.SetRequests(*buildDatadogListStreamRequests(&terraformRequest))
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
	return datadogDefinition
}

func buildDatadogListStreamRequests(terraformRequests *[]interface{}) *[]datadogV1.ListStreamWidgetRequest {
	datadogRequests := make([]datadogV1.ListStreamWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build ListStream Request
		datadogListStreamRequest := datadogV1.NewListStreamWidgetRequestWithDefaults()

		datadogQuery := datadogV1.NewListStreamQueryWithDefaults()

		if terraformQuery, ok := terraformRequest["query"].([]interface{}); ok && len(terraformQuery) > 0 {
			q := terraformQuery[0].(map[string]interface{})
			if v, ok := q["data_source"].(string); ok && len(v) > 0 {
				ds := datadogV1.ListStreamSource(v)
				datadogQuery.SetDataSource(ds)
				if v, ok := q["event_size"].(string); ds == datadogV1.LISTSTREAMSOURCE_EVENT_STREAM && ok {
					datadogQuery.SetEventSize(datadogV1.WidgetEventSize(v))
				}
			}
			if v, ok := q["query_string"].(string); ok {
				datadogQuery.SetQueryString(v)
			}
			if v, ok := q["storage"].(string); ok && v != "" {
				datadogQuery.SetStorage(v)
			}
			if v, ok := q["indexes"].([]interface{}); ok {
				var indexes []string
				for _, s := range v {
					indexes = append(indexes, s.(string))
				}
				datadogQuery.SetIndexes(indexes)
			}
			if terraformSort, ok := q["sort"].([]interface{}); ok && len(terraformSort) > 0 {
				sortMap := terraformSort[0].(map[string]interface{})
				datadogSort := datadogV1.NewWidgetFieldSortWithDefaults()
				if v, ok := sortMap["column"].(string); ok {
					datadogSort.SetColumn(v)
				}
				if v, ok := sortMap["order"].(string); ok {
					order, _ := datadogV1.NewWidgetSortFromValue(v)
					datadogSort.SetOrder(*order)
				}
				datadogQuery.SetSort(*datadogSort)
			}
			datadogListStreamRequest.SetQuery(*datadogQuery)

			if v, ok := terraformRequest["response_format"].(string); ok && len(v) != 0 {
				rf := datadogV1.ListStreamResponseFormat(v)
				datadogListStreamRequest.SetResponseFormat(rf)
			}
		}

		terraformColumns := terraformRequest["columns"].([]interface{})
		var datadogColumns []datadogV1.ListStreamColumn
		for _, c := range terraformColumns {
			column := c.(map[string]interface{})
			width := datadogV1.ListStreamColumnWidth(column["width"].(string))
			field := column["field"].(string)
			streamColumn := datadogV1.NewListStreamColumn(field, width)
			datadogColumns = append(datadogColumns, *streamColumn)
		}
		datadogListStreamRequest.SetColumns(datadogColumns)

		datadogRequests[i] = *datadogListStreamRequest
	}
	return &datadogRequests
}

// Geomap Widget Definition helpers
func getGeomapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `log_query` or `rum_query` is required within the `request` block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getGeomapRequestSchema(),
			},
		},
		"style": {
			Description: "The style of the widget graph. One nested block is allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"palette": {
						Description: "The color palette to apply to the widget.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"palette_flip": {
						Description: "A Boolean indicating whether to flip the palette tones.",
						Type:        schema.TypeBool,
						Required:    true,
					},
				},
			},
		},
		"view": {
			Description: "The view of the world that the map should render.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"focus": {
						Description: "The two-letter ISO code of a country to focus the map on (or `WORLD`).",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}

func getGeomapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":         getMetricQuerySchema(),
		"log_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"rum_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
	}
}

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

	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
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
			terraformRequest["formula"] = buildTerraformFormula(v)
		}

		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Timeseries Widget Definition helpers
//

func getTimeseriesDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `network_query`, `security_query` or `process_query` is required within the `request` block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getTimeseriesRequestSchema(),
			},
		},
		"marker": {
			Description: "A nested block describing the marker to use when displaying the widget. The structure of this block is described below. Multiple `marker` blocks are allowed within a given `tile_def` block.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetMarkerSchema(),
			},
		},
		"event": {
			Description: "The definition of the event to overlay on the graph. Multiple `event` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetEventSchema(),
			},
		},
		"yaxis": {
			Description: "A nested block describing the Y-Axis Controls. The structure of this block is described below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"right_yaxis": {
			Description: "A nested block describing the right Y-Axis Controls. See the `on_right_yaxis` property for which request will use this axis. The structure of this block is described below.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"show_legend": {
			Description: "Whether or not to show the legend on this widget.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"legend_size": {
			Description:  "The size of the legend displayed in the widget.",
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateTimeseriesWidgetLegendSize,
		},
		"legend_layout": {
			Description:      "The layout of the legend displayed in the widget.",
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTimeseriesWidgetLegendLayoutFromValue),
		},
		"legend_columns": {
			Description: "A list of columns to display in the legend.",
			Type:        schema.TypeSet,
			Optional:    true,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTimeseriesWidgetLegendColumnFromValue),
			},
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
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

//
// Sunburst Widget Definition helpers
//

func getSunburstDefinitionschema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `log_query` or `rum_query` is required within the `request` block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getSunburstRequestSchema(),
			},
		},
		"hide_total": {
			Description: "Whether or not to show the total value in the widget.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"legend_inline": {
			Description: "Used to configure the inline legend. Cannot be used in conjunction with legend_table.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description:      "The type of legend (inline or automatic).",
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSunburstWidgetLegendInlineAutomaticTypeFromValue),
						Required:         true,
					},
					"hide_value": {
						Description: "Whether to hide the values of the groups.",
						Type:        schema.TypeBool,
						Optional:    true,
					},
					"hide_percent": {
						Description: "Whether to hide the percentages of the groups.",
						Type:        schema.TypeBool,
						Optional:    true,
					},
				},
			},
		},
		"legend_table": {
			Description: "Used to configure the table legend. Cannot be used in conjunction with legend_inline.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description:      "The type of legend (table or none).",
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSunburstWidgetLegendTableTypeFromValue),
						Required:         true,
					},
				},
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title. Default is 16.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title. One of `left`, `center`, or `right`.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "Nested block describing a custom link. Multiple `custom_link` blocks are allowed with the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}

func getSunburstRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"network_query":  getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"audit_query":    getApmLogNetworkRumSecurityAuditQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to Sunburst requests
		"style": {
			Description: "Define style for the widget's request.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
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

	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
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
			terraformRequest["formula"] = buildTerraformFormula(v)
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}

	return terraformDefinition
}

func getScatterplotFormulaSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"formula_expression": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A string expression built from queries, formulas, and functions.",
				},
				"dimension": {
					Description:      "Dimension of the Scatterplot.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewScatterplotDimensionFromValue),
				},
				"alias": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "An expression alias.",
				},
			},
		},
	}
}

func getFormulaSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"formula_expression": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "A string expression built from queries, formulas, and functions.",
				},
				"limit": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The options for limiting results returned.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"count": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "The number of results to return.",
							},
							"order": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
								Default:          "desc",
								Description:      "The direction of the sort.",
							},
						},
					},
				},
				"alias": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "An expression alias.",
				},
				"conditional_formats": {
					Description: "Conditional formats allow you to set the color of your widget content or background depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: getWidgetConditionalFormatSchema(),
					},
				},
				"cell_display_mode": {
					Description:      "A list of display modes for each table cell.",
					Type:             schema.TypeString,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTableWidgetCellDisplayModeFromValue),
					Optional:         true,
				},
				"style": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Styling options for widget formulas.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"palette": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The color palette used to display the formula. A guide to the available color palettes can be found at https://docs.datadoghq.com/dashboards/guide/widget_colors.",
							},
							"palette_index": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "Index specifying which color to use within the palette.",
							},
						},
					},
				},
			},
		},
	}
}

func getFormulaQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "A timeseries formula and functions metrics query.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "metrics",
								Description: "The data source for metrics queries.",
							},
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The metrics query definition.",
							},
							"aggregator": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricAggregationFromValue),
								Description:      "The aggregation methods available for metrics queries.",
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the query for use in formulas.",
							},
						},
					},
				},
				"event_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "A timeseries formula and functions events query.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventsDataSourceFromValue),
								Description:      "The data source for event platform-based queries.",
							},
							"storage": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Storage location (private beta).",
							},
							"search": {
								Type:        schema.TypeList,
								Optional:    true,
								MaxItems:    1,
								Description: "The search options.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"query": {
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
											Required:     true,
											Description:  "The events search string.",
										},
									},
								},
							},
							"indexes": {
								Type:        schema.TypeList,
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "An array of index names to query in the stream.",
							},
							"compute": {
								Type:        schema.TypeList,
								Required:    true,
								Description: "The compute options.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"aggregation": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventAggregationFromValue),
											Description:      "The aggregation methods for event platform queries.",
										},
										"interval": {
											Type:        schema.TypeInt,
											Optional:    true,
											Description: "A time interval in milliseconds.",
										},
										"metric": {
											Type:        schema.TypeString,
											Optional:    true,
											Description: "The measurable attribute to compute.",
										},
									},
								},
							},
							"group_by": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Group by options.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"facet": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "The event facet.",
										},
										"limit": {
											Type:        schema.TypeInt,
											Optional:    true,
											Description: "The number of groups to return.",
										},
										"sort": {
											Type:        schema.TypeList,
											Optional:    true,
											MaxItems:    1,
											Description: "The options for sorting group by results.",
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"aggregation": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionEventAggregationFromValue),
														Description:      "The aggregation methods for the event platform queries.",
													},
													"metric": {
														Type:        schema.TypeString,
														Optional:    true,
														Description: "The metric used for sorting group by results.",
													},
													"order": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
														Description:      "Direction of sort.",
													},
												},
											},
										},
									},
								},
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of query for use in formulas.",
							},
						},
					},
				},
				"process_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "The process query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionProcessQueryDataSourceFromValue),
								Description:      "The data source for process queries.",
							},
							"metric": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The process metric name.",
							},
							"text_filter": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The text to use as a filter.",
							},
							"tag_filters": {
								Type:        schema.TypeList,
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "An array of tags to filter by.",
							},
							"limit": {
								Type:        schema.TypeInt,
								Optional:    true,
								Description: "The number of hits to return.",
							},
							"sort": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
								Description:      "The direction of the sort.",
								Default:          "desc",
							},
							"aggregator": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionMetricAggregationFromValue),
								Description:      "The aggregation methods available for metrics queries.",
							},
							"is_normalized_cpu": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Whether to normalize the CPU percentages.",
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of query for use in formulas.",
							},
						},
					},
				},
				"apm_dependency_stats_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "The APM Dependency Stats query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionApmDependencyStatsDataSourceFromValue),
								Description:      "The data source for APM Dependency Stats queries.",
							},
							"env": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "APM environment.",
							},
							"stat": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionApmDependencyStatNameFromValue),
								Description:      "APM statistic.",
							},
							"operation_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of operation on service.",
							},
							"resource_name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "APM resource.",
							},
							"service": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "APM service.",
							},
							"primary_tag_name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The name of the second primary tag used within APM; required when `primary_tag_value` is specified. See https://docs.datadoghq.com/tracing/guide/setting_primary_tags_to_scope/#add-a-second-primary-tag-in-datadog.",
							},
							"primary_tag_value": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Filter APM data by the second primary tag. `primary_tag_name` must also be specified.",
							},
							"is_upstream": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Determines whether stats for upstream or downstream dependencies should be queried.",
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of query for use in formulas.",
							},
						},
					},
				},
				"apm_resource_stats_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "The APM Resource Stats query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionApmResourceStatsDataSourceFromValue),
								Description:      "The data source for APM Resource Stats queries.",
							},
							"env": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "APM environment.",
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of query for use in formulas.",
							},
							"stat": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionApmResourceStatNameFromValue),
								Description:      "APM statistic.",
							},
							"operation_name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Name of operation on service.",
							},
							"resource_name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "APM resource.",
							},
							"service": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "APM service.",
							},
							"primary_tag_name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The name of the second primary tag used within APM; required when `primary_tag_value` is specified. See https://docs.datadoghq.com/tracing/guide/setting_primary_tags_to_scope/#add-a-second-primary-tag-in-datadog.",
							},
							"primary_tag_value": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Filter APM data by the second primary tag. `primary_tag_name` must also be specified.",
							},
							"group_by": {
								Type:        schema.TypeList,
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "Array of fields to group results by.",
							},
						},
					},
				},
				"slo_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "The SLO query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionSLODataSourceFromValue),
								Description:      "The data source for SLO queries.",
							},
							"slo_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "ID of an SLO to query.",
							},
							"measure": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionSLOMeasureFromValue),
								Description:      "SLO measures queries.",
							},
							"name": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The name of query for use in formulas.",
							},
							"group_mode": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          "overall",
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionSLOGroupModeFromValue),
								Description:      "Group mode to query measures.",
							},
							"slo_query_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          "metric",
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewFormulaAndFunctionSLOQueryTypeFromValue),
								Description:      "type of the SLO to query.",
							},
							"additional_query_filters": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Additional filters applied to the SLO query.",
							},
						},
					},
				},
			},
		},
	}
}

func getTimeseriesRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"network_query":  getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"audit_query":    getApmLogNetworkRumSecurityAuditQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to Timeseries requests
		"style": {
			Description: "The style of the widget graph. Exactly one `style` block is allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"palette": {
						Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"line_type": {
						Description:      "The type of lines displayed.",
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLineTypeFromValue),
						Optional:         true,
					},
					"line_width": {
						Description:      "The width of line displayed.",
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLineWidthFromValue),
						Optional:         true,
					},
				},
			},
		},
		"metadata": {
			Description: "Used to define expression aliases. Multiple `metadata` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"expression": {
						Description: "The expression name.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"alias_name": {
						Description: "The expression alias.",
						Type:        schema.TypeString,
						Optional:    true,
					},
				},
			},
		},
		"display_type": {
			Description:      "How to display the marker lines.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetDisplayTypeFromValue),
			Optional:         true,
		},
		"on_right_yaxis": {
			Description: "A Boolean indicating whether the request uses the right or left Y-Axis.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}
}

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

	definition := datadogV1.FormulaAndFunctionMetricQueryDefinitionAsFormulaAndFunctionQueryDefinition(metricQuery)
	return &definition
}

func buildDatadogFormulaAndFunctionAPMResourceStatsQuery(data map[string]interface{}) *datadogV1.FormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.FormulaAndFunctionApmResourceStatsDataSource(data["data_source"].(string))
	stat := datadogV1.FormulaAndFunctionApmResourceStatName(data["stat"].(string))
	apmResourceStatsQuery := datadogV1.NewFormulaAndFunctionApmResourceStatsQueryDefinition(dataSource, data["env"].(string), data["name"].(string), data["service"].(string), stat)

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
			terraformRequest["formula"] = buildTerraformFormula(v)
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

//
// Toplist Widget Definition helpers
//

func getToplistDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block).",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getToplistRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}

func getToplistRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"log_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmLogNetworkRumSecurityAuditQuerySchema(),
		"security_query": getApmLogNetworkRumSecurityAuditQuerySchema(),
		"audit_query":    getApmLogNetworkRumSecurityAuditQuerySchema(),
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
		// Settings specific to Toplist requests
		"conditional_formats": {
			Description: "Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetConditionalFormatSchema(),
			},
		},
		"style": {
			Description: "Define request for the widget's style.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
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
			terraformRequest["formula"] = buildTerraformFormula(v)
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

//
// Trace Service Widget Definition helpers
//

func getTraceServiceDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"env": {
			Description: "APM environment.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"service": {
			Description: "APM service.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"span_name": {
			Description: "APM span name",
			Type:        schema.TypeString,
			Required:    true,
		},
		"show_hits": {
			Description: "Whether to show the hits metrics or not",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_errors": {
			Description: "Whether to show the error metrics or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_latency": {
			Description: "Whether to show the latency metrics or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_breakdown": {
			Description: "Whether to show the latency breakdown or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_distribution": {
			Description: "Whether to show the latency distribution or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"show_resource_list": {
			Description: "Whether to show the resource list or not.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"size_format": {
			Description:      "The size of the widget.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSizeFormatFromValue),
			Optional:         true,
		},
		"display_format": {
			Description:      "The number of columns to display.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetServiceSummaryDisplayFormatFromValue),
			Optional:         true,
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
	}
}

func buildDatadogTraceServiceDefinition(terraformDefinition map[string]interface{}) *datadogV1.ServiceSummaryWidgetDefinition {
	datadogDefinition := datadogV1.NewServiceSummaryWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetEnv(terraformDefinition["env"].(string))
	datadogDefinition.SetService(terraformDefinition["service"].(string))
	datadogDefinition.SetSpanName(terraformDefinition["span_name"].(string))
	// Optional params
	if v, ok := terraformDefinition["show_hits"].(bool); ok {
		datadogDefinition.SetShowHits(v)
	}
	if v, ok := terraformDefinition["show_errors"].(bool); ok {
		datadogDefinition.SetShowErrors(v)
	}
	if v, ok := terraformDefinition["show_latency"].(bool); ok {
		datadogDefinition.SetShowLatency(v)
	}
	if v, ok := terraformDefinition["show_breakdown"].(bool); ok {
		datadogDefinition.SetShowBreakdown(v)
	}
	if v, ok := terraformDefinition["show_distribution"].(bool); ok {
		datadogDefinition.SetShowDistribution(v)
	}
	if v, ok := terraformDefinition["show_resource_list"].(bool); ok {
		datadogDefinition.SetShowResourceList(v)
	}
	if v, ok := terraformDefinition["size_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSizeFormat(datadogV1.WidgetSizeFormat(v))
	}
	if v, ok := terraformDefinition["display_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetDisplayFormat(datadogV1.WidgetServiceSummaryDisplayFormat(v))
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
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}

	return datadogDefinition
}

//
// Run workflow Widget Definition helpers
//

func getRunWorkflowDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"workflow_id": {
			Description: "Workflow ID",
			Type:        schema.TypeString,
			Required:    true,
		},
		"input": {
			Description: "Array of workflow inputs to map to dashboard template variables.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Name of the workflow input.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"value": {
						Description: "Dashboard template variable. Can be suffixed with `.value` or `.key`.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_size": {
			Description: "The size of the widget's title (defaults to 16).",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"title_align": {
			Description:      "The alignment of the widget's title.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetTextAlignFromValue),
			Optional:         true,
		},
		"live_span": getWidgetLiveSpanSchema(),
		"custom_link": {
			Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getWidgetCustomLinkSchema(),
			},
		},
	}
}

func buildDatadogRunWorkflowDefinition(terraformDefinition map[string]interface{}) *datadogV1.RunWorkflowWidgetDefinition {
	datadogDefinition := datadogV1.NewRunWorkflowWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetWorkflowId(terraformDefinition["workflow_id"].(string))
	// Optional params
	if inputs, ok := terraformDefinition["input"].([]interface{}); ok && len(inputs) != 0 {
		datadogDefinition.Inputs = *buildDatadogRunWorkflowInputs(&inputs)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.PtrString(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.PtrString(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if ls, ok := terraformDefinition["live_span"].(string); ok && ls != "" {
		datadogDefinition.Time = &datadogV1.WidgetTime{
			LiveSpan: datadogV1.WidgetLiveSpan(ls).Ptr(),
		}
	}
	if v, ok := terraformDefinition["custom_link"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetCustomLinks(*buildDatadogWidgetCustomLinks(&v))
	}
	return datadogDefinition
}
func buildDatadogRunWorkflowInputs(terraformInputs *[]interface{}) *[]datadogV1.RunWorkflowWidgetInput {
	datadogRunWorkflowInputs := make([]datadogV1.RunWorkflowWidgetInput, len(*terraformInputs))
	for i, terraformInput := range *terraformInputs {
		terraformRunWorkflowInput := terraformInput.(map[string]interface{})
		if terraformRunWorkflowInput == nil {
			continue
		}
		datadogInput := datadogV1.NewRunWorkflowWidgetInput(terraformRunWorkflowInput["name"].(string), terraformRunWorkflowInput["value"].(string))
		datadogRunWorkflowInputs[i] = *datadogInput
	}
	return &datadogRunWorkflowInputs
}

func buildTerraformRunWorkflowDefinition(datadogDefinition *datadogV1.RunWorkflowWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["workflow_id"] = datadogDefinition.GetWorkflowId()
	// Optional params
	if v, ok := datadogDefinition.GetInputsOk(); ok {
		terraformDefinition["input"] = buildTerraformRunWorkflowInputs(v)
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	if v, ok := datadogDefinition.GetCustomLinksOk(); ok {
		terraformDefinition["custom_link"] = buildTerraformWidgetCustomLinks(v)
	}
	return terraformDefinition
}
func buildTerraformRunWorkflowInputs(datadogInputs *[]datadogV1.RunWorkflowWidgetInput) *[]map[string]interface{} {
	terraformRunWorkflowInputs := make([]map[string]interface{}, len(*datadogInputs))
	for i, input := range *datadogInputs {
		terraformInput := map[string]interface{}{}
		// Required params
		terraformInput["name"] = input.GetName()
		terraformInput["value"] = input.GetValue()
		terraformRunWorkflowInputs[i] = terraformInput
	}
	return &terraformRunWorkflowInputs
}

//
// Treemap Widget Definition Helpers
//

func getTreemapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Description: "Nested block describing the request to use when displaying the widget.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: getTreemapRequestSchema(),
			},
		},
		"title": {
			Description: "The title of the widget.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

func getTreemapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// "query" and "formula" go together
		"query":   getFormulaQuerySchema(),
		"formula": getFormulaSchema(),
	}
}

func buildDatadogTreemapDefinition(terraformDefinition map[string]interface{}) *datadogV1.TreeMapWidgetDefinition {
	datadogDefinition := datadogV1.NewTreeMapWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogTreemapRequests(&terraformRequests)

	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
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
			terraformRequest["formula"] = buildTerraformFormula(v)
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
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}

	return terraformDefinition
}

func buildTerraformListStreamDefinition(datadogDefinition *datadogV1.ListStreamWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformListStreamWidgetRequests(&datadogDefinition.Requests)
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
	return terraformDefinition
}

func buildTerraformListStreamWidgetRequests(datadogListStreamRequests *[]datadogV1.ListStreamWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogListStreamRequests))
	for i, d := range *datadogListStreamRequests {
		terraformRequest := map[string]interface{}{}

		terraformRequest["response_format"] = string(d.GetResponseFormat())

		terraformColumns := make([]interface{}, len(d.GetColumns()))
		for i, c := range d.GetColumns() {
			column := map[string]interface{}{
				"field": c.GetField(),
				"width": string(c.GetWidth()),
			}
			terraformColumns[i] = column
		}
		terraformRequest["columns"] = terraformColumns

		q := d.GetQuery()
		queryRequest := map[string]interface{}{}
		if indexes, ok := q.GetIndexesOk(); ok && indexes != nil && len(*indexes) > 0 {
			queryRequest["indexes"] = *indexes
		}
		if queryString, ok := q.GetQueryStringOk(); ok {
			queryRequest["query_string"] = queryString
		}
		if storage, ok := q.GetStorageOk(); ok {
			queryRequest["storage"] = storage
		}
		queryRequest["data_source"] = string(q.GetDataSource())
		if eventSize, ok := q.GetEventSizeOk(); ok && q.GetDataSource() == datadogV1.LISTSTREAMSOURCE_EVENT_STREAM {
			queryRequest["event_size"] = eventSize
		}
		if v, ok := q.GetSortOk(); ok {
			terraformSort := map[string]interface{}{}
			if v, ok := v.GetColumnOk(); ok {
				terraformSort["column"] = v
			}
			if v, ok := v.GetOrderOk(); ok {
				terraformSort["order"] = v
			}
			queryRequest["sort"] = []interface{}{terraformSort}
		}
		terraformRequest["query"] = []map[string]interface{}{queryRequest}

		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

func buildTerraformTraceServiceDefinition(datadogDefinition *datadogV1.ServiceSummaryWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["env"] = datadogDefinition.GetEnv()
	terraformDefinition["service"] = datadogDefinition.GetService()
	terraformDefinition["span_name"] = datadogDefinition.GetSpanName()
	// Optional params
	if v, ok := datadogDefinition.GetShowHitsOk(); ok {
		terraformDefinition["show_hits"] = v
	}
	if v, ok := datadogDefinition.GetShowErrorsOk(); ok {
		terraformDefinition["show_errors"] = v
	}
	if v, ok := datadogDefinition.GetShowLatencyOk(); ok {
		terraformDefinition["show_latency"] = v
	}
	if v, ok := datadogDefinition.GetShowBreakdownOk(); ok {
		terraformDefinition["show_breakdown"] = v
	}
	if v, ok := datadogDefinition.GetShowDistributionOk(); ok {
		terraformDefinition["show_distribution"] = v
	}
	if v, ok := datadogDefinition.GetShowResourceListOk(); ok {
		terraformDefinition["show_resource_list"] = v
	}
	if v, ok := datadogDefinition.GetSizeFormatOk(); ok {
		terraformDefinition["size_format"] = v
	}
	if v, ok := datadogDefinition.GetDisplayFormatOk(); ok {
		terraformDefinition["display_format"] = v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["live_span"] = v.GetLiveSpan()
	}
	return terraformDefinition
}

// Widget Conditional Format helpers
func getWidgetConditionalFormatSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"comparator": {
			Description:      "The comparator to use.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetComparatorFromValue),
			Required:         true,
		},
		"value": {
			Description: "A value for the comparator.",
			Type:        schema.TypeFloat,
			Required:    true,
		},
		"palette": {
			Description:      "The color palette to apply.",
			Type:             schema.TypeString,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetPaletteFromValue),
			Required:         true,
		},
		"custom_bg_color": {
			Description: "The color palette to apply to the background, same values available as palette.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"custom_fg_color": {
			Description: "The color palette to apply to the foreground, same values available as palette.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"image_url": {
			Description: "Displays an image as the background.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"hide_value": {
			Description: "Setting this to True hides values.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"timeframe": {
			Description: "Defines the displayed timeframe.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"metric": {
			Description: "The metric from the request to correlate with this conditional format.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}
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

func getWidgetCustomLinkSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"label": {
			Description: "The label for the custom link URL.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"link": {
			Description: "The URL of the custom link.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"is_hidden": {
			Description: "The flag for toggling context menu link visibility.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
		"override_label": {
			Description: "The label ID that refers to a context menu link item. When `override_label` is provided, the client request omits the label field.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}

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

func getWidgetEventSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"q": {
			Description: "The event query to use in the widget.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"tags_execution": {
			Description: "The execution method for multi-value filters.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}
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
func getWidgetLiveSpanSchema() *schema.Schema {
	return &schema.Schema{
		Description:      "The timeframe to use when displaying the widget.",
		Type:             schema.TypeString,
		ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetLiveSpanFromValue),
		Optional:         true,
	}
}

// Widget Marker helpers
func getWidgetMarkerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"value": {
			Description: "A mathematical expression describing the marker, for example: `y > 1`, `-5 < y < 0`, `y = 19`.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"display_type": {
			Description: "How the marker lines are displayed, options are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"label": {
			Description: "A label for the line or range.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}
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
func getMetricQuerySchema() *schema.Schema {
	return &schema.Schema{
		Description: "The metric query to use for this widget.",
		Type:        schema.TypeString,
		Optional:    true,
	}
}

// APM, Log, Network, RUM or Audit Query
func getApmLogNetworkRumSecurityAuditQuerySchema() *schema.Schema {
	return &schema.Schema{
		Description: "The query to use for this widget.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"index": {
					Description: "The name of the index to query.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"compute_query": {
					Description: "`compute_query` or `multi_compute` is required. The map keys are listed below.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem:        getComputeSchema(),
				},
				"multi_compute": {
					Description: "`compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed using the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        getComputeSchema(),
				},
				"search_query": {
					Description: "The search query to use.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"group_by": {
					Description: "Multiple `group_by` blocks are allowed using the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"facet": {
								Description: "The facet name.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"limit": {
								Description: "The maximum number of items in the group.",
								Type:        schema.TypeInt,
								Optional:    true,
							},
							"sort_query": {
								Description: "A list of exactly one element describing the sort query to use.",
								Type:        schema.TypeList,
								MaxItems:    1,
								Optional:    true,
								Elem:        getQueryGroupBySortSchema(),
							},
						},
					},
				},
			},
		},
	}
}

func getQueryGroupBySortSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"aggregation": {
				Description: "The aggregation method.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"order": {
				Description:      "Widget sorting methods.",
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
				Required:         true,
			},
			"facet": {
				Description: "The facet name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func getComputeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"aggregation": {
				Description: "The aggregation method.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"facet": {
				Description: "The facet name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"interval": {
				Description: "Define the time interval in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
		},
	}
}

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
			groupBy := g.(map[string]interface{})
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
			if metricQuery, ok := terraformMetricQueryDefinition.GetQueryOk(); ok {
				terraformQuery["query"] = metricQuery
			}
			if aggregator, ok := terraformMetricQueryDefinition.GetAggregatorOk(); ok {
				terraformQuery["aggregator"] = aggregator
			}
			if name, ok := terraformMetricQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
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

func buildTerraformFormula(datadogFormulas *[]datadogV1.WidgetFormula) []map[string]interface{} {
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
func getProcessQuerySchema() *schema.Schema {
	return &schema.Schema{
		Description: "The process query to use in the widget. The structure of this block is described below.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric": {
					Description: "Your chosen metric.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"search_by": {
					Description: "Your chosen search term.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"filter_by": {
					Description: "A list of processes.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"limit": {
					Description: "The max number of items in the filter list.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
			},
		},
	}
}
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
func getApmStatsQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"service": {
					Description: "The service name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"name": {
					Description: "The operation name associated with the service.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"env": {
					Description: "The environment name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"primary_tag": {
					Description: "The organization's host group name and value.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"row_type": {
					Description:      "The level of detail for the request.",
					Type:             schema.TypeString,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewApmStatsQueryRowTypeFromValue),
					Required:         true,
				},
				"resource": {
					Description: "The resource name.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"columns": {
					Description: "Column properties used by the front end for display.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Description: "The column name.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"alias": {
								Description: "A user-assigned alias for the column.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"order": {
								Description:      "Widget sorting methods.",
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWidgetSortFromValue),
								Optional:         true,
							},
							"cell_display_mode": {
								Description:      "A list of display modes for each table cell.",
								Type:             schema.TypeString,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewTableWidgetCellDisplayModeFromValue),
								Optional:         true,
							},
						},
					},
				},
			},
		},
	}
}

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

func getWidgetAxisSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"label": {
			Description: "The label of the axis to display on the graph.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"scale": {
			Description: "Specify the scale type, options: `linear`, `log`, `pow`, `sqrt`.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"min": {
			Description: "Specify the minimum value to show on the Y-axis.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"max": {
			Description: "Specify the maximum value to show on the Y-axis.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"include_zero": {
			Description: "Always include zero or fit the axis to the data range.",
			Type:        schema.TypeBool,
			Optional:    true,
		},
	}
}

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

// Widget Style helpers

func getWidgetRequestStyle() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"palette": {
			Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.",
			Type:        schema.TypeString,
			Optional:    true,
		},
	}
}
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

func buildDatadogHostmapRequestStyle(terraformStyle map[string]interface{}) *datadogV1.HostMapWidgetDefinitionStyle {
	datadogStyle := &datadogV1.HostMapWidgetDefinitionStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}
	if v, ok := terraformStyle["palette_flip"].(bool); ok {
		datadogStyle.SetPaletteFlip(v)
	}
	if v, ok := terraformStyle["fill_min"].(string); ok && len(v) != 0 {
		datadogStyle.SetFillMin(v)
	}
	if v, ok := terraformStyle["fill_max"].(string); ok && len(v) != 0 {
		datadogStyle.SetFillMax(v)
	}

	return datadogStyle
}
func buildTerraformHostmapRequestStyle(datadogStyle *datadogV1.HostMapWidgetDefinitionStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if datadogStyle.Palette != nil {
		terraformStyle["palette"] = datadogStyle.GetPalette()
	}
	if datadogStyle.PaletteFlip != nil {
		terraformStyle["palette_flip"] = datadogStyle.GetPaletteFlip()
	}
	if datadogStyle.FillMin != nil {
		terraformStyle["fill_min"] = datadogStyle.GetFillMin()
	}
	if datadogStyle.FillMax != nil {
		terraformStyle["fill_max"] = datadogStyle.GetFillMax()
	}
	return terraformStyle
}

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
