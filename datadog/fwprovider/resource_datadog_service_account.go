package fwprovider

import (
	"context"
	"fmt"
	"log"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &serviceAccountResource{}
	_ resource.ResourceWithImportState = &serviceAccountResource{}
)

type serviceAccountResource struct {
	Api                 *datadogV2.UsersApi
	ServiceAccountApiV2 *datadogV2.ServiceAccountsApi
	RolesApiV2          *datadogV2.RolesApi
	Auth                context.Context
}
type serviceAccountResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Email    types.String `tfsdk:"email"`
	Name     types.String `tfsdk:"name"`
	Roles    types.Set    `tfsdk:"roles"`
}

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

func (*serviceAccountResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "service_account"
}

func (r *serviceAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetUsersApiV2()
	r.ServiceAccountApiV2 = providerData.DatadogApiInstances.GetServiceAccountsApiV2()
	r.RolesApiV2 = providerData.DatadogApiInstances.GetRolesApiV2()
	r.Auth = providerData.Auth
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog service account resource. This can be used to create and manage Datadog service accounts.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name for the service account.",
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the service account is disabled.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"email": schema.StringAttribute{
				Description: "Email of the associated user.",
				Required:    true,
			},
			"roles": schema.SetAttribute{
				Description: "A list of role IDs to assign to the service account.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *serviceAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *serviceAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state serviceAccountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	userResponse, httpResp, err := r.Api.GetUser(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			state.ID = types.String{}
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting user"))
		return
	}
	if diags := updateServiceAccountStateV2(ctx, &state, &userResponse); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state serviceAccountResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	serviceAccountRequest, diags := buildDatadogServiceAccountV2Request(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	var userID string

	createResponse, httpresp, err := r.ServiceAccountApiV2.CreateServiceAccount(r.Auth, *serviceAccountRequest)
	if err != nil {
		// Datadog does not actually delete users, so CreateUser might return a 409.
		// We ignore that case and proceed, likely re-enabling the user.
		if httpresp == nil || httpresp.StatusCode != 409 {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error creating service account"))
			return
		}
		email := state.Email.ValueString()
		log.Printf("[INFO] Linking existing Datadog email %s to service account", email)

		var existingServiceAccount *datadogV2.User
		// Find user ID by listing user and filtering by email
		listResponse, _, err := r.Api.ListUsers(r.Auth,
			*datadogV2.NewListUsersOptionalParameters().WithFilter(email))

		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error searching for service account"))
			return
		}

		if err := utils.CheckForUnparsed(listResponse); err != nil {
			response.Diagnostics.AddError("", err.Error())
		}

		responseData := listResponse.GetData()

		if len(responseData) > 1 {
			for _, user := range responseData {
				if user.Attributes.GetEmail() == email && user.Attributes.GetServiceAccount() {
					existingServiceAccount = &user
					break
				}
			}

			if existingServiceAccount == nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("could not find service account with email %s", email)))
				return
			}
		} else {
			existingServiceAccount = &responseData[0]
		}

		userID = existingServiceAccount.GetId()
		userRequest := buildDatadogUserV2UpdateStructFw(state, userID)

		updatedUser, _, err := r.Api.UpdateUser(r.Auth, userID, *userRequest)

		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error updating service account"))
			return
		}
		if err := utils.CheckForUnparsed(updatedUser); err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error updating service account"))
			return
		}

		// Update roles
		newRoles := state.Roles
		oldRoles, _ := types.SetValueFrom(ctx, types.StringType, &newRoles)
		oldRolesSlice := []string{}
		diags.Append(oldRoles.ElementsAs(ctx, &oldRolesSlice, false)...)

		for _, existingRole := range updatedUser.Data.Relationships.Roles.GetData() {
			oldRolesWithExisting := append(oldRolesSlice, existingRole.GetId())
			oldRoles, _ = types.SetValueFrom(ctx, types.StringType, &oldRolesWithExisting)
		}

		if err := r.updateRolesFw(ctx, userID, oldRoles, newRoles); err != nil {
			response.Diagnostics.Append(err)
			return
		}

		if diags := updateServiceAccountStateV2(ctx, &state, &updatedUser); diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	} else {
		if err := utils.CheckForUnparsed(createResponse); err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
			return
		}
		userData := createResponse.GetData()
		userID = userData.GetId()
	}

	state.ID = types.StringValue(userID)
	if diags := updateServiceAccountStateV2(ctx, &state, &createResponse); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, prev_state serviceAccountResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	response.Diagnostics.Append(request.State.Get(ctx, &prev_state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if !prev_state.Roles.Equal(state.Roles) {
		newRoles := state.Roles
		oldRoles := prev_state.Roles

		if err := r.updateRolesFw(ctx, state.ID.ValueString(), oldRoles, newRoles); err != nil {
			response.Diagnostics.Append(err)
			return
		}
	}

	userRequest := buildDatadogUserV2UpdateStructFw(state, state.ID.ValueString())
	updatedUser, _, err := r.Api.UpdateUser(r.Auth, state.ID.ValueString(), *userRequest)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating service account"))
		return
	}
	if err := utils.CheckForUnparsed(updatedUser); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
		return
	}
	if diags := updateServiceAccountStateV2(ctx, &state, &updatedUser); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *serviceAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state serviceAccountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if httpResponse, err := r.Api.DisableUser(r.Auth, state.ID.ValueString()); err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error disabling user"))
		return
	}
}

