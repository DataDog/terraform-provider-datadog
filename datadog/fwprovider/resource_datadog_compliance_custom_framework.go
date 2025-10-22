package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var _ resource.Resource = &complianceCustomFrameworkResource{}

type complianceCustomFrameworkResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

// to handle a larger input, requirements and controls had to be lists even though order doesn't matter
// but rules can be sets since requirements and controls are lists (the performance issue happened when all were sets)
type complianceCustomFrameworkModel struct {
	ID           types.String                                 `tfsdk:"id"`
	Version      types.String                                 `tfsdk:"version"`
	Handle       types.String                                 `tfsdk:"handle"`
	Name         types.String                                 `tfsdk:"name"`
	IconURL      types.String                                 `tfsdk:"icon_url"`
	Requirements []complianceCustomFrameworkRequirementsModel `tfsdk:"requirements"`
}

type complianceCustomFrameworkRequirementsModel struct {
	Name     types.String                             `tfsdk:"name"`
	Controls []complianceCustomFrameworkControlsModel `tfsdk:"controls"`
}

type complianceCustomFrameworkControlsModel struct {
	Name    types.String `tfsdk:"name"`
	RulesID types.Set    `tfsdk:"rules_id"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"handle": schema.StringAttribute{
				Description: "The framework handle.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"requirements": schema.ListNestedBlock{
				Description: "The requirements of the framework.",
				Validators: []validator.List{
					validators.DuplicateRequirementControlValidator(),
					listvalidator.IsRequired(),
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
						"controls": schema.ListNestedBlock{
							Description: "The controls of the requirement.",
							Validators: []validator.List{
								listvalidator.IsRequired(),
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
										Description: "The set of rules IDs for the control.",
										ElementType: types.StringType,
										Required:    true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
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
			_, httpResp, getErr := r.Api.GetCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString())
			// if the framework with the same handle and version does not exist, throw an error because
			// only the handle matches which has to be unique
			if httpResp != nil && httpResp.StatusCode == 400 {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Framework with same handle already exists. Currently there is no support for two frameworks with the same handle."))
				return
			}
			if getErr != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(getErr, "error getting existing compliance custom framework"))
				return
			}
			// if the framework with the same handle and version exists, update it
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
	// If the framework does not exist, remove it from terraform state
	// This is to avoid the provider to return an error when the framework is deleted in the UI prior
	if httpResp != nil && httpResp.StatusCode == 400 {
		// 400 could only mean the framework does not exist
		// because terraform would have already validated the framework in the create function
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading compliance custom framework"))
		return
	}
	databaseState := readStateFromDatabase(data, state.Handle.ValueString(), state.Version.ValueString(), &state)
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

	state.ID = types.StringValue(state.Handle.ValueString() + "-" + state.Version.ValueString())

	_, _, err := r.Api.UpdateCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString(), *buildUpdateFrameworkRequest(state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating compliance custom framework"))
		return
	}
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func convertControlToModel(control datadogV2.CustomFrameworkControl) complianceCustomFrameworkControlsModel {
	rulesID := make([]attr.Value, len(control.GetRulesId()))
	for k, v := range control.GetRulesId() {
		rulesID[k] = types.StringValue(v)
	}
	return complianceCustomFrameworkControlsModel{
		Name:    types.StringValue(control.GetName()),
		RulesID: types.SetValueMust(types.StringType, rulesID),
	}
}

func convertRequirementToModel(req datadogV2.CustomFrameworkRequirement) complianceCustomFrameworkRequirementsModel {
	controls := make([]complianceCustomFrameworkControlsModel, len(req.GetControls()))
	for j, control := range req.GetControls() {
		controls[j] = convertControlToModel(control)
	}
	return complianceCustomFrameworkRequirementsModel{
		Name:     types.StringValue(req.GetName()),
		Controls: controls,
	}
}

func readStateFromDatabase(data datadogV2.GetCustomFrameworkResponse, handle string, version string, currentState *complianceCustomFrameworkModel) complianceCustomFrameworkModel {
	var state complianceCustomFrameworkModel
	state.ID = types.StringValue(handle + "-" + version)
	state.Handle = types.StringValue(handle)
	state.Version = types.StringValue(version)
	state.Name = types.StringValue(data.GetData().Attributes.Name)
	if data.GetData().Attributes.IconUrl != nil {
		state.IconURL = types.StringValue(*data.GetData().Attributes.IconUrl)
	}

	apiReqMap := make(map[string]datadogV2.CustomFrameworkRequirement)
	apiControlMap := make(map[string]map[string]datadogV2.CustomFrameworkControl)

	for _, req := range data.GetData().Attributes.Requirements {
		apiReqMap[req.GetName()] = req
		apiControlMap[req.GetName()] = make(map[string]datadogV2.CustomFrameworkControl)
		for _, control := range req.GetControls() {
			apiControlMap[req.GetName()][control.GetName()] = control
		}
	}

	// since the requirements and controls from the API response might be in a different order than the state
	// we need to sort them to match the state so terraform can detect the changes
	// without taking order into account
	sortedRequirements := make([]complianceCustomFrameworkRequirementsModel, 0, len(data.GetData().Attributes.Requirements))

	if currentState != nil {
		for _, currentReq := range currentState.Requirements {
			currentReqName := currentReq.Name.ValueString()
			if apiReq, exists := apiReqMap[currentReqName]; exists {
				sortedControls := make([]complianceCustomFrameworkControlsModel, 0, len(apiReq.GetControls()))

				for _, currentControl := range currentReq.Controls {
					currentControlName := currentControl.Name.ValueString()
					if apiControl, exists := apiControlMap[currentReqName][currentControlName]; exists {
						sortedControls = append(sortedControls, convertControlToModel(apiControl))
						delete(apiControlMap[currentReqName], currentControlName)
					}
				}

				for _, apiControl := range apiControlMap[currentReqName] {
					sortedControls = append(sortedControls, convertControlToModel(apiControl))
				}

				sortedReq := complianceCustomFrameworkRequirementsModel{
					Name:     types.StringValue(apiReq.GetName()),
					Controls: sortedControls,
				}
				sortedRequirements = append(sortedRequirements, sortedReq)
				delete(apiReqMap, currentReqName)
			}
		}
	}

	for _, apiReq := range apiReqMap {
		sortedRequirements = append(sortedRequirements, convertRequirementToModel(apiReq))
	}

	state.Requirements = sortedRequirements
	return state
}

func convertStateRequirementsToFrameworkRequirements(requirements []complianceCustomFrameworkRequirementsModel) []datadogV2.CustomFrameworkRequirement {
	frameworkRequirements := make([]datadogV2.CustomFrameworkRequirement, len(requirements))
	for i, requirement := range requirements {
		controls := make([]datadogV2.CustomFrameworkControl, len(requirement.Controls))
		for j, control := range requirement.Controls {
			rulesID := make([]string, 0)
			for _, v := range control.RulesID.Elements() {
				rulesID = append(rulesID, v.(types.String).ValueString())
			}
			controls[j] = *datadogV2.NewCustomFrameworkControl(control.Name.ValueString(), rulesID)
		}
		frameworkRequirements[i] = *datadogV2.NewCustomFrameworkRequirement(controls, requirement.Name.ValueString())
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
