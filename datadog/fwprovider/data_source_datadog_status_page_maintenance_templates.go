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

var _ datasource.DataSourceWithConfigure = &statusPageMaintenanceTemplatesDataSource{}

func NewStatusPageMaintenanceTemplatesDataSource() datasource.DataSource {
	return &statusPageMaintenanceTemplatesDataSource{}
}

type statusPageMaintenanceTemplateListItemModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	MaintenanceTitle types.String `tfsdk:"maintenance_title"`
}

var statusPageMaintenanceTemplateListItemAttrTypes = map[string]attr.Type{
	"id":                types.StringType,
	"name":              types.StringType,
	"maintenance_title": types.StringType,
}

type statusPageMaintenanceTemplatesDataSourceModel struct {
	ID                   types.String                                  `tfsdk:"id"`
	PageID               types.String                                  `tfsdk:"page_id"`
	Name                 types.String                                  `tfsdk:"name"`
	MaintenanceTemplates []*statusPageMaintenanceTemplateListItemModel `tfsdk:"maintenance_templates"`
}

type statusPageMaintenanceTemplatesDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageMaintenanceTemplatesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *statusPageMaintenanceTemplatesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_maintenance_templates"
}

func (d *statusPageMaintenanceTemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to list the maintenance templates of a Datadog status page, for discovery or import. An empty result is not an error.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"page_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the status page whose maintenance templates to list.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter the results to templates with this exact name.",
			},
			"maintenance_templates": schema.ListAttribute{
				Computed:    true,
				Description: "The list of matching maintenance templates.",
				ElementType: types.ObjectType{
					AttrTypes: statusPageMaintenanceTemplateListItemAttrTypes,
				},
			},
		},
	}
}

func (d *statusPageMaintenanceTemplatesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageMaintenanceTemplatesDataSourceModel
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

	resp, httpResp, err := d.Api.ListMaintenanceTemplates(d.Auth, pageID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page not found", state.PageID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing maintenance templates"))
		return
	}

	nameFilter := state.Name.ValueString()
	items := make([]*statusPageMaintenanceTemplateListItemModel, 0)
	for _, template := range resp.GetData() {
		attrs := template.GetAttributes()
		if !state.Name.IsNull() && attrs.GetName() != nameFilter {
			continue
		}
		items = append(items, &statusPageMaintenanceTemplateListItemModel{
			ID:               types.StringValue(template.GetId()),
			Name:             types.StringValue(attrs.GetName()),
			MaintenanceTitle: types.StringValue(attrs.GetMaintenanceTitle()),
		})
	}

	state.ID = state.PageID
	state.MaintenanceTemplates = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