func updateServiceAccountStateV2(ctx context.Context, state *serviceAccountResourceModel, user *datadogV2.UserResponse) diag.Diagnostics {
	userData := user.GetData()
	userAttributes := userData.GetAttributes()

	state.Email = types.StringValue(userAttributes.GetEmail())
	state.Name = types.StringValue(userAttributes.GetName())
	state.Disabled = types.BoolValue(userAttributes.GetDisabled())
	diags := diag.Diagnostics{}
	if !state.Roles.IsNull() {
		userRelations := userData.GetRelationships()
		userRolesRelations := userRelations.GetRoles()
		userRoles := userRolesRelations.GetData()

		roles := make([]string, len(userRoles))
		for i, userRole := range userRoles {
			roles[i] = userRole.GetId()
		}
		state.Roles, diags = types.SetValueFrom(ctx, types.StringType, roles)
	}
	return diags
}

func buildDatadogServiceAccountV2Request(ctx context.Context, state *serviceAccountResourceModel) (*datadogV2.ServiceAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	serviceAccountAttributes := datadogV2.NewServiceAccountCreateAttributesWithDefaults()
	serviceAccountAttributes.SetServiceAccount(true)
	serviceAccountAttributes.SetEmail(state.Email.ValueString())
	if !state.Name.IsNull() {
		serviceAccountAttributes.SetName(state.Name.ValueString())
	}
	serviceAccountCreate := datadogV2.NewServiceAccountCreateDataWithDefaults()
	serviceAccountCreate.SetAttributes(*serviceAccountAttributes)

	var roles []string
	if !state.Roles.IsUnknown() {
		diags.Append(state.Roles.ElementsAs(ctx, &roles, false)...)
	}
	rolesData := make([]datadogV2.RelationshipToRoleData, len(roles))
	for i, role := range roles {
		roleData := datadogV2.NewRelationshipToRoleData()
		roleData.SetId(role)
		rolesData[i] = *roleData
	}
	toRoles := datadogV2.NewRelationshipToRoles()
	toRoles.SetData(rolesData)

	userRelationships := datadogV2.NewUserRelationships()
	userRelationships.SetRoles(*toRoles)

	serviceAccountCreate.SetRelationships(*userRelationships)

	serviceAccountRequest := datadogV2.NewServiceAccountCreateRequestWithDefaults()
	serviceAccountRequest.SetData(*serviceAccountCreate)

	return serviceAccountRequest, diags
}

func buildDatadogUserV2UpdateStructFw(state serviceAccountResourceModel, userID string) *datadogV2.UserUpdateRequest {
	userAttributes := datadogV2.NewUserUpdateAttributesWithDefaults()
	userAttributes.SetEmail(state.Email.ValueString())
	if !state.Name.IsNull() {
		userAttributes.SetName(state.Name.ValueString())
	}
	userAttributes.SetDisabled(state.Disabled.ValueBool())
	userUpdate := datadogV2.NewUserUpdateDataWithDefaults()
	userUpdate.SetAttributes(*userAttributes)
	userUpdate.SetId(userID)

	userRequest := datadogV2.NewUserUpdateRequestWithDefaults()
	userRequest.SetData(*userUpdate)

	return userRequest
}

func (r *serviceAccountResource) updateRolesFw(ctx context.Context, userID string, oldRoles types.Set, newRoles types.Set) diag.Diagnostic {

	oldRolesSlice := []string{}
	newRolesSlice := []string{}

	oldRoles.ElementsAs(ctx, &oldRolesSlice, false)
	newRoles.ElementsAs(ctx, &newRolesSlice, false)

	rolesToRemove := utils.StringSliceDifference(oldRolesSlice, newRolesSlice)
	rolesToAdd := utils.StringSliceDifference(newRolesSlice, oldRolesSlice)

	for _, role := range rolesToAdd {
		roleRelation := datadogV2.NewRelationshipToUserWithDefaults()
		roleRelationData := datadogV2.NewRelationshipToUserDataWithDefaults()
		roleRelationData.SetId(userID)
		roleRelation.SetData(*roleRelationData)
		_, _, err := r.RolesApiV2.AddUserToRole(r.Auth, role, *roleRelation)
		if err != nil {
			return diag.NewErrorDiagnostic("error adding user to role: ", err.Error())
		}
	}
	for _, role := range rolesToRemove {
		userRelation := datadogV2.NewRelationshipToUserWithDefaults()
		userRelationData := datadogV2.NewRelationshipToUserDataWithDefaults()
		userRelationData.SetId(userID)
		userRelation.SetData(*userRelationData)

		_, _, err := r.RolesApiV2.RemoveUserFromRole(r.Auth, role, *userRelation)
		if err != nil {
			return diag.NewErrorDiagnostic("error removing user from role: ", err.Error())
		}
	}

	return nil
}
