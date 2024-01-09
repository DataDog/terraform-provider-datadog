package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogUsersDataSource{}
)

type UserModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Email types.String `tfsdk:"email"`
}

type datadogUsersDataSourceModel struct {
	// Query Parameters
	Filter       types.String `tfsdk:"filter"`
	FilterStatus types.String `tfsdk:"filter_status"`

	// Results
	ID    types.String `tfsdk:"id"`
	Users []*UserModel `tfsdk:"users"`
}

type datadogUsersDataSource struct {
	Api  *datadogV2.UsersApi
	Auth context.Context
}

func NewDatadogUsersDataSource() datasource.DataSource {
	return &datadogUsersDataSource{}
}

func (d *datadogUsersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetUsersApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "users"
}

func (d *datadogUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing users for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"filter": schema.StringAttribute{
				Optional:    true,
				Description: "Filter all users by the given string.",
			},
			"filter_status": schema.StringAttribute{
				Optional:    true,
				Description: "Filter on status attribute. Comma-separated list with possible values of Active, Pending, and Disabled.",
			},

			// computed values
			"users": schema.ListAttribute{
				Computed:    true,
				Description: "List of users",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":    types.StringType,
						"name":  types.StringType,
						"email": types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogUsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var optionalParams datadogV2.ListUsersOptionalParameters
	if !state.Filter.IsNull() {
		optionalParams.Filter = state.Filter.ValueStringPointer()
	}
	if !state.FilterStatus.IsNull() {
		optionalParams.FilterStatus = state.FilterStatus.ValueStringPointer()
	}

	pageSize := 100
	pageNumber := int64(0)
	optionalParams.WithPageSize(int64(pageSize))

	var users []datadogV2.User
	for {
		optionalParams.WithPageNumber(pageNumber)

		ddResp, _, err := d.Api.ListUsers(d.Auth, optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting users"))
			return
		}

		users = append(users, ddResp.GetData()...)
		if len(ddResp.GetData()) < pageSize {
			break
		}
		pageNumber++
	}

	d.updateState(&state, &users)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogUsersDataSource) updateState(state *datadogUsersDataSourceModel, usersData *[]datadogV2.User) {
	var users []*UserModel
	for _, user := range *usersData {
		u := UserModel{
			ID:    types.StringValue(user.GetId()),
			Email: types.StringValue(user.Attributes.GetEmail()),
			Name:  types.StringValue(user.Attributes.GetName()),
		}

		users = append(users, &u)
	}

	hashingData := fmt.Sprintf("%s:%s", state.Filter.ValueString(), state.FilterStatus.ValueString())

	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Users = users
}
