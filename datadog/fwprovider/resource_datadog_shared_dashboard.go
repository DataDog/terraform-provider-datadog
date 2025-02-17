package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &sharedDashboardResource{}
	_ resource.ResourceWithImportState = &sharedDashboardResource{}
)

type sharedDashboardResource struct {
	Api  *datadogV1.DashboardsApi
	Auth context.Context
}

type sharedDashboardModel struct {
	ID                          types.String                   `tfsdk:"id"`
	CreatedAt                   types.String                   `tfsdk:"created_at"`
	DashboardId                 types.String                   `tfsdk:"dashboard_id"`
	DashboardType               types.String                   `tfsdk:"dashboard_type"`
	GlobalTimeSelectableEnabled types.Bool                     `tfsdk:"global_time_selectable_enabled"`
	PublicUrl                   types.String                   `tfsdk:"public_url"`
	ShareType                   types.String                   `tfsdk:"share_type"`
	Token                       types.String                   `tfsdk:"token"`
	ShareList                   types.List                     `tfsdk:"share_list"`
	SelectableTemplateVars      []*selectableTemplateVarsModel `tfsdk:"selectable_template_vars"`
	Author                      *authorModel                   `tfsdk:"author"`
	GlobalTime                  *globalTimeModel               `tfsdk:"global_time"`
}

type selectableTemplateVarsModel struct {
	DefaultValue types.String `tfsdk:"default_value"`
	Name         types.String `tfsdk:"name"`
	Prefix       types.String `tfsdk:"prefix"`
	VisibleTags  types.List   `tfsdk:"visible_tags"`
}

type authorModel struct {
	Handle types.String `tfsdk:"handle"`
	Name   types.String `tfsdk:"name"`
}

type globalTimeModel struct {
	LiveSpan types.String `tfsdk:"live_span"`
}

func NewSharedDashboardResource() resource.Resource {
	return &sharedDashboardResource{}
}

func (r *sharedDashboardResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetDashboardsApiV1()
	r.Auth = providerData.Auth
}

func (r *sharedDashboardResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "shared_dashboard"
}

func (r *sharedDashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog SharedDashboard resource. This can be used to create and manage Datadog shared_dashboard.",
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Date the dashboard was shared.",
			},
			"dashboard_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the dashboard to share.",
			},
			"dashboard_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the associated private dashboard.",
			},
			"global_time_selectable_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to allow viewers to select a different global time setting for the shared dashboard.",
			},
			"public_url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the shared dashboard.",
			},
			"share_type": schema.StringAttribute{
				Optional:    true,
				Description: "Type of sharing access (either open to anyone who has the public URL or invite-only).",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Description: "A unique token assigned to the shared dashboard.",
			},
			"share_list": schema.ListAttribute{
				Optional:    true,
				Description: "List of email addresses that can receive an invitation to access to the shared dashboard.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"selectable_template_vars": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"default_value": schema.StringAttribute{
							Optional:    true,
							Description: "The default value of the template variable.",
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the template variable.",
						},
						"prefix": schema.StringAttribute{
							Optional:    true,
							Description: "The tag/attribute key associated with the template variable.",
						},
						"visible_tags": schema.ListAttribute{
							Optional:    true,
							Description: "List of visible tag values on the shared dashboard.",
							ElementType: types.StringType,
						},
					},
				},
			},
			"author": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"handle": schema.StringAttribute{
						Computed:    true,
						Description: "Identifier of the user who shared the dashboard.",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "Name of the user who shared the dashboard.",
					},
				},
			},
			"global_time": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"live_span": schema.StringAttribute{
						Optional:    true,
						Description: "Dashboard global time live_span selection",
					},
				},
			},
		},
	}
}

