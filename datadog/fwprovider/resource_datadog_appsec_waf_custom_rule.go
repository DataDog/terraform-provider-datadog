package fwprovider

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &appsecWafCustomRuleResource{}
	_ resource.ResourceWithImportState = &appsecWafCustomRuleResource{}
)

type appsecWafCustomRuleResource struct {
	Api  *datadogV2.ApplicationSecurityApi
	Auth context.Context
}

type appsecWafCustomRuleModel struct {
	ID         types.String       `tfsdk:"id"`
	Blocking   types.Bool         `tfsdk:"blocking"`
	Enabled    types.Bool         `tfsdk:"enabled"`
	Name       types.String       `tfsdk:"name"`
	PathGlob   types.String       `tfsdk:"path_glob"`
	Conditions []*conditionsModel `tfsdk:"conditions"`
	Scope      []*scopeModel      `tfsdk:"scope"`
	Action     *actionModel       `tfsdk:"action"`
	Tags       types.Map          `tfsdk:"tags"`
}

type conditionsModel struct {
	Operator   types.String     `tfsdk:"operator"`
	Parameters *parametersModel `tfsdk:"parameters"`
}
type parametersModel struct {
	Data    types.String   `tfsdk:"data"`
	Regex   types.String   `tfsdk:"regex"`
	Value   types.String   `tfsdk:"value"`
	List    types.List     `tfsdk:"list"`
	Inputs  []*inputsModel `tfsdk:"inputs"`
	Options *optionsModel  `tfsdk:"options"`
}
type inputsModel struct {
	Address types.String `tfsdk:"address"`
	KeyPath types.List   `tfsdk:"key_path"`
}

type optionsModel struct {
	CaseSensitive types.Bool  `tfsdk:"case_sensitive"`
	MinLength     types.Int64 `tfsdk:"min_length"`
}

type actionModel struct {
	Action     types.String           `tfsdk:"action"`
	Parameters *actionParametersModel `tfsdk:"parameters"`
}
type actionParametersModel struct {
	Location   types.String `tfsdk:"location"`
	StatusCode types.Int64  `tfsdk:"status_code"`
}

func NewAppsecWafCustomRuleResource() resource.Resource {
	return &appsecWafCustomRuleResource{}
}

func (r *appsecWafCustomRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetApplicationSecurityApiV2()
	r.Auth = providerData.Auth
}

func (r *appsecWafCustomRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "appsec_waf_custom_rule"
}

