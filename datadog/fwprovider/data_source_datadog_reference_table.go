package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogReferenceTableDataSource{}
)

type datadogReferenceTableDataSource struct {
	Api  *datadogV2.ReferenceTablesApi
	Auth context.Context
}

type datadogReferenceTableDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	FilterStatus            types.String `tfsdk:"filter[status]"`
	FilterTableNameExact    types.String `tfsdk:"filter[table_name][exact]"`
	FilterTableNameContains types.String `tfsdk:"filter[table_name][contains]"`

	// Computed values
	CreatedBy     types.String                       `tfsdk:"created_by"`
	Description   types.String                       `tfsdk:"description"`
	LastUpdatedBy types.String                       `tfsdk:"last_updated_by"`
	RowCount      types.Int64                        `tfsdk:"row_count"`
	Source        datadogV2.ReferenceTableSourceType `tfsdk:"source"`
	Status        types.String                       `tfsdk:"status"`
	TableName     types.String                       `tfsdk:"table_name"`
	UpdatedAt     types.String                       `tfsdk:"updated_at"`
	Tags          types.List                         `tfsdk:"tags"`
	FileMetadata  *fileMetadataModel                 `tfsdk:"file_metadata"`
	Schema        *schemaModel                       `tfsdk:"schema"`
}

type tableResultV2DataAttributesFileMetadataCloudStorageModel struct {
	ErrorMessage  types.String        `tfsdk:"error_message"`
	ErrorRowCount types.Int64         `tfsdk:"error_row_count"`
	ErrorType     types.String        `tfsdk:"error_type"`
	SyncEnabled   types.Bool          `tfsdk:"sync_enabled"`
	AccessDetails *accessDetailsModel `tfsdk:"access_details"`
}
type accessDetailsModel struct {
	AwsDetail   *awsDetailModel   `tfsdk:"aws_detail"`
	AzureDetail *azureDetailModel `tfsdk:"azure_detail"`
	GcpDetail   *gcpDetailModel   `tfsdk:"gcp_detail"`
}
type awsDetailModel struct {
	AwsAccountId  types.String `tfsdk:"aws_account_id"`
	AwsBucketName types.String `tfsdk:"aws_bucket_name"`
	FilePath      types.String `tfsdk:"file_path"`
}
type azureDetailModel struct {
	AzureClientId           types.String `tfsdk:"azure_client_id"`
	AzureContainerName      types.String `tfsdk:"azure_container_name"`
	AzureStorageAccountName types.String `tfsdk:"azure_storage_account_name"`
	AzureTenantId           types.String `tfsdk:"azure_tenant_id"`
	FilePath                types.String `tfsdk:"file_path"`
}
type gcpDetailModel struct {
	FilePath               types.String `tfsdk:"file_path"`
	GcpBucketName          types.String `tfsdk:"gcp_bucket_name"`
	GcpProjectId           types.String `tfsdk:"gcp_project_id"`
	GcpServiceAccountEmail types.String `tfsdk:"gcp_service_account_email"`
}
type tableResultV2DataAttributesFileMetadataLocalFileModel struct {
	ErrorMessage  types.String `tfsdk:"error_message"`
	ErrorRowCount types.Int64  `tfsdk:"error_row_count"`
	UploadId      types.String `tfsdk:"upload_id"`
}

type schemaModel struct {
	PrimaryKeys types.List     `tfsdk:"primary_keys"`
	Fields      []*fieldsModel `tfsdk:"fields"`
}
type fieldsModel struct {
	Name types.String                            `tfsdk:"name"`
	Type datadogV2.ReferenceTableSchemaFieldType `tfsdk:"type"`
}

func NewDatadogReferenceTableDataSource() datasource.DataSource {
	return &datadogReferenceTableDataSource{}
}

func (d *datadogReferenceTableDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetReferenceTablesApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogReferenceTableDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "reference_table"
}

