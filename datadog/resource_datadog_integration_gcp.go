package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
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
	client := meta.(*datadog.Client)

	integrations, err := client.ListIntegrationGCP()
	if err != nil {
		return false, err
	}
	projectID := d.Id()
	for _, integration := range integrations {
		if integration.GetProjectID() == projectID {
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
	client := meta.(*datadog.Client)

	projectID := d.Get("project_id").(string)

	if err := client.CreateIntegrationGCP(
		&datadog.IntegrationGCPCreateRequest{
			Type:                    datadog.String(defaultType),
			ProjectID:               datadog.String(projectID),
			PrivateKeyID:            datadog.String(d.Get("private_key_id").(string)),
			PrivateKey:              datadog.String(d.Get("private_key").(string)),
			ClientEmail:             datadog.String(d.Get("client_email").(string)),
			ClientID:                datadog.String(d.Get("client_id").(string)),
			AuthURI:                 datadog.String(defaultAuthURI),
			TokenURI:                datadog.String(defaultTokenURI),
			AuthProviderX509CertURL: datadog.String(defaultAuthProviderX509CertURL),
			ClientX509CertURL:       datadog.String(defaultClientX509CertURLPrefix + d.Get("client_email").(string)),
			HostFilters:             datadog.String(d.Get("host_filters").(string)),
		},
	); err != nil {
		return fmt.Errorf("error creating a Google Cloud Platform integration: %s", err.Error())
	}

	d.SetId(projectID)

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	projectID := d.Id()

	integrations, err := client.ListIntegrationGCP()
	if err != nil {
		return err
	}
	for _, integration := range integrations {
		if integration.GetProjectID() == projectID {
			d.Set("project_id", integration.GetProjectID())
			d.Set("client_email", integration.GetClientEmail())
			d.Set("host_filters", integration.GetHostFilters())
			return nil
		}
	}
	return fmt.Errorf("error getting a Google Cloud Platform integration: project_id=%s", projectID)
}

func resourceDatadogIntegrationGcpUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.UpdateIntegrationGCP(
		&datadog.IntegrationGCPUpdateRequest{
			ProjectID:   datadog.String(d.Id()),
			ClientEmail: datadog.String(d.Get("client_email").(string)),
			HostFilters: datadog.String(d.Get("host_filters").(string)),
		},
	); err != nil {
		return fmt.Errorf("error updating a Google Cloud Platform integration: %s", err.Error())
	}

	return resourceDatadogIntegrationGcpRead(d, meta)
}

func resourceDatadogIntegrationGcpDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteIntegrationGCP(
		&datadog.IntegrationGCPDeleteRequest{
			ProjectID:   datadog.String(d.Id()),
			ClientEmail: datadog.String(d.Get("client_email").(string)),
		},
	); err != nil {
		return fmt.Errorf("error deleting a Google Cloud Platform integration: %s", err.Error())
	}

	return nil
}

func resourceDatadogIntegrationGcpImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationGcpRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
