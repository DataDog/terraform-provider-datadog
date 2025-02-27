package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	ID       types.String                 `tfsdk:"id"`
	Policies []csmThreatsPolicyEntryModel `tfsdk:"policies"`
}

type csmThreatsPolicyEntryModel struct {
	PolicyLabel types.String `tfsdk:"policy_label"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Tags        types.Set    `tfsdk:"tags"`
}

func NewCSMThreatsPoliciesListResource() resource.Resource {
	return &csmThreatsPoliciesListResource{}
}

func (r *csmThreatsPoliciesListResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "csm_threats_policies"
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
		Description: "Manages multiple Datadog CSM Threats policies in a single resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"policies": schema.SetNestedBlock{
				Description: "Set of policy blocks. Each block requires a unique policy_label.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"policy_label": schema.StringAttribute{
							Description: "The ID of the policy to manage (from csm_threats_policy).",
							Required:    true,
						},
						"id": schema.StringAttribute{
							Description: "The Datadog-assigned policy ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the policy.",
							Optional:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description for the policy.",
							Optional:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Indicates whether the policy is enabled.",
							Optional:    true,
							Default:     booldefault.StaticBool(false),
							Computed:    true,
						},
						"tags": schema.SetAttribute{
							Description: "Host tags that define where the policy is deployed.",
							Optional:    true,
							ElementType: types.StringType,
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

	updatedPolicies, err := r.applyBatchPolicies(ctx, []csmThreatsPolicyEntryModel{}, plan.Policies, &response.Diagnostics)
	if err != nil {
		return
	}

	plan.Policies = updatedPolicies
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *csmThreatsPoliciesListResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsPoliciesListModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
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

	apiMap := make(map[string]datadogV2.CloudWorkloadSecurityAgentPolicyAttributes)
	for _, policy := range listResponse.GetData() {
		policyID := policy.GetId()
		if policy.Attributes != nil {
			apiMap[policyID] = *policy.Attributes
		}
	}

	newPolicies := make([]csmThreatsPolicyEntryModel, 0, len(state.Policies))

	// update the state with the latest data from the API, but only for the policies that are already present in the state
	for _, policy := range state.Policies {
		policyID := policy.ID.ValueString()
		attr, found := apiMap[policyID]
		if !found {
			// policy was deleted outside of Terraform
			continue
		}

		tags, _ := types.SetValueFrom(ctx, types.StringType, attr.GetHostTags())
		newPolicies = append(newPolicies, csmThreatsPolicyEntryModel{
			PolicyLabel: policy.PolicyLabel,
			ID:          types.StringValue(policyID),
			Name:        types.StringValue(attr.GetName()),
			Description: types.StringValue(attr.GetDescription()),
			Enabled:     types.BoolValue(attr.GetEnabled()),
			Tags:        tags,
		})
	}

	state.Policies = newPolicies
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsPoliciesListResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, old csmThreatsPoliciesListModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	updatedPolicies, err := r.applyBatchPolicies(ctx, old.Policies, plan.Policies, &response.Diagnostics)
	if err != nil {
		return
	}

	plan.Policies = updatedPolicies
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *csmThreatsPoliciesListResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state csmThreatsPoliciesListModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := r.applyBatchPolicies(ctx, state.Policies, []csmThreatsPolicyEntryModel{}, &response.Diagnostics)
	if err != nil {
		return
	}
	response.State.RemoveResource(ctx)
}

func (r *csmThreatsPoliciesListResource) applyBatchPolicies(ctx context.Context, oldPolicies []csmThreatsPolicyEntryModel, newPolicies []csmThreatsPolicyEntryModel, diags *diag.Diagnostics) ([]csmThreatsPolicyEntryModel, error) {
	oldPoliciesMap := make(map[string]csmThreatsPolicyEntryModel)
	for _, policy := range oldPolicies {
		oldPoliciesMap[policy.PolicyLabel.ValueString()] = policy
	}

	newPoliciesMap := make(map[string]csmThreatsPolicyEntryModel)
	for _, policy := range newPolicies {
		newPoliciesMap[policy.PolicyLabel.ValueString()] = policy
	}

	// check policies that should be deleted (present in old but not in new)
	var toDelete []csmThreatsPolicyEntryModel

	for policyLabel, oldPolicy := range oldPoliciesMap {
		if _, found := newPoliciesMap[policyLabel]; !found {
			toDelete = append(toDelete, oldPolicy)
		}
	}

	// add policies that should be created or updated (even if they are not modified, we send all policies in the batch request)
	var toUpsert []csmThreatsPolicyEntryModel

	// get IDs of existing policies
	for _, policy := range newPolicies {
		policyLabel := policy.PolicyLabel.ValueString()
		if oldPolicy, found := oldPoliciesMap[policyLabel]; found {
			policy.ID = oldPolicy.ID
		}
		toUpsert = append(toUpsert, policy)
	}

	var batchItems []datadogV2.CloudWorkloadSecurityAgentPolicyBatchUpdateAttributesPoliciesItems

	// add deleted policies to the batch request
	for _, policy := range toDelete {
		policyID := policy.PolicyLabel.ValueString()
		DeleteTrue := true
		item := datadogV2.CloudWorkloadSecurityAgentPolicyBatchUpdateAttributesPoliciesItems{
			Id:     &policyID,
			Delete: &DeleteTrue,
		}
		batchItems = append(batchItems, item)
	}

	// add updated or new policies to the batch request
	for _, policy := range toUpsert {
		policyID := policy.ID.ValueString()
		name := policy.Name.ValueString()
		description := policy.Description.ValueString()
		enabled := policy.Enabled.ValueBool()
		var tags []string
		if !policy.Tags.IsNull() && !policy.Tags.IsUnknown() {
			for _, tag := range policy.Tags.Elements() {
				tagStr, ok := tag.(types.String)
				if !ok {
					return nil, fmt.Errorf("expected item to be of type types.String, got %T", tag)
				}
				tags = append(tags, tagStr.ValueString())
			}
		}

		items := datadogV2.CloudWorkloadSecurityAgentPolicyBatchUpdateAttributesPoliciesItems{
			Name:        &name,
			Description: &description,
			Enabled:     &enabled,
			HostTags:    tags,
		}
		// if policyID is not empty, it means it's not a new policy: we add the id parameter to the request
		if policyID != "" {
			items.Id = &policyID
		}
		batchItems = append(batchItems, items)
	}

	if len(batchItems) == 0 {
		return newPolicies, nil
	}

	patchID := "batch_req"
	typ := datadogV2.CLOUDWORKLOADSECURITYAGENTPOLICYBATCHUPDATEDATATYPE_POLICIES
	attrs := datadogV2.NewCloudWorkloadSecurityAgentPolicyBatchUpdateAttributes()
	attrs.SetPolicies(batchItems)
	data := datadogV2.NewCloudWorkloadSecurityAgentPolicyBatchUpdateData(*attrs, patchID, typ)
	batchReq := datadogV2.NewCloudWorkloadSecurityAgentPolicyBatchUpdateRequest(*data)

	batchResp, _, err := r.api.BatchUpdateCSMThreatsAgentPolicy(r.auth, *batchReq)
	if err != nil {
		*diags = append(*diags, utils.FrameworkErrorDiag(err, "error applying batch policy changes"))
		return nil, err
	}

	for _, policy := range toDelete {
		delete(newPoliciesMap, policy.PolicyLabel.ValueString())
	}

	// get the policies from the response using the ID for modified policies and the name for new policies (because new policies don't have an ID yet)
	respMapByID := make(map[string]datadogV2.CloudWorkloadSecurityAgentPolicyAttributes)
	respMapByName := make(map[string]datadogV2.CloudWorkloadSecurityAgentPolicyAttributes)

	for _, policy := range batchResp.GetData() {
		respID := policy.GetId()
		respAttr := policy.Attributes
		if respAttr == nil {
			continue
		}
		respMapByID[respID] = *respAttr
		respMapByName[respAttr.GetName()] = *respAttr

	}

	// final state of the policies updated with the response from the API
	finalMap := make(map[string]csmThreatsPolicyEntryModel, len(newPoliciesMap))

	for label, policy := range newPoliciesMap {
		oldID := policy.ID.ValueString()
		oldName := policy.Name.ValueString()

		// if the ID is not empty, it means the policy was either modified or left unchanged
		if oldID != "" {
			if attr, found := respMapByID[oldID]; found {
				tags, _ := types.SetValueFrom(ctx, types.StringType, attr.GetHostTags())
				finalMap[label] = csmThreatsPolicyEntryModel{
					PolicyLabel: policy.PolicyLabel,
					ID:          types.StringValue(oldID),
					Name:        types.StringValue(attr.GetName()),
					Description: types.StringValue(attr.GetDescription()),
					Enabled:     types.BoolValue(attr.GetEnabled()),
					Tags:        tags,
				}
				continue
			}
		}

		// if the ID is empty, it means the policy was created
		if attr, found := respMapByName[oldName]; found {
			finalID := findIDByName(oldName, batchResp.GetData())
			tags, _ := types.SetValueFrom(ctx, types.StringType, attr.GetHostTags())
			finalMap[label] = csmThreatsPolicyEntryModel{
				PolicyLabel: policy.PolicyLabel,
				ID:          types.StringValue(finalID),
				Name:        types.StringValue(attr.GetName()),
				Description: types.StringValue(attr.GetDescription()),
				Enabled:     types.BoolValue(attr.GetEnabled()),
				Tags:        tags,
			}
		}
	}

	finalSlice := make([]csmThreatsPolicyEntryModel, 0, len(finalMap))
	for _, policy := range newPolicies {
		if updated, ok := finalMap[policy.PolicyLabel.ValueString()]; ok {
			finalSlice = append(finalSlice, updated)
		}
	}

	return finalSlice, nil
}

func findIDByName(name string, items []datadogV2.CloudWorkloadSecurityAgentPolicyData) string {
	for _, it := range items {
		if it.Attributes != nil && it.Attributes.GetName() == name {
			return it.GetId()
		}
	}
	return ""
}
