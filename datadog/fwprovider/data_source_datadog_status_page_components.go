package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &statusPageComponentsDataSource{}

func NewDatadogStatusPageComponentsDataSource() datasource.DataSource {
	return &statusPageComponentsDataSource{}
}

type statusPageComponentListItemModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Position types.Int64  `tfsdk:"position"`
	GroupID  types.String `tfsdk:"group_id"`
}

type statusPageComponentsDataSourceModel struct {
	ID           types.String                        `tfsdk:"id"`
	StatusPageID types.String                        `tfsdk:"status_page_id"`
	Name         types.String                        `tfsdk:"name"`
	Type         types.String                        `tfsdk:"type"`
	Components   []*statusPageComponentListItemModel `tfsdk:"components"`
}

type statusPageComponentsDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageComponentsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		return
	}
	d.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	d.Auth = providerData.Auth
}

func (d *statusPageComponentsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_components"
}

func (d *statusPageComponentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to list the components and groups of a Datadog status page, for discovery or import. An empty result is not an error.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"status_page_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the status page whose components to list.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter the results to components with this exact name.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter the results by type (`component` or `group`).",
			},
			"components": schema.ListAttribute{
				Computed:    true,
				Description: "The list of matching components and groups.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":       types.StringType,
						"name":     types.StringType,
						"type":     types.StringType,
						"position": types.Int64Type,
						"group_id": types.StringType,
					},
				},
			},
		},
	}
}

func (d *statusPageComponentsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageComponentsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, err := uuid.Parse(state.StatusPageID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status_page_id"))
		return
	}
	resp, httpResp, err := d.Api.GetStatusPage(d.Auth, pid)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page not found", state.StatusPageID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving status page"))
		return
	}

	data := resp.GetData()
	attrs := data.GetAttributes()
	items := make([]*statusPageComponentListItemModel, 0)
	for _, top := range attrs.GetComponents() {
		items = d.appendItem(&state, items, top.GetId().String(), top.GetName(), string(top.GetType()), top.GetPosition(), types.StringNull())
		for _, child := range top.GetComponents() {
			items = d.appendItem(&state, items, child.GetId().String(), child.GetName(), string(child.GetType()), child.GetPosition(), types.StringValue(top.GetId().String()))
		}
	}

	state.ID = state.StatusPageID
	state.Components = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *statusPageComponentsDataSource) appendItem(state *statusPageComponentsDataSourceModel, items []*statusPageComponentListItemModel, id, name, typ string, position int64, groupID types.String) []*statusPageComponentListItemModel {
	if !state.Name.IsNull() && name != state.Name.ValueString() {
		return items
	}
	if !state.Type.IsNull() && typ != state.Type.ValueString() {
		return items
	}
	return append(items, &statusPageComponentListItemModel{
		ID:       types.StringValue(id),
		Name:     types.StringValue(name),
		Type:     types.StringValue(typ),
		Position: types.Int64Value(position),
		GroupID:  groupID,
	})
}
