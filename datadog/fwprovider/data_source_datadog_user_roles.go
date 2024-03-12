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
	_ datasource.DataSource = &datadogUserRolesDataSource{}
)

type datadogUserRolesDataSourceModel struct {
	// Query Parameters
	RoleID     types.String `tfsdk:"role_id"`
	Filter     types.String `tfsdk:"filter"`
	ExactMatch types.Bool   `tfsdk:"exact_match"`
	// Results
	UserRoles []*UserRoleModel `tfsdk:"user_roles"`
}

func NewDatadogUserRolesDataSource() datasource.DataSource {
	return &datadogUserRolesDataSource{}
}

type datadogUserRolesDataSource struct {
	Api      *datadogV2.RolesApi
	UsersApi *datadogV2.UsersApi
	Auth     context.Context
}

func (r *datadogUserRolesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRolesApiV2()
	r.UsersApi = providerData.DatadogApiInstances.GetUsersApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogUserRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "user_roles"
}

func (d *datadogUserRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Datadog user role assignments.",
		Attributes: map[string]schema.Attribute{
			// Datasource Parameters
			"role_id": schema.StringAttribute{
				Description: "The role's identifier.",
				Required:    true,
			},
			"filter": schema.StringAttribute{
				Description: "Search query, can be user name.",
				Optional:    true,
			},
			"exact_match": schema.BoolAttribute{
				Description: "When true, `filter_keyword` string is exact matched against the user's `name`.",
				Optional:    true,
			},
			// Computed values
			"user_roles": schema.ListAttribute{
				Computed:    true,
				Description: "List of users assigned to role.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"role_id": types.StringType,
						"user_id": types.StringType,
					},
				},
			},
		},
	}

}

func (d *datadogUserRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogUserRolesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var optionalParams datadogV2.ListRoleUsersOptionalParameters
	roleID := state.RoleID.ValueString()

	if !state.Filter.IsNull() {
		optionalParams.Filter = state.Filter.ValueStringPointer()
	}

	pageSize := int64(100)
	pageNumber := int64(0)

	var roleUsers []datadogV2.User
	for {
		optionalParams.PageNumber = &pageNumber
		optionalParams.PageSize = &pageSize

		ddResp, _, err := d.Api.ListRoleUsers(d.Auth, roleID, optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting user roles"))
			return
		}

		roleUsers = append(roleUsers, ddResp.GetData()...)
		if len(ddResp.GetData()) < 100 {
			break
		}
		pageNumber++
	}

	d.updateState(&state, &roleUsers)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogUserRolesDataSource) updateState(state *datadogUserRolesDataSourceModel, roleUsers *[]datadogV2.User) {
	exactMatch := state.ExactMatch.ValueBool()
	filterKeyword := state.Filter.ValueString()
	var userRoles []*UserRoleModel

	for _, user := range *roleUsers {
		if exactMatch {
			if u, _, err := r.UsersApi.GetUser(r.Auth, user.GetId()); err == nil {
				attributes := u.Data.GetAttributes()
				if attributes.GetName() == filterKeyword {
					userRole := UserRoleModel{
						RoleId: types.StringValue(string(state.RoleID.ValueString())),
						UserId: types.StringValue(user.GetId()),
					}

					userRoles = append(userRoles, &userRole)
				}
			}
		} else {
			userRole := UserRoleModel{
				RoleId: types.StringValue(state.RoleID.ValueString()),
				UserId: types.StringValue(user.GetId()),
			}

			userRoles = append(userRoles, &userRole)
		}
	}

	state.UserRoles = userRoles
}
