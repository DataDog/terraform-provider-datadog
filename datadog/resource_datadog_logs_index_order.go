package datadog

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogLogsIndexOrder() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsIndexOrderCreate,
		Update: resourceDatadogLogsIndexOrderUpdate,
		Read:   resourceDatadogLogsIndexOrderRead,
		Delete: resourceDatadogLogsIndexOrderDelete,
		Exists: resourceDatadogLogsIndexOrderExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
			"indexes": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsIndexOrderCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceDatadogLogsIndexOrderUpdate(d, meta)
}

func resourceDatadogLogsIndexOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	var ddIndexList datadog.LogsIndexList
	tfList := d.Get("indexes").([]interface{})
	ddList := make([]string, len(tfList))
	for i, tfName := range tfList {
		ddList[i] = tfName.(string)
	}
	ddIndexList.IndexNames = ddList
	var tfId string
	if name, exists := d.GetOk("name"); exists {
		tfId = name.(string)
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if _, err := client.UpdateLogsIndexList(&ddIndexList); err != nil {
		return translateClientError(err, "error updating logs index list")
	}
	d.SetId(tfId)
	return resourceDatadogLogsIndexOrderRead(d, meta)
}

func resourceDatadogLogsIndexOrderRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	ddIndexList, err := client.GetLogsIndexList()
	if err != nil {
		return translateClientError(err, "error getting logs index list")
	}
	if err := d.Set("indexes", ddIndexList.IndexNames); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsIndexOrderDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogLogsIndexOrderExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}
