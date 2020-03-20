package datadog

import (
	"log"
	"strings"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogUserCreate,
		Read:   resourceDatadogUserRead,
		Update: resourceDatadogUserUpdate,
		Delete: resourceDatadogUserDelete,
		Exists: resourceDatadogUserExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogUserImport,
		},

		Schema: map[string]*schema.Schema{
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"handle": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_admin": {
				Type:       schema.TypeBool,
				Computed:   true,
				Optional:   true,
				Deprecated: "This parameter has been replaced by `access_role` and has no effect",
			},
			"access_role": {
				Type:     schema.TypeString,
				Optional: true,
				Required: false,
				Default:  "st",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "This parameter was removed from the API and has no effect",
			},
			"verified": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceDatadogUserExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if _, _, err := client.UsersApi.GetUser(auth, d.Id()).Execute(); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking user exists")
	}

	return true, nil
}

func buildDatadogUserStruct(d *schema.ResourceData) *datadog.User {
	var u datadog.User
	u.SetDisabled(d.Get("disabled").(bool))
	u.SetEmail(d.Get("email").(string))
	u.SetHandle(d.Get("handle").(string))
	//u.SetIsAdmin(d.Get("is_admin").(bool))
	u.SetName(d.Get("name").(string))
	u.SetAccessRole(datadog.AccessRole(d.Get("access_role").(string)))

	return &u
}

func resourceDatadogUserCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	u := buildDatadogUserStruct(d)

	// Datadog does not actually delete users, so CreateUser might return a 409.
	// We ignore that case and proceed, likely re-enabling the user.
	if _, httpresp, err := client.UsersApi.CreateUser(auth).Body(*u).Execute(); err != nil {
		if httpresp.StatusCode != 409 {
			return translateClientError(err, "error creating user")
		}
		log.Printf("[INFO] Updating existing Datadog user %s", *u.Handle)
	}

	resp, _, err := client.UsersApi.UpdateUser(auth, d.Get("handle").(string)).Body(*u).Execute()
	if err != nil {
		return translateClientError(err, "error updating user")
	}
	user := resp.GetUser()
	d.SetId(user.GetHandle())

	return resourceDatadogUserRead(d, meta)
}

func resourceDatadogUserRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	resp, _, err := client.UsersApi.GetUser(auth, d.Id()).Execute()
	if err != nil {
		return err
	}

	u := resp.GetUser()
	d.Set("disabled", u.GetDisabled())
	d.Set("email", u.GetEmail())
	d.Set("handle", u.GetHandle())
	d.Set("name", u.GetName())
	d.Set("verified", u.GetVerified())
	d.Set("access_role", u.GetAccessRole())
	//d.Set("is_admin", u.GetIsAdmin())
	return nil
}

func resourceDatadogUserUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	u := buildDatadogUserStruct(d)
	u.SetHandle(d.Id())

	if _, _, err := client.UsersApi.UpdateUser(auth, u.GetHandle()).Body(*u).Execute(); err != nil {
		return translateClientError(err, "error updating user")
	}

	return resourceDatadogUserRead(d, meta)
}

func resourceDatadogUserDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	// Datadog does not actually delete users, but instead marks them as disabled.
	// Bypass DeleteUser if GetUser returns User.Disabled == true, otherwise it will 400.
	resp, _, err := client.UsersApi.GetUser(auth, d.Id()).Execute()
	u := resp.GetUser()
	if err == nil && u.GetDisabled() {
		return nil
	}

	if _, _, err := client.UsersApi.DisableUser(auth, d.Id()).Execute(); err != nil {
		return translateClientError(err, "error deleting user")
	}

	return nil
}

func resourceDatadogUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogUserRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
