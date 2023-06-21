package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogServiceAccountDatasource{}
)

func NewDatadogServiceAccountDatasource() datasource.DataSource {
	return &datadogServiceAccountDatasource{}
}

type datadogServiceAccountDatasourceModel struct {
	// Query Parameters
	ID           types.String `tfsdk:"id"`
	Filter       types.String `tfsdk:"filter"`
	FilterStatus types.String `tfsdk:"filter_status"`
	// Results
	Disabled types.Bool   `tfsdk:"disabled"`
	Email    types.String `tfsdk:"email"`
	Handle   types.String `tfsdk:"handle"`
	Icon     types.String `tfsdk:"icon"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	Title    types.String `tfsdk:"title"`
	Verified types.Bool   `tfsdk:"verified"`
	Roles    types.List   `tfsdk:"roles"`
}

type datadogServiceAccountDatasource struct {
	Api  *datadogV2.UsersApi
	Auth context.Context
}

func (r *datadogServiceAccountDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetUsersApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogServiceAccountDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "service_account"
}

func (d *datadogServiceAccountDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog service account.",
		Attributes: map[string]schema.Attribute{
			// Datasource Parameters
			"id": schema.StringAttribute{
				Description: "The service account's ID.",
				Optional:    true,
				Computed:    true,
			},
			"filter": schema.StringAttribute{
				Description: "Filter all users and service accounts by name, email, or role.",
				Optional:    true,
			},
			"filter_status": schema.StringAttribute{
				Description: "Filter on status attribute. Comma separated list, with possible values `Active`, `Pending`, and `Disabled`.",
				Optional:    true,
			},
			// Computed values
			"disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the user is disabled.",
			},
			"email": schema.StringAttribute{
				Description: "Email of the user.",
				Computed:    true,
			},
			"handle": schema.StringAttribute{
				Description: "Handle of the user.",
				Computed:    true,
			},
			"icon": schema.StringAttribute{
				Description: "URL of the user's icon.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the user.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the user.",
			},
			"title": schema.StringAttribute{
				Computed:    true,
				Description: "Title of the user.",
			},
			"verified": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the user is verified.",
			},
			"roles": schema.ListAttribute{
				Computed:    true,
				Description: "Roles assigned to this service account.",
				ElementType: types.StringType,
			},
		},
	}
}

func (d *datadogServiceAccountDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogServiceAccountDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var userData *datadogV2.User
	if !state.ID.IsNull() {
		serviceAccountID := state.ID.ValueString()
		ddResp, _, err := d.Api.GetUser(d.Auth, serviceAccountID)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog service account by ID"))
			return
		}
		attr := ddResp.Data.GetAttributes()
		if !attr.GetServiceAccount() {
			resp.Diagnostics.AddError("Obtained entity was not a service account", "")
			return
		}
		userData = ddResp.Data
	} else {
		optionalParams := datadogV2.ListUsersOptionalParameters{}
		optionalParams.WithFilter(state.Filter.ValueString())
		if !state.FilterStatus.IsNull() {
			optionalParams.WithFilterStatus(state.FilterStatus.ValueString())
		}

		ddResp, _, err := d.Api.ListUsers(d.Auth, optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datadog users and service accounts"))
			return
		}

		var serviceAccounts []datadogV2.User
		for _, user := range ddResp.Data {
			attr := user.GetAttributes()
			if attr.GetServiceAccount() {
				serviceAccounts = append(serviceAccounts, user)
			}
		}
		if len(serviceAccounts) > 1 {
			resp.Diagnostics.AddError("filter keyword returned more than one result, use more specific search criteria", "")
			return
		}
		if len(serviceAccounts) == 0 {
			resp.Diagnostics.AddError("filter keyword returned no results", "")
			return
		}
		userData = &serviceAccounts[0]
	}
	d.updateState(ctx, &state, userData)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogServiceAccountDatasource) updateState(ctx context.Context, state *datadogServiceAccountDatasourceModel, userData *datadogV2.User) {
	state.ID = types.StringValue(userData.GetId())
	attributes := userData.GetAttributes()
	if v, ok := attributes.GetDisabledOk(); ok && v != nil {
		state.Disabled = types.BoolValue(*v)
	}
	if v, ok := attributes.GetEmailOk(); ok && v != nil {
		state.Email = types.StringValue(*v)
	}
	if v, ok := attributes.GetHandleOk(); ok && v != nil {
		state.Handle = types.StringValue(*v)
	}
	if v, ok := attributes.GetIconOk(); ok && v != nil {
		state.Icon = types.StringValue(*v)
	}
	if v, ok := attributes.GetNameOk(); ok && v != nil {
		state.Name = types.StringValue(*v)
	}
	if v, ok := attributes.GetStatusOk(); ok && v != nil {
		state.Status = types.StringValue(*v)
	}
	if v, ok := attributes.GetTitleOk(); ok && v != nil {
		state.Title = types.StringValue(*v)
	}
	if v, ok := attributes.GetVerifiedOk(); ok && v != nil {
		state.Verified = types.BoolValue(*v)
	}
	var roles []string
	if v, ok := userData.GetRelationshipsOk(); ok {
		if r, ok := v.GetRolesOk(); ok {
			if data, ok := r.GetDataOk(); ok {
				for _, v := range *data {
					roles = append(roles, v.GetId())
				}
			}
		}
	}
	state.Roles, _ = types.ListValueFrom(ctx, types.StringType, roles)
}
