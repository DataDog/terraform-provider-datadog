package datadog

import (
	"context"
	"log"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service account resource. This can be used to create and manage Datadog service accounts.",
		CreateContext: resourceDatadogServiceAccountCreate,
		ReadContext:   resourceDatadogServiceAccountRead,
		UpdateContext: resourceDatadogServiceAccountUpdate,
		DeleteContext: resourceDatadogUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"disabled": {
					Description: "Whether the service account is disabled.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
				"email": {
					Description: "Email of the associated user.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"name": {
					Description: "Name for the service account.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"roles": {
					Description: "A list a role IDs to assign to the service account.",
					Type:        schema.TypeSet,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

// Builds a service account creation payload
func buildDatadogServiceAccountV2Request(d *schema.ResourceData) *datadogV2.ServiceAccountCreateRequest {
	serviceAccountAttributes := datadogV2.NewServiceAccountCreateAttributesWithDefaults()
	serviceAccountAttributes.SetServiceAccount(true)
	serviceAccountAttributes.SetEmail(d.Get("email").(string))
	if v, ok := d.GetOk("name"); ok {
		serviceAccountAttributes.SetName(v.(string))
	}

	serviceAccountCreate := datadogV2.NewServiceAccountCreateDataWithDefaults()
	serviceAccountCreate.SetAttributes(*serviceAccountAttributes)

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

	serviceAccountCreate.SetRelationships(*userRelationships)

	serviceAccountRequest := datadogV2.NewServiceAccountCreateRequestWithDefaults()
	serviceAccountRequest.SetData(*serviceAccountCreate)

	return serviceAccountRequest
}

// Creates a service account, which is a special subclass of the User model
func resourceDatadogServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	serviceAccountRequest := buildDatadogServiceAccountV2Request(d)
	var userID string
	updated := false

	createResponse, httpresp, err := apiInstances.GetServiceAccountsApiV2().CreateServiceAccount(auth, *serviceAccountRequest)
	if err != nil {
		// Datadog does not actually delete users, so CreateUser might return a 409.
		// We ignore that case and proceed, likely re-enabling the user.
		if httpresp == nil || httpresp.StatusCode != 409 {
			return utils.TranslateClientErrorDiag(err, httpresp, "Error creating service account")
		}
		email := d.Get("email").(string)
		log.Printf("[INFO] Linking existing Datadog email %s to service account", email)

		var existingServiceAccount *datadogV2.User
		// Find user ID by listing user and filtering by email
		listResponse, _, err := apiInstances.GetUsersApiV2().ListUsers(auth,
			*datadogV2.NewListUsersOptionalParameters().WithFilter(email))

		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "Error searching for service account")
		}

		if err := utils.CheckForUnparsed(listResponse); err != nil {
			return diag.FromErr(err)
		}

		responseData := listResponse.GetData()

		if len(responseData) > 1 {
			for _, user := range responseData {
				// TODO: Why is this greedy?
				// TODO: Isn't this filtered by the parameters?
				if user.Attributes.GetEmail() == email && user.Attributes.GetServiceAccount() == d.Get("service_account").(bool) {
					existingServiceAccount = &user
					break
				}
			}

			if existingServiceAccount == nil {
				// TODO: This shouldn't be possible, literally just created said user
				return diag.Errorf("could not find service account with email %s", email)
			}
		} else {
			existingServiceAccount = &responseData[0]
		}

		userID = existingServiceAccount.GetId()
		userRequest := buildDatadogUserV2UpdateStruct(d, userID)

		updatedUser, _, err := apiInstances.GetUsersApiV2().UpdateUser(auth, userID, *userRequest)

		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "Error updating service account")
		}

		if err := utils.CheckForUnparsed(updatedUser); err != nil {
			return diag.FromErr(err)
		}

		// Update roles
		_, newRolesI := d.GetChange("roles")
		newRoles := newRolesI.(*schema.Set)
		oldRoles := schema.NewSet(newRoles.F, []interface{}{})

		for _, existingRole := range updatedUser.Data.Relationships.Roles.GetData() {
			oldRoles.Add(existingRole.GetId())
		}

		if err := updateRoles(meta, userID, oldRoles, newRoles); err != nil {
			return err
		}

		if err := updateServiceAccountStateV2(d, &updatedUser); err != nil {
			return err
		}
		updated = true
	} else {
		if err := utils.CheckForUnparsed(createResponse); err != nil {
			return diag.FromErr(err)
		}
		userData := createResponse.GetData()
		userID = userData.GetId()
	}

	d.SetId(userID)
	if updated {
		return nil
	}

	return updateServiceAccountStateV2(d, &createResponse)
}

func updateServiceAccountStateV2(d *schema.ResourceData, user *datadogV2.UserResponse) diag.Diagnostics {
	userData := user.GetData()
	userAttributes := userData.GetAttributes()
	userRelations := userData.GetRelationships()
	userRolesRelations := userRelations.GetRoles()
	userRoles := userRolesRelations.GetData()

	if err := d.Set("email", userAttributes.GetEmail()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", userAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("disabled", userAttributes.GetDisabled()); err != nil {
		return diag.FromErr(err)
	}

	roles := make([]string, len(userRoles))
	for i, userRole := range userRoles {
		roles[i] = userRole.GetId()
	}

	if err := d.Set("roles", roles); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	userResponse, httpResponse, err := apiInstances.GetUsersApiV2().GetUser(auth, d.Id())

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			// TODO: Why?
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting user")
	}

	if err := utils.CheckForUnparsed(userResponse); err != nil {
		return diag.FromErr(err)
	}

	return updateServiceAccountStateV2(d, &userResponse)
}

func resourceDatadogServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if d.HasChange("roles") {
		oldRolesI, newRolesI := d.GetChange("roles")
		oldRoles := oldRolesI.(*schema.Set)
		newRoles := newRolesI.(*schema.Set)

		if err := updateRoles(meta, d.Id(), oldRoles, newRoles); err != nil {
			return err
		}
	}

	userRequest := buildDatadogUserV2UpdateStruct(d, d.Id())
	updatedUser, httpResponse, err := apiInstances.GetUsersApiV2().UpdateUser(auth, d.Id(), *userRequest)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating service account")
	}
	if err := utils.CheckForUnparsed(updatedUser); err != nil {
		return diag.FromErr(err)
	}
	// Update state once after we do the UpdateUser operation. At this point, the roles have already been changed
	// so the updated list is available in the update response.
	return updateServiceAccountStateV2(d, &updatedUser)
}

func resourceDatadogServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetUsersApiV2().DisableUser(auth, d.Id()); err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error disabling user")
	}

	return nil
}
