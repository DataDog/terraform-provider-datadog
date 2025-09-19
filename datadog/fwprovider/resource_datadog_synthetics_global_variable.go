package fwprovider

import (
	"context"
	"regexp"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &syntheticsGlobalVariableResource{}
	_ resource.ResourceWithImportState = &syntheticsGlobalVariableResource{}
	_ resource.ResourceWithModifyPlan  = &syntheticsGlobalVariableResource{}
)

type syntheticsGlobalVariableResource struct {
	Api  *datadogV1.SyntheticsApi
	Auth context.Context
}

type syntheticsGlobalVariableModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Tags             types.List   `tfsdk:"tags"`
	Value            types.String `tfsdk:"value"`
	Secure           types.Bool   `tfsdk:"secure"`
	ParseTestId      types.String `tfsdk:"parse_test_id"`
	ParseTestOptions types.List   `tfsdk:"parse_test_options"`
	Options          types.List   `tfsdk:"options"`
	RestrictedRoles  types.Set    `tfsdk:"restricted_roles"`
	IsTotp           types.Bool   `tfsdk:"is_totp"`
	IsFido           types.Bool   `tfsdk:"is_fido"`
}

type syntheticsGlobalVariableParseTestOptionsModel struct {
	Field             types.String                          `tfsdk:"field"`
	Type              types.String                          `tfsdk:"type"`
	Parser            []syntheticsGlobalVariableParserModel `tfsdk:"parser"`
	LocalVariableName types.String                          `tfsdk:"local_variable_name"`
}

func (syntheticsGlobalVariableParseTestOptionsModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"field":               types.StringType,
		"type":                types.StringType,
		"parser":              types.ListType{ElemType: types.ObjectType{AttrTypes: syntheticsGlobalVariableParserModel{}.AttributeTypes(ctx)}},
		"local_variable_name": types.StringType,
	}
}

type syntheticsGlobalVariableParserModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func (syntheticsGlobalVariableParserModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
}

type syntheticsGlobalVariableOptionsModel struct {
	TotpParameters []syntheticsGlobalVariableTotpParametersModel `tfsdk:"totp_parameters"`
}

func (syntheticsGlobalVariableOptionsModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"totp_parameters": types.ListType{ElemType: types.ObjectType{AttrTypes: syntheticsGlobalVariableTotpParametersModel{}.AttributeTypes(ctx)}},
	}
}

type syntheticsGlobalVariableTotpParametersModel struct {
	Digits          types.Int64 `tfsdk:"digits"`
	RefreshInterval types.Int64 `tfsdk:"refresh_interval"`
}

func (syntheticsGlobalVariableTotpParametersModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"digits":           types.Int64Type,
		"refresh_interval": types.Int64Type,
	}
}

func NewSyntheticsGlobalVariableResource() resource.Resource {
	return &syntheticsGlobalVariableResource{}
}

func (r *syntheticsGlobalVariableResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSyntheticsApiV1()
	r.Auth = providerData.Auth
}

func (r *syntheticsGlobalVariableResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "synthetics_global_variable"
}

