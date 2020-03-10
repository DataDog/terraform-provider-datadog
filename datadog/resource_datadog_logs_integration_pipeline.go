package datadog

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
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
			"is_enabled": {Type: schema.TypeBool, Optional: true},
		},
	}
}

func resourceDatadogLogsIntegrationPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("cannot create an integration pipeline, please import it first to make changes")
}

func resourceDatadogLogsIntegrationPipelineRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	ddPipeline, _, err := client.LogsPipelinesApi.GetLogsPipeline(auth, d.Id()).Execute()
	if err != nil {
		return translateClientError(err,"error getting logs pipeline")
	}
	if err := d.Set("is_enabled", ddPipeline.GetIsEnabled()); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsIntegrationPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	var ddPipeline datadog.LogsPipeline
	ddPipeline.SetIsEnabled(d.Get("is_enabled").(bool))
	updatedPipeline, _, err := client.LogsPipelinesApi.UpdateLogsPipeline(auth, d.Id()).Body(ddPipeline).Execute()
	if err != nil {
		return translateClientError(err,"error updating logs pipeline")
	}
	d.SetId(*updatedPipeline.Id)
	return resourceDatadogLogsIntegrationPipelineRead(d, meta)
}

func resourceDatadogLogsIntegrationPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogLogsIntegrationPipelineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	ddPipeline, _, err := client.LogsPipelinesApi.GetLogsPipeline(auth, d.Id()).Execute()
	if err != nil {
		// API returns 400 when the specific pipeline id doesn't exist through GET request.
		if strings.Contains(err.Error(), "400 Bad Request") {
			return false, nil
		}
		return false, translateClientError(err,"error getting logs pipeline")
	}
	return ddPipeline.GetIsReadOnly(), nil
}