func (d *datadogReferenceTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog reference_table.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"filter[status]": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by processing status.",
			},
			"filter[table_name][exact]": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by table name exactly.",
			},
			"filter[table_name][contains]": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by table name contains.",
			},
			// Computed values
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the user who created the reference table.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the reference table.",
			},
			"last_updated_by": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the user who last updated the reference table.",
			},
			"row_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of successfully processed rows in the reference table.",
			},
			"source": schema.StringAttribute{
				Computed:    true,
				Description: "The source type for reference table data. Includes all possible source types that can appear in responses.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the reference table.",
			},
			"table_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the reference table.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of the last update to the reference table in ISO 8601 format.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				Description: "The tags of the reference table.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			// Computed values
			"file_metadata": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"table_result_v2_data_attributes_file_metadata_cloud_storage": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"error_message": schema.StringAttribute{
								Computed:    true,
								Description: "The error message returned from the sync.",
							},
							"error_row_count": schema.Int64Attribute{
								Computed:    true,
								Description: "The number of rows that failed to sync.",
							},
							"error_type": schema.StringAttribute{
								Computed:    true,
								Description: "The type of error that occurred during file processing. This field provides high-level error categories for easier troubleshooting and is only present when there are errors.",
							},
							"sync_enabled": schema.BoolAttribute{
								Computed:    true,
								Description: "Whether this table is synced automatically.",
							},
						},
						Blocks: map[string]schema.Block{
							"access_details": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{},
								Blocks: map[string]schema.Block{
									"aws_detail": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"aws_account_id": schema.StringAttribute{
												Computed:    true,
												Description: "The ID of the AWS account.",
											},
											"aws_bucket_name": schema.StringAttribute{
												Computed:    true,
												Description: "The name of the AWS bucket.",
											},
											"file_path": schema.StringAttribute{
												Computed:    true,
												Description: "The relative file path from the S3 bucket root to the CSV file.",
											},
										},
									},
									"azure_detail": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"azure_client_id": schema.StringAttribute{
												Computed:    true,
												Description: "The Azure client ID.",
											},
											"azure_container_name": schema.StringAttribute{
												Computed:    true,
												Description: "The name of the Azure container.",
											},
											"azure_storage_account_name": schema.StringAttribute{
												Computed:    true,
												Description: "The name of the Azure storage account.",
											},
											"azure_tenant_id": schema.StringAttribute{
												Computed:    true,
												Description: "The ID of the Azure tenant.",
											},
											"file_path": schema.StringAttribute{
												Computed:    true,
												Description: "The relative file path from the Azure container root to the CSV file.",
											},
										},
									},
									"gcp_detail": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"file_path": schema.StringAttribute{
												Computed:    true,
												Description: "The relative file path from the GCS bucket root to the CSV file.",
											},
											"gcp_bucket_name": schema.StringAttribute{
												Computed:    true,
												Description: "The name of the GCP bucket.",
											},
											"gcp_project_id": schema.StringAttribute{
												Computed:    true,
												Description: "The ID of the GCP project.",
											},
											"gcp_service_account_email": schema.StringAttribute{
												Computed:    true,
												Description: "The email of the GCP service account.",
											},
										},
									},
								},
							},
						},
					},
					"table_result_v2_data_attributes_file_metadata_local_file": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"error_message": schema.StringAttribute{
								Computed:    true,
								Description: "The error message returned from the creation/update.",
							},
							"error_row_count": schema.Int64Attribute{
								Computed:    true,
								Description: "The number of rows that failed to create/update.",
							},
							"upload_id": schema.StringAttribute{
								Computed:    true,
								Description: "The upload ID that was used to create/update the table.",
							},
						},
					},
				},
			},
			"schema": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"primary_keys": schema.ListAttribute{
						Computed:    true,
						Description: "List of field names that serve as primary keys for the table. Only one primary key is supported, and it is used as an ID to retrieve rows.",
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					"fields": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The field name.",
								},
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "The field type for reference table schema fields.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *datadogReferenceTableDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogReferenceTableDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		tableId := state.ID.ValueString()
		ddResp, _, err := d.Api.GetTable(d.Auth, tableId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog referenceTable"))
			return
		}

		d.updateState(ctx, &state, ddResp.Data)
	} else {
		filterStatus := state.FilterStatus.ValueString()
		filterTableNameExact := state.FilterTableNameExact.ValueString()
		filterTableNameContains := state.FilterTableNameContains.ValueString()

		optionalParams := datadogV2.ListTablesOptionalParameters{
			FilterStatus:            &filterStatus,
			FilterTableNameExact:    &filterTableNameExact,
			FilterTableNameContains: &filterTableNameContains,
		}

		ddResp, _, err := d.Api.ListTables(d.Auth, optionalParams)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datadog referenceTable"))
			return
		}

		if len(ddResp.Data) > 1 {
			response.Diagnostics.AddError("filters returned more than one result, use more specific search criteria", "")
			return
		}
		if len(ddResp.Data) == 0 {
			response.Diagnostics.AddError("filters returned no results", "")
			return
		}

		d.updateStateFromListResponse(ctx, &state, &ddResp.Data[0])
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogReferenceTableDataSource) updateState(ctx context.Context, state *datadogReferenceTableDataSourceModel, referenceTableData *datadogV2.TableResultV2Data) {
	state.ID = types.StringValue(referenceTableData.GetId())

	attributes := referenceTableData.GetAttributes()
	state.CreatedBy = types.StringValue(attributes.GetCreatedBy())
	state.Description = types.StringValue(attributes.GetDescription())
	// cloud tables
	state.FileMetadata.SyncEnabled = types.BoolValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetSyncEnabled())
	state.FileMetadata.AccessDetails = &accessDetailsModel{
		AwsDetail: &awsDetailModel{
			AwsAccountId:  types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AwsDetail.GetAwsAccountId()),
			AwsBucketName: types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AwsDetail.GetAwsBucketName()),
			FilePath:      types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AwsDetail.GetFilePath()),
		},
		AzureDetail: &azureDetailModel{
			AzureClientId:           types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AzureDetail.GetAzureClientId()),
			AzureContainerName:      types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AzureDetail.GetAzureContainerName()),
			AzureStorageAccountName: types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AzureDetail.GetAzureStorageAccountName()),
			AzureTenantId:           types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AzureDetail.GetAzureTenantId()),
			FilePath:                types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().AzureDetail.GetFilePath()),
		},
		GcpDetail: &gcpDetailModel{
			FilePath:               types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().GcpDetail.GetFilePath()),
			GcpBucketName:          types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().GcpDetail.GetGcpBucketName()),
			GcpProjectId:           types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().GcpDetail.GetGcpProjectId()),
			GcpServiceAccountEmail: types.StringValue(attributes.GetFileMetadata().TableResultV2DataAttributesFileMetadataCloudStorage.GetAccessDetails().GcpDetail.GetGcpServiceAccountEmail()),
		},
	}
	// for local tables, file path is not stored (upload_id is used by the backend instead)
	state.LastUpdatedBy = types.StringValue(attributes.GetLastUpdatedBy())
	state.RowCount = types.Int64Value(int64(attributes.GetRowCount()))
	primaryKeys := make([]string, len(attributes.GetSchema().PrimaryKeys))
	for i, primaryKey := range attributes.GetSchema().PrimaryKeys {
		primaryKeys[i] = primaryKey
	}
	primaryKeysList, _ := types.ListValueFrom(ctx, types.StringType, primaryKeys)
	state.Schema = &schemaModel{
		PrimaryKeys: primaryKeysList,
		Fields:      make([]*fieldsModel, len(attributes.GetSchema().Fields)),
	}
	for i, field := range attributes.GetSchema().Fields {
		state.Schema.Fields[i] = &fieldsModel{
			Name: types.StringValue(field.GetName()),
			Type: field.GetType(),
		}
	}
	state.Source = attributes.GetSource()
	state.Status = types.StringValue(attributes.GetStatus())
	state.TableName = types.StringValue(attributes.GetTableName())
	state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt())
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, attributes.GetTags())

}

func (d *datadogReferenceTableDataSource) updateStateFromListResponse(ctx context.Context, state *datadogReferenceTableDataSourceModel, referenceTableData *datadogV2.ReferenceTable) {
	state.ID = types.StringValue(referenceTableData.GetId())
	state.Id = types.StringValue(referenceTableData.GetId())

	attributes := referenceTableData.GetAttributes()
	state.CreatedBy = types.StringValue(attributes.GetCreatedBy())
	state.Description = types.StringValue(attributes.GetDescription())
	state.LastUpdatedBy = types.StringValue(attributes.GetLastUpdatedBy())
	state.RowCount = types.Int64Value(int64(attributes.GetRowCount()))
	state.Source = types.StringValue(attributes.GetSource())
	state.Status = types.StringValue(attributes.GetStatus())
	state.TableName = types.StringValue(attributes.GetTableName())
	state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt())
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, attributes.GetTags())
	state.FileMetadata = types.StringValue(attributes.GetFileMetadata())
	state.Schema = types.BlockValue(attributes.GetSchema())
}
