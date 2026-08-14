package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &statusPageDataSource{}

func NewDatadogStatusPageDataSource() datasource.DataSource {
	return &statusPageDataSource{}
}

type statusPageDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	DomainPrefix      types.String `tfsdk:"domain_prefix"`
	Type              types.String `tfsdk:"type"`
	VisualizationType types.String `tfsdk:"visualization_type"`
	PageURL           types.String `tfsdk:"page_url"`
	Enabled           types.Bool   `tfsdk:"enabled"`
}

type statusPageDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
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
				Required:    true,
				Description: "The ID of the status page.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the status page.",
			},
			"domain_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "The domain prefix for the status page URL.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of status page (`public` or `internal`).",
			},
			"visualization_type": schema.StringAttribute{
				Computed:    true,
				Description: "How component status is visualized.",
			},
			"page_url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the status page.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the status page is published/enabled.",
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
	pid, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status page id"))
		return
	}
	resp, httpResp, err := d.Api.GetStatusPage(d.Auth, pid)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page not found", state.ID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving status page"))
		return
	}
	data := resp.GetData()
	attrs := data.GetAttributes()
	state.Name = types.StringValue(attrs.GetName())
	state.DomainPrefix = types.StringValue(attrs.GetDomainPrefix())
	state.Type = types.StringValue(string(attrs.GetType()))
	state.VisualizationType = types.StringValue(string(attrs.GetVisualizationType()))
	state.PageURL = types.StringValue(attrs.GetPageUrl())
	state.Enabled = types.BoolValue(attrs.GetEnabled())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
