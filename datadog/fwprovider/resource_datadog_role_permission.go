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
	_ resource.ResourceWithConfigure   = &rolePermissionResource{}
	_ resource.ResourceWithImportState = &rolePermissionResource{}
)

type rolePermissionResource struct {
	Api  *datadogV2.RolesApi
	Auth context.Context
}

type RolePermissionModel struct {
	ID           types.String `tfsdk:"id"`
	RoleId       types.String `tfsdk:"role_id"`
	PermissionId types.String `tfsdk:"permission_id"`
}

func NewRolePermissionResource() resource.Resource {
	return &rolePermissionResource{}
}

func (r *rolePermissionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRolesApiV2()
	r.Auth = providerData.Auth
}

func (r *rolePermissionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "role_permission"
}

func (r *rolePermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RolePermission resource. This can be used to create and manage Datadog Role Permissions. Conflicts may occur if used together with the `datadog_role` resource's `permission` attribute. This resource is in beta and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"role_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the role that the permission is assigned to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"permission_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the permission.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *rolePermissionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving role_id or permission_id from given ID", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("role_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("permission_id"), result[1])...)
}

func (r *rolePermissionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state RolePermissionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	roleId := state.RoleId.ValueString()

	var rolePermissions []datadogV2.Permission
	resp, httpResp, err := r.Api.ListRolePermissions(r.Auth, roleId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Role no longer exists, remove the mapping
			response.State.RemoveResource(ctx)
			return
		}

		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RolePermissions"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	rolePermissions = append(rolePermissions, resp.GetData()...)
	updated := r.updatedStateFromRoleResponse(ctx, &state, rolePermissions, false)

	// Delete state if updated is false, since that means the permission doesn't exist
	if !updated {
		response.State.RemoveResource(ctx)
	}
}

func (r *rolePermissionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state RolePermissionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildRolePermissionRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	roleId := state.RoleId.ValueString()
	resp, _, err := r.Api.AddPermissionToRole(r.Auth, roleId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RolePermission"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Save data into Terraform state
	r.updatedStateFromRoleResponse(ctx, &state, resp.GetData(), true)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rolePermissionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported for this resource", "RolePermissions assignments should be updated by deleting the old assignment and creating a new one.")
}

func (r *rolePermissionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state RolePermissionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildRolePermissionRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	roleId := state.RoleId.ValueString()
	resp, httpResp, err := r.Api.RemovePermissionFromRole(r.Auth, roleId, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting role_permission"))
		return
	}

	// Save data into Terraform state
	r.updatedStateFromRoleResponse(ctx, &state, resp.GetData(), true)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rolePermissionResource) buildRolePermissionRequestBody(ctx context.Context, state *RolePermissionModel) (*datadogV2.RelationshipToPermission, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	relationship := &datadogV2.RelationshipToPermission{
		Data: datadogV2.NewRelationshipToPermissionDataWithDefaults(),
	}
	relationship.Data.Id = state.PermissionId.ValueStringPointer()

	return relationship, diags
}

func (r *rolePermissionResource) updatedStateFromRoleResponse(ctx context.Context, state *RolePermissionModel, resp []datadogV2.Permission, force bool) bool {
	// if force, just set the state to the user and role ID that's already in the state
	// this is useful for create/delete since the API doesn't return all the users
	if force {
		state.ID = types.StringValue(state.RoleId.ValueString() + ":" + state.PermissionId.ValueString())
		return true
	}

	for _, permission := range resp {
		if permission.GetId() == state.PermissionId.ValueString() {
			userId := permission.GetId()
			state.ID = types.StringValue(state.RoleId.ValueString() + ":" + userId)
			state.PermissionId = types.StringValue(userId)
			return true
		}
	}
	return false
}
