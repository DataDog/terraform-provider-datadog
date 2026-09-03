package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &OrgGroupPolicyResource{}
	_ resource.ResourceWithImportState    = &OrgGroupPolicyResource{}
	_ resource.ResourceWithValidateConfig = &OrgGroupPolicyResource{}
)

type OrgGroupPolicyResource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupPolicyModel struct {
	ID              types.String         `tfsdk:"id"`
	OrgGroupID      types.String         `tfsdk:"org_group_id"`
	PolicyName      types.String         `tfsdk:"policy_name"`
	Content         jsontypes.Normalized `tfsdk:"content"`
	EnforcementTier types.String         `tfsdk:"enforcement_tier"`
	PolicyType      types.String         `tfsdk:"policy_type"`
}

func NewOrgGroupPolicyResource() resource.Resource {
	return &OrgGroupPolicyResource{}
}

func (r *OrgGroupPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	r.Auth = providerData.Auth
}

func (r *OrgGroupPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "org_group_policy"
}

func (r *OrgGroupPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Org Group Policy resource. This can be used to create and manage policies attached to an org group.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group this policy belongs to.",
				Validators:  []validator.String{uuidValidator},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the policy. This becomes the name of the resource created across orgs in the group (for example, for `role` policies, the name of the created role). Can be renamed in place for `role` policies; renaming an `org_config` policy replaces the resource.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var policyType types.String
							diags := req.State.GetAttribute(ctx, frameworkPath.Root("policy_type"), &policyType)
							resp.Diagnostics.Append(diags...)
							if diags.HasError() {
								return
							}
							resp.RequiresReplace = policyType.ValueString() != string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE)
						},
						"Renaming requires replacing the resource unless policy_type is \"role\"; the API only supports in-place rename for role policies.",
						"Renaming requires replacing the resource unless `policy_type` is `role`; the API only supports in-place rename for `role` policies.",
					),
				},
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "The policy content as a JSON-encoded string.",
				CustomType:  jsontypes.NormalizedType{},
			},
			"enforcement_tier": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The enforcement tier of the policy. `OVERRIDE_ALLOWED` means the policy is set but member orgs may mutate it. `GROUP_MANAGED` means the policy is strictly controlled and mutations are blocked for affected orgs. `DELEGATE` means each member org controls its own value. Not all policy types support every tier.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_OVERRIDE_ALLOWED),
						string(datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_GROUP_MANAGED),
						string(datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_DELEGATE),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The type of the policy.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ORG_CONFIG),
						string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *OrgGroupPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

// ValidateConfig catches role + OVERRIDE_ALLOWED at plan time; the API only 400s.
func (r *OrgGroupPolicyResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config OrgGroupPolicyModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	if config.PolicyType.ValueString() != string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE) {
		return
	}
	if config.EnforcementTier.ValueString() == string(datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_OVERRIDE_ALLOWED) {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("enforcement_tier"),
			"Invalid enforcement_tier for policy_type \"role\"",
			"role policies only support GROUP_MANAGED and DELEGATE.",
		)
	}
}