func (r *appsecWafCustomRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AppsecWafCustomRule resource. This can be used to create and manage Datadog appsec_waf_custom_rule.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the WAF custom rule is enabled.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The Name of the WAF custom rule.",
			},
			"path_glob": schema.StringAttribute{
				Optional:    true,
				Description: "The path glob for the WAF custom rule.",
			},
			"blocking": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the WAF custom rule will block the request.",
			},
			"tags": schema.MapAttribute{
				Required:    true,
				Description: "Tags associated with the WAF custom rule. `category` and `type` tags are required. Supported categories include `business_logic`, `attack_attempt` and `security_response`.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"conditions": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"operator": schema.StringAttribute{
							Optional:    true,
							Description: "Operator to use for the WAF Condition.",
						},
					},
					Blocks: map[string]schema.Block{
						"parameters": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"data": schema.StringAttribute{
									Optional:    true,
									Description: "Identifier of a list of data from the denylist. Can only be used as substitution from the list parameter.",
								},
								"regex": schema.StringAttribute{
									Optional:    true,
									Description: "Regex to use with the condition. Only used with match_regex and !match_regex operator.",
								},
								"value": schema.StringAttribute{
									Optional:    true,
									Description: "Store the captured value in the specified tag name. Only used with the capture_data operator.",
								},
								"list": schema.ListAttribute{
									Optional:    true,
									Description: "List of value to use with the condition. Only used with the phrase_match, !phrase_match, exact_match and !exact_match operator.",
									ElementType: types.StringType,
								},
							},
							Blocks: map[string]schema.Block{
								"inputs": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"address": schema.StringAttribute{
												Optional:    true,
												Description: "Input from the request on which the condition should apply.",
											},
											"key_path": schema.ListAttribute{
												Optional:    true,
												Description: "Specific path for the input.",
												ElementType: types.StringType,
											},
										},
									},
								},
								"options": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{
										"case_sensitive": schema.BoolAttribute{
											Optional:    true,
											Description: "Evaluate the value as case sensitive.",
										},
										"min_length": schema.Int64Attribute{
											Optional:    true,
											Description: "Only evaluate this condition if the value has a minimum amount of characters.",
										},
									},
								},
							},
						},
					},
				},
			},
			"scope": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"env": schema.StringAttribute{
							Optional:    true,
							Description: "The environment scope for the WAF custom rule.",
						},
						"service": schema.StringAttribute{
							Optional:    true,
							Description: "The service scope for the WAF custom rule.",
						},
					},
				},
			},
			"action": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						Optional:    true,
						Description: "Override the default action to take when the WAF custom rule would block.",
					},
				},
				Blocks: map[string]schema.Block{
					"parameters": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"location": schema.StringAttribute{
								Optional:    true,
								Description: "The location to redirect to when the WAF custom rule triggers.",
							},
							"status_code": schema.Int64Attribute{
								Optional:    true,
								Description: "The status code to return when the WAF custom rule triggers.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *appsecWafCustomRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *appsecWafCustomRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state appsecWafCustomRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetApplicationSecurityWafCustomRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AppsecWafCustomRule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appsecWafCustomRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state appsecWafCustomRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAppsecWafCustomRuleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	appsecWafConcurrencyMutex.Lock()
	defer appsecWafConcurrencyMutex.Unlock()

	var resp datadogV2.ApplicationSecurityWafCustomRuleResponse
	var err error
	err = retry.RetryContext(ctx, appsecRetryOnConflictTimeout, func() *retry.RetryError {
		var httpResp *http.Response
		resp, httpResp, err = r.Api.CreateApplicationSecurityWafCustomRule(r.Auth, *body)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AppsecWafCustomRule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appsecWafCustomRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state appsecWafCustomRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildAppsecWafCustomRuleUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	appsecWafConcurrencyMutex.Lock()
	defer appsecWafConcurrencyMutex.Unlock()

	var resp datadogV2.ApplicationSecurityWafCustomRuleResponse
	var err error
	err = retry.RetryContext(ctx, appsecRetryOnConflictTimeout, func() *retry.RetryError {
		var httpResp *http.Response
		resp, httpResp, err = r.Api.UpdateApplicationSecurityWafCustomRule(r.Auth, id, *body)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AppsecWafCustomRule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *appsecWafCustomRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state appsecWafCustomRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	appsecWafConcurrencyMutex.Lock()
	defer appsecWafConcurrencyMutex.Unlock()

	var httpResp *http.Response
	var err error
	err = retry.RetryContext(ctx, appsecRetryOnConflictTimeout, func() *retry.RetryError {
		httpResp, err = r.Api.DeleteApplicationSecurityWafCustomRule(r.Auth, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting appsec_waf_custom_rule"))
		return
	}
}

func (r *appsecWafCustomRuleResource) updateState(ctx context.Context, state *appsecWafCustomRuleModel, resp *datadogV2.ApplicationSecurityWafCustomRuleResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if blocking, ok := attributes.GetBlockingOk(); ok {
		state.Blocking = types.BoolValue(*blocking)
	}

	if enabled, ok := attributes.GetEnabledOk(); ok {
		state.Enabled = types.BoolValue(*enabled)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if pathGlob, ok := attributes.GetPathGlobOk(); ok {
		state.PathGlob = types.StringValue(*pathGlob)
	}

	if conditions, ok := attributes.GetConditionsOk(); ok && len(*conditions) > 0 {
		state.Conditions = []*conditionsModel{}
		for _, conditionsDd := range *conditions {
			conditionsTfItem := conditionsModel{}

			if operator, ok := conditionsDd.GetOperatorOk(); ok {
				conditionsTfItem.Operator = types.StringValue(string(*operator))
			}
			if parameters, ok := conditionsDd.GetParametersOk(); ok {

				parametersTf := parametersModel{}
				if data, ok := parameters.GetDataOk(); ok {
					parametersTf.Data = types.StringValue(*data)
				}
				if inputs, ok := parameters.GetInputsOk(); ok && len(*inputs) > 0 {
					parametersTf.Inputs = []*inputsModel{}
					for _, inputsDd := range *inputs {
						inputsTfItem := inputsModel{}

						if address, ok := inputsDd.GetAddressOk(); ok {
							inputsTfItem.Address = types.StringValue(string(*address))
						}
						inputsTfItem.KeyPath, _ = types.ListValueFrom(ctx, types.StringType, inputsDd.GetKeyPath())
						parametersTf.Inputs = append(parametersTf.Inputs, &inputsTfItem)
					}
				}
				parametersTf.List, _ = types.ListValueFrom(ctx, types.StringType, parameters.GetList())
				if regex, ok := parameters.GetRegexOk(); ok {
					parametersTf.Regex = types.StringValue(*regex)
				}
				if value, ok := parameters.GetValueOk(); ok {
					parametersTf.Value = types.StringValue(*value)
				}
				if options, ok := parameters.GetOptionsOk(); ok {
					optionsTf := optionsModel{}
					if caseSensitive, ok := options.GetCaseSensitiveOk(); ok {
						optionsTf.CaseSensitive = types.BoolValue(*caseSensitive)
					}
					if minLength, ok := options.GetMinLengthOk(); ok {
						optionsTf.MinLength = types.Int64Value(*minLength)
					}
					parametersTf.Options = &optionsTf
				}
				conditionsTfItem.Parameters = &parametersTf
			}
			state.Conditions = append(state.Conditions, &conditionsTfItem)
		}
	}

	if scope, ok := attributes.GetScopeOk(); ok && len(*scope) > 0 {
		state.Scope = []*scopeModel{}
		for _, scopeDd := range *scope {
			scopeTfItem := scopeModel{}

			if env, ok := scopeDd.GetEnvOk(); ok {
				scopeTfItem.Env = types.StringValue(*env)
			}
			if service, ok := scopeDd.GetServiceOk(); ok {
				scopeTfItem.Service = types.StringValue(*service)
			}

			state.Scope = append(state.Scope, &scopeTfItem)
		}
	}

	if action, ok := attributes.GetActionOk(); ok {

		actionTf := actionModel{}
		if action, ok := action.GetActionOk(); ok {
			actionTf.Action = types.StringValue(string(*action))
		}
		if parameters, ok := action.GetParametersOk(); ok {

			parametersTf := actionParametersModel{}
			if location, ok := parameters.GetLocationOk(); ok {
				parametersTf.Location = types.StringValue(*location)
			}
			if statusCode, ok := parameters.GetStatusCodeOk(); ok {
				parametersTf.StatusCode = types.Int64Value(*statusCode)
			}

			actionTf.Parameters = &parametersTf
		}

		state.Action = &actionTf
	}

	tagsTf := map[string]string{}
	tags := attributes.GetTags()
	for k, v := range tags.AdditionalProperties {
		tagsTf[k] = v
	}
	if category, ok := tags.GetCategoryOk(); ok {
		tagsTf["category"] = string(*category)
	}
	if typeVar, ok := tags.GetTypeOk(); ok {
		tagsTf["type"] = *typeVar
	}
	state.Tags, _ = types.MapValueFrom(ctx, types.StringType, tagsTf)
}

func (r *appsecWafCustomRuleResource) buildAppsecWafCustomRuleRequestBody(ctx context.Context, state *appsecWafCustomRuleModel) (*datadogV2.ApplicationSecurityWafCustomRuleCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationSecurityWafCustomRuleCreateAttributesWithDefaults()

	if !state.Blocking.IsNull() {
		attributes.SetBlocking(state.Blocking.ValueBool())
	}
	if !state.Enabled.IsNull() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}
	if !state.PathGlob.IsNull() {
		attributes.SetPathGlob(state.PathGlob.ValueString())
	}

	var conditions []datadogV2.ApplicationSecurityWafCustomRuleCondition
	for _, conditionsTFItem := range state.Conditions {
		conditionsDDItem := datadogV2.NewApplicationSecurityWafCustomRuleConditionWithDefaults()

		if !conditionsTFItem.Operator.IsNull() {
			conditionsDDItem.SetOperator(datadogV2.ApplicationSecurityWafCustomRuleConditionOperator(conditionsTFItem.Operator.ValueString()))
		}

		if conditionsTFItem.Parameters != nil {
			var parameters datadogV2.ApplicationSecurityWafCustomRuleConditionParameters

			if !conditionsTFItem.Parameters.Data.IsNull() {
				parameters.SetData(conditionsTFItem.Parameters.Data.ValueString())
			}
			if !conditionsTFItem.Parameters.Regex.IsNull() {
				parameters.SetRegex(conditionsTFItem.Parameters.Regex.ValueString())
			}
			if !conditionsTFItem.Parameters.Value.IsNull() {
				parameters.SetValue(conditionsTFItem.Parameters.Value.ValueString())
			}

			if !conditionsTFItem.Parameters.List.IsNull() {
				var list []string
				diags.Append(conditionsTFItem.Parameters.List.ElementsAs(ctx, &list, false)...)
				parameters.SetList(list)
			}

			if conditionsTFItem.Parameters.Inputs != nil {
				var inputs []datadogV2.ApplicationSecurityWafCustomRuleConditionInput
				for _, inputsTFItem := range conditionsTFItem.Parameters.Inputs {
					inputsDDItem := datadogV2.NewApplicationSecurityWafCustomRuleConditionInputWithDefaults()

					if !inputsTFItem.Address.IsNull() {
						inputsDDItem.SetAddress(datadogV2.ApplicationSecurityWafCustomRuleConditionInputAddress(inputsTFItem.Address.ValueString()))
					}

					if !inputsTFItem.KeyPath.IsNull() {
						var keyPath []string
						diags.Append(inputsTFItem.KeyPath.ElementsAs(ctx, &keyPath, false)...)
						inputsDDItem.SetKeyPath(keyPath)
					}

					inputs = append(inputs, *inputsDDItem)
				}
				parameters.SetInputs(inputs)
			}

			if conditionsTFItem.Parameters.Options != nil {
				var options datadogV2.ApplicationSecurityWafCustomRuleConditionOptions

				if !conditionsTFItem.Parameters.Options.CaseSensitive.IsNull() {
					options.SetCaseSensitive(conditionsTFItem.Parameters.Options.CaseSensitive.ValueBool())
				}
				if !conditionsTFItem.Parameters.Options.MinLength.IsNull() {
					options.SetMinLength(conditionsTFItem.Parameters.Options.MinLength.ValueInt64())
				}
				parameters.Options = &options
			}

			conditionsDDItem.SetParameters(parameters)
		}

		conditions = append(conditions, *conditionsDDItem)
	}
	attributes.SetConditions(conditions)

	if state.Scope != nil {
		var scope []datadogV2.ApplicationSecurityWafCustomRuleScope
		for _, scopeTFItem := range state.Scope {
			scopeDDItem := datadogV2.NewApplicationSecurityWafCustomRuleScopeWithDefaults()

			if !scopeTFItem.Env.IsNull() {
				scopeDDItem.SetEnv(scopeTFItem.Env.ValueString())
			}
			if !scopeTFItem.Service.IsNull() {
				scopeDDItem.SetService(scopeTFItem.Service.ValueString())
			}

			scope = append(scope, *scopeDDItem)
		}
		attributes.SetScope(scope)
	}

	if state.Action != nil {
		var action datadogV2.ApplicationSecurityWafCustomRuleAction

		if !state.Action.Action.IsNull() {
			action.SetAction(datadogV2.ApplicationSecurityWafCustomRuleActionAction(state.Action.Action.ValueString()))
		}

		if state.Action.Parameters != nil {
			var parameters datadogV2.ApplicationSecurityWafCustomRuleActionParameters

			if !state.Action.Parameters.Location.IsNull() {
				parameters.SetLocation(state.Action.Parameters.Location.ValueString())
			}
			if !state.Action.Parameters.StatusCode.IsNull() {
				parameters.SetStatusCode(state.Action.Parameters.StatusCode.ValueInt64())
			}
			action.Parameters = &parameters
		}
		attributes.Action = &action
	}

	if !state.Tags.IsNull() {
		var tags datadogV2.ApplicationSecurityWafCustomRuleTags
		diags.Append(state.Tags.ElementsAs(ctx, &tags.AdditionalProperties, false)...)
		attributes.SetTags(tags)
	}

	req := datadogV2.NewApplicationSecurityWafCustomRuleCreateRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationSecurityWafCustomRuleCreateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *appsecWafCustomRuleResource) buildAppsecWafCustomRuleUpdateRequestBody(ctx context.Context, state *appsecWafCustomRuleModel) (*datadogV2.ApplicationSecurityWafCustomRuleUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewApplicationSecurityWafCustomRuleUpdateAttributesWithDefaults()

	if !state.Blocking.IsNull() {
		attributes.SetBlocking(state.Blocking.ValueBool())
	}
	if !state.Enabled.IsNull() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}
	if !state.Name.IsNull() {
		attributes.SetName(state.Name.ValueString())
	}
	if !state.PathGlob.IsNull() {
		attributes.SetPathGlob(state.PathGlob.ValueString())
	}

	if state.Conditions != nil {
		var conditions []datadogV2.ApplicationSecurityWafCustomRuleCondition
		for _, conditionsTFItem := range state.Conditions {
			conditionsDDItem := datadogV2.NewApplicationSecurityWafCustomRuleConditionWithDefaults()

			if !conditionsTFItem.Operator.IsNull() {
				conditionsDDItem.SetOperator(datadogV2.ApplicationSecurityWafCustomRuleConditionOperator(conditionsTFItem.Operator.ValueString()))
			}

			if conditionsTFItem.Parameters != nil {
				var parameters datadogV2.ApplicationSecurityWafCustomRuleConditionParameters

				if !conditionsTFItem.Parameters.Data.IsNull() {
					parameters.SetData(conditionsTFItem.Parameters.Data.ValueString())
				}
				if !conditionsTFItem.Parameters.Regex.IsNull() {
					parameters.SetRegex(conditionsTFItem.Parameters.Regex.ValueString())
				}
				if !conditionsTFItem.Parameters.Value.IsNull() {
					parameters.SetValue(conditionsTFItem.Parameters.Value.ValueString())
				}

				if !conditionsTFItem.Parameters.List.IsNull() {
					var list []string
					diags.Append(conditionsTFItem.Parameters.List.ElementsAs(ctx, &list, false)...)
					parameters.SetList(list)
				}

				if conditionsTFItem.Parameters.Inputs != nil {
					var inputs []datadogV2.ApplicationSecurityWafCustomRuleConditionInput
					for _, inputsTFItem := range conditionsTFItem.Parameters.Inputs {
						inputsDDItem := datadogV2.NewApplicationSecurityWafCustomRuleConditionInputWithDefaults()

						if !inputsTFItem.Address.IsNull() {
							inputsDDItem.SetAddress(datadogV2.ApplicationSecurityWafCustomRuleConditionInputAddress(inputsTFItem.Address.ValueString()))
						}

						if !inputsTFItem.KeyPath.IsNull() {
							var keyPath []string
							diags.Append(inputsTFItem.KeyPath.ElementsAs(ctx, &keyPath, false)...)
							inputsDDItem.SetKeyPath(keyPath)
						}

						inputs = append(inputs, *inputsDDItem)
					}
					parameters.SetInputs(inputs)
				}

				if conditionsTFItem.Parameters.Options != nil {
					var options datadogV2.ApplicationSecurityWafCustomRuleConditionOptions

					if !conditionsTFItem.Parameters.Options.CaseSensitive.IsNull() {
						options.SetCaseSensitive(conditionsTFItem.Parameters.Options.CaseSensitive.ValueBool())
					}
					if !conditionsTFItem.Parameters.Options.MinLength.IsNull() {
						options.SetMinLength(conditionsTFItem.Parameters.Options.MinLength.ValueInt64())
					}
					parameters.Options = &options
				}

				conditionsDDItem.SetParameters(parameters)
			}

			conditions = append(conditions, *conditionsDDItem)
		}
		attributes.SetConditions(conditions)
	}

	if state.Scope != nil {
		var scope []datadogV2.ApplicationSecurityWafCustomRuleScope
		for _, scopeTFItem := range state.Scope {
			scopeDDItem := datadogV2.NewApplicationSecurityWafCustomRuleScopeWithDefaults()

			if !scopeTFItem.Env.IsNull() {
				scopeDDItem.SetEnv(scopeTFItem.Env.ValueString())
			}
			if !scopeTFItem.Service.IsNull() {
				scopeDDItem.SetService(scopeTFItem.Service.ValueString())
			}

			scope = append(scope, *scopeDDItem)
		}
		attributes.SetScope(scope)
	}

	if state.Action != nil {
		var action datadogV2.ApplicationSecurityWafCustomRuleAction

		if !state.Action.Action.IsNull() {
			action.SetAction(datadogV2.ApplicationSecurityWafCustomRuleActionAction(state.Action.Action.ValueString()))
		}

		if state.Action.Parameters != nil {
			var parameters datadogV2.ApplicationSecurityWafCustomRuleActionParameters

			if !state.Action.Parameters.Location.IsNull() {
				parameters.SetLocation(state.Action.Parameters.Location.ValueString())
			}
			if !state.Action.Parameters.StatusCode.IsNull() {
				parameters.SetStatusCode(state.Action.Parameters.StatusCode.ValueInt64())
			}
			action.Parameters = &parameters
		}
		attributes.Action = &action
	}

	if !state.Tags.IsNull() {
		var tags datadogV2.ApplicationSecurityWafCustomRuleTags
		diags.Append(state.Tags.ElementsAs(ctx, &tags.AdditionalProperties, false)...)
		attributes.SetTags(tags)
	}

	req := datadogV2.NewApplicationSecurityWafCustomRuleUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewApplicationSecurityWafCustomRuleUpdateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
