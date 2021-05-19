package datadog

import (
	"errors"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var computedFields = []string{"id", "author_handle", "author_name", "created_at", "modified_at", "url"}

const path = "/api/v1/dashboard"

func resourceDatadogDashboardJSON() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog dashboard JSON resource. This can be used to create and manage Datadog dashboards using the JSON definition.",
		Create:      resourceDatadogDashboardJSONCreate,
		Read:        resourceDatadogDashboardJSONRead,
		Update:      resourceDatadogDashboardJSONUpdate,
		Delete:      resourceDatadogDashboardJSONDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceDatadogDashboardJSONRead(d *schema.ResourceData, meta interface{}) error {
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
		return err
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return err
	}

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dashboard := d.Get("dashboard").(string)

	respByte, _, err := utils.SendRequest(authV1, datadogClientV1, "POST", path, &dashboard)
	if err != nil {
		return utils.TranslateClientError(err, "error creating resource")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return err
	}

	id, ok := respMap["id"]
	if !ok {
		return errors.New("error retrieving id from response")
	}
	d.SetId(id.(string))

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	dashboard := d.Get("dashboard")
	id := d.Id()

	respByte, _, err := utils.SendRequest(authV1, datadogClientV1, "PUT", path+"/"+id, &dashboard)
	if err != nil {
		return utils.TranslateClientError(err, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return err
	}

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	id := d.Id()

	_, _, err := utils.SendRequest(authV1, datadogClientV1, "DELETE", path+"/"+id, nil)
	if err != nil {
		return utils.TranslateClientError(err, "error deleting dashboard")
	}

	return nil
}

func updateDashboardJSONState(d *schema.ResourceData, dashboard map[string]interface{}) error {
	if v, ok := dashboard["url"]; ok {
		if err := d.Set("url", v.(string)); err != nil {
			return err
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
		return err
	}

	if err = d.Set("dashboard", dashboardString); err != nil {
		return err
	}
	return nil
}
