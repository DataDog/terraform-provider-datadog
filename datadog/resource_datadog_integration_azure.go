package datadog

import (
	"fmt"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogIntegrationAzure() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationAzureCreate,
		Read:   resourceDatadogIntegrationAzureRead,
		Update: resourceDatadogIntegrationAzureUpdate,
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
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tenantName, _, err := tenantAndClientFromID(d.Id())
	if err != nil {
		return err
	}

	integrations, _, err := datadogClientV1.AzureIntegrationApi.ListAzureIntegration(authV1).Execute()
	if err != nil {
		return err
	}
	for _, integration := range integrations {
		if integration.GetTenantName() == tenantName {
			d.Set("tenant_name", integration.GetTenantName())
			d.Set("client_id", integration.GetClientId())
			d.Set("host_filters", integration.GetHostFilters())
			return nil
		}
	}
	return fmt.Errorf("error getting an Azure integration: tenant_name=%s", tenantName)
}

func resourceDatadogIntegrationAzureCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tenantName := d.Get("tenant_name").(string)
	iazure := datadogV1.NewAzureAccount()
	iazure.SetTenantName(tenantName)
	iazure.SetClientId(d.Get("client_id").(string))
	iazure.SetClientSecret(d.Get("client_secret").(string))
	iazure.SetHostFilters(d.Get("host_filters").(string))

	if _, _, err := datadogClientV1.AzureIntegrationApi.CreateAzureIntegration(authV1).Body(*iazure).Execute(); err != nil {
		return fmt.Errorf("error creating an Azure integration: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", iazure.GetTenantName(), iazure.GetClientId()))

	return resourceDatadogIntegrationAzureRead(d, meta)
}

func resourceDatadogIntegrationAzureUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	existingTenantName, existingClientID, err := tenantAndClientFromID(d.Id())
	if err != nil {
		return err
	}
	newTenantName := d.Get("tenant_name").(string)
	newClientID := d.Get("client_id").(string)

	iazure := datadogV1.NewAzureAccount()
	iazure.SetTenantName(existingTenantName)
	iazure.SetClientId(existingClientID)
	iazure.SetNewTenantName(newTenantName)
	iazure.SetNewClientId(newClientID)
	iazure.SetHostFilters(d.Get("host_filters").(string))
	iazure.SetClientSecret(d.Get("client_secret").(string))

	if _, _, err := datadogClientV1.AzureIntegrationApi.UpdateAzureIntegration(authV1).Body(*iazure).Execute(); err != nil {
		return fmt.Errorf("error updating an Azure integration: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%s:%s", newTenantName, newClientID))

	return resourceDatadogIntegrationAzureRead(d, meta)
}

func resourceDatadogIntegrationAzureDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tenantName, clientID, err := tenantAndClientFromID(d.Id())
	if err != nil {
		return err
	}
	iazure := datadogV1.NewAzureAccount()
	iazure.SetTenantName(tenantName)
	iazure.SetClientId(clientID)

	if _, _, err := datadogClientV1.AzureIntegrationApi.DeleteAzureIntegration(authV1).Body(*iazure).Execute(); err != nil {
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

func tenantAndClientFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting tenant name and client ID from an Azure integration id: %s", id)
	}
	return result[0], result[1], nil
}
