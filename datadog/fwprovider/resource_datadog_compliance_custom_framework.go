package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var _ resource.Resource = &complianceCustomFrameworkResource{}

type complianceCustomFrameworkResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type complianceCustomFrameworkModel struct {
	ID           types.String `tfsdk:"id"`
	Version      types.String `tfsdk:"version"`
	Handle       types.String `tfsdk:"handle"`
	Name         types.String `tfsdk:"name"`
	IconURL      types.String `tfsdk:"icon_url"`
	Requirements types.Set    `tfsdk:"requirements"` // have to define requirements as a set to be unordered
}

func NewComplianceCustomFrameworkResource() resource.Resource {
	return &complianceCustomFrameworkResource{}
}

func (r *complianceCustomFrameworkResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "compliance_custom_framework"
}

func (r *complianceCustomFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Compliance Custom Framework resource, which is used to create and manage compliance custom frameworks.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the compliance custom framework resource.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The framework version.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Required: true,
			},
			"handle": schema.StringAttribute{
				Description: "The framework handle.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Required: true,
			},
			"name": schema.StringAttribute{
				Description: "The framework name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Required: true,
			},
			"icon_url": schema.StringAttribute{
				Description: "The URL of the icon representing the framework",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"requirements": schema.SetNestedBlock{
				Description: "The requirements of the framework.",
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					validators.RequirementNameValidator(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the requirement.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"controls": schema.SetNestedBlock{
							Description: "The controls of the requirement.",
							Validators: []validator.Set{
								setvalidator.IsRequired(),
								validators.ControlNameValidator(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of the control.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
									},
									"rules_id": schema.SetAttribute{
										Description: "The list of rules IDs for the control.",
										ElementType: types.StringType,
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *complianceCustomFrameworkResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *complianceCustomFrameworkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state complianceCustomFrameworkModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, httpResp, err := r.Api.CreateCustomFramework(r.Auth, *buildCreateFrameworkRequest(state))
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 409 { // if framework already exists, try to update it with the new state
			_, _, updateErr := r.Api.UpdateCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString(), *buildUpdateFrameworkRequest(state))
			if updateErr != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(updateErr, "error updating existing compliance custom framework"))
				return
			}
		} else {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating compliance custom framework"))
			return
		}
	}
	state.ID = types.StringValue(state.Handle.ValueString() + string('-') + state.Version.ValueString())
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *complianceCustomFrameworkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state complianceCustomFrameworkModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	_, _, err := r.Api.DeleteCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting framework"))
		return
	}
}

func (r *complianceCustomFrameworkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state complianceCustomFrameworkModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data, httpResp, err := r.Api.GetCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString())
	if err != nil && httpResp != nil && httpResp.StatusCode != 400 {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading compliance custom framework"))
		return
	}
	// If the framework does not exist, remove it from terraform state
	// This is to avoid the provider to return an error when the framework is deleted in the UI prior
	if err != nil && httpResp != nil && httpResp.StatusCode == 400 {
		// 400 could only mean the framework does not exist
		// because terraform would have already validated the framework in the create function
		response.State.RemoveResource(ctx)
		return
	}
	databaseState := readStateFromDatabase(data, state.Handle.ValueString(), state.Version.ValueString())
	diags = response.State.Set(ctx, &databaseState)
	response.Diagnostics.Append(diags...)
}

func (r *complianceCustomFrameworkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state complianceCustomFrameworkModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.Api.UpdateCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString(), *buildUpdateFrameworkRequest(state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating compliance custom framework"))
		return
	}
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func setControl(name string, rulesID []attr.Value) types.Object {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"name":     types.StringType,
			"rules_id": types.SetType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"name":     types.StringValue(name),
			"rules_id": types.SetValueMust(types.StringType, rulesID),
		},
	)
}

func setRequirement(name string, controls []attr.Value) types.Object {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"name": types.StringType,
			"controls": types.SetType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":     types.StringType,
				"rules_id": types.SetType{ElemType: types.StringType},
			}}},
		},
		map[string]attr.Value{
			"name": types.StringValue(name),
			"controls": types.SetValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":     types.StringType,
				"rules_id": types.SetType{ElemType: types.StringType},
			}}, controls),
		},
	)
}

