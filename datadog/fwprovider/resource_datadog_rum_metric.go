package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &rumMetricResource{}
	_ resource.ResourceWithImportState = &rumMetricResource{}
)

type rumMetricResource struct {
	Api  *datadogV2.RumMetricsApi
	Auth context.Context
}

type rumMetricModel struct {
	ID         types.String     `tfsdk:"id"`
	EventType  types.String     `tfsdk:"event_type"`
	GroupBy    []*groupByModel  `tfsdk:"group_by"`
	Compute    *computeModel    `tfsdk:"compute"`
	Filter     *filterModel     `tfsdk:"filter"`
	Uniqueness *uniquenessModel `tfsdk:"uniqueness"`
}

type groupByModel struct {
	Path    types.String `tfsdk:"path"`
	TagName types.String `tfsdk:"tag_name"`
}

type computeModel struct {
	AggregationType    types.String `tfsdk:"aggregation_type"`
	IncludePercentiles types.Bool   `tfsdk:"include_percentiles"`
	Path               types.String `tfsdk:"path"`
}

type filterModel struct {
	Query types.String `tfsdk:"query"`
}

type uniquenessModel struct {
	When types.String `tfsdk:"when"`
}

func NewRumMetricResource() resource.Resource {
	return &rumMetricResource{}
}

func (r *rumMetricResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRumMetricsApiV2()
	r.Auth = providerData.Auth
}

func (r *rumMetricResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_metric"
}

func (r *rumMetricResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RumMetric resource. This can be used to create and manage Datadog rum_metric.",
		Attributes: map[string]schema.Attribute{
			"event_type": schema.StringAttribute{
				Optional:    true,
				Description: "The type of RUM events to filter on.",
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"group_by": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Optional:    true,
							Description: "The path to the value the rum-based metric will be aggregated over.",
						},
						"tag_name": schema.StringAttribute{
							Optional:    true,
							Description: "Eventual name of the tag that gets created. By default, `path` is used as the tag name.",
						},
					},
				},
			},
			"compute": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"aggregation_type": schema.StringAttribute{
						Optional:    true,
						Description: "The type of aggregation to use.",
					},
					"include_percentiles": schema.BoolAttribute{
						Optional:    true,
						Description: "Toggle to include or exclude percentile aggregations for distribution metrics. Only present when `aggregation_type` is `distribution`.",
					},
					"path": schema.StringAttribute{
						Optional:    true,
						Description: "The path to the value the rum-based metric will aggregate on. Only present when `aggregation_type` is `distribution`.",
					},
				},
			},
			"filter": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Optional:    true,
						Description: "The search query - following the RUM search syntax.",
					},
				},
			},
			"uniqueness": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"when": schema.StringAttribute{
						Optional:    true,
						Description: "When to count updatable events. `match` when the event is first seen, or `end` when the event is complete.",
					},
				},
			},
		},
	}
}

