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
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ordered", "free"}, false)),
				},
				"reflow_type": {
					Type:             schema.TypeString,
					Optional:         true,
					Description:      "The reflow type of a new dashboard layout. Set this only when layout type is `ordered`. If set to `fixed`, the dashboard expects all widgets to have a layout, and if it's set to `auto`, widgets should not have layouts.",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"auto", "fixed"}, false)),
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

func updateDashboardLists(d *schema.ResourceData, providerConf *ProviderConfiguration, dashboardID string, layoutType string) {
	dashTypeString := "custom_screenboard"
	if layoutType == "ordered" {
		dashTypeString = "custom_timeboard"
	}

	type dashboardListItem struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}
	type dashboardListRequest struct {
		Dashboards []dashboardListItem `json:"dashboards"`
	}

	requestBody := dashboardListRequest{
		Dashboards: []dashboardListItem{
			{Type: dashTypeString, ID: dashboardID},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[DEBUG] Got error marshaling dashboard list request: %v", err)
		return
	}
	bodyStr := string(bodyBytes)

	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if v, ok := d.GetOk("dashboard_lists"); ok && v.(*schema.Set).Len() > 0 {
		for _, id := range v.(*schema.Set).List() {
			path := fmt.Sprintf("/api/v2/dashboard/lists/manual/%d/dashboards", id.(int))
			_, _, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", path, &bodyStr)
			if err != nil {
				log.Printf("[DEBUG] Got error adding to dashboard list %d: %v", id.(int), err)
			}
		}
	}

	if v, ok := d.GetOk("dashboard_lists_removed"); ok && v.(*schema.Set).Len() > 0 {
		for _, id := range v.(*schema.Set).List() {
			path := fmt.Sprintf("/api/v2/dashboard/lists/manual/%d/dashboards", id.(int))
			_, _, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", path, &bodyStr)
			if err != nil {
				log.Printf("[DEBUG] Got error removing from dashboard list %d: %v", id.(int), err)
			}
		}
	}
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
	s := dashboardmapping.AllWidgetSchemasMap(false)
	// Inject recursive group widget sub-schema
	groupSchema := s["group_definition"]
	if groupSchema != nil {
		groupSchema.Elem.(*schema.Resource).Schema["widget"] = &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of widgets in this group.",
			Elem: &schema.Resource{
				Schema: dashboardmapping.AllWidgetSchemasMap(false),
			},
		}
	}
	return s
}
