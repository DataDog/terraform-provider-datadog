package fwprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &ApmRetentionFilterResource{}
	_ resource.ResourceWithImportState = &ApmRetentionFilterResource{}
)

type ApmRetentionFilterResource struct {
	Api  *datadogV2.APMRetentionFiltersApi
	Auth context.Context
}

type ApmRetentionFilterModel struct {
	ID         types.String          `tfsdk:"id"`
	Name       types.String          `tfsdk:"name"`
	Rate       types.String          `tfsdk:"rate"`
	Enabled    types.Bool            `tfsdk:"enabled"`
	FilterType types.String          `tfsdk:"filter_type"`
	Filter     *retentionFilterModel `tfsdk:"filter"`
}

type retentionFilterModel struct {
	Query types.String `tfsdk:"query"`
}

func NewApmRetentionFilterResource() resource.Resource {
	return &ApmRetentionFilterResource{}
}

const spansRetentionFilterType = "spans-sampling-processor"

func (r *ApmRetentionFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetApmRetentionFiltersApiV2()
	r.Auth = providerData.Auth
}

func (r *ApmRetentionFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "apm_retention_filter"
}

func (r *ApmRetentionFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "The object describing the configuration of the retention filter to create/update.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the retention filter.",
				Required:    true,
			},
			"id": utils.ResourceIDAttribute(),
			"enabled": schema.BoolAttribute{
				Description: "the status of the retention filter.",
				Required:    true,
			},
			"filter_type": schema.StringAttribute{
				Description: "The type of the retention filter, currently only spans-processing-sampling is available.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.OneOf(spansRetentionFilterType)},
			},
			"rate": schema.StringAttribute{
				Description: "Sample rate to apply to spans going through this retention filter as a string, a value of 1.0 keeps all spans matching the query.",
				Required:    true,
				Validators:  []validator.String{validators.Float64Between(0, 1)}},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Description: "The spans filter. Spans matching this filter will be indexed and stored.",
				Attributes: map[string]schema.Attribute{
					"query": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The search query - following the span search syntax.",
						Default:     stringdefault.StaticString("*"),
					},
				},
				Validators: []validator.Object{objectvalidator.IsRequired()},
			},
		},
	}
}

func (r *ApmRetentionFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *ApmRetentionFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ApmRetentionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetApmRetentionFilter(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving retention filter"))
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

func (r *ApmRetentionFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state ApmRetentionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildRetentionFilterCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateApmRetentionFilter(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving spans metric"))
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

func (r *ApmRetentionFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state ApmRetentionFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildApmRetentionFilterUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateApmRetentionFilter(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving spans metric"))
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

func (r *ApmRetentionFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ApmRetentionFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteApmRetentionFilter(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting spans metric"))
		return
	}
}

func (r *ApmRetentionFilterResource) updateState(ctx context.Context, state *ApmRetentionFilterModel, resp *datadogV2.RetentionFilterResponse) {
	state.ID = types.StringValue(resp.Data.GetId())
	state.Name = types.StringValue(resp.Data.Attributes.GetName())
	state.Rate = types.StringValue(strconv.FormatFloat(resp.Data.Attributes.GetRate(), 'f', -1, 64))
	state.Filter.Query = types.StringValue(*resp.Data.Attributes.GetFilter().Query)
	state.Enabled = types.BoolValue(*resp.Data.Attributes.Enabled)
}

func (r *ApmRetentionFilterResource) buildRetentionFilterCreateRequestBody(ctx context.Context, state *ApmRetentionFilterModel) (*datadogV2.RetentionFilterCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	attributes := datadogV2.NewRetentionFilterCreateAttributesWithDefaults()
	attributes.SetName(state.Name.ValueString())
	attributes.SetEnabled(state.Enabled.ValueBool())
	attributes.SetFilterType(datadogV2.RetentionFilterType(state.FilterType.ValueString()))
	fValue, err := strconv.ParseFloat(state.Rate.ValueString(), 64)
	if err != nil {
		diags.AddError("rate", fmt.Sprintf("error parsing rate: %s", err))
	}
	attributes.SetRate(fValue)
	attributes.Filter.Query = state.Filter.Query.ValueString()

	req := datadogV2.NewRetentionFilterCreateRequestWithDefaults()
	req.Data = *datadogV2.NewRetentionFilterCreateDataWithDefaults()
	req.Data.SetType("apm_retention_filter")
	req.Data.SetAttributes(*attributes)
	return req, diags
}

func (r *ApmRetentionFilterResource) buildApmRetentionFilterUpdateRequestBody(ctx context.Context, state *ApmRetentionFilterModel) (*datadogV2.RetentionFilterUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewRetentionFilterCreateAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	attributes.SetFilterType(datadogV2.RetentionFilterType(state.FilterType.ValueString()))
	fValue, err := strconv.ParseFloat(state.Rate.ValueString(), 64)
	if err != nil {
		diags.AddError("rate", fmt.Sprintf("error parsing rate: %s", err))
	}
	attributes.SetRate(fValue)
	attributes.SetEnabled(state.Enabled.ValueBool())
	attributes.Filter.Query = state.Filter.Query.ValueString()

	req := datadogV2.NewRetentionFilterUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewRetentionFilterUpdateDataWithDefaults()
	req.Data.SetType("apm_retention_filter")
	req.Data.SetAttributes(*attributes)
	return req, diags
}
