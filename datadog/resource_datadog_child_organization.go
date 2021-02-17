package datadog

import (
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceDatadogChildOrganisation() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog Organization resource. This can be used to create and manage Datadog organizations.",
		Create:      resourceDatadogChildOrganizationCreate,
		Read:        resourceDatadogChildOrganizationRead,
		Update:      resourceDatadogChildOrganizationUpdate,
		Delete:      resourceDatadogChildOrganizationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogOrganizationImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "The name of the organization",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"api_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"application_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
		},
	}
}

func resourceDatadogChildOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	var userRequest datadogV1.OrganizationCreateBody
	if v, ok := d.GetOk("name"); ok {
		userRequest.SetName(v.(string))
	}

	// Datadog does not actually delete users, so CreateUser might return a 409.
	// We ignore that case and proceed, likely re-enabling the user.
	resp, _, err := datadogClientV1.OrganizationsApi.CreateChildOrg(authV1).Body(userRequest).Execute()
	if err != nil {
		return translateClientError(err, "error creating child organization")
	}
	ordData := resp.GetOrg()
	d.SetId(ordData.GetPublicId())
	d.Set("api_key", resp.GetApiKey().Key)
	d.Set("application_key", resp.GetApplicationKey().Hash)

	return resourceDatadogChildOrganizationRead(d, meta)
}

func resourceDatadogChildOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	resp, httpresp, err := client.OrganizationsApi.GetOrg(auth, d.Id()).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting organization")
	}
	orgData := resp.GetOrg()
	d.Set("name", orgData.GetName())

	return nil
}

func resourceDatadogChildOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	if d.HasChange("name") {
		var userRequest datadogV1.Organization
		if v, ok := d.GetOk("name"); ok {
			userRequest.SetName(v.(string))
		}
		_, _, err := client.OrganizationsApi.UpdateOrg(auth, d.Id()).Body(userRequest).Execute()
		if err != nil {
			return translateClientError(err, "error updating role")
		}
	}

	return resourceDatadogChildOrganizationRead(d, meta)
}

func resourceDatadogChildOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogOrganizationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogUserRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
