package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &statusPageResource{}
	_ resource.ResourceWithImportState = &statusPageResource{}
)

type statusPageResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

type statusPageModel struct {
	ID                        types.String               `tfsdk:"id"`
	Name                      types.String               `tfsdk:"name"`
	Type                      types.String               `tfsdk:"type"`
	Enabled                   types.Bool                 `tfsdk:"enabled"`
	DomainPrefix              types.String               `tfsdk:"domain_prefix"`
	CustomDomain              types.String               `tfsdk:"custom_domain"`
	CustomDomainEnabled       types.Bool                 `tfsdk:"custom_domain_enabled"`
	PageURL                   types.String               `tfsdk:"page_url"`
	VisualizationType         types.String               `tfsdk:"visualization_type"`
	SubscriptionsEnabled      types.Bool                 `tfsdk:"subscriptions_enabled"`
	SlackSubscriptionsEnabled types.Bool                 `tfsdk:"slack_subscriptions_enabled"`
	CompanyLogo               types.String               `tfsdk:"company_logo"`
	Favicon                   types.String               `tfsdk:"favicon"`
	EmailHeaderImage          types.String               `tfsdk:"email_header_image"`
	SlackAppIcon              types.String               `tfsdk:"slack_app_icon"`
	CreatedAt                 types.String               `tfsdk:"created_at"`
	ModifiedAt                types.String               `tfsdk:"modified_at"`
	Components                []statusPageComponentModel `tfsdk:"components"`
}

type statusPageComponentModel struct {
	ID         types.String                  `tfsdk:"id"`
	Name       types.String                  `tfsdk:"name"`
	Type       types.String                  `tfsdk:"type"`
	Position   types.Int64                   `tfsdk:"position"`
	Status     types.String                  `tfsdk:"status"`
	Components []statusPageSubComponentModel `tfsdk:"components"`
}

type statusPageSubComponentModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Position types.Int64  `tfsdk:"position"`
	Status   types.String `tfsdk:"status"`
}

func NewStatusPageResource() resource.Resource {
	return &statusPageResource{}
}

func (r *statusPageResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "status_page"
}

