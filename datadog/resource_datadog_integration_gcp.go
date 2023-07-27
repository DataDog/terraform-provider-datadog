package datadog

import (
	"context"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var integrationGcpMutex = sync.Mutex{}

func resourceDatadogIntegrationGcp() *schema.Resource {
	return &schema.Resource{
		Description:        "This resource is deprecated — use the `datadog_integration_gcp_sts resource` instead. Provides a Datadog - Google Cloud Platform integration resource. This can be used to create and manage Datadog - Google Cloud Platform integration.",
		DeprecationMessage: "This resource is deprecated — use the datadog_integration_gcp_sts resource instead.",
		CreateContext:      resourceDatadogIntegrationGcpCreate,
		ReadContext:        resourceDatadogIntegrationGcpRead,
		UpdateContext:      resourceDatadogIntegrationGcpUpdate,
		DeleteContext:      resourceDatadogIntegrationGcpDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"project_id": {
					Description: "Your Google Cloud project ID found in your JSON service account key.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"private_key_id": {
					Description: "Your private key ID found in your JSON service account key.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"private_key": {
					Description: "Your private key name found in your JSON service account key.",
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					ForceNew:    true,
				},
				"client_email": {
					Description: "Your email found in your JSON service account key.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"client_id": {
					Description: "Your ID found in your JSON service account key.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"host_filters": {
					Description: "Limit the GCE instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"automute": {
					Description: "Silence monitors for expected GCE instance shutdowns.",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
				"cspm_resource_collection_enabled": {
					Description: "Whether Datadog collects cloud security posture management resources from your GCP project.",
					Type:        schema.TypeBool,
					Default:     false,
					Optional:    true,
				},
			}
		},
	}
}

const (
	defaultType                    = "service_account"
	defaultAuthURI                 = "https://accounts.google.com/o/oauth2/auth"
	defaultTokenURI                = "https://oauth2.googleapis.com/token"
	defaultAuthProviderX509CertURL = "https://www.googleapis.com/oauth2/v1/certs"
	defaultClientX509CertURLPrefix = "https://www.googleapis.com/robot/v1/metadata/x509/"
)

func resourceDatadogIntegrationGcpCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationGcpMutex.Lock()
	defer integrationGcpMutex.Unlock()

	projectID := d.Get("project_id").(string)

	if _, httpresp, err := apiInstances.GetGCPIntegrationApiV1().CreateGCPIntegration(auth,
		datadogV1.GCPAccount{
			Type:                    datadog.PtrString(defaultType),
			ProjectId:               datadog.PtrString(projectID),
			PrivateKeyId:            datadog.PtrString(d.Get("private_key_id").(string)),
			PrivateKey:              datadog.PtrString(d.Get("private_key").(string)),
			ClientEmail:             datadog.PtrString(d.Get("client_email").(string)),
			ClientId:                datadog.PtrString(d.Get("client_id").(string)),
			AuthUri:                 datadog.PtrString(defaultAuthURI),
			TokenUri:                datadog.PtrString(defaultTokenURI),
			AuthProviderX509CertUrl: datadog.PtrString(defaultAuthProviderX509CertURL),
			ClientX509CertUrl:       datadog.PtrString(defaultClientX509CertURLPrefix + d.Get("client_email").(string)),
			HostFilters:             datadog.PtrString(d.Get("host_filters").(string)),
			Automute:                datadog.PtrBool(d.Get("automute").(bool)),
			IsCspmEnabled:           datadog.PtrBool(d.Get("cspm_resource_collection_enabled").(bool)),
		},
	); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating GCP integration")
	}

	d.SetId(projectID)

	return resourceDatadogIntegrationGcpRead(ctx, d, meta)
}

func resourceDatadogIntegrationGcpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	projectID := d.Id()

	integrations, httpresp, err := apiInstances.GetGCPIntegrationApiV1().ListGCPIntegration(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting GCP integration")
	}
	if err := utils.CheckForUnparsed(integrations); err != nil {
		return diag.FromErr(err)
	}
	for _, integration := range integrations {
		if integration.GetProjectId() == projectID {
			d.Set("project_id", integration.GetProjectId())
			d.Set("client_email", integration.GetClientEmail())
			d.Set("host_filters", integration.GetHostFilters())
			d.Set("automute", integration.GetAutomute())
			d.Set("cspm_resource_collection_enabled", integration.GetIsCspmEnabled())
			return nil
		}
	}
	d.SetId("")
	return nil
}

func resourceDatadogIntegrationGcpUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationGcpMutex.Lock()
	defer integrationGcpMutex.Unlock()

	if _, httpresp, err := apiInstances.GetGCPIntegrationApiV1().UpdateGCPIntegration(auth,
		datadogV1.GCPAccount{
			ProjectId:     datadog.PtrString(d.Id()),
			ClientEmail:   datadog.PtrString(d.Get("client_email").(string)),
			HostFilters:   datadog.PtrString(d.Get("host_filters").(string)),
			Automute:      datadog.PtrBool(d.Get("automute").(bool)),
			IsCspmEnabled: datadog.PtrBool(d.Get("cspm_resource_collection_enabled").(bool)),
		},
	); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating GCP integration")
	}

	return resourceDatadogIntegrationGcpRead(ctx, d, meta)
}

func resourceDatadogIntegrationGcpDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationGcpMutex.Lock()
	defer integrationGcpMutex.Unlock()

	if _, httpresp, err := apiInstances.GetGCPIntegrationApiV1().DeleteGCPIntegration(auth,
		datadogV1.GCPAccount{
			ProjectId:   datadog.PtrString(d.Id()),
			ClientEmail: datadog.PtrString(d.Get("client_email").(string)),
		},
	); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting GCP integration")
	}

	return nil
}
