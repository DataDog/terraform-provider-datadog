package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &dashboardSecureEmbedResource{}
	_ resource.ResourceWithImportState = &dashboardSecureEmbedResource{}
)

// --- Go structs for the JSONAPI payloads ---

type secureEmbedGlobalTime struct {
	LiveSpan string `json:"live_span,omitempty"`
}

type secureEmbedTemplateVar struct {
	Name          string   `json:"name"`
	Prefix        string   `json:"prefix,omitempty"`
	Type          string   `json:"type,omitempty"`
	DefaultValues []string `json:"default_values,omitempty"`
	VisibleTags   []string `json:"visible_tags,omitempty"`
}

type secureEmbedViewingPreferences struct {
	HighDensity bool   `json:"high_density,omitempty"`
	Theme       string `json:"theme,omitempty"`
}

type secureEmbedAttributes struct {
	Title                  string                         `json:"title"`
	Status                 string                         `json:"status,omitempty"`
	GlobalTime             *secureEmbedGlobalTime         `json:"global_time,omitempty"`
	GlobalTimeSelectable   *bool                          `json:"global_time_selectable,omitempty"`
	SelectableTemplateVars []secureEmbedTemplateVar       `json:"selectable_template_vars,omitempty"`
	ViewingPreferences     *secureEmbedViewingPreferences `json:"viewing_preferences,omitempty"`
}

type secureEmbedData struct {
	Type       string                `json:"type"`
	Attributes secureEmbedAttributes `json:"attributes"`
}

type secureEmbedRequest struct {
	Data secureEmbedData `json:"data"`
}

// Response types (GET/POST/PATCH return the same shape)
type secureEmbedResponseAttributes struct {
	Title                  string                         `json:"title"`
	Status                 string                         `json:"status"`
	Token                  string                         `json:"token"`
	URL                    string                         `json:"url"`
	Credential             string                         `json:"credential,omitempty"` // only in POST response
	GlobalTime             *secureEmbedGlobalTime         `json:"global_time,omitempty"`
	GlobalTimeSelectable   bool                           `json:"global_time_selectable"`
	SelectableTemplateVars []secureEmbedTemplateVar       `json:"selectable_template_vars,omitempty"`
	ViewingPreferences     *secureEmbedViewingPreferences `json:"viewing_preferences,omitempty"`
}

type secureEmbedResponseData struct {
	Type       string                        `json:"type"`
	Attributes secureEmbedResponseAttributes `json:"attributes"`
}

type secureEmbedResponse struct {
	Data secureEmbedResponseData `json:"data"`
}

// --- Terraform model ---

type secureEmbedTemplateVarModel struct {
	Name          types.String `tfsdk:"name"`
	Prefix        types.String `tfsdk:"prefix"`
	Type          types.String `tfsdk:"type"`
	DefaultValues types.List   `tfsdk:"default_values"`
	VisibleTags   types.List   `tfsdk:"visible_tags"`
}

type secureEmbedModel struct {
	ID                      types.String                  `tfsdk:"id"`
	DashboardID             types.String                  `tfsdk:"dashboard_id"`
	Title                   types.String                  `tfsdk:"title"`
	Status                  types.String                  `tfsdk:"status"`
	GlobalTimeLiveSpan      types.String                  `tfsdk:"global_time_live_span"`
	GlobalTimeSelectable    types.Bool                    `tfsdk:"global_time_selectable"`
	SelectableTemplateVars  []secureEmbedTemplateVarModel `tfsdk:"selectable_template_vars"`
	ViewingPrefsTheme       types.String                  `tfsdk:"viewing_preferences_theme"`
	ViewingPrefsHighDensity types.Bool                    `tfsdk:"viewing_preferences_high_density"`
	// Computed
	Token      types.String `tfsdk:"token"`
	URL        types.String `tfsdk:"url"`
	Credential types.String `tfsdk:"credential"`
}

// --- Resource ---

type dashboardSecureEmbedResource struct {
	Api  *datadog.APIClient
	Auth context.Context
}