func (r *rumMetricResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *rumMetricResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumMetricModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetRumMetric(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumMetric"))
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

func (r *rumMetricResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumMetricModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildRumMetricRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateRumMetric(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumMetric"))
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

func (r *rumMetricResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumMetricModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildRumMetricUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateRumMetric(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumMetric"))
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

func (r *rumMetricResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumMetricModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteRumMetric(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting rum_metric"))
		return
	}
}

func (r *rumMetricResource) updateState(ctx context.Context, state *rumMetricModel, resp *datadogV2.RumMetricResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.EventType = types.StringValue(attributes.GetEventType())

	if groupBy, ok := attributes.GetGroupByOk(); ok && len(*groupBy) > 0 {
		state.GroupBy = []*groupByModel{}
		for _, groupByDd := range *groupBy {
			groupByTfItem := groupByModel{}

			if groupBy, ok := groupByDd.GetGroupByOk(); ok {

				groupByTf := groupByModel{}
				if path, ok := groupBy.GetPathOk(); ok {
					groupByTf.Path = types.StringValue(*path)
				}
				if tagName, ok := groupBy.GetTagNameOk(); ok {
					groupByTf.TagName = types.StringValue(*tagName)
				}

				groupByTfItem.GroupBy = &groupByTf
			}
			state.GroupBy = append(state.GroupBy, &groupByTfItem)
		}
	}

	if compute, ok := attributes.GetComputeOk(); ok {

		computeTf := computeModel{}
		if aggregationType, ok := compute.GetAggregationTypeOk(); ok {
			computeTf.AggregationType = types.StringValue(*aggregationType)
		}
		if includePercentiles, ok := compute.GetIncludePercentilesOk(); ok {
			computeTf.IncludePercentiles = types.BoolValue(*includePercentiles)
		}
		if path, ok := compute.GetPathOk(); ok {
			computeTf.Path = types.StringValue(*path)
		}

		state.Compute = &computeTf
	}

	if filter, ok := attributes.GetFilterOk(); ok {

		filterTf := filterModel{}
		if query, ok := filter.GetQueryOk(); ok {
			filterTf.Query = types.StringValue(*query)
		}

		state.Filter = &filterTf
	}

	if uniqueness, ok := attributes.GetUniquenessOk(); ok {

		uniquenessTf := uniquenessModel{}
		if when, ok := uniqueness.GetWhenOk(); ok {
			uniquenessTf.When = types.StringValue(*when)
		}

		state.Uniqueness = &uniquenessTf
	}
}

func (r *rumMetricResource) buildRumMetricRequestBody(ctx context.Context, state *rumMetricModel) (*datadogV2.RumMetricCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRumMetricCreateAttributesWithDefaults()

	attributes.SetEventType(state.EventType.ValueString())

	if state.GroupBy != nil {
		var groupBy []datadogV2.RumMetricGroupBy
		for _, groupByTFItem := range state.GroupBy {
			groupByDDItem := datadogV2.NewRumMetricGroupBy()

			groupByDDItem.SetPath(groupByTFItem.Path.ValueString())
			if !groupByTFItem.TagName.IsNull() {
				groupByDDItem.SetTagName(groupByTFItem.TagName.ValueString())
			}
		}
		attributes.SetGroupBy(groupBy)
	}

	var compute datadogV2.RumMetricCompute

	compute.SetAggregationType(state.Compute.AggregationType.ValueString())
	if !state.Compute.IncludePercentiles.IsNull() {
		compute.SetIncludePercentiles(state.Compute.IncludePercentiles.ValueBool())
	}
	if !state.Compute.Path.IsNull() {
		compute.SetPath(state.Compute.Path.ValueString())
	}

	attributes.Compute = compute

	if state.Filter != nil {
		var filter datadogV2.RumMetricFilter

		filter.SetQuery(state.Filter.Query.ValueString())

		attributes.Filter = &filter
	}

	if state.Uniqueness != nil {
		var uniqueness datadogV2.RumMetricUniqueness

		uniqueness.SetWhen(state.Uniqueness.When.ValueString())

		attributes.Uniqueness = &uniqueness
	}

	req := datadogV2.NewRumMetricCreateRequestWithDefaults()
	req.Data = *datadogV2.NewRumMetricCreateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *rumMetricResource) buildRumMetricUpdateRequestBody(ctx context.Context, state *rumMetricModel) (*datadogV2.RumMetricUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRumMetricUpdateAttributesWithDefaults()

	if state.GroupBy != nil {
		var groupBy []datadogV2.RumMetricGroupBy
		for _, groupByTFItem := range state.GroupBy {
			groupByDDItem := datadogV2.NewRumMetricGroupBy()

			groupByDDItem.SetPath(groupByTFItem.Path.ValueString())
			if !groupByTFItem.TagName.IsNull() {
				groupByDDItem.SetTagName(groupByTFItem.TagName.ValueString())
			}
		}
		attributes.SetGroupBy(groupBy)
	}

	if state.Compute != nil {
		var compute datadogV2.RumMetricUpdateCompute

		if !state.Compute.IncludePercentiles.IsNull() {
			compute.SetIncludePercentiles(state.Compute.IncludePercentiles.ValueBool())
		}

		attributes.Compute = &compute
	}

	if state.Filter != nil {
		var filter datadogV2.RumMetricFilter

		filter.SetQuery(state.Filter.Query.ValueString())

		attributes.Filter = &filter
	}

	req := datadogV2.NewRumMetricUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewRumMetricUpdateDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
