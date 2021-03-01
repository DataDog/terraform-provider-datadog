package datadog

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var computedFields = []string{"id", "author_handle", "author_name", "created_at", "modified_at", "url"}

func resourceDatadogDashboardJSON() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogDashboardJSONCreate,
		Read:   resourceDatadogDashboardJSONRead,
		Update: resourceDatadogDashboardJSONUpdate,
		Delete: resourceDatadogDashboardJSONDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"dashboard": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (warns []string, errs []error) {
					_, err := utils.NormalizeJSONorYamlString(v.(string))
					if err != nil {
						errs = append(errs, fmt.Errorf("invalid JSON or YAML: \n %s", v))
					}
					return
				},
				StateFunc: func(v interface{}) string {
					k, _ := utils.NormalizeJSONorYamlString(v.(string))
					// Remove computed fields when comparing diffs
					attrMap, _ := structure.ExpandJsonFromString(k)
					for _, f := range computedFields {
						delete(attrMap, f)
					}
					res, _ := structure.FlattenJsonToString(attrMap)
					return res
				},
				Description: "The JSON or YAML formatted definition of the Dashboard.",
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

func resourceDatadogDashboardJSONRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	path := "/api/v1/dashboard/" + d.Id()

	respByte, httpresp, err := utils.SendRequest(authV1, datadogClientV1, "GET", path, nil)
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

	return updateDashboardJSONState(d, respMap)
}

func resourceDatadogDashboardJSONCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	path := "/api/v1/dashboard"

	test, err := utils.NormalizeJSONorYamlString(d.Get("dashboard").(string))
	if err != nil {
		return err
	}

	respByte, _, err := utils.SendRequest(authV1, datadogClientV1, "POST", path, test)
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
	path := "/api/v1/dashboard/" + d.Id()

	jsonString, err := utils.NormalizeJSONorYamlString(d.Get("dashboard").(string))
	if err != nil {
		return err
	}

	respByte, _, err := utils.SendRequest(authV1, datadogClientV1, "PUT", path, jsonString)
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
	path := "/api/v1/dashboard/" + d.Id()

	_, _, err := utils.SendRequest(authV1, datadogClientV1, "DELETE", path, nil)
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

	dashboardString, err := structure.FlattenJsonToString(dashboard)
	if err != nil {
		return err
	}

	if err = d.Set("dashboard", dashboardString); err != nil {
		return err
	}
	return nil
}
