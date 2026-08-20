package fwprovider

import (
	"context"
	"fmt"
	"reflect"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
	_ resource.ResourceWithModifyPlan  = &statusPageResource{}
)

type statusPageResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

type statusPageModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Type                      types.String `tfsdk:"type"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	DomainPrefix              types.String `tfsdk:"domain_prefix"`
	CustomDomain              types.String `tfsdk:"custom_domain"`
	CustomDomainEnabled       types.Bool   `tfsdk:"custom_domain_enabled"`
	PageURL                   types.String `tfsdk:"page_url"`
	VisualizationType         types.String `tfsdk:"visualization_type"`
	SubscriptionsEnabled      types.Bool   `tfsdk:"subscriptions_enabled"`
	SlackSubscriptionsEnabled types.Bool   `tfsdk:"slack_subscriptions_enabled"`
	CompanyLogo               types.String `tfsdk:"company_logo"`
	Favicon                   types.String `tfsdk:"favicon"`
	EmailHeaderImage          types.String `tfsdk:"email_header_image"`
	SlackAppIcon              types.String `tfsdk:"slack_app_icon"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	ModifiedAt                types.String `tfsdk:"modified_at"`
}

func NewStatusPageResource() resource.Resource {
	return &statusPageResource{}
}

// ModifyPlan pins modified_at back to its prior state value when no other
// attribute is actually changing. modified_at genuinely changes on real
// updates, so it can't use UseStateForUnknown() (which is unconditional),
// but without any plan modifier Terraform 1.1.5's core marks it Unknown even
// on a no-op replan, unlike newer Terraform-core versions.
func (r *statusPageResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var planModel, stateModel statusPageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priorModifiedAt := stateModel.ModifiedAt
	planModel.ModifiedAt = types.StringNull()
	stateModel.ModifiedAt = types.StringNull()
	if !reflect.DeepEqual(planModel, stateModel) {
		// A real change is happening elsewhere in the plan; leave modified_at
		// as Terraform's core computed it (Unknown).
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("modified_at"), priorModifiedAt)...)
}

// statusPageURLShape buckets a status page type by the page_url format the
// API derives from it. "internal" pages URL to a Datadog-domain viewer path;
// every other type (public, and the upcoming private) URLs to a
// domain_prefix-based subdomain, so they share a shape.
func statusPageURLShape(pageType types.String) string {
	if pageType.ValueString() == "internal" {
		return "internal"
	}
	return "domain_prefix"
}

// statusPageURLPlanModifier behaves like stringplanmodifier.UseStateForUnknown(),
// except it lets page_url stay Unknown (instead of reusing the prior state
// value) when domain_prefix changes, or when type changes across a URL-shape
// boundary, since the API derives page_url from those fields server-side.
type statusPageURLPlanModifier struct{}

func (m statusPageURLPlanModifier) Description(_ context.Context) string {
	return "Recomputes page_url when domain_prefix changes, or type changes across a URL-shape boundary, instead of reusing the prior state value."
}

func (m statusPageURLPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m statusPageURLPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var stateType, planType, stateDomainPrefix, planDomainPrefix types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("type"), &stateType)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("type"), &planType)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("domain_prefix"), &stateDomainPrefix)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("domain_prefix"), &planDomainPrefix)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if statusPageURLShape(stateType) != statusPageURLShape(planType) || !stateDomainPrefix.Equal(planDomainPrefix) {
		// page_url is derived server-side from type/domain_prefix; force a
		// recompute instead of letting Terraform's default proposal carry
		// forward the now-stale prior value.
		resp.PlanValue = types.StringUnknown()
		return
	}

	if req.PlanValue.IsUnknown() {
		resp.PlanValue = req.StateValue
	}
}

func (r *statusPageResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "status_page"
}

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog status page resource. This can be used to create and manage Datadog status pages.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the status page.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the status page.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the status page. Valid values are: public, internal.",
				Required:    true,
			},
			"domain_prefix": schema.StringAttribute{
				Description: "The subdomain prefix used to build the status page's URL.",
				Required:    true,
			},
			"visualization_type": schema.StringAttribute{
				Description: "How component statuses are visualized on the page. Valid values are: bars_and_uptime_percentage, bars_only, component_name_only.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the status page is published. Managed by the publish/unpublish API operations, not by this resource; always reflects the page's current state.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_domain": schema.StringAttribute{
				Description: "The custom domain configured for the status page, if any. Managed via a separate custom-domain flow, not by this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_domain_enabled": schema.BoolAttribute{
				Description: "Whether the custom domain is enabled for the status page. Managed via a separate custom-domain flow, not by this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"page_url": schema.StringAttribute{
				Description: "The URL of the status page.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					statusPageURLPlanModifier{},
				},
			},
			"subscriptions_enabled": schema.BoolAttribute{
				Description: "Whether subscriber notifications are enabled for the status page.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"slack_subscriptions_enabled": schema.BoolAttribute{
				Description: "Whether Slack subscriber notifications are enabled for the status page.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"company_logo": schema.StringAttribute{
				Description: "The company logo displayed on the status page.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"favicon": schema.StringAttribute{
				Description: "The favicon displayed for the status page.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email_header_image": schema.StringAttribute{
				Description: "The header image included in subscriber emails.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slack_app_icon": schema.StringAttribute{
				Description: "The icon used for the status page's Slack app integration.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the status page was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp when the status page was last modified. Changes on every real update, so this must stay Unknown during an actual update plan rather than using UseStateForUnknown.",
				Computed:    true,
			},
		},
	}
}

