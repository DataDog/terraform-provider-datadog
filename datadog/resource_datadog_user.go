package datadog

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

var uuidRegex = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")

func isV2User(id string) bool {
	return uuidRegex.MatchString(id)
}

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
				Optional: true,
			},
			"title": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"is_admin": {
				Type:       schema.TypeBool,
				Computed:   true,
				Optional:   true,
				Deprecated: "This parameter will be replaced by `access_role` and will be removed from the next Major version",
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

	if isV2User(d.Id()) {
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if _, httpResponse, err := datadogClientV2.UsersApi.GetUser(authV2, d.Id()).Execute(); err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return false, nil
			}
			return false, translateClientError(err, "error checking user exists")
		}
	} else {
		client := providerConf.CommunityClient

		if _, err := client.GetUser(d.Id()); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				return false, nil
			}
			return false, translateClientError(err, "error checking user exists")
		}
	}

	return true, nil
}

func buildDatadogUserStruct(d *schema.ResourceData) *datadog.User {
	var u datadog.User
	u.SetDisabled(d.Get("disabled").(bool))
	u.SetEmail(d.Get("email").(string))
	u.SetHandle(d.Get("handle").(string))
	u.SetIsAdmin(d.Get("is_admin").(bool))
	u.SetName(d.Get("name").(string))
	u.SetAccessRole(d.Get("access_role").(string))

	return &u
}

func buildDatadogUserV2Struct(d *schema.ResourceData) *datadogV2.UserCreateRequest {
	userAttributes := datadogV2.NewUserCreateAttributesWithDefaults()
	userAttributes.SetEmail(d.Get("email").(string))
	userAttributes.SetName(d.Get("name").(string))
	userAttributes.SetTitle(d.Get("title").(string))

	userCreate := datadogV2.NewUserCreateDataWithDefaults()
	userCreate.SetAttributes(*userAttributes)
	userRequest := datadogV2.NewUserCreateRequestWithDefaults()

	userRequest.SetData(*userCreate)

	return userRequest
}

func buildDatadogUserV2UpdateStruct(d *schema.ResourceData, userId string) *datadogV2.UserUpdateRequest {
	userAttributes := datadogV2.NewUserUpdateAttributesWithDefaults()
	userAttributes.SetEmail(d.Get("email").(string))
	userAttributes.SetName(d.Get("name").(string))
	userAttributes.SetDisabled(d.Get("disabled").(bool))

	userUpdate := datadogV2.NewUserUpdateDataWithDefaults()
	userUpdate.SetAttributes(*userAttributes)
	userUpdate.SetId(userId)

	userRequest := datadogV2.NewUserUpdateRequestWithDefaults()
	userRequest.SetData(*userUpdate)

	return userRequest
}

func resourceDatadogUserCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	userRequest := buildDatadogUserV2Struct(d)

	// Datadog does not actually delete users, so CreateUser might return a 409.
	// We ignore that case and proceed, likely re-enabling the user.
	createResponse, httpresp, err := datadogClientV2.UsersApi.CreateUser(authV2).Body(*userRequest).Execute()
	if err != nil {
		if httpresp == nil || httpresp.StatusCode != 409 {
			return translateClientError(err, "error creating user")
		}
		email := d.Get("email").(string)
		log.Printf("[INFO] Updating existing Datadog user %s", email)
		// Find user ID by listing user and filtering by email
		ulr, _, err := datadogClientV2.UsersApi.ListUsers(authV2).Filter(email).Execute()
		if err != nil {
			return translateClientError(err, "error searching user")
		}
		if len(ulr.GetData()) != 1 {
			return fmt.Errorf("could not find single user with email %s", email)
		}
		userId := ulr.GetData()[0].GetId()
		userRequest := buildDatadogUserV2UpdateStruct(d, userId)

		res, _, err := datadogClientV2.UsersApi.UpdateUser(authV2, userId).Body(*userRequest).Execute()
		if err != nil {
			return translateClientError(err, "error updating user")
		}
		userData := res.GetData()
		d.SetId(userData.GetId())
	} else {
		userData := createResponse.GetData()
		d.SetId(userData.GetId())
	}

	return resourceDatadogUserRead(d, meta)
}

func resourceDatadogUserRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)

	if isV2User(d.Id()) {
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		userResponse, _, err := datadogClientV2.UsersApi.GetUser(authV2, d.Id()).Execute()
		if err != nil {
			return translateClientError(err, "error getting user")
		}

		userData := userResponse.GetData()
		userAttributes := userData.GetAttributes()
		d.Set("email", userAttributes.GetEmail())
		d.Set("name", userAttributes.GetName())
		d.Set("title", userAttributes.GetTitle())
		d.Set("verified", userAttributes.GetVerified())
	} else {
		client := providerConf.CommunityClient
		u, err := client.GetUser(d.Id())
		if err != nil {
			return err
		}

		d.Set("disabled", u.GetDisabled())
		d.Set("email", u.GetEmail())
		d.Set("handle", u.GetHandle())
		d.Set("name", u.GetName())
		d.Set("verified", u.GetVerified())
		d.Set("access_role", u.GetAccessRole())
		d.Set("is_admin", u.GetIsAdmin())
	}
	return nil
}

func resourceDatadogUserUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)

	if isV2User(d.Id()) {
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		userRequest := buildDatadogUserV2UpdateStruct(d, d.Id())

		_, _, err := datadogClientV2.UsersApi.UpdateUser(authV2, d.Id()).Body(*userRequest).Execute()
		if err != nil {
			return translateClientError(err, "error updating user")
		}
	} else {
		client := providerConf.CommunityClient

		u := buildDatadogUserStruct(d)
		u.SetHandle(d.Id())

		if err := client.UpdateUser(*u); err != nil {
			return translateClientError(err, "error updating user")
		}
	}

	return resourceDatadogUserRead(d, meta)
}

func resourceDatadogUserDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)

	if isV2User(d.Id()) {
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		if httpResponse, err := datadogClientV2.UsersApi.DisableUser(authV2, d.Id()).Execute(); err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return nil
			}
			return translateClientError(err, "error disabling user")
		}
	} else {
		client := providerConf.CommunityClient

		// Datadog does not actually delete users, but instead marks them as disabled.
		// Bypass DeleteUser if GetUser returns User.Disabled == true, otherwise it will 400.
		if u, err := client.GetUser(d.Id()); err == nil && u.GetDisabled() {
			return nil
		}

		if err := client.DeleteUser(d.Id()); err != nil {
			return translateClientError(err, "error deleting user")
		}
	}

	return nil
}

func resourceDatadogUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogUserRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
