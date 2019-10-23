package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogIntegrationAzure() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationAzureCreate,
		Read:   resourceDatadogIntegrationAzureRead,
		Delete: resourceDatadogIntegrationAzureDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationAzureImport,
		},

		Schema: map[string]*schema.Schema{
			"tenant_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"host_filters": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceDatadogIntegrationAzureRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	tenantName := d.Id()

	integrations, err := client.ListIntegrationAzure()
	if err != nil {
		return err
	}
	for _, integration := range integrations {
		if integration.GetTenantName() == tenantName {
			d.Set("tenant_name", integration.GetTenantName())
			d.Set("client_id", integration.GetClientID())
			d.Set("host_filters", integration.GetHostFilters())
			return nil
		}
	}
	return fmt.Errorf("error getting an Azure integration: tenant_name=%s", tenantName)
}

func resourceDatadogIntegrationAzureCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	tenantName := d.Get("tenant_name").(string)

	if err := client.CreateIntegrationAzure(
		&datadog.IntegrationAzure{
			TenantName:   datadog.String(tenantName),
			ClientID:     datadog.String(d.Get("client_id").(string)),
			ClientSecret: datadog.String(d.Get("client_secret").(string)),
			HostFilters:  datadog.String(d.Get("host_filters").(string)),
		},
	); err != nil {
		return fmt.Errorf("error creating an Azure integration: %s", err.Error())
	}

	d.SetId(tenantName)

	return resourceDatadogIntegrationAzureRead(d, meta)
}

func resourceDatadogIntegrationAzureDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteIntegrationAzure(
		&datadog.IntegrationAzure{
			TenantName: datadog.String(d.Id()),
			ClientID:   datadog.String(d.Get("client_id").(string)),
		},
	); err != nil {
		return fmt.Errorf("error deleting an Azure integration: %s", err.Error())
	}

	return nil
}

func resourceDatadogIntegrationAzureImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationAzureRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
