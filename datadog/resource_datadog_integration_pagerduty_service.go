package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogIntegrationPagerdutyService() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationPagerdutyServiceCreate,
		Read:   resourceDatadogIntegrationPagerdutyServiceRead,
		Exists: resourceDatadogIntegrationPagerdutyServiceExists,
		Update: resourceDatadogIntegrationPagerdutyServiceUpdate,
		Delete: resourceDatadogIntegrationPagerdutyServiceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationPagerdutyServiceImport,
		},

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
			"notify_handle": {
				Type:      schema.TypeString,
				Computed:  true,
			},
		},
	}
}

func resourceDatadogIntegrationPagerdutyServiceExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*datadog.Client)

	pd, err := client.GetIntegrationPD()
	if err != nil {
		return false, err
	}

	serviceName := d.Get("service_name").(string)
	for _, service := range pd.Services {
		if service.GetServiceName() == serviceName {
			return true, nil
		}
	}
	return false, nil
}

func resourceDatadogIntegrationPagerdutyServiceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	serviceName := d.Get("service_name").(string)
	serviceKey := d.Get("service_key").(string)

	pdServiceReq := datadog.ServicePDRequest{}
	pdServiceReq.SetServiceName(serviceName)
	pdServiceReq.SetServiceKey(serviceKey)

	pdReq := &datadog.IntegrationPDRequest{}
	pdReq.Services = []datadog.ServicePDRequest{pdServiceReq}

	if err := client.CreateIntegrationPD(pdReq); err != nil {
		return fmt.Errorf("failed to create pagerduty service mapping using Datadog API: %s", err.Error())
	}

	d.SetId(serviceName)
	d.Set("notify_handle", "@pagerduty-" + serviceName)

	return nil
}

func resourceDatadogIntegrationPagerdutyServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	pd, err := client.GetIntegrationPD()
	if err != nil {
		return err
	}

	serviceName := d.Get("service_name").(string)
	serviceKey := d.Get("service_key").(string)
	entryFound := false

	services := []datadog.ServicePDRequest{}
	for _, service := range pd.Services {
		pdServiceReq := datadog.ServicePDRequest{}
		pdServiceReq.SetServiceName(service.GetServiceName())
		pdServiceReq.SetServiceKey(service.GetServiceKey())

		if service.GetServiceName() == serviceName {
			entryFound = true
			pdServiceReq.SetServiceKey(serviceKey)
		}
		services = append(services, pdServiceReq)
	}

	if !entryFound {
		pdServiceReq := datadog.ServicePDRequest{}
		pdServiceReq.SetServiceName(serviceName)
		pdServiceReq.SetServiceKey(serviceKey)
		services = append(services, pdServiceReq)
	}

	pdReq := &datadog.IntegrationPDRequest{}
	pdReq.Services = services

	if err := client.UpdateIntegrationPD(pdReq); err != nil {
		return fmt.Errorf("failed to update pagerduty service mapping using Datadog API: %s", err.Error())
	}

	d.SetId(serviceName)
	d.Set("service_name", serviceName)
	d.Set("service_key", serviceKey)
	d.Set("notify_handle", "@pagerduty-" + serviceName)
	return nil
}

func resourceDatadogIntegrationPagerdutyServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	pd, err := client.GetIntegrationPD()
	if err != nil {
		return err
	}

	serviceName := d.Get("service_name").(string)

	for _, service := range pd.Services {
		if service.GetServiceName() == serviceName {
			d.Set("service_name", serviceName)
			return nil
		}
	}

	return fmt.Errorf("failed to loacate serivce with name: %s", serviceName)
}

func resourceDatadogIntegrationPagerdutyServiceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	pd, err := client.GetIntegrationPD()
	if err != nil {
		return err
	}

	serviceName := d.Get("service_name").(string)

	services := []datadog.ServicePDRequest{}
	for _, service := range pd.Services {
		if service.GetServiceName() != serviceName {
			pdServiceReq := datadog.ServicePDRequest{}
			pdServiceReq.SetServiceName(service.GetServiceName())
			pdServiceReq.SetServiceKey(service.GetServiceKey())
			services = append(services, pdServiceReq)
		}
	}

	pdr := &datadog.IntegrationPDRequest{}
	pdr.Services = services

	if err := client.UpdateIntegrationPD(pdr); err != nil {
		return fmt.Errorf("failed to delete pagerduty service mapping using Datadog API: %s", err.Error())
	}

	return nil
}

func resourceDatadogIntegrationPagerdutyServiceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationPagerdutyServiceRead(d, meta); err != nil {
		return nil, err
	}
	d.Set("notify_handle", "@pagerduty-" + d.Get("service_name").(string))
	return []*schema.ResourceData{d}, nil
}