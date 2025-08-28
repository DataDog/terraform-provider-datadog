package fwprovider

import (
	"context"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &DatasetResource{}
	_ resource.ResourceWithImportState = &DatasetResource{}
)

type DatasetResource struct {
	API  *datadogV2.DatasetsApi
	Auth context.Context
}

type DatasetModel struct {
	ID             types.String           `tfsdk:"id"`
	Name           types.String           `tfsdk:"name"`
	Principals     types.Set              `tfsdk:"principals"`
	ProductFilters []*ProductFiltersModel `tfsdk:"product_filters"`
	CreatedAt      timetypes.RFC3339      `tfsdk:"created_at"`
	CreatedBy      types.String           `tfsdk:"created_by"`
}

type ProductFiltersModel struct {
	Product types.String `tfsdk:"product"`
	Filters types.Set    `tfsdk:"filters"`
}

func NewDatasetResource() resource.Resource {
	return &DatasetResource{}
}

func (r *DatasetResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetDatasetsApiV2()
	r.Auth = providerData.Auth
}

func (r *DatasetResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "dataset"
}

func (r *DatasetResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Dataset resource. This can be used to create and manage Datadog datasets.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the dataset.",
				Required:    true,
			},
			"principals": schema.SetAttribute{
				Description: "An array of principals. A principal is a subject or group of subjects. Each principal is formatted as `type:id`. Supported types: `role` and `team`.",
				ElementType: types.StringType,
				Required:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Indicates when the dataset was created (in ISO 8601).",
				CustomType:  timetypes.RFC3339Type{},
				Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "Indicates who created the dataset.",
				Computed:    true,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"product_filters": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"product": schema.StringAttribute{
							Description: "The product type of the dataset. Supported types: `apm`, `rum`, `synthetics`, `metrics`, `logs`, `sd_repoinfo`, `error_tracking`, `cloud_cost`, and `ml_obs`.",
							Required:    true,
						},
						"filters": schema.SetAttribute{
							Description: "A list of tag-based filters used to restrict access to the product type. Each filter is formatted as `@tag.key:value`.",
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *DatasetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *DatasetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data DatasetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	resp, httpResp, err := r.API.GetDataset(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving dataset"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp)
}

func (r *DatasetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data DatasetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildCreateDatasetRequestBody(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.CreateDataset(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating dataset"))
		return
	}
	if err = utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data DatasetModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	datasetId := data.ID.ValueString()
	body, diags := r.buildUpdateDatasetRequestBody(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateDataset(r.Auth, datasetId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating dataset"))
		return
	}

	if err = utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &data, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data DatasetModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	httpResp, err := r.API.DeleteDataset(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting dataset"))
	}

}

func (r *DatasetResource) updateState(ctx context.Context, state *DatasetModel, response *datadogV2.DatasetResponseSingle) {
	state.ID = types.StringValue(response.Data.GetId())
	attributes := response.Data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	} else {
		state.Name = types.StringValue("")
	}

	if principals, ok := attributes.GetPrincipalsOk(); ok {
		state.Principals, _ = types.SetValueFrom(ctx, types.StringType, principals)
	} else {
		state.Principals, _ = types.SetValue(types.StringType, []attr.Value{})
	}

	if productFilters, ok := attributes.GetProductFiltersOk(); ok {
		state.ProductFilters = []*ProductFiltersModel{}
		for _, filter := range *productFilters {
			tfProductFilter := ProductFiltersModel{}
			tfProductFilter.Product = types.StringValue(filter.Product)
			tfProductFilter.Filters, _ = types.SetValueFrom(ctx, types.StringType, filter.Filters)
			state.ProductFilters = append(state.ProductFilters, &tfProductFilter)
		}
	}

	if createdAt, ok := attributes.GetCreatedAtOk(); ok && !createdAt.IsZero() {
		state.CreatedAt = timetypes.NewRFC3339TimeValue(*createdAt)
	} else {
		state.CreatedAt = timetypes.NewRFC3339Null()
	}

	if createdBy, ok := attributes.GetCreatedByOk(); ok {
		state.CreatedBy = types.StringValue(createdBy.String())
	} else {
		state.CreatedBy = types.StringValue("")
	}
}

func (r *DatasetResource) buildCreateDatasetRequestBody(ctx context.Context, data *DatasetModel) (*datadogV2.DatasetCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	body := datadogV2.NewDatasetRequestWithDefaults()
	attributes := *datadogV2.NewDatasetAttributesRequestWithDefaults()

	attributes.Name = data.Name.ValueString()
	var principals []string
	diags.Append(data.Principals.ElementsAs(ctx, &principals, false)...)
	attributes.Principals = principals
	ddProductFilters := []datadogV2.FiltersPerProduct{}
	for _, product := range data.ProductFilters {
		productFilters := []string{}
		diags.Append(product.Filters.ElementsAs(ctx, &productFilters, false)...)
		ddFilter := datadogV2.FiltersPerProduct{
			Product: product.Product.ValueString(),
			Filters: productFilters,
		}
		ddProductFilters = append(ddProductFilters, ddFilter)
	}

	attributes.ProductFilters = ddProductFilters
	body.SetAttributes(attributes)
	req := datadogV2.NewDatasetCreateRequest(*body)

	return req, diags
}

func (r *DatasetResource) buildUpdateDatasetRequestBody(ctx context.Context, data *DatasetModel) (*datadogV2.DatasetUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	body := datadogV2.NewDatasetRequestWithDefaults()
	attributes := *datadogV2.NewDatasetAttributesRequestWithDefaults()

	attributes.Name = data.Name.ValueString()
	var principals []string
	diags.Append(data.Principals.ElementsAs(ctx, &principals, false)...)
	attributes.Principals = principals
	ddProductFilters := []datadogV2.FiltersPerProduct{}
	for _, product := range data.ProductFilters {
		productFilters := []string{}
		diags.Append(product.Filters.ElementsAs(ctx, &productFilters, false)...)
		ddFilter := datadogV2.FiltersPerProduct{
			Product: product.Product.ValueString(),
			Filters: productFilters,
		}
		ddProductFilters = append(ddProductFilters, ddFilter)
	}
	attributes.ProductFilters = ddProductFilters

	body.SetAttributes(attributes)
	req := datadogV2.NewDatasetUpdateRequest(*body)

	return req, diags
}
