package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogAzureUcConfigDataSource{}
)

type datadogAzureUcConfigDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type datadogAzureUcConfigDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	CloudAccountId types.Int64 `tfsdk:"cloud_account_id"`

	// Computed values
	AccountId           types.String              `tfsdk:"account_id"`
	ClientId            types.String              `tfsdk:"client_id"`
	Scope               types.String              `tfsdk:"scope"`
	Status              types.String              `tfsdk:"status"`
	CreatedAt           types.String              `tfsdk:"created_at"`
	ActualBillConfig    *actualBillConfigModel    `tfsdk:"actual_bill_config"`
	AmortizedBillConfig *amortizedBillConfigModel `tfsdk:"amortized_bill_config"`
}

func NewDatadogAzureUcConfigDataSource() datasource.DataSource {
	return &datadogAzureUcConfigDataSource{}
}

func (d *datadogAzureUcConfigDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogAzureUcConfigDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "azure_uc_config"
}

func (d *datadogAzureUcConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific Datadog Azure Usage Cost configuration. This allows you to fetch details about an existing Cloud Cost Management configuration for Azure billing data access.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"cloud_account_id": schema.Int64Attribute{
				Required:    true,
				Description: "The Datadog cloud account ID for the Azure Usage Cost configuration you want to retrieve information about.",
			},
			// Computed values
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The tenant ID of the Azure account.",
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "The client ID of the Azure account.",
			},
			"scope": schema.StringAttribute{
				Computed:    true,
				Description: "The scope of your observed subscription.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the Azure Usage Cost configuration.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the Azure Usage Cost configuration was created.",
			},
		},
		Blocks: map[string]schema.Block{
			"actual_bill_config": schema.SingleNestedBlock{
				Description: "Configuration for the actual cost export.",
				Attributes: map[string]schema.Attribute{
					"export_name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the configured Azure Export.",
					},
					"export_path": schema.StringAttribute{
						Computed:    true,
						Description: "The path where the Azure Export is saved.",
					},
					"storage_account": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the storage account where the Azure Export is saved.",
					},
					"storage_container": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the storage container where the Azure Export is saved.",
					},
				},
			},
			"amortized_bill_config": schema.SingleNestedBlock{
				Description: "Configuration for the amortized cost export.",
				Attributes: map[string]schema.Attribute{
					"export_name": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the configured Azure Export.",
					},
					"export_path": schema.StringAttribute{
						Computed:    true,
						Description: "The path where the Azure Export is saved.",
					},
					"storage_account": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the storage account where the Azure Export is saved.",
					},
					"storage_container": schema.StringAttribute{
						Computed:    true,
						Description: "The name of the storage container where the Azure Export is saved.",
					},
				},
			},
		},
	}
}

func (d *datadogAzureUcConfigDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogAzureUcConfigDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	cloudAccountId := state.CloudAccountId.ValueInt64()
	ddResp, _, err := d.Api.GetCostAzureUCConfig(d.Auth, cloudAccountId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog Azure UC config"))
		return
	}

	response.Diagnostics.Append(d.updateState(ctx, &state, &ddResp)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogAzureUcConfigDataSource) updateState(ctx context.Context, state *datadogAzureUcConfigDataSourceModel, ucConfigPair *datadogV2.UCConfigPair) diag.Diagnostics {
	var diags diag.Diagnostics
	if data, ok := ucConfigPair.GetDataOk(); ok {
		if id, ok := data.GetIdOk(); ok {
			state.ID = types.StringValue(*id)
		}

		if attributes, ok := data.GetAttributesOk(); ok {
			if configs, ok := attributes.GetConfigsOk(); ok && len(*configs) > 0 {
				configs := *configs

				// Set shared attributes from first config
				firstConfig := configs[0]
				state.AccountId = types.StringValue(firstConfig.GetAccountId())
				state.ClientId = types.StringValue(firstConfig.GetClientId())
				state.Scope = types.StringValue(firstConfig.GetScope())
				state.Status = types.StringValue(firstConfig.GetStatus())
				state.CreatedAt = types.StringValue(firstConfig.GetCreatedAt())

				// Separate configs by dataset_type and populate respective blocks
				for _, configData := range configs {
					datasetType := configData.GetDatasetType()
					switch datasetType {
					case "actual":
						state.ActualBillConfig = &actualBillConfigModel{}
						state.ActualBillConfig.ExportName = types.StringValue(configData.GetExportName())
						state.ActualBillConfig.ExportPath = types.StringValue(configData.GetExportPath())
						state.ActualBillConfig.StorageAccount = types.StringValue(configData.GetStorageAccount())
						state.ActualBillConfig.StorageContainer = types.StringValue(configData.GetStorageContainer())
					case "amortized":
						state.AmortizedBillConfig = &amortizedBillConfigModel{}
						state.AmortizedBillConfig.ExportName = types.StringValue(configData.GetExportName())
						state.AmortizedBillConfig.ExportPath = types.StringValue(configData.GetExportPath())
						state.AmortizedBillConfig.StorageAccount = types.StringValue(configData.GetStorageAccount())
						state.AmortizedBillConfig.StorageContainer = types.StringValue(configData.GetStorageContainer())
					default:
						diags.AddError(
							"Unexpected dataset type",
							fmt.Sprintf("Received unexpected dataset type '%s'. Expected 'actual' or 'amortized'.", datasetType),
						)
						return diags
					}
				}
			}
		}
	}
	return diags
}
