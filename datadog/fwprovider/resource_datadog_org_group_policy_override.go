package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
		Description: "Provides a Datadog Org Group Policy Override resource. An override exempts a specific organization from a policy applied at the org group level.\n\n" +
			"**Server-side auto-creation.** Overrides are also created automatically by the server when an org's existing config differs from a policy's value. Two triggers: (1) when an org is moved into an org group, the server compares that org's current config to each non-`ENFORCE` policy in the group and creates an override for any mismatch; (2) when a policy is created or updated with a non-`ENFORCE` tier, the server performs the same comparison against every member org. Auto-created and user-declared overrides are indistinguishable at the API level and can be adopted into Terraform via `terraform import` or by iterating the `datadog_org_group_policy_overrides` data source with `for_each` + `import` blocks.\n\n" +
			"**Delete behavior.** Removing an override does **not** just remove the exemption marker — it resets the target org's config value to match the parent policy's current value. The server treats override deletion as \"re-apply the policy to this org\" and propagates the policy value accordingly. To stop managing an override without changing the org's value, use `terraform state rm` instead of removing the resource block.\n\n" +
			"**ENFORCE tier cascade.** When the parent `datadog_org_group_policy` transitions to `enforcement_tier = ENFORCE`, the server atomically deletes every override for that policy. Creating an override against an already-`ENFORCE`-tier policy also fails with a `FailedPrecondition` error. If you plan to flip a policy to `ENFORCE`, remove the `datadog_org_group_policy_override` resource blocks for that policy in the same commit.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group that owns the policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group policy the override applies to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_uuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the organization being exempted from the policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_site": schema.StringAttribute{
				Required:    true,
				Description: "The short site name of the organization (e.g. `us1`, `eu1`, `us1-fed`).",
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
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
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
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from org group policy override response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupPolicyOverrideResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All fields are RequiresReplace; Update is unreachable.
}

func (r *OrgGroupPolicyOverrideResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state OrgGroupPolicyOverrideModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
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

	if rels, ok := data.GetRelationshipsOk(); ok && rels != nil {
		if orgGroup, ok := rels.GetOrgGroupOk(); ok && orgGroup != nil {
			orgGroupData := orgGroup.GetData()
			state.OrgGroupID = types.StringValue(orgGroupData.GetId().String())
		}
		if policy, ok := rels.GetOrgGroupPolicyOk(); ok && policy != nil {
			policyData := policy.GetData()
			state.PolicyID = types.StringValue(policyData.GetId().String())
		}
	}

	return nil
}