func (r *syntheticsGlobalVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Synthetics global variable name.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the global variable.",
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Description: "A list of tags to associate with your synthetics global variable.",
				ElementType: types.StringType,
				Computed:    true,
				Optional:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"value": schema.StringAttribute{
				Description: "The value of the global variable. Required unless `is_fido` is set to `true`.",
				Optional:    true,
				Sensitive:   true,
			},
			"secure": schema.BoolAttribute{
				Description: "If set to true, the value of the global variable is hidden. This setting is automatically set to `true` if `is_totp` or `is_fido` is set to `true`.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"parse_test_id": schema.StringAttribute{
				Description: "Id of the Synthetics test to use for a variable from test.",
				Optional:    true,
			},
			"restricted_roles": schema.SetAttribute{
				Description:        "A list of role identifiers to associate with the Synthetics global variable. **Deprecated.** This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
				DeprecationMessage: "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
				ElementType:        types.StringType,
				Optional:           true,
			},
			"is_totp": schema.BoolAttribute{
				Description: "If set to true, the global variable is a TOTP variable.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_fido": schema.BoolAttribute{
				Description: "If set to true, the global variable is a FIDO variable.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"parse_test_options": schema.ListNestedBlock{
				Description: "ID of the Synthetics test to use a source of the global variable value.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "Required when type = `http_header`. Defines the header to use to extract the value",
							Optional:    true,
						},
						"type": schema.StringAttribute{
							Description: "Defines the source to use to extract the value.",
							Required:    true,
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsGlobalVariableParseTestOptionsTypeFromValue),
							},
						},
						"local_variable_name": schema.StringAttribute{
							Description: "When type is `local_variable`, name of the local variable to use to extract the value.",
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"parser": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:    true,
										Description: "Type of parser to extract the value.",
										Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV1.NewSyntheticsGlobalVariableParserTypeFromValue)},
									},
									"value": schema.StringAttribute{
										Description: "Value for the parser to use, required for type `json_path` or `regex`.",
										Optional:    true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"options": schema.ListNestedBlock{
				Description: "Additional options for the variable, such as a MFA token.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{},
					Blocks: map[string]schema.Block{
						"totp_parameters": schema.ListNestedBlock{
							Description: "Parameters needed for MFA/TOTP.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"digits": schema.Int64Attribute{
										Description: "Number of digits for the OTP.",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(4, 10),
										},
									},
									"refresh_interval": schema.Int64Attribute{
										Description: "Interval for which to refresh the token (in seconds).",
										Required:    true,
										Validators: []validator.Int64{
											int64validator.Between(0, 999),
										},
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					}},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *syntheticsGlobalVariableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *syntheticsGlobalVariableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state syntheticsGlobalVariableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	resp, httpResp, err := r.Api.GetGlobalVariable(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in backend
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsGlobalVariable"))
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

func (r *syntheticsGlobalVariableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state syntheticsGlobalVariableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildSyntheticsGlobalVariableRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateGlobalVariable(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsGlobalVariable"))
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

func (r *syntheticsGlobalVariableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state syntheticsGlobalVariableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	body, diags := r.buildSyntheticsGlobalVariableRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.EditGlobalVariable(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsGlobalVariable"))
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

func (r *syntheticsGlobalVariableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state syntheticsGlobalVariableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	httpResp, err := r.Api.DeleteGlobalVariable(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// The resource is assumed to be destroyed, and all state is removed.
			return
		}
		// The resource is assumed to still exist, and all prior state is preserved.
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting synthetics_global_variable"))
		return
	}
}

func (r *syntheticsGlobalVariableResource) updateState(ctx context.Context, state *syntheticsGlobalVariableModel, resp *datadogV1.SyntheticsGlobalVariable) {
	state.Id = types.StringValue(resp.GetId())

	state.Name = types.StringValue(resp.GetName())
	state.Description = types.StringValue(resp.GetDescription())
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, resp.GetTags())
	state.IsTotp = types.BoolValue(resp.GetIsTotp())
	state.IsFido = types.BoolValue(resp.GetIsFido())

	if value, ok := resp.GetValueOk(); ok {
		if !value.GetSecure() {
			// Only change the value in state if the global variable is not secure
			// Otherwise it will not be returned by the api, so we keep the config value
			state.Value = types.StringValue(value.GetValue())
		}
		if secure, ok := value.GetSecureOk(); ok {
			state.Secure = types.BoolValue(*secure)
		}
		if options, ok := value.GetOptionsOk(); ok {
			var optionsList []syntheticsGlobalVariableOptionsModel
			localVariableOptions := syntheticsGlobalVariableOptionsModel{}
			if totpParameters, ok := options.GetTotpParametersOk(); ok {
				localTotpParameters := syntheticsGlobalVariableTotpParametersModel{}
				localTotpParameters.Digits = types.Int64Value(int64(totpParameters.GetDigits()))
				localTotpParameters.RefreshInterval = types.Int64Value(int64(totpParameters.GetRefreshInterval()))
				localVariableOptions.TotpParameters = []syntheticsGlobalVariableTotpParametersModel{localTotpParameters}
			}
			optionsList = append(optionsList, localVariableOptions)
			state.Options, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: syntheticsGlobalVariableOptionsModel{}.AttributeTypes(ctx)}, optionsList)
		}
	}

	if attributes, ok := resp.GetAttributesOk(); ok {
		state.RestrictedRoles, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetRestrictedRoles())
	}

	if parseTestId, ok := resp.GetParseTestPublicIdOk(); ok {
		state.ParseTestId = types.StringValue(*parseTestId)

		if parseTestOptions, ok := resp.GetParseTestOptionsOk(); ok {
			var parseTestOptionsList []syntheticsGlobalVariableParseTestOptionsModel
			localParseTestOptions := syntheticsGlobalVariableParseTestOptionsModel{}
			localParseTestOptions.Type = types.StringValue(string(parseTestOptions.GetType()))
			if field, ok := parseTestOptions.GetFieldOk(); ok {
				localParseTestOptions.Field = types.StringValue(*field)
			}
			if parser, ok := parseTestOptions.GetParserOk(); ok {
				localParser := syntheticsGlobalVariableParserModel{}
				localParser.Type = types.StringValue(string(parser.GetType()))
				if value, ok := parser.GetValueOk(); ok {
					localParser.Value = types.StringValue(*value)
				}
				localParseTestOptions.Parser = []syntheticsGlobalVariableParserModel{localParser}
			}
			if localVariableName, ok := parseTestOptions.GetLocalVariableNameOk(); ok {
				localParseTestOptions.LocalVariableName = types.StringValue(*localVariableName)
			}

			parseTestOptionsList = append(parseTestOptionsList, localParseTestOptions)
			state.ParseTestOptions, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: syntheticsGlobalVariableParseTestOptionsModel{}.AttributeTypes(ctx)}, parseTestOptionsList)
		}
	}
}

