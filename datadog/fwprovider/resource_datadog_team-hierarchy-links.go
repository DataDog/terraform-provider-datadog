package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	ID   types.String `tfsdk:"id"`
	Data *dataModel   `tfsdk:"data"`
}

type dataModel struct {
	Type          types.String        `tfsdk:"type"`
	Relationships *relationshipsModel `tfsdk:"relationships"`
}
type relationshipsModel struct {
	ParentTeam *parentTeamModel `tfsdk:"parent_team"`
	SubTeam    *subTeamModel    `tfsdk:"sub_team"`
}
type parentTeamModel struct {
	Data *dataModel `tfsdk:"data"`
}
type dataModel struct {
	Id   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}
type subTeamModel struct {
	Data *dataModel `tfsdk:"data"`
}
type dataModel struct {
	Id   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
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
		},
		Blocks: map[string]schema.Block{
			"data": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Optional:    true,
						Description: "Team hierarchy link type",
					},
				},
				Blocks: map[string]schema.Block{
					"relationships": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{},
						Blocks: map[string]schema.Block{
							"parent_team": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{},
								Blocks: map[string]schema.Block{
									"data": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Optional:    true,
												Description: "The team's identifier",
											},
											"type": schema.StringAttribute{
												Optional:    true,
												Description: "Team type",
											},
										},
									},
								},
							},
							"sub_team": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{},
								Blocks: map[string]schema.Block{
									"data": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Optional:    true,
												Description: "The team's identifier",
											},
											"type": schema.StringAttribute{
												Optional:    true,
												Description: "Team type",
											},
										},
									},
								},
							},
						},
					},
				},
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
	response.Diagnostics.AddError("Update not supported for this resource", "Hierarchy links should be updated by deleting the old link and creating a new one.")
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
	state.ID = types.StringValue(resp.GetLinkId())

	if createdAt, ok := resp.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if provisionedBy, ok := resp.GetProvisionedByOk(); ok {
		state.ProvisionedBy = types.StringValue(*provisionedBy)
	}
}

func (r *teamHierarchyLinksResource) buildTeamHierarchyLinksRequestBody(ctx context.Context, state *teamHierarchyLinksModel) (*datadogV2.TeamHierarchyLinkCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.TeamHierarchyLinkCreateRequest{}

	if state.Data != nil {
		var data datadogV2.TeamHierarchyLinkCreate

		data.SetType(datadogV2.TeamHierarchyLinkType(state.Data.Type.ValueString()))

		var relationships datadogV2.TeamHierarchyLinkCreateRelationships

		if state.Data.Relationships.ParentTeam != nil {
			var parentTeam datadogV2.TeamHierarchyLinkCreateTeamRelationship

			if state.Data.Relationships.ParentTeam.Data != nil {
				var data datadogV2.TeamHierarchyLinkCreateTeam

				if !state.Data.Relationships.ParentTeam.Data.Id.IsNull() {
					data.SetId(state.Data.Relationships.ParentTeam.Data.Id.ValueString())
				}
				if !state.Data.Relationships.ParentTeam.Data.Type.IsNull() {
					data.SetType(datadogV2.TeamType(state.Data.Relationships.ParentTeam.Data.Type.ValueString()))
				}
				parentTeam.Data = &data
			}
			relationships.ParentTeam = &parentTeam
		}

		if state.Data.Relationships.SubTeam != nil {
			var subTeam datadogV2.TeamHierarchyLinkCreateTeamRelationship

			if state.Data.Relationships.SubTeam.Data != nil {
				var data datadogV2.TeamHierarchyLinkCreateTeam

				if !state.Data.Relationships.SubTeam.Data.Id.IsNull() {
					data.SetId(state.Data.Relationships.SubTeam.Data.Id.ValueString())
				}
				if !state.Data.Relationships.SubTeam.Data.Type.IsNull() {
					data.SetType(datadogV2.TeamType(state.Data.Relationships.SubTeam.Data.Type.ValueString()))
				}
				subTeam.Data = &data
			}
			relationships.SubTeam = &subTeam
		}
		data.Relationships = relationships
		req.Data = &data
	}

	return req, diags
}
