package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &statusPageDataSource{}

func NewStatusPageDataSource() datasource.DataSource {
	return &statusPageDataSource{}
}

type statusPageDataSourceModel struct {
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

type statusPageDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	d.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	d.Auth = providerData.Auth
}

func (d *statusPageDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page"
}

func (d *statusPageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the status page.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the status page.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the status page.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the status page is published.",
				Computed:    true,
			},
			"domain_prefix": schema.StringAttribute{
				Description: "The subdomain prefix used to build the status page's URL.",
				Computed:    true,
			},
			"custom_domain": schema.StringAttribute{
				Description: "The custom domain configured for the status page, if any.",
				Computed:    true,
			},
			"custom_domain_enabled": schema.BoolAttribute{
				Description: "Whether the custom domain is enabled for the status page.",
				Computed:    true,
			},
			"page_url": schema.StringAttribute{
				Description: "The URL of the status page.",
				Computed:    true,
			},
			"visualization_type": schema.StringAttribute{
				Description: "How component statuses are visualized on the page.",
				Computed:    true,
			},
			"subscriptions_enabled": schema.BoolAttribute{
				Description: "Whether subscriber notifications are enabled for the status page.",
				Computed:    true,
			},
			"slack_subscriptions_enabled": schema.BoolAttribute{
				Description: "Whether Slack subscriber notifications are enabled for the status page.",
				Computed:    true,
			},
			"company_logo": schema.StringAttribute{
				Description: "The company logo displayed on the status page.",
				Computed:    true,
			},
			"favicon": schema.StringAttribute{
				Description: "The favicon displayed for the status page.",
				Computed:    true,
			},
			"email_header_image": schema.StringAttribute{
				Description: "The header image included in subscriber emails.",
				Computed:    true,
			},
			"slack_app_icon": schema.StringAttribute{
				Description: "The icon used for the status page's Slack app integration.",
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
		},
	}
}

func (d *statusPageDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
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

	resp, httpResp, err := d.Api.GetStatusPage(d.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page not found", state.ID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving status page"))
		return
	}

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

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
