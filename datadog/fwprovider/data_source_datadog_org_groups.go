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

var _ datasource.DataSource = &datadogOrgGroupsDataSource{}

type datadogOrgGroupsDataSource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupItemModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	OwnerOrgSite types.String `tfsdk:"owner_org_site"`
	OwnerOrgUuid types.String `tfsdk:"owner_org_uuid"`
}

type datadogOrgGroupsDataSourceModel struct {
	ID     types.String         `tfsdk:"id"`
	Groups []*OrgGroupItemModel `tfsdk:"groups"`
}

func NewDatadogOrgGroupsDataSource() datasource.DataSource {
	return &datadogOrgGroupsDataSource{}
}

func (d *datadogOrgGroupsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogOrgGroupsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "org_groups"
}

func (d *datadogOrgGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve all org groups in the organization.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"groups": schema.ListAttribute{
				Computed:    true,
				Description: "The list of org groups.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":             types.StringType,
						"name":           types.StringType,
						"owner_org_site": types.StringType,
						"owner_org_uuid": types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogOrgGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogOrgGroupsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	const pageSize = int64(100)
	var groups []datadogV2.OrgGroupData
	for page := int64(0); ; page++ {
		opts := datadogV2.NewListOrgGroupsOptionalParameters().WithPageNumber(page).WithPageSize(pageSize)
		resp, _, err := d.API.ListOrgGroups(d.Auth, *opts)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing org groups"))
			return
		}
		data := resp.GetData()
		groups = append(groups, data...)
		if int64(len(data)) < pageSize {
			break
		}
	}

	items := make([]*OrgGroupItemModel, 0, len(groups))
	for _, g := range groups {
		attrs := g.GetAttributes()
		items = append(items, &OrgGroupItemModel{
			ID:           types.StringValue(g.GetId().String()),
			Name:         types.StringValue(attrs.GetName()),
			OwnerOrgSite: types.StringValue(attrs.GetOwnerOrgSite()),
			OwnerOrgUuid: types.StringValue(attrs.GetOwnerOrgUuid().String()),
		})
	}

	state.ID = types.StringValue("all")
	state.Groups = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
