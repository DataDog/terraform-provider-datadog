package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
)

func NewStatusPageResource() resource.Resource {
	return &statusPageResource{}
}

type statusPageResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	DomainPrefix              types.String `tfsdk:"domain_prefix"`
	Type                      types.String `tfsdk:"type"`
	VisualizationType         types.String `tfsdk:"visualization_type"`
	CompanyLogo               types.String `tfsdk:"company_logo"`
	Favicon                   types.String `tfsdk:"favicon"`
	EmailHeaderImage          types.String `tfsdk:"email_header_image"`
	SlackAppIcon              types.String `tfsdk:"slack_app_icon"`
	SubscriptionsEnabled      types.Bool   `tfsdk:"subscriptions_enabled"`
	SlackSubscriptionsEnabled types.Bool   `tfsdk:"slack_subscriptions_enabled"`
	PageURL                   types.String `tfsdk:"page_url"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	CustomDomain              types.String `tfsdk:"custom_domain"`
	CustomDomainEnabled       types.Bool   `tfsdk:"custom_domain_enabled"`
}

type statusPageResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (r *statusPageResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		return
	}
	r.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	r.Auth = providerData.Auth
}

func (r *statusPageResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "status_page"
}

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Status Page resource. This can be used to create and manage the page-level structure of a Datadog status page. Components are managed via the `datadog_status_page_component` resource.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the status page.",
			},
			"domain_prefix": schema.StringAttribute{
				Required:    true,
				Description: "The domain prefix for the status page URL.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of status page.",
				Validators:  []validator.String{stringvalidator.OneOf("public", "internal")},
			},
			"visualization_type": schema.StringAttribute{
				Required:    true,
				Description: "How component status is visualized.",
				Validators:  []validator.String{stringvalidator.OneOf("bars_and_uptime_percentage", "bars_only", "component_name_only")},
			},
			"company_logo": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the company logo displayed on the status page.",
			},
			"favicon": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the favicon for the status page.",
			},
			"email_header_image": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the image used in subscription email headers.",
			},
			"slack_app_icon": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the Slack app icon.",
			},
			"subscriptions_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether subscriptions are enabled for the status page.",
			},
			"slack_subscriptions_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether Slack subscriptions are enabled for the status page.",
			},
			"page_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the status page.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the status page is published/enabled.",
			},
			"custom_domain": schema.StringAttribute{
				Computed:    true,
				Description: "The custom domain configured for the status page.",
			},
			"custom_domain_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether a custom domain is enabled for the status page.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *statusPageResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state statusPageResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildCreateRequest(&state)
	resp, _, err := r.Api.CreateStatusPage(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating status page"))
		return
	}
	data := resp.GetData()
	r.updateState(&state, data.GetAttributes(), data.GetId().String())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status page id"))
		return
	}
	resp, httpResp, err := r.Api.GetStatusPage(r.Auth, pid)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving status page"))
		return
	}
	data := resp.GetData()
	r.updateState(&state, data.GetAttributes(), data.GetId().String())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state statusPageResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status page id"))
		return
	}
	body := r.buildPatchRequest(&state, pid)
	resp, _, err := r.Api.UpdateStatusPage(r.Auth, pid, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating status page"))
		return
	}
	data := resp.GetData()
	r.updateState(&state, data.GetAttributes(), data.GetId().String())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status page id"))
		return
	}
	if _, err := r.Api.DeleteStatusPage(r.Auth, pid); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting status page"))
	}
}

func (r *statusPageResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *statusPageResource) buildCreateRequest(state *statusPageResourceModel) *datadogV2.CreateStatusPageRequest {
	attrs := datadogV2.NewCreateStatusPageRequestDataAttributesWithDefaults()
	attrs.SetName(state.Name.ValueString())
	attrs.SetDomainPrefix(state.DomainPrefix.ValueString())
	attrs.SetType(datadogV2.CreateStatusPageRequestDataAttributesType(state.Type.ValueString()))
	attrs.SetVisualizationType(datadogV2.CreateStatusPageRequestDataAttributesVisualizationType(state.VisualizationType.ValueString()))
	r.applyOptionalAttributes(state, attrs.SetCompanyLogo, attrs.SetFavicon, attrs.SetEmailHeaderImage, attrs.SetSlackAppIcon, attrs.SetSubscriptionsEnabled, attrs.SetSlackSubscriptionsEnabled)
	data := datadogV2.NewCreateStatusPageRequestData(*attrs, datadogV2.STATUSPAGEDATATYPE_STATUS_PAGES)
	req := datadogV2.NewCreateStatusPageRequest()
	req.SetData(*data)
	return req
}

func (r *statusPageResource) buildPatchRequest(state *statusPageResourceModel, pid uuid.UUID) *datadogV2.PatchStatusPageRequest {
	attrs := datadogV2.NewPatchStatusPageRequestDataAttributesWithDefaults()
	attrs.SetName(state.Name.ValueString())
	attrs.SetDomainPrefix(state.DomainPrefix.ValueString())
	attrs.SetType(datadogV2.CreateStatusPageRequestDataAttributesType(state.Type.ValueString()))
	attrs.SetVisualizationType(datadogV2.CreateStatusPageRequestDataAttributesVisualizationType(state.VisualizationType.ValueString()))
	r.applyOptionalAttributes(state, attrs.SetCompanyLogo, attrs.SetFavicon, attrs.SetEmailHeaderImage, attrs.SetSlackAppIcon, attrs.SetSubscriptionsEnabled, attrs.SetSlackSubscriptionsEnabled)
	data := datadogV2.NewPatchStatusPageRequestData(*attrs, pid, datadogV2.STATUSPAGEDATATYPE_STATUS_PAGES)
	req := datadogV2.NewPatchStatusPageRequest()
	req.SetData(*data)
	return req
}

func (r *statusPageResource) applyOptionalAttributes(
	state *statusPageResourceModel,
	setCompanyLogo, setFavicon, setEmailHeaderImage, setSlackAppIcon func(string),
	setSubscriptionsEnabled, setSlackSubscriptionsEnabled func(bool),
) {
	if !state.CompanyLogo.IsNull() {
		setCompanyLogo(state.CompanyLogo.ValueString())
	}
	if !state.Favicon.IsNull() {
		setFavicon(state.Favicon.ValueString())
	}
	if !state.EmailHeaderImage.IsNull() {
		setEmailHeaderImage(state.EmailHeaderImage.ValueString())
	}
	if !state.SlackAppIcon.IsNull() {
		setSlackAppIcon(state.SlackAppIcon.ValueString())
	}
	if !state.SubscriptionsEnabled.IsNull() && !state.SubscriptionsEnabled.IsUnknown() {
		setSubscriptionsEnabled(state.SubscriptionsEnabled.ValueBool())
	}
	if !state.SlackSubscriptionsEnabled.IsNull() && !state.SlackSubscriptionsEnabled.IsUnknown() {
		setSlackSubscriptionsEnabled(state.SlackSubscriptionsEnabled.ValueBool())
	}
}

func (r *statusPageResource) updateState(state *statusPageResourceModel, attrs datadogV2.StatusPageDataAttributes, id string) {
	state.ID = types.StringValue(id)
	state.Name = types.StringValue(attrs.GetName())
	state.DomainPrefix = types.StringValue(attrs.GetDomainPrefix())
	state.Type = types.StringValue(string(attrs.GetType()))
	state.VisualizationType = types.StringValue(string(attrs.GetVisualizationType()))
	state.SubscriptionsEnabled = types.BoolValue(attrs.GetSubscriptionsEnabled())
	state.SlackSubscriptionsEnabled = types.BoolValue(attrs.GetSlackSubscriptionsEnabled())
	state.PageURL = types.StringValue(attrs.GetPageUrl())
	state.Enabled = types.BoolValue(attrs.GetEnabled())
	state.CustomDomain = types.StringValue(attrs.GetCustomDomain())
	state.CustomDomainEnabled = types.BoolValue(attrs.GetCustomDomainEnabled())
	state.CompanyLogo = optionalString(attrs.GetCompanyLogo())
	state.Favicon = optionalString(attrs.GetFavicon())
	state.EmailHeaderImage = optionalString(attrs.GetEmailHeaderImage())
	state.SlackAppIcon = optionalString(attrs.GetSlackAppIcon())
}

// optionalString maps an empty API value back to null so an unset optional
// attribute stays null in state (the API echoes unset branding fields as null,
// which the NullableString accessors surface as an empty string).
func optionalString(v string) types.String {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}
