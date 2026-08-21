package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &statusPagesDataSource{}

func NewDatadogStatusPagesDataSource() datasource.DataSource {
	return &statusPagesDataSource{}
}

type statusPageListItemModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	DomainPrefix      types.String `tfsdk:"domain_prefix"`
	Type              types.String `tfsdk:"type"`
	VisualizationType types.String `tfsdk:"visualization_type"`
	PageURL           types.String `tfsdk:"page_url"`
	Enabled           types.Bool   `tfsdk:"enabled"`
}

type statusPagesDataSourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Name        types.String               `tfsdk:"name"`
	StatusPages []*statusPageListItemModel `tfsdk:"status_pages"`
}

type statusPagesDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPagesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		return
	}
	d.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	d.Auth = providerData.Auth
}

func (d *statusPagesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_pages"
}

func (d *statusPagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to list existing Datadog status pages, for discovery or import. An empty result is not an error.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter the results to pages with this exact name.",
			},
			"status_pages": schema.ListAttribute{
				Computed:    true,
				Description: "The list of matching status pages.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                 types.StringType,
						"name":               types.StringType,
						"domain_prefix":      types.StringType,
						"type":               types.StringType,
						"visualization_type": types.StringType,
						"page_url":           types.StringType,
						"enabled":            types.BoolType,
					},
				},
			},
		},
	}
}

func (d *statusPagesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPagesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := d.Api.ListStatusPages(d.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing status pages"))
		return
	}

	nameFilter := state.Name.ValueString()
	items := make([]*statusPageListItemModel, 0)
	for _, page := range resp.GetData() {
		attrs := page.GetAttributes()
		if !state.Name.IsNull() && attrs.GetName() != nameFilter {
			continue
		}
		items = append(items, &statusPageListItemModel{
			ID:                types.StringValue(page.GetId().String()),
			Name:              types.StringValue(attrs.GetName()),
			DomainPrefix:      types.StringValue(attrs.GetDomainPrefix()),
			Type:              types.StringValue(string(attrs.GetType())),
			VisualizationType: types.StringValue(string(attrs.GetVisualizationType())),
			PageURL:           types.StringValue(attrs.GetPageUrl()),
			Enabled:           types.BoolValue(attrs.GetEnabled()),
		})
	}

	state.ID = types.StringValue("status-pages")
	state.StatusPages = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
