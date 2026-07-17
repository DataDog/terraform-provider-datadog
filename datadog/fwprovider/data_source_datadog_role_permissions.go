package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogRolePermissionsDataSource{}
)

type RolePermissionModel struct {
	PermissionId types.String `tfsdk:"permission_id"`
	Name         types.String `tfsdk:"name"`
}

type datadogRolePermissionsDataSourceModel struct {
	// Query Parameters
	RoleID types.String `tfsdk:"role_id"`
	// Results
	ID          types.String           `tfsdk:"id"`
	Permissions []*RolePermissionModel `tfsdk:"permissions"`
}

func NewDatadogRolePermissionsDataSource() datasource.DataSource {
	return &datadogRolePermissionsDataSource{}
}

type datadogRolePermissionsDataSource struct {
	Api  *datadogV2.RolesApi
	Auth context.Context
}

func (r *datadogRolePermissionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRolesApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogRolePermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "role_permissions"
}

func (d *datadogRolePermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the list of permissions assigned to a Datadog role.",
		Attributes: map[string]schema.Attribute{
			// Datasource Parameters
			"id": utils.ResourceIDAttribute(),
			"role_id": schema.StringAttribute{
				Description: "The role's identifier.",
				Required:    true,
			},
			// Computed values
			"permissions": schema.ListAttribute{
				Computed:    true,
				Description: "List of permissions assigned to the role.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"permission_id": types.StringType,
						"name":          types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogRolePermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogRolePermissionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := state.RoleID.ValueString()

	ddResp, _, err := d.Api.ListRolePermissions(d.Auth, roleID)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing role permissions"))
		return
	}

	var permissions []*RolePermissionModel
	for _, perm := range ddResp.GetData() {
		permissions = append(permissions, &RolePermissionModel{
			PermissionId: types.StringValue(perm.GetId()),
			Name:         types.StringValue(perm.Attributes.GetName()),
		})
	}

	state.ID = types.StringValue(roleID)
	state.Permissions = permissions

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
