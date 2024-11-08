package fwprovider

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &userRoleResource{}
	_ resource.ResourceWithImportState = &userRoleResource{}
)

type userRoleResource struct {
	Api   *datadogV2.RolesApi
	Auth  context.Context
	Users *datadogV2.UsersApi
}

type UserRoleModel struct {
	ID     types.String `tfsdk:"id"`
	RoleId types.String `tfsdk:"role_id"`
	UserId types.String `tfsdk:"user_id"`
}

func NewUserRoleResource() resource.Resource {
	return &userRoleResource{}
}

func (r *userRoleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRolesApiV2()
	r.Auth = providerData.Auth
	r.Users = providerData.DatadogApiInstances.GetUsersApiV2()
}

func (r *userRoleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "user_role"
}

func (r *userRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog UserRole resource. This can be used to create and manage Datadog User Roles. Conflicts may occur if used together with the `datadog_user` resource's `roles` attribute or the `datadog_service_account` resource's `roles` attribute. This resource is in beta and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"role_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the role that the user is assigned to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userRoleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving role_id or user_id from given ID", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("role_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_id"), result[1])...)
}

func (r *userRoleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state UserRoleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// get User
	userId := state.UserId.ValueString()
	resp, httpResp, err := r.Users.GetUser(r.Auth, userId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RoleUsers"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// check if User already has Role
	for _, role := range resp.GetData().Relationships.GetRoles().Data {
		if roleId := role.GetId(); roleId == state.RoleId.ValueString() {
			state.ID = types.StringValue(roleId + ":" + userId)
			return
		}
	}

	// User doesn't have Role
	response.State.RemoveResource(ctx)
}

func (r *userRoleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state UserRoleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildUserRoleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	roleId := state.RoleId.ValueString()
	resp, _, err := r.Api.AddUserToRole(r.Auth, roleId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving UserRole"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Save data into Terraform state
	state.ID = types.StringValue(state.RoleId.ValueString() + ":" + state.UserId.ValueString())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *userRoleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported for this resource", "UserRoles assignments should be updated by deleting the old assignment and creating a new one.")
}

func (r *userRoleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state UserRoleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildUserRoleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	roleId := state.RoleId.ValueString()
	_, httpResp, err := r.Api.RemoveUserFromRole(r.Auth, roleId, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting user_role"))
		return
	}

	// Save data into Terraform state
	state.ID = types.StringValue(state.RoleId.ValueString() + ":" + state.UserId.ValueString())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *userRoleResource) buildUserRoleRequestBody(_ context.Context, state *UserRoleModel) (*datadogV2.RelationshipToUser, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	relationship := &datadogV2.RelationshipToUser{
		Data: *datadogV2.NewRelationshipToUserDataWithDefaults(),
	}
	relationship.Data.Id = state.UserId.ValueString()

	return relationship, diags
}