func (r *OrgGroupPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state OrgGroupPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	orgGroupID, err := uuid.Parse(state.OrgGroupID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
		return
	}

	var content map[string]interface{}
	if err := json.Unmarshal([]byte(state.Content.ValueString()), &content); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "content must be valid JSON"))
		return
	}

	attributes := datadogV2.NewOrgGroupPolicyCreateAttributes(content, state.PolicyName.ValueString())
	if !state.EnforcementTier.IsNull() && !state.EnforcementTier.IsUnknown() {
		tier := datadogV2.OrgGroupPolicyEnforcementTier(state.EnforcementTier.ValueString())
		attributes.SetEnforcementTier(tier)
	}
	if !state.PolicyType.IsNull() && !state.PolicyType.IsUnknown() {
		policyType := datadogV2.OrgGroupPolicyPolicyType(state.PolicyType.ValueString())
		attributes.SetPolicyType(policyType)
	}

	orgGroupRefData := datadogV2.NewOrgGroupRelationshipToOneData(orgGroupID, datadogV2.ORGGROUPTYPE_ORG_GROUPS)
	orgGroupRef := datadogV2.NewOrgGroupRelationshipToOne(*orgGroupRefData)
	relationships := datadogV2.NewOrgGroupPolicyCreateRelationships(*orgGroupRef)

	data := datadogV2.NewOrgGroupPolicyCreateData(*attributes, *relationships, datadogV2.ORGGROUPPOLICYTYPE_ORG_GROUP_POLICIES)
	body := datadogV2.NewOrgGroupPolicyCreateRequest(*data)

	resp, _, err := r.API.CreateOrgGroupPolicy(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating org group policy"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_policy: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from org group policy response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state OrgGroupPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group policy ID must be a valid UUID"))
		return
	}

	resp, httpResp, err := r.API.GetOrgGroupPolicy(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving org group policy"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_policy: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from org group policy response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state OrgGroupPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	var priorState OrgGroupPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &priorState)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group policy ID must be a valid UUID"))
		return
	}

	var content map[string]interface{}
	if err := json.Unmarshal([]byte(state.Content.ValueString()), &content); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "content must be valid JSON"))
		return
	}

	attributes := datadogV2.NewOrgGroupPolicyUpdateAttributes()
	attributes.SetContent(content)
	if state.PolicyName.ValueString() != priorState.PolicyName.ValueString() {
		attributes.SetPolicyName(state.PolicyName.ValueString())
	}
	if !state.EnforcementTier.IsNull() && !state.EnforcementTier.IsUnknown() {
		tier := datadogV2.OrgGroupPolicyEnforcementTier(state.EnforcementTier.ValueString())
		attributes.SetEnforcementTier(tier)
	}

	data := datadogV2.NewOrgGroupPolicyUpdateData(*attributes, id, datadogV2.ORGGROUPPOLICYTYPE_ORG_GROUP_POLICIES)
	body := datadogV2.NewOrgGroupPolicyUpdateRequest(*data)

	resp, _, err := r.API.UpdateOrgGroupPolicy(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating org group policy"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_policy: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from org group policy response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state OrgGroupPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group policy ID must be a valid UUID"))
		return
	}

	// role policies are never hard-deleted server-side (API returns 405). If it's already
	// disabled (enforcement_tier = DELEGATE), there's nothing left for the API to do; just
	// drop it from state. Otherwise reject and tell the user to disable it first.
	if state.PolicyType.ValueString() == string(datadogV2.ORGGROUPPOLICYPOLICYTYPE_ROLE) {
		if state.EnforcementTier.ValueString() != string(datadogV2.ORGGROUPPOLICYENFORCEMENTTIER_DELEGATE) {
			response.Diagnostics.AddError(
				"role policies cannot be deleted",
				`Set enforcement_tier = "DELEGATE" to disable this role policy before removing it from your configuration.`,
			)
		}
		return
	}

	httpResp, err := r.API.DeleteOrgGroupPolicy(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting org group policy"))
	}
}

func (r *OrgGroupPolicyResource) updateState(state *OrgGroupPolicyModel, resp *datadogV2.OrgGroupPolicyResponse) error {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId().String())

	attributes := data.GetAttributes()
	state.PolicyName = types.StringValue(attributes.GetPolicyName())
	state.EnforcementTier = types.StringValue(string(attributes.GetEnforcementTier()))
	// policy_type is immutable (RequiresReplace). If the response omits it, keep the
	// prior state value so we don't force a spurious replace on the next plan.
	if pt := string(attributes.GetPolicyType()); pt != "" {
		state.PolicyType = types.StringValue(pt)
	}

	contentBytes, err := json.Marshal(attributes.GetContent())
	if err != nil {
		return fmt.Errorf("error marshaling policy content: %w", err)
	}
	state.Content = jsontypes.NewNormalizedValue(string(contentBytes))

	rels, ok := data.GetRelationshipsOk()
	if !ok || rels == nil {
		return fmt.Errorf("org group policy response missing relationships")
	}
	orgGroup, ok := rels.GetOrgGroupOk()
	if !ok || orgGroup == nil {
		return fmt.Errorf("org group policy response missing org_group relationship")
	}
	orgGroupData, ok := orgGroup.GetDataOk()
	if !ok {
		return fmt.Errorf("org group policy response missing org_group.data")
	}
	state.OrgGroupID = types.StringValue(orgGroupData.GetId().String())

	return nil
}
