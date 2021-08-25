package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationAzure() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog - Microsoft Azure integration resource. This can be used to create and manage the integrations.",
		CreateContext: resourceDatadogIntegrationAzureCreate,
		ReadContext:   resourceDatadogIntegrationAzureRead,
		UpdateContext: resourceDatadogIntegrationAzureUpdate,
		DeleteContext: resourceDatadogIntegrationAzureDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceDatadogIntegrationAzureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tenantName, _, err := utils.TenantAndClientFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	integrations, httpresp, err := datadogClientV1.AzureIntegrationApi.ListAzureIntegration(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error listing azure integration")
	}
	if err := utils.CheckForUnparsed(integrations); err != nil {
		return diag.FromErr(err)
	}
	for _, integration := range integrations {
		if integration.GetTenantName() == tenantName {
			d.Set("tenant_name", integration.GetTenantName())
			d.Set("client_id", integration.GetClientId())
			hostFilters, exists := integration.GetHostFiltersOk()
			if exists {
				d.Set("host_filters", hostFilters)
			}
			return nil
		}
	}
	return diag.Errorf("error getting an Azure integration: tenant_name=%s", tenantName)
}

func resourceDatadogIntegrationAzureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tenantName := d.Get("tenant_name").(string)
	clientID := d.Get("client_id").(string)

	iazure := buildDatadogAzureIntegrationDefinition(d, tenantName, clientID, false)

	if _, httpresp, err := datadogClientV1.AzureIntegrationApi.CreateAzureIntegration(authV1, *iazure); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating an Azure integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", iazure.GetTenantName(), iazure.GetClientId()))

	return resourceDatadogIntegrationAzureRead(ctx, d, meta)
}

func resourceDatadogIntegrationAzureUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	existingTenantName, existingClientID, err := utils.TenantAndClientFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	iazure := buildDatadogAzureIntegrationDefinition(d, existingTenantName, existingClientID, true)

	if _, httpresp, err := datadogClientV1.AzureIntegrationApi.UpdateAzureIntegration(authV1, *iazure); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating an Azure integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", iazure.GetNewTenantName(), iazure.GetNewClientId()))

	return resourceDatadogIntegrationAzureRead(ctx, d, meta)
}

func resourceDatadogIntegrationAzureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tenantName, clientID, err := utils.TenantAndClientFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	iazure := buildDatadogAzureIntegrationDefinition(d, tenantName, clientID, false)

	if _, httpresp, err := datadogClientV1.AzureIntegrationApi.DeleteAzureIntegration(authV1, *iazure); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting an Azure integration")
	}

	return nil
}

func buildDatadogAzureIntegrationDefinition(terraformDefinition *schema.ResourceData, tenantName string, clientID string, update bool) *datadogV1.AzureAccount {
	datadogDefinition := datadogV1.NewAzureAccount()
	// Required params
	datadogDefinition.SetTenantName(tenantName)
	datadogDefinition.SetClientId(clientID)
	// Optional params
	hostFilters := terraformDefinition.Get("host_filters")
	datadogDefinition.SetHostFilters(hostFilters.(string))

	clientSecret, exists := terraformDefinition.GetOk("client_secret")
	if exists {
		datadogDefinition.SetClientSecret(clientSecret.(string))
	}
	// Only do the following if building for the Update
	if update {
		newTenantName, exists := terraformDefinition.GetOk("tenant_name")
		if exists {
			datadogDefinition.SetNewTenantName(newTenantName.(string))
		}
		newClientID, exists := terraformDefinition.GetOk("client_id")
		if exists {
			datadogDefinition.SetNewClientId(newClientID.(string))
		}
	}
	return datadogDefinition
}
