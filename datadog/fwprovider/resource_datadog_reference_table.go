package fwprovider

import (
	"context"
	"fmt"
	"slices"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &referenceTableResource{}
	_ resource.ResourceWithImportState    = &referenceTableResource{}
	_ resource.ResourceWithValidateConfig = &referenceTableResource{}
	_ resource.ResourceWithModifyPlan     = &referenceTableResource{}
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
	ErrorMessage  types.String        `tfsdk:"error_message"`
	ErrorRowCount types.Int64         `tfsdk:"error_row_count"`
	ErrorType     types.String        `tfsdk:"error_type"`
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
		Description: "Provides a Datadog Reference Table resource for cloud storage sources (S3, GCS, Azure). This can be used to create and manage Datadog reference tables that sync data from cloud storage. For setup instructions including granting Datadog read access to your cloud storage bucket, see the [Reference Tables documentation](https://docs.datadoghq.com/reference_tables/?tab=cloudstorage#create-a-reference-table).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"table_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the reference table. This must be unique within your organization.",
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "The source type for the reference table.",
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
			},
			"last_updated_by": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the user who last updated the reference table.",
			},
			"row_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of successfully processed rows in the reference table.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the reference table (e.g., DONE, PROCESSING, ERROR).",
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
						Description: "The type of error that occurred during file processing.",
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
										Description: "The relative file path from the AWS S3 bucket root to the CSV file.",
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
				Description: "The schema definition for the reference table, including field definitions and primary keys. Schema is only set on create; updates are derived from the file asynchronously.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"primary_keys": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						Description: "List of field names that serve as primary keys for the table. Currently only one primary key is supported.",
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"fields": schema.ListNestedBlock{
						Description: "List of fields in the table schema. Must include at least one field. Schema is only set on create.",
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "The name of the field.",
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"type": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "The data type of the field. Must be one of: STRING, INT32.",
									Validators: []validator.String{
										stringvalidator.OneOf("STRING", "INT32"),
									},
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
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

func (r *referenceTableResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config referenceTableModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate that access_details matches source type
	if config.FileMetadata != nil && config.FileMetadata.AccessDetails != nil {
		source := config.Source.ValueString()
		ad := config.FileMetadata.AccessDetails

		hasAws := ad.AwsDetail != nil
		hasGcp := ad.GcpDetail != nil
		hasAzure := ad.AzureDetail != nil

		// Count how many detail types are specified
		detailCount := 0
		if hasAws {
			detailCount++
		}
		if hasGcp {
			detailCount++
		}
		if hasAzure {
			detailCount++
		}

		// Exactly one detail type must be specified
		if detailCount == 0 {
			response.Diagnostics.AddError(
				"Missing access_details configuration",
				"Exactly one of aws_detail, gcp_detail, or azure_detail must be specified in access_details.",
			)
			return
		}
		if detailCount > 1 {
			response.Diagnostics.AddError(
				"Multiple access_details configurations",
				"Only one of aws_detail, gcp_detail, or azure_detail can be specified in access_details.",
			)
			return
		}

		// Validate that the detail type matches the source
		switch source {
		case "S3":
			if !hasAws {
				response.Diagnostics.AddError(
					"Invalid access_details for source",
					"Source 'S3' requires aws_detail in access_details.",
				)
			}
		case "GCS":
			if !hasGcp {
				response.Diagnostics.AddError(
					"Invalid access_details for source",
					"Source 'GCS' requires gcp_detail in access_details.",
				)
			}
		case "AZURE":
			if !hasAzure {
				response.Diagnostics.AddError(
					"Invalid access_details for source",
					"Source 'AZURE' requires azure_detail in access_details.",
				)
			}
		}
	}
	// Note: schema.fields and schema.primary_keys validation is handled by listvalidator.SizeAtLeast(1) in schema definition
}

func (r *referenceTableResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// Skip if this is a create or destroy operation
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var state referenceTableModel
	var plan referenceTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if schema is being modified on an existing resource
	if state.Schema != nil && plan.Schema != nil {
		// Compare primary keys using slices.Equal
		var statePKs, planPKs []string
		state.Schema.PrimaryKeys.ElementsAs(ctx, &statePKs, false)
		plan.Schema.PrimaryKeys.ElementsAs(ctx, &planPKs, false)
		if !slices.Equal(statePKs, planPKs) {
			response.Diagnostics.AddError(
				"Primary key modification not supported",
				"Reference table primary keys cannot be modified after creation. "+
					"To change the primary key, you must delete and recreate the table.",
			)
			return
		}

		// Compare fields
		fieldsChanged := false
		if len(state.Schema.Fields) != len(plan.Schema.Fields) {
			fieldsChanged = true
		} else {
			for i := range state.Schema.Fields {
				if state.Schema.Fields[i].Name.ValueString() != plan.Schema.Fields[i].Name.ValueString() ||
					state.Schema.Fields[i].Type.ValueString() != plan.Schema.Fields[i].Type.ValueString() {
					fieldsChanged = true
					break
				}
			}
		}

		if fieldsChanged {
			response.Diagnostics.AddError(
				"Schema field modification not supported",
				"Reference table schema fields cannot be modified through Terraform after creation. The schema is derived from the CSV file in cloud storage. "+
					"To change the schema, update the CSV file and the table will sync automatically if sync_enabled is true.",
			)
		}
	}
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
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

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

	_, httpResp, err := r.Api.CreateReferenceTable(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating ReferenceTable"))
		return
	}

	// The create API returns an empty body with 201 status, so we need to fetch the table by name
	if httpResp == nil || httpResp.StatusCode != 201 {
		statusCode := 0
		if httpResp != nil {
			statusCode = httpResp.StatusCode
		}
		response.Diagnostics.AddError("API Error", fmt.Sprintf("CreateReferenceTable returned unexpected status (HTTP %d).", statusCode))
		return
	}

	// List tables with exact name filter to find the created one
	tableName := state.TableName.ValueString()
	optionalParams := datadogV2.ListTablesOptionalParameters{
		FilterTableNameExact: &tableName,
	}
	listResp, _, listErr := r.Api.ListTables(r.Auth, optionalParams)
	if listErr != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(listErr, "table created but error listing tables"))
		return
	}

	if len(listResp.Data) == 0 {
		response.Diagnostics.AddError("API Error", fmt.Sprintf("Table %s was created but not found in list", tableName))
		return
	}

	// Use the table data from the list response
	tableData := listResp.Data[0]
	resp := datadogV2.TableResultV2{Data: &tableData}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *referenceTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var planState referenceTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &planState)...)
	if response.Diagnostics.HasError() {
		return
	}

	var currentState referenceTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &currentState)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := planState.ID.ValueString()

	body, diags := r.buildReferenceTableUpdateRequestBody(ctx, &planState, &currentState)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := r.Api.UpdateReferenceTable(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ReferenceTable"))
		return
	}

	// Read back the updated resource to populate state
	// Note: Schema updates happen asynchronously, so the schema may not be updated immediately.
	// Terraform will refresh state on the next plan/apply cycle to pick up async changes.
	resp, _, err := r.Api.GetTable(r.Auth, id)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading ReferenceTable after update"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &planState, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &planState)...)
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

	if fileMetadata, ok := attributes.GetFileMetadataOk(); ok {
		fileMetadataTf := &fileMetadataModel{}

		if syncEnabled, ok := fileMetadata.GetSyncEnabledOk(); ok {
			fileMetadataTf.SyncEnabled = types.BoolValue(*syncEnabled)
		} else {
			// If sync_enabled is not in API response, preserve existing value from state
			// This handles cases where the API doesn't return sync_enabled in the response
			if state.FileMetadata != nil && !state.FileMetadata.SyncEnabled.IsNull() {
				fileMetadataTf.SyncEnabled = state.FileMetadata.SyncEnabled
			}
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

			if aws := state.FileMetadata.AccessDetails.AwsDetail; aws != nil {
				awsDetail := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetailsAwsDetail{}
				awsDetail.SetAwsAccountId(aws.AwsAccountId.ValueString())
				awsDetail.SetAwsBucketName(aws.AwsBucketName.ValueString())
				awsDetail.SetFilePath(aws.FilePath.ValueString())
				accessDetails.AwsDetail = &awsDetail
			}

			if gcp := state.FileMetadata.AccessDetails.GcpDetail; gcp != nil {
				gcpDetail := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetailsGcpDetail{}
				gcpDetail.SetGcpProjectId(gcp.GcpProjectId.ValueString())
				gcpDetail.SetGcpBucketName(gcp.GcpBucketName.ValueString())
				gcpDetail.SetFilePath(gcp.FilePath.ValueString())
				gcpDetail.SetGcpServiceAccountEmail(gcp.GcpServiceAccountEmail.ValueString())
				accessDetails.GcpDetail = &gcpDetail
			}

			if azure := state.FileMetadata.AccessDetails.AzureDetail; azure != nil {
				azureDetail := datadogV2.CreateTableRequestDataAttributesFileMetadataOneOfAccessDetailsAzureDetail{}
				azureDetail.SetAzureTenantId(azure.AzureTenantId.ValueString())
				azureDetail.SetAzureClientId(azure.AzureClientId.ValueString())
				azureDetail.SetAzureStorageAccountName(azure.AzureStorageAccountName.ValueString())
				azureDetail.SetAzureContainerName(azure.AzureContainerName.ValueString())
				azureDetail.SetFilePath(azure.FilePath.ValueString())
				accessDetails.AzureDetail = &azureDetail
			}

			cloudStorageMetadata.SetAccessDetails(accessDetails)
		}

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

func (r *referenceTableResource) buildReferenceTableUpdateRequestBody(ctx context.Context, planState *referenceTableModel, currentState *referenceTableModel) (*datadogV2.PatchTableRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewPatchTableRequestDataAttributesWithDefaults()

	if !planState.Description.IsNull() {
		attributes.SetDescription(planState.Description.ValueString())
	}

	if !planState.Tags.IsNull() {
		var tags []string
		diags.Append(planState.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	// Note: Schema updates are not supported via PATCH - schema is only set on create.
	// The schema will be derived from the file asynchronously.

	// Build file_metadata for cloud storage updates
	if planState.FileMetadata != nil {
		cloudStorageMetadata := datadogV2.PatchTableRequestDataAttributesFileMetadataCloudStorage{}

		if !planState.FileMetadata.SyncEnabled.IsNull() {
			cloudStorageMetadata.SetSyncEnabled(planState.FileMetadata.SyncEnabled.ValueBool())
		}

		accessDetailsToUse := planState.FileMetadata.AccessDetails

		// Check if we have valid access_details (at least one detail field must be set)
		hasValidAccessDetails := accessDetailsToUse != nil &&
			(accessDetailsToUse.AwsDetail != nil ||
				accessDetailsToUse.GcpDetail != nil ||
				accessDetailsToUse.AzureDetail != nil)

		if hasValidAccessDetails {
			accessDetails := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetails{}

			if aws := accessDetailsToUse.AwsDetail; aws != nil {
				awsDetail := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetailsAwsDetail{}
				awsDetail.SetAwsAccountId(aws.AwsAccountId.ValueString())
				awsDetail.SetAwsBucketName(aws.AwsBucketName.ValueString())
				awsDetail.SetFilePath(aws.FilePath.ValueString())
				accessDetails.AwsDetail = &awsDetail
			}

			if gcp := accessDetailsToUse.GcpDetail; gcp != nil {
				gcpDetail := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetailsGcpDetail{}
				gcpDetail.SetGcpProjectId(gcp.GcpProjectId.ValueString())
				gcpDetail.SetGcpBucketName(gcp.GcpBucketName.ValueString())
				gcpDetail.SetFilePath(gcp.FilePath.ValueString())
				gcpDetail.SetGcpServiceAccountEmail(gcp.GcpServiceAccountEmail.ValueString())
				accessDetails.GcpDetail = &gcpDetail
			}

			if azure := accessDetailsToUse.AzureDetail; azure != nil {
				azureDetail := datadogV2.PatchTableRequestDataAttributesFileMetadataOneOfAccessDetailsAzureDetail{}
				azureDetail.SetAzureTenantId(azure.AzureTenantId.ValueString())
				azureDetail.SetAzureClientId(azure.AzureClientId.ValueString())
				azureDetail.SetAzureStorageAccountName(azure.AzureStorageAccountName.ValueString())
				azureDetail.SetAzureContainerName(azure.AzureContainerName.ValueString())
				azureDetail.SetFilePath(azure.FilePath.ValueString())
				accessDetails.AzureDetail = &azureDetail
			}

			cloudStorageMetadata.SetAccessDetails(accessDetails)
			fileMetadata := datadogV2.PatchTableRequestDataAttributesFileMetadataCloudStorageAsPatchTableRequestDataAttributesFileMetadata(&cloudStorageMetadata)
			attributes.SetFileMetadata(fileMetadata)
		}
	}

	// Note: Schema is not included in PATCH requests. Schema is only set on create.
	// The schema will be derived from the file asynchronously by the backend.

	req := datadogV2.NewPatchTableRequestWithDefaults()
	req.Data = datadogV2.NewPatchTableRequestData(datadogV2.PATCHTABLEREQUESTDATATYPE_REFERENCE_TABLE)
	req.Data.SetAttributes(*attributes)

	return req, diags
}
