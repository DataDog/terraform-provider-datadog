package datadog

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogAppKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogAppKeyCreate,
		Read:   resourceDatadogAppKeyRead,
		Update: resourceDatadogAppKeyUpdate,
		Delete: resourceDatadogAppKeyDelete,
		Exists: resourceDatadogAppKeyExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"hash": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceDatadogAppKeyExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	if _, err := client.GetAPPKey(d.Id()); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking app key exists")
	}

	return true, nil
}

func resourceDatadogAppKeyCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	key, err := client.CreateAPPKey(d.Get("name").(string))
	if err != nil {
		return translateClientError(err, "error creating app key")
	}

	d.SetId(key.GetHash())

	return resourceDatadogAppKeyRead(d, meta)
}

func resourceDatadogAppKeyRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	k, err := client.GetAPPKey(d.Id())
	if err != nil {
		return err
	}

	d.Set("name", k.GetName())
	d.Set("hash", k.GetHash())
	return nil
}

func resourceDatadogAppKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	k := &datadog.APPKey{}
	k.SetHash(d.Id())
	k.SetName(d.Get("name").(string))

	if err := client.UpdateAPPKey(k); err != nil {
		return translateClientError(err, "error updating app key")
	}

	return resourceDatadogAppKeyRead(d, meta)
}

func resourceDatadogAppKeyDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	if err := client.DeleteAPPKey(d.Id()); err != nil {
		return translateClientError(err, "error deleting app key")
	}

	return nil
}
