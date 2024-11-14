package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	ID         types.String              `tfsdk:"id"`
	Name       types.String              `tfsdk:"name"`
	EventType  types.String              `tfsdk:"event_type"`
	GroupBy    []*rumMetricGroupByModel  `tfsdk:"group_by"`
	Compute    *rumMetricComputeModel    `tfsdk:"compute"`
	Filter     *rumMetricFilterModel     `tfsdk:"filter"`
	Uniqueness *rumMetricUniquenessModel `tfsdk:"uniqueness"`
}

type rumMetricGroupByModel struct {
	Path    types.String `tfsdk:"path"`
	TagName types.String `tfsdk:"tag_name"`
}

type rumMetricComputeModel struct {
	AggregationType    types.String `tfsdk:"aggregation_type"`
	IncludePercentiles types.Bool   `tfsdk:"include_percentiles"`
	Path               types.String `tfsdk:"path"`
}

type rumMetricFilterModel struct {
	Query types.String `tfsdk:"query"`
}

type rumMetricUniquenessModel struct {
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
			"name": schema.StringAttribute{
				Description: "The name of the RUM-based metric. This field can't be updated after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"event_type": schema.StringAttribute{
				Description: "The type of RUM events to filter on.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"compute": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"aggregation_type": schema.StringAttribute{
						Description: "The type of aggregation to use.",
						Required:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"include_percentiles": schema.BoolAttribute{
						Description: "Toggle to include or exclude percentile aggregations for distribution metrics. Only present when `aggregation_type` is `distribution`.",
						Optional:    true,
					},
					"path": schema.StringAttribute{
						Description: "The path to the value the RUM-based metric will aggregate on. Only present when `aggregation_type` is `distribution`.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"filter": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Description: "The search query. Follows RUM search syntax.",
						Optional:    true,
					},
				},
			},
			"group_by": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Description: "The path to the value the RUM-based metric will be aggregated over.",
							Optional:    true,
						},
						"tag_name": schema.StringAttribute{
							Description: "Name of the tag that gets created. By default, `path` is used as the tag name.",
							Optional:    true,
						},
					},
				},
			},
			"uniqueness": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"when": schema.StringAttribute{
						Description: "When to count updatable events. `match` when the event is first seen, or `end` when the event is complete.",
						Optional:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
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

	id := state.Name.ValueString()

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
	state.Name = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.EventType = types.StringValue(string(attributes.GetEventType()))

	if compute, ok := attributes.GetComputeOk(); ok {

		computeTf := rumMetricComputeModel{}
		if aggregationType, ok := compute.GetAggregationTypeOk(); ok {
			computeTf.AggregationType = types.StringValue(string(*aggregationType))
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

		filterTf := rumMetricFilterModel{}
		if query, ok := filter.GetQueryOk(); ok {
			filterTf.Query = types.StringValue(*query)
		}

		state.Filter = &filterTf
	}

	if groupBy, ok := attributes.GetGroupByOk(); ok && len(*groupBy) > 0 {
		state.GroupBy = []*rumMetricGroupByModel{}
		for _, groupByDdItem := range *groupBy {
			groupByTfItem := rumMetricGroupByModel{}
			if path, ok := groupByDdItem.GetPathOk(); ok {
				groupByTfItem.Path = types.StringValue(*path)
			}
			if tagName, ok := groupByDdItem.GetTagNameOk(); ok {
				groupByTfItem.TagName = types.StringValue(*tagName)
			}

			state.GroupBy = append(state.GroupBy, &groupByTfItem)
		}
	}

	if uniqueness, ok := attributes.GetUniquenessOk(); ok {

		uniquenessTf := rumMetricUniquenessModel{}
		if when, ok := uniqueness.GetWhenOk(); ok {
			uniquenessTf.When = types.StringValue(string(*when))
		}

		state.Uniqueness = &uniquenessTf
	}
}

func (r *rumMetricResource) buildRumMetricRequestBody(ctx context.Context, state *rumMetricModel) (*datadogV2.RumMetricCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRumMetricCreateAttributesWithDefaults()

	attributes.SetEventType(datadogV2.RumMetricEventType(state.EventType.ValueString()))

	if state.GroupBy != nil {
		var groupBy []datadogV2.RumMetricGroupBy
		for _, groupByTFItem := range state.GroupBy {
			groupByDDItem := datadogV2.NewRumMetricGroupBy(groupByTFItem.Path.ValueString())

			if !groupByTFItem.TagName.IsNull() {
				groupByDDItem.SetTagName(groupByTFItem.TagName.ValueString())
			}
			groupBy = append(groupBy, *groupByDDItem)
		}
		attributes.SetGroupBy(groupBy)
	}

	var compute datadogV2.RumMetricCompute

	compute.SetAggregationType(datadogV2.RumMetricComputeAggregationType(state.Compute.AggregationType.ValueString()))
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

		uniqueness.SetWhen(datadogV2.RumMetricUniquenessWhen(state.Uniqueness.When.ValueString()))

		attributes.Uniqueness = &uniqueness
	}

	req := datadogV2.NewRumMetricCreateRequestWithDefaults()
	req.Data = *datadogV2.NewRumMetricCreateDataWithDefaults()
	req.Data.SetId(state.Name.ValueString())
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *rumMetricResource) buildRumMetricUpdateRequestBody(ctx context.Context, state *rumMetricModel) (*datadogV2.RumMetricUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRumMetricUpdateAttributesWithDefaults()

	if state.GroupBy != nil {
		var groupBy []datadogV2.RumMetricGroupBy
		for _, groupByTFItem := range state.GroupBy {
			groupByDDItem := datadogV2.NewRumMetricGroupBy(groupByTFItem.Path.ValueString())

			if !groupByTFItem.TagName.IsNull() {
				groupByDDItem.SetTagName(groupByTFItem.TagName.ValueString())
			}

			groupBy = append(groupBy, *groupByDDItem)
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
	req.Data.SetId(state.Name.ValueString())
	req.Data.SetAttributes(*attributes)

	return req, diags
}
