package datadog

import (
	"fmt"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogIntegrationAzure() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.",
		Create:      resourceDatadogIntegrationAzureCreate,
		Read:        resourceDatadogIntegrationAzureRead,
		Update:      resourceDatadogIntegrationAzureUpdate,
		Delete:      resourceDatadogIntegrationAzureDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationAzureImport,
		},

		Schema: map[string]*schema.Schema{
			"tenant_name": {
				Description: "Your Azure Active Directory ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"client_id": {
				Description: "Your Azure web application ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"client_secret": {
				Description: "(Required for Initial Creation) Your Azure web application secret key.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"host_filters": {
				Description: "String of host tag(s) (in the form `key:value,key:value`) defines a filter that Datadog will use when collecting metrics from Azure. Limit the Azure instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog. e.x. `env:production,deploymentgroup:red`",
				Type:        schema.TypeString,
				Optional:    true,
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
		return translateClientError(err, "error listing azure integration")
	}
	for _, integration := range integrations {
		if integration.GetTenantName() == tenantName {
			d.Set("tenant_name", integration.GetTenantName())
			d.Set("client_id", integration.GetClientId())
			hostFilters, exists := integration.GetHostFiltersOk()
			if exists == true {
				d.Set("host_filters", hostFilters)
			}
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
	clientID := d.Get("client_id").(string)

	iazure := buildDatadogAzureIntegrationDefinition(d, tenantName, clientID, false)

	if _, _, err := datadogClientV1.AzureIntegrationApi.CreateAzureIntegration(authV1).Body(*iazure).Execute(); err != nil {
		return translateClientError(err, "error creating an Azure integration")
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

	iazure := buildDatadogAzureIntegrationDefinition(d, existingTenantName, existingClientID, true)

	if _, _, err := datadogClientV1.AzureIntegrationApi.UpdateAzureIntegration(authV1).Body(*iazure).Execute(); err != nil {
		return translateClientError(err, "error updating an Azure integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", iazure.GetNewTenantName(), iazure.GetNewClientId()))

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
	iazure := buildDatadogAzureIntegrationDefinition(d, tenantName, clientID, false)

	if _, _, err := datadogClientV1.AzureIntegrationApi.DeleteAzureIntegration(authV1).Body(*iazure).Execute(); err != nil {
		return translateClientError(err, "error deleting an Azure integration")
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

func buildDatadogAzureIntegrationDefinition(terraformDefinition *schema.ResourceData, tenantName string, clientID string, update bool) *datadogV1.AzureAccount {
	datadogDefinition := datadogV1.NewAzureAccount()
	// Required params
	datadogDefinition.SetTenantName(tenantName)
	datadogDefinition.SetClientId(clientID)
	// Optional params
	hostFilters, exists := terraformDefinition.GetOk("host_filters")
	if exists == true {
		datadogDefinition.SetHostFilters(hostFilters.(string))
	}
	clientSecret, exists := terraformDefinition.GetOk("client_secret")
	if exists == true {
		datadogDefinition.SetClientSecret(clientSecret.(string))
	}
	// Only do the following if building for the Update
	if update == true {
		newTenantName, exists := terraformDefinition.GetOk("tenant_name")
		if exists == true {
			datadogDefinition.SetNewTenantName(newTenantName.(string))
		}
		newClientID, exists := terraformDefinition.GetOk("client_id")
		if exists == true {
			datadogDefinition.SetNewClientId(newClientID.(string))
		}
	}
	return datadogDefinition
}
