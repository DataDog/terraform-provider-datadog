package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"
)

func resourceDatadogIntegrationPagerduty() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationPagerdutyCreate,
		Read:   resourceDatadogIntegrationPagerdutyRead,
		Update: resourceDatadogIntegrationPagerdutyUpdate,
		Delete: resourceDatadogIntegrationPagerdutyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"services": &schema.Schema{
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of service names and service keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"service_key": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
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
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func buildIntegrationPagerduty(d *schema.ResourceData) (*datadog.IntegrationPDRequest, error) {
	pd := &datadog.IntegrationPDRequest{}
	pd.SetSubdomain(d.Get("subdomain").(string))
	pd.SetAPIToken(d.Get("api_token").(string))

	schedules := []string{}
	for _, s := range d.Get("schedules").([]interface{}) {
		schedules = append(schedules, s.(string))
	}
	pd.Schedules = schedules

	services := []datadog.ServicePDRequest{}
	for _, sInterface := range d.Get("services").([]interface{}) {
		s := sInterface.(map[string]interface{})

		service := datadog.ServicePDRequest{}
		service.SetServiceName(s["service_name"].(string))
		service.SetServiceKey(s["service_key"].(string))

		services = append(services, service)
	}
	pd.Services = services

	return pd, nil
}

func resourceDatadogIntegrationPagerdutyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if err := client.CreateIntegrationPD(pd); err != nil {
		return fmt.Errorf("Failed to create integration pagerduty using Datadog API: %s", err.Error())
	}

	pdIntegration, err := client.GetIntegrationPD()
	if err != nil {
		return fmt.Errorf("error retrieving integration pagerduty: %s", err.Error())
	}

	d.SetId(pdIntegration.GetSubdomain())

	return nil
}

func resourceDatadogIntegrationPagerdutyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	pd, err := client.GetIntegrationPD()
	if err != nil {
		return err
	}

	services := []map[string]string{}
	for _, service := range pd.Services {
		services = append(services, map[string]string{
			"service_name": service.GetServiceName(),
			"service_key":  service.GetServiceKey(),
		})
	}

	d.Set("services", services)
	d.Set("subdomain", pd.GetSubdomain())
	d.Set("schedules", pd.Schedules)
	d.Set("api_token", pd.GetAPIToken())

	return nil
}

func resourceDatadogIntegrationPagerdutyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	pd, err := buildIntegrationPagerduty(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if err := client.UpdateIntegrationPD(pd); err != nil {
		return fmt.Errorf("Failed to create integration pagerduty using Datadog API: %s", err.Error())
	}

	return resourceDatadogIntegrationPagerdutyRead(d, meta)
}

func resourceDatadogIntegrationPagerdutyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteIntegrationPD(); err != nil {
		return fmt.Errorf("Error while deleting integration: %v", err)
	}

	return nil
}
