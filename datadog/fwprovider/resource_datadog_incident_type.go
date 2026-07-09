package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.ResourceWithConfigure   = &incidentTypeResource{}
	_ resource.ResourceWithImportState = &incidentTypeResource{}
)

type incidentTypeResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentTypeModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	Configuration types.Object `tfsdk:"configuration"`
}

type incidentTypeConfigurationModel struct {
	PrivateIncidents                     types.Bool   `tfsdk:"private_incidents"`
	PrivateIncidentsByDefault            types.Bool   `tfsdk:"private_incidents_by_default"`
	AllowWorkflows                       types.Bool   `tfsdk:"allow_workflows"`
	AllowIncidentDeletion                types.Bool   `tfsdk:"allow_incident_deletion"`
	EditableTimestamps                   types.Bool   `tfsdk:"editable_timestamps"`
	TestIncidents                        types.Bool   `tfsdk:"test_incidents"`
	CreateMessage                        types.String `tfsdk:"create_message"`
	DisableOutOfTheBoxPostmortemTemplate types.Bool   `tfsdk:"disable_out_of_the_box_postmortem_template"`
	SlugSource                           types.String `tfsdk:"slug_source"`
}

var incidentTypeConfigurationAttrTypes = map[string]attr.Type{
	"private_incidents":                          types.BoolType,
	"private_incidents_by_default":               types.BoolType,
	"allow_workflows":                            types.BoolType,
	"allow_incident_deletion":                    types.BoolType,
	"editable_timestamps":                        types.BoolType,
	"test_incidents":                             types.BoolType,
	"create_message":                             types.StringType,
	"disable_out_of_the_box_postmortem_template": types.BoolType,
	"slug_source":                                types.StringType,
}

func NewIncidentTypeResource() resource.Resource {
	return &incidentTypeResource{}
}

func (r *incidentTypeResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	r.Auth = providerData.Auth
}

func (r *incidentTypeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "incident_type"
}

func (r *incidentTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog incident type resource. This can be used to create and manage Datadog incident types.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident type.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the incident type. Must be between 1 and 50 characters.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the incident type. The description can have a maximum of 512 characters.",
				Optional:    true,
			},
			"is_default": schema.BoolAttribute{
				Description:   "Whether this incident type is the default type.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"configuration": schema.SingleNestedAttribute{
				Description: "The incident type's behavior settings. Any field left unset takes its server-side default. This block is applied in a separate call after the incident type is created.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"private_incidents": schema.BoolAttribute{
						Description:   "Whether responders can create private incidents of this type. Defaults to `false`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"private_incidents_by_default": schema.BoolAttribute{
						Description:   "Whether the private toggle is enabled by default in the incident creation modal for this type. Defaults to `false`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"allow_workflows": schema.BoolAttribute{
						Description:   "Whether users can manually run a workflow from an incident of this type. Defaults to `true`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"allow_incident_deletion": schema.BoolAttribute{
						Description:   "Whether incidents of this type can be deleted. Defaults to `false`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"editable_timestamps": schema.BoolAttribute{
						Description:   "Whether responders can edit incident timestamps for incidents of this type. Defaults to `false`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"test_incidents": schema.BoolAttribute{
						Description:   "Whether test incidents of this type can be created. Defaults to `true`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"create_message": schema.StringAttribute{
						Description:   "An optional message shown to users when they declare an incident of this type. Defaults to an empty string.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"disable_out_of_the_box_postmortem_template": schema.BoolAttribute{
						Description:   "When enabled, incidents of this type do not use Datadog's out-of-the-box postmortem template. Defaults to `false`.",
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
					},
					"slug_source": schema.StringAttribute{
						Description:   "The source used to derive the incident slug. When set to `servicenow`, incidents display the ServiceNow record ID instead of the public ID. If no ServiceNow integration exists, the public ID is displayed. Defaults to `default`.",
						Optional:      true,
						Computed:      true,
						Validators:    []validator.String{stringvalidator.OneOf("default", "servicenow")},
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
			},
		},
	}
}