func (r *syntheticsGlobalVariableResource) buildSyntheticsGlobalVariableRequestBody(ctx context.Context, state *syntheticsGlobalVariableModel) (*datadogV1.SyntheticsGlobalVariableRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	syntheticsGlobalVariableRequest := datadogV1.NewSyntheticsGlobalVariableRequestWithDefaults()

	if !state.Name.IsNull() {
		syntheticsGlobalVariableRequest.SetName(state.Name.ValueString())
	}
	if !state.Description.IsNull() {
		syntheticsGlobalVariableRequest.SetDescription(state.Description.ValueString())
	}
	if !state.IsFido.IsNull() {
		syntheticsGlobalVariableRequest.SetIsFido(state.IsFido.ValueBool())
	}
	if !state.IsTotp.IsNull() {
		syntheticsGlobalVariableRequest.SetIsTotp(state.IsTotp.ValueBool())
	}

	tags := make([]string, 0)
	if !state.Tags.IsNull() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	}
	syntheticsGlobalVariableRequest.SetTags(tags)

	if !state.RestrictedRoles.IsNull() {
		var attributes datadogV1.SyntheticsGlobalVariableAttributes

		var restrictedRoles []string
		diags.Append(state.RestrictedRoles.ElementsAs(ctx, &restrictedRoles, false)...)
		attributes.SetRestrictedRoles(restrictedRoles)

		syntheticsGlobalVariableRequest.Attributes = &attributes
	}

	if !state.ParseTestId.IsNull() {
		if !state.ParseTestOptions.IsNull() && len(state.ParseTestOptions.Elements()) > 0 {
			syntheticsGlobalVariableRequest.SetParseTestPublicId(state.ParseTestId.ValueString())

			var parseTestOptionsList []syntheticsGlobalVariableParseTestOptionsModel
			diags.Append(state.ParseTestOptions.ElementsAs(ctx, &parseTestOptionsList, false)...)
			if diags.HasError() {
				return nil, diags
			}

			if len(parseTestOptionsList) > 0 {
				var parseTestOptions datadogV1.SyntheticsGlobalVariableParseTestOptions
				if !parseTestOptionsList[0].Field.IsNull() {
					parseTestOptions.SetField(parseTestOptionsList[0].Field.ValueString())
				}
				if !parseTestOptionsList[0].LocalVariableName.IsNull() {
					parseTestOptions.SetLocalVariableName(parseTestOptionsList[0].LocalVariableName.ValueString())
				}
				if !parseTestOptionsList[0].Type.IsNull() {
					parseTestOptions.SetType(datadogV1.SyntheticsGlobalVariableParseTestOptionsType(parseTestOptionsList[0].Type.ValueString()))
				}

				if len(parseTestOptionsList[0].Parser) > 0 {
					var parser datadogV1.SyntheticsVariableParser

					if !parseTestOptionsList[0].Parser[0].Type.IsNull() {
						parser.SetType(datadogV1.SyntheticsGlobalVariableParserType(parseTestOptionsList[0].Parser[0].Type.ValueString()))
					}
					if !parseTestOptionsList[0].Parser[0].Value.IsNull() {
						parser.SetValue(parseTestOptionsList[0].Parser[0].Value.ValueString())
					}
					parseTestOptions.Parser = &parser
				}
				syntheticsGlobalVariableRequest.ParseTestOptions = &parseTestOptions
			}
		}
	}

	if !state.Value.IsNull() {
		var value datadogV1.SyntheticsGlobalVariableValue

		value.SetValue(state.Value.ValueString())
		if !state.Secure.IsNull() {
			value.SetSecure(state.Secure.ValueBool())
		}

		if !state.Options.IsNull() && len(state.Options.Elements()) > 0 {
			var optionsList []syntheticsGlobalVariableOptionsModel
			diags.Append(state.Options.ElementsAs(ctx, &optionsList, false)...)
			if diags.HasError() {
				return nil, diags
			}

			if len(optionsList) > 0 {
				var options datadogV1.SyntheticsGlobalVariableOptions

				if len(optionsList[0].TotpParameters) > 0 {
					var totpParameters datadogV1.SyntheticsGlobalVariableTOTPParameters

					if !optionsList[0].TotpParameters[0].Digits.IsNull() {
						totpParameters.SetDigits(int32(optionsList[0].TotpParameters[0].Digits.ValueInt64()))
					}
					if !optionsList[0].TotpParameters[0].RefreshInterval.IsNull() {
						totpParameters.SetRefreshInterval(int32(optionsList[0].TotpParameters[0].RefreshInterval.ValueInt64()))
					}
					options.TotpParameters = &totpParameters
				}
				value.Options = &options
			}
		}
		syntheticsGlobalVariableRequest.Value = &value
	}

	return syntheticsGlobalVariableRequest, diags
}

func (r syntheticsGlobalVariableResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config syntheticsGlobalVariableModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	isFido := config.IsFido.ValueBool()
	isTotp := config.IsTotp.ValueBool()

	// If `is_fido` is `true` and `value` is set, return an error.
	if isFido && !config.Value.IsNull() {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("value"),
			"Invalid Configuration",
			"`value` cannot be set when `is_fido` is `true`.",
		)
	}

	// If `is_fido` is `false` and `value` is not set, return an error.
	if !isFido && config.Value.IsNull() {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("value"),
			"Invalid Configuration",
			"`value` must be set.",
		)
	}

	// If `is_fido` or `is_totp` is `true`, is `secure` should not be set or should be set to `true`
	if (isFido || isTotp) && !config.Secure.IsNull() && !config.Secure.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("secure"),
			"Invalid Configuration",
			"`secure` must not be set to `false` if `is_totp` or `is_fido` is set to `true`.",
		)
	}
}

func (r syntheticsGlobalVariableResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var config syntheticsGlobalVariableModel
	diags := req.Plan.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isTotp := config.IsTotp.ValueBool()
	isFido := config.IsFido.ValueBool()

	// Default to true for secure when is_fido or is_totp is true
	if isFido || isTotp {
		resp.Plan.SetAttribute(ctx, frameworkPath.Root("secure"), types.BoolValue(true))
		return
	}
}
