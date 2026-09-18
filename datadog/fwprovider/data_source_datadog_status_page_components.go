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

var _ datasource.DataSourceWithConfigure = &statusPageComponentsDataSource{}

func NewStatusPageComponentsDataSource() datasource.DataSource {
	return &statusPageComponentsDataSource{}
}

type statusPageComponentListItemModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Position types.Int64  `tfsdk:"position"`
	GroupID  types.String `tfsdk:"group_id"`
}

var statusPageComponentListItemAttrTypes = map[string]attr.Type{
	"id":       types.StringType,
	"name":     types.StringType,
	"type":     types.StringType,
	"position": types.Int64Type,
	"group_id": types.StringType,
}

type statusPageComponentsDataSourceModel struct {
	ID         types.String                        `tfsdk:"id"`
	PageID     types.String                        `tfsdk:"page_id"`
	Name       types.String                        `tfsdk:"name"`
	Type       types.String                        `tfsdk:"type"`
	Components []*statusPageComponentListItemModel `tfsdk:"components"`
}

type statusPageComponentsDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageComponentsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *statusPageComponentsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_components"
}

func (d *statusPageComponentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to list the components and groups of a Datadog status page, for discovery or import. An empty result is not an error.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"page_id": schema.StringAttribute{
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
					AttrTypes: statusPageComponentListItemAttrTypes,
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

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := d.Api.ListComponents(d.Auth, pageID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page not found", state.PageID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing status page components"))
		return
	}

	nameFilter := state.Name.ValueString()
	typeFilter := state.Type.ValueString()
	items := make([]*statusPageComponentListItemModel, 0)
	for _, component := range resp.GetData() {
		attrs := component.GetAttributes()
		if !state.Name.IsNull() && attrs.GetName() != nameFilter {
			continue
		}
		if !state.Type.IsNull() && string(attrs.GetType()) != typeFilter {
			continue
		}
		item := &statusPageComponentListItemModel{
			ID:       types.StringValue(component.GetId().String()),
			Name:     types.StringValue(attrs.GetName()),
			Type:     types.StringValue(string(attrs.GetType())),
			Position: types.Int64Value(attrs.GetPosition()),
			GroupID:  types.StringNull(),
		}
		items = append(items, item)

		for _, sub := range attrs.GetComponents() {
			if !state.Name.IsNull() && sub.GetName() != nameFilter {
				continue
			}
			if !state.Type.IsNull() && string(sub.GetType()) != typeFilter {
				continue
			}
			items = append(items, &statusPageComponentListItemModel{
				ID:       types.StringValue(sub.GetId().String()),
				Name:     types.StringValue(sub.GetName()),
				Type:     types.StringValue(string(sub.GetType())),
				Position: types.Int64Value(sub.GetPosition()),
				GroupID:  types.StringValue(component.GetId().String()),
			})
		}
	}

	state.ID = state.PageID
	state.Components = items
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
