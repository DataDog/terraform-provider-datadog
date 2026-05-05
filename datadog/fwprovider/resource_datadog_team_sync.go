package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &teamSyncResource{}
	_ resource.ResourceWithImportState = &teamSyncResource{}
)

type teamSyncResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamSyncModel struct {
	ID             types.String                  `tfsdk:"id"`
	Source         types.String                  `tfsdk:"source"`
	Type           types.String                  `tfsdk:"type"`
	Frequency      types.String                  `tfsdk:"frequency"`
	SyncMembership types.Bool                    `tfsdk:"sync_membership"`
	SelectionState []teamSyncSelectionStateModel `tfsdk:"selection_state"`
}

type teamSyncSelectionStateModel struct {
	Operation  types.String             `tfsdk:"operation"`
	Scope      types.String             `tfsdk:"scope"`
	ExternalId *teamSyncExternalIdModel `tfsdk:"external_id"`
}

type teamSyncExternalIdModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func NewTeamSyncResource() resource.Resource {
	return &teamSyncResource{}
}

func (r *teamSyncResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamSyncResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_sync"
}

func (r *teamSyncResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Team Sync resource. This can be used to configure team synchronization from external sources (e.g. GitHub) into Datadog.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"source": schema.StringAttribute{
				Required:    true,
				Description: "The external source platform for team synchronization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("github"),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of synchronization operation. `link` connects teams by matching names. `provision` creates new teams when no match is found.",
				Validators: []validator.String{
					stringvalidator.OneOf("link", "provision"),
				},
			},
			"frequency": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("once"),
				Description: "How often the sync process should run.",
				Validators: []validator.String{
					stringvalidator.OneOf("once", "continuously", "paused"),
				},
			},
			"sync_membership": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to sync members from the external team to the Datadog team.",
			},
		},
		Blocks: map[string]schema.Block{
			"selection_state": schema.ListNestedBlock{
				Description: "Specifies which teams or organizations to sync. When provided, synchronization is limited to the specified items and their subtrees.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"operation": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("include"),
							Description: "The operation to perform on the selected hierarchy.",
							Validators: []validator.String{
								stringvalidator.OneOf("include"),
							},
						},
						"scope": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("subtree"),
							Description: "The scope of the selection.",
							Validators: []validator.String{
								stringvalidator.OneOf("subtree"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"external_id": schema.SingleNestedBlock{
							Description: "The external identifier for a team or organization in the source platform.",
							Validators: []validator.Object{
								objectvalidator.IsRequired(),
							},
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The type of external identifier.",
									Validators: []validator.String{
										stringvalidator.OneOf("team", "organization"),
									},
								},
								"value": schema.StringAttribute{
									Required:    true,
									Description: "The external identifier value from the source platform (e.g. a GitHub organization ID or team ID).",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *teamSyncResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	source := request.ID
	if source != "github" {
		response.Diagnostics.AddError("invalid import ID", "Import ID must be the source name (e.g. \"github\").")
		return
	}

	resp, _, err := r.Api.GetTeamSync(r.Auth, datadogV2.TeamSyncAttributesSource(source))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamSync for import"))
		return
	}

	data := resp.GetData()
	if len(data) == 0 {
		response.Diagnostics.AddError("not found", "no team sync configuration found for source \""+source+"\"")
		return
	}

	var state teamSyncModel
	r.updateState(&state, &data[0])
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamSyncResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamSyncModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	source := datadogV2.TeamSyncAttributesSource(state.Source.ValueString())
	resp, httpResp, err := r.Api.GetTeamSync(r.Auth, source)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamSync"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := resp.GetData()
	if len(data) == 0 {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(&state, &data[0])
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamSyncResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamSyncModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Guard: check if sync config already exists for this source
	source := datadogV2.TeamSyncAttributesSource(state.Source.ValueString())
	existResp, httpResp, err := r.Api.GetTeamSync(r.Auth, source)
	if err == nil {
		data := existResp.GetData()
		if len(data) > 0 {
			attrs := data[0].GetAttributes()
			freq := attrs.GetFrequency()
			if freq != datadogV2.TEAMSYNCATTRIBUTESFREQUENCY_PAUSED {
				response.Diagnostics.AddError(
					"team sync already exists",
					"A team sync configuration for source \""+state.Source.ValueString()+"\" already exists. "+
						"Only one datadog_team_sync resource per source is allowed. "+
						"Import the existing resource with: terraform import datadog_team_sync.<name> "+state.Source.ValueString(),
				)
				return
			}
		}
	} else if httpResp != nil && httpResp.StatusCode != 404 {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error checking existing TeamSync"))
		return
	}

	r.syncAndRead(ctx, &state, &response.Diagnostics, "creating")
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamSyncResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state teamSyncModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	r.syncAndRead(ctx, &state, &response.Diagnostics, "updating")
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamSyncResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamSyncModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Delete by setting frequency to paused
	source := datadogV2.TeamSyncAttributesSource(state.Source.ValueString())
	syncType := datadogV2.TeamSyncAttributesType(state.Type.ValueString())
	attrs := datadogV2.NewTeamSyncAttributes(source, syncType)
	attrs.SetFrequency(datadogV2.TEAMSYNCATTRIBUTESFREQUENCY_PAUSED)

	syncData := datadogV2.NewTeamSyncData(*attrs, datadogV2.TEAMSYNCBULKTYPE_TEAM_SYNC_BULK)
	body := datadogV2.NewTeamSyncRequest(*syncData)

	httpResp, err := r.Api.SyncTeams(r.Auth, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting TeamSync"))
		return
	}
}

func (r *teamSyncResource) syncAndRead(_ context.Context, state *teamSyncModel, diags *diag.Diagnostics, action string) {
	body := r.buildRequestBody(state)

	_, err := r.Api.SyncTeams(r.Auth, *body)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error "+action+" TeamSync"))
		return
	}

	// SyncTeams returns no body, so re-read to populate state
	source := datadogV2.TeamSyncAttributesSource(state.Source.ValueString())
	resp, _, err := r.Api.GetTeamSync(r.Auth, source)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error reading TeamSync after "+action))
		return
	}

	data := resp.GetData()
	if len(data) == 0 {
		diags.AddError("empty response", "no team sync data returned after "+action)
		return
	}

	r.updateState(state, &data[0])
}

func (r *teamSyncResource) updateState(state *teamSyncModel, data *datadogV2.TeamSyncData) {
	attrs := data.GetAttributes()

	state.ID = types.StringValue(string(attrs.GetSource()))
	state.Source = types.StringValue(string(attrs.GetSource()))
	state.Type = types.StringValue(string(attrs.GetType()))

	if freq, ok := attrs.GetFrequencyOk(); ok {
		state.Frequency = types.StringValue(string(*freq))
	}

	if sm, ok := attrs.GetSyncMembershipOk(); ok {
		state.SyncMembership = types.BoolValue(*sm)
	}

	if selState, ok := attrs.GetSelectionStateOk(); ok {
		selModels := make([]teamSyncSelectionStateModel, 0, len(*selState))
		for _, item := range *selState {
			m := teamSyncSelectionStateModel{}
			if op, ok := item.GetOperationOk(); ok {
				m.Operation = types.StringValue(string(*op))
			}
			if scope, ok := item.GetScopeOk(); ok {
				m.Scope = types.StringValue(string(*scope))
			}
			extId := item.GetExternalId()
			m.ExternalId = &teamSyncExternalIdModel{
				Type:  types.StringValue(string(extId.GetType())),
				Value: types.StringValue(extId.GetValue()),
			}
			selModels = append(selModels, m)
		}
		state.SelectionState = selModels
	} else {
		state.SelectionState = nil
	}
}

func (r *teamSyncResource) buildRequestBody(state *teamSyncModel) *datadogV2.TeamSyncRequest {
	source := datadogV2.TeamSyncAttributesSource(state.Source.ValueString())
	syncType := datadogV2.TeamSyncAttributesType(state.Type.ValueString())

	attrs := datadogV2.NewTeamSyncAttributes(source, syncType)

	if !state.Frequency.IsNull() && !state.Frequency.IsUnknown() {
		freq := datadogV2.TeamSyncAttributesFrequency(state.Frequency.ValueString())
		attrs.SetFrequency(freq)
	}

	if !state.SyncMembership.IsNull() && !state.SyncMembership.IsUnknown() {
		attrs.SetSyncMembership(state.SyncMembership.ValueBool())
	}

	if state.SelectionState != nil {
		items := make([]datadogV2.TeamSyncSelectionStateItem, 0, len(state.SelectionState))
		for _, sel := range state.SelectionState {
			extIdType := datadogV2.TeamSyncSelectionStateExternalIdType(sel.ExternalId.Type.ValueString())
			extId := datadogV2.NewTeamSyncSelectionStateExternalId(extIdType, sel.ExternalId.Value.ValueString())
			item := datadogV2.NewTeamSyncSelectionStateItem(*extId)

			if !sel.Operation.IsNull() && !sel.Operation.IsUnknown() {
				op := datadogV2.TeamSyncSelectionStateOperation(sel.Operation.ValueString())
				item.SetOperation(op)
			}
			if !sel.Scope.IsNull() && !sel.Scope.IsUnknown() {
				scope := datadogV2.TeamSyncSelectionStateScope(sel.Scope.ValueString())
				item.SetScope(scope)
			}
			items = append(items, *item)
		}
		attrs.SetSelectionState(items)
	}

	syncData := datadogV2.NewTeamSyncData(*attrs, datadogV2.TEAMSYNCBULKTYPE_TEAM_SYNC_BULK)
	return datadogV2.NewTeamSyncRequest(*syncData)
}
