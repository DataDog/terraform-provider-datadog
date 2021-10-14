package datadog

import (
	"context"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsPipelineOrder() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Pipeline API resource, which is used to manage Datadog log pipelines order.",
		CreateContext: resourceDatadogLogsPipelineOrderCreate,
		UpdateContext: resourceDatadogLogsPipelineOrderUpdate,
		ReadContext:   resourceDatadogLogsPipelineOrderRead,
		DeleteContext: resourceDatadogLogsPipelineOrderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceDatadogLogsPipelineOrderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDatadogLogsPipelineOrderUpdate(ctx, d, meta)
}

func updateLogsPipelineOrderState(d *schema.ResourceData, order *datadogV1.LogsPipelinesOrder) diag.Diagnostics {
	if err := d.Set("pipelines", order.PipelineIds); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogLogsPipelineOrderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	order, httpResponse, err := datadogClientV1.LogsPipelinesApi.GetLogsPipelineOrder(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs pipeline order")
	}
	if err := utils.CheckForUnparsed(order); err != nil {
		return diag.FromErr(err)
	}

	return updateLogsPipelineOrderState(d, &order)
}

func resourceDatadogLogsPipelineOrderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	updatedOrder, httpResponse, err := datadogClientV1.LogsPipelinesApi.UpdateLogsPipelineOrder(authV1, ddPipelineList)
	if err != nil {
		// Cannot map pipelines to existing ones
		if strings.Contains(err.Error(), "422 Unprocessable Entity") {
			ddPipelineOrder, httpResponse, getErr := datadogClientV1.LogsPipelinesApi.GetLogsPipelineOrder(authV1)
			if getErr != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error getting logs pipeline order")
			}
			return diag.Errorf("cannot map pipelines to existing ones\n existing pipelines: %s\n pipeline to be updated: %s",
				ddPipelineOrder.PipelineIds,
				ddList)
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs pipeline order")
	}
	if err := utils.CheckForUnparsed(updatedOrder); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tfID)
	return updateLogsPipelineOrderState(d, &updatedOrder)
}

// The deletion of pipeline order is not supported from config API.
// This function simply delete the pipeline order resource from terraform state.
func resourceDatadogLogsPipelineOrderDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {

	return nil
}
