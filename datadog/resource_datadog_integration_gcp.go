package datadog

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
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
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrations, _, err := client.GCPIntegrationApi.ListGCPIntegration(auth).Execute()
	if err != nil {
		return false, err
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
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	projectID := d.Get("project_id").(string)

	if _, _, err := client.GCPIntegrationApi.CreateGCPIntegration(auth).Body(
		datadog.GCPAccount{
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
		},
	).Execute(); err != nil {
		return translateClientError(err, "error creating a Google Cloud Platform integration")
	}

	d.SetId(projectID)

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	projectID := d.Id()

	integrations, _, err := client.GCPIntegrationApi.ListGCPIntegration(auth).Execute()
	if err != nil {
		return err
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
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if _, _, err := client.GCPIntegrationApi.UpdateGCPIntegration(auth).Body(
		datadog.GCPAccount{
			ProjectId:   datadog.PtrString(d.Id()),
			ClientEmail: datadog.PtrString(d.Get("client_email").(string)),
			HostFilters: datadog.PtrString(d.Get("host_filters").(string)),
		},
	).Execute(); err != nil {
		return translateClientError(err, "error updating a Google Cloud Platform integration")
	}

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if _, _, err := client.GCPIntegrationApi.DeleteGCPIntegration(auth).Body(
		datadog.GCPAccount{
			ProjectId:   datadog.PtrString(d.Id()),
			ClientEmail: datadog.PtrString(d.Get("client_email").(string)),
		},
	).Execute(); err != nil {
		return translateClientError(err, "error deleting a Google Cloud Platform integration")
	}

	return nil
}

func resourceDatadogIntegrationGcpImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationGcpRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
