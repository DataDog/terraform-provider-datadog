package datadog

import (
	"context"
	"log"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog user resource. This can be used to create and manage Datadog users.",
		CreateContext: resourceDatadogUserCreate,
		ReadContext:   resourceDatadogUserRead,
		UpdateContext: resourceDatadogUserUpdate,
		DeleteContext: resourceDatadogUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disabled": {
				Description: "Whether the user is disabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"email": {
				Description: "Email address for user.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name for user.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"roles": {
				Description: "A list a role IDs to assign to the user.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"send_user_invitation": {
				Description: "Whether an invitation email should be sent when the user is created.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This is only used on create, so don't generate diff when the resource already exists
					return d.Id() != ""
				},
			},
			"verified": {
				Description: "Returns `true` if the user is verified.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"user_invitation_id": {
				Description: "The ID of the user invitation that was sent when creating the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func buildDatadogUserV2Struct(d *schema.ResourceData) *datadogV2.UserCreateRequest {
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

func buildDatadogUserV2UpdateStruct(d *schema.ResourceData, userID string) *datadogV2.UserUpdateRequest {
	userAttributes := datadogV2.NewUserUpdateAttributesWithDefaults()
	userAttributes.SetEmail(d.Get("email").(string))
	if v, ok := d.GetOk("name"); ok {
		userAttributes.SetName(v.(string))
	}
	userAttributes.SetDisabled(d.Get("disabled").(bool))

	userUpdate := datadogV2.NewUserUpdateDataWithDefaults()
	userUpdate.SetAttributes(*userAttributes)
	userUpdate.SetId(userID)

	userRequest := datadogV2.NewUserUpdateRequestWithDefaults()
	userRequest.SetData(*userUpdate)

	return userRequest
}

func resourceDatadogUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	userRequest := buildDatadogUserV2Struct(d)
	var userID string
	updated := false

	// Datadog does not actually delete users, so CreateUser might return a 409.
	// We ignore that case and proceed, likely re-enabling the user.
	createResponse, httpresp, err := datadogClientV2.UsersApi.CreateUser(authV2, *userRequest)
	if err != nil {
		if httpresp == nil || httpresp.StatusCode != 409 {
			return utils.TranslateClientErrorDiag(err, httpresp, "error creating user")
		}
		email := d.Get("email").(string)
		log.Printf("[INFO] Updating existing Datadog user %s", email)
		// Find user ID by listing user and filtering by email
		listResponse, _, err := datadogClientV2.UsersApi.ListUsers(authV2,
			*datadogV2.NewListUsersOptionalParameters().WithFilter(email))
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error searching user")
		}
		if err := utils.CheckForUnparsed(listResponse); err != nil {
			return diag.FromErr(err)
		}
		responseData := listResponse.GetData()
		if len(responseData) != 1 {
			return diag.Errorf("could not find single user with email %s", email)
		}
		userID = responseData[0].GetId()
		userRequest := buildDatadogUserV2UpdateStruct(d, userID)

		updatedUser, _, err := datadogClientV2.UsersApi.UpdateUser(authV2, userID, *userRequest)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error updating user")
		}
		if err := utils.CheckForUnparsed(updatedUser); err != nil {
			return diag.FromErr(err)
		}
		if err := updateUserStateV2(d, &updatedUser); err != nil {
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

	// Send invitation email to newly created users
	if d.Get("send_user_invitation").(bool) {
		if err := sendUserInvitation(userID, d, meta); err != nil {
			return err
		}
	}

	d.SetId(userID)
	if updated {
		return nil
	}
	return updateUserStateV2(d, &createResponse)
}

func sendUserInvitation(userID string, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	userInviteRelationData := datadogV2.NewRelationshipToUserDataWithDefaults()
	userInviteRelationData.SetId(userID)
	userInviteUserRelation := datadogV2.NewRelationshipToUserWithDefaults()
	userInviteUserRelation.SetData(*userInviteRelationData)
	userInviteRelationships := datadogV2.NewUserInvitationRelationshipsWithDefaults()
	userInviteRelationships.SetUser(*userInviteUserRelation)
	userInviteData := datadogV2.NewUserInvitationDataWithDefaults()
	userInviteData.SetRelationships(*userInviteRelationships)
	userInvite := []datadogV2.UserInvitationData{*userInviteData}
	body := *datadogV2.NewUserInvitationsRequestWithDefaults()
	body.SetData(userInvite)

	res, httpResponse, err := datadogClientV2.UsersApi.SendInvitations(authV2, body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error sending user invitation")
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_invitation_id", res.GetData()[0].GetId()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateUserStateV2(d *schema.ResourceData, user *datadogV2.UserResponse) diag.Diagnostics {
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
	if err := d.Set("verified", userAttributes.GetVerified()); err != nil {
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
func resourceDatadogUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	userResponse, httpResponse, err := datadogClientV2.UsersApi.GetUser(authV2, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting user")
	}
	if err := utils.CheckForUnparsed(userResponse); err != nil {
		return diag.FromErr(err)
	}
	return updateUserStateV2(d, &userResponse)
}
func resourceDatadogUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

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
			_, httpResponse, err := datadogClientV2.RolesApi.RemoveUserFromRole(authV2, role, *userRelation)
			if err != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error removing user from role")
			}
		}
		for _, roleI := range rolesToAdd.List() {
			role := roleI.(string)
			roleRelation := datadogV2.NewRelationshipToUserWithDefaults()
			roleRelationData := datadogV2.NewRelationshipToUserDataWithDefaults()
			roleRelationData.SetId(d.Id())
			roleRelation.SetData(*roleRelationData)
			_, httpResponse, err := datadogClientV2.RolesApi.AddUserToRole(authV2, role, *roleRelation)
			if err != nil {
				return utils.TranslateClientErrorDiag(err, httpResponse, "error adding user to role")
			}
		}
	}

	userRequest := buildDatadogUserV2UpdateStruct(d, d.Id())
	updatedUser, httpResponse, err := datadogClientV2.UsersApi.UpdateUser(authV2, d.Id(), *userRequest)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating user")
	}
	if err := utils.CheckForUnparsed(updatedUser); err != nil {
		return diag.FromErr(err)
	}
	// Update state once after we do the UpdateUser operation. At this point, the roles have already been changed
	// so the updated list is available in the update response.
	return updateUserStateV2(d, &updatedUser)
}
func resourceDatadogUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if httpResponse, err := datadogClientV2.UsersApi.DisableUser(authV2, d.Id()); err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error disabling user")
	}

	return nil
}
