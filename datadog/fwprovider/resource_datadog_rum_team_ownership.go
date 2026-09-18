package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &datadogRumTeamOwnershipResource{}
	_ resource.ResourceWithImportState = &datadogRumTeamOwnershipResource{}
)

type datadogRumTeamOwnershipResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationId types.String `tfsdk:"application_id"`
	MatchType     types.String `tfsdk:"match_type"`
	OrgId         types.Int64  `tfsdk:"org_id"`
	Service       types.String `tfsdk:"service"`
	TeamHandle    types.String `tfsdk:"team_handle"`
	ViewName      types.String `tfsdk:"view_name"`
}

type datadogRumTeamOwnershipResource struct {
	Api  *datadogV2.RumTeamsOwnershipApi
	Auth context.Context
}

func NewDatadogRumTeamOwnershipResource() resource.Resource {
	return &datadogRumTeamOwnershipResource{}
}

func (r *datadogRumTeamOwnershipResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = datadogV2.NewRumTeamsOwnershipApi(providerData.DatadogApiInstances.HttpClient)
	r.Auth = providerData.Auth
}

func (r *datadogRumTeamOwnershipResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_team_ownership"
}

func (r *datadogRumTeamOwnershipResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RUM team ownership mapping resource. If both `application_id` and `service` are omitted, the mapping applies across all applications and services.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"application_id": schema.StringAttribute{
				Description: "The ID of the RUM application this mapping applies to.\nFor browser applications, this is the real application UUID.\nFor mobile applications, this is the nil UUID `00000000-0000-0000-0000-000000000000` (wildcard), meaning the ownership applies across all applications.",
				Optional:    true,
				Computed:    true,

				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"match_type": schema.StringAttribute{
				Description: "How the `view_name` is matched against RUM view names.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("exact", "prefix"),
				},

				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"org_id": schema.Int64Attribute{
				Description: "The ID of the organization that owns this mapping.",
				Computed:    true,
			},
			"service": schema.StringAttribute{
				Description: "The RUM application's service name. For browser applications, may be empty. For mobile applications, this is the service that scopes the ownership.",
				Optional:    true,
				Computed:    true,

				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_handle": schema.StringAttribute{
				Description: "The handle of the team that owns the matched RUM views.",
				Required:    true,

				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"view_name": schema.StringAttribute{
				Description: "The RUM view name to match, or its prefix when `match_type` is `prefix`.",
				Required:    true,

				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *datadogRumTeamOwnershipResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *datadogRumTeamOwnershipResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state datadogRumTeamOwnershipResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	bodyAttributes := datadogV2.NewTeamsOwnershipMappingCreateDataAttributesWithDefaults()

	if !state.ApplicationId.IsNull() && !state.ApplicationId.IsUnknown() {
		applicationIdParsed, err := uuid.Parse(state.ApplicationId.ValueString())
		if err != nil {
			response.Diagnostics.AddError("Invalid application_id", err.Error())
			return
		}
		bodyAttributes.SetApplicationId(applicationIdParsed)
	}

	if !state.MatchType.IsNull() && !state.MatchType.IsUnknown() {
		bodyAttributes.SetMatchType(datadogV2.TeamsOwnershipMatchType(state.MatchType.ValueString()))
	}

	if !state.Service.IsNull() && !state.Service.IsUnknown() {
		bodyAttributes.SetService(state.Service.ValueString())
	}

	bodyAttributes.SetTeamHandle(state.TeamHandle.ValueString())

	bodyAttributes.SetViewName(state.ViewName.ValueString())
	bodyData := datadogV2.NewTeamsOwnershipMappingCreateDataWithDefaults()
	bodyData.SetType(datadogV2.TeamsOwnershipMappingType("teams_ownership_mappings"))
	bodyData.SetAttributes(*bodyAttributes)
	body := datadogV2.NewTeamsOwnershipMappingCreateRequestWithDefaults()
	body.SetData(*bodyData)
	resp, _, err := r.Api.CreateTeamsOwnershipMapping(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating rum_team_ownership"))
		return
	}
	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datadogRumTeamOwnershipResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state datadogRumTeamOwnershipResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetTeamsOwnershipMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading rum_team_ownership"))
		return
	}
	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datadogRumTeamOwnershipResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update should not be called", "Updating this resource should replace it.")
}

func (r *datadogRumTeamOwnershipResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state datadogRumTeamOwnershipResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteTeamsOwnershipMapping(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting rum_team_ownership"))
	}
}

// updateState maps *datadogV2.TeamsOwnershipMappingResponse into state, returning diagnostics
// rather than writing them so it stays free of the resource request/response
// types and can fail at a specific schema path. Shared by Create, Read and
// Update, which all resolve the same response type.
func (r *datadogRumTeamOwnershipResource) updateState(ctx context.Context, state *datadogRumTeamOwnershipResourceModel, resp *datadogV2.TeamsOwnershipMappingResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := resp.Data.GetAttributes()
	if id, ok := resp.Data.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(*id)
	}
	if applicationId, ok := attributes.GetApplicationIdOk(); ok && applicationId != nil {
		state.ApplicationId = types.StringValue(*applicationId)
	}
	if matchType, ok := attributes.GetMatchTypeOk(); ok && matchType != nil {
		state.MatchType = types.StringValue(string(*matchType))
	}
	if orgId, ok := attributes.GetOrgIdOk(); ok && orgId != nil {
		state.OrgId = types.Int64Value(int64(*orgId))
	}
	if service, ok := attributes.GetServiceOk(); ok && service != nil {
		state.Service = types.StringValue(*service)
	}
	if teamHandle, ok := attributes.GetTeamHandleOk(); ok && teamHandle != nil {
		state.TeamHandle = types.StringValue(*teamHandle)
	}
	if viewName, ok := attributes.GetViewNameOk(); ok && viewName != nil {
		state.ViewName = types.StringValue(*viewName)
	}
	return diags
}
