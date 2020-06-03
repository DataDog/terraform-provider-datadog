package datadog

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

// creating/modifying/deleting PD integration and its service objects in parallel on one account
// is unsupported by the API right now; therefore we use the mutex to only operate on one at a time
var integrationPdMutex = sync.Mutex{}

func resourceDatadogIntegrationPagerduty() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated. Instead use the Pagerduty Service Object resource",
		Create:             resourceDatadogIntegrationPagerdutyCreate,
		Read:               resourceDatadogIntegrationPagerdutyRead,
		Exists:             resourceDatadogIntegrationPagerdutyExists,
		Update:             resourceDatadogIntegrationPagerdutyUpdate,
		Delete:             resourceDatadogIntegrationPagerdutyDelete,
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

func buildIntegrationPagerduty(d *schema.ResourceData) (*datadog.IntegrationPDRequest, error) {
	pd := &datadog.IntegrationPDRequest{}
	pd.SetSubdomain(d.Get("subdomain").(string))
	pd.SetAPIToken(d.Get("api_token").(string))

	var schedules []string
	for _, s := range d.Get("schedules").([]interface{}) {
		schedules = append(schedules, s.(string))
	}
	pd.Schedules = schedules

	var services []datadog.ServicePDRequest
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = nil
	} else {
		configServices, ok := d.GetOk("services")
		if ok {
			for _, sInterface := range configServices.([]interface{}) {
				s := sInterface.(map[string]interface{})

				service := datadog.ServicePDRequest{}
				service.SetServiceName(s["service_name"].(string))
				service.SetServiceKey(s["service_key"].(string))

				services = append(services, service)
			}
		}
	}
	pd.Services = services

	return pd, nil
}

func resourceDatadogIntegrationPagerdutyCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}

	if err := client.CreateIntegrationPD(pd); err != nil {
		return translateClientError(err, "error creating PagerDuty integration")
	}

	pdIntegration, err := client.GetIntegrationPD()
	if err != nil {
		return translateClientError(err, "error getting PagerDuty integration")
	}

	d.SetId(pdIntegration.GetSubdomain())

	return resourceDatadogIntegrationPagerdutyRead(d, meta)
}

func resourceDatadogIntegrationPagerdutyRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	pd, err := client.GetIntegrationPD()
	if err != nil {
		return translateClientError(err, "error getting PagerDuty integration")
	}

	var services []map[string]string
	if value, ok := d.GetOk("individual_services"); ok && value.(bool) {
		services = nil
	} else {
		for _, service := range pd.Services {
			services = append(services, map[string]string{
				"service_name": service.GetServiceName(),
				"service_key":  service.GetServiceKey(),
			})
		}
	}

	d.Set("services", services)
	d.Set("subdomain", pd.GetSubdomain())
	d.Set("schedules", pd.Schedules)

	return nil
}

func resourceDatadogIntegrationPagerdutyExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	_, err := client.GetIntegrationPD()
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
	client := providerConf.CommunityClient

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}

	if err := client.UpdateIntegrationPD(pd); err != nil {
		return translateClientError(err, "error updating PagerDuty integration")
	}

	// if there are none currently configured services, we actually
	// have to remove them explicitly, otherwise the underlying API client
	// would not send the "services" key at all and they wouldn't get deleted
	if value, ok := d.GetOk("individual_services"); !ok || !value.(bool) {
		currentServices := d.Get("services").([]interface{})
		if len(currentServices) == 0 {
			pd, err := client.GetIntegrationPD()
			if err != nil {
				return translateClientError(err, "error getting PagerDuty integration")
			}
			for _, service := range pd.Services {
				if err := client.DeleteIntegrationPDService(*service.ServiceName); err != nil {
					return translateClientError(err, "error deleting PagerDuty integration service")
				}
			}
		}
	}
	return resourceDatadogIntegrationPagerdutyRead(d, meta)
}

func resourceDatadogIntegrationPagerdutyDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if err := client.DeleteIntegrationPD(); err != nil {
		return translateClientError(err, "error deleting PagerDuty integration")
	}

	return nil
}