func NewDashboardSecureEmbedResource() resource.Resource {
	return &dashboardSecureEmbedResource{}
}

func (r *dashboardSecureEmbedResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.HttpClient
	r.Auth = providerData.Auth
}

func (r *dashboardSecureEmbedResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "dashboard_secure_embed"
}

func (r *dashboardSecureEmbedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Datadog secure embed shared dashboard. " +
			"NOTE: The HMAC `credential` is only returned by the API once on creation and is stored in Terraform state. " +
			"Ensure your state backend uses encryption at rest and has appropriate access controls.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"dashboard_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the dashboard to create a secure embed for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Required:    true,
				Description: "Title of the secure embed share.",
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("active"),
				Description: "Status of the secure embed. Valid values are `active` and `paused`.",
			},
			"global_time_live_span": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1h"),
				Description: "The live span for the global time, e.g. `1h`, `4h`, `1d`, `2d`, `1w`.",
			},
			"global_time_selectable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether viewers can change the global time range.",
			},
			"viewing_preferences_theme": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("system"),
				Description: "Display theme for the embedded dashboard. Valid values are `system`, `dark`, `light`.",
			},
			"viewing_preferences_high_density": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to display the dashboard in high density mode.",
			},
			// Computed outputs
			"token": schema.StringAttribute{
				Computed:    true,
				Description: "The share token for the secure embed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The public URL for the embedded dashboard.",
			},
			"credential": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The HMAC credential granting access to this secure embed. Only available on initial creation; stored in state.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"selectable_template_vars": schema.ListNestedBlock{
				Description: "Template variables that viewers can filter by.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the template variable.",
						},
						"prefix": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "The tag prefix for this template variable.",
						},
						"type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "The type of the template variable.",
						},
						"default_values": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "The default values for this template variable.",
						},
						"visible_tags": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "The visible tag values for this template variable.",
						},
					},
				},
			},
		},
	}
}

func (r *dashboardSecureEmbedResource) apiPath(dashboardID string) string {
	return fmt.Sprintf("/api/v2/dashboard/%s/shared/secure-embed", dashboardID)
}

func (r *dashboardSecureEmbedResource) apiPathWithToken(dashboardID, token string) string {
	return fmt.Sprintf("/api/v2/dashboard/%s/shared/secure-embed/%s", dashboardID, token)
}

// stringListFromTypes extracts []string from a types.List of strings.
func stringListFromTypes(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	elems := l.Elements()
	result := make([]string, 0, len(elems))
	for _, e := range elems {
		if s, ok := e.(types.String); ok {
			result = append(result, s.ValueString())
		}
	}
	return result
}

func (r *dashboardSecureEmbedResource) buildRequest(plan secureEmbedModel, reqType string) secureEmbedRequest {
	globalTimeSelectable := plan.GlobalTimeSelectable.ValueBool()
	highDensity := plan.ViewingPrefsHighDensity.ValueBool()

	var templateVars []secureEmbedTemplateVar
	for _, tv := range plan.SelectableTemplateVars {
		templateVars = append(templateVars, secureEmbedTemplateVar{
			Name:          tv.Name.ValueString(),
			Prefix:        tv.Prefix.ValueString(),
			Type:          tv.Type.ValueString(),
			DefaultValues: stringListFromTypes(tv.DefaultValues),
			VisibleTags:   stringListFromTypes(tv.VisibleTags),
		})
	}

	return secureEmbedRequest{
		Data: secureEmbedData{
			Type: reqType,
			Attributes: secureEmbedAttributes{
				Title:  plan.Title.ValueString(),
				Status: plan.Status.ValueString(),
				GlobalTime: &secureEmbedGlobalTime{
					LiveSpan: plan.GlobalTimeLiveSpan.ValueString(),
				},
				GlobalTimeSelectable:   &globalTimeSelectable,
				SelectableTemplateVars: templateVars,
				ViewingPreferences: &secureEmbedViewingPreferences{
					Theme:       plan.ViewingPrefsTheme.ValueString(),
					HighDensity: highDensity,
				},
			},
		},
	}
}

