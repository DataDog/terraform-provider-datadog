package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &datadogOrgGroupMembershipsDataSource{}

type datadogOrgGroupMembershipsDataSource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupMembershipItemModel struct {
	ID         types.String `tfsdk:"id"`
	OrgGroupID types.String `tfsdk:"org_group_id"`
	OrgUuid    types.String `tfsdk:"org_uuid"`
	OrgSite    types.String `tfsdk:"org_site"`
	OrgName    types.String `tfsdk:"org_name"`
}

type datadogOrgGroupMembershipsDataSourceModel struct {
	// Query parameters
	OrgGroupID types.String `tfsdk:"org_group_id"`
	OrgUuid    types.String `tfsdk:"org_uuid"`

	// Results
	ID          types.String                   `tfsdk:"id"`
	Memberships []*OrgGroupMembershipItemModel `tfsdk:"memberships"`
}

func NewDatadogOrgGroupMembershipsDataSource() datasource.DataSource {
	return &datadogOrgGroupMembershipsDataSource{}
}

func (d *datadogOrgGroupMembershipsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogOrgGroupMembershipsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "org_group_memberships"
}

func (d *datadogOrgGroupMembershipsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve org group memberships filtered by org group or organization. At least one of `org_group_id` or `org_uuid` is required.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter memberships to those within the given org group. At least one of `org_group_id` or `org_uuid` is required.",
				Validators:  []validator.String{uuidValidator},
			},
			"org_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "Filter memberships to those for the given organization. At least one of `org_group_id` or `org_uuid` is required.",
				Validators:  []validator.String{uuidValidator},
			},
			"memberships": schema.ListAttribute{
				Computed:    true,
				Description: "The list of org group memberships.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"org_group_id": types.StringType,
						"org_uuid":     types.StringType,
						"org_site":     types.StringType,
						"org_name":     types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogOrgGroupMembershipsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogOrgGroupMembershipsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	opts := datadogV2.NewListOrgGroupMembershipsOptionalParameters()
	if !state.OrgGroupID.IsNull() {
		parsed, err := uuid.Parse(state.OrgGroupID.ValueString())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
			return
		}
		opts.WithFilterOrgGroupId(parsed)
	}
	if !state.OrgUuid.IsNull() {
		parsed, err := uuid.Parse(state.OrgUuid.ValueString())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_uuid must be a valid UUID"))
			return
		}
		opts.WithFilterOrgUuid(parsed)
	}

	const pageSize = int64(100)
	var memberships []datadogV2.OrgGroupMembershipData
	for page := int64(0); ; page++ {
		opts.WithPageNumber(page).WithPageSize(pageSize)
		resp, _, err := d.API.ListOrgGroupMemberships(d.Auth, *opts)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing org group memberships"))
			return
		}
		data := resp.GetData()
		memberships = append(memberships, data...)
		if int64(len(data)) < pageSize {
			break
		}
	}

	items := make([]*OrgGroupMembershipItemModel, 0, len(memberships))
	for _, m := range memberships {
		attrs := m.GetAttributes()
		item := &OrgGroupMembershipItemModel{
			ID:      types.StringValue(m.GetId().String()),
			OrgUuid: types.StringValue(attrs.GetOrgUuid().String()),
			OrgSite: types.StringValue(attrs.GetOrgSite()),
			OrgName: types.StringValue(attrs.GetOrgName()),
		}
		rels, ok := m.GetRelationshipsOk()
		if !ok || rels == nil {
			response.Diagnostics.AddError("datadog_org_group_memberships: response missing relationships", fmt.Sprintf("membership %s has no relationships block", item.ID.ValueString()))
			return
		}
		orgGroup, ok := rels.GetOrgGroupOk()
		if !ok || orgGroup == nil {
			response.Diagnostics.AddError("datadog_org_group_memberships: response missing org_group relationship", fmt.Sprintf("membership %s has no org_group relationship", item.ID.ValueString()))
			return
		}
		ogData, ok := orgGroup.GetDataOk()
		if !ok {
			response.Diagnostics.AddError("datadog_org_group_memberships: response missing org_group.data", fmt.Sprintf("membership %s has no org_group.data", item.ID.ValueString()))
			return
		}
		item.OrgGroupID = types.StringValue(ogData.GetId().String())
		items = append(items, item)
	}

	state.ID = types.StringValue(synthesizeOrgGroupMembershipsID(state))
	state.Memberships = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func synthesizeOrgGroupMembershipsID(state datadogOrgGroupMembershipsDataSourceModel) string {
	parts := make([]string, 0, 2)
	if !state.OrgGroupID.IsNull() {
		parts = append(parts, state.OrgGroupID.ValueString())
	}
	if !state.OrgUuid.IsNull() {
		parts = append(parts, state.OrgUuid.ValueString())
	}
	if len(parts) == 0 {
		return "all"
	}
	return strings.Join(parts, ":")
}