func (r *sharedDashboardResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *sharedDashboardResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state sharedDashboardModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetPublicDashboard(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SharedDashboard"))
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

func (r *sharedDashboardResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state sharedDashboardModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildSharedDashboardRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreatePublicDashboard(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SharedDashboard"))
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

func (r *sharedDashboardResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state sharedDashboardModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildSharedDashboardUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdatePublicDashboard(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SharedDashboard"))
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

func (r *sharedDashboardResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state sharedDashboardModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	_, httpResp, err := r.Api.DeletePublicDashboard(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting shared_dashboard"))
		return
	}
}

func (r *sharedDashboardResource) updateState(ctx context.Context, state *sharedDashboardModel, resp *datadogV1.SharedDashboard) {
	state.ID = types.StringValue(resp.GetToken())

	if createdAt, ok := resp.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(*createdAt)
	}

	state.DashboardId = types.StringValue(resp.GetDashboardId())

	state.DashboardType = types.StringValue(string(resp.GetDashboardType()))

	if globalTimeSelectableEnabled, ok := resp.GetGlobalTimeSelectableEnabledOk(); ok {
		state.GlobalTimeSelectableEnabled = types.BoolValue(*globalTimeSelectableEnabled)
	}

	if publicUrl, ok := resp.GetPublicUrlOk(); ok {
		state.PublicUrl = types.StringValue(*publicUrl)
	}

	if shareType, ok := resp.GetShareTypeOk(); ok {
		state.ShareType = types.StringValue(string(*shareType))
	}

	if token, ok := resp.GetTokenOk(); ok {
		state.Token = types.StringValue(*token)
	}

	if shareList, ok := resp.GetShareListOk(); ok && len(*shareList) > 0 {
		state.ShareList, _ = types.ListValueFrom(ctx, types.StringType, *shareList)
	}

	if selectableTemplateVars, ok := resp.GetSelectableTemplateVarsOk(); ok && len(*selectableTemplateVars) > 0 {
		state.SelectableTemplateVars = []*selectableTemplateVarsModel{}
		for _, selectableTemplateVarsDd := range *selectableTemplateVars {
			selectableTemplateVarsTfItem := selectableTemplateVarsModel{}

			if selectableTemplateVars, ok := selectableTemplateVarsDd.GetSelectableTemplateVarsOk(); ok {

				selectableTemplateVarsTf := selectableTemplateVarsModel{}
				if defaultValue, ok := selectableTemplateVars.GetDefaultValueOk(); ok {
					selectableTemplateVarsTf.DefaultValue = types.StringValue(*defaultValue)
				}
				if name, ok := selectableTemplateVars.GetNameOk(); ok {
					selectableTemplateVarsTf.Name = types.StringValue(*name)
				}
				if prefix, ok := selectableTemplateVars.GetPrefixOk(); ok {
					selectableTemplateVarsTf.Prefix = types.StringValue(*prefix)
				}
				if visibleTags, ok := selectableTemplateVars.GetVisibleTagsOk(); ok && len(*visibleTags) > 0 {
					selectableTemplateVarsTf.VisibleTags, _ = types.ListValueFrom(ctx, types.StringType, *visibleTags)
				}

				selectableTemplateVarsTfItem.SelectableTemplateVars = &selectableTemplateVarsTf
			}
			state.SelectableTemplateVars = append(state.SelectableTemplateVars, &selectableTemplateVarsTfItem)
		}
	}

	if author, ok := resp.GetAuthorOk(); ok {

		authorTf := authorModel{}
		if handle, ok := author.GetHandleOk(); ok {
			authorTf.Handle = types.StringValue(*handle)
		}
		if name, ok := author.GetNameOk(); ok {
			authorTf.Name = types.StringValue(*name)
		}

		state.Author = &authorTf
	}

	if globalTime, ok := resp.GetGlobalTimeOk(); ok {

		globalTimeTf := globalTimeModel{}
		if liveSpan, ok := globalTime.GetLiveSpanOk(); ok {
			globalTimeTf.LiveSpan = types.StringValue(string(*liveSpan))
		}

		state.GlobalTime = &globalTimeTf
	}
}

func (r *sharedDashboardResource) buildSharedDashboardRequestBody(ctx context.Context, state *sharedDashboardModel) (*datadogV1.SharedDashboard, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV1.SharedDashboard{}

	if !state.DashboardId.IsNull() {
		req.SetDashboardId(state.DashboardId.ValueString())
	}
	if !state.DashboardType.IsNull() {
		req.SetDashboardType(datadogV1.DashboardType(state.DashboardType.ValueString()))
	}
	if !state.GlobalTimeSelectableEnabled.IsNull() {
		req.SetGlobalTimeSelectableEnabled(state.GlobalTimeSelectableEnabled.ValueBool())
	}
	if !state.ShareType.IsNull() {
		req.SetShareType(datadogV1.DashboardShareType(state.ShareType.ValueString()))
	}

	if !state.ShareList.IsNull() {
		var shareList []string
		diags.Append(state.ShareList.ElementsAs(ctx, &shareList, false)...)
		req.SetShareList(shareList)
	}

	if state.SelectableTemplateVars != nil {
		var selectableTemplateVars []datadogV1.SelectableTemplateVariableItems
		for _, selectableTemplateVarsTFItem := range state.SelectableTemplateVars {
			selectableTemplateVarsDDItem := datadogV1.NewSelectableTemplateVariableItems()

			if !selectableTemplateVarsTFItem.DefaultValue.IsNull() {
				selectableTemplateVarsDDItem.SetDefaultValue(selectableTemplateVarsTFItem.DefaultValue.ValueString())
			}
			if !selectableTemplateVarsTFItem.Name.IsNull() {
				selectableTemplateVarsDDItem.SetName(selectableTemplateVarsTFItem.Name.ValueString())
			}
			if !selectableTemplateVarsTFItem.Prefix.IsNull() {
				selectableTemplateVarsDDItem.SetPrefix(selectableTemplateVarsTFItem.Prefix.ValueString())
			}

			if !selectableTemplateVarsTFItem.VisibleTags.IsNull() {
				var visibleTags []string
				diags.Append(selectableTemplateVarsTFItem.VisibleTags.ElementsAs(ctx, &visibleTags, false)...)
				selectableTemplateVarsDDItem.SetVisibleTags(visibleTags)
			}
		}
		req.SetSelectableTemplateVars(selectableTemplateVars)
	}

	if state.GlobalTime != nil {
		var globalTime datadogV1.DashboardGlobalTime

		if !state.GlobalTime.LiveSpan.IsNull() {
			globalTime.SetLiveSpan(datadogV1.DashboardGlobalTimeLiveSpan(state.GlobalTime.LiveSpan.ValueString()))
		}
		req.GlobalTime = &globalTime
	}

	return req, diags
}

func (r *sharedDashboardResource) buildSharedDashboardUpdateRequestBody(ctx context.Context, state *sharedDashboardModel) (*datadogV1.SharedDashboardUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV1.SharedDashboardUpdateRequest{}

	if !state.GlobalTimeSelectableEnabled.IsNull() {
		req.SetGlobalTimeSelectableEnabled(state.GlobalTimeSelectableEnabled.ValueBool())
	}
	if !state.ShareType.IsNull() {
		req.SetShareType(datadogV1.DashboardShareType(state.ShareType.ValueString()))
	}

	if !state.ShareList.IsNull() {
		var shareList []string
		diags.Append(state.ShareList.ElementsAs(ctx, &shareList, false)...)
		req.SetShareList(shareList)
	}

	if state.SelectableTemplateVars != nil {
		var selectableTemplateVars []datadogV1.SelectableTemplateVariableItems
		for _, selectableTemplateVarsTFItem := range state.SelectableTemplateVars {
			selectableTemplateVarsDDItem := datadogV1.NewSelectableTemplateVariableItems()

			if !selectableTemplateVarsTFItem.DefaultValue.IsNull() {
				selectableTemplateVarsDDItem.SetDefaultValue(selectableTemplateVarsTFItem.DefaultValue.ValueString())
			}
			if !selectableTemplateVarsTFItem.Name.IsNull() {
				selectableTemplateVarsDDItem.SetName(selectableTemplateVarsTFItem.Name.ValueString())
			}
			if !selectableTemplateVarsTFItem.Prefix.IsNull() {
				selectableTemplateVarsDDItem.SetPrefix(selectableTemplateVarsTFItem.Prefix.ValueString())
			}

			if !selectableTemplateVarsTFItem.VisibleTags.IsNull() {
				var visibleTags []string
				diags.Append(selectableTemplateVarsTFItem.VisibleTags.ElementsAs(ctx, &visibleTags, false)...)
				selectableTemplateVarsDDItem.SetVisibleTags(visibleTags)
			}
		}
		req.SetSelectableTemplateVars(selectableTemplateVars)
	}

	if state.GlobalTime != nil {
		var globalTime datadogV1.SharedDashboardUpdateRequestGlobalTime

		if !state.GlobalTime.LiveSpan.IsNull() {
			globalTime.SetLiveSpan(datadogV1.DashboardGlobalTimeLiveSpan(state.GlobalTime.LiveSpan.ValueString()))
		}
		req.GlobalTime = *datadogV1.NewNullableSharedDashboardUpdateRequestGlobalTime(&globalTime)
	}

	return req, diags
}
