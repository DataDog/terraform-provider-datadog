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

var _ datasource.DataSourceWithConfigure = &statusPageDegradationTemplateDataSource{}

func NewStatusPageDegradationTemplateDataSource() datasource.DataSource {
	return &statusPageDegradationTemplateDataSource{}
}

type statusPageDegradationTemplateDataSourceModel struct {
	ID                 types.String                                          `tfsdk:"id"`
	PageID             types.String                                          `tfsdk:"page_id"`
	Name               types.String                                          `tfsdk:"name"`
	DegradationTitle   types.String                                          `tfsdk:"degradation_title"`
	ComponentsAffected []statusPageDegradationTemplateComponentAffectedModel `tfsdk:"components_affected"`
	Updates            []statusPageDegradationTemplateUpdateModel            `tfsdk:"updates"`
	CreatedAt          types.String                                          `tfsdk:"created_at"`
	ModifiedAt         types.String                                          `tfsdk:"modified_at"`
}

type statusPageDegradationTemplateDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageDegradationTemplateDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *statusPageDegradationTemplateDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_degradation_template"
}

func (d *statusPageDegradationTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog status page degradation template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the degradation template.",
				Required:    true,
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the status page this degradation template belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the degradation template.",
				Computed:    true,
			},
			"degradation_title": schema.StringAttribute{
				Description: "The title used for a degradation created from this template.",
				Computed:    true,
			},
			"components_affected": schema.ListNestedAttribute{
				Description: "The components affected by a degradation created from this template.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the affected component.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the affected component.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The pre-filled status for this component.",
							Computed:    true,
						},
					},
				},
			},
			"updates": schema.ListNestedAttribute{
				Description: "The pre-filled updates for a degradation created from this template.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"message": schema.StringAttribute{
							Description: "The pre-filled message for this update.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The pre-filled degradation status for this update.",
							Computed:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the degradation template was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp when the degradation template was last modified.",
				Computed:    true,
			},
		},
	}
}

func (d *statusPageDegradationTemplateDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageDegradationTemplateDataSourceModel
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

	templateID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse degradation template ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := d.Api.GetDegradationTemplate(d.Auth, pageID, templateID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("degradation template not found", state.ID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving degradation template"))
		return
	}

	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		if name, ok := attributes.GetNameOk(); ok && name != nil {
			state.Name = types.StringValue(*name)
		}
		if degradationTitle, ok := attributes.GetDegradationTitleOk(); ok && degradationTitle != nil {
			state.DegradationTitle = types.StringValue(*degradationTitle)
		} else {
			state.DegradationTitle = types.StringNull()
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.String())
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.String())
		}
		if components, ok := attributes.GetComponentsAffectedOk(); ok && components != nil {
			state.ComponentsAffected = make([]statusPageDegradationTemplateComponentAffectedModel, len(*components))
			for i, component := range *components {
				componentModel := statusPageDegradationTemplateComponentAffectedModel{
					ID:     types.StringValue(component.Id),
					Status: types.StringValue(string(component.Status)),
				}
				if component.Name != nil {
					componentModel.Name = types.StringValue(*component.Name)
				}
				state.ComponentsAffected[i] = componentModel
			}
		} else {
			state.ComponentsAffected = []statusPageDegradationTemplateComponentAffectedModel{}
		}
		if updates, ok := attributes.GetUpdatesOk(); ok && updates != nil {
			state.Updates = make([]statusPageDegradationTemplateUpdateModel, len(*updates))
			for i, update := range *updates {
				updateModel := statusPageDegradationTemplateUpdateModel{
					Status: types.StringValue(string(update.Status)),
				}
				if update.Message != nil {
					updateModel.Message = types.StringValue(*update.Message)
				}
				state.Updates[i] = updateModel
			}
		} else {
			state.Updates = []statusPageDegradationTemplateUpdateModel{}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