func (r *statusPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog status page resource. This can be used to create and manage Datadog status pages. Component/degradation-template/maintenance-template management is not yet supported by this resource.",
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
			},
			"custom_domain": schema.StringAttribute{
				Description: "The custom domain configured for the status page, if any. Managed via a separate custom-domain flow, not by this resource.",
				Computed:    true,
			},
			"custom_domain_enabled": schema.BoolAttribute{
				Description: "Whether the custom domain is enabled for the status page. Managed via a separate custom-domain flow, not by this resource.",
				Computed:    true,
			},
			"page_url": schema.StringAttribute{
				Description: "The URL of the status page.",
				Computed:    true,
			},
			"subscriptions_enabled": schema.BoolAttribute{
				Description: "Whether subscriber notifications are enabled for the status page.",
				Optional:    true,
				Computed:    true,
			},
			"slack_subscriptions_enabled": schema.BoolAttribute{
				Description: "Whether Slack subscriber notifications are enabled for the status page.",
				Optional:    true,
				Computed:    true,
			},
			"company_logo": schema.StringAttribute{
				Description: "The company logo displayed on the status page.",
				Optional:    true,
				Computed:    true,
			},
			"favicon": schema.StringAttribute{
				Description: "The favicon displayed for the status page.",
				Optional:    true,
				Computed:    true,
			},
			"email_header_image": schema.StringAttribute{
				Description: "The header image included in subscriber emails.",
				Optional:    true,
				Computed:    true,
			},
			"slack_app_icon": schema.StringAttribute{
				Description: "The icon used for the status page's Slack app integration.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the status page was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp when the status page was last modified.",
				Computed:    true,
			},
			"components": schema.ListNestedAttribute{
				Description: "The components (and component groups) displayed on the status page.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the component.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the component.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the component. Valid values are: component, group.",
							Required:    true,
						},
						"position": schema.Int64Attribute{
							Description: "The zero-indexed position of the component.",
							Optional:    true,
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the component. Server-managed; not settable.",
							Computed:    true,
						},
						"components": schema.ListNestedAttribute{
							Description: "If this component is of type `group`, the components nested within the group.",
							Optional:    true,
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "The ID of the component.",
										Computed:    true,
									},
									"name": schema.StringAttribute{
										Description: "The name of the component.",
										Required:    true,
									},
									"type": schema.StringAttribute{
										Description: "The type of the component. Valid values are: component, group.",
										Required:    true,
									},
									"position": schema.Int64Attribute{
										Description: "The zero-indexed position of the component.",
										Optional:    true,
										Computed:    true,
									},
									"status": schema.StringAttribute{
										Description: "The current status of the component. Server-managed; not settable.",
										Computed:    true,
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

func buildStatusPageComponents(components []statusPageComponentModel) []datadogV2.CreateStatusPageRequestDataAttributesComponentsItems {
	if components == nil {
		return nil
	}
	result := make([]datadogV2.CreateStatusPageRequestDataAttributesComponentsItems, len(components))
	for i, component := range components {
		name := component.Name.ValueString()
		componentType := datadogV2.CreateComponentRequestDataAttributesType(component.Type.ValueString())
		item := datadogV2.CreateStatusPageRequestDataAttributesComponentsItems{
			Name: &name,
			Type: &componentType,
		}
		if !component.Position.IsNull() && !component.Position.IsUnknown() {
			item.Position = component.Position.ValueInt64Pointer()
		}
		if len(component.Components) > 0 {
			subItems := make([]datadogV2.CreateStatusPageRequestDataAttributesComponentsItemsComponentsItems, len(component.Components))
			for j, sub := range component.Components {
				subName := sub.Name.ValueString()
				subType := datadogV2.StatusPagesComponentGroupAttributesComponentsItemsType(sub.Type.ValueString())
				subItem := datadogV2.CreateStatusPageRequestDataAttributesComponentsItemsComponentsItems{
					Name: &subName,
					Type: &subType,
				}
				if !sub.Position.IsNull() && !sub.Position.IsUnknown() {
					subItem.Position = sub.Position.ValueInt64Pointer()
				}
				subItems[j] = subItem
			}
			item.Components = subItems
		}
		result[i] = item
	}
	return result
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
		Components:        buildStatusPageComponents(plan.Components),
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

func (r *statusPageResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan statusPageModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
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
			// A status page must be unpublished (see enabled/PublishStatusPage/UnpublishStatusPage,
			// not managed by this resource) before it can be deleted; the API surfaces that as a 409.
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

		if components, ok := attributes.GetComponentsOk(); ok && components != nil {
			state.Components = make([]statusPageComponentModel, len(*components))
			for i, component := range *components {
				componentModel := statusPageComponentModel{}
				if component.Id != nil {
					componentModel.ID = types.StringValue(component.Id.String())
				}
				if component.Name != nil {
					componentModel.Name = types.StringValue(*component.Name)
				}
				if component.Type != nil {
					componentModel.Type = types.StringValue(string(*component.Type))
				}
				if component.Position != nil {
					componentModel.Position = types.Int64Value(*component.Position)
				}
				if component.Status != nil {
					componentModel.Status = types.StringValue(string(*component.Status))
				}
				if len(component.Components) > 0 {
					componentModel.Components = make([]statusPageSubComponentModel, len(component.Components))
					for j, sub := range component.Components {
						subModel := statusPageSubComponentModel{}
						if sub.Id != nil {
							subModel.ID = types.StringValue(sub.Id.String())
						}
						if sub.Name != nil {
							subModel.Name = types.StringValue(*sub.Name)
						}
						if sub.Type != nil {
							subModel.Type = types.StringValue(string(*sub.Type))
						}
						if sub.Position != nil {
							subModel.Position = types.Int64Value(*sub.Position)
						}
						if sub.Status != nil {
							subModel.Status = types.StringValue(string(*sub.Status))
						}
						componentModel.Components[j] = subModel
					}
				}
				state.Components[i] = componentModel
			}
		}
	}
}
