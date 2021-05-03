package datadog

import (
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogIntegrationGcp() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog - Google Cloud Platform integration resource. This can be used to create and manage Datadog - Google Cloud Platform integration.",
		Create:      resourceDatadogIntegrationGcpCreate,
		Read:        resourceDatadogIntegrationGcpRead,
		Update:      resourceDatadogIntegrationGcpUpdate,
		Delete:      resourceDatadogIntegrationGcpDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationGcpImport,
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

const (
	defaultType                    = "service_account"
	defaultAuthURI                 = "https://accounts.google.com/o/oauth2/auth"
	defaultTokenURI                = "https://accounts.google.com/o/oauth2/token"
	defaultAuthProviderX509CertURL = "https://www.googleapis.com/oauth2/v1/certs"
	defaultClientX509CertURLPrefix = "https://www.googleapis.com/robot/v1/metadata/x509/"
)

func resourceDatadogIntegrationGcpCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	projectID := d.Get("project_id").(string)

	if _, _, err := datadogClientV1.GCPIntegrationApi.CreateGCPIntegration(authV1,
		datadogV1.GCPAccount{
			Type:                    datadogV1.PtrString(defaultType),
			ProjectId:               datadogV1.PtrString(projectID),
			PrivateKeyId:            datadogV1.PtrString(d.Get("private_key_id").(string)),
			PrivateKey:              datadogV1.PtrString(d.Get("private_key").(string)),
			ClientEmail:             datadogV1.PtrString(d.Get("client_email").(string)),
			ClientId:                datadogV1.PtrString(d.Get("client_id").(string)),
			AuthUri:                 datadogV1.PtrString(defaultAuthURI),
			TokenUri:                datadogV1.PtrString(defaultTokenURI),
			AuthProviderX509CertUrl: datadogV1.PtrString(defaultAuthProviderX509CertURL),
			ClientX509CertUrl:       datadogV1.PtrString(defaultClientX509CertURLPrefix + d.Get("client_email").(string)),
			HostFilters:             datadogV1.PtrString(d.Get("host_filters").(string)),
			Automute:                datadogV1.PtrBool(d.Get("automute").(bool)),
		},
	); err != nil {
		return utils.TranslateClientError(err, "error creating GCP integration")
	}

	d.SetId(projectID)

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	projectID := d.Id()

	integrations, _, err := datadogClientV1.GCPIntegrationApi.ListGCPIntegration(authV1)
	if err != nil {
		return utils.TranslateClientError(err, "error getting GCP integration")
	}
	for _, integration := range integrations {
		if integration.GetProjectId() == projectID {
			d.Set("project_id", integration.GetProjectId())
			d.Set("client_email", integration.GetClientEmail())
			d.Set("host_filters", integration.GetHostFilters())
			d.Set("automute", integration.GetAutomute())
			return nil
		}
	}
	d.SetId("")
	return nil
}

func resourceDatadogIntegrationGcpUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.GCPIntegrationApi.UpdateGCPIntegration(authV1,
		datadogV1.GCPAccount{
			ProjectId:   datadogV1.PtrString(d.Id()),
			ClientEmail: datadogV1.PtrString(d.Get("client_email").(string)),
			HostFilters: datadogV1.PtrString(d.Get("host_filters").(string)),
			Automute:    datadogV1.PtrBool(d.Get("automute").(bool)),
		},
	); err != nil {
		return utils.TranslateClientError(err, "error updating GCP integration")
	}

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.GCPIntegrationApi.DeleteGCPIntegration(authV1,
		datadogV1.GCPAccount{
			ProjectId:   datadogV1.PtrString(d.Id()),
			ClientEmail: datadogV1.PtrString(d.Get("client_email").(string)),
		},
	); err != nil {
		return utils.TranslateClientError(err, "error deleting GCP integration")
	}

	return nil
}

func resourceDatadogIntegrationGcpImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationGcpRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
