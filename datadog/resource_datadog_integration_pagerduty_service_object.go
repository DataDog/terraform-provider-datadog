package datadog

import (
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const maskedSecret = "*****"

func resourceDatadogIntegrationPagerdutySO() *schema.Resource {
	return &schema.Resource{
		Description: "Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.",
		Create:      resourceDatadogIntegrationPagerdutySOCreate,
		Read:        resourceDatadogIntegrationPagerdutySORead,
		Update:      resourceDatadogIntegrationPagerdutySOUpdate,
		Delete:      resourceDatadogIntegrationPagerdutySODelete,
		// since the API never returns service_key, it's impossible to meaningfully import resources
		Importer: nil,

		Schema: map[string]*schema.Schema{
			"service_name": {
				Description: "Your Service name in PagerDuty.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"service_key": {
				Description: "Your Service name associated service key in PagerDuty. Note: Since the Datadog API never returns service keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform?_ga=2.15990198.1091155358.1609189257-888022054.1605547463). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html?_ga=2.15990198.1091155358.1609189257-888022054.1605547463) to have it destroyed and recreated.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
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

	so, httpresp, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegrationService(authV1, d.Id()).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
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
