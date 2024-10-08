package fwprovider

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	asmWafExclusionFiltersMutex sync.Mutex
	_                           resource.ResourceWithConfigure   = &asmWafExclusionFiltersResource{}
	_                           resource.ResourceWithImportState = &asmWafExclusionFiltersResource{}
)

type asmWafExclusionFiltersModel struct {
	Id          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	PathGlob    types.String `tfsdk:"path_glob"`
	Parameters  types.List   `tfsdk:"parameters"`
	Scope       types.List   `tfsdk:"scope"`
	RulesTarget types.List   `tfsdk:"rules_target"`
}

type asmWafExclusionFiltersResource struct {
	api  *datadogV2.ASMExclusionFiltersApi
	auth context.Context
}

func NewAsmWafExclusionFiltersResource() resource.Resource {
	return &asmWafExclusionFiltersResource{}
}

func (r *asmWafExclusionFiltersResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "asm_waf_exclusion_filters"
}

func (r *asmWafExclusionFiltersResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetASMExclusionFiltersApiV2() // to change
	r.auth = providerData.Auth
}

func (r *asmWafExclusionFiltersResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog ASM WAF Exclusion Filters API resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"description": schema.StringAttribute{
				Required:    true,
				Description: "A description for the exclusion filter.",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the exclusion filter is enabled.",
			},
			"path_glob": schema.StringAttribute{
				Required:    true,
				Description: "The path glob for the exclusion filter.",
			},
			"parameters": schema.ListAttribute{
				Description: "List of parameters for the exclusion filters.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"scope": schema.ListAttribute{
				Description: "The scope of the exclusion filter. Each entry contains 'env' and 'service'.",
				Optional:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"env":     types.StringType,
						"service": types.StringType,
					},
				},
			},
			"rules_target": schema.ListAttribute{
				Description: "The rules target of the exclusion filter with 'rule_id'.",
				Optional:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"rule_id": types.StringType,
					},
				},
			},
		},
	}
}

func (r *asmWafExclusionFiltersResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *asmWafExclusionFiltersResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state asmWafExclusionFiltersModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	asmWafExclusionFiltersMutex.Lock()
	defer asmWafExclusionFiltersMutex.Unlock()

	exclusionFilterPayload, err := r.buildCreateASMExclusionFilterPayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
	}

	res, _, err := r.api.CreateASMExclusionFilter(r.auth, *exclusionFilterPayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating exclusion filter"))
		return
	}

	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *asmWafExclusionFiltersResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state asmWafExclusionFiltersModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	exclusionFilterId := state.Id.ValueString()

	res, httpResponse, err := r.api.GetASMExclusionFilters(r.auth, exclusionFilterId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching exclusion filter"))
		return
	}

	if len(res.Data) == 0 {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(fmt.Errorf("no data found in response"), "error extracting exclusion filter data"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *asmWafExclusionFiltersResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state asmWafExclusionFiltersModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	asmWafExclusionFiltersMutex.Lock()
	defer asmWafExclusionFiltersMutex.Unlock()

	exclusionFiltersPayload, err := r.buildUpdateAsmWafExclusionFiltersPayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
	}

	res, _, err := r.api.UpdateASMExclusionFilter(r.auth, state.Id.ValueString(), *exclusionFiltersPayload) // to change: endpoint PATCH/PUT
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating agent rule"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *asmWafExclusionFiltersResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state asmWafExclusionFiltersModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	asmWafExclusionFiltersMutex.Lock()
	defer asmWafExclusionFiltersMutex.Unlock()

	id := state.Id.ValueString()

	httpResp, err := r.api.DeleteASMExclusionFilter(r.auth, id)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting exclusion filter"))
		return
	}
}

func (r *asmWafExclusionFiltersResource) buildUpdateAsmWafExclusionFiltersPayload(state *asmWafExclusionFiltersModel) (*datadogV2.ASMExclusionFilterUpdateRequest, error) {
	exclusionFiltersId, enabled, description, pathGlob, parameters, scopeList, rulesTargetList := r.extractExclusionFilterAttributesFromResource(state)

	attributes := datadogV2.ASMExclusionFilterUpdateAttributes{
		Description: description,
		Enabled:     enabled,
		PathGlob:    pathGlob,
	}

	if len(parameters) > 0 {
		attributes.Parameters = parameters
	}

	if len(scopeList) > 0 {
		var newScopeList []datadogV2.ASMExclusionFilterScope
		for _, scopeItem := range scopeList {
			newScopeList = append(newScopeList, datadogV2.ASMExclusionFilterScope{
				Env:     scopeItem.Env,
				Service: scopeItem.Service,
			})
		}
		attributes.Scope = newScopeList
	}

	if len(rulesTargetList) > 0 {
		var newRulesTargetList []datadogV2.ASMExclusionFilterRulesTarget
		for _, targetItem := range rulesTargetList {
			newRulesTargetList = append(newRulesTargetList, datadogV2.ASMExclusionFilterRulesTarget{
				RuleId: targetItem.RuleId,
			})
		}
		attributes.RulesTarget = newRulesTargetList
	}

	data := datadogV2.NewASMExclusionFilterUpdateData(attributes, datadogV2.ASMEXCLUSIONFILTERTYPE_EXCLUSION_FILTER)
	data.Id = &exclusionFiltersId
	return datadogV2.NewASMExclusionFilterUpdateRequest(*data), nil
}

