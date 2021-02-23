package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"reflect"
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: dashboardJsonDiffSuppress,
				StateFunc: func(v interface{}) string {
					jsonString, _ := structure.NormalizeJsonString(v)
					return jsonString
				},
				Description: "",
			},
		},
	}
}

func dashboardJsonDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	oldMap, err := structure.ExpandJsonFromString(old)
	if err != nil {
		return false
	}
	newMap, err := structure.ExpandJsonFromString(new)
	if err != nil {
		return false
	}

	for _, f := range computedFields {
		delete(oldMap, f)
		delete(newMap, f)
	}

	return reflect.DeepEqual(oldMap, newMap)
}

func resourceDatadogDashboardJsonRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	httpClient := providerConf.HttpClient
	path := "/api/v1/dashboard/" + d.Id()

	result, err := httpClient.SendRequest("GET", path, nil)
	if err != nil {
		return err
	}

	return updateDashboardJsonState(d, meta, result)
}

func resourceDatadogDashboardJsonCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	httpClient := providerConf.HttpClient
	path := "/api/v1/dashboard"

	obj, err := structure.ExpandJsonFromString(d.Get("dashboard_json").(string))
	if err != nil {
		return err
	}

	result, err := httpClient.SendRequest("POST", path, obj)
	if err != nil {
		return err
	}

	id := result["id"]
	d.SetId(id.(string))

	return updateDashboardJsonState(d, meta, result)
}

func updateDashboardJsonState(d *schema.ResourceData, meta interface{}, dashboard map[string]interface{}) error {
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
	httpClient := providerConf.HttpClient
	path := "/api/v1/dashboard/" + d.Id()

	obj, _ := structure.ExpandJsonFromString(d.Get("dashboard_json").(string))

	result, err := httpClient.SendRequest("PUT", path, obj)
	if err != nil {
		return err
	}

	return updateDashboardJsonState(d, meta, result)
}

func resourceDatadogDashboardJsonDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	httpClient := providerConf.HttpClient
	path := "/api/v1/dashboard/" + d.Id()

	_, err := httpClient.SendRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	return nil
}
