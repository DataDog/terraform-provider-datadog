package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &datadogOrgGroupPoliciesDataSource{}

type datadogOrgGroupPoliciesDataSource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupPolicyItemModel struct {
	ID              types.String `tfsdk:"id"`
	OrgGroupID      types.String `tfsdk:"org_group_id"`
	PolicyName      types.String `tfsdk:"policy_name"`
	Content         types.String `tfsdk:"content"`
	EnforcementTier types.String `tfsdk:"enforcement_tier"`
	PolicyType      types.String `tfsdk:"policy_type"`
}

type datadogOrgGroupPoliciesDataSourceModel struct {
	// Query parameters
	OrgGroupID types.String `tfsdk:"org_group_id"`
	PolicyName types.String `tfsdk:"policy_name"`

	// Results
	ID       types.String               `tfsdk:"id"`
	Policies []*OrgGroupPolicyItemModel `tfsdk:"policies"`
}

func NewDatadogOrgGroupPoliciesDataSource() datasource.DataSource {
	return &datadogOrgGroupPoliciesDataSource{}
}

func (d *datadogOrgGroupPoliciesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogOrgGroupPoliciesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "org_group_policies"
}

func (d *datadogOrgGroupPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve the policies attached to an org group, optionally filtered by policy name.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group whose policies to list.",
			},
			"policy_name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter policies by name.",
			},
			"policies": schema.ListAttribute{
				Computed:    true,
				Description: "The list of policies.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":               types.StringType,
						"org_group_id":     types.StringType,
						"policy_name":      types.StringType,
						"content":          types.StringType,
						"enforcement_tier": types.StringType,
						"policy_type":      types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogOrgGroupPoliciesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogOrgGroupPoliciesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	orgGroupID, err := uuid.Parse(state.OrgGroupID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
		return
	}

	opts := datadogV2.NewListOrgGroupPoliciesOptionalParameters()
	if !state.PolicyName.IsNull() {
		opts.WithFilterPolicyName(state.PolicyName.ValueString())
	}

	const pageSize = int64(100)
	const maxPages = int64(100)

	var policies []datadogV2.OrgGroupPolicyData
	for page := int64(0); page < maxPages; page++ {
		opts.WithPageNumber(page).WithPageSize(pageSize)
		resp, _, err := d.API.ListOrgGroupPolicies(d.Auth, orgGroupID, *opts)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing org group policies"))
			return
		}
		data := resp.GetData()
		policies = append(policies, data...)
		if int64(len(data)) < pageSize {
			break
		}
	}

	items := make([]*OrgGroupPolicyItemModel, 0, len(policies))
	for _, p := range policies {
		attrs := p.GetAttributes()
		item := &OrgGroupPolicyItemModel{
			ID:              types.StringValue(p.GetId().String()),
			PolicyName:      types.StringValue(attrs.GetPolicyName()),
			EnforcementTier: types.StringValue(string(attrs.GetEnforcementTier())),
			PolicyType:      types.StringValue(string(attrs.GetPolicyType())),
		}

		contentBytes, err := json.Marshal(attrs.GetContent())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error marshaling policy content"))
			return
		}
		item.Content = types.StringValue(string(contentBytes))

		// OrgGroupID left as null if the API omitted the relationship — distinguishable
		// from an empty string so callers can detect server data integrity issues.
		if rels, ok := p.GetRelationshipsOk(); ok && rels != nil {
			if orgGroup, ok := rels.GetOrgGroupOk(); ok && orgGroup != nil {
				ogData := orgGroup.GetData()
				item.OrgGroupID = types.StringValue(ogData.GetId().String())
			}
		}

		items = append(items, item)
	}

	state.ID = types.StringValue(synthesizeOrgGroupPoliciesID(state))
	state.Policies = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func synthesizeOrgGroupPoliciesID(state datadogOrgGroupPoliciesDataSourceModel) string {
	id := state.OrgGroupID.ValueString()
	if !state.PolicyName.IsNull() {
		id = fmt.Sprintf("%s:%s", id, state.PolicyName.ValueString())
	}
	return id
}
