package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	// Computed fields
	CreatedAt       types.String `tfsdk:"created_at"`
	Dataset         types.String `tfsdk:"dataset"`
	Months          types.Int64  `tfsdk:"months"`
	Status          types.String `tfsdk:"status"`
	StatusUpdatedAt types.String `tfsdk:"status_updated_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
	ErrorMessages   types.List   `tfsdk:"error_messages"`
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
				Required:      true,
				Description:   "The Google Cloud account ID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"bucket_name": schema.StringAttribute{
				Required:      true,
				Description:   "The Google Cloud bucket name used to store the Usage Cost export.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"export_dataset_name": schema.StringAttribute{
				Required:      true,
				Description:   "The export dataset name used for the Google Cloud Usage Cost report.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"export_prefix": schema.StringAttribute{
				Optional:      true,
				Description:   "The export prefix used for the Google Cloud Usage Cost report.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"export_project_name": schema.StringAttribute{
				Required:      true,
				Description:   "The name of the Google Cloud Usage Cost report.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"service_account": schema.StringAttribute{
				Required:      true,
				Description:   "The unique Google Cloud service account email.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": utils.ResourceIDAttribute(),
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the GCP UC configuration was created.",
			},
			"dataset": schema.StringAttribute{
				Computed:    true,
				Description: "The dataset name used for the GCP Usage Cost export.",
			},
			"months": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of months of usage data to include in the export.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the GCP UC configuration.",
			},
			"status_updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the configuration status was last updated.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the GCP UC configuration was last modified.",
			},
			"error_messages": schema.ListAttribute{
				Computed:    true,
				Description: "List of error messages if the GCP UC configuration encountered any issues during setup or data processing.",
				ElementType: types.StringType,
			},
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
	response.Diagnostics.AddError(
		"Update Not Supported",
		"GCP UC Config resources do not support updates. Changes require resource recreation.",
	)
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

		// Set computed fields
		state.CreatedAt = types.StringValue(attributes.GetCreatedAt())
		state.Dataset = types.StringValue(attributes.GetDataset())
		state.Months = types.Int64Value(int64(attributes.GetMonths()))
		state.Status = types.StringValue(attributes.GetStatus())
		state.StatusUpdatedAt = types.StringValue(attributes.GetStatusUpdatedAt())
		state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt())
		if errorMessages, ok := attributes.GetErrorMessagesOk(); ok && errorMessages != nil {
			state.ErrorMessages, _ = types.ListValueFrom(ctx, types.StringType, *errorMessages)
		} else {
			state.ErrorMessages = types.ListValueMust(types.StringType, []attr.Value{})
		}
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

		// Set computed fields
		state.CreatedAt = types.StringValue(attributes.GetCreatedAt())
		state.Dataset = types.StringValue(attributes.GetDataset())
		state.Months = types.Int64Value(int64(attributes.GetMonths()))
		state.Status = types.StringValue(attributes.GetStatus())
		state.StatusUpdatedAt = types.StringValue(attributes.GetStatusUpdatedAt())
		state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt())
		if errorMessages, ok := attributes.GetErrorMessagesOk(); ok && errorMessages != nil {
			state.ErrorMessages, _ = types.ListValueFrom(ctx, types.StringType, *errorMessages)
		} else {
			state.ErrorMessages = types.ListValueMust(types.StringType, []attr.Value{})
		}
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
