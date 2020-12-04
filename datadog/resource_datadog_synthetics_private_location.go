package datadog

import (
	"encoding/json"
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogSyntheticsPrivateLocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSyntheticsPrivateLocationCreate,
		Read:   resourceDatadogSyntheticsPrivateLocationRead,
		Update: resourceDatadogSyntheticsPrivateLocationUpdate,
		Delete: resourceDatadogSyntheticsPrivateLocationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"config": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceDatadogSyntheticsPrivateLocationCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsPrivateLocation := buildSyntheticsPrivateLocationStruct(d)
	createdSyntheticsPrivateLocationResponse, _, err := datadogClientV1.SyntheticsApi.CreatePrivateLocation(authV1).Body(*syntheticsPrivateLocation).Execute()
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return translateClientError(err, "error creating synthetics private location")
	}

	createdSyntheticsPrivateLocation := createdSyntheticsPrivateLocationResponse.GetPrivateLocation()
	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsPrivateLocation.GetId())

	// set the config that is only returned when creating the private location
	conf, _ := json.Marshal(createdSyntheticsPrivateLocationResponse.GetConfig())
	d.Set("config", string(conf))

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsPrivateLocationRead(d, meta)
}

func resourceDatadogSyntheticsPrivateLocationRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsPrivateLocation, httpresp, err := datadogClientV1.SyntheticsApi.GetPrivateLocation(authV1, d.Id()).Execute()

	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting synthetics private location")
	}

	return updateSyntheticsPrivateLocationLocalState(d, &syntheticsPrivateLocation)
}

func resourceDatadogSyntheticsPrivateLocationUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsPrivateLocation := buildSyntheticsPrivateLocationStruct(d)
	if _, _, err := datadogClientV1.SyntheticsApi.UpdatePrivateLocation(authV1, d.Id()).Body(*syntheticsPrivateLocation).Execute(); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return translateClientError(err, "error updating synthetics private location")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsPrivateLocationRead(d, meta)
}

func resourceDatadogSyntheticsPrivateLocationDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, err := datadogClientV1.SyntheticsApi.DeletePrivateLocation(authV1, d.Id()).Execute(); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return translateClientError(err, "error deleting synthetics private location")
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func buildSyntheticsPrivateLocationStruct(d *schema.ResourceData) *datadogV1.SyntheticsPrivateLocation {
	syntheticsPrivateLocation := datadogV1.NewSyntheticsPrivateLocationWithDefaults()

	syntheticsPrivateLocation.SetName(d.Get("name").(string))

	if description, ok := d.GetOk("description"); ok {
		syntheticsPrivateLocation.SetDescription(description.(string))
	}

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
	}
	syntheticsPrivateLocation.SetTags(tags)

	return syntheticsPrivateLocation
}

func updateSyntheticsPrivateLocationLocalState(d *schema.ResourceData, syntheticsPrivateLocation *datadogV1.SyntheticsPrivateLocation) error {
	d.Set("name", syntheticsPrivateLocation.GetName())
	d.Set("description", syntheticsPrivateLocation.GetDescription())
	d.Set("tags", syntheticsPrivateLocation.Tags)

	return nil
}
