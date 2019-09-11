package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
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
	pipelineList, err := meta.(*datadog.Client).GetLogsPipelineList()
	if err != nil {
		return err
	}
	tfList := make([]string, len(pipelineList.PipelineIds))
	for i, id := range pipelineList.PipelineIds {
		tfList[i] = id
	}
	if err = d.Set("pipelines", tfList); err != nil {
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
	if _, err := meta.(*datadog.Client).UpdateLogsPipelineList(&ddPipelineList); err != nil {
		return fmt.Errorf("error updating logs pipeline list: (%s)", err.Error())
	}
	if name, ok := d.GetOk("name"); ok {
		d.SetId(name.(string))
	}
	return resourceDatadogLogsPipelineOrderRead(d, meta)
}

// The deletion of pipeline order is not supported from config API.
// This function simply delete the pipeline order resource from terraform state.
func resourceDatadogLogsPipelineOrderDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceDatadogLogsPipelineOrderExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	if pipelineList, err := meta.(*datadog.Client).GetLogsPipelineList(); err == nil && len(pipelineList.PipelineIds) > 0 {
		return true, nil
	}
	return false, nil
}
