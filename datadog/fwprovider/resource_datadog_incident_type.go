package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ID            types.String                    `tfsdk:"id"`
	Name          types.String                    `tfsdk:"name"`
	Description   types.String                    `tfsdk:"description"`
	IsDefault     types.Bool                      `tfsdk:"is_default"`
	Configuration *incidentTypeConfigurationModel `tfsdk:"configuration"`
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

// incidentTypeConfigurationDefault is the object used when the configuration block is omitted,
// mirroring the API's server-side defaults. It keeps the block known (never unknown) so the
// nested-struct model can decode it.
func incidentTypeConfigurationDefault() types.Object {
	return types.ObjectValueMust(incidentTypeConfigurationAttrTypes, map[string]attr.Value{
		"private_incidents":                          types.BoolValue(false),
		"private_incidents_by_default":               types.BoolValue(false),
		"allow_workflows":                            types.BoolValue(true),
		"allow_incident_deletion":                    types.BoolValue(false),
		"editable_timestamps":                        types.BoolValue(false),
		"test_incidents":                             types.BoolValue(true),
		"create_message":                             types.StringValue(""),
		"disable_out_of_the_box_postmortem_template": types.BoolValue(false),
		"slug_source":                                types.StringValue("default"),
	})
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
				Description: "Whether this incident type is the default type.",
				Optional:    true,
				Computed:    true,
			},
			"configuration": schema.SingleNestedAttribute{
				Description: "The incident type's behavior settings. Fields left unset default to their server-side values. Note: this block is applied after creation via a separate update call, since the create endpoint does not accept configuration.",
				Optional:    true,
				Computed:    true,
				Default:     objectdefault.StaticValue(incidentTypeConfigurationDefault()),
				Attributes: map[string]schema.Attribute{
					"private_incidents": schema.BoolAttribute{
						Description: "Whether responders can create private incidents of this type.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"private_incidents_by_default": schema.BoolAttribute{
						Description: "Whether incidents of this type are created as private by default.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"allow_workflows": schema.BoolAttribute{
						Description: "Whether automation workflows can be triggered for incidents of this type.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"allow_incident_deletion": schema.BoolAttribute{
						Description: "Whether incidents of this type can be deleted.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"editable_timestamps": schema.BoolAttribute{
						Description: "Whether responders can edit incident timestamps for incidents of this type.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"test_incidents": schema.BoolAttribute{
						Description: "Whether incidents of this type are treated as test incidents.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"create_message": schema.StringAttribute{
						Description: "An optional message shown to users when they declare an incident of this type.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
					},
					"disable_out_of_the_box_postmortem_template": schema.BoolAttribute{
						Description: "Whether the out-of-the-box postmortem template is disabled for incidents of this type.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"slug_source": schema.StringAttribute{
						Description: "When set to `servicenow`, incidents display the ServiceNow record ID instead of the public ID. If no ServiceNow integration exists, the public ID is displayed.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("default"),
						Validators:  []validator.String{stringvalidator.OneOf("default", "servicenow")},
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
		response.Diagnostics.AddError(
			"Error creating incident type",
			errorMsg,
		)
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
	if state.Configuration != nil {
		patchBody := datadogV2.IncidentTypePatchRequest{
			Data: datadogV2.IncidentTypePatchData{
				Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
				Id:   resp.Data.GetId(),
				Attributes: datadogV2.IncidentTypeUpdateAttributes{
					Configuration: buildIncidentTypeConfiguration(state.Configuration),
				},
			},
		}
		patchResp, patchHTTPResp, err := r.Api.UpdateIncidentType(r.Auth, resp.Data.GetId(), patchBody)
		if err != nil {
			errorMsg := "Could not apply configuration to created incident type, unexpected error: " + err.Error()
			if patchHTTPResp != nil {
				errorMsg += fmt.Sprintf(" (Status: %d)", patchHTTPResp.StatusCode)
			}
			response.Diagnostics.AddError("Error applying incident type configuration", errorMsg)
			return
		}
		if patchHTTPResp.StatusCode != 200 {
			response.Diagnostics.AddError(
				"Error applying incident type configuration",
				fmt.Sprintf("Could not apply configuration, status code: %d", patchHTTPResp.StatusCode),
			)
			return
		}
		r.updateStateFromResponse(&state, &patchResp)
	} else {
		r.updateStateFromResponse(&state, &resp)
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

	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentTypeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state incidentTypeModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
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

	if state.Configuration != nil {
		body.Data.Attributes.Configuration = buildIncidentTypeConfiguration(state.Configuration)
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

	r.updateStateFromResponse(&state, &resp)
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

func (r *incidentTypeResource) updateStateFromResponse(state *incidentTypeModel, resp *datadogV2.IncidentTypeResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	if attributes, ok := resp.Data.GetAttributesOk(); ok {
		state.Name = types.StringValue(attributes.GetName())
		state.Description = types.StringValue(attributes.GetDescription())
		state.IsDefault = types.BoolValue(attributes.GetIsDefault())

		if cfg, ok := attributes.GetConfigurationOk(); ok {
			state.Configuration = &incidentTypeConfigurationModel{
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
		}
	}
}

// buildIncidentTypeConfiguration maps the Terraform configuration model into the API model,
// sending only the fields the user explicitly set so the API's partial-update semantics apply.
func buildIncidentTypeConfiguration(m *incidentTypeConfigurationModel) *datadogV2.IncidentTypeConfiguration {
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
	return cfg
}
