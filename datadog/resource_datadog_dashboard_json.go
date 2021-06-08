package datadog

import (
	"context"
	"errors"
	"log"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		Schema: map[string]*schema.Schema{
			"dashboard": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					// Remove computed fields when comparing diffs
					attrMap, _ := structure.ExpandJsonFromString(v.(string))
					for _, f := range computedFields {
						delete(attrMap, f)
					}
					// Remove every widget id too
					deleteWidgetID(attrMap["widgets"].([]interface{}))
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
				Description: "The list of dashboard lists this dashboard belongs to.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"dashboard_lists_removed": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The list of dashboard lists this dashboard should be removed from. Internal only.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func deleteWidgetID(widgets []interface{}) {
	for _, w := range widgets {
		widget := w.(map[string]interface{})
		def := widget["definition"].(map[string]interface{})
		if def["type"] == "group" {
			deleteWidgetID(def["widgets"].([]interface{}))
		}
		delete(widget, "id")
	}
}

func resourceDatadogDashboardJSONRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()

	respByte, httpResp, err := utils.SendRequest(authV1, datadogClientV1, "GET", path+"/"+id, nil)
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dashboard := d.Get("dashboard").(string)

	respByte, _, err := utils.SendRequest(authV1, datadogClientV1, "POST", path, &dashboard)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error creating resource")
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
	updateDashboardJSONLists(d, providerConf, id.(string), layoutType.(string))

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dashboard := d.Get("dashboard").(string)
	id := d.Id()

	respByte, _, err := utils.SendRequest(authV1, datadogClientV1, "PUT", path+"/"+id, &dashboard)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}
	updateDashboardJSONLists(d, providerConf, id, layoutType.(string))

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()

	_, _, err := utils.SendRequest(authV1, datadogClientV1, "DELETE", path+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, "error deleting dashboard")
	}

	return nil
}

func updateDashboardJSONState(d *schema.ResourceData, dashboard map[string]interface{}) diag.Diagnostics {
	if v, ok := dashboard["url"]; ok {
		if err := d.Set("url", v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Remove computed fields from the object
	for _, f := range computedFields {
		delete(dashboard, f)
	}

	// Remove every widget id too
	deleteWidgetID(dashboard["widgets"].([]interface{}))

	dashboardString, err := structure.FlattenJsonToString(dashboard)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("dashboard", dashboardString); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateDashboardJSONLists(d *schema.ResourceData, providerConf *ProviderConfiguration, dashboardID, layoutType string) {
	dashTypeString := "custom_screenboard"
	if layoutType == "ordered" {
		dashTypeString = "custom_timeboard"
	}
	dashType := datadogV2.DashboardType(dashTypeString)
	itemsRequest := []datadogV2.DashboardListItemRequest{*datadogV2.NewDashboardListItemRequest(dashboardID, dashType)}
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if v, ok := d.GetOk("dashboard_lists"); ok && v.(*schema.Set).Len() > 0 {
		items := datadogV2.NewDashboardListAddItemsRequest()
		items.SetDashboards(itemsRequest)

		for _, id := range v.(*schema.Set).List() {
			_, _, err := datadogClientV2.DashboardListsApi.CreateDashboardListItems(authV2, int64(id.(int)), *items)
			if err != nil {
				log.Printf("[DEBUG] Got error adding to dashboard list %d: %v", id.(int), err)
			}
		}
	}

	if v, ok := d.GetOk("dashboard_lists_removed"); ok && v.(*schema.Set).Len() > 0 {
		items := datadogV2.NewDashboardListDeleteItemsRequest()
		items.SetDashboards(itemsRequest)

		for _, id := range v.(*schema.Set).List() {
			_, _, err := datadogClientV2.DashboardListsApi.DeleteDashboardListItems(authV2, int64(id.(int)), *items)
			if err != nil {
				log.Printf("[DEBUG] Got error removing from dashboard list %d: %v", id.(int), err)
			}
		}
	}
}
