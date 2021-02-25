package datadog

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var computedFields = []string{"id", "author_handle", "author_name", "created_at", "modified_at", "url"}

func resourceDatadogDashboardJson() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogDashboardJsonCreate,
		Read:   resourceDatadogDashboardJsonRead,
		Update: resourceDatadogDashboardJsonUpdate,
		Delete: resourceDatadogDashboardJsonDelete,
		Schema: map[string]*schema.Schema{
			"dashboard_json": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					jsonString, _ := structure.NormalizeJsonString(v)
					// Remove computed fields from the object
					attrMap, _ := structure.ExpandJsonFromString(jsonString)
					for _, f := range computedFields {
						delete(attrMap, f)
					}
					jsonString, _ = structure.FlattenJsonToString(attrMap)

					return jsonString
				},
				Description: "",
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

func resourceDatadogDashboardJsonRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	path := "/api/v1/dashboard/" + d.Id()

	respByte, httpresp, err := utils.SendRequest(datadogClientV1, authV1, "GET", path, nil)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return err
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return err
	}

	return updateDashboardJsonState(d, meta, respMap)
}

func resourceDatadogDashboardJsonCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	path := "/api/v1/dashboard"


	jsonString, err := structure.NormalizeJsonString(d.Get("dashboard_json").(string))
	if err != nil {
		return err
	}

	respByte, _, err := utils.SendRequest(datadogClientV1, authV1, "POST", path, jsonString)
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

	return updateDashboardJsonState(d, meta, respMap)
}

func updateDashboardJsonState(d *schema.ResourceData, meta interface{}, dashboard map[string]interface{}) error {
	if v, ok := dashboard["url"]; ok {
		d.Set("url", v.(string))
	}

	// Remove computed fields from the object
	for _, f := range computedFields {
		delete(dashboard, f)
	}

	dashboardString, err := structure.FlattenJsonToString(dashboard)
	if err != nil {
		return err
	}

	if err = d.Set("dashboard_json", dashboardString); err != nil {
		return err
	}
	return nil
}

func resourceDatadogDashboardJsonUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	path := "/api/v1/dashboard/" + d.Id()

	jsonString, err := structure.NormalizeJsonString(d.Get("dashboard_json").(string))
	if err != nil {
		return err
	}

	respByte, _, err := utils.SendRequest(datadogClientV1, authV1, "PUT", path, jsonString)
	if err != nil {
		return utils.TranslateClientError(err, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return err
	}

	return updateDashboardJsonState(d, meta, respMap)
}

func resourceDatadogDashboardJsonDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	path := "/api/v1/dashboard/" + d.Id()

	_, _, err := utils.SendRequest(datadogClientV1, authV1, "DELETE", path, nil)
	if err != nil {
		return utils.TranslateClientError(err, "error deleting dashboard")
	}

	return nil
}