func (r *incidentTypeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Read the raw config (not the plan) to tell whether the user actually declared a
	// configuration block. Because the attribute is Computed, the plan always carries a
	// value, so we must only issue the follow-up PATCH when the user explicitly set it.
	var config incidentTypeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.IncidentTypeCreateRequest{
		Data: datadogV2.IncidentTypeCreateData{
			Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			Attributes: datadogV2.IncidentTypeAttributes{
				Name: state.Name.ValueString(),
			},
		},
	}

	if !state.Description.IsNull() {
		body.Data.Attributes.SetDescription(state.Description.ValueString())
	}

	if !state.IsDefault.IsNull() {
		body.Data.Attributes.SetIsDefault(state.IsDefault.ValueBool())
	}

	resp, httpResp, err := r.Api.CreateIncidentType(r.Auth, body)
	if err != nil {
		errorMsg := "Could not create incident type, unexpected error: " + err.Error()
		if httpResp != nil {
			errorMsg += fmt.Sprintf(" (Status: %d)", httpResp.StatusCode)
		}
		response.Diagnostics.AddError("Error creating incident type", errorMsg)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident type",
			fmt.Sprintf("Could not create incident type, status code: %d", httpResp.StatusCode),
		)
		return
	}

	// The create endpoint ignores configuration, so when the user specifies it we apply it
	// with a follow-up PATCH against the newly created type.
	if !config.Configuration.IsNull() {
		cfg, diags := buildIncidentTypeConfiguration(ctx, state.Configuration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		patchBody := datadogV2.IncidentTypePatchRequest{
			Data: datadogV2.IncidentTypePatchData{
				Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
				Id:   resp.Data.GetId(),
				Attributes: datadogV2.IncidentTypeUpdateAttributes{
					Configuration: cfg,
				},
			},
		}
		patchResp, patchHTTPResp, err := r.Api.UpdateIncidentType(r.Auth, resp.Data.GetId(), patchBody)
		if err != nil {
			errorMsg := "Could not apply configuration to created incident type, unexpected error: " + err.Error()
			if patchHTTPResp != nil {
				errorMsg += fmt.Sprintf(" (Status: %d)", patchHTTPResp.StatusCode)
			}
			// The type exists but its configuration could not be applied. Since we never
			// persist it to state, roll back the create to avoid orphaning the type.
			r.cleanupOrphanedIncidentType(ctx, resp.Data.GetId(), &response.Diagnostics)
			response.Diagnostics.AddError("Error applying incident type configuration", errorMsg)
			return
		}
		if patchHTTPResp.StatusCode != 200 {
			r.cleanupOrphanedIncidentType(ctx, resp.Data.GetId(), &response.Diagnostics)
			response.Diagnostics.AddError(
				"Error applying incident type configuration",
				fmt.Sprintf("Could not apply configuration, status code: %d", patchHTTPResp.StatusCode),
			)
			return
		}
		response.Diagnostics.Append(r.updateStateFromResponse(ctx, &state, &patchResp)...)
	} else {
		response.Diagnostics.Append(r.updateStateFromResponse(ctx, &state, &resp)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetIncidentType(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident type",
			"Could not read incident type, unexpected error: "+err.Error(),
		)
		return
	}

	response.Diagnostics.Append(r.updateStateFromResponse(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Only send configuration when the user actually declared the block (the plan always
	// carries one because the attribute is Computed).
	var config incidentTypeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.IncidentTypePatchRequest{
		Data: datadogV2.IncidentTypePatchData{
			Type:       datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			Id:         state.ID.ValueString(),
			Attributes: datadogV2.IncidentTypeUpdateAttributes{},
		},
	}

	if !state.Name.IsNull() {
		body.Data.Attributes.SetName(state.Name.ValueString())
	}

	if !state.Description.IsNull() {
		body.Data.Attributes.SetDescription(state.Description.ValueString())
	}

	if !state.IsDefault.IsNull() {
		body.Data.Attributes.SetIsDefault(state.IsDefault.ValueBool())
	}

	if !config.Configuration.IsNull() {
		cfg, diags := buildIncidentTypeConfiguration(ctx, state.Configuration)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		body.Data.Attributes.Configuration = cfg
	}

	resp, httpResp, err := r.Api.UpdateIncidentType(r.Auth, state.ID.ValueString(), body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident type",
			"Could not update incident type, unexpected error: "+err.Error(),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident type",
			fmt.Sprintf("Could not update incident type, status code: %d", httpResp.StatusCode),
		)
		return
	}

	response.Diagnostics.Append(r.updateStateFromResponse(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteIncidentType(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting incident type",
			fmt.Sprintf("Could not delete incident type, unexpected error: %s (Status: %d)", err.Error(), httpResp.StatusCode),
		)
		return
	}
}

func (r *incidentTypeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// cleanupOrphanedIncidentType best-effort deletes an incident type that was created but whose
// follow-up configuration PATCH failed. Because the resource is never written to state in that
// path, leaving the type in place would orphan it. A failure to delete is surfaced as a warning
// so the original error remains the primary diagnostic.
func (r *incidentTypeResource) cleanupOrphanedIncidentType(ctx context.Context, id string, diags *diag.Diagnostics) {
	if httpResp, err := r.Api.DeleteIncidentType(r.Auth, id); err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		diags.AddWarning(
			"Could not clean up incident type after failed configuration",
			fmt.Sprintf("Incident type %s was created but its configuration could not be applied, and the automatic cleanup delete also failed: %s. You may need to delete it manually.", id, err.Error()),
		)
	}
}

func (r *incidentTypeResource) updateStateFromResponse(ctx context.Context, state *incidentTypeModel, resp *datadogV2.IncidentTypeResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	state.ID = types.StringValue(resp.Data.GetId())

	if attributes, ok := resp.Data.GetAttributesOk(); ok {
		state.Name = types.StringValue(attributes.GetName())
		state.Description = types.StringValue(attributes.GetDescription())
		state.IsDefault = types.BoolValue(attributes.GetIsDefault())

		if cfg, ok := attributes.GetConfigurationOk(); ok {
			m := incidentTypeConfigurationModel{
				PrivateIncidents:                     types.BoolValue(cfg.GetPrivateIncidents()),
				PrivateIncidentsByDefault:            types.BoolValue(cfg.GetPrivateIncidentsByDefault()),
				AllowWorkflows:                       types.BoolValue(cfg.GetAllowWorkflows()),
				AllowIncidentDeletion:                types.BoolValue(cfg.GetAllowIncidentDeletion()),
				EditableTimestamps:                   types.BoolValue(cfg.GetEditableTimestamps()),
				TestIncidents:                        types.BoolValue(cfg.GetTestIncidents()),
				CreateMessage:                        types.StringValue(cfg.GetCreateMessage()),
				DisableOutOfTheBoxPostmortemTemplate: types.BoolValue(cfg.GetDisableOutOfTheBoxPostmortemTemplate()),
				SlugSource:                           types.StringValue(string(cfg.GetSlugSource())),
			}
			obj, d := types.ObjectValueFrom(ctx, incidentTypeConfigurationAttrTypes, m)
			diags.Append(d...)
			state.Configuration = obj
		} else {
			state.Configuration = types.ObjectNull(incidentTypeConfigurationAttrTypes)
		}
	}
	return diags
}

// buildIncidentTypeConfiguration maps the Terraform configuration object into the API model,
// sending only the fields the user explicitly set so the API's partial-update semantics apply.
func buildIncidentTypeConfiguration(ctx context.Context, obj types.Object) (*datadogV2.IncidentTypeConfiguration, diag.Diagnostics) {
	var m incidentTypeConfigurationModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	cfg := datadogV2.NewIncidentTypeConfiguration()
	if !m.PrivateIncidents.IsNull() && !m.PrivateIncidents.IsUnknown() {
		cfg.SetPrivateIncidents(m.PrivateIncidents.ValueBool())
	}
	if !m.PrivateIncidentsByDefault.IsNull() && !m.PrivateIncidentsByDefault.IsUnknown() {
		cfg.SetPrivateIncidentsByDefault(m.PrivateIncidentsByDefault.ValueBool())
	}
	if !m.AllowWorkflows.IsNull() && !m.AllowWorkflows.IsUnknown() {
		cfg.SetAllowWorkflows(m.AllowWorkflows.ValueBool())
	}
	if !m.AllowIncidentDeletion.IsNull() && !m.AllowIncidentDeletion.IsUnknown() {
		cfg.SetAllowIncidentDeletion(m.AllowIncidentDeletion.ValueBool())
	}
	if !m.EditableTimestamps.IsNull() && !m.EditableTimestamps.IsUnknown() {
		cfg.SetEditableTimestamps(m.EditableTimestamps.ValueBool())
	}
	if !m.TestIncidents.IsNull() && !m.TestIncidents.IsUnknown() {
		cfg.SetTestIncidents(m.TestIncidents.ValueBool())
	}
	if !m.CreateMessage.IsNull() && !m.CreateMessage.IsUnknown() {
		cfg.SetCreateMessage(m.CreateMessage.ValueString())
	}
	if !m.DisableOutOfTheBoxPostmortemTemplate.IsNull() && !m.DisableOutOfTheBoxPostmortemTemplate.IsUnknown() {
		cfg.SetDisableOutOfTheBoxPostmortemTemplate(m.DisableOutOfTheBoxPostmortemTemplate.ValueBool())
	}
	if !m.SlugSource.IsNull() && !m.SlugSource.IsUnknown() {
		cfg.SetSlugSource(datadogV2.IncidentTypeSlugSource(m.SlugSource.ValueString()))
	}
	return cfg, diags
}
