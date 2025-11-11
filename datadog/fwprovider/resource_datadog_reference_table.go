package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &referenceTableResource{}
	_ resource.ResourceWithImportState = &referenceTableResource{}
	_ resource.ResourceWithModifyPlan  = &referenceTableResource{}
)

type referenceTableResource struct {
	Api  *datadogV2.ReferenceTablesApi
	Auth context.Context
}

type referenceTableModel struct {
	ID            types.String       `tfsdk:"id"`
	Source        types.String       `tfsdk:"source"`
	TableName     types.String       `tfsdk:"table_name"`
	FileMetadata  *fileMetadataModel `tfsdk:"file_metadata"`
	Schema        *schemaModel       `tfsdk:"schema"`
	CreatedBy     types.String       `tfsdk:"created_by"`
	LastUpdatedBy types.String       `tfsdk:"last_updated_by"`
	RowCount      types.Int64        `tfsdk:"row_count"`
	Status        types.String       `tfsdk:"status"`
	UpdatedAt     types.String       `tfsdk:"updated_at"`
	Tags          types.List         `tfsdk:"tags"`
	Description   types.String       `tfsdk:"description"`
}

type fileMetadataModel struct {
	SyncEnabled   types.Bool          `tfsdk:"sync_enabled"`
	AccessDetails *accessDetailsModel `tfsdk:"access_details"`
}

func NewReferenceTableResource() resource.Resource {
	return &referenceTableResource{}
}

func (r *referenceTableResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetReferenceTablesApiV2()
	r.Auth = providerData.Auth
}

func (r *referenceTableResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "reference_table"
}

