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

var _ datasource.DataSourceWithConfigure = &statusPageComponentDataSource{}

func NewStatusPageComponentDataSource() datasource.DataSource {
	return &statusPageComponentDataSource{}
}

type statusPageComponentDataSourceModel struct {
	ID         types.String                  `tfsdk:"id"`
	PageID     types.String                  `tfsdk:"page_id"`
	Name       types.String                  `tfsdk:"name"`
	Type       types.String                  `tfsdk:"type"`
	Position   types.Int64                   `tfsdk:"position"`
	Status     types.String                  `tfsdk:"status"`
	Components []statusPageSubComponentModel `tfsdk:"components"`
	CreatedAt  types.String                  `tfsdk:"created_at"`
	ModifiedAt types.String                  `tfsdk:"modified_at"`
}

type statusPageComponentDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageComponentDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *statusPageComponentDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_component"
}

func (d *statusPageComponentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog status page component.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the component.",
				Required:    true,
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the status page this component belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the component.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the component.",
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "The position of the component on the status page.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the component.",
				Computed:    true,
			},
			"components": schema.ListNestedAttribute{
				Description: "The sub-components of a component of type `group`.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the sub-component.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the sub-component.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the sub-component.",
							Computed:    true,
						},
						"position": schema.Int64Attribute{
							Description: "The position of the sub-component within the group.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the sub-component.",
							Computed:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the component was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp of when the component was last modified.",
				Computed:    true,
			},
		},
	}
}

func (d *statusPageComponentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageComponentDataSourceModel
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

	componentID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page component ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := d.Api.GetComponent(d.Auth, pageID, componentID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("status page component not found", state.ID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving status page component"))
		return
	}

	data := resp.GetData()
	if id, ok := data.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(id.String())
	}

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		if name, ok := attributes.GetNameOk(); ok && name != nil {
			state.Name = types.StringValue(*name)
		}
		state.Type = types.StringValue(string(attributes.GetType()))
		if position, ok := attributes.GetPositionOk(); ok && position != nil {
			state.Position = types.Int64Value(*position)
		}
		if status, ok := attributes.GetStatusOk(); ok && status != nil {
			state.Status = types.StringValue(string(*status))
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.Format("2006-01-02T15:04:05Z"))
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.Format("2006-01-02T15:04:05Z"))
		}
		if components, ok := attributes.GetComponentsOk(); ok && components != nil && len(*components) > 0 {
			state.Components = make([]statusPageSubComponentModel, len(*components))
			for i, sub := range *components {
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
				state.Components[i] = subModel
			}
		} else {
			state.Components = []statusPageSubComponentModel{}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