func setRequirements(requirements []attr.Value) types.Set {
	return types.SetValueMust(
		types.ObjectType{AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"controls": types.SetType{ElemType: types.ObjectType{AttrTypes: map[string]attr.Type{
				"name":     types.StringType,
				"rules_id": types.SetType{ElemType: types.StringType},
			}}},
		}},
		requirements,
	)
}
func readStateFromDatabase(data datadogV2.GetCustomFrameworkResponse, handle string, version string) complianceCustomFrameworkModel {
	var state complianceCustomFrameworkModel
	state.ID = types.StringValue(handle + "-" + version)
	state.Handle = types.StringValue(handle)
	state.Version = types.StringValue(version)
	state.Name = types.StringValue(data.GetData().Attributes.Name)
	if data.GetData().Attributes.IconUrl != nil {
		state.IconURL = types.StringValue(*data.GetData().Attributes.IconUrl)
	}

	requirements := make([]attr.Value, len(data.GetData().Attributes.Requirements))
	for i, requirement := range data.GetData().Attributes.Requirements {
		controls := make([]attr.Value, len(requirement.Controls))
		for j, control := range requirement.Controls {
			rulesID := make([]attr.Value, len(control.RulesId))
			for k, ruleID := range control.RulesId {
				rulesID[k] = types.StringValue(ruleID)
			}
			controls[j] = setControl(control.Name, rulesID)
		}
		requirements[i] = setRequirement(requirement.Name, controls)
	}
	state.Requirements = setRequirements(requirements)
	return state
}

// using sets for requirements in state to be unordered
func convertStateRequirementsToFrameworkRequirements(requirements types.Set) []datadogV2.CustomFrameworkRequirement {
	frameworkRequirements := make([]datadogV2.CustomFrameworkRequirement, len(requirements.Elements()))
	for i, requirement := range requirements.Elements() {
		requirementState := requirement.(types.Object)
		controls := make([]datadogV2.CustomFrameworkControl, len(requirementState.Attributes()["controls"].(types.Set).Elements()))
		for j, control := range requirementState.Attributes()["controls"].(types.Set).Elements() {
			controlState := control.(types.Object)
			rulesID := make([]string, len(controlState.Attributes()["rules_id"].(types.Set).Elements()))
			for k, ruleID := range controlState.Attributes()["rules_id"].(types.Set).Elements() {
				rulesID[k] = ruleID.(types.String).ValueString()
			}
			controls[j] = *datadogV2.NewCustomFrameworkControl(controlState.Attributes()["name"].(types.String).ValueString(), rulesID)
		}
		frameworkRequirements[i] = *datadogV2.NewCustomFrameworkRequirement(controls, requirementState.Attributes()["name"].(types.String).ValueString())
	}
	return frameworkRequirements
}

func buildCreateFrameworkRequest(state complianceCustomFrameworkModel) *datadogV2.CreateCustomFrameworkRequest {
	var iconURL *string
	if !state.IconURL.IsNull() && !state.IconURL.IsUnknown() {
		iconURLStr := state.IconURL.ValueString()
		iconURL = &iconURLStr
	}
	createFrameworkRequest := datadogV2.NewCreateCustomFrameworkRequestWithDefaults()
	createFrameworkRequest.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:       state.Handle.ValueString(),
			Name:         state.Name.ValueString(),
			IconUrl:      iconURL,
			Version:      state.Version.ValueString(),
			Requirements: convertStateRequirementsToFrameworkRequirements(state.Requirements),
		},
	})
	return createFrameworkRequest
}

func buildUpdateFrameworkRequest(state complianceCustomFrameworkModel) *datadogV2.UpdateCustomFrameworkRequest {
	var iconURL *string
	if !state.IconURL.IsNull() && !state.IconURL.IsUnknown() {
		iconURLStr := state.IconURL.ValueString()
		iconURL = &iconURLStr
	}
	updateFrameworkRequest := datadogV2.NewUpdateCustomFrameworkRequestWithDefaults()
	updateFrameworkRequest.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:       state.Handle.ValueString(),
			Name:         state.Name.ValueString(),
			Version:      state.Version.ValueString(),
			IconUrl:      iconURL,
			Requirements: convertStateRequirementsToFrameworkRequirements(state.Requirements),
		},
	})
	return updateFrameworkRequest
}
