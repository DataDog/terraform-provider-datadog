package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &gcpUcConfigResource{}
	_ resource.ResourceWithImportState = &gcpUcConfigResource{}
)

type gcpUcConfigResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type gcpUcConfigModel struct {
	ID                types.String `tfsdk:"id"`
	BillingAccountId  types.String `tfsdk:"billing_account_id"`
	BucketName        types.String `tfsdk:"bucket_name"`
	ExportDatasetName types.String `tfsdk:"export_dataset_name"`
	ExportPrefix      types.String `tfsdk:"export_prefix"`
	ExportProjectName types.String `tfsdk:"export_project_name"`
	ServiceAccount    types.String `tfsdk:"service_account"`
}

func NewGcpUcConfigResource() resource.Resource {
	return &gcpUcConfigResource{}
}

func (r *gcpUcConfigResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *gcpUcConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "gcp_uc_config"
}

func (r *gcpUcConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog GcpUcConfig resource. This can be used to create and manage Datadog gcp_uc_config.",
		Attributes: map[string]schema.Attribute{
			"billing_account_id": schema.StringAttribute{
				Required:    true,
				Description: "The Google Cloud account ID.",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The Google Cloud bucket name used to store the Usage Cost export.",
			},
			"export_dataset_name": schema.StringAttribute{
				Required:    true,
				Description: "The export dataset name used for the Google Cloud Usage Cost report.",
			},
			"export_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "The export prefix used for the Google Cloud Usage Cost report.",
			},
			"export_project_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Google Cloud Usage Cost report.",
			},
			"service_account": schema.StringAttribute{
				Required:    true,
				Description: "The unique Google Cloud service account email.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *gcpUcConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *gcpUcConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state gcpUcConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, _ := strconv.ParseInt(state.ID.ValueString(), 10, 64)

	resp, httpResp, err := r.Api.GetCostGCPUsageCostConfig(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving GcpUcConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	responseData := resp.GetData()
	r.updateStateFromGcpUcConfigResponseData(ctx, &state, &responseData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *gcpUcConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state gcpUcConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildGcpUcConfigRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateCostGCPUsageCostConfig(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving GcpUcConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	responseData := resp.GetData()
	r.updateStateFromResponseData(ctx, &state, &responseData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *gcpUcConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state gcpUcConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildGcpUcConfigUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	idInt, _ := strconv.ParseInt(id, 10, 64)
	resp, _, err := r.Api.UpdateCostGCPUsageCostConfig(r.Auth, idInt, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving GcpUcConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	responseData := resp.GetData()
	r.updateStateFromResponseData(ctx, &state, &responseData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *gcpUcConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state gcpUcConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, _ := strconv.ParseInt(state.ID.ValueString(), 10, 64)

	httpResp, err := r.Api.DeleteCostGCPUsageCostConfig(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting gcp_uc_config"))
		return
	}
}

func (r *gcpUcConfigResource) updateStateFromResponseData(ctx context.Context, state *gcpUcConfigModel, responseData *datadogV2.GCPUsageCostConfig) {
	state.ID = types.StringValue(responseData.GetId())

	if attributes, ok := responseData.GetAttributesOk(); ok {
		state.BillingAccountId = types.StringValue(attributes.GetAccountId())
		state.BucketName = types.StringValue(attributes.GetBucketName())
		state.ExportDatasetName = types.StringValue(attributes.GetDataset())
		state.ExportPrefix = types.StringValue(attributes.GetExportPrefix())
		state.ExportProjectName = types.StringValue(attributes.GetExportProjectName())
		state.ServiceAccount = types.StringValue(attributes.GetServiceAccount())
	}
}

func (r *gcpUcConfigResource) updateStateFromGcpUcConfigResponseData(ctx context.Context, state *gcpUcConfigModel, responseData *datadogV2.GcpUcConfigResponseData) {
	state.ID = types.StringValue(responseData.GetId())

	if attributes, ok := responseData.GetAttributesOk(); ok {
		state.BillingAccountId = types.StringValue(attributes.GetAccountId())
		state.BucketName = types.StringValue(attributes.GetBucketName())
		state.ExportDatasetName = types.StringValue(attributes.GetDataset())
		state.ExportPrefix = types.StringValue(attributes.GetExportPrefix())
		state.ExportProjectName = types.StringValue(attributes.GetExportProjectName())
		state.ServiceAccount = types.StringValue(attributes.GetServiceAccount())
	}
}

func (r *gcpUcConfigResource) buildGcpUcConfigRequestBody(ctx context.Context, state *gcpUcConfigModel) (*datadogV2.GCPUsageCostConfigPostRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewGCPUsageCostConfigPostRequestAttributesWithDefaults()

	if !state.BillingAccountId.IsNull() {
		attributes.SetBillingAccountId(state.BillingAccountId.ValueString())
	}
	if !state.BucketName.IsNull() {
		attributes.SetBucketName(state.BucketName.ValueString())
	}
	if !state.ExportDatasetName.IsNull() {
		attributes.SetExportDatasetName(state.ExportDatasetName.ValueString())
	}
	if !state.ExportPrefix.IsNull() {
		attributes.SetExportPrefix(state.ExportPrefix.ValueString())
	}
	if !state.ExportProjectName.IsNull() {
		attributes.SetExportProjectName(state.ExportProjectName.ValueString())
	}
	if !state.ServiceAccount.IsNull() {
		attributes.SetServiceAccount(state.ServiceAccount.ValueString())
	}

	req := datadogV2.NewGCPUsageCostConfigPostRequestWithDefaults()
	req.Data = *datadogV2.NewGCPUsageCostConfigPostDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *gcpUcConfigResource) buildGcpUcConfigUpdateRequestBody(ctx context.Context, state *gcpUcConfigModel) (*datadogV2.GCPUsageCostConfigPatchRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewGCPUsageCostConfigPatchRequestAttributesWithDefaults()

	// IsEnabled is not part of the resource model for creation/update in this context
	// It's handled through separate patch operations similar to AWS

	req := datadogV2.NewGCPUsageCostConfigPatchRequestWithDefaults()
	req.Data = *datadogV2.NewGCPUsageCostConfigPatchDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
