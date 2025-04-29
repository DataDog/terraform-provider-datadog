package fwprovider

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ resource.Resource = &customFrameworkResource{}

type customFrameworkResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type customFrameworkModel struct {
	ID           types.String `tfsdk:"id"`
	Description  types.String `tfsdk:"description"`
	Version      types.String `tfsdk:"version"`
	Handle       types.String `tfsdk:"handle"`
	Name         types.String `tfsdk:"name"`
	IconURL      types.String `tfsdk:"icon_url"`
	Requirements types.Set    `tfsdk:"requirements"` // have to define requirements as a set to be unordered
}

func NewCustomFrameworkResource() resource.Resource {
	return &customFrameworkResource{}
}

func (r *customFrameworkResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "custom_framework"
}

func (r *customFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages custom framework in Datadog.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the custom framework resource.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "The framework version.",
				Required:    true,
			},
			"handle": schema.StringAttribute{
				Description: "The framework handle.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The framework name.",
				Required:    true,
			},
			"icon_url": schema.StringAttribute{
				Description: "The URL of the icon representing the framework.",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the framework.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"requirements": schema.SetNestedBlock{
				Description: "The requirements of the framework.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the requirement.",
							Required:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"controls": schema.SetNestedBlock{
							Description: "The controls of the requirement.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of the control.",
										Required:    true,
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

func (r *customFrameworkResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *customFrameworkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state customFrameworkModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.Api.CreateCustomFramework(r.Auth, *buildCreateFrameworkRequest(state))

	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating custom framework"))
		return
	}
	state.ID = types.StringValue(state.Handle.ValueString() + string('-') + state.Version.ValueString())
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *customFrameworkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state customFrameworkModel
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

func (r *customFrameworkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state customFrameworkModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data, _, err := r.Api.GetCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading custom framework"))
		return
	}
	databaseState := readStateFromDatabase(data, state.Handle.ValueString(), state.Version.ValueString())
	diags = response.State.Set(ctx, &databaseState)
	response.Diagnostics.Append(diags...)
}

func (r *customFrameworkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state customFrameworkModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, _, err := r.Api.UpdateCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString(), *buildUpdateFrameworkRequest(state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating custom framework"))
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
func readStateFromDatabase(data datadogV2.GetCustomFrameworkResponse, handle string, version string) customFrameworkModel {
	// Set the state
	var state customFrameworkModel
	state.ID = types.StringValue(handle + "-" + version)
	state.Handle = types.StringValue(handle)
	state.Version = types.StringValue(version)
	state.Name = types.StringValue(data.GetData().Attributes.Name)
	state.Description = types.StringValue(data.GetData().Attributes.Description)
	state.IconURL = types.StringValue(data.GetData().Attributes.IconUrl)

	// Convert requirements to set
	requirements := make([]attr.Value, len(data.GetData().Attributes.Requirements))
	for i, requirement := range data.GetData().Attributes.Requirements {
		// Convert controls to set
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

// ImportState is used to import a resource from an existing framework so we can update it if it exists in the database and not in terraform
func (r *customFrameworkResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Split the ID into handle and version
	// The last hyphen separates handle and version
	lastHyphenIndex := strings.LastIndex(request.ID, "-")
	if lastHyphenIndex == -1 {
		response.Diagnostics.AddError("Invalid import ID", "Import ID must contain a hyphen to separate handle and version")
		return
	}
	handle := request.ID[:lastHyphenIndex]
	version := request.ID[lastHyphenIndex+1:]

	data, _, err := r.Api.GetCustomFramework(r.Auth, handle, version)
	if err != nil {
		response.Diagnostics.AddError("Error importing resource", err.Error())
		return
	}

	state := readStateFromDatabase(data, handle, version)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

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

func buildCreateFrameworkRequest(state customFrameworkModel) *datadogV2.CreateCustomFrameworkRequest {
	createFrameworkRequest := datadogV2.NewCreateCustomFrameworkRequestWithDefaults()
	description := state.Description.ValueString()
	iconURL := state.IconURL.ValueString()
	createFrameworkRequest.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:       state.Handle.ValueString(),
			Name:         state.Name.ValueString(),
			Description:  &description,
			IconUrl:      &iconURL,
			Version:      state.Version.ValueString(),
			Requirements: convertStateRequirementsToFrameworkRequirements(state.Requirements),
		},
	})
	return createFrameworkRequest
}

func buildUpdateFrameworkRequest(state customFrameworkModel) *datadogV2.UpdateCustomFrameworkRequest {
	updateFrameworkRequest := datadogV2.NewUpdateCustomFrameworkRequestWithDefaults()
	description := state.Description.ValueString()
	iconURL := state.IconURL.ValueString()
	updateFrameworkRequest.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:       state.Handle.ValueString(),
			Name:         state.Name.ValueString(),
			Description:  &description,
			IconUrl:      &iconURL,
			Version:      state.Version.ValueString(),
			Requirements: convertStateRequirementsToFrameworkRequirements(state.Requirements),
		},
	})
	return updateFrameworkRequest
}
