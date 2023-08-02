package datadog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var computedFields = []string{"id", "author_handle", "author_name", "created_at", "modified_at", "url"}

const path = "/api/v1/dashboard"

func resourceDatadogDashboardJSON() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog dashboard JSON resource. This can be used to create and manage Datadog dashboards using the JSON definition.",
		CreateContext: resourceDatadogDashboardJSONCreate,
		ReadContext:   resourceDatadogDashboardJSONRead,
		UpdateContext: resourceDatadogDashboardJSONUpdate,
		DeleteContext: resourceDatadogDashboardJSONDelete,
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
				"dashboard": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsJSON,
					StateFunc: func(v interface{}) string {
						attrMap, _ := structure.ExpandJsonFromString(v.(string))
						prepResource(attrMap)
						res, _ := structure.FlattenJsonToString(attrMap)
						return res
					},
					Description: "The JSON formatted definition of the Dashboard.",
				},
				"url": {
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
					Description: "The URL of the dashboard.",
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
					Description: "The list of dashboard lists this dashboard should be removed from. Internal only.",
					Elem:        &schema.Schema{Type: schema.TypeInt},
				},
			}
		},
	}
}

func deleteWidgetID(widgets []interface{}) {
	for _, w := range widgets {
		if widget, ok := w.(map[string]interface{}); ok {
			if def, ok := widget["definition"].(map[string]interface{}); ok {
				if def["type"] == "group" {
					if group, ok := def["widgets"].([]interface{}); ok {
						deleteWidgetID(group)
					}
				}
				delete(widget, "id")
			}
		}
	}
}

func resourceDatadogDashboardJSONRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	respByte, httpResp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", path+"/"+id, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	dashboard := d.Get("dashboard").(string)

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", path, &dashboard)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating resource")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	id, ok := respMap["id"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving id from response"))
	}
	d.SetId(id.(string))

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	var httpResponse *http.Response
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		_, httpResponse, err = utils.SendRequest(auth, apiInstances.HttpClient, "GET", path+"/"+id.(string), nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("dashboard not created yet"))
			}

			return retry.NonRetryableError(err)
		}

		// We only log the error, as failing to update the list shouldn't fail dashboard creation
		// Method imported from dashboard resource
		updateDashboardLists(d, providerConf, id.(string), layoutType.(string))

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	dashboard := d.Get("dashboard").(string)
	id := d.Id()

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "PUT", path+"/"+id, &dashboard)
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

	// Method imported from dashboard resource
	updateDashboardLists(d, providerConf, id, layoutType.(string))

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	_, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", path+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard")
	}

	return nil
}

func updateDashboardJSONState(d *schema.ResourceData, dashboard map[string]interface{}) diag.Diagnostics {
	if v, ok := dashboard["url"]; ok {
		if err := d.Set("url", v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	prepResource(dashboard)

	dashboardString, err := structure.FlattenJsonToString(dashboard)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("dashboard", dashboardString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func prepResource(attrMap map[string]interface{}) map[string]interface{} {
	// This is an edge case where refresh might be called with an empty definition.
	if attrMap == nil {
		return attrMap
	}

	// Remove computed fields when comparing diffs
	for _, f := range computedFields {
		delete(attrMap, f)
	}
	// Remove every widget id too
	if widgets, ok := attrMap["widgets"].([]interface{}); ok {
		deleteWidgetID(widgets)
	}
	// 'restricted_roles' takes precedence over 'is_read_only'
	if _, ok := attrMap["restricted_roles"].([]interface{}); ok {
		delete(attrMap, "is_read_only")
	} else {
		// `is_read_only` defaults to false.
		// We set it manually to avoid continous diff when not set.
		if _, ok := attrMap["is_read_only"]; !ok {
			attrMap["is_read_only"] = false
		}
	}
	// handle `notify_list` order
	if notifyList, ok := attrMap["notify_list"].([]interface{}); ok {
		sort.SliceStable(notifyList, func(i, j int) bool {
			return notifyList[i].(string) < notifyList[j].(string)
		})
	}

	return attrMap
}