func (r *asmWafExclusionFiltersResource) buildCreateASMExclusionFilterPayload(state *asmWafExclusionFiltersModel) (*datadogV2.ASMExclusionFilterCreateRequest, error) {
	_, enabled, description, pathGlob, parameters, scopeList, rulesTargetList := r.extractExclusionFilterAttributesFromResource(state)

	attributes := datadogV2.ASMExclusionFilterCreateAttributes{
		Description: description,
		Enabled:     enabled,
		PathGlob:    pathGlob,
	}

	if len(parameters) > 0 {
		attributes.Parameters = parameters
	}

	if len(scopeList) > 0 {
		var newScopeList []datadogV2.ASMExclusionFilterScope
		for _, scopeItem := range scopeList {
			newScopeList = append(newScopeList, datadogV2.ASMExclusionFilterScope{
				Env:     scopeItem.Env,
				Service: scopeItem.Service,
			})
		}
		attributes.Scope = newScopeList
	}

	if len(rulesTargetList) > 0 {
		var newRulesTargetList []datadogV2.ASMExclusionFilterRulesTarget
		for _, targetItem := range rulesTargetList {
			newRulesTargetList = append(newRulesTargetList, datadogV2.ASMExclusionFilterRulesTarget{
				RuleId: targetItem.RuleId,
			})
		}
		attributes.RulesTarget = newRulesTargetList
	}

	data := datadogV2.NewASMExclusionFilterCreateData(attributes, datadogV2.ASMEXCLUSIONFILTERTYPE_EXCLUSION_FILTER)
	return datadogV2.NewASMExclusionFilterCreateRequest(*data), nil
}

func (r *asmWafExclusionFiltersResource) extractExclusionFilterAttributesFromResource(state *asmWafExclusionFiltersModel) (string, bool, string, string, []string, []datadogV2.ASMExclusionFilterScope, []datadogV2.ASMExclusionFilterRulesTarget) {
	id := state.Id.ValueString()
	enabled := state.Enabled.ValueBool()
	description := state.Description.ValueString()
	pathGlob := state.PathGlob.ValueString()

	var parameters []string
	if !state.Parameters.IsNull() && len(state.Parameters.Elements()) > 0 {
		for _, param := range state.Parameters.Elements() {
			parameters = append(parameters, param.(types.String).ValueString())
		}
	}

	var scopeList []datadogV2.ASMExclusionFilterScope
	if !state.Scope.IsNull() && len(state.Scope.Elements()) > 0 {
		for _, scopeItem := range state.Scope.Elements() {
			scopeMap := scopeItem.(types.Object).Attributes()

			env := scopeMap["env"].(types.String).ValueString()
			service := scopeMap["service"].(types.String).ValueString()

			envPtr := &env
			servicePtr := &service

			scopeList = append(scopeList, datadogV2.ASMExclusionFilterScope{
				Env:     envPtr,
				Service: servicePtr,
			})
		}
	}

	var rulesTargetList []datadogV2.ASMExclusionFilterRulesTarget
	if !state.RulesTarget.IsNull() && len(state.RulesTarget.Elements()) > 0 {
		for _, targetItem := range state.RulesTarget.Elements() {
			rulesMap := targetItem.(types.Object).Attributes()

			ruleId := rulesMap["rule_id"].(types.String).ValueString()

			ruleIdPtr := &ruleId

			rulesTargetList = append(rulesTargetList, datadogV2.ASMExclusionFilterRulesTarget{
				RuleId: ruleIdPtr,
			})
		}
	}

	return id, enabled, description, pathGlob, parameters, scopeList, rulesTargetList
}

func (r *asmWafExclusionFiltersResource) updateStateFromResponse(ctx context.Context, state *asmWafExclusionFiltersModel, res *datadogV2.ASMExclusionFilterResponse) {

	if len(res.Data) == 0 {
		return
	}

	filterData := res.Data[0]

	attributes := filterData.Attributes

	state.Id = types.StringValue(filterData.GetId())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.PathGlob = types.StringValue(attributes.GetPathGlob())

	var parameters []attr.Value
	for _, param := range attributes.GetParameters() {
		parameters = append(parameters, types.StringValue(param))
	}
	state.Parameters, _ = types.ListValue(types.StringType, parameters)

	var scopes []attr.Value
	if scopeList := attributes.GetScope(); len(scopeList) > 0 {
		for _, scopeItem := range scopeList {
			scopeObject := map[string]attr.Value{
				"env":     types.StringValue(scopeItem.GetEnv()),
				"service": types.StringValue(scopeItem.GetService()),
			}
			scopeValue, _ := types.ObjectValue(map[string]attr.Type{
				"env":     types.StringType,
				"service": types.StringType,
			}, scopeObject)
			scopes = append(scopes, scopeValue)
		}
	}
	state.Scope, _ = types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"env":     types.StringType,
			"service": types.StringType,
		},
	}, scopes)

	var rulesTarget []attr.Value
	if rulesTargetList := attributes.GetRulesTarget(); len(rulesTargetList) > 0 {
		for _, targetItem := range rulesTargetList {
			ruleValues := map[string]attr.Value{
				"rule_id": types.StringValue(targetItem.GetRuleId()),
			}
			ruleObject, _ := types.ObjectValue(map[string]attr.Type{
				"rule_id": types.StringType,
			}, ruleValues)
			rulesTarget = append(rulesTarget, ruleObject)
		}
	}
	state.RulesTarget, _ = types.ListValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"rule_id": types.StringType,
		},
	}, rulesTarget)

}