func (r *statusPageResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	r.Auth = providerData.Auth
}

func (r *statusPageResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan statusPageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := datadogV2.CreateStatusPageRequestDataAttributes{
		Name:              plan.Name.ValueString(),
		Type:              datadogV2.CreateStatusPageRequestDataAttributesType(plan.Type.ValueString()),
		DomainPrefix:      plan.DomainPrefix.ValueString(),
		VisualizationType: datadogV2.CreateStatusPageRequestDataAttributesVisualizationType(plan.VisualizationType.ValueString()),
	}

	if !plan.SubscriptionsEnabled.IsNull() && !plan.SubscriptionsEnabled.IsUnknown() {
		attributes.SubscriptionsEnabled = plan.SubscriptionsEnabled.ValueBoolPointer()
	}
	if !plan.SlackSubscriptionsEnabled.IsNull() && !plan.SlackSubscriptionsEnabled.IsUnknown() {
		attributes.SlackSubscriptionsEnabled = plan.SlackSubscriptionsEnabled.ValueBoolPointer()
	}
	if !plan.CompanyLogo.IsNull() && !plan.CompanyLogo.IsUnknown() {
		attributes.CompanyLogo = plan.CompanyLogo.ValueStringPointer()
	}
	if !plan.Favicon.IsNull() && !plan.Favicon.IsUnknown() {
		attributes.Favicon = plan.Favicon.ValueStringPointer()
	}
	if !plan.EmailHeaderImage.IsNull() && !plan.EmailHeaderImage.IsUnknown() {
		attributes.EmailHeaderImage = plan.EmailHeaderImage.ValueStringPointer()
	}
	if !plan.SlackAppIcon.IsNull() && !plan.SlackAppIcon.IsUnknown() {
		attributes.SlackAppIcon = plan.SlackAppIcon.ValueStringPointer()
	}

	body := datadogV2.CreateStatusPageRequest{
		Data: &datadogV2.CreateStatusPageRequestData{
			Type:       datadogV2.STATUSPAGEDATATYPE_STATUS_PAGES,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.CreateStatusPage(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating status page",
			fmt.Sprintf("Could not create status page, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating status page",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state statusPageModel
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetStatusPage(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading status page",
			"Could not read status page ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// requireUnpublished returns a diagnostic if the status page is currently published, since the API
// rejects changing a page's type or deleting a page while it is enabled (published). trailingClause
// completes "Status page must be unpublished before <trailingClause>", e.g. "its type can be changed".
func requireUnpublished(state *statusPageModel, trailingClause string) diag.Diagnostic {
	if state.Enabled.ValueBool() {
		return diag.NewErrorDiagnostic(
			fmt.Sprintf("Status page must be unpublished before %s", trailingClause),
			"This status page is currently published (enabled=true). Publishing/unpublishing is managed outside this resource, via the status page publish/unpublish API or UI; unpublish the page first, then retry.",
		)
	}
	return nil
}

func (r *statusPageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan statusPageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state statusPageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Type.Equal(state.Type) {
		if d := requireUnpublished(&state, "its type can be changed"); d != nil {
			response.Diagnostics.Append(d)
			return
		}
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	name := plan.Name.ValueString()
	domainPrefix := plan.DomainPrefix.ValueString()
	pageType := datadogV2.CreateStatusPageRequestDataAttributesType(plan.Type.ValueString())
	visualizationType := datadogV2.CreateStatusPageRequestDataAttributesVisualizationType(plan.VisualizationType.ValueString())

	attributes := datadogV2.PatchStatusPageRequestDataAttributes{
		Name:              &name,
		Type:              &pageType,
		DomainPrefix:      &domainPrefix,
		VisualizationType: &visualizationType,
	}

	if !plan.SubscriptionsEnabled.IsNull() && !plan.SubscriptionsEnabled.IsUnknown() {
		attributes.SubscriptionsEnabled = plan.SubscriptionsEnabled.ValueBoolPointer()
	}
	if !plan.SlackSubscriptionsEnabled.IsNull() && !plan.SlackSubscriptionsEnabled.IsUnknown() {
		attributes.SlackSubscriptionsEnabled = plan.SlackSubscriptionsEnabled.ValueBoolPointer()
	}
	if !plan.CompanyLogo.IsNull() && !plan.CompanyLogo.IsUnknown() {
		attributes.CompanyLogo = plan.CompanyLogo.ValueStringPointer()
	}
	if !plan.Favicon.IsNull() && !plan.Favicon.IsUnknown() {
		attributes.Favicon = plan.Favicon.ValueStringPointer()
	}
	if !plan.EmailHeaderImage.IsNull() && !plan.EmailHeaderImage.IsUnknown() {
		attributes.EmailHeaderImage = plan.EmailHeaderImage.ValueStringPointer()
	}
	if !plan.SlackAppIcon.IsNull() && !plan.SlackAppIcon.IsUnknown() {
		attributes.SlackAppIcon = plan.SlackAppIcon.ValueStringPointer()
	}

	body := datadogV2.PatchStatusPageRequest{
		Data: &datadogV2.PatchStatusPageRequestData{
			Id:         id,
			Type:       datadogV2.STATUSPAGEDATATYPE_STATUS_PAGES,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.UpdateStatusPage(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating status page",
			fmt.Sprintf("Could not update status page ID %s, unexpected error: %s. HTTP Response: %v", plan.ID.ValueString(), err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating status page",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *statusPageResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if d := requireUnpublished(&state, "it can be deleted"); d != nil {
		response.Diagnostics.Append(d)
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteStatusPage(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting status page",
			"Could not delete status page ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *statusPageResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *statusPageResource) updateStateFromResponse(state *statusPageModel, resp *datadogV2.StatusPage) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		if name, ok := attributes.GetNameOk(); ok && name != nil {
			state.Name = types.StringValue(*name)
		}
		if pageType, ok := attributes.GetTypeOk(); ok && pageType != nil {
			state.Type = types.StringValue(string(*pageType))
		}
		if enabled, ok := attributes.GetEnabledOk(); ok && enabled != nil {
			state.Enabled = types.BoolValue(*enabled)
		}
		if domainPrefix, ok := attributes.GetDomainPrefixOk(); ok && domainPrefix != nil {
			state.DomainPrefix = types.StringValue(*domainPrefix)
		}
		if customDomain, ok := attributes.GetCustomDomainOk(); ok && customDomain != nil {
			state.CustomDomain = types.StringValue(*customDomain)
		} else {
			state.CustomDomain = types.StringNull()
		}
		if customDomainEnabled, ok := attributes.GetCustomDomainEnabledOk(); ok && customDomainEnabled != nil {
			state.CustomDomainEnabled = types.BoolValue(*customDomainEnabled)
		}
		if pageURL, ok := attributes.GetPageUrlOk(); ok && pageURL != nil {
			state.PageURL = types.StringValue(*pageURL)
		}
		if visualizationType, ok := attributes.GetVisualizationTypeOk(); ok && visualizationType != nil {
			state.VisualizationType = types.StringValue(string(*visualizationType))
		}
		if subscriptionsEnabled, ok := attributes.GetSubscriptionsEnabledOk(); ok && subscriptionsEnabled != nil {
			state.SubscriptionsEnabled = types.BoolValue(*subscriptionsEnabled)
		}
		if slackSubscriptionsEnabled, ok := attributes.GetSlackSubscriptionsEnabledOk(); ok && slackSubscriptionsEnabled != nil {
			state.SlackSubscriptionsEnabled = types.BoolValue(*slackSubscriptionsEnabled)
		}
		if companyLogo, ok := attributes.GetCompanyLogoOk(); ok && companyLogo != nil {
			state.CompanyLogo = types.StringValue(*companyLogo)
		} else {
			state.CompanyLogo = types.StringNull()
		}
		if favicon, ok := attributes.GetFaviconOk(); ok && favicon != nil {
			state.Favicon = types.StringValue(*favicon)
		} else {
			state.Favicon = types.StringNull()
		}
		if emailHeaderImage, ok := attributes.GetEmailHeaderImageOk(); ok && emailHeaderImage != nil {
			state.EmailHeaderImage = types.StringValue(*emailHeaderImage)
		} else {
			state.EmailHeaderImage = types.StringNull()
		}
		if slackAppIcon, ok := attributes.GetSlackAppIconOk(); ok && slackAppIcon != nil {
			state.SlackAppIcon = types.StringValue(*slackAppIcon)
		} else {
			state.SlackAppIcon = types.StringNull()
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.Format("2006-01-02T15:04:05Z"))
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.Format("2006-01-02T15:04:05Z"))
		}
	}
}
