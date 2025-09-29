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
	_ resource.ResourceWithConfigure   = &awsCurConfigResource{}
	_ resource.ResourceWithImportState = &awsCurConfigResource{}
)

type awsCurConfigResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type awsCurConfigModel struct {
	ID             types.String         `tfsdk:"id"`
	AccountId      types.String         `tfsdk:"account_id"`
	BucketName     types.String         `tfsdk:"bucket_name"`
	BucketRegion   types.String         `tfsdk:"bucket_region"`
	ReportName     types.String         `tfsdk:"report_name"`
	ReportPrefix   types.String         `tfsdk:"report_prefix"`
	AccountFilters *accountFiltersModel `tfsdk:"account_filters"`
}

type accountFiltersModel struct {
	IncludeNewAccounts types.Bool `tfsdk:"include_new_accounts"`
	ExcludedAccounts   types.List `tfsdk:"excluded_accounts"`
	IncludedAccounts   types.List `tfsdk:"included_accounts"`
}

func NewAwsCurConfigResource() resource.Resource {
	return &awsCurConfigResource{}
}

func (r *awsCurConfigResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *awsCurConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cur_config"
}

func (r *awsCurConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AwsCurConfig resource. This can be used to create and manage Datadog aws_cur_config.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "The AWS account ID.",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "The AWS bucket name used to store the Cost and Usage Report.",
			},
			"bucket_region": schema.StringAttribute{
				Optional:    true,
				Description: "The region the bucket is located in.",
			},
			"report_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Cost and Usage Report.",
			},
			"report_prefix": schema.StringAttribute{
				Required:    true,
				Description: "The report prefix used for the Cost and Usage Report.",
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"account_filters": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"include_new_accounts": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether or not to automatically include new member accounts by default in your billing dataset.",
					},
					"excluded_accounts": schema.ListAttribute{
						Optional:    true,
						Description: "The AWS account IDs to be excluded from your billing dataset. This field is used when `include_new_accounts` is `true`.",
						ElementType: types.StringType,
					},
					"included_accounts": schema.ListAttribute{
						Optional:    true,
						Description: "The AWS account IDs to be included in your billing dataset. This field is used when `include_new_accounts` is `false`.",
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (r *awsCurConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *awsCurConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state awsCurConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	cloudAccountId, _ := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	resp, httpResp, err := r.Api.GetCostAWSCURConfig(r.Auth, cloudAccountId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AwsCurConfig"))
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

func (r *awsCurConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state awsCurConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAwsCurConfigRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateCostAWSCURConfig(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AwsCurConfig"))
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

func (r *awsCurConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state awsCurConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildAwsCurConfigUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	cloudAccountId, _ := strconv.ParseInt(id, 10, 64)
	resp, _, err := r.Api.UpdateCostAWSCURConfig(r.Auth, cloudAccountId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating AwsCurConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	// Find the updated config in the response
	var foundConfig *datadogV2.AwsCURConfig
	for _, config := range resp.Data {
		if config.GetId() == id {
			foundConfig = &config
			break
		}
	}
	if foundConfig != nil {
		r.updateStateFromSingleConfig(ctx, &state, foundConfig)
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *awsCurConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state awsCurConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	cloudAccountId, _ := strconv.ParseInt(id, 10, 64)
	httpResp, err := r.Api.DeleteCostAWSCURConfig(r.Auth, cloudAccountId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting aws_cur_config"))
		return
	}
}

func (r *awsCurConfigResource) updateStateFromSingleConfig(ctx context.Context, state *awsCurConfigModel, config *datadogV2.AwsCURConfig) {
	state.ID = types.StringValue(config.GetId())

	if attributes, ok := config.GetAttributesOk(); ok {
		state.AccountId = types.StringValue(attributes.GetAccountId())
		state.BucketName = types.StringValue(attributes.GetBucketName())
		state.BucketRegion = types.StringValue(attributes.GetBucketRegion())
		state.ReportName = types.StringValue(attributes.GetReportName())
		state.ReportPrefix = types.StringValue(attributes.GetReportPrefix())

		// Set AccountFilters if present in API response and was originally specified in config
		if accountFilters, ok := attributes.GetAccountFiltersOk(); ok && state.AccountFilters != nil {
			state.AccountFilters = mapAccountFilters(ctx, accountFilters)
		}
	}
}

func (r *awsCurConfigResource) updateStateFromResponseData(ctx context.Context, state *awsCurConfigModel, config *datadogV2.AwsCurConfigResponseData) {
	state.ID = types.StringValue(config.GetId())

	if attributes, ok := config.GetAttributesOk(); ok {
		state.AccountId = types.StringValue(attributes.GetAccountId())
		state.BucketName = types.StringValue(attributes.GetBucketName())
		state.BucketRegion = types.StringValue(attributes.GetBucketRegion())
		state.ReportName = types.StringValue(attributes.GetReportName())
		state.ReportPrefix = types.StringValue(attributes.GetReportPrefix())

		// Set AccountFilters if present in API response and was originally specified in config
		if accountFilters, ok := attributes.GetAccountFiltersOk(); ok && state.AccountFilters != nil {
			state.AccountFilters = mapAccountFiltersFromResponseData(ctx, accountFilters)
		}
	}
}

func (r *awsCurConfigResource) buildAwsCurConfigRequestBody(ctx context.Context, state *awsCurConfigModel) (*datadogV2.AwsCURConfigPostRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAwsCURConfigPostRequestAttributesWithDefaults()

	if !state.AccountId.IsNull() {
		attributes.SetAccountId(state.AccountId.ValueString())
	}
	if !state.BucketName.IsNull() {
		attributes.SetBucketName(state.BucketName.ValueString())
	}
	if !state.BucketRegion.IsNull() {
		attributes.SetBucketRegion(state.BucketRegion.ValueString())
	}
	if !state.ReportName.IsNull() {
		attributes.SetReportName(state.ReportName.ValueString())
	}
	if !state.ReportPrefix.IsNull() {
		attributes.SetReportPrefix(state.ReportPrefix.ValueString())
	}

	if state.AccountFilters != nil {
		var accountFilters datadogV2.AccountFilteringConfig

		if !state.AccountFilters.IncludeNewAccounts.IsNull() {
			accountFilters.SetIncludeNewAccounts(state.AccountFilters.IncludeNewAccounts.ValueBool())
		}

		if !state.AccountFilters.ExcludedAccounts.IsNull() {
			var excludedAccounts []string
			diags.Append(state.AccountFilters.ExcludedAccounts.ElementsAs(ctx, &excludedAccounts, false)...)
			accountFilters.SetExcludedAccounts(excludedAccounts)
		}

		if !state.AccountFilters.IncludedAccounts.IsNull() {
			var includedAccounts []string
			diags.Append(state.AccountFilters.IncludedAccounts.ElementsAs(ctx, &includedAccounts, false)...)
			accountFilters.SetIncludedAccounts(includedAccounts)
		}
		attributes.AccountFilters = &accountFilters
	}

	req := datadogV2.NewAwsCURConfigPostRequestWithDefaults()
	req.Data = *datadogV2.NewAwsCURConfigPostDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *awsCurConfigResource) buildAwsCurConfigUpdateRequestBody(ctx context.Context, state *awsCurConfigModel) (*datadogV2.AwsCURConfigPatchRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAwsCURConfigPatchRequestAttributesWithDefaults()

	// IsEnabled is not part of the resource model for creation/update in this context
	// It's handled through separate patch operations

	if state.AccountFilters != nil {
		var accountFilters datadogV2.AccountFilteringConfig

		if !state.AccountFilters.IncludeNewAccounts.IsNull() {
			accountFilters.SetIncludeNewAccounts(state.AccountFilters.IncludeNewAccounts.ValueBool())
		}

		if !state.AccountFilters.ExcludedAccounts.IsNull() {
			var excludedAccounts []string
			diags.Append(state.AccountFilters.ExcludedAccounts.ElementsAs(ctx, &excludedAccounts, false)...)
			accountFilters.SetExcludedAccounts(excludedAccounts)
		}

		if !state.AccountFilters.IncludedAccounts.IsNull() {
			var includedAccounts []string
			diags.Append(state.AccountFilters.IncludedAccounts.ElementsAs(ctx, &includedAccounts, false)...)
			accountFilters.SetIncludedAccounts(includedAccounts)
		}
		attributes.AccountFilters = &accountFilters
	}

	req := datadogV2.NewAwsCURConfigPatchRequestWithDefaults()
	req.Data = *datadogV2.NewAwsCURConfigPatchDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

// mapAccountFilters is a helper function to convert API account filters to Terraform model
func mapAccountFilters(ctx context.Context, accountFilters *datadogV2.AccountFilteringConfig) *accountFiltersModel {
	model := &accountFiltersModel{}

	// Handle include_new_accounts
	if includeNew, ok := accountFilters.GetIncludeNewAccountsOk(); ok {
		if includeNew != nil {
			model.IncludeNewAccounts = types.BoolValue(*includeNew)
		} else {
			model.IncludeNewAccounts = types.BoolNull()
		}
	}

	// Handle excluded_accounts list
	if excluded, ok := accountFilters.GetExcludedAccountsOk(); ok && excluded != nil && len(*excluded) > 0 {
		model.ExcludedAccounts, _ = types.ListValueFrom(ctx, types.StringType, *excluded)
	} else {
		model.ExcludedAccounts = types.ListNull(types.StringType)
	}

	// Handle included_accounts list
	if included, ok := accountFilters.GetIncludedAccountsOk(); ok && included != nil && len(*included) > 0 {
		model.IncludedAccounts, _ = types.ListValueFrom(ctx, types.StringType, *included)
	} else {
		model.IncludedAccounts = types.ListNull(types.StringType)
	}

	return model
}

// mapAccountFiltersFromResponseData is a helper function to convert API account filters from response data to Terraform model
func mapAccountFiltersFromResponseData(ctx context.Context, accountFilters *datadogV2.AwsCurConfigResponseDataAttributesAccountFilters) *accountFiltersModel {
	model := &accountFiltersModel{}

	// Handle include_new_accounts
	if includeNew, ok := accountFilters.GetIncludeNewAccountsOk(); ok {
		if includeNew != nil {
			model.IncludeNewAccounts = types.BoolValue(*includeNew)
		} else {
			model.IncludeNewAccounts = types.BoolNull()
		}
	}

	// Handle excluded_accounts list
	if excluded, ok := accountFilters.GetExcludedAccountsOk(); ok && excluded != nil && len(*excluded) > 0 {
		model.ExcludedAccounts, _ = types.ListValueFrom(ctx, types.StringType, *excluded)
	} else {
		model.ExcludedAccounts = types.ListNull(types.StringType)
	}

	// Handle included_accounts list
	if included, ok := accountFilters.GetIncludedAccountsOk(); ok && included != nil && len(*included) > 0 {
		model.IncludedAccounts, _ = types.ListValueFrom(ctx, types.StringType, *included)
	} else {
		model.IncludedAccounts = types.ListNull(types.StringType)
	}

	return model
}

// accountFiltersHasData checks if account filters contains meaningful data
func accountFiltersHasData(accountFilters *datadogV2.AccountFilteringConfig) bool {
	if accountFilters == nil {
		return false
	}

	// Check if include_new_accounts is set
	if includeNew, ok := accountFilters.GetIncludeNewAccountsOk(); ok && includeNew != nil {
		return true
	}

	// Check if excluded_accounts has data
	if excluded, ok := accountFilters.GetExcludedAccountsOk(); ok && excluded != nil && len(*excluded) > 0 {
		return true
	}

	// Check if included_accounts has data
	if included, ok := accountFilters.GetIncludedAccountsOk(); ok && included != nil && len(*included) > 0 {
		return true
	}

	return false
}

// accountFiltersFromResponseDataHasData checks if response data account filters contains meaningful data
func accountFiltersFromResponseDataHasData(accountFilters *datadogV2.AwsCurConfigResponseDataAttributesAccountFilters) bool {
	if accountFilters == nil {
		return false
	}

	// Check if include_new_accounts is set
	if includeNew, ok := accountFilters.GetIncludeNewAccountsOk(); ok && includeNew != nil {
		return true
	}

	// Check if excluded_accounts has data
	if excluded, ok := accountFilters.GetExcludedAccountsOk(); ok && excluded != nil && len(*excluded) > 0 {
		return true
	}

	// Check if included_accounts has data
	if included, ok := accountFilters.GetIncludedAccountsOk(); ok && included != nil && len(*included) > 0 {
		return true
	}

	return false
}
