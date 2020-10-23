package datadog

import (
	"fmt"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogIntegrationGcp() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationGcpCreate,
		Read:   resourceDatadogIntegrationGcpRead,
		Update: resourceDatadogIntegrationGcpUpdate,
		Delete: resourceDatadogIntegrationGcpDelete,
		Exists: resourceDatadogIntegrationGcpExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationGcpImport,
		},

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_key_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"client_email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host_filters": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceDatadogIntegrationGcpExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrations, _, err := datadogClientV1.GCPIntegrationApi.ListGCPIntegration(authV1).Execute()
	if err != nil {
		return false, translateClientError(err, "error checking GCP integration exists")
	}
	projectID := d.Id()
	for _, integration := range integrations {
		if integration.GetProjectId() == projectID {
			return true, nil
		}
	}
	return false, nil
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

	if _, _, err := datadogClientV1.GCPIntegrationApi.CreateGCPIntegration(authV1).Body(
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
		},
	).Execute(); err != nil {
		return translateClientError(err, "error creating GCP integration")
	}

	d.SetId(projectID)

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	projectID := d.Id()

	integrations, _, err := datadogClientV1.GCPIntegrationApi.ListGCPIntegration(authV1).Execute()
	if err != nil {
		return translateClientError(err, "error getting GCP integration")
	}
	for _, integration := range integrations {
		if integration.GetProjectId() == projectID {
			d.Set("project_id", integration.GetProjectId())
			d.Set("client_email", integration.GetClientEmail())
			d.Set("host_filters", integration.GetHostFilters())
			return nil
		}
	}
	return fmt.Errorf("error getting a Google Cloud Platform integration: project_id=%s", projectID)
}

func resourceDatadogIntegrationGcpUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.GCPIntegrationApi.UpdateGCPIntegration(authV1).Body(
		datadogV1.GCPAccount{
			ProjectId:   datadogV1.PtrString(d.Id()),
			ClientEmail: datadogV1.PtrString(d.Get("client_email").(string)),
			HostFilters: datadogV1.PtrString(d.Get("host_filters").(string)),
		},
	).Execute(); err != nil {
		return translateClientError(err, "error updating GCP integration")
	}

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.GCPIntegrationApi.DeleteGCPIntegration(authV1).Body(
		datadogV1.GCPAccount{
			ProjectId:   datadogV1.PtrString(d.Id()),
			ClientEmail: datadogV1.PtrString(d.Get("client_email").(string)),
		},
	).Execute(); err != nil {
		return translateClientError(err, "error deleting GCP integration")
	}

	return nil
}

func resourceDatadogIntegrationGcpImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationGcpRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
