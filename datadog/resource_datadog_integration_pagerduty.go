package datadog

import (
	"fmt"
	"strings"
	"sync"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// creating/modifying/deleting PD integration and its service objects in parallel on one account
// is unsupported by the API right now; therefore we use the mutex to only operate on one at a time
var integrationPdMutex = sync.Mutex{}

func resourceDatadogIntegrationPagerduty() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationPagerdutyCreate,
		Read:   resourceDatadogIntegrationPagerdutyRead,
		Exists: resourceDatadogIntegrationPagerdutyExists,
		Update: resourceDatadogIntegrationPagerdutyUpdate,
		Delete: resourceDatadogIntegrationPagerdutyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"individual_services": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"services": {
				ConflictsWith: []string{"individual_services"},
				Deprecated:    "set \"individual_services\" to true and use datadog_pagerduty_integration_service_object",
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "A list of service names and service keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"service_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"subdomain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schedules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"api_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func buildIntegrationPagerduty(d *schema.ResourceData) (*datadog.PagerDutyIntegration, error) {
	pd := &datadog.PagerDutyIntegration{}
	pd.SetSubdomain(d.Get("subdomain").(string))
	pd.SetApiToken(d.Get("api_token").(string))

	schedules := []string{}
	for _, s := range d.Get("schedules").([]interface{}) {
		if s.(string) != "" {
			schedules = append(schedules, s.(string))
		}
	}
	pd.SetSchedules(schedules)

	services := []datadog.PagerDutyService{}
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		//services = nil
	} else {
		configServices, ok := d.GetOk("services")
		if ok {
			for _, sInterface := range configServices.([]interface{}) {
				s := sInterface.(map[string]interface{})

				if s["service_name"].(string) != "" && s["service_key"].(string) != "" {
					service := datadog.PagerDutyService{}
					service.SetServiceName(s["service_name"].(string))
					service.SetServiceKey(s["service_key"].(string))

					services = append(services, service)
				}
			}
		}
	}
	pd.SetServices(services)

	return pd, nil
}

func resourceDatadogIntegrationPagerdutyCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if _, err := client.PagerDutyIntegrationApi.CreatePagerDutyIntegration(auth).Body(*pd).Execute(); err != nil {
		return translateClientError(err, "Failed to create integration pagerduty using Datadog API")
	}

	pdIntegration, _, err := client.PagerDutyIntegrationApi.GetPagerDutyIntegration(auth).Execute()
	if err != nil {
		return translateClientError(err, "error retrieving integration pagerduty")
	}

	d.SetId(pdIntegration.GetSubdomain())

	return nil
}

func resourceDatadogIntegrationPagerdutyRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	pd, _, err := client.PagerDutyIntegrationApi.GetPagerDutyIntegration(auth).Execute()
	if err != nil {
		return err
	}

	services := []map[string]string{}
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = nil
	} else {
		for _, service := range pd.GetServices() {
			if service.GetServiceName() != "" && service.GetServiceKey() != "" {
				services = append(services, map[string]string{
					"service_name": service.GetServiceName(),
					"service_key":  service.GetServiceKey(),
				})
			}
		}
	}

	d.Set("services", services)
	d.Set("subdomain", pd.GetSubdomain())
	d.Set("schedules", pd.GetSchedules())
	d.Set("api_token", pd.GetApiToken())

	return nil
}

func resourceDatadogIntegrationPagerdutyExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	_, _, err := client.PagerDutyIntegrationApi.GetPagerDutyIntegration(auth).Execute()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resourceDatadogIntegrationPagerdutyUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	pdss := datadog.NewPagerDutyServicesAndSchedules()
	pdss.SetServices(pd.GetServices())
	pdss.SetSchedules(pd.GetSchedules())
	if _, err := client.PagerDutyIntegrationApi.UpdatePagerDutyIntegration(auth).Body(*pdss).Execute(); err != nil {
		return translateClientError(err, "Failed to update integration pagerduty using Datadog API")
	}

	return resourceDatadogIntegrationPagerdutyRead(d, meta)
}

func resourceDatadogIntegrationPagerdutyDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if _, err := client.PagerDutyIntegrationApi.DeletePagerDutyIntegration(auth).Execute(); err != nil {
		return translateClientError(err, "Error while deleting integration")
	}

	return nil
}
