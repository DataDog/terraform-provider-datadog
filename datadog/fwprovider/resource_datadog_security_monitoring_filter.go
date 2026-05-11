package fwprovider

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

const securityFilterType = "security_filters"

var filterWriteMutex = sync.Mutex{}

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringFilterResource{}
	_ resource.ResourceWithImportState = &securityMonitoringFilterResource{}
)

type securityMonitoringFilterResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

type securityMonitoringFilterResourceModel struct {
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	Version          types.Int64            `tfsdk:"version"`
	Query            types.String           `tfsdk:"query"`
	IsEnabled        types.Bool             `tfsdk:"is_enabled"`
	FilteredDataType types.String           `tfsdk:"filtered_data_type"`
	ExclusionFilter  []exclusionFilterModel `tfsdk:"exclusion_filter"`
}

type exclusionFilterModel struct {
	Name  types.String `tfsdk:"name"`
	Query types.String `tfsdk:"query"`
}

func NewSecurityMonitoringFilterResource() resource.Resource {
	return &securityMonitoringFilterResource{}
}

func (r *securityMonitoringFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_filter"
}

func (r *securityMonitoringFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource for security filters.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the security filter.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The version of the security filter.",
			},
			"query": schema.StringAttribute{
				Required:    true,
				Description: "The query of the security filter.",
			},
			"is_enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the security filter is enabled.",
			},
			"filtered_data_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("logs"),
				Description: "The filtered data type.",
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewSecurityFilterFilteredDataTypeFromValue),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"exclusion_filter": schema.ListNestedBlock{
				Description: "Exclusion filters to exclude some logs from the security filter.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Exclusion filter name.",
						},
						"query": schema.StringAttribute{
							Required:    true,
							Description: "Exclusion filter query. Logs that match this query are excluded from the security filter.",
						},
					},
				},
			},
		},
	}
}

func (r *securityMonitoringFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *securityMonitoringFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan securityMonitoringFilterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	filterCreate := buildSecMonFilterCreatePayload(&plan)

	filterWriteMutex.Lock()
	defer filterWriteMutex.Unlock()

	filterResponse, httpResponse, err := r.api.CreateSecurityFilter(r.auth, *filterCreate)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error creating security monitoring filter"), ""))
		return
	}
	if err := utils.CheckForUnparsed(filterResponse); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	updateResourceDataFilterFromResponse(&plan, filterResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *securityMonitoringFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityMonitoringFilterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	filterResponse, httpResponse, err := r.api.GetSecurityFilter(r.auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error fetching security monitoring filter"), ""))
		return
	}
	if err := utils.CheckForUnparsed(filterResponse); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	updateResourceDataFilterFromResponse(&state, filterResponse)

	// handle warning
	if filterResponse.HasMeta() {
		filterMeta := filterResponse.GetMeta()
		if warning := filterMeta.GetWarning(); warning != "" {
			response.Diagnostics.AddWarning(warning, "")
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityMonitoringFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan securityMonitoringFilterResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state securityMonitoringFilterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	filterUpdate := buildSecMonFilterUpdatePayload(&plan, int32(state.Version.ValueInt64()))

	filterWriteMutex.Lock()
	defer filterWriteMutex.Unlock()

	filterResponse, httpResponse, err := r.api.UpdateSecurityFilter(r.auth, state.ID.ValueString(), *filterUpdate)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error updating security monitoring filter"), ""))
		return
	}
	if err := utils.CheckForUnparsed(filterResponse); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	updateResourceDataFilterFromResponse(&plan, filterResponse)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *securityMonitoringFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityMonitoringFilterResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	filterWriteMutex.Lock()
	defer filterWriteMutex.Unlock()

	httpResponse, err := r.api.DeleteSecurityFilter(r.auth, state.ID.ValueString())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error deleting security monitoring filter"), ""))
	}
}

func updateResourceDataFilterFromResponse(state *securityMonitoringFilterResourceModel, filterResponse datadogV2.SecurityFilterResponse) {
	data := filterResponse.GetData()
	state.ID = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	state.Version = types.Int64Value(int64(attributes.GetVersion()))
	state.Name = types.StringValue(attributes.GetName())
	state.Query = types.StringValue(attributes.GetQuery())
	state.IsEnabled = types.BoolValue(attributes.GetIsEnabled())
	state.FilteredDataType = types.StringValue(string(attributes.GetFilteredDataType()))

	if _, ok := attributes.GetExclusionFiltersOk(); ok {
		state.ExclusionFilter = extractExclusionFiltersTF(attributes)
	}
}

func extractExclusionFiltersTF(attributes datadogV2.SecurityFilterAttributes) []exclusionFilterModel {
	exclusionFilters := make([]exclusionFilterModel, len(attributes.GetExclusionFilters()))
	for i, ef := range attributes.GetExclusionFilters() {
		exclusionFilters[i] = exclusionFilterModel{
			Name:  types.StringValue(ef.GetName()),
			Query: types.StringValue(ef.GetQuery()),
		}
	}
	return exclusionFilters
}

func buildSecMonFilterUpdatePayload(plan *securityMonitoringFilterResourceModel, currentVersion int32) *datadogV2.SecurityFilterUpdateRequest {
	payload := datadogV2.SecurityFilterUpdateRequest{}
	payload.Data.Type = securityFilterType
	// set the version from current state
	payload.Data.Attributes.SetVersion(currentVersion)

	isEnabled, name, filteredDataType, query, filters := extractFilterAttributedFromResource(plan)

	payload.Data.Attributes.SetIsEnabled(isEnabled)
	payload.Data.Attributes.SetName(name)
	payload.Data.Attributes.SetFilteredDataType(filteredDataType)
	payload.Data.Attributes.SetQuery(query)
	payload.Data.Attributes.SetExclusionFilters(filters)

	return &payload
}

func buildSecMonFilterCreatePayload(plan *securityMonitoringFilterResourceModel) *datadogV2.SecurityFilterCreateRequest {
	payload := datadogV2.SecurityFilterCreateRequest{}
	payload.Data.Type = securityFilterType

	isEnabled, name, filteredDataType, query, filters := extractFilterAttributedFromResource(plan)

	payload.Data.Attributes.SetIsEnabled(isEnabled)
	payload.Data.Attributes.SetName(name)
	payload.Data.Attributes.SetFilteredDataType(filteredDataType)
	payload.Data.Attributes.SetQuery(query)
	payload.Data.Attributes.SetExclusionFilters(filters)

	return &payload
}

func extractFilterAttributedFromResource(plan *securityMonitoringFilterResourceModel) (bool, string, datadogV2.SecurityFilterFilteredDataType, string, []datadogV2.SecurityFilterExclusionFilter) {
	isEnabled := plan.IsEnabled.ValueBool()
	name := plan.Name.ValueString()
	filteredDataType := datadogV2.SecurityFilterFilteredDataType(plan.FilteredDataType.ValueString())
	query := plan.Query.ValueString()

	filters := make([]datadogV2.SecurityFilterExclusionFilter, len(plan.ExclusionFilter))
	for i, ef := range plan.ExclusionFilter {
		filters[i].SetName(ef.Name.ValueString())
		filters[i].SetQuery(ef.Query.ValueString())
	}

	return isEnabled, name, filteredDataType, query, filters
}