func (r *referenceTableResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Reference Table resource for cloud storage sources (S3, GCS, Azure). This can be used to create and manage Datadog reference tables that sync data from cloud storage.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"table_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the reference table. This must be unique within your organization.",
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "The source type for the reference table. Must be one of: S3, GCS, AZURE.",
				Validators: []validator.String{
					stringvalidator.OneOf("S3", "GCS", "AZURE"),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the reference table.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Description: "A list of tags to associate with the reference table.",
				ElementType: types.StringType,
			},
			// Computed attributes
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the user who created the reference table.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated_by": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the user who last updated the reference table.",
			},
			"row_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of successfully processed rows in the reference table.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the reference table (e.g., DONE, PROCESSING, ERROR).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of the last update to the reference table in ISO 8601 format.",
			},
		},
		Blocks: map[string]schema.Block{
			"file_metadata": schema.SingleNestedBlock{
				Description: "Configuration for cloud storage file access and sync settings.",
				Attributes: map[string]schema.Attribute{
					"sync_enabled": schema.BoolAttribute{
						Required:    true,
						Description: "Whether this table should automatically sync with the cloud storage source.",
					},
				},
				Blocks: map[string]schema.Block{
					"access_details": schema.SingleNestedBlock{
						Description: "Cloud storage access configuration. Exactly one of aws_detail, gcp_detail, or azure_detail must be specified.",
						Blocks: map[string]schema.Block{
							"aws_detail": schema.SingleNestedBlock{
								Description: "AWS S3 access configuration. Required when source is S3.",
								Attributes: map[string]schema.Attribute{
									"aws_account_id": schema.StringAttribute{
										Optional:    true,
										Description: "The ID of the AWS account.",
									},
									"aws_bucket_name": schema.StringAttribute{
										Optional:    true,
										Description: "The name of the Amazon S3 bucket.",
									},
									"file_path": schema.StringAttribute{
										Optional:    true,
										Description: "The relative file path from the S3 bucket root to the CSV file.",
									},
								},
							},
							"gcp_detail": schema.SingleNestedBlock{
								Description: "Google Cloud Storage access configuration. Required when source is GCS.",
								Attributes: map[string]schema.Attribute{
									"gcp_project_id": schema.StringAttribute{
										Optional:    true,
										Description: "The ID of the GCP project.",
									},
									"gcp_bucket_name": schema.StringAttribute{
										Optional:    true,
										Description: "The name of the GCP bucket.",
									},
									"file_path": schema.StringAttribute{
										Optional:    true,
										Description: "The relative file path from the GCS bucket root to the CSV file.",
									},
									"gcp_service_account_email": schema.StringAttribute{
										Optional:    true,
										Description: "The email of the GCP service account used to access the bucket.",
									},
								},
							},
							"azure_detail": schema.SingleNestedBlock{
								Description: "Azure Blob Storage access configuration. Required when source is AZURE.",
								Attributes: map[string]schema.Attribute{
									"azure_tenant_id": schema.StringAttribute{
										Optional:    true,
										Description: "The ID of the Azure tenant.",
									},
									"azure_client_id": schema.StringAttribute{
										Optional:    true,
										Description: "The Azure client ID (application ID).",
									},
									"azure_storage_account_name": schema.StringAttribute{
										Optional:    true,
										Description: "The name of the Azure storage account.",
									},
									"azure_container_name": schema.StringAttribute{
										Optional:    true,
										Description: "The name of the Azure container.",
									},
									"file_path": schema.StringAttribute{
										Optional:    true,
										Description: "The relative file path from the Azure container root to the CSV file.",
									},
								},
							},
						},
					},
				},
			},
			"schema": schema.SingleNestedBlock{
				Description: "The schema definition for the reference table, including field definitions and primary keys.",
				Attributes: map[string]schema.Attribute{
					"primary_keys": schema.ListAttribute{
						Required:    true,
						Description: "List of field names that serve as primary keys for the table. Currently only one primary key is supported.",
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					"fields": schema.ListNestedBlock{
						Description: "List of fields in the table schema. Must include at least one field.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "The name of the field.",
								},
								"type": schema.StringAttribute{
									Required:    true,
									Description: "The data type of the field. Must be one of: STRING, INT32.",
									Validators: []validator.String{
										stringvalidator.OneOf("STRING", "INT32"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *referenceTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *referenceTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state referenceTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetTable(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ReferenceTable"))
		return
	}
	// Note: Skipping CheckForUnparsed because file_metadata OneOf always fails to unmarshal
	// due to Go client bug. We handle it manually in updateState.
	// if err := utils.CheckForUnparsed(resp); err != nil {
	// 	response.Diagnostics.AddError("response contains unparsedObject", err.Error())
	// 	return
	// }

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *referenceTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state referenceTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildReferenceTableRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.CreateReferenceTable(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating ReferenceTable"))
		return
	}
	// Note: Skipping CheckForUnparsed because file_metadata OneOf always fails to unmarshal
	// due to Go client bug. We handle it manually in updateState.
	// if err := utils.CheckForUnparsed(resp); err != nil {
	// 	response.Diagnostics.AddError("response contains unparsedObject", err.Error())
	// 	return
	// }

	// If the create response doesn't include data, fetch it with a list+filter request
	if resp.Data == nil {
		if httpResp != nil && httpResp.StatusCode == 201 {
			// Table was created successfully, but response was empty - list tables and find by exact name
			tableName := state.TableName.ValueString()
			listResp, _, listErr := r.Api.ListTables(r.Auth)
			if listErr != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(listErr, "table created but error listing tables"))
				return
			}

			// Find the table by exact name match
			var foundTable *datadogV2.TableResultV2Data
			if listResp.Data != nil {
				for _, table := range listResp.Data {
					if attrs, ok := table.GetAttributesOk(); ok {
						if name, nameOk := attrs.GetTableNameOk(); nameOk && *name == tableName {
							tableCopy := table
							foundTable = &tableCopy
							break
						}
					}
				}
			}

			if foundTable == nil {
				response.Diagnostics.AddError("API Error", fmt.Sprintf("Table %s was created but not found in list", tableName))
				return
			}

			// Get the full table details by ID
			tableID := foundTable.GetId()
			getResp, _, getErr := r.Api.GetTable(r.Auth, tableID)
			if getErr != nil {
				response.Diagnostics.Append(utils.FrameworkErrorDiag(getErr, fmt.Sprintf("table created but error fetching details for ID %s", tableID)))
				return
			}
			resp = getResp
		} else {
			statusCode := 0
			if httpResp != nil {
				statusCode = httpResp.StatusCode
			}
			response.Diagnostics.AddError("API Error", fmt.Sprintf("CreateReferenceTable returned an empty response (HTTP %d).", statusCode))
			return
		}
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *referenceTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state referenceTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildReferenceTableUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := r.Api.UpdateReferenceTable(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ReferenceTable"))
		return
	}

	// Read back the updated resource to get computed fields
	resp, _, err := r.Api.GetTable(r.Auth, id)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading ReferenceTable after update"))
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *referenceTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state referenceTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteTable(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting reference_table"))
		return
	}
}

func (r *referenceTableResource) updateState(ctx context.Context, state *referenceTableModel, resp *datadogV2.TableResultV2) {
	// Check if Data is present
	if resp == nil || resp.Data == nil {
		return
	}

	attributes := resp.Data.GetAttributes()

	state.ID = types.StringValue(*resp.GetData().Id)

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
		state.RowCount = types.Int64Value(*rowCount)
	}

	if source, ok := attributes.GetSourceOk(); ok {
		state.Source = types.StringValue(string(*source))
	}

	if status, ok := attributes.GetStatusOk(); ok {
		state.Status = types.StringValue(*status)
	}

	if tableName, ok := attributes.GetTableNameOk(); ok {
		state.TableName = types.StringValue(*tableName)
	}

	if updatedAt, ok := attributes.GetUpdatedAtOk(); ok {
		state.UpdatedAt = types.StringValue(*updatedAt)
	}

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		state.Tags, _ = types.ListValueFrom(ctx, types.StringType, *tags)
	}

	// Handle FileMetadata from API response (OneOf union type)
	if fileMetadata, ok := attributes.GetFileMetadataOk(); ok {
		fileMetadataTf := &fileMetadataModel{}

		// Handle UnparsedObject case - manually distinguish between CloudStorage and LocalFile
		// The Go client's OneOf unmarshaler fails because both types can match, so we handle it manually.
		if fileMetadata.UnparsedObject != nil {
			if unparsedMap, ok := fileMetadata.UnparsedObject.(map[string]interface{}); ok {
				// Check if it has access_details (CloudStorage) or not (LocalFile)
				if accessDetails, hasAccessDetails := unparsedMap["access_details"]; hasAccessDetails && accessDetails != nil {
					// It's CloudStorage
					if syncEnabled, ok := unparsedMap["sync_enabled"].(bool); ok {
						fileMetadataTf.SyncEnabled = types.BoolValue(syncEnabled)
					}

					// Parse access_details
					if accessDetailsMap, ok := accessDetails.(map[string]interface{}); ok {
						accessDetailsTf := &accessDetailsModel{}

						// Check for AWS details
						if awsDetail, ok := accessDetailsMap["aws_detail"].(map[string]interface{}); ok {
							awsDetailTf := &awsDetailModel{}
							if awsAccountId, ok := awsDetail["aws_account_id"].(string); ok {
								awsDetailTf.AwsAccountId = types.StringValue(awsAccountId)
							}
							if awsBucketName, ok := awsDetail["aws_bucket_name"].(string); ok {
								awsDetailTf.AwsBucketName = types.StringValue(awsBucketName)
							}
							if filePath, ok := awsDetail["file_path"].(string); ok {
								awsDetailTf.FilePath = types.StringValue(filePath)
							}
							accessDetailsTf.AwsDetail = awsDetailTf
						}

						// Check for GCP details
						if gcpDetail, ok := accessDetailsMap["gcp_detail"].(map[string]interface{}); ok {
							gcpDetailTf := &gcpDetailModel{}
							if gcpProjectId, ok := gcpDetail["gcp_project_id"].(string); ok {
								gcpDetailTf.GcpProjectId = types.StringValue(gcpProjectId)
							}
							if gcpBucketName, ok := gcpDetail["gcp_bucket_name"].(string); ok {
								gcpDetailTf.GcpBucketName = types.StringValue(gcpBucketName)
							}
							if filePath, ok := gcpDetail["file_path"].(string); ok {
								gcpDetailTf.FilePath = types.StringValue(filePath)
							}
							if gcpServiceAccountEmail, ok := gcpDetail["gcp_service_account_email"].(string); ok {
								gcpDetailTf.GcpServiceAccountEmail = types.StringValue(gcpServiceAccountEmail)
							}
							accessDetailsTf.GcpDetail = gcpDetailTf
						}

						// Check for Azure details
						if azureDetail, ok := accessDetailsMap["azure_detail"].(map[string]interface{}); ok {
							azureDetailTf := &azureDetailModel{}
							if azureTenantId, ok := azureDetail["azure_tenant_id"].(string); ok {
								azureDetailTf.AzureTenantId = types.StringValue(azureTenantId)
							}
							if azureClientId, ok := azureDetail["azure_client_id"].(string); ok {
								azureDetailTf.AzureClientId = types.StringValue(azureClientId)
							}
							if azureStorageAccountName, ok := azureDetail["azure_storage_account_name"].(string); ok {
								azureDetailTf.AzureStorageAccountName = types.StringValue(azureStorageAccountName)
							}
							if azureContainerName, ok := azureDetail["azure_container_name"].(string); ok {
								azureDetailTf.AzureContainerName = types.StringValue(azureContainerName)
							}
							if filePath, ok := azureDetail["file_path"].(string); ok {
								azureDetailTf.FilePath = types.StringValue(filePath)
							}
							accessDetailsTf.AzureDetail = azureDetailTf
						}

						fileMetadataTf.AccessDetails = accessDetailsTf
					}

					state.FileMetadata = fileMetadataTf
					return // Skip the normal OneOf handling
				}
			}
		}

		// Check if it's CloudStorage type
		if cloudStorage := fileMetadata.TableResultV2DataAttributesFileMetadataCloudStorage; cloudStorage != nil {
			if syncEnabled, ok := cloudStorage.GetSyncEnabledOk(); ok {
				fileMetadataTf.SyncEnabled = types.BoolValue(*syncEnabled)
			}

			// Extract access_details
			if accessDetails, ok := cloudStorage.GetAccessDetailsOk(); ok {
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
		}

		state.FileMetadata = fileMetadataTf
	}

	// Handle Schema
	if schema, ok := attributes.GetSchemaOk(); ok {
		schemaTf := schemaModel{}
		if fields, ok := schema.GetFieldsOk(); ok && len(*fields) > 0 {
			schemaTf.Fields = []*fieldsModel{}
			for _, fieldsDd := range *fields {
				fieldsTf := fieldsModel{}
				if name, ok := fieldsDd.GetNameOk(); ok {
					fieldsTf.Name = types.StringValue(*name)
				}
				if typeVar, ok := fieldsDd.GetTypeOk(); ok {
					fieldsTf.Type = types.StringValue(string(*typeVar))
				}
				schemaTf.Fields = append(schemaTf.Fields, &fieldsTf)
			}
		}
		if primaryKeys, ok := schema.GetPrimaryKeysOk(); ok && len(*primaryKeys) > 0 {
			schemaTf.PrimaryKeys, _ = types.ListValueFrom(ctx, types.StringType, *primaryKeys)
		}
		state.Schema = &schemaTf
	}
}

func (r *referenceTableResource) buildReferenceTableRequestBody(ctx context.Context, state *referenceTableModel) (*datadogV2.CreateTableRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateTableRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	if !state.Source.IsNull() {
		attributes.SetSource(datadogV2.ReferenceTableCreateSourceType(state.Source.ValueString()))
	}
	if !state.TableName.IsNull() {
		attributes.SetTableName(state.TableName.ValueString())
	}

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	// Build file_metadata for cloud storage
	if state.FileMetadata != nil {
		cloudStorageMetadata := datadogV2.CreateTableRequestDataAttributesFileMetadataCloudStorage{}
		cloudStorageMetadata.SetSyncEnabled(state.FileMetadata.SyncEnabled.ValueBool())

		if state.FileMetadata.AccessDetails != nil {
			accessDetails := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetails{}

			// AWS details
			if state.FileMetadata.AccessDetails.AwsDetail != nil {
				awsDetail := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetailsAwsDetail{}
				awsDetail.SetAwsAccountId(state.FileMetadata.AccessDetails.AwsDetail.AwsAccountId.ValueString())
				awsDetail.SetAwsBucketName(state.FileMetadata.AccessDetails.AwsDetail.AwsBucketName.ValueString())
				awsDetail.SetFilePath(state.FileMetadata.AccessDetails.AwsDetail.FilePath.ValueString())
				accessDetails.AwsDetail = &awsDetail
			}

			// GCP details
			if state.FileMetadata.AccessDetails.GcpDetail != nil {
				gcpDetail := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetailsGcpDetail{}
				gcpDetail.SetGcpProjectId(state.FileMetadata.AccessDetails.GcpDetail.GcpProjectId.ValueString())
				gcpDetail.SetGcpBucketName(state.FileMetadata.AccessDetails.GcpDetail.GcpBucketName.ValueString())
				gcpDetail.SetFilePath(state.FileMetadata.AccessDetails.GcpDetail.FilePath.ValueString())
				gcpDetail.SetGcpServiceAccountEmail(state.FileMetadata.AccessDetails.GcpDetail.GcpServiceAccountEmail.ValueString())
				accessDetails.GcpDetail = &gcpDetail
			}

			// Azure details
			if state.FileMetadata.AccessDetails.AzureDetail != nil {
				azureDetail := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetailsAzureDetail{}
				azureDetail.SetAzureTenantId(state.FileMetadata.AccessDetails.AzureDetail.AzureTenantId.ValueString())
				azureDetail.SetAzureClientId(state.FileMetadata.AccessDetails.AzureDetail.AzureClientId.ValueString())
				azureDetail.SetAzureStorageAccountName(state.FileMetadata.AccessDetails.AzureDetail.AzureStorageAccountName.ValueString())
				azureDetail.SetAzureContainerName(state.FileMetadata.AccessDetails.AzureDetail.AzureContainerName.ValueString())
				azureDetail.SetFilePath(state.FileMetadata.AccessDetails.AzureDetail.FilePath.ValueString())
				accessDetails.AzureDetail = &azureDetail
			}

			cloudStorageMetadata.SetAccessDetails(accessDetails)
		}

		// Set the file_metadata as a oneOf union type
		fileMetadata := datadogV2.CreateTableRequestDataAttributesFileMetadataCloudStorageAsCreateTableRequestDataAttributesFileMetadata(&cloudStorageMetadata)
		attributes.SetFileMetadata(fileMetadata)
	}

	// Build schema
	if state.Schema != nil {
		schema := datadogV2.CreateTableRequestDataAttributesSchema{}

		var primaryKeys []string
		diags.Append(state.Schema.PrimaryKeys.ElementsAs(ctx, &primaryKeys, false)...)
		schema.SetPrimaryKeys(primaryKeys)

		if state.Schema.Fields != nil {
			var fields []datadogV2.CreateTableRequestDataAttributesSchemaFieldsItems
			for _, fieldsTFItem := range state.Schema.Fields {
				if !fieldsTFItem.Name.IsNull() && !fieldsTFItem.Type.IsNull() {
					fieldsDDItem := datadogV2.NewCreateTableRequestDataAttributesSchemaFieldsItems(
						fieldsTFItem.Name.ValueString(),
						datadogV2.ReferenceTableSchemaFieldType(fieldsTFItem.Type.ValueString()),
					)
					fields = append(fields, *fieldsDDItem)
				}
			}
			schema.SetFields(fields)
		}
		attributes.Schema = schema
	}

	req := datadogV2.NewCreateTableRequestWithDefaults()
	req.Data = datadogV2.NewCreateTableRequestData(datadogV2.CREATETABLEREQUESTDATATYPE_REFERENCE_TABLE)
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *referenceTableResource) buildReferenceTableUpdateRequestBody(ctx context.Context, state *referenceTableModel) (*datadogV2.PatchTableRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewPatchTableRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	// Build file_metadata for cloud storage updates
	if state.FileMetadata != nil {
		cloudStorageMetadata := datadogV2.PatchTableRequestDataAttributesFileMetadataCloudStorage{}

		if !state.FileMetadata.SyncEnabled.IsNull() {
			cloudStorageMetadata.SetSyncEnabled(state.FileMetadata.SyncEnabled.ValueBool())
		}

		if state.FileMetadata.AccessDetails != nil {
			accessDetails := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetails{}

			// AWS details
			if state.FileMetadata.AccessDetails.AwsDetail != nil {
				awsDetail := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetailsAwsDetail{}
				awsDetail.SetAwsAccountId(state.FileMetadata.AccessDetails.AwsDetail.AwsAccountId.ValueString())
				awsDetail.SetAwsBucketName(state.FileMetadata.AccessDetails.AwsDetail.AwsBucketName.ValueString())
				awsDetail.SetFilePath(state.FileMetadata.AccessDetails.AwsDetail.FilePath.ValueString())
				accessDetails.AwsDetail = &awsDetail
			}

			// GCP details
			if state.FileMetadata.AccessDetails.GcpDetail != nil {
				gcpDetail := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetailsGcpDetail{}
				gcpDetail.SetGcpProjectId(state.FileMetadata.AccessDetails.GcpDetail.GcpProjectId.ValueString())
				gcpDetail.SetGcpBucketName(state.FileMetadata.AccessDetails.GcpDetail.GcpBucketName.ValueString())
				gcpDetail.SetFilePath(state.FileMetadata.AccessDetails.GcpDetail.FilePath.ValueString())
				gcpDetail.SetGcpServiceAccountEmail(state.FileMetadata.AccessDetails.GcpDetail.GcpServiceAccountEmail.ValueString())
				accessDetails.GcpDetail = &gcpDetail
			}

			// Azure details
			if state.FileMetadata.AccessDetails.AzureDetail != nil {
				azureDetail := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetailsAzureDetail{}
				azureDetail.SetAzureTenantId(state.FileMetadata.AccessDetails.AzureDetail.AzureTenantId.ValueString())
				azureDetail.SetAzureClientId(state.FileMetadata.AccessDetails.AzureDetail.AzureClientId.ValueString())
				azureDetail.SetAzureStorageAccountName(state.FileMetadata.AccessDetails.AzureDetail.AzureStorageAccountName.ValueString())
				azureDetail.SetAzureContainerName(state.FileMetadata.AccessDetails.AzureDetail.AzureContainerName.ValueString())
				azureDetail.SetFilePath(state.FileMetadata.AccessDetails.AzureDetail.FilePath.ValueString())
				accessDetails.AzureDetail = &azureDetail
			}

			cloudStorageMetadata.SetAccessDetails(accessDetails)
		}

		// Set the file_metadata as a oneOf union type
		fileMetadata := datadogV2.PatchTableRequestDataAttributesFileMetadataCloudStorageAsPatchTableRequestDataAttributesFileMetadata(&cloudStorageMetadata)
		attributes.SetFileMetadata(fileMetadata)
	}

	// Build schema for updates
	if state.Schema != nil {
		schema := datadogV2.PatchTableRequestDataAttributesSchema{}

		if !state.Schema.PrimaryKeys.IsNull() {
			var primaryKeys []string
			diags.Append(state.Schema.PrimaryKeys.ElementsAs(ctx, &primaryKeys, false)...)
			schema.SetPrimaryKeys(primaryKeys)
		}

		if state.Schema.Fields != nil {
			var fields []datadogV2.PatchTableRequestDataAttributesSchemaFieldsItems
			for _, fieldsTFItem := range state.Schema.Fields {
				if !fieldsTFItem.Name.IsNull() && !fieldsTFItem.Type.IsNull() {
					fieldsDDItem := datadogV2.NewPatchTableRequestDataAttributesSchemaFieldsItems(
						fieldsTFItem.Name.ValueString(),
						datadogV2.ReferenceTableSchemaFieldType(fieldsTFItem.Type.ValueString()),
					)
					fields = append(fields, *fieldsDDItem)
				}
			}
			schema.SetFields(fields)
		}
		attributes.Schema = &schema
	}

	req := datadogV2.NewPatchTableRequestWithDefaults()
	req.Data = datadogV2.NewPatchTableRequestData(datadogV2.PATCHTABLEREQUESTDATATYPE_REFERENCE_TABLE)
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *referenceTableResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the plan is null (resource is being destroyed) or no state exists yet, return early
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, state referenceTableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate schema changes
	if plan.Schema != nil && state.Schema != nil {
		planSchema := plan.Schema
		stateSchema := state.Schema

		// Check primary keys changes (destructive)
		var planPrimaryKeys, statePrimaryKeys []string
		if !planSchema.PrimaryKeys.IsNull() && !stateSchema.PrimaryKeys.IsNull() {
			planSchema.PrimaryKeys.ElementsAs(ctx, &planPrimaryKeys, false)
			stateSchema.PrimaryKeys.ElementsAs(ctx, &statePrimaryKeys, false)

			// Check if primary keys have changed
			if len(planPrimaryKeys) != len(statePrimaryKeys) {
				resp.Diagnostics.AddError(
					"Destructive schema changes are not supported",
					fmt.Sprintf("Cannot change primary keys from %v to %v.\n\n"+
						"The planned schema change would modify primary keys, which requires table recreation and causes downtime.\n\n"+
						"To proceed:\n"+
						"1. Remove the resource from Terraform state: terraform state rm datadog_reference_table.%s\n"+
						"2. Update your configuration with the new schema\n"+
						"3. Run terraform apply to recreate the table\n\n"+
						"Note: The table will be unavailable during recreation, causing enrichment processors to fail.",
						statePrimaryKeys, planPrimaryKeys, state.TableName.ValueString()),
				)
				return
			}

			for i, planKey := range planPrimaryKeys {
				if i >= len(statePrimaryKeys) || planKey != statePrimaryKeys[i] {
					resp.Diagnostics.AddError(
						"Destructive schema changes are not supported",
						fmt.Sprintf("Cannot change primary keys from %v to %v.\n\n"+
							"The planned schema change would modify primary keys, which requires table recreation and causes downtime.\n\n"+
							"To proceed:\n"+
							"1. Remove the resource from Terraform state: terraform state rm datadog_reference_table.%s\n"+
							"2. Update your configuration with the new schema\n"+
							"3. Run terraform apply to recreate the table\n\n"+
							"Note: The table will be unavailable during recreation, causing enrichment processors to fail.",
							statePrimaryKeys, planPrimaryKeys, state.TableName.ValueString()),
					)
					return
				}
			}
		}

		// Build field maps for comparison
		stateFieldMap := make(map[string]string) // field name -> type
		if stateSchema.Fields != nil {
			for _, field := range stateSchema.Fields {
				if !field.Name.IsNull() && !field.Type.IsNull() {
					stateFieldMap[field.Name.ValueString()] = field.Type.ValueString()
				}
			}
		}

		planFieldMap := make(map[string]string)
		if planSchema.Fields != nil {
			for _, field := range planSchema.Fields {
				if !field.Name.IsNull() && !field.Type.IsNull() {
					planFieldMap[field.Name.ValueString()] = field.Type.ValueString()
				}
			}
		}

		// Check for removed fields (destructive)
		for fieldName := range stateFieldMap {
			if _, exists := planFieldMap[fieldName]; !exists {
				resp.Diagnostics.AddError(
					"Destructive schema changes are not supported",
					fmt.Sprintf("Cannot remove field '%s' from the schema.\n\n"+
						"The planned schema change would remove fields, which requires table recreation and causes downtime.\n\n"+
						"To proceed:\n"+
						"1. Remove the resource from Terraform state: terraform state rm datadog_reference_table.%s\n"+
						"2. Update your configuration with the new schema\n"+
						"3. Run terraform apply to recreate the table\n\n"+
						"Note: The table will be unavailable during recreation, causing enrichment processors to fail.",
						fieldName, state.TableName.ValueString()),
				)
				return
			}
		}

		// Check for field type changes (destructive)
		for fieldName, planType := range planFieldMap {
			if stateType, exists := stateFieldMap[fieldName]; exists {
				if stateType != planType {
					resp.Diagnostics.AddError(
						"Destructive schema changes are not supported",
						fmt.Sprintf("Cannot change type of field '%s' from '%s' to '%s'.\n\n"+
							"The planned schema change would modify field types, which requires table recreation and causes downtime.\n\n"+
							"To proceed:\n"+
							"1. Remove the resource from Terraform state: terraform state rm datadog_reference_table.%s\n"+
							"2. Update your configuration with the new schema\n"+
							"3. Run terraform apply to recreate the table\n\n"+
							"Note: The table will be unavailable during recreation, causing enrichment processors to fail.",
							fieldName, stateType, planType, state.TableName.ValueString()),
					)
					return
				}
			}
			// New fields (additive) are allowed - no error
		}
	}
}
