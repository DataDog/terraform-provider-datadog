package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ resource.Resource = &complianceCustomFrameworkResource{}

type complianceCustomFrameworkResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

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

// Custom plan modifier to handle list ordering consistently
type listOrderPlanModifier struct{}

func (m listOrderPlanModifier) Description(ctx context.Context) string {
	return "Preserves the order of list elements as specified in the configuration"
}

func (m listOrderPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Preserves the order of list elements as specified in the configuration"
}

func (m listOrderPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	// If the plan is null, we don't need to do anything
	if req.PlanValue.IsNull() {
		return
	}

	// If the state is null, we don't need to do anything
	if req.StateValue.IsNull() {
		return
	}

	// Get the config value
	configValue := req.ConfigValue

	// If the config is null, we don't need to do anything
	if configValue.IsNull() {
		return
	}

	// Get the state value
	stateValue := req.StateValue

	// Check if the elements are the same (ignoring order)
	configElems := configValue.Elements()
	stateElems := stateValue.Elements()

	// If lengths are different, there's a real change
	if len(configElems) != len(stateElems) {
		resp.PlanValue = configValue
		return
	}

	// Create maps to track elements by their name
	configMap := make(map[string]attr.Value)
	stateMap := make(map[string]attr.Value)

	// Extract names from config elements
	for _, elem := range configElems {
		if obj, ok := elem.(types.Object); ok {
			if name, ok := obj.Attributes()["name"]; ok {
				if strName, ok := name.(types.String); ok {
					configMap[strName.ValueString()] = elem
				}
			}
		}
	}

	// Extract names from state elements
	for _, elem := range stateElems {
		if obj, ok := elem.(types.Object); ok {
			if name, ok := obj.Attributes()["name"]; ok {
				if strName, ok := name.(types.String); ok {
					stateMap[strName.ValueString()] = elem
				}
			}
		}
	}

	// Check if all elements exist in both maps
	hasChanges := false
	for name, configElem := range configMap {
		stateElem, exists := stateMap[name]
		if !exists {
			hasChanges = true
			break
		}
		// Compare the elements (excluding order of nested lists)
		if !compareElements(configElem, stateElem) {
			hasChanges = true
			break
		}
	}

	// If there are real changes, use the config value
	// Otherwise, use the state value to preserve existing order
	if hasChanges {
		resp.PlanValue = configValue
	} else {
		resp.PlanValue = stateValue
	}
}

// Helper function to compare elements while ignoring order of nested lists
func compareElements(config, state attr.Value) bool {
	configObj, ok1 := config.(types.Object)
	stateObj, ok2 := state.(types.Object)
	if !ok1 || !ok2 {
		return config.Equal(state)
	}

	configAttrs := configObj.Attributes()
	stateAttrs := stateObj.Attributes()

	// Compare all attributes except nested lists
	for name, configAttr := range configAttrs {
		stateAttr, exists := stateAttrs[name]
		if !exists {
			return false
		}

		// Skip comparison of nested lists (they're handled by their own plan modifier)
		if _, ok := configAttr.(types.List); ok {
			continue
		}

		if !configAttr.Equal(stateAttr) {
			return false
		}
	}

	return true
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
				PlanModifiers: []planmodifier.List{
					listOrderPlanModifier{},
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
							PlanModifiers: []planmodifier.List{
								listOrderPlanModifier{},
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

	_, _, err := r.Api.UpdateCustomFramework(r.Auth, state.Handle.ValueString(), state.Version.ValueString(), *buildUpdateFrameworkRequest(state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating compliance custom framework"))
		return
	}
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}

func (r *complianceCustomFrameworkResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the plan is null, we don't need to do anything
	if req.Plan.Raw.IsNull() {
		return
	}

	// Let the plan modifiers handle the ordering
	// They will be called automatically for the requirements and controls lists
	return
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

	// Create maps to track requirements and controls by name
	reqMap := make(map[string]datadogV2.CustomFrameworkRequirement)
	ctrlMap := make(map[string]map[string]datadogV2.CustomFrameworkControl)

	// Build maps of requirements and controls from API response
	for _, req := range data.GetData().Attributes.Requirements {
		reqMap[req.GetName()] = req
		ctrlMap[req.GetName()] = make(map[string]datadogV2.CustomFrameworkControl)
		for _, ctrl := range req.GetControls() {
			ctrlMap[req.GetName()][ctrl.GetName()] = ctrl
		}
	}

	// Check if API response has same elements as current state
	if currentState != nil {
		// Check if all requirements and controls match
		hasChanges := false
		stateReqMap := make(map[string]bool)
		stateCtrlMap := make(map[string]map[string]bool)

		// Build maps of current state requirements and controls
		for _, req := range currentState.Requirements {
			reqName := req.Name.ValueString()
			stateReqMap[reqName] = true
			stateCtrlMap[reqName] = make(map[string]bool)
			for _, ctrl := range req.Controls {
				stateCtrlMap[reqName][ctrl.Name.ValueString()] = true
			}
		}

		// Check if API response matches current state
		for reqName := range reqMap {
			if !stateReqMap[reqName] {
				hasChanges = true
				break
			}
			for ctrlName := range ctrlMap[reqName] {
				if !stateCtrlMap[reqName][ctrlName] {
					hasChanges = true
					break
				}
			}
		}

		// If no changes, use current state
		if !hasChanges {
			state.Requirements = currentState.Requirements
			return state
		}
	}

	// If there are changes or no current state, use API order
	state.Requirements = make([]complianceCustomFrameworkRequirementsModel, len(data.GetData().Attributes.Requirements))
	for i, req := range data.GetData().Attributes.Requirements {
		state.Requirements[i] = complianceCustomFrameworkRequirementsModel{
			Name:     types.StringValue(req.GetName()),
			Controls: make([]complianceCustomFrameworkControlsModel, len(req.GetControls())),
		}

		for j, ctrl := range req.GetControls() {
			rulesID := make([]attr.Value, len(ctrl.GetRulesId()))
			for k, v := range ctrl.GetRulesId() {
				rulesID[k] = types.StringValue(v)
			}

			state.Requirements[i].Controls[j] = complianceCustomFrameworkControlsModel{
				Name:    types.StringValue(ctrl.GetName()),
				RulesID: types.SetValueMust(types.StringType, rulesID),
			}
		}
	}

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
