package datadog

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogLogsPipelineOrder() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsPipelineOrderCreate,
		Update: resourceDatadogLogsPipelineOrderUpdate,
		Read:   resourceDatadogLogsPipelineOrderRead,
		Delete: resourceDatadogLogsPipelineOrderDelete,
		Exists: resourceDatadogLogsPipelineOrderExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString, Required: true},
			"pipelines": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsPipelineOrderCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceDatadogLogsPipelineOrderUpdate(d, meta)
}

func resourceDatadogLogsPipelineOrderRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	ddList, err := client.GetLogsPipelineList()
	if err != nil {
		return translateClientError(err, "error getting logs pipeline order")
	}

	if err = d.Set("pipelines", ddList.PipelineIds); err != nil {
		return err
	}

	return nil
}

func resourceDatadogLogsPipelineOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	var ddPipelineList datadog.LogsPipelineList
	tfList := d.Get("pipelines").([]interface{})
	ddList := make([]string, len(tfList))
	for i, id := range tfList {
		ddList[i] = id.(string)
	}
	ddPipelineList.PipelineIds = ddList
	var tfId string
	if name, exists := d.GetOk("name"); exists {
		tfId = name.(string)
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if _, err := client.UpdateLogsPipelineList(&ddPipelineList); err != nil {
		// Cannot map pipelines to existing ones
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			ddPipelineOrder, getErr := client.GetLogsPipelineList()
			if getErr != nil {
				return translateClientError(err, "error getting logs pipeline order")
			}
			return fmt.Errorf("cannot map pipelines to existing ones\n existing pipelines: %s\n pipeline to be updated: %s",
				ddPipelineOrder.PipelineIds,
				ddList)
		}
		return translateClientError(err, "error updating logs pipeline order")
	}
	d.SetId(tfId)
	return resourceDatadogLogsPipelineOrderRead(d, meta)
}

// The deletion of pipeline order is not supported from config API.
// This function simply delete the pipeline order resource from terraform state.
func resourceDatadogLogsPipelineOrderDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceDatadogLogsPipelineOrderExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}
