package datadog

import (
	"fmt"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogLogsIntegrationPipeline() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog Logs Pipeline API resource to manage the integrations. Integration pipelines are the pipelines that are automatically installed for your organization when sending the logs with specific sources. You don't need to maintain or update these types of pipelines. Keeping them as resources, however, allows you to manage the order of your pipelines by referencing them in your `datadog_logs_pipeline_order` resource. If you don't need the `pipeline_order` feature, this resource declaration can be omitted.",
		Create:      resourceDatadogLogsIntegrationPipelineCreate,
		Update:      resourceDatadogLogsIntegrationPipelineUpdate,
		Read:        resourceDatadogLogsIntegrationPipelineRead,
		Delete:      resourceDatadogLogsIntegrationPipelineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceDatadogLogsIntegrationPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("cannot create an integration pipeline, please import it first to make changes")
}

func resourceDatadogLogsIntegrationPipelineRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	ddPipeline, httpresp, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id()).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 400 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientError(err, "error getting logs integration pipeline")
	}
	if !ddPipeline.GetIsReadOnly() {
		d.SetId("")
		return nil
	}
	if err := d.Set("is_enabled", ddPipeline.GetIsEnabled()); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsIntegrationPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	var ddPipeline datadogV1.LogsPipeline
	ddPipeline.SetIsEnabled(d.Get("is_enabled").(bool))
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	updatedPipeline, _, err := datadogClientV1.LogsPipelinesApi.UpdateLogsPipeline(authV1, d.Id()).Body(ddPipeline).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error updating logs integration pipeline")
	}
	d.SetId(*updatedPipeline.Id)
	return resourceDatadogLogsIntegrationPipelineRead(d, meta)
}

func resourceDatadogLogsIntegrationPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
