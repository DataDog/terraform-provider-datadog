package datadog

import (
	"fmt"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogLogsIntegrationPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsIntegrationPipelineCreate,
		Update: resourceDatadogLogsIntegrationPipelineUpdate,
		Read:   resourceDatadogLogsIntegrationPipelineRead,
		Delete: resourceDatadogLogsIntegrationPipelineDelete,
		Exists: resourceDatadogLogsIntegrationPipelineExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"is_enabled": {
				Description: "Boolean value to enable your pipeline.",
				Type: schema.TypeBool,
				Optional: true,
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
	ddPipeline, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id()).Execute()
	if err != nil {
		return translateClientError(err, "error getting logs integration pipeline")
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
		return translateClientError(err, "error updating logs integration pipeline")
	}
	d.SetId(*updatedPipeline.Id)
	return resourceDatadogLogsIntegrationPipelineRead(d, meta)
}

func resourceDatadogLogsIntegrationPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogLogsIntegrationPipelineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	ddPipeline, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, d.Id()).Execute()
	if err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through GET request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return false, nil
		}
		return false, translateClientError(err, "error checking logs integration pipeline exists")
	}
	return ddPipeline.GetIsReadOnly(), nil
}
