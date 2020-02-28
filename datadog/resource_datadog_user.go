package datadog

import (
	"fmt"
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
		return false, err
	}

	return true, nil
}

func resourceDatadogUserCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	var userCreate datadog.User
	userCreate.SetEmail(d.Get("email").(string))
	userCreate.SetHandle(d.Get("handle").(string))
	userCreate.SetName(d.Get("name").(string))
	userCreate.SetAccessRole(d.Get("access_role").(datadog.AccessRole))

	var userUpdate datadog.User
	userUpdate.SetEmail(d.Get("email").(string))
	userUpdate.SetName(d.Get("name").(string))
	userUpdate.SetAccessRole(d.Get("access_role").(datadog.AccessRole))
	userUpdate.SetDisabled(d.Get("disabled").(bool))

	// Datadog does not actually delete users, so CreateUser might return a 409.
	// We ignore that case and proceed, likely re-enabling the user.
	if _, _, err := client.UsersApi.CreateUser(auth).Body(userCreate).Execute(); err != nil {
		if !strings.Contains(err.Error(), "API error 409 Conflict") {
			return fmt.Errorf("error creating user: %s", err.Error())
		}
		log.Printf("[INFO] Updating existing Datadog user %s", *userCreate.Handle)
	}

	res, _, err := client.UsersApi.UpdateUser(auth, d.Get("handle").(string)).Body(userUpdate).Execute()

	if err != nil {
		return fmt.Errorf("error creating user: %s", err.Error())
	}

	u := res.GetUser()
	d.SetId(u.GetHandle())

	return resourceDatadogUserRead(d, meta)
}

func resourceDatadogUserRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	userResponse, _, err := client.UsersApi.GetUser(auth, d.Id()).Execute()
	if err != nil {
		return err
	}
	u := userResponse.GetUser()
	d.Set("disabled", u.GetDisabled())
	d.Set("email", u.GetEmail())
	d.Set("handle", u.GetHandle())
	d.Set("name", u.GetName())
	d.Set("verified", u.GetVerified())
	d.Set("access_role", u.GetAccessRole())
	return nil
}

func resourceDatadogUserUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	var userUpdate datadog.User
	userUpdate.SetEmail(d.Get("email").(string))
	userUpdate.SetName(d.Get("name").(string))
	userUpdate.SetAccessRole(d.Get("access_role").(datadog.AccessRole))
	userUpdate.SetDisabled(d.Get("disabled").(bool))

	if _, _, err := client.UsersApi.UpdateUser(auth, d.Get("handle").(string)).Body(userUpdate).Execute(); err != nil {
		return fmt.Errorf("error updating user: %s", err.Error())
	}

	return resourceDatadogUserRead(d, meta)
}

func resourceDatadogUserDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	// Datadog does not actually delete users, but instead marks them as disabled.
	// Bypass DeleteUser if GetUser returns User.Disabled == true, otherwise it will 400.
	userResponse, _, err := client.UsersApi.GetUser(auth, d.Id()).Execute()
	u := userResponse.GetUser()
	if err == nil && u.GetDisabled() {
		return nil
	}

	if _, _, err := client.UsersApi.DisableUser(auth, d.Id()).Execute(); err != nil {
		return err
	}

	return nil
}

func resourceDatadogUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogUserRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
