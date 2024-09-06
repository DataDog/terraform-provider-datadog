package fwprovider

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	asmWafExclusionFiltersMutex sync.Mutex
	_                           resource.ResourceWithConfigure   = &asmWafExclusionFiltersResource{}
	_                           resource.ResourceWithImportState = &asmWafExclusionFiltersResource{}
)

type asmWafExclusionFiltersModel struct {
	Id           types.String `tfsdk:"id"`
	Description  types.String `tfsdk:"description"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Search_Query types.String `tfsdk:"search_query"`
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

func (r *asmWafExclusionFiltersResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) { // to change
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
			"search_query": schema.StringAttribute{
				Required:    true,
				Description: "The search query of the exclusion filter",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	exclusionFiltersId := state.Id.ValueString()
	res, httpResponse, err := r.api.ListASMExclusionFilters(r.auth)

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching exclusion filters"))
		return
	}

	var matchedExclusionFilter *datadogV2.ASMExclusionFilter
	for _, exclusionFilter := range res.GetData() {
		if exclusionFilter.GetId() == exclusionFiltersId {
			matchedExclusionFilter = &exclusionFilter
			break
		}
	}

	if matchedExclusionFilter == nil {
		response.State.RemoveResource(ctx)
		return
	}

	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, matchedExclusionFilter)
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

// to change: from the Create and Update functions
func (r *asmWafExclusionFiltersResource) updateStateFromResponse(ctx context.Context, state *asmWafExclusionFiltersModel, exclusionFilter *datadogV2.ASMExclusionFilter) {
	// Met à jour l'état avec les attributs de l'API
	state.Id = types.StringValue(exclusionFilter.GetId())
	attributes := exclusionFilter.GetAttributes()

	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Search_Query = types.StringValue(attributes.GetSearchQuery())
}
