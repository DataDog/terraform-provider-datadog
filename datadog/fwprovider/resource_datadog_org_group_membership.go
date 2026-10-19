package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var uuidValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
	"must be a valid UUID",
)

var (
	_ resource.ResourceWithConfigure   = &OrgGroupMembershipResource{}
	_ resource.ResourceWithImportState = &OrgGroupMembershipResource{}
)

type OrgGroupMembershipResource struct {
	API  *datadogV2.OrgGroupsApi
	Auth context.Context
}

type OrgGroupMembershipModel struct {
	ID         types.String `tfsdk:"id"`
	OrgGroupID types.String `tfsdk:"org_group_id"`
	OrgUuid    types.String `tfsdk:"org_uuid"`
	OrgSite    types.String `tfsdk:"org_site"`
	OrgName    types.String `tfsdk:"org_name"`
}

func NewOrgGroupMembershipResource() resource.Resource {
	return &OrgGroupMembershipResource{}
}

func (r *OrgGroupMembershipResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetOrgGroupsApiV2()
	r.Auth = providerData.Auth
}

func (r *OrgGroupMembershipResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "org_group_membership"
}

func (r *OrgGroupMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Org Group Membership resource. This can be used to manage an organization's membership in an org group.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"org_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the org group to assign the organization to.",
				Validators:  []validator.String{uuidValidator},
			},
			"org_uuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the organization.",
				Validators:  []validator.String{uuidValidator},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_site": schema.StringAttribute{
				Computed:    true,
				Description: "The site of the organization. Server-managed (derived from the organization's own settings).",
			},
			"org_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the organization.",
			},
		},
	}
}

func (r *OrgGroupMembershipResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *OrgGroupMembershipResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state OrgGroupMembershipModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	orgUuid, err := uuid.Parse(state.OrgUuid.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_uuid must be a valid UUID"))
		return
	}

	params := datadogV2.NewListOrgGroupMembershipsOptionalParameters().WithFilterOrgUuid(orgUuid)
	listResp, _, err := r.API.ListOrgGroupMemberships(r.Auth, *params)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing org group memberships"))
		return
	}

	memberships := listResp.GetData()
	if len(memberships) == 0 {
		response.Diagnostics.AddError(
			"datadog_org_group_membership: no membership found",
			fmt.Sprintf("no membership returned for org %s", state.OrgUuid.ValueString()),
		)
		return
	}
	membershipID := memberships[0].GetId()

	targetGroupID, err := uuid.Parse(state.OrgGroupID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
		return
	}

	body := r.buildUpdateRequest(membershipID, targetGroupID)

	resp, _, err := r.API.UpdateOrgGroupMembership(r.Auth, membershipID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating org group membership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_membership: response contains unparsedObject", err.Error())
		return
	}

	diags := r.updateState(&state, &resp)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupMembershipResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state OrgGroupMembershipModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "membership ID must be a valid UUID"))
		return
	}

	resp, httpResp, err := r.API.GetOrgGroupMembership(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving org group membership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_membership: response contains unparsedObject", err.Error())
		return
	}

	diags := r.updateState(&state, &resp)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupMembershipResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state OrgGroupMembershipModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	membershipID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "membership ID must be a valid UUID"))
		return
	}

	targetGroupID, err := uuid.Parse(state.OrgGroupID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "org_group_id must be a valid UUID"))
		return
	}

	body := r.buildUpdateRequest(membershipID, targetGroupID)

	resp, _, err := r.API.UpdateOrgGroupMembership(r.Auth, membershipID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating org group membership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_org_group_membership: response contains unparsedObject", err.Error())
		return
	}

	diags := r.updateState(&state, &resp)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *OrgGroupMembershipResource) Delete(_ context.Context, _ resource.DeleteRequest, response *resource.DeleteResponse) {
	// No API call — memberships cannot be deleted, only reassigned.
	// The org remains in its current group. Emit a warning so users who ran
	// `terraform destroy` understand the org isn't detached from the group.
	response.Diagnostics.AddWarning(
		"datadog_org_group_membership destroy is state-only",
		"Memberships cannot be deleted via the API; only reassigned. The organization remains in its current org group. "+
			"To destroy a `datadog_org_group`, the group must have zero memberships pointing at it — "+
			"move member orgs to another group first (e.g. by updating `org_group_id` on each membership resource).",
	)
}

func (r *OrgGroupMembershipResource) buildUpdateRequest(membershipID uuid.UUID, targetGroupID uuid.UUID) *datadogV2.OrgGroupMembershipUpdateRequest {
	orgGroupRef := datadogV2.NewOrgGroupRelationshipToOneData(targetGroupID, datadogV2.ORGGROUPTYPE_ORG_GROUPS)
	relationship := datadogV2.NewOrgGroupRelationshipToOne(*orgGroupRef)
	relationships := datadogV2.NewOrgGroupMembershipUpdateRelationships(*relationship)
	data := datadogV2.NewOrgGroupMembershipUpdateData(membershipID, *relationships, datadogV2.ORGGROUPMEMBERSHIPTYPE_ORG_GROUP_MEMBERSHIPS)
	return datadogV2.NewOrgGroupMembershipUpdateRequest(*data)
}

func (r *OrgGroupMembershipResource) updateState(state *OrgGroupMembershipModel, resp *datadogV2.OrgGroupMembershipResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId().String())

	attributes := data.GetAttributes()
	state.OrgUuid = types.StringValue(attributes.GetOrgUuid().String())
	state.OrgSite = types.StringValue(attributes.GetOrgSite())
	state.OrgName = types.StringValue(attributes.GetOrgName())

	rels, ok := data.GetRelationshipsOk()
	if !ok || rels == nil {
		diags.AddError("missing relationships", "org group membership response does not contain relationships")
		return diags
	}
	orgGroup, ok := rels.GetOrgGroupOk()
	if !ok || orgGroup == nil {
		diags.AddError("missing relationships", "org group membership response does not contain org_group relationship")
		return diags
	}
	orgGroupData, ok := orgGroup.GetDataOk()
	if !ok {
		diags.AddError("missing relationships", "org group membership response does not contain org_group.data")
		return diags
	}
	state.OrgGroupID = types.StringValue(orgGroupData.GetId().String())

	return diags
}
