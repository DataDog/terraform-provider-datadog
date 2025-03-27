package fwprovider

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/DataDog/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogUsersDataSource{}
)

type UserModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Email          types.String `tfsdk:"email"`
	ServiceAccount types.Bool   `tfsdk:"service_account"`
	MfaEnabled     types.Bool   `tfsdk:"mfa_enabled"`
	Status         types.String `tfsdk:"status"`
	CreatedAt      types.String `tfsdk:"created_at"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	Title          types.String `tfsdk:"title"`
	Handle         types.String `tfsdk:"handle"`
	Disabled       types.Bool   `tfsdk:"disabled"`
	Verified       types.Bool   `tfsdk:"verified"`
	Icon           types.String `tfsdk:"icon"`
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
						"id":              types.StringType,
						"name":            types.StringType,
						"email":           types.StringType,
						"service_account": types.BoolType,
						"mfa_enabled":     types.BoolType,
						"status":          types.StringType,
						"created_at":      types.StringType,
						"modified_at":     types.StringType,
						"title":           types.StringType,
						"handle":          types.StringType,
						"disabled":        types.BoolType,
						"verified":        types.BoolType,
						"icon":            types.StringType,
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
			ID:             types.StringValue(user.GetId()),
			Name:           types.StringValue(user.Attributes.GetName()),
			Email:          types.StringValue(user.Attributes.GetEmail()),
			ServiceAccount: types.BoolValue(user.Attributes.GetServiceAccount()),
			MfaEnabled:     types.BoolValue(user.Attributes.GetMfaEnabled()),
			Status:         types.StringValue(user.Attributes.GetStatus()),
			CreatedAt:      types.StringValue(user.Attributes.GetCreatedAt().Format(time.RFC3339)),
			ModifiedAt:     types.StringValue(user.Attributes.GetModifiedAt().Format(time.RFC3339)),
			Title:          types.StringValue(user.Attributes.GetTitle()),
			Handle:         types.StringValue(user.Attributes.GetHandle()),
			Disabled:       types.BoolValue(user.Attributes.GetDisabled()),
			Verified:       types.BoolValue(user.Attributes.GetVerified()),
			Icon:           types.StringValue(user.Attributes.GetIcon()),
		}

		users = append(users, &u)
	}

	hashingData := fmt.Sprintf("%s:%s", state.Filter.ValueString(), state.FilterStatus.ValueString())

	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Users = users
}
