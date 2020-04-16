package datadog

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogAPIKeyCreate,
		Read:   resourceDatadogAPIKeyRead,
		Update: resourceDatadogAPIKeyUpdate,
		Delete: resourceDatadogAPIKeyDelete,
		Exists: resourceDatadogAPIKeyExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceDatadogAPIKeyExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	if _, err := client.GetAPIKey(d.Id()); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking API key exists")
	}

	return true, nil
}

func resourceDatadogAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	key, err := client.CreateAPIKey(d.Get("name").(string))
	if err != nil {
		return translateClientError(err, "error creating API key")
	}

	d.SetId(key.GetKey())

	return resourceDatadogAPIKeyRead(d, meta)
}

func resourceDatadogAPIKeyRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	k, err := client.GetAPIKey(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", k.GetName())
	d.Set("key", k.GetKey())
	return nil
}

func resourceDatadogAPIKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	k := &datadog.APIKey{}
	k.SetKey(d.Id())
	k.SetName(d.Get("name").(string))

	if err := client.UpdateAPIKey(k); err != nil {
		return translateClientError(err, "error updating API key")
	}

	return resourceDatadogAPIKeyRead(d, meta)
}

func resourceDatadogAPIKeyDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	if err := client.DeleteAPIKey(d.Id()); err != nil {
		return translateClientError(err, "error deleting API key")
	}

	return nil
}
