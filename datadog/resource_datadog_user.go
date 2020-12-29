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
		Importer: &schema.ResourceImporter{
			State: resourceDatadogUserImport,
		},

		Schema: map[string]*schema.Schema{
			"disabled": {
				Description: "Whether the user is disabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"email": {
				Description: "Email address for user",
				Type:        schema.TypeString,
				Required:    true,
			},
			"handle": {
				Description: "The user handle, must be a valid email.",
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "This parameter is deprecated and will be removed from the next Major version",
			},
			"is_admin": {
				Description: "Whether the user is an administrator. Warning: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.",
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Deprecated:  "This parameter is replaced by `roles` and will be removed from the next Major version",
			},
			"access_role": {
				Description: "Role description for user. Can be st (standard user), adm (admin user) or ro (read-only user). Default is st. access_role is ignored for new users created with this resource. New users have to use the roles attribute.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "st",
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					return (d.Get("roles").(*schema.Set)).Len() > 0
				},
			},
			"name": {
				Description: "Name for user",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"role": {
				Description: "Role description for user. Warning: the corresponding query parameter is ignored by the Datadog API, thus the argument would always trigger an execution plan.",
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "This parameter was removed from the API and has no effect",
			},
			"roles": {
				Description: "A list a role IDs to assign to the user.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"verified": {
				Description: "Returns true if Datadog user is verified",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
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

func buildDatadogUserV2Struct(d *schema.ResourceData, providerConf *ProviderConfiguration) *datadogV2.UserCreateRequest {
	userAttributes := datadogV2.NewUserCreateAttributesWithDefaults()
	userAttributes.SetEmail(d.Get("email").(string))
	if v, ok := d.GetOk("name"); ok {
		userAttributes.SetName(v.(string))
	}

	userCreate := datadogV2.NewUserCreateDataWithDefaults()
	userCreate.SetAttributes(*userAttributes)

	roles := d.Get("roles").(*schema.Set).List()
	rolesData := make([]datadogV2.RelationshipToRoleData, len(roles))
	for i, role := range roles {
		roleData := datadogV2.NewRelationshipToRoleData()
		roleData.SetId(role.(string))
		rolesData[i] = *roleData
	}

	toRoles := datadogV2.NewRelationshipToRoles()
	toRoles.SetData(rolesData)

	userRelationships := datadogV2.NewUserRelationships()
	userRelationships.SetRoles(*toRoles)
	userCreate.SetRelationships(*userRelationships)

	userRequest := datadogV2.NewUserCreateRequestWithDefaults()
	userRequest.SetData(*userCreate)

	return userRequest
}

func buildDatadogUserV2UpdateStruct(d *schema.ResourceData, userId string) *datadogV2.UserUpdateRequest {
	userAttributes := datadogV2.NewUserUpdateAttributesWithDefaults()
	userAttributes.SetEmail(d.Get("email").(string))
	if v, ok := d.GetOk("name"); ok {
		userAttributes.SetName(v.(string))
	}
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

	userRequest := buildDatadogUserV2Struct(d, providerConf)

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
		listResponse, _, err := datadogClientV2.UsersApi.ListUsers(authV2).Filter(email).Execute()
		if err != nil {
			return translateClientError(err, "error searching user")
		}
		responseData := listResponse.GetData()
		if len(responseData) != 1 {
			return fmt.Errorf("could not find single user with email %s", email)
		}
		userId := responseData[0].GetId()
		userRequest := buildDatadogUserV2UpdateStruct(d, userId)

		if _, _, err = datadogClientV2.UsersApi.UpdateUser(authV2, userId).Body(*userRequest).Execute(); err != nil {
			return translateClientError(err, "error updating user")
		}
		d.SetId(userId)
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

		userResponse, httpResponse, err := datadogClientV2.UsersApi.GetUser(authV2, d.Id()).Execute()
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				d.SetId("")
				return nil
			}
			return translateClientError(err, "error getting user")
		}

		userData := userResponse.GetData()
		userAttributes := userData.GetAttributes()

		userRelations := userData.GetRelationships()
		userRolesRelations := userRelations.GetRoles()
		userRoles := userRolesRelations.GetData()

		if err := d.Set("email", userAttributes.GetEmail()); err != nil {
			return err
		}
		if err := d.Set("name", userAttributes.GetName()); err != nil {
			return err
		}
		if err := d.Set("verified", userAttributes.GetVerified()); err != nil {
			return err
		}
		if err := d.Set("disabled", userAttributes.GetDisabled()); err != nil {
			return err
		}
		roles := make([]string, len(userRoles))
		for i, userRole := range userRoles {
			roles[i] = userRole.GetId()
		}
		if err := d.Set("roles", roles); err != nil {
			return err
		}
	} else {
		client := providerConf.CommunityClient
		u, err := client.GetUser(d.Id())
		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				d.SetId("")
				return nil
			}
			return err
		}

		if err := d.Set("disabled", u.GetDisabled()); err != nil {
			return err
		}
		if err := d.Set("email", u.GetEmail()); err != nil {
			return err
		}
		if err := d.Set("handle", u.GetHandle()); err != nil {
			return err
		}
		if err := d.Set("name", u.GetName()); err != nil {
			return err
		}
		if err := d.Set("verified", u.GetVerified()); err != nil {
			return err
		}
		if err := d.Set("access_role", u.GetAccessRole()); err != nil {
			return err
		}
		if err := d.Set("is_admin", u.GetIsAdmin()); err != nil {
			return err
		}
	}
	return nil
}

func resourceDatadogUserUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)

	if !isV2User(d.Id()) && (d.Get("roles").(*schema.Set)).Len() > 0 {
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2
		email := d.Get("email").(string)
		log.Printf("[INFO] Migrating existing Datadog user %s", email)
		// Find user ID by listing user and filtering by email
		listResponse, _, err := datadogClientV2.UsersApi.ListUsers(authV2).Filter(email).Execute()
		if err != nil {
			return translateClientError(err, "error searching user")
		}
		responseData := listResponse.GetData()
		if len(responseData) != 1 {
			return fmt.Errorf("could not find single user with email %s", email)
		}
		userId := responseData[0].GetId()
		d.SetId(userId)
	}

	if isV2User(d.Id()) {
		datadogClientV2 := providerConf.DatadogClientV2
		authV2 := providerConf.AuthV2

		userRequest := buildDatadogUserV2UpdateStruct(d, d.Id())

		if _, _, err := datadogClientV2.UsersApi.UpdateUser(authV2, d.Id()).Body(*userRequest).Execute(); err != nil {
			return translateClientError(err, "error updating user")
		}
		if d.HasChange("roles") {
			oldRolesI, newRolesI := d.GetChange("roles")
			oldRoles := oldRolesI.(*schema.Set)
			newRoles := newRolesI.(*schema.Set)
			rolesToRemove := oldRoles.Difference(newRoles)
			rolesToAdd := newRoles.Difference(oldRoles)
			for _, roleI := range rolesToRemove.List() {
				role := roleI.(string)
				userRelation := datadogV2.NewRelationshipToUserWithDefaults()
				userRelationData := datadogV2.NewRelationshipToUserDataWithDefaults()
				userRelationData.SetId(d.Id())
				userRelation.SetData(*userRelationData)
				_, _, err := datadogClientV2.RolesApi.RemoveUserFromRole(authV2, role).Body(*userRelation).Execute()
				if err != nil {
					return translateClientError(err, "error removing user from role")
				}
			}
			for _, roleI := range rolesToAdd.List() {
				role := roleI.(string)
				roleRelation := datadogV2.NewRelationshipToUserWithDefaults()
				roleRelationData := datadogV2.NewRelationshipToUserDataWithDefaults()
				roleRelationData.SetId(d.Id())
				roleRelation.SetData(*roleRelationData)
				_, _, err := datadogClientV2.RolesApi.AddUserToRole(authV2, role).Body(*roleRelation).Execute()
				if err != nil {
					return translateClientError(err, "error adding user to role")
				}
			}
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
