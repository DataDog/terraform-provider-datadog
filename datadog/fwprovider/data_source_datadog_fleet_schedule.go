package fwprovider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &fleetScheduleDataSource{}

type fleetScheduleDataSource struct {
	Api  fleetSchedulesAPI
	Auth context.Context
}

func NewFleetScheduleDataSource() datasource.DataSource {
	return &fleetScheduleDataSource{}
}

func (d *fleetScheduleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}
	d.Api = providerData.DatadogApiInstances.GetFleetAutomationApiV2()
	d.Auth = providerData.Auth
}

func (d *fleetScheduleDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "fleet_schedule"
}

func (d *fleetScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Retrieves a Fleet Automation Agent upgrade schedule by ID. Reading schedules requires an application key with the `hosts_read` permission.",
		Attributes:  fleetScheduleDataSourceAttributes(true),
	}
}

func (d *fleetScheduleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config fleetScheduleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiResponse, httpResponse, err := d.Api.GetFleetScheduleV2(d.Auth, config.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.Diagnostics.AddError("Fleet Automation schedule not found", fmt.Sprintf("No schedule exists with ID %q.", config.ID.ValueString()))
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error retrieving Fleet Automation schedule"), ""))
		return
	}

	state, diags := fleetScheduleDataSourceModelFromAPI(ctx, apiResponse.GetData())
	response.Diagnostics.Append(diags...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func fleetScheduleDataSourceAttributes(idRequired bool) map[string]schema.Attribute {
	id := schema.StringAttribute{
		Description: "Unique identifier of the Fleet Automation schedule.",
		Computed:    !idRequired,
		Required:    idRequired,
	}
	return map[string]schema.Attribute{
		"id": id,
		"name": schema.StringAttribute{
			Description: "Human-readable name of the schedule.",
			Computed:    true,
		},
		"query": schema.StringAttribute{
			Description: "Datadog host query used to select the Agent upgrade targets.",
			Computed:    true,
		},
		"status": schema.StringAttribute{
			Description: "Whether the schedule is `active` or `inactive`.",
			Computed:    true,
		},
		"version_to_latest": schema.Int64Attribute{
			Description: "Number of major Agent versions behind the latest version targeted by the schedule.",
			Computed:    true,
		},
		"rule": schema.SingleNestedAttribute{
			Description: "Recurrence and maintenance-window configuration for the schedule.",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"days_of_week": schema.SetAttribute{
					Description: "Days when the schedule may run.",
					Computed:    true,
					ElementType: types.StringType,
				},
				"interval": schema.Int64Attribute{
					Description: "Interval between schedule runs in weeks.",
					Computed:    true,
				},
				"maintenance_window_duration": schema.Int64Attribute{
					Description: "Duration of the maintenance window in minutes.",
					Computed:    true,
				},
				"start_maintenance_window": schema.StringAttribute{
					Description: "Start of the maintenance window in canonical 24-hour `HH:MM` format.",
					Computed:    true,
				},
				"timezone": schema.StringAttribute{
					Description: "IANA time zone used to interpret the maintenance window.",
					Computed:    true,
				},
			},
		},
		"created_at": schema.StringAttribute{
			Description: "RFC 3339 timestamp when the schedule was created.",
			Computed:    true,
		},
		"created_by": schema.StringAttribute{
			Description: "User handle of the person who created the schedule.",
			Computed:    true,
		},
		"updated_at": schema.StringAttribute{
			Description: "RFC 3339 timestamp when the schedule was last updated.",
			Computed:    true,
		},
		"updated_by": schema.StringAttribute{
			Description: "User handle of the person who last updated the schedule.",
			Computed:    true,
		},
		"is_default": schema.BoolAttribute{
			Description: "Whether this is the organization's default schedule.",
			Computed:    true,
		},
		"next_run": schema.StringAttribute{
			Description: "RFC 3339 timestamp of the next maintenance window, or null when no next run can be computed.",
			Computed:    true,
		},
		"notification_rule": schema.SingleNestedAttribute{
			Description: "Notification configuration attached to the schedule, when available.",
			Computed:    true,
			Attributes: map[string]schema.Attribute{
				"handles": schema.SetAttribute{
					Description: "Notification handles, such as Slack channels or PagerDuty integrations.",
					Computed:    true,
					ElementType: types.StringType,
				},
				"tags": schema.SetAttribute{
					Description: "Tags associated with the notification rule.",
					Computed:    true,
					ElementType: types.StringType,
				},
			},
		},
	}
}
