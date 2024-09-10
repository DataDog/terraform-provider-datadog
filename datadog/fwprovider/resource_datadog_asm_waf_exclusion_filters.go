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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
				Optional:    true,
				Description: "A description for the exclusion filter.",
				Default:     stringdefault.StaticString(""),
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates whether the exclusion filter is enabled.",
			},
			"path_glob": schema.StringAttribute{
				Required:    true,
				Description: "The path glob for the exclusion filter.",
			},
			"scope": schema.ListAttribute{
				Description: "The scope of the exclusion filter. Each entry is a map with 'env' and 'service' keys.",
				Optional:    true,
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
			"rules_target": schema.ListAttribute{
				Description: "The rules target of the exclusion filter. Each entry contains tags with 'category' and 'type'.",
				Optional:    true,
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
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

	r.updateStateFromCreateResponse(ctx, &state, &res)
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
	exclusionFiltersId, enabled, description, pathGlob := r.extractExclusionFilterAttributesFromResource(state)

	attributes := datadogV2.ASMExclusionFilterUpdateAttributes{}
	attributes.Description = &description
	attributes.Enabled = &enabled
	attributes.PathGlob = &pathGlob

	data := datadogV2.NewASMExclusionFilterUpdateData(attributes, datadogV2.ASMEXCLUSIONFILTERTYPE_EXCLUSION_FILTER)
	data.Id = &exclusionFiltersId
	return datadogV2.NewASMExclusionFilterUpdateRequest(*data), nil
}

func (r *asmWafExclusionFiltersResource) buildCreateASMExclusionFilterPayload(state *asmWafExclusionFiltersModel) (*datadogV2.ASMExclusionFilterCreateRequest, error) {
	_, enabled, description, pathGlob := r.extractExclusionFilterAttributesFromResource(state)

	attributes := datadogV2.ASMExclusionFilterCreateAttributes{}
	attributes.Description = description
	attributes.Enabled = enabled
	attributes.PathGlob = &pathGlob

	data := datadogV2.NewASMExclusionFilterCreateData(attributes, datadogV2.ASMEXCLUSIONFILTERTYPE_EXCLUSION_FILTER)
	return datadogV2.NewASMExclusionFilterCreateRequest(*data), nil
}

func (r *asmWafExclusionFiltersResource) extractExclusionFilterAttributesFromResource(state *asmWafExclusionFiltersModel) (string, bool, string, string) {
	id := state.Id.ValueString()
	enabled := state.Enabled.ValueBool()
	description := state.Description.ValueString()
	pathGlob := state.PathGlob.ValueString()

	return id, enabled, description, pathGlob
}

func (r *asmWafExclusionFiltersResource) updateStateFromCreateResponse(ctx context.Context, state *asmWafExclusionFiltersModel, res *datadogV2.ASMExclusionFilterResponse) {
	if len(res.GetData()) == 0 {
		return
	}

	filterData := res.GetData()[0]
	attributes := filterData.Attributes

	state.Id = types.StringValue(filterData.GetId())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.PathGlob = types.StringValue(attributes.GetPathGlob())
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

	var scopes []attr.Value
	if scopeList := attributes.GetScope(); len(scopeList) > 0 {
		for _, scopeItem := range scopeList {
			scopeValues := map[string]attr.Value{}

			if envValue := scopeItem.GetEnv(); envValue != "" {
				scopeValues["env"] = types.StringValue(envValue)
			}

			if serviceValue := scopeItem.GetService(); serviceValue != "" {
				scopeValues["service"] = types.StringValue(serviceValue)
			}

			if len(scopeValues) > 0 {
				scopeValue, _ := types.MapValue(types.StringType, scopeValues)
				scopes = append(scopes, scopeValue)
			}
		}
	}
	state.Scope, _ = types.ListValue(types.MapType{ElemType: types.StringType}, scopes)

	var rulesTarget []attr.Value
	if rulesTargetList := attributes.GetRulesTarget(); len(rulesTargetList) > 0 {
		for _, targetItem := range rulesTargetList {
			tags, tagsOk := targetItem.GetTagsOk()
			if tagsOk && tags != nil {
				tagValues := map[string]attr.Value{
					"category": types.StringValue(tags.GetCategory()),
					"type":     types.StringValue(tags.GetType()),
				}
				tagMapValue, _ := types.MapValue(types.StringType, tagValues)
				rulesTarget = append(rulesTarget, tagMapValue)
			}
		}
	}
	state.RulesTarget, _ = types.ListValue(types.MapType{ElemType: types.StringType}, rulesTarget)
}
