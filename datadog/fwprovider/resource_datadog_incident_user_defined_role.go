package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.ResourceWithConfigure   = &incidentUserDefinedRoleResource{}
	_ resource.ResourceWithImportState = &incidentUserDefinedRoleResource{}
)

type incidentUserDefinedRoleResource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentUserDefinedRoleModel struct {
	ID           types.String                        `tfsdk:"id"`
	Name         types.String                        `tfsdk:"name"`
	Description  types.String                        `tfsdk:"description"`
	IncidentType types.String                        `tfsdk:"incident_type"`
	Policy       *incidentUserDefinedRolePolicyModel `tfsdk:"policy"`
	Created      types.String                        `tfsdk:"created"`
	Modified     types.String                        `tfsdk:"modified"`
}

type incidentUserDefinedRolePolicyModel struct {
	IsSingle types.Bool `tfsdk:"is_single"`
}

func NewIncidentUserDefinedRoleResource() resource.Resource {
	return &incidentUserDefinedRoleResource{}
}

func (r *incidentUserDefinedRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "incident_user_defined_role"
}

func (r *incidentUserDefinedRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog incident user-defined role resource. This can be used to create and manage custom responder roles that are available for a given incident type. **Note**: This resource targets an endpoint that is in preview and is subject to change.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident user-defined role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the user-defined role. Cannot be a reserved name (\"Incident Commander\" or \"Responder\") and must be at most 255 characters.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the user-defined role. At most 1024 characters.",
				Optional:    true,
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this user-defined role is associated with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: "Policy configuration for the user-defined role. Defaults to a multi-assignee policy when omitted.",
				Optional:    true,
				Computed:    true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{"is_single": types.BoolType},
					map[string]attr.Value{"is_single": types.BoolValue(false)},
				)),
				Attributes: map[string]schema.Attribute{
					"is_single": schema.BoolAttribute{
						Description: "Whether this role can only be assigned to one responder at a time.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the user-defined role was created.",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the user-defined role was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *incidentUserDefinedRoleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	r.Auth = providerData.Auth
}

func (r *incidentUserDefinedRoleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan incidentUserDefinedRoleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	incidentTypeID, err := uuid.Parse(plan.IncidentType.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing incident type ID",
			"Could not parse incident type ID: "+err.Error(),
		)
		return
	}

	attributes := datadogV2.IncidentUserDefinedRoleDataAttributesRequest{
		Name: plan.Name.ValueString(),
		Policy: datadogV2.IncidentUserDefinedRolePolicy{
			IsSingle: plan.Policy.IsSingle.ValueBool(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		attributes.Description = *datadog.NewNullableString(plan.Description.ValueStringPointer())
	}

	body := datadogV2.IncidentUserDefinedRoleRequest{
		Data: datadogV2.IncidentUserDefinedRoleDataRequest{
			Type:       datadogV2.INCIDENTUSERDEFINEDROLETYPE_INCIDENT_USER_DEFINED_ROLES,
			Attributes: attributes,
			Relationships: datadogV2.IncidentUserDefinedRoleRelationshipsRequest{
				IncidentType: datadogV2.IncidentUserDefinedRoleIncidentTypeRelationship{
					Data: datadogV2.IncidentUserDefinedRoleIncidentTypeRelationshipData{
						Id:   incidentTypeID,
						Type: "incident_types",
					},
				},
			},
		},
	}

	resp, httpResp, err := r.Api.CreateIncidentUserDefinedRole(r.Auth, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating incident user-defined role",
			fmt.Sprintf("Could not create incident user-defined role, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating incident user-defined role",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state incidentUserDefinedRoleModel
	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentUserDefinedRoleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state incidentUserDefinedRoleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident user-defined role ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetIncidentUserDefinedRole(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident user-defined role",
			"Could not read incident user-defined role ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *incidentUserDefinedRoleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan incidentUserDefinedRoleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident user-defined role ID: "+err.Error(),
		)
		return
	}

	attributes := datadogV2.IncidentUserDefinedRolePatchDataAttributesRequest{
		Name: plan.Name.ValueStringPointer(),
		Policy: &datadogV2.IncidentUserDefinedRolePolicy{
			IsSingle: plan.Policy.IsSingle.ValueBool(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		attributes.Description = *datadog.NewNullableString(plan.Description.ValueStringPointer())
	} else {
		attributes.Description = *datadog.NewNullableString(nil)
	}

	body := datadogV2.IncidentUserDefinedRolePatchRequest{
		Data: datadogV2.IncidentUserDefinedRolePatchDataRequest{
			Id:         id,
			Type:       datadogV2.INCIDENTUSERDEFINEDROLETYPE_INCIDENT_USER_DEFINED_ROLES,
			Attributes: &attributes,
		},
	}

	resp, httpResp, err := r.Api.UpdateIncidentUserDefinedRole(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating incident user-defined role",
			fmt.Sprintf("Could not update incident user-defined role ID %s, unexpected error: %s. HTTP Response: %v", plan.ID.ValueString(), err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating incident user-defined role",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *incidentUserDefinedRoleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state incidentUserDefinedRoleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident user-defined role ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteIncidentUserDefinedRole(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting incident user-defined role",
			"Could not delete incident user-defined role ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *incidentUserDefinedRoleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *incidentUserDefinedRoleResource) updateStateFromResponse(state *incidentUserDefinedRoleModel, resp *datadogV2.IncidentUserDefinedRoleResponse) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		state.Name = types.StringValue(attributes.GetName())

		if description, ok := attributes.GetDescriptionOk(); ok && description != nil {
			state.Description = types.StringValue(*description)
		} else {
			state.Description = types.StringNull()
		}

		if policy, ok := attributes.GetPolicyOk(); ok && policy != nil {
			state.Policy = &incidentUserDefinedRolePolicyModel{
				IsSingle: types.BoolValue(policy.GetIsSingle()),
			}
		}

		if created, ok := attributes.GetCreatedOk(); ok && created != nil {
			state.Created = types.StringValue(created.Format("2006-01-02T15:04:05Z"))
		}

		if modified, ok := attributes.GetModifiedOk(); ok && modified != nil {
			state.Modified = types.StringValue(modified.Format("2006-01-02T15:04:05Z"))
		}
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId().String())
			}
		}
	}
}
