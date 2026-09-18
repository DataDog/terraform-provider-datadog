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
	_ datasource.DataSource = &currentUserDataSource{}
)

// NewDatadogCurrentUserDataSource creates a new current user data source
func NewDatadogCurrentUserDataSource() datasource.DataSource {
	return &currentUserDataSource{}
}

// currentUserDataSourceModel represents the Terraform state for the current user data source
type currentUserDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Email          types.String `tfsdk:"email"`
	Name           types.String `tfsdk:"name"`
	Handle         types.String `tfsdk:"handle"`
	ServiceAccount types.Bool   `tfsdk:"service_account"`
	OrgId          types.String `tfsdk:"org_id"`
	OrgPublicId    types.String `tfsdk:"org_public_id"`
	OrgName        types.String `tfsdk:"org_name"`
}

// currentUserDataSource is the data source implementation
type currentUserDataSource struct {
	Api  *utils.ApiInstances
	Auth context.Context
}

// Configure sets up the data source with provider data
func (d *currentUserDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances
	d.Auth = providerData.Auth
}

// Metadata returns the data source type name
func (d *currentUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "current_user"
}

// Schema defines the data source schema
func (d *currentUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve information about the user associated with the authentication context used by " +
			"the Datadog provider. This data source is also useful for retrieving organization metadata.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"email": schema.StringAttribute{
				Description: "Email of the user.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the user.",
				Computed:    true,
			},
			"handle": schema.StringAttribute{
				Description: "The user's handle.",
				Computed:    true,
			},
			"service_account": schema.BoolAttribute{
				Description: "Indicates whether the user is a service account.",
				Computed:    true,
			},
			"org_id": schema.StringAttribute{
				Description: "The UUID of the user's organization.",
				Computed:    true,
			},
			"org_public_id": schema.StringAttribute{
				Description: "The public_id of the user's organization.",
				Computed:    true,
			},
			"org_name": schema.StringAttribute{
				Description: "Name of the user's organization.",
				Computed:    true,
			},
		},
	}
}

// Read retrieves the current user's information from the Datadog API
func (d *currentUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state currentUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userResp, httpResp, err := d.Api.GetUsersApiV2().GetCurrentUser(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error getting current user"), ""))
		return
	}

	user := userResp.GetData()
	attributes := user.GetAttributes()

	state.ID = types.StringValue(user.GetId())
	state.Email = types.StringValue(attributes.GetEmail())
	state.Name = types.StringValue(attributes.GetName())
	state.Handle = types.StringValue(attributes.GetHandle())
	state.ServiceAccount = types.BoolValue(attributes.GetServiceAccount())

	// The current user's organization is the only "orgs"-type resource in the
	// included array.
	for _, included := range userResp.GetIncluded() {
		if org, ok := included.GetActualInstance().(*datadogV2.Organization); ok {
			orgAttributes := org.GetAttributes()
			state.OrgId = types.StringValue(org.GetId())
			state.OrgPublicId = types.StringValue(orgAttributes.GetPublicId())
			state.OrgName = types.StringValue(orgAttributes.GetName())
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
