package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &teamHierarchyLinksResource{}
	_ resource.ResourceWithImportState = &teamHierarchyLinksResource{}
)

type teamHierarchyLinksResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamHierarchyLinksModel struct {
	ID            types.String `tfsdk:"id"`
	ParentTeamId  types.String `tfsdk:"parent_team_id"`
	SubTeamId     types.String `tfsdk:"sub_team_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ProvisionedBy types.String `tfsdk:"provisioned_by"`
}

func NewTeamHierarchyLinksResource() resource.Resource {
	return &teamHierarchyLinksResource{}
}

func (r *teamHierarchyLinksResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamHierarchyLinksResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_hierarchy_links"
}

func (r *teamHierarchyLinksResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog TeamHierarchyLinks resource. This can be used to create and manage Datadog team-hierarchy-links.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"parent_team_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the parent team the team hierarchy link is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sub_team_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the sub team the team hierarchy link is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the team hierarchy link was created.",
			},
			"provisioned_by": schema.StringAttribute{
				Computed:    true,
				Description: "The user who created the team hierarchy link.",
			},
		},
	}
}

func (r *teamHierarchyLinksResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *teamHierarchyLinksResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamHierarchyLinksModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetTeamHierarchyLink(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamHierarchyLinks"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamHierarchyLinksResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamHierarchyLinksModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildTeamHierarchyLinksRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.AddTeamHierarchyLink(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamHierarchyLinks"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamHierarchyLinksResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported for this resource", "Hierarchy links are immutable.")
}

func (r *teamHierarchyLinksResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamHierarchyLinksModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.RemoveTeamHierarchyLink(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting team-hierarchy-links"))
		return
	}
}

func (r *teamHierarchyLinksResource) updateState(ctx context.Context, state *teamHierarchyLinksModel, resp *datadogV2.TeamHierarchyLinkResponse) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId())

	attributes := data.GetAttributes()

	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if provisionedBy, ok := attributes.GetProvisionedByOk(); ok {
		state.ProvisionedBy = types.StringValue(*provisionedBy)
	}
}

func (r *teamHierarchyLinksResource) buildTeamHierarchyLinksRequestBody(ctx context.Context, state *teamHierarchyLinksModel) (*datadogV2.TeamHierarchyLinkCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.TeamHierarchyLinkCreateRequest{}

	var data datadogV2.TeamHierarchyLinkCreate

	data.SetType(datadogV2.TEAMHIERARCHYLINKTYPE_TEAM_HIERARCHY_LINKS)

	var relationships datadogV2.TeamHierarchyLinkCreateRelationships
	var parentTeam datadogV2.TeamHierarchyLinkCreateTeamRelationship
	var parentTeamData datadogV2.TeamHierarchyLinkCreateTeam
	parentTeamData.SetType(datadogV2.TEAMTYPE_TEAM)
	parentTeamData.SetId(state.ParentTeamId.ValueString())

	parentTeam.Data = parentTeamData

	relationships.ParentTeam = parentTeam

	var subTeam datadogV2.TeamHierarchyLinkCreateTeamRelationship
	var subTeamData datadogV2.TeamHierarchyLinkCreateTeam
	subTeamData.SetType(datadogV2.TEAMTYPE_TEAM)
	subTeamData.SetId(state.SubTeamId.ValueString())

	subTeam.Data = subTeamData

	relationships.SubTeam = subTeam

	data.Relationships = relationships

	req.Data = data

	return req, diags
}
