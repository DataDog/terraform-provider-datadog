package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &derivedMetricResource{}
	_ resource.ResourceWithImportState = &derivedMetricResource{}
)

type derivedMetricResource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

type derivedMetricModel struct {
	ID      types.String              `tfsdk:"id"`
	Name    types.String              `tfsdk:"name"`
	Formula types.String              `tfsdk:"formula"`
	Queries []derivedMetricQueryModel `tfsdk:"queries"`
}

type derivedMetricQueryModel struct {
	Name       types.String `tfsdk:"name"`
	DataSource types.String `tfsdk:"data_source"`
	Query      types.String `tfsdk:"query"`
}

func NewDerivedMetricResource() resource.Resource {
	return &derivedMetricResource{}
}

func (r *derivedMetricResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	r.Auth = providerData.Auth
}

func (r *derivedMetricResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "derived_metric"
}

func (r *derivedMetricResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog derived metric resource. A derived metric is computed from a formula over one or more metric queries. Managing derived metrics requires the `Manage Tags for Metrics` permission.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the derived metric. This field can't be updated after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"formula": schema.StringAttribute{
				Description: "The formula that combines the queries into the value of the derived metric. Queries are referenced by their `name`.",
				Required:    true,
			},
			"queries": schema.ListNestedAttribute{
				Description: "The list of queries referenced by the formula.",
				Required:    true,
				Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name used to reference this query from the formula.",
							Required:    true,
						},
						"data_source": schema.StringAttribute{
							Description: "The data source that the query is run against, for example `metrics`.",
							Required:    true,
						},
						"query": schema.StringAttribute{
							Description: "The query string executed against the data source.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *derivedMetricResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *derivedMetricResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state derivedMetricModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetDerivedMetric(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving derived metric"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *derivedMetricResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state derivedMetricModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildDerivedMetricRequestBody(&state)

	resp, _, err := r.Api.CreateDerivedMetric(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating derived metric"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *derivedMetricResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state derivedMetricModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Name.ValueString()
	body := r.buildDerivedMetricRequestBody(&state)

	resp, _, err := r.Api.UpdateDerivedMetric(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating derived metric"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *derivedMetricResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state derivedMetricModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteDerivedMetric(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting derived metric"))
		return
	}
}

func (r *derivedMetricResource) updateState(state *derivedMetricModel, resp *datadogV2.DerivedMetricResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())
	state.Name = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	definition := attributes.GetDefinition()

	state.Formula = types.StringValue(definition.GetFormula())

	state.Queries = []derivedMetricQueryModel{}
	for _, query := range definition.GetQueries() {
		state.Queries = append(state.Queries, derivedMetricQueryModel{
			Name:       types.StringValue(query.GetName()),
			DataSource: types.StringValue(query.GetDataSource()),
			Query:      types.StringValue(query.GetQuery()),
		})
	}
}

// The API uses the same request body for both create and update.
func (r *derivedMetricResource) buildDerivedMetricRequestBody(state *derivedMetricModel) *datadogV2.DerivedMetricCreateRequest {
	queries := []datadogV2.DerivedMetricQuery{}
	for _, query := range state.Queries {
		queries = append(queries, *datadogV2.NewDerivedMetricQuery(
			query.DataSource.ValueString(),
			query.Name.ValueString(),
			query.Query.ValueString(),
		))
	}

	definition := datadogV2.NewDerivedMetricDefinition(state.Formula.ValueString(), queries)
	attributes := datadogV2.NewDerivedMetricAttributes(*definition)

	req := datadogV2.NewDerivedMetricCreateRequestWithDefaults()
	req.Data = *datadogV2.NewDerivedMetricCreateDataWithDefaults()
	req.Data.SetId(state.Name.ValueString())
	req.Data.SetAttributes(*attributes)

	return req
}
