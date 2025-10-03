package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithImportState = &azureUcConfigResource{}
)

type azureUcConfigResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type azureUcConfigModel struct {
	ID                  types.String              `tfsdk:"id"`
	AccountId           types.String              `tfsdk:"account_id"`
	ClientId            types.String              `tfsdk:"client_id"`
	Scope               types.String              `tfsdk:"scope"`
	ActualBillConfig    *actualBillConfigModel    `tfsdk:"actual_bill_config"`
	AmortizedBillConfig *amortizedBillConfigModel `tfsdk:"amortized_bill_config"`
}

type actualBillConfigModel struct {
	ExportName       types.String `tfsdk:"export_name"`
	ExportPath       types.String `tfsdk:"export_path"`
	StorageAccount   types.String `tfsdk:"storage_account"`
	StorageContainer types.String `tfsdk:"storage_container"`
}

type amortizedBillConfigModel struct {
	ExportName       types.String `tfsdk:"export_name"`
	ExportPath       types.String `tfsdk:"export_path"`
	StorageAccount   types.String `tfsdk:"storage_account"`
	StorageContainer types.String `tfsdk:"storage_container"`
}

func NewAzureUcConfigResource() resource.Resource {
	return &azureUcConfigResource{}
}

func (r *azureUcConfigResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *azureUcConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "azure_uc_config"
}

func (r *azureUcConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Azure Usage Cost configuration resource. This can be used to create and manage Azure Cost Export configurations for Cloud Cost Management. Azure configurations require both actual and amortized cost export settings.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:      true,
				Description:   "The tenant ID of the Azure account.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"client_id": schema.StringAttribute{
				Required:      true,
				Description:   "The client ID of the Azure account.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"scope": schema.StringAttribute{
				Required:      true,
				Description:   "The scope of your observed subscription.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"actual_bill_config": schema.SingleNestedBlock{
				Description:   "Configuration for the actual cost export.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"export_name": schema.StringAttribute{
						Required:      true,
						Description:   "The name of the configured Azure Export.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"export_path": schema.StringAttribute{
						Required:      true,
						Description:   "The path where the Azure Export is saved.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"storage_account": schema.StringAttribute{
						Required:      true,
						Description:   "The name of the storage account where the Azure Export is saved.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"storage_container": schema.StringAttribute{
						Required:      true,
						Description:   "The name of the storage container where the Azure Export is saved.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
				},
			},
			"amortized_bill_config": schema.SingleNestedBlock{
				Description:   "Configuration for the amortized cost export.",
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"export_name": schema.StringAttribute{
						Required:      true,
						Description:   "The name of the configured Azure Export.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"export_path": schema.StringAttribute{
						Required:      true,
						Description:   "The path where the Azure Export is saved.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"storage_account": schema.StringAttribute{
						Required:      true,
						Description:   "The name of the storage account where the Azure Export is saved.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"storage_container": schema.StringAttribute{
						Required:      true,
						Description:   "The name of the storage container where the Azure Export is saved.",
						PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
				},
			},
		},
	}
}

