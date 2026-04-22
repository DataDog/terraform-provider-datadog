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
	_ resource.ResourceWithConfigure   = &OrgGroupPolicyOverrideResource{}
	_ resource.ResourceWithImportState = &OrgGroupPolicyOverrideResource{}
)

type OrgGroupPolicyOverrideResource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupPolicyOverrideModel struct {
	ID         types.String         `tfsdk:"id"`
	OrgGroupID types.String         `tfsdk:"org_group_id"`
	PolicyID   types.String         `tfsdk:"policy_id"`
	OrgUuid    types.String         `tfsdk:"org_uuid"`
	OrgSite    types.String         `tfsdk:"org_site"`
	Content    jsontypes.Normalized `tfsdk:"content"`
}

func NewOrgGroupPolicyOverrideResource() resource.Resource {
	return &OrgGroupPolicyOverrideResource{}
}

func (r *OrgGroupPolicyOverrideResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	r.Auth = providerData.Auth
}

func (r *OrgGroupPolicyOverrideResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "org_group_policy_override"
}

func (r *OrgGroupPolicyOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Org Group Policy Override resource. An override exempts a specific organization from a policy applied at the org group level.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group that owns the policy.",
				Validators:  []validator.String{uuidValidator},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group policy the override applies to.",
				Validators:  []validator.String{uuidValidator},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_uuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the organization being exempted from the policy.",
				Validators:  []validator.String{uuidValidator},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_site": schema.StringAttribute{
				Required:    true,
				Description: "The short site name of the organization (e.g. `us1`, `eu1`, `us1-fed`). Part of the override's server-side identity; changing it replaces the resource.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "The org's config value at the time the override was created, as a JSON-encoded string. Server-managed.",
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *OrgGroupPolicyOverrideResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *OrgGroupPolicyOverrideResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state OrgGroupPolicyOverrideModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	orgGroupID, err := uuid.Parse(state.OrgGroupID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
		return
	}
	policyID, err := uuid.Parse(state.PolicyID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "policy_id must be a valid UUID"))
		return
	}
	orgUUID, err := uuid.Parse(state.OrgUuid.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_uuid must be a valid UUID"))
		return
	}

	orgGroupRefData := datadogV2.NewOrgGroupRelationshipToOneData(orgGroupID, datadogV2.ORGGROUPTYPE_ORG_GROUPS)
	orgGroupRef := datadogV2.NewOrgGroupRelationshipToOne(*orgGroupRefData)
	policyRefData := datadogV2.NewOrgGroupPolicyRelationshipToOneData(policyID, datadogV2.ORGGROUPPOLICYTYPE_ORG_GROUP_POLICIES)
	policyRef := datadogV2.NewOrgGroupPolicyRelationshipToOne(*policyRefData)
	relationships := datadogV2.NewOrgGroupPolicyOverrideCreateRelationships(*orgGroupRef, *policyRef)

	attributes := datadogV2.NewOrgGroupPolicyOverrideCreateAttributes(state.OrgSite.ValueString(), orgUUID)
	data := datadogV2.NewOrgGroupPolicyOverrideCreateData(*attributes, *relationships, datadogV2.ORGGROUPPOLICYOVERRIDETYPE_ORG_GROUP_POLICY_OVERRIDES)
	body := datadogV2.NewOrgGroupPolicyOverrideCreateRequest(*data)

	resp, _, err := r.API.CreateOrgGroupPolicyOverride(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating org group policy override"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_policy_override: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from org group policy override response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupPolicyOverrideResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state OrgGroupPolicyOverrideModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group policy override ID must be a valid UUID"))
		return
	}

	resp, httpResp, err := r.API.GetOrgGroupPolicyOverride(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving org group policy override"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_policy_override: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from org group policy override response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupPolicyOverrideResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	// All user-settable fields (org_group_id, policy_id, org_uuid, org_site) are
	// RequiresReplace; `content` is Computed-only. Terraform should never invoke
	// Update. If a settable, non-Replace field is added in the future, implement
	// the real Update logic here.
	response.Diagnostics.AddError(
		"datadog_org_group_policy_override: unexpected Update call",
		"all fields should be RequiresReplace or Computed; Update is unreachable. This indicates a provider bug.",
	)
}

func (r *OrgGroupPolicyOverrideResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state OrgGroupPolicyOverrideModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org group policy override ID must be a valid UUID"))
		return
	}

	httpResp, err := r.API.DeleteOrgGroupPolicyOverride(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting org group policy override"))
	}
}

func (r *OrgGroupPolicyOverrideResource) updateState(state *OrgGroupPolicyOverrideModel, resp *datadogV2.OrgGroupPolicyOverrideResponse) error {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId().String())

	attributes := data.GetAttributes()
	state.OrgUuid = types.StringValue(attributes.GetOrgUuid().String())
	state.OrgSite = types.StringValue(attributes.GetOrgSite())

	if attributes.HasContent() {
		bytes, err := json.Marshal(attributes.GetContent())
		if err != nil {
			return fmt.Errorf("error marshaling override content: %w", err)
		}
		state.Content = jsontypes.NewNormalizedValue(string(bytes))
	} else {
		state.Content = jsontypes.NewNormalizedValue("{}")
	}

	rels, ok := data.GetRelationshipsOk()
	if !ok || rels == nil {
		return fmt.Errorf("org group policy override response missing relationships")
	}
	orgGroup, ok := rels.GetOrgGroupOk()
	if !ok || orgGroup == nil {
		return fmt.Errorf("org group policy override response missing org_group relationship")
	}
	orgGroupData, ok := orgGroup.GetDataOk()
	if !ok {
		return fmt.Errorf("org group policy override response missing org_group.data")
	}
	state.OrgGroupID = types.StringValue(orgGroupData.GetId().String())

	policy, ok := rels.GetOrgGroupPolicyOk()
	if !ok || policy == nil {
		return fmt.Errorf("org group policy override response missing org_group_policy relationship")
	}
	policyData, ok := policy.GetDataOk()
	if !ok {
		return fmt.Errorf("org group policy override response missing org_group_policy.data")
	}
	state.PolicyID = types.StringValue(policyData.GetId().String())

	return nil
}
