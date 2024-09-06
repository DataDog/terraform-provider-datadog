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
	// 	var state asmWafExclusionFiltersModel
	// 	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	// 	if response.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	asmWafExclusionFiltersMutex.Lock()
	// 	defer asmWafExclusionFiltersMutex.Unlock()

	// 	exclusionFiltersPayload, err := r.buildCreateAsmWafExclusionFiltersPayload(&state)
	// 	if err != nil {
	// 		response.Diagnostics.AddError("error while parsing resource", err.Error())
	// 	}

	// 	res, _, err := r.api.handlePostExclusionFilters(r.auth, *exclusionFiltersPayload) // to change: endpoint POST
	// 	if err != nil {
	// 		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating agent rule"))
	// 		return
	// 	}
	// 	if err := utils.CheckForUnparsed(response); err != nil {
	// 		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
	// 		return
	// 	}

	// r.updateStateFromResponse(ctx, &state, &res)
	// response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *asmWafExclusionFiltersResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state asmWafExclusionFiltersModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Récupérer l'ID de l'exclusion filter
	exclusionFilterId := state.Id.ValueString()

	// Appel à l'API pour obtenir le filtre d'exclusion correspondant à l'ID
	res, httpResponse, err := r.api.GetASMExclusionFilters(r.auth, exclusionFilterId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching exclusion filter"))
		return
	}

	// Extraire les données à partir de la réponse
	dataList, ok := res.AdditionalProperties["data"].([]interface{})
	if !ok || len(dataList) == 0 {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(fmt.Errorf("no data found in response"), "error extracting exclusion filter data"))
		return
	}

	// Extraire les informations du premier élément (car la requête renvoie un seul filtre d'exclusion)
	filterData := dataList[0].(map[string]interface{})
	attributes := filterData["attributes"].(map[string]interface{})

	// Mettre à jour l'état en fonction des attributs extraits
	state.Id = types.StringValue(filterData["id"].(string))
	state.Description = types.StringValue(attributes["description"].(string))
	state.Enabled = types.BoolValue(attributes["enabled"].(bool))

	// Extraire le scope
	var scopes []attr.Value
	if scopeList, ok := attributes["scope"].([]interface{}); ok {
		for _, scopeItem := range scopeList {
			scopeMap := scopeItem.(map[string]interface{})
			scopeValue, _ := types.MapValue(types.StringType, map[string]attr.Value{
				"env":     types.StringValue(scopeMap["env"].(string)),
				"service": types.StringValue(scopeMap["service"].(string)),
			})
			scopes = append(scopes, scopeValue)
		}
	}
	state.Scope, _ = types.ListValue(types.MapType{ElemType: types.StringType}, scopes)

	// Mettre à jour l'état
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *asmWafExclusionFiltersResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// 	var state asmWafExclusionFiltersModel
	// 	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	// 	if response.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	asmWafExclusionFiltersMutex.Lock()
	// 	defer asmWafExclusionFiltersMutex.Unlock()

	// 	exclusionFiltersPayload, err := r.buildUpdateAsmWafExclusionFiltersPayload(&state)
	// 	if err != nil {
	// 		response.Diagnostics.AddError("error while parsing resource", err.Error())
	// 	}

	// 	res, _, err := r.api.handlePutExclusionFilter(r.auth, state.Id.ValueString(), *exclusionFiltersPayload) // to change: endpoint PATCH/PUT
	// 	if err != nil {
	// 		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating agent rule"))
	// 		return
	// 	}
	// 	if err := utils.CheckForUnparsed(response); err != nil {
	// 		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
	// 		return
	// 	}

	// r.updateStateFromResponse(ctx, &state, &res)
	// response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *asmWafExclusionFiltersResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// 	var state asmWafExclusionFiltersModel
	// 	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	// 	if response.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	asmWafExclusionFiltersMutex.Lock()
	// 	defer asmWafExclusionFiltersMutex.Unlock()

	// 	id := state.Id.ValueString()

	// httpResp, err := r.api.handleDeleteExclusionFilterByID(r.auth, id) // to change: endpoint DELETE
	//
	//	if err != nil {
	//		if httpResp != nil && httpResp.StatusCode == 404 {
	//			return
	//		}
	//		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting agent rule"))
	//		return
	//	}
}

// // to change: payload from the Create function
// func (r *asmWafExclusionFiltersResource) buildCreateAsmWafExclusionFiltersPayload(state *asmWafExclusionFiltersModel) (*datadogV2.CloudWorkloadSecurityExclusionFiltersCreateRequest, error) {
// 	_, description, enabled, search_query := r.extractExclusionFiltersAttributesFromResource(state)

// 	attributes := datadogV2.CloudWorkloadSecurityExclusionFiltersCreateAttributes{}
// 	attributes.Search_Query = search_query
// 	attributes.Description = description
// 	attributes.Enabled = &enabled

// 	data := datadogV2.NewCloudWorkloadSecurityExclusionFiltersCreateData(attributes, datadogV2.CLOUDWORKLOADSECURITYEXCLUSIONFILTERSTYPE_AGENT_RULE)
// 	return datadogV2.NewCloudWorkloadSecurityExclusionFiltersCreateRequest(*data), nil
// }

// // to change: payload from the Update function
// func (r *asmWafExclusionFiltersResource) buildUpdateAsmWafExclusionFiltersPayload(state *asmWafExclusionFiltersModel) (*datadogV2.CloudWorkloadSecurityExclusionFiltersUpdateRequest, error) {
// 	exclusionFiltersId, _, description, enabled, _ := r.extractExclusionFiltersAttributesFromResource(state)

// 	attributes := datadogV2.CloudWorkloadSecurityExclusionFiltersUpdateAttributes{}
// 	attributes.Description = description
// 	attributes.Enabled = &enabled

// 	data := datadogV2.NewCloudWorkloadSecurityExclusionFiltersUpdateData(attributes, datadogV2.CLOUDWORKLOADSECURITYEXCLUSIONFILTERSTYPE_AGENT_RULE)
// 	data.Id = &exclusionFiltersId
// 	return datadogV2.NewCloudWorkloadSecurityExclusionFiltersUpdateRequest(*data), nil
// }

// // called from the payloads above
// func (r *asmWafExclusionFiltersResource) extractExclusionFiltersAttributesFromResource(state *asmWafExclusionFiltersModel) (string, *string, bool, string) {
// 	// Mandatory fields
// 	id := state.Id.ValueString()
// 	enabled := state.Enabled.ValueBool()
// 	search_query := state.Search_Query.ValueString()
// 	description := state.Description.ValueStringPointer()

// 	return id, description, enabled, search_query
// }
