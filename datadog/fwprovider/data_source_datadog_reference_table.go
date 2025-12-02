package fwprovider

import (
	"context"
	"fmt"

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
	// Query Parameters (mutually exclusive)
	ID        types.String `tfsdk:"id"`
	TableName types.String `tfsdk:"table_name"`

	// Computed values
	CreatedBy     types.String       `tfsdk:"created_by"`
	Description   types.String       `tfsdk:"description"`
	LastUpdatedBy types.String       `tfsdk:"last_updated_by"`
	RowCount      types.Int64        `tfsdk:"row_count"`
	Source        types.String       `tfsdk:"source"`
	Status        types.String       `tfsdk:"status"`
	UpdatedAt     types.String       `tfsdk:"updated_at"`
	Tags          types.List         `tfsdk:"tags"`
	FileMetadata  *fileMetadataModel `tfsdk:"file_metadata"`
	Schema        *schemaModel       `tfsdk:"schema"`
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
		Description: "Use this data source to retrieve information about an existing Datadog reference table. Query by either table_name or id (mutually exclusive). Supports all source types including cloud storage (S3, GCS, Azure) and external integrations (ServiceNow, Salesforce, Databricks, Snowflake, LOCAL_FILE).",
		Attributes: map[string]schema.Attribute{
			// Query Parameters (mutually exclusive)
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The UUID of the reference table. Either id or table_name must be specified, but not both.",
			},
			"table_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the reference table. Either id or table_name must be specified, but not both.",
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
				Description: "The source type for the reference table (e.g., S3, GCS, AZURE, SERVICENOW, SALESFORCE, DATABRICKS, SNOWFLAKE, LOCAL_FILE).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the reference table (e.g., DONE, PROCESSING, ERROR).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of the last update to the reference table in ISO 8601 format.",
			},
			"tags": schema.ListAttribute{
				Computed:    true,
				Description: "The tags associated with the reference table.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"file_metadata": schema.SingleNestedBlock{
				Description: "File metadata for the reference table. Contains sync settings for cloud storage sources.",
				Attributes: map[string]schema.Attribute{
					"sync_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether automatic sync is enabled for this table. Only present for cloud storage sources (S3, GCS, Azure).",
					},
					"error_message": schema.StringAttribute{
						Computed:    true,
						Description: "Error message from the last sync attempt, if any.",
					},
					"error_row_count": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of rows that failed to sync.",
					},
					"error_type": schema.StringAttribute{
						Computed:    true,
						Description: "The type of error that occurred during file processing. Only present for cloud storage sources.",
					},
				},
				Blocks: map[string]schema.Block{
					"access_details": schema.SingleNestedBlock{
						Description: "Cloud storage access configuration. Only present for cloud storage sources (S3, GCS, Azure).",
						Blocks: map[string]schema.Block{
							"aws_detail": schema.SingleNestedBlock{
								Description: "AWS S3 access configuration.",
								Attributes: map[string]schema.Attribute{
									"aws_account_id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the AWS account.",
									},
									"aws_bucket_name": schema.StringAttribute{
										Computed:    true,
										Description: "The name of the AWS S3 bucket.",
									},
									"file_path": schema.StringAttribute{
										Computed:    true,
										Description: "The relative file path from the S3 bucket root.",
									},
								},
							},
							"gcp_detail": schema.SingleNestedBlock{
								Description: "Google Cloud Storage access configuration.",
								Attributes: map[string]schema.Attribute{
									"gcp_project_id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the GCP project.",
									},
									"gcp_bucket_name": schema.StringAttribute{
										Computed:    true,
										Description: "The name of the GCP bucket.",
									},
									"file_path": schema.StringAttribute{
										Computed:    true,
										Description: "The relative file path from the GCS bucket root.",
									},
									"gcp_service_account_email": schema.StringAttribute{
										Computed:    true,
										Description: "The email of the GCP service account.",
									},
								},
							},
							"azure_detail": schema.SingleNestedBlock{
								Description: "Azure Blob Storage access configuration.",
								Attributes: map[string]schema.Attribute{
									"azure_tenant_id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the Azure tenant.",
									},
									"azure_client_id": schema.StringAttribute{
										Computed:    true,
										Description: "The Azure client ID.",
									},
									"azure_storage_account_name": schema.StringAttribute{
										Computed:    true,
										Description: "The name of the Azure storage account.",
									},
									"azure_container_name": schema.StringAttribute{
										Computed:    true,
										Description: "The name of the Azure container.",
									},
									"file_path": schema.StringAttribute{
										Computed:    true,
										Description: "The relative file path from the Azure container root.",
									},
								},
							},
						},
					},
				},
			},
			"schema": schema.SingleNestedBlock{
				Description: "The schema definition for the reference table.",
				Attributes: map[string]schema.Attribute{
					"primary_keys": schema.ListAttribute{
						Computed:    true,
						Description: "List of field names that serve as primary keys for the table.",
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					"fields": schema.ListNestedBlock{
						Description: "List of fields in the table schema.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The name of the field.",
								},
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "The data type of the field (e.g., STRING, INT32).",
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

	// Validate that exactly one of id or table_name is specified
	hasID := !state.ID.IsNull() && state.ID.ValueString() != ""
	hasTableName := !state.TableName.IsNull() && state.TableName.ValueString() != ""

	if !hasID && !hasTableName {
		response.Diagnostics.AddError(
			"Missing required argument",
			"Either 'id' or 'table_name' must be specified",
		)
		return
	}

	if hasID && hasTableName {
		response.Diagnostics.AddError(
			"Conflicting arguments",
			"Only one of 'id' or 'table_name' can be specified, not both",
		)
		return
	}

	// Query by ID
	if hasID {
		tableId := state.ID.ValueString()
		ddResp, _, err := d.Api.GetTable(d.Auth, tableId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting reference table"))
			return
		}

		d.updateState(ctx, &state, ddResp.Data)
	} else {
		// Query by table_name using list endpoint with exact match
		tableName := state.TableName.ValueString()
		optionalParams := datadogV2.ListTablesOptionalParameters{
			FilterTableNameExact: &tableName,
		}

		ddResp, _, err := d.Api.ListTables(d.Auth, optionalParams)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datadog reference tables"))
			return
		}

		if len(ddResp.Data) == 0 {
			response.Diagnostics.AddError(
				"Reference table not found",
				fmt.Sprintf("No reference table found with table_name='%s'", tableName),
			)
			return
		}

		if len(ddResp.Data) > 1 {
			response.Diagnostics.AddError(
				"Multiple reference tables found",
				fmt.Sprintf("Found %d reference tables with table_name='%s', expected exactly 1", len(ddResp.Data), tableName),
			)
			return
		}

		// Get full details using the ID from list response
		tableId := ddResp.Data[0].GetId()
		fullResp, _, err := d.Api.GetTable(d.Auth, tableId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting reference table"))
			return
		}

		d.updateState(ctx, &state, fullResp.Data)
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogReferenceTableDataSource) updateState(ctx context.Context, state *datadogReferenceTableDataSourceModel, referenceTableData *datadogV2.TableResultV2Data) {
	attributes := referenceTableData.GetAttributes()

	// Set basic attributes
	state.ID = types.StringValue(referenceTableData.GetId())
	state.TableName = types.StringValue(attributes.GetTableName())

	if createdBy, ok := attributes.GetCreatedByOk(); ok {
		state.CreatedBy = types.StringValue(*createdBy)
	}

	if description, ok := attributes.GetDescriptionOk(); ok {
		state.Description = types.StringValue(*description)
	}

	if lastUpdatedBy, ok := attributes.GetLastUpdatedByOk(); ok {
		state.LastUpdatedBy = types.StringValue(*lastUpdatedBy)
	}

	if rowCount, ok := attributes.GetRowCountOk(); ok {
		state.RowCount = types.Int64Value(int64(*rowCount))
	}

	if source, ok := attributes.GetSourceOk(); ok {
		state.Source = types.StringValue(string(*source))
	}

	if status, ok := attributes.GetStatusOk(); ok {
		state.Status = types.StringValue(*status)
	}

	if updatedAt, ok := attributes.GetUpdatedAtOk(); ok {
		state.UpdatedAt = types.StringValue(*updatedAt)
	}

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		state.Tags, _ = types.ListValueFrom(ctx, types.StringType, *tags)
	}

	// Handle FileMetadata from API response (flattened structure, no longer OneOf)
	if fileMetadata, ok := attributes.GetFileMetadataOk(); ok {
		fileMetadataTf := &fileMetadataModel{}

		// FileMetadata is now a flattened struct with direct fields
		if syncEnabled, ok := fileMetadata.GetSyncEnabledOk(); ok {
			fileMetadataTf.SyncEnabled = types.BoolValue(*syncEnabled)
		}

		if errorMessage, ok := fileMetadata.GetErrorMessageOk(); ok {
			fileMetadataTf.ErrorMessage = types.StringValue(*errorMessage)
		}

		if errorRowCount, ok := fileMetadata.GetErrorRowCountOk(); ok {
			fileMetadataTf.ErrorRowCount = types.Int64Value(*errorRowCount)
		}

		if errorType, ok := fileMetadata.GetErrorTypeOk(); ok {
			fileMetadataTf.ErrorType = types.StringValue(string(*errorType))
		}

		// Extract access_details (only present for cloud storage sources)
		if accessDetails, ok := fileMetadata.GetAccessDetailsOk(); ok {
			accessDetailsTf := &accessDetailsModel{}

			// AWS details
			if awsDetail := accessDetails.AwsDetail; awsDetail != nil {
				awsDetailTf := &awsDetailModel{}
				if awsAccountId, ok := awsDetail.GetAwsAccountIdOk(); ok {
					awsDetailTf.AwsAccountId = types.StringValue(*awsAccountId)
				}
				if awsBucketName, ok := awsDetail.GetAwsBucketNameOk(); ok {
					awsDetailTf.AwsBucketName = types.StringValue(*awsBucketName)
				}
				if filePath, ok := awsDetail.GetFilePathOk(); ok {
					awsDetailTf.FilePath = types.StringValue(*filePath)
				}
				accessDetailsTf.AwsDetail = awsDetailTf
			}

			// GCP details
			if gcpDetail := accessDetails.GcpDetail; gcpDetail != nil {
				gcpDetailTf := &gcpDetailModel{}
				if gcpProjectId, ok := gcpDetail.GetGcpProjectIdOk(); ok {
					gcpDetailTf.GcpProjectId = types.StringValue(*gcpProjectId)
				}
				if gcpBucketName, ok := gcpDetail.GetGcpBucketNameOk(); ok {
					gcpDetailTf.GcpBucketName = types.StringValue(*gcpBucketName)
				}
				if filePath, ok := gcpDetail.GetFilePathOk(); ok {
					gcpDetailTf.FilePath = types.StringValue(*filePath)
				}
				if gcpServiceAccountEmail, ok := gcpDetail.GetGcpServiceAccountEmailOk(); ok {
					gcpDetailTf.GcpServiceAccountEmail = types.StringValue(*gcpServiceAccountEmail)
				}
				accessDetailsTf.GcpDetail = gcpDetailTf
			}

			// Azure details
			if azureDetail := accessDetails.AzureDetail; azureDetail != nil {
				azureDetailTf := &azureDetailModel{}
				if azureTenantId, ok := azureDetail.GetAzureTenantIdOk(); ok {
					azureDetailTf.AzureTenantId = types.StringValue(*azureTenantId)
				}
				if azureClientId, ok := azureDetail.GetAzureClientIdOk(); ok {
					azureDetailTf.AzureClientId = types.StringValue(*azureClientId)
				}
				if azureStorageAccountName, ok := azureDetail.GetAzureStorageAccountNameOk(); ok {
					azureDetailTf.AzureStorageAccountName = types.StringValue(*azureStorageAccountName)
				}
				if azureContainerName, ok := azureDetail.GetAzureContainerNameOk(); ok {
					azureDetailTf.AzureContainerName = types.StringValue(*azureContainerName)
				}
				if filePath, ok := azureDetail.GetFilePathOk(); ok {
					azureDetailTf.FilePath = types.StringValue(*filePath)
				}
				accessDetailsTf.AzureDetail = azureDetailTf
			}

			fileMetadataTf.AccessDetails = accessDetailsTf
		}

		state.FileMetadata = fileMetadataTf
	}

	// Handle Schema
	if schema, ok := attributes.GetSchemaOk(); ok {
		schemaTf := &schemaModel{}

		if primaryKeys, ok := schema.GetPrimaryKeysOk(); ok && len(*primaryKeys) > 0 {
			schemaTf.PrimaryKeys, _ = types.ListValueFrom(ctx, types.StringType, *primaryKeys)
		}

		if fields, ok := schema.GetFieldsOk(); ok && len(*fields) > 0 {
			schemaTf.Fields = []*fieldsModel{}
			for _, fieldDd := range *fields {
				fieldTf := &fieldsModel{}
				if name, ok := fieldDd.GetNameOk(); ok {
					fieldTf.Name = types.StringValue(*name)
				}
				if typeVar, ok := fieldDd.GetTypeOk(); ok {
					fieldTf.Type = types.StringValue(string(*typeVar))
				}
				schemaTf.Fields = append(schemaTf.Fields, fieldTf)
			}
		}

		state.Schema = schemaTf
	}
}
