package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsIntegrationPipeline() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Pipeline API resource to manage the integrations. Integration pipelines are the pipelines that are automatically installed for your organization when sending the logs with specific sources. You don't need to maintain or update these types of pipelines. Keeping them as resources, however, allows you to manage the order of your pipelines by referencing them in your `datadog_logs_pipeline_order` resource. If you don't need the `pipeline_order` feature, this resource declaration can be omitted.",
		CreateContext: resourceDatadogLogsIntegrationPipelineCreate,
		UpdateContext: resourceDatadogLogsIntegrationPipelineUpdate,
		ReadContext:   resourceDatadogLogsIntegrationPipelineRead,
		DeleteContext: resourceDatadogLogsIntegrationPipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"is_enabled": {
				Description: "Boolean value to enable your pipeline.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func resourceDatadogLogsIntegrationPipelineCreate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return diag.Errorf("cannot create an integration pipeline, please import it first to make changes")
}

func updateLogsIntegrationPipelineState(d *schema.ResourceData, pipeline *datadogV1.LogsPipeline) diag.Diagnostics {
	if err := d.Set("is_enabled", pipeline.GetIsEnabled()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogLogsIntegrationPipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	ddPipeline, httpresp, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 400 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting logs integration pipeline")
	}
	if err := utils.CheckForUnparsed(ddPipeline); err != nil {
		return diag.FromErr(err)
	}
	if !ddPipeline.GetIsReadOnly() {
		d.SetId("")
		return nil
	}
	return updateLogsIntegrationPipelineState(d, &ddPipeline)
}

func resourceDatadogLogsIntegrationPipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var ddPipeline datadogV1.LogsPipeline
	ddPipeline.SetIsEnabled(d.Get("is_enabled").(bool))
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	updatedPipeline, httpResponse, err := datadogClientV1.LogsPipelinesApi.UpdateLogsPipeline(authV1, d.Id(), ddPipeline)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs integration pipeline")
	}
	if err := utils.CheckForUnparsed(updatedPipeline); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*updatedPipeline.Id)
	return updateLogsIntegrationPipelineState(d, &updatedPipeline)
}

func resourceDatadogLogsIntegrationPipelineDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
