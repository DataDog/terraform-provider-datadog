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

var _ datasource.DataSourceWithConfigure = &statusPageMaintenanceTemplateDataSource{}

func NewStatusPageMaintenanceTemplateDataSource() datasource.DataSource {
	return &statusPageMaintenanceTemplateDataSource{}
}

type statusPageMaintenanceTemplateDataSourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	PageID                types.String   `tfsdk:"page_id"`
	Name                  types.String   `tfsdk:"name"`
	MaintenanceTitle      types.String   `tfsdk:"maintenance_title"`
	ComponentIds          []types.String `tfsdk:"component_ids"`
	ScheduledDescription  types.String   `tfsdk:"scheduled_description"`
	InProgressDescription types.String   `tfsdk:"in_progress_description"`
	CompletedDescription  types.String   `tfsdk:"completed_description"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	ModifiedAt            types.String   `tfsdk:"modified_at"`
}

type statusPageMaintenanceTemplateDataSource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (d *statusPageMaintenanceTemplateDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *statusPageMaintenanceTemplateDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "status_page_maintenance_template"
}

func (d *statusPageMaintenanceTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog status page maintenance template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the maintenance template.",
				Required:    true,
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the status page this maintenance template belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the maintenance template.",
				Computed:    true,
			},
			"maintenance_title": schema.StringAttribute{
				Description: "The title used for a maintenance created from this template.",
				Computed:    true,
			},
			"component_ids": schema.ListAttribute{
				Description: "The IDs of the components affected by a maintenance created from this template.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"scheduled_description": schema.StringAttribute{
				Description: "The pre-filled description shown while the maintenance is scheduled.",
				Computed:    true,
			},
			"in_progress_description": schema.StringAttribute{
				Description: "The pre-filled description shown while the maintenance is in progress.",
				Computed:    true,
			},
			"completed_description": schema.StringAttribute{
				Description: "The pre-filled description shown once the maintenance is completed.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the maintenance template was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp of when the maintenance template was last modified.",
				Computed:    true,
			},
		},
	}
}

func (d *statusPageMaintenanceTemplateDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state statusPageMaintenanceTemplateDataSourceModel
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
			"Could not parse maintenance template ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := d.Api.GetMaintenanceTemplate(d.Auth, pageID, templateID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError("maintenance template not found", state.ID.ValueString())
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving maintenance template"))
		return
	}

	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		if name, ok := attributes.GetNameOk(); ok && name != nil {
			state.Name = types.StringValue(*name)
		}
		if maintenanceTitle, ok := attributes.GetMaintenanceTitleOk(); ok && maintenanceTitle != nil {
			state.MaintenanceTitle = types.StringValue(*maintenanceTitle)
		} else {
			state.MaintenanceTitle = types.StringNull()
		}
		if componentIDs, ok := attributes.GetComponentIdsOk(); ok && componentIDs != nil {
			state.ComponentIds = make([]types.String, len(*componentIDs))
			for i, id := range *componentIDs {
				state.ComponentIds[i] = types.StringValue(id)
			}
		} else {
			state.ComponentIds = []types.String{}
		}
		if scheduledDescription, ok := attributes.GetScheduledDescriptionOk(); ok && scheduledDescription != nil {
			state.ScheduledDescription = types.StringValue(*scheduledDescription)
		} else {
			state.ScheduledDescription = types.StringNull()
		}
		if inProgressDescription, ok := attributes.GetInProgressDescriptionOk(); ok && inProgressDescription != nil {
			state.InProgressDescription = types.StringValue(*inProgressDescription)
		} else {
			state.InProgressDescription = types.StringNull()
		}
		if completedDescription, ok := attributes.GetCompletedDescriptionOk(); ok && completedDescription != nil {
			state.CompletedDescription = types.StringValue(*completedDescription)
		} else {
			state.CompletedDescription = types.StringNull()
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.String())
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.String())
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