func (r *azureUcConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *azureUcConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state azureUcConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, _ := strconv.ParseInt(state.ID.ValueString(), 10, 64)

	resp, httpResp, err := r.Api.GetCostAzureUCConfig(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving AzureUcConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateStateFromUCConfigPair(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *azureUcConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state azureUcConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildAzureUcConfigRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateCostAzureUCConfigs(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating AzureUcConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateStateFromAzureUCConfigPairsResponse(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *azureUcConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError(
		"Update Not Supported",
		"Azure UC Config resources do not support updates. Changes require resource recreation.",
	)
}

func (r *azureUcConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state azureUcConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, _ := strconv.ParseInt(state.ID.ValueString(), 10, 64)

	httpResp, err := r.Api.DeleteCostAzureUCConfig(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting azure_uc_config"))
		return
	}
}

func (r *azureUcConfigResource) updateStateFromUCConfigPair(ctx context.Context, state *azureUcConfigModel, resp *datadogV2.UCConfigPair) {
	if data, ok := resp.GetDataOk(); ok {
		if id, ok := data.GetIdOk(); ok {
			state.ID = types.StringValue(*id)
		}

		if attributes, ok := data.GetAttributesOk(); ok {
			if configs, ok := attributes.GetConfigsOk(); ok && len(*configs) > 0 {
				// Populate basic fields from first config
				firstConfig := (*configs)[0]
				if accountId, ok := firstConfig.GetAccountIdOk(); ok {
					state.AccountId = types.StringValue(*accountId)
				}
				if clientId, ok := firstConfig.GetClientIdOk(); ok {
					state.ClientId = types.StringValue(*clientId)
				}
				if scope, ok := firstConfig.GetScopeOk(); ok {
					state.Scope = types.StringValue(*scope)
				}

				// Separate configs by dataset_type and populate bill config blocks
				for _, configData := range *configs {
					datasetType := configData.GetDatasetType()

					if datasetType == "actual" {
						state.ActualBillConfig = &actualBillConfigModel{}
						if exportName, ok := configData.GetExportNameOk(); ok {
							state.ActualBillConfig.ExportName = types.StringValue(*exportName)
						}
						if exportPath, ok := configData.GetExportPathOk(); ok {
							state.ActualBillConfig.ExportPath = types.StringValue(*exportPath)
						}
						if storageAccount, ok := configData.GetStorageAccountOk(); ok {
							state.ActualBillConfig.StorageAccount = types.StringValue(*storageAccount)
						}
						if storageContainer, ok := configData.GetStorageContainerOk(); ok {
							state.ActualBillConfig.StorageContainer = types.StringValue(*storageContainer)
						}
					} else if datasetType == "amortized" {
						state.AmortizedBillConfig = &amortizedBillConfigModel{}
						if exportName, ok := configData.GetExportNameOk(); ok {
							state.AmortizedBillConfig.ExportName = types.StringValue(*exportName)
						}
						if exportPath, ok := configData.GetExportPathOk(); ok {
							state.AmortizedBillConfig.ExportPath = types.StringValue(*exportPath)
						}
						if storageAccount, ok := configData.GetStorageAccountOk(); ok {
							state.AmortizedBillConfig.StorageAccount = types.StringValue(*storageAccount)
						}
						if storageContainer, ok := configData.GetStorageContainerOk(); ok {
							state.AmortizedBillConfig.StorageContainer = types.StringValue(*storageContainer)
						}
					}
				}
			}
		}
	}
}

func (r *azureUcConfigResource) updateStateFromAzureUCConfigPairsResponse(ctx context.Context, state *azureUcConfigModel, resp *datadogV2.AzureUCConfigPairsResponse) {
	if data, ok := resp.GetDataOk(); ok {
		// Convert AzureUCConfigPair to UCConfigPair format for consistency
		if id, ok := data.GetIdOk(); ok {
			state.ID = types.StringValue(*id)
		}

		if attributes := data.GetAttributes(); len(attributes.GetConfigs()) > 0 {
			configs := attributes.GetConfigs()

			// Populate basic fields from first config
			firstConfig := configs[0]
			if accountId, ok := firstConfig.GetAccountIdOk(); ok {
				state.AccountId = types.StringValue(*accountId)
			}
			if clientId, ok := firstConfig.GetClientIdOk(); ok {
				state.ClientId = types.StringValue(*clientId)
			}
			if scope, ok := firstConfig.GetScopeOk(); ok {
				state.Scope = types.StringValue(*scope)
			}

			// Separate configs by dataset_type and populate bill config blocks
			for _, configData := range configs {
				datasetType := configData.GetDatasetType()

				if datasetType == "actual" {
					state.ActualBillConfig = &actualBillConfigModel{}
					if exportName, ok := configData.GetExportNameOk(); ok {
						state.ActualBillConfig.ExportName = types.StringValue(*exportName)
					}
					if exportPath, ok := configData.GetExportPathOk(); ok {
						state.ActualBillConfig.ExportPath = types.StringValue(*exportPath)
					}
					if storageAccount, ok := configData.GetStorageAccountOk(); ok {
						state.ActualBillConfig.StorageAccount = types.StringValue(*storageAccount)
					}
					if storageContainer, ok := configData.GetStorageContainerOk(); ok {
						state.ActualBillConfig.StorageContainer = types.StringValue(*storageContainer)
					}
				} else if datasetType == "amortized" {
					state.AmortizedBillConfig = &amortizedBillConfigModel{}
					if exportName, ok := configData.GetExportNameOk(); ok {
						state.AmortizedBillConfig.ExportName = types.StringValue(*exportName)
					}
					if exportPath, ok := configData.GetExportPathOk(); ok {
						state.AmortizedBillConfig.ExportPath = types.StringValue(*exportPath)
					}
					if storageAccount, ok := configData.GetStorageAccountOk(); ok {
						state.AmortizedBillConfig.StorageAccount = types.StringValue(*storageAccount)
					}
					if storageContainer, ok := configData.GetStorageContainerOk(); ok {
						state.AmortizedBillConfig.StorageContainer = types.StringValue(*storageContainer)
					}
				}
			}
		}
	}
}

func (r *azureUcConfigResource) buildAzureUcConfigRequestBody(ctx context.Context, state *azureUcConfigModel) (*datadogV2.AzureUCConfigPostRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAzureUCConfigPostRequestAttributesWithDefaults()

	if !state.AccountId.IsNull() {
		attributes.SetAccountId(state.AccountId.ValueString())
	}
	if !state.ClientId.IsNull() {
		attributes.SetClientId(state.ClientId.ValueString())
	}
	if !state.Scope.IsNull() {
		attributes.SetScope(state.Scope.ValueString())
	}

	if state.ActualBillConfig != nil {
		var actualBillConfig datadogV2.BillConfig
		if !state.ActualBillConfig.ExportName.IsNull() {
			actualBillConfig.SetExportName(state.ActualBillConfig.ExportName.ValueString())
		}
		if !state.ActualBillConfig.ExportPath.IsNull() {
			actualBillConfig.SetExportPath(state.ActualBillConfig.ExportPath.ValueString())
		}
		if !state.ActualBillConfig.StorageAccount.IsNull() {
			actualBillConfig.SetStorageAccount(state.ActualBillConfig.StorageAccount.ValueString())
		}
		if !state.ActualBillConfig.StorageContainer.IsNull() {
			actualBillConfig.SetStorageContainer(state.ActualBillConfig.StorageContainer.ValueString())
		}
		attributes.SetActualBillConfig(actualBillConfig)
	}

	if state.AmortizedBillConfig != nil {
		var amortizedBillConfig datadogV2.BillConfig
		if !state.AmortizedBillConfig.ExportName.IsNull() {
			amortizedBillConfig.SetExportName(state.AmortizedBillConfig.ExportName.ValueString())
		}
		if !state.AmortizedBillConfig.ExportPath.IsNull() {
			amortizedBillConfig.SetExportPath(state.AmortizedBillConfig.ExportPath.ValueString())
		}
		if !state.AmortizedBillConfig.StorageAccount.IsNull() {
			amortizedBillConfig.SetStorageAccount(state.AmortizedBillConfig.StorageAccount.ValueString())
		}
		if !state.AmortizedBillConfig.StorageContainer.IsNull() {
			amortizedBillConfig.SetStorageContainer(state.AmortizedBillConfig.StorageContainer.ValueString())
		}
		attributes.SetAmortizedBillConfig(amortizedBillConfig)
	}

	req := datadogV2.NewAzureUCConfigPostRequestWithDefaults()
	req.Data = *datadogV2.NewAzureUCConfigPostDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
