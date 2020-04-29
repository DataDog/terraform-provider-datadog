package datadog

import (
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const maskedSecret = "*****"

func resourceDatadogIntegrationPagerdutySO() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationPagerdutySOCreate,
		Read:   resourceDatadogIntegrationPagerdutySORead,
		Exists: resourceDatadogIntegrationPagerdutySOExists,
		Update: resourceDatadogIntegrationPagerdutySOUpdate,
		Delete: resourceDatadogIntegrationPagerdutySODelete,
		// since the API never returns service_key, it's impossible to meaningfully import resources
		Importer: nil,

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func buildIntegrationPagerdutySO(d *schema.ResourceData) *datadogV1.PagerDutyService {
	so := &datadogV1.PagerDutyService{}
	if v, ok := d.GetOk("service_name"); ok {
		so.SetServiceName(v.(string))
	}
	if v, ok := d.GetOk("service_key"); ok {
		so.SetServiceKey(v.(string))
	}

	return so
}

func buildIntegrationPagerdutyServiceKey(d *schema.ResourceData) *datadogV1.PagerDutyServiceKey {
	serviceKey := &datadogV1.PagerDutyServiceKey{}
	if v, ok := d.GetOk("service_key"); ok {
		serviceKey.SetServiceKey(v.(string))
	}

	return serviceKey
}

func resourceDatadogIntegrationPagerdutySOCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	so := buildIntegrationPagerdutySO(d)
	if _, _, err := datadogClientV1.PagerDutyIntegrationApi.CreatePagerDutyIntegrationService(authV1).Body(*so).Execute(); err != nil {
		// TODO: warn user that PD integration must be enabled to be able to create service objects
		return translateClientError(err, "error creating PagerDuty integration service")
	}
	d.SetId(so.GetServiceName())

	return resourceDatadogIntegrationPagerdutySORead(d, meta)
}

func resourceDatadogIntegrationPagerdutySORead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	so, _, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegrationService(authV1, d.Id()).Execute()
	if err != nil {
		return translateClientError(err, "error getting PagerDuty integration service")
	}

	d.Set("service_name", so.GetServiceName())
	// Only update service_key if not set on d - the API endpoints never return
	// the keys, so this is how we recognize new values.
	if _, ok := d.GetOk("service_key"); !ok {
		d.Set("service_key", maskedSecret)
	}

	return nil
}

func resourceDatadogIntegrationPagerdutySOExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	_, _, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegrationService(authV1, d.Id()).Execute()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking PagerDuty integration service exists")
	}

	return true, nil
}

func resourceDatadogIntegrationPagerdutySOUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	serviceKey := buildIntegrationPagerdutyServiceKey(d)
	if _, err := datadogClientV1.PagerDutyIntegrationApi.UpdatePagerDutyIntegrationService(authV1, d.Id()).Body(*serviceKey).Execute(); err != nil {
		return translateClientError(err, "error updating PagerDuty integration service")
	}

	return resourceDatadogIntegrationPagerdutySORead(d, meta)
}

func resourceDatadogIntegrationPagerdutySODelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if _, err := datadogClientV1.PagerDutyIntegrationApi.DeletePagerDutyIntegrationService(authV1, d.Id()).Execute(); err != nil {
		return translateClientError(err, "error deleting PagerDuty integration service")
	}

	return nil
}
