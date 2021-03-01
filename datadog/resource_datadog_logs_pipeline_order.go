package datadog

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogLogsPipelineOrder() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog Logs Pipeline API resource, which is used to manage Datadog log pipelines order.",
		Create:      resourceDatadogLogsPipelineOrderCreate,
		Update:      resourceDatadogLogsPipelineOrderUpdate,
		Read:        resourceDatadogLogsPipelineOrderRead,
		Delete:      resourceDatadogLogsPipelineOrderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name attribute in the resource `datadog_logs_pipeline_order` needs to be unique. It's recommended to use the same value as the resource name. No related field is available in [Logs Pipeline API](https://docs.datadoghq.com/api/v1/logs-pipelines/#get-pipeline-order).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"pipelines": {
				Description: "The pipeline IDs list. The order of pipeline IDs in this attribute defines the overall pipeline order for logs.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogLogsPipelineOrderCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceDatadogLogsPipelineOrderUpdate(d, meta)
}

func updateLogsPipelineOrderState(d *schema.ResourceData, order *datadogV1.LogsPipelinesOrder) error {
	if err := d.Set("pipelines", order.PipelineIds); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsPipelineOrderRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	order, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipelineOrder(authV1).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error getting logs pipeline order")
	}

	return updateLogsPipelineOrderState(d, &order)
}

func resourceDatadogLogsPipelineOrderUpdate(d *schema.ResourceData, meta interface{}) error {
	var ddPipelineList datadogV1.LogsPipelinesOrder
	tfList := d.Get("pipelines").([]interface{})
	ddList := make([]string, len(tfList))
	for i, id := range tfList {
		ddList[i] = id.(string)
	}
	ddPipelineList.PipelineIds = ddList
	var tfID string
	if name, exists := d.GetOk("name"); exists {
		tfID = name.(string)
	}
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	updatedOrder, _, err := datadogClientV1.LogsPipelinesApi.UpdateLogsPipelineOrder(authV1).Body(ddPipelineList).Execute()
	if err != nil {
		// Cannot map pipelines to existing ones
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			ddPipelineOrder, _, getErr := datadogClientV1.LogsPipelinesApi.GetLogsPipelineOrder(authV1).Execute()
			if getErr != nil {
				return utils.TranslateClientError(err, "error getting logs pipeline order")
			}
			return fmt.Errorf("cannot map pipelines to existing ones\n existing pipelines: %s\n pipeline to be updated: %s",
				ddPipelineOrder.PipelineIds,
				ddList)
		}
		return utils.TranslateClientError(err, "error updating logs pipeline order")
	}
	d.SetId(tfID)
	return updateLogsPipelineOrderState(d, &updatedOrder)
}

// The deletion of pipeline order is not supported from config API.
// This function simply delete the pipeline order resource from terraform state.
func resourceDatadogLogsPipelineOrderDelete(_ *schema.ResourceData, _ interface{}) error {

	return nil
}