func (r *dashboardSecureEmbedResource) updateModelFromResponse(ctx context.Context, model *secureEmbedModel, resp secureEmbedResponse) {
	attr := resp.Data.Attributes
	model.Token = types.StringValue(attr.Token)
	model.URL = types.StringValue(attr.URL)
	model.Title = types.StringValue(attr.Title)
	model.Status = types.StringValue(attr.Status)
	model.GlobalTimeSelectable = types.BoolValue(attr.GlobalTimeSelectable)
	if attr.GlobalTime != nil {
		model.GlobalTimeLiveSpan = types.StringValue(attr.GlobalTime.LiveSpan)
	}
	if attr.ViewingPreferences != nil {
		model.ViewingPrefsTheme = types.StringValue(attr.ViewingPreferences.Theme)
		model.ViewingPrefsHighDensity = types.BoolValue(attr.ViewingPreferences.HighDensity)
	}

	tvModels := make([]secureEmbedTemplateVarModel, 0, len(attr.SelectableTemplateVars))
	for _, tv := range attr.SelectableTemplateVars {
		defaultValues, _ := types.ListValueFrom(ctx, types.StringType, tv.DefaultValues)
		visibleTags, _ := types.ListValueFrom(ctx, types.StringType, tv.VisibleTags)
		tvModels = append(tvModels, secureEmbedTemplateVarModel{
			Name:          types.StringValue(tv.Name),
			Prefix:        types.StringValue(tv.Prefix),
			Type:          types.StringValue(tv.Type),
			DefaultValues: defaultValues,
			VisibleTags:   visibleTags,
		})
	}
	model.SelectableTemplateVars = tvModels

	// Only set credential if present (POST response only)
	if attr.Credential != "" {
		model.Credential = types.StringValue(attr.Credential)
	}
	// ID is dashboardID:token for import
	model.ID = types.StringValue(model.DashboardID.ValueString() + ":" + attr.Token)
}

func (r *dashboardSecureEmbedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secureEmbedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := r.buildRequest(plan, "secure_embed_request")
	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, "POST", r.apiPath(plan.DashboardID.ValueString()), &body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating secure embed",
			fmt.Sprintf("API error (status %d): %s", httpResp.StatusCode, err.Error()))
		return
	}

	var apiResp secureEmbedResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing create response", err.Error())
		return
	}

	r.updateModelFromResponse(ctx, &plan, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardSecureEmbedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secureEmbedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, "GET",
		r.apiPathWithToken(state.DashboardID.ValueString(), state.Token.ValueString()), nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading secure embed", err.Error())
		return
	}

	var apiResp secureEmbedResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	r.updateModelFromResponse(ctx, &state, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardSecureEmbedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan secureEmbedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Preserve token and credential from state (not in plan)
	var state secureEmbedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Token = state.Token
	plan.Credential = state.Credential

	body := r.buildRequest(plan, "secure_embed_update_request")
	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, "PATCH",
		r.apiPathWithToken(plan.DashboardID.ValueString(), plan.Token.ValueString()), &body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating secure embed",
			fmt.Sprintf("API error (status %d): %s", httpResp.StatusCode, err.Error()))
		return
	}

	var apiResp secureEmbedResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error parsing update response", err.Error())
		return
	}

	r.updateModelFromResponse(ctx, &plan, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardSecureEmbedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secureEmbedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := utils.SendRequest(r.Auth, r.Api, "DELETE",
		r.apiPathWithToken(state.DashboardID.ValueString(), state.Token.ValueString()), nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError("Error deleting secure embed", err.Error())
		return
	}
}

// ImportState supports `terraform import datadog_dashboard_secure_embed.x <dashboard_id>:<token>`
func (r *dashboardSecureEmbedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in format <dashboard_id>:<token>",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, frameworkPath.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, frameworkPath.Root("dashboard_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, frameworkPath.Root("token"), parts[1])...)
}
