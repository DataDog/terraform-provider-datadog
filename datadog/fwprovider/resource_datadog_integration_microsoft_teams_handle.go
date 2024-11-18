package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &tenantBasedHandleResource{}
	_ resource.ResourceWithImportState = &tenantBasedHandleResource{}
)

type tenantBasedHandleResource struct {
	Api  *datadogV2.MicrosoftTeamsIntegrationApi
	Auth context.Context
}

type tenantBasedHandleModel struct {
	ID          types.String `tfsdk:"id"`
	ChannelName types.String `tfsdk:"channel_name"`
	TeamName    types.String `tfsdk:"team_name"`
	TenantName  types.String `tfsdk:"tenant_name"`
	Name        types.String `tfsdk:"name"`
}

func NewTenantBasedHandleResource() resource.Resource {
	return &tenantBasedHandleResource{}
}

func (r *tenantBasedHandleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMicrosoftTeamsIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *tenantBasedHandleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_ms_teams_tenant_based_handle"
}

func (r *tenantBasedHandleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Resource for interacting with Datadog Microsoft Teams Integration tenant-based handles.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Your tenant-based handle name.",
			},
			"tenant_name": schema.StringAttribute{
				Required:    true,
				Description: "Your tenant name.",
			},
			"team_name": schema.StringAttribute{
				Description: "Your team name.",
				Required:    true,
			},
			"channel_name": schema.StringAttribute{
				Description: "Your channel name.",
				Required:    true,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *tenantBasedHandleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *tenantBasedHandleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tenantBasedHandleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	// Check if handle exists
	resp, httpResp, err := r.Api.GetTenantBasedHandle(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving tenant-based handle"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	diags := r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tenantBasedHandleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state tenantBasedHandleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildTenantBasedHandleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateTenantBasedHandle(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating tenant-based handle"))
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

func (r *tenantBasedHandleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state tenantBasedHandleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildTenantBasedHandleUpdateRequestBody(ctx, &state)

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateTenantBasedHandle(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating tenant-based handle"))
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

func (r *tenantBasedHandleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state tenantBasedHandleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteTenantBasedHandle(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting tenant-based handle"))
		return
	}
}

func (r *tenantBasedHandleResource) updateState(ctx context.Context, state *tenantBasedHandleModel, resp *datadogV2.MicrosoftTeamsTenantBasedHandleResponse) diag.Diagnostics {
	diags := diag.Diagnostics{}
	state.ID = types.StringValue(resp.Data.GetId())
	fullHandleDataList, _, err := r.Api.ListTenantBasedHandles(r.Auth, datadogV2.ListTenantBasedHandlesOptionalParameters{Name: resp.Data.Attributes.Name})
	if err != nil {
		diags.AddError("Could not get remote state: ", err.Error())
		return diags
	}
	var fullHandleData datadogV2.MicrosoftTeamsTenantBasedHandleInfoResponseData
	if len(fullHandleDataList.Data) == 0 {
		diags.AddError("No matches for handle with name: "+*resp.Data.Attributes.Name, "")
		return diags
	}

	fullHandleData = fullHandleDataList.Data[0]
	attributes := fullHandleData.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok && name != nil {
		state.Name = types.StringValue(*name)
	}

	if tenantName, ok := attributes.GetTenantNameOk(); ok && tenantName != nil {
		state.TenantName = types.StringValue(*tenantName)
	}

	if teamName, ok := attributes.GetTeamNameOk(); ok && teamName != nil {
		state.TeamName = types.StringValue(*teamName)
	}

	if channelName, ok := attributes.GetChannelNameOk(); ok && channelName != nil {
		state.ChannelName = types.StringValue(*channelName)
	}
	return diags
}

func (r *tenantBasedHandleResource) buildTenantBasedHandleRequestBody(ctx context.Context, state *tenantBasedHandleModel) (*datadogV2.MicrosoftTeamsCreateTenantBasedHandleRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewMicrosoftTeamsTenantBasedHandleRequestAttributesWithDefaults()
	channelData, _, err := r.Api.GetChannelByName(r.Auth, strings.ReplaceAll(state.TenantName.ValueString(), "\"", ""), state.TeamName.ValueString(), state.ChannelName.ValueString())
	if err != nil {
		channelInfo := fmt.Sprintf("Tenant Name: %s\nTeam Name: %s\nChannel Name: %s\n", state.TenantName.ValueString(), state.TeamName.ValueString(), state.ChannelName.ValueString())
		diags.AddError("Channel data not found for: \n"+channelInfo, err.Error())
		return nil, diags
	}

	attributes.SetName(state.Name.ValueString())
	attributes.SetTenantId(*channelData.Data.Attributes.TenantId)
	attributes.SetTeamId(*channelData.Data.Attributes.TeamId)
	attributes.SetChannelId(*channelData.Data.Id)

	req := datadogV2.NewMicrosoftTeamsCreateTenantBasedHandleRequestWithDefaults()
	req.Data = *datadogV2.NewMicrosoftTeamsTenantBasedHandleRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *tenantBasedHandleResource) buildTenantBasedHandleUpdateRequestBody(ctx context.Context, state *tenantBasedHandleModel) (*datadogV2.MicrosoftTeamsUpdateTenantBasedHandleRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewMicrosoftTeamsTenantBasedHandleAttributesWithDefaults()
	channelData, _, err := r.Api.GetChannelByName(r.Auth, strings.ReplaceAll(state.TenantName.ValueString(), "\"", ""), state.TeamName.ValueString(), state.ChannelName.ValueString())
	if err != nil {
		channelInfo := fmt.Sprintf("Tenant Name: %s\nTeam Name: %s\nChannel Name: %s\n", state.TenantName.ValueString(), state.TeamName.ValueString(), state.ChannelName.ValueString())
		diags.AddError("Channel data not found for: \n"+channelInfo, err.Error())
		return nil, diags
	}

	attributes.SetName(state.Name.ValueString())
	attributes.SetTenantId(*channelData.Data.Attributes.TenantId)
	attributes.SetTeamId(*channelData.Data.Attributes.TeamId)
	attributes.SetChannelId(*channelData.Data.Id)

	req := datadogV2.NewMicrosoftTeamsUpdateTenantBasedHandleRequestWithDefaults()
	req.Data = *datadogV2.NewMicrosoftTeamsUpdateTenantBasedHandleRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
