package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
	ID                   types.String                        `tfsdk:"id"`
	Name                 types.String                        `tfsdk:"name"`
	Description          types.String                        `tfsdk:"description"`
	IncidentType         types.String                        `tfsdk:"incident_type_id"`
	Policy               *incidentUserDefinedRolePolicyModel `tfsdk:"policy"`
	Created              types.String                        `tfsdk:"created"`
	Modified             types.String                        `tfsdk:"modified"`
	CreatedByUserID      types.String                        `tfsdk:"created_by_user_id"`
	LastModifiedByUserID types.String                        `tfsdk:"last_modified_by_user_id"`
}

type incidentUserDefinedRolePolicyModel struct {
	IsSingle types.Bool `tfsdk:"is_single"`
}

func NewIncidentUserDefinedRoleResource() resource.Resource {
	return &incidentUserDefinedRoleResource{}
}

func (r *incidentUserDefinedRoleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "incident_user_defined_role"
}

func (r *incidentUserDefinedRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog incident user-defined role resource. This can be used to create and manage custom incident roles that responders can be assigned to, scoped to an incident type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the user-defined role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the user-defined role. Must be between 1 and 255 characters and cannot be a reserved name (\"Incident Commander\" or \"Responder\").",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the user-defined role. Can have a maximum of 1024 characters.",
				Optional:    true,
			},
			"incident_type_id": schema.StringAttribute{
				Description: "The ID of the incident type this role is scoped to. Changing this forces a new resource to be created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: "The policy governing how the role can be assigned.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"is_single": schema.BoolAttribute{
						Description: "Whether at most one responder can hold this role at a time on a given incident.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the user-defined role was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the user-defined role was last modified.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by_user_id": schema.StringAttribute{
				Description: "The ID of the user who created the role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_modified_by_user_id": schema.StringAttribute{
				Description: "The ID of the user who last modified the role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	attributes := datadogV2.IncidentUserDefinedRoleCreateAttributes{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() {
		attributes.SetDescription(plan.Description.ValueString())
	}
	if plan.Policy != nil && !plan.Policy.IsSingle.IsNull() && !plan.Policy.IsSingle.IsUnknown() {
		attributes.Policy = &datadogV2.IncidentUserDefinedRolePolicy{
			IsSingle: plan.Policy.IsSingle.ValueBool(),
		}
	}

	relationships := &datadogV2.IncidentUserDefinedRoleCreateDataRelationships{
		IncidentType: &datadogV2.RelationshipToIncidentType{
			Data: datadogV2.RelationshipToIncidentTypeData{
				Id:   plan.IncidentType.ValueString(),
				Type: datadogV2.INCIDENTTYPETYPE_INCIDENT_TYPES,
			},
		},
	}

	body := datadogV2.CreateIncidentUserDefinedRoleRequest{
		Data: datadogV2.IncidentUserDefinedRoleCreateData{
			Type:          datadogV2.INCIDENTUSERDEFINEDROLETYPE_INCIDENT_USER_DEFINED_ROLES,
			Attributes:    attributes,
			Relationships: relationships,
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

	// The PATCH surface only supports name, description, and policy. The
	// incident type relationship is immutable (RequiresReplace handles that).
	attributes := datadogV2.IncidentUserDefinedRoleUpdateAttributes{}
	attributes.SetName(plan.Name.ValueString())
	if !plan.Description.IsNull() {
		attributes.SetDescription(plan.Description.ValueString())
	}
	if plan.Policy != nil && !plan.Policy.IsSingle.IsNull() && !plan.Policy.IsSingle.IsUnknown() {
		attributes.Policy = &datadogV2.IncidentUserDefinedRolePolicy{
			IsSingle: plan.Policy.IsSingle.ValueBool(),
		}
	}

	body := datadogV2.UpdateIncidentUserDefinedRoleRequest{
		Data: datadogV2.IncidentUserDefinedRoleUpdateData{
			Id:         id,
			Type:       datadogV2.INCIDENTUSERDEFINEDROLETYPE_INCIDENT_USER_DEFINED_ROLES,
			Attributes: attributes,
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

func (r *incidentUserDefinedRoleResource) updateStateFromResponse(state *incidentUserDefinedRoleModel, resp *datadogV2.IncidentUserDefinedRole) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		state.Name = types.StringValue(attributes.GetName())

		if description, descriptionOk := attributes.GetDescriptionOk(); descriptionOk && description != nil {
			state.Description = types.StringValue(*description)
		}

		if policy, policyOk := attributes.GetPolicyOk(); policyOk && policy != nil {
			state.Policy = &incidentUserDefinedRolePolicyModel{
				IsSingle: types.BoolValue(policy.GetIsSingle()),
			}
		}

		if created, createdOk := attributes.GetCreatedOk(); createdOk && created != nil {
			state.Created = types.StringValue(created.Format("2006-01-02T15:04:05Z"))
		}

		if modified, modifiedOk := attributes.GetModifiedOk(); modifiedOk && modified != nil {
			state.Modified = types.StringValue(modified.Format("2006-01-02T15:04:05Z"))
		}
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId())
			}
		}

		if createdByUser, ok := relationships.GetCreatedByUserOk(); ok && createdByUser != nil {
			if userData, ok := createdByUser.GetDataOk(); ok && userData != nil {
				state.CreatedByUserID = types.StringValue(userData.GetId())
			}
		}

		if lastModifiedByUser, ok := relationships.GetLastModifiedByUserOk(); ok && lastModifiedByUser != nil {
			if userData, ok := lastModifiedByUser.GetDataOk(); ok && userData != nil {
				state.LastModifiedByUserID = types.StringValue(userData.GetId())
			}
		}
	}
}
