package fwprovider

import (
	"context"
	"encoding/json"
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

var _ datasource.DataSource = &datadogOrgGroupPolicyOverridesDataSource{}

type datadogOrgGroupPolicyOverridesDataSource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupPolicyOverrideItemModel struct {
	ID         types.String `tfsdk:"id"`
	OrgGroupID types.String `tfsdk:"org_group_id"`
	PolicyID   types.String `tfsdk:"policy_id"`
	OrgUuid    types.String `tfsdk:"org_uuid"`
	OrgSite    types.String `tfsdk:"org_site"`
	Content    types.String `tfsdk:"content"`
}

type datadogOrgGroupPolicyOverridesDataSourceModel struct {
	// Query parameters
	OrgGroupID types.String `tfsdk:"org_group_id"`
	PolicyID   types.String `tfsdk:"policy_id"`
	OrgUuid    types.String `tfsdk:"org_uuid"`

	// Results
	ID        types.String                       `tfsdk:"id"`
	Overrides []*OrgGroupPolicyOverrideItemModel `tfsdk:"overrides"`
}

func NewDatadogOrgGroupPolicyOverridesDataSource() datasource.DataSource {
	return &datadogOrgGroupPolicyOverridesDataSource{}
}

func (d *datadogOrgGroupPolicyOverridesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogOrgGroupPolicyOverridesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "org_group_policy_overrides"
}

func (d *datadogOrgGroupPolicyOverridesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve org group policy overrides. Supports filtering by policy ID (server-side) and organization UUID (client-side).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group whose overrides to list.",
				Validators:  []validator.String{uuidValidator},
			},
			"policy_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter overrides to those on the given policy.",
				Validators:  []validator.String{uuidValidator},
			},
			"org_uuid": schema.StringAttribute{
				Optional:    true,
				Description: "Filter overrides to those for the given organization. Applied client-side after the List call since the API does not accept an org_uuid filter on this endpoint.",
				Validators:  []validator.String{uuidValidator},
			},
			"overrides": schema.ListAttribute{
				Computed:    true,
				Description: "The list of policy overrides.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"org_group_id": types.StringType,
						"policy_id":    types.StringType,
						"org_uuid":     types.StringType,
						"org_site":     types.StringType,
						"content":      types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogOrgGroupPolicyOverridesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogOrgGroupPolicyOverridesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	orgGroupID, err := uuid.Parse(state.OrgGroupID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
		return
	}

	opts := datadogV2.NewListOrgGroupPolicyOverridesOptionalParameters()
	if !state.PolicyID.IsNull() {
		parsed, err := uuid.Parse(state.PolicyID.ValueString())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "policy_id must be a valid UUID"))
			return
		}
		opts.WithFilterPolicyId(parsed)
	}

	// Parse the client-side org_uuid filter up front so bad values fail before any API call.
	var orgUuidFilter uuid.UUID
	if !state.OrgUuid.IsNull() {
		parsed, err := uuid.Parse(state.OrgUuid.ValueString())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_uuid must be a valid UUID"))
			return
		}
		orgUuidFilter = parsed
	}

	// See datadog_org_groups for the pagination invariant (duplicate row = bail).
	const pageSize = int64(100)
	seen := make(map[string]struct{})

	var overrides []datadogV2.OrgGroupPolicyOverrideData
	for page := int64(0); ; page++ {
		opts.WithPageNumber(page).WithPageSize(pageSize)
		resp, _, err := d.API.ListOrgGroupPolicyOverrides(d.Auth, orgGroupID, *opts)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing org group policy overrides"))
			return
		}
		data := resp.GetData()
		for _, item := range data {
			id := item.GetId().String()
			if _, ok := seen[id]; ok {
				response.Diagnostics.AddError(
					"datadog_org_group_policy_overrides: pagination returned duplicate row",
					fmt.Sprintf("override %s appeared on more than one page; aborting to avoid an infinite loop", id),
				)
				return
			}
			seen[id] = struct{}{}
			overrides = append(overrides, item)
		}
		if int64(len(data)) < pageSize {
			break
		}
	}

	items := make([]*OrgGroupPolicyOverrideItemModel, 0, len(overrides))
	for _, o := range overrides {
		attrs := o.GetAttributes()
		ou := attrs.GetOrgUuid()
		// Defensive: flag zero-UUID rows. The server should never return these, so
		// hitting this branch indicates a malformed response rather than a filter miss.
		if ou == uuid.Nil {
			response.Diagnostics.AddWarning(
				"datadog_org_group_policy_overrides: skipping row with zero org_uuid",
				fmt.Sprintf("override %s returned a zero UUID for org_uuid; server-side data integrity issue", o.GetId().String()),
			)
			continue
		}
		if orgUuidFilter != uuid.Nil && ou != orgUuidFilter {
			continue
		}

		item := &OrgGroupPolicyOverrideItemModel{
			ID:      types.StringValue(o.GetId().String()),
			OrgUuid: types.StringValue(ou.String()),
			OrgSite: types.StringValue(attrs.GetOrgSite()),
		}

		if attrs.HasContent() {
			bytes, err := json.Marshal(attrs.GetContent())
			if err != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error marshaling override content"))
				return
			}
			item.Content = types.StringValue(string(bytes))
		} else {
			item.Content = types.StringValue("{}")
		}

		rels, ok := o.GetRelationshipsOk()
		if !ok || rels == nil {
			response.Diagnostics.AddError("datadog_org_group_policy_overrides: response missing relationships", fmt.Sprintf("override %s has no relationships block", item.ID.ValueString()))
			return
		}
		orgGroup, ok := rels.GetOrgGroupOk()
		if !ok || orgGroup == nil {
			response.Diagnostics.AddError("datadog_org_group_policy_overrides: response missing org_group relationship", fmt.Sprintf("override %s has no org_group relationship", item.ID.ValueString()))
			return
		}
		ogData, ok := orgGroup.GetDataOk()
		if !ok {
			response.Diagnostics.AddError("datadog_org_group_policy_overrides: response missing org_group.data", fmt.Sprintf("override %s has no org_group.data", item.ID.ValueString()))
			return
		}
		item.OrgGroupID = types.StringValue(ogData.GetId().String())

		policy, ok := rels.GetOrgGroupPolicyOk()
		if !ok || policy == nil {
			response.Diagnostics.AddError("datadog_org_group_policy_overrides: response missing org_group_policy relationship", fmt.Sprintf("override %s has no org_group_policy relationship", item.ID.ValueString()))
			return
		}
		pData, ok := policy.GetDataOk()
		if !ok {
			response.Diagnostics.AddError("datadog_org_group_policy_overrides: response missing org_group_policy.data", fmt.Sprintf("override %s has no org_group_policy.data", item.ID.ValueString()))
			return
		}
		item.PolicyID = types.StringValue(pData.GetId().String())

		items = append(items, item)
	}

	state.ID = types.StringValue(synthesizeOrgGroupPolicyOverridesID(state))
	state.Overrides = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func synthesizeOrgGroupPolicyOverridesID(state datadogOrgGroupPolicyOverridesDataSourceModel) string {
	parts := []string{state.OrgGroupID.ValueString()}
	if !state.PolicyID.IsNull() {
		parts = append(parts, state.PolicyID.ValueString())
	}
	if !state.OrgUuid.IsNull() {
		parts = append(parts, state.OrgUuid.ValueString())
	}
	return strings.Join(parts, ":")
}
