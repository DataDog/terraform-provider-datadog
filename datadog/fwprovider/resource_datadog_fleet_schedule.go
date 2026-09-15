package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	ddvalidators "github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &fleetScheduleResource{}
	_ resource.ResourceWithImportState = &fleetScheduleResource{}
)

type fleetScheduleResource struct {
	Api  fleetSchedulesAPI
	Auth context.Context
}

func NewFleetScheduleResource() resource.Resource {
	return &fleetScheduleResource{}
}

func (r *fleetScheduleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}
	r.Api = providerData.DatadogApiInstances.GetFleetAutomationApiV2()
	r.Auth = providerData.Auth
}

func (r *fleetScheduleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "fleet_schedule"
}

func (r *fleetScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Fleet Automation Agent upgrade schedule. Schedule mutations use Preview API endpoints and require an application key with the `agent_upgrade_write` and `hosts_read` permissions.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Human-readable name for the schedule.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"query": schema.StringAttribute{
				Description: "Datadog host query used to select the Agent upgrade targets.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"status": schema.StringAttribute{
				Description: "Whether the schedule creates deployments. Valid values are `active` and `inactive`. The API default is used when omitted.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.String{stringvalidator.OneOf("active", "inactive")},
			},
			"version_to_latest": schema.Int64Attribute{
				Description: "Number of major Agent versions behind the latest version to target: `0`, `1`, or `2`. The API default is used when omitted.",
				Optional:    true,
				Computed:    true,
				Validators:  []validator.Int64{int64validator.Between(0, 2)},
			},
			"rule": schema.SingleNestedAttribute{
				Description: "Weekly recurrence and maintenance-window configuration for the schedule.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"days_of_week": schema.SetAttribute{
						Description: "Days when the schedule may run. Valid values are `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`, and `Sun`.",
						Required:    true,
						ElementType: types.StringType,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
							setvalidator.ValueStringsAre(stringvalidator.OneOf("Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun")),
						},
					},
					"maintenance_window_duration": schema.Int64Attribute{
						Description: "Duration of the maintenance window in minutes.",
						Required:    true,
						Validators:  []validator.Int64{int64validator.AtLeast(1)},
					},
					"start_maintenance_window": schema.StringAttribute{
						Description: "Start of the maintenance window in 24-hour `HH:MM` format.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^\d{2}:\d{2}$`), "must use HH:MM format"),
							ddvalidators.TimeFormatValidator("15:04"),
						},
					},
					"timezone": schema.StringAttribute{
						Description: "IANA time zone used to interpret the maintenance window, for example `America/New_York` or `UTC`.",
						Required:    true,
						Validators:  []validator.String{ddvalidators.IANATimezoneValidator()},
					},
				},
			},
		},
	}
}

func (r *fleetScheduleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *fleetScheduleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan fleetScheduleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := buildFleetScheduleCreateRequest(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	created, httpResponse, err := r.Api.CreateFleetSchedule(r.Auth, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error creating Fleet Automation schedule"), ""))
		return
	}
	data, ok := created.GetDataOk()
	if !ok || data == nil || data.GetId() == "" {
		response.Diagnostics.AddError("Invalid Fleet Automation schedule response", "The create API returned no schedule ID.")
		return
	}

	plan.ID = types.StringValue(data.GetId())
	if plan.Status.IsUnknown() && data.Attributes.Status != nil {
		plan.Status = types.StringValue(string(*data.Attributes.Status))
	} else if plan.Status.IsUnknown() {
		plan.Status = types.StringNull()
	}
	if plan.VersionToLatest.IsUnknown() && data.Attributes.VersionToLatest != nil {
		plan.VersionToLatest = types.Int64Value(*data.Attributes.VersionToLatest)
	} else if plan.VersionToLatest.IsUnknown() {
		plan.VersionToLatest = types.Int64Null()
	}

	// The Preview create endpoint may apply its defaults instead of explicit
	// optional values. Reconcile those values before the stable refresh so an
	// explicitly inactive schedule never remains active after creation.
	reconciliation, changed := buildFleetScheduleCreateReconciliation(&plan)
	if changed {
		_, reconciliationResponse, reconciliationErr := r.Api.UpdateFleetSchedule(r.Auth, plan.ID.ValueString(), reconciliation)
		if reconciliationErr != nil {
			// A failed reconciliation can leave a schedule active. Best-effort
			// deletion is safer than leaving an untracked schedule behind.
			_, _ = r.Api.DeleteFleetSchedule(r.Auth, plan.ID.ValueString())
			response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(reconciliationErr, reconciliationResponse, "error reconciling Fleet Automation schedule after creation"), ""))
			return
		}
	}

	// Persist the create response first so the schedule remains tracked if the
	// stable read endpoint is temporarily unavailable.
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	r.readAfterMutation(ctx, plan.ID.ValueString(), &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *fleetScheduleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state fleetScheduleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	apiResponse, httpResponse, err := r.Api.GetFleetScheduleV2(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error retrieving Fleet Automation schedule"), ""))
		return
	}
	response.Diagnostics.Append(updateFleetScheduleResourceModel(ctx, &state, apiResponse.GetData())...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *fleetScheduleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state fleetScheduleResourceModel
	var plan fleetScheduleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, changed, diags := buildFleetSchedulePatchRequest(ctx, &state, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if changed {
		_, httpResponse, err := r.Api.UpdateFleetSchedule(r.Auth, state.ID.ValueString(), body)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error updating Fleet Automation schedule"), ""))
			return
		}
	}

	r.readAfterMutation(ctx, state.ID.ValueString(), &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *fleetScheduleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state fleetScheduleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResponse, err := r.Api.DeleteFleetSchedule(r.Auth, state.ID.ValueString())
	if err != nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error deleting Fleet Automation schedule"), ""))
	}
}

func (r *fleetScheduleResource) readAfterMutation(ctx context.Context, id string, model *fleetScheduleResourceModel, diagnostics *diag.Diagnostics) {
	apiResponse, httpResponse, err := r.Api.GetFleetScheduleV2(r.Auth, id)
	if err != nil {
		diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error refreshing Fleet Automation schedule"), ""))
		return
	}
	diagnostics.Append(updateFleetScheduleResourceModel(ctx, model, apiResponse.GetData())...)
}
