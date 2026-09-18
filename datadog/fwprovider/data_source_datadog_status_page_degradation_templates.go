package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &statusPageDegradationTemplatesDataSource{}

func NewStatusPageDegradationTemplatesDataSource() datasource.DataSource {
	return &statusPageDegradationTemplatesDataSource{}
}

type statusPageDegradationTemplateListItemModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DegradationTitle types.String `tfsdk:"degradation_title"`
}

var statusPageDegradationTemplateListItemAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"name":              types.StringType,
	"degradation_title": types.StringType,
}

type statusPageDegradationTemplatesDataSourceModel struct {
	ID                   types.String                                  `tfsdk:"id"`
	PageID               types.String                                  `tfsdk:"page_id"`
	Name                 types.String                                  `tfsdk:"name"`
	DegradationTemplates []*statusPageDegradationTemplateListItemModel `tfsdk:"degradation_templates"`
}

type statusPageDegradationTemplatesDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageDegradationTemplatesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *statusPageDegradationTemplatesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_degradation_templates"
}

func (d *statusPageDegradationTemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to list the degradation templates of a Datadog status page, for discovery or import. An empty result is not an error.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"page_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the status page whose degradation templates to list.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter the results to templates with this exact name.",
			},
			"degradation_templates": schema.ListAttribute{
				Computed:    true,
				Description: "The list of matching degradation templates.",
				ElementType: types.ObjectType{
					AttrTypes: statusPageDegradationTemplateListItemAttrTypes,
				},
			},
		},
	}
}

func (d *statusPageDegradationTemplatesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageDegradationTemplatesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := d.Api.ListDegradationTemplates(d.Auth, pageID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page not found", state.PageID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing degradation templates"))
		return
	}

	nameFilter := state.Name.ValueString()
	items := make([]*statusPageDegradationTemplateListItemModel, 0)
	for _, template := range resp.GetData() {
		attrs := template.GetAttributes()
		if !state.Name.IsNull() && attrs.GetName() != nameFilter {
			continue
		}
		items = append(items, &statusPageDegradationTemplateListItemModel{
			ID:               types.StringValue(template.GetId()),
			Name:             types.StringValue(attrs.GetName()),
			DegradationTitle: types.StringValue(attrs.GetDegradationTitle()),
		})
	}

	state.ID = state.PageID
	state.DegradationTemplates = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
