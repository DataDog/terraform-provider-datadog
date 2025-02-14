package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &csmThreatsPoliciesListResource{}
	_ resource.ResourceWithImportState = &csmThreatsPoliciesListResource{}
)

type csmThreatsPoliciesListResource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsPoliciesListModel struct {
	ID      types.String                       `tfsdk:"id"`
	Entries []csmThreatsPoliciesListEntryModel `tfsdk:"entries"`
}

type csmThreatsPoliciesListEntryModel struct {
	PolicyID types.String `tfsdk:"policy_id"`
	Name     types.String `tfsdk:"name"`
	Priority types.Int64  `tfsdk:"priority"`
}

func NewCSMThreatsPoliciesListResource() resource.Resource {
	return &csmThreatsPoliciesListResource{}
}

func (r *csmThreatsPoliciesListResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "csm_threats_policies_list"
}

func (r *csmThreatsPoliciesListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (r *csmThreatsPoliciesListResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *csmThreatsPoliciesListResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog CSM Threats policies API resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"entries": schema.SetNestedBlock{
				Description: "A set of policies that belong to this list. Only one policies_list resource can be defined in Terraform, containing all unique policies. All non-listed policies get deleted.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"policy_id": schema.StringAttribute{
							Description: "The ID of the policy to manage (from csm_threats_policy).",
							Required:    true,
						},
						"priority": schema.Int64Attribute{
							Description: "The priority of the policy in this list.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Optional name. If omitted, fallback to the policy_id as name.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *csmThreatsPoliciesListResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan csmThreatsPoliciesListModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue("policies_list")

	updatedEntries, err := r.applyBatchPolicies(ctx, plan.Entries, &response.Diagnostics)
	if err != nil {
		return
	}

	plan.Entries = updatedEntries
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *csmThreatsPoliciesListResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsPoliciesListModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if state.ID.IsUnknown() || state.ID.IsNull() || state.ID.ValueString() == "" {
		response.State.RemoveResource(ctx)
		return
	}

	listResponse, httpResp, err := r.api.ListCSMThreatsAgentPolicies(r.auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching agent policies"))
		return
	}

	newEntries := make([]csmThreatsPoliciesListEntryModel, 0)
	for _, policyData := range listResponse.GetData() {
		policyID := policyData.GetId()
		if policyID == "CWS_DD" {
			continue
		}
		attributes := policyData.Attributes

		name := attributes.GetName()
		priorirty := attributes.GetPriority()

		entry := csmThreatsPoliciesListEntryModel{
			PolicyID: types.StringValue(policyID),
			Name:     types.StringValue(name),
			Priority: types.Int64Value(int64(priorirty)),
		}
		newEntries = append(newEntries, entry)
	}

	state.Entries = newEntries
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsPoliciesListResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan csmThreatsPoliciesListModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updatedEntries, err := r.applyBatchPolicies(ctx, plan.Entries, &response.Diagnostics)
	if err != nil {
		return
	}

	plan.Entries = updatedEntries
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *csmThreatsPoliciesListResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	_, err := r.applyBatchPolicies(ctx, []csmThreatsPoliciesListEntryModel{}, &response.Diagnostics)
	if err != nil {
		return
	}
	response.State.RemoveResource(ctx)
}

func (r *csmThreatsPoliciesListResource) applyBatchPolicies(ctx context.Context, entries []csmThreatsPoliciesListEntryModel, diags *diag.Diagnostics) ([]csmThreatsPoliciesListEntryModel, error) {
	listResp, httpResp, err := r.api.ListCSMThreatsAgentPolicies(r.auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			diags.Append(utils.FrameworkErrorDiag(err, "error while fetching agent policies"))
			return nil, err
		}
	}

	existingPolicies := make(map[string]struct{})
	for _, policy := range listResp.GetData() {
		if policy.GetId() == "CWS_DD" {
			continue
		}
		existingPolicies[policy.GetId()] = struct{}{}
	}

	var batchItems []datadogV2.CloudWorkloadSecurityAgentPolicyBatchUpdateAttributesPoliciesItems

	for i := range entries {
		policyID := entries[i].PolicyID.ValueString()
		name := entries[i].Name.ValueString()

		if name == "" {
			name = policyID
			entries[i].Name = types.StringValue(name)
		}
		priority := entries[i].Priority.ValueInt64()

		item := datadogV2.CloudWorkloadSecurityAgentPolicyBatchUpdateAttributesPoliciesItems{
			Id:       &policyID,
			Name:     &name,
			Priority: &priority,
		}

		batchItems = append(batchItems, item)
		delete(existingPolicies, policyID)
	}

	for policyID := range existingPolicies {
		DeleteTrue := true
		item := datadogV2.CloudWorkloadSecurityAgentPolicyBatchUpdateAttributesPoliciesItems{
			Id:     &policyID,
			Delete: &DeleteTrue,
		}
		batchItems = append(batchItems, item)
	}

	patchID := "batch_update_req"
	typ := datadogV2.CLOUDWORKLOADSECURITYAGENTPOLICYBATCHUPDATEDATATYPE_POLICIES
	attributes := datadogV2.NewCloudWorkloadSecurityAgentPolicyBatchUpdateAttributes()
	attributes.SetPolicies(batchItems)
	data := datadogV2.NewCloudWorkloadSecurityAgentPolicyBatchUpdateData(*attributes, patchID, typ)
	batchReq := datadogV2.NewCloudWorkloadSecurityAgentPolicyBatchUpdateRequest(*data)

	response, _, err := r.api.BatchUpdateCSMThreatsAgentPolicy(
		r.auth,
		*batchReq,
	)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error while applying batch policies"))
		return nil, err
	}

	finalEntries := make([]csmThreatsPoliciesListEntryModel, 0)
	for _, policy := range response.GetData() {
		policyID := policy.GetId()
		attributes := policy.Attributes

		name := ""
		if attributes.GetName() == "" {
			name = policyID
		}
		name = attributes.GetName()
		priority := attributes.GetPriority()

		entry := csmThreatsPoliciesListEntryModel{
			PolicyID: types.StringValue(policyID),
			Name:     types.StringValue(name),
			Priority: types.Int64Value(int64(priority)),
		}
		finalEntries = append(finalEntries, entry)
	}

	return finalEntries, nil
}
