package datadog

import (
	"fmt"
	"strings"
	"sync"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
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

func buildIntegrationPagerduty(d *schema.ResourceData) (datadogV1.PagerDutyIntegration, error) {
	pd := &datadogV1.PagerDutyIntegration{}
	pd.SetSubdomain(d.Get("subdomain").(string))
	pd.SetApiToken(d.Get("api_token").(string))

	var schedules []string
	if v, ok := d.GetOk("schedules"); ok {
		for _, s := range v.([]interface{}) {
			schedules = append(schedules, s.(string))
		}
	} else {
		// Explicitly return an empty array. The API will respond with a 400 if the value is null
		schedules = []string{}
	}
	pd.SetSchedules(schedules)

	var services []datadogV1.PagerDutyService
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = []datadogV1.PagerDutyService{}
	} else {
		configServices, ok := d.GetOk("services")
		if ok {
			for _, sInterface := range configServices.([]interface{}) {
				s := sInterface.(map[string]interface{})

				service := datadogV1.PagerDutyService{}
				service.SetServiceName(s["service_name"].(string))
				service.SetServiceKey(s["service_key"].(string))

				services = append(services, service)
			}
		} else {
			// Explicitly return an empty array. The API will respond with a 400 if the value is null
			services = []datadogV1.PagerDutyService{}
		}
	}
	pd.SetServices(services)

	return *pd, nil
}

func resourceDatadogIntegrationPagerdutyCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}

	if _, err := datadogClientV1.PagerDutyIntegrationApi.CreatePagerDutyIntegration(authV1).Body(pd).Execute(); err != nil {
		return translateClientError(err, "error creating PagerDuty integration")
	}

	pdIntegration, _, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegration(authV1).Execute()
	if err != nil {
		return translateClientError(err, "error getting PagerDuty integration")
	}

	d.SetId(pdIntegration.GetSubdomain())

	return resourceDatadogIntegrationPagerdutyRead(d, meta)
}

func resourceDatadogIntegrationPagerdutyRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	pd, _, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegration(authV1).Execute()
	if err != nil {
		return translateClientError(err, "error getting PagerDuty integration")
	}

	var services []map[string]string
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = nil
	} else {
		for _, service := range pd.GetServices() {
			services = append(services, map[string]string{
				"service_name": service.GetServiceName(),
				"service_key":  service.GetServiceKey(),
			})
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	_, _, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegration(authV1).Execute()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error getting PagerDuty integration")
	}

	return true, nil
}

func resourceDatadogIntegrationPagerdutyUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}

	// Use CreatePagerDutyIntegration method to update the test. UpdatePagerDutyIntegration() accepts type
	// PagerDutyServicesAndSchedules which does does not have field to update subdomain and api_token
	if _, err := datadogClientV1.PagerDutyIntegrationApi.CreatePagerDutyIntegration(authV1).Body(pd).Execute(); err != nil {
		return translateClientError(err, "error updating PagerDuty integration")
	}

	// if there are none currently configured services, we actually
	// have to remove them explicitly, otherwise the underlying API client
	// would not send the "services" key at all and they wouldn't get deleted
	if value, ok := d.GetOk("individual_services"); !ok || !value.(bool) {
		currentServices := d.Get("services").([]interface{})
		if len(currentServices) == 0 {
			pd, _, err := datadogClientV1.PagerDutyIntegrationApi.GetPagerDutyIntegration(authV1).Execute()
			if err != nil {
				return translateClientError(err, "error getting PagerDuty integration")
			}
			for _, service := range pd.GetServices() {
				if _, err := datadogClientV1.PagerDutyIntegrationApi.DeletePagerDutyIntegrationService(authV1, service.GetServiceName()).Execute(); err != nil {
					return translateClientError(err, "error deleting PagerDuty integration service")
				}
			}
		}
	}
	return resourceDatadogIntegrationPagerdutyRead(d, meta)
}

func resourceDatadogIntegrationPagerdutyDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if _, err := datadogClientV1.PagerDutyIntegrationApi.DeletePagerDutyIntegration(authV1).Execute(); err != nil {
		return translateClientError(err, "error deleting PagerDuty integration")
	}

	return nil
}
