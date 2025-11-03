package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &referenceTableResource{}
	_ resource.ResourceWithImportState = &referenceTableResource{}
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
	createTableRequestDataAttributesFileMetadataCloudStorageModel
	createTableRequestDataAttributesFileMetadataLocalFileModel
}
type createTableRequestDataAttributesFileMetadataCloudStorageModel struct {
	SyncEnabled   types.Bool          `tfsdk:"sync_enabled"`
	AccessDetails *accessDetailsModel `tfsdk:"access_details"`
}

type createTableRequestDataAttributesFileMetadataLocalFileModel struct {
	FilePath types.String `tfsdk:"file_path"` // for local files we accept the file path and will perform the upload as part of the resource creation
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
		Description: "Provides a Datadog ReferenceTable resource. This can be used to create and manage Datadog reference_table.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the reference table.",
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "The source type for creating reference table data. Only these source types can be created through this API.",
			},
			"table_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the reference table.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Description: "The tags of the reference table.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"file_metadata": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"create_table_request_data_attributes_file_metadata_cloud_storage": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"sync_enabled": schema.BoolAttribute{
								Optional:    true,
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
									"azure_detail": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"azure_client_id": schema.StringAttribute{
												Optional:    true,
												Description: "The Azure client ID.",
											},
											"azure_container_name": schema.StringAttribute{
												Optional:    true,
												Description: "The name of the Azure container.",
											},
											"azure_storage_account_name": schema.StringAttribute{
												Optional:    true,
												Description: "The name of the Azure storage account.",
											},
											"azure_tenant_id": schema.StringAttribute{
												Optional:    true,
												Description: "The ID of the Azure tenant.",
											},
											"file_path": schema.StringAttribute{
												Optional:    true,
												Description: "The relative file path from the Azure container root to the CSV file.",
											},
										},
									},
									"gcp_detail": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"file_path": schema.StringAttribute{
												Optional:    true,
												Description: "The relative file path from the GCS bucket root to the CSV file.",
											},
											"gcp_bucket_name": schema.StringAttribute{
												Optional:    true,
												Description: "The name of the GCP bucket.",
											},
											"gcp_project_id": schema.StringAttribute{
												Optional:    true,
												Description: "The ID of the GCP project.",
											},
											"gcp_service_account_email": schema.StringAttribute{
												Optional:    true,
												Description: "The email of the GCP service account.",
											},
										},
									},
								},
							},
						},
					},
					"create_table_request_data_attributes_file_metadata_local_file": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"upload_id": schema.StringAttribute{
								Optional:    true,
								Description: "The upload ID.",
							},
						},
					},
				},
			},
			"schema": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"primary_keys": schema.ListAttribute{
						Optional:    true,
						Description: "List of field names that serve as primary keys for the table. Only one primary key is supported, and it is used as an ID to retrieve rows.",
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					"fields": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Optional:    true,
									Description: "The field name.",
								},
								"type": schema.StringAttribute{
									Optional:    true,
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

	resp, _, err := r.Api.CreateReferenceTable(r.Auth, *body)
	if err != nil {
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

	if schema, ok := attributes.GetSchemaOk(); ok {

		schemaTf := schemaModel{}
		if fields, ok := schema.GetFieldsOk(); ok && len(*fields) > 0 {

			schemaTf.Fields = []*fieldsModel{}
			for _, fieldsDd := range *fields {
				fieldsTfItem := fieldsModel{}

				fieldsTf := fieldsModel{}
				if name, ok := fieldsDd.GetNameOk(); ok {
					fieldsTf.Name = types.StringValue(*name)
				}
				if typeVar, ok := fieldsDd.GetTypeOk(); ok {
					fieldsTf.Type = types.StringValue(string(*typeVar))
				}
				fieldsTfItem = fieldsTf

				schemaTf.Fields = append(schemaTf.Fields, &fieldsTfItem)
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
	req := &datadogV2.CreateTableRequest{}
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

	if state.Schema != nil {
		var schema datadogV2.CreateTableRequestDataAttributesSchema

		var primaryKeys []string
		diags.Append(state.Schema.PrimaryKeys.ElementsAs(ctx, &primaryKeys, false)...)
		schema.SetPrimaryKeys(primaryKeys)

		if state.Schema.Fields != nil {
			var fields []datadogV2.CreateTableRequestDataAttributesSchemaFieldsItems
			for _, fieldsTFItem := range state.Schema.Fields {
				if !fieldsTFItem.Name.IsNull() && !fieldsTFItem.Type.IsNull() {
					fieldsDDItem := datadogV2.NewCreateTableRequestDataAttributesSchemaFieldsItems(fieldsTFItem.Name.ValueString(), datadogV2.ReferenceTableSchemaFieldType(fieldsTFItem.Type.ValueString()))
					fields = append(fields, *fieldsDDItem)
				}
			}
			schema.SetFields(fields)
		}
		attributes.Schema = schema
	}

	req = datadogV2.NewCreateTableRequestWithDefaults()
	req.Data = datadogV2.NewCreateTableRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *referenceTableResource) buildReferenceTableUpdateRequestBody(ctx context.Context, state *referenceTableModel) (*datadogV2.PatchTableRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.PatchTableRequest{}
	attributes := datadogV2.NewPatchTableRequestDataAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	if !state.FileMetadata.SyncEnabled.IsNull() {
		attributes.SetSyncEnabled(state.FileMetadata.SyncEnabled.ValueBool())
	}

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	if state.Schema != nil {
		var schema datadogV2.PatchTableRequestDataAttributesSchema

		if !state.Schema.PrimaryKeys.IsNull() {
			var primaryKeys []string
			diags.Append(state.Schema.PrimaryKeys.ElementsAs(ctx, &primaryKeys, false)...)
			schema.SetPrimaryKeys(primaryKeys)
		}

		if state.Schema.Fields != nil {
			var fields []datadogV2.PatchTableRequestDataAttributesSchemaFieldsItems
			for _, fieldsTFItem := range state.Schema.Fields {
				if !fieldsTFItem.Name.IsNull() && !fieldsTFItem.Type.IsNull() {
					fieldsDDItem := datadogV2.NewPatchTableRequestDataAttributesSchemaFieldsItems(fieldsTFItem.Name.ValueString(), datadogV2.ReferenceTableSchemaFieldType(fieldsTFItem.Type.ValueString()))
					fields = append(fields, *fieldsDDItem)
				}
			}
			schema.SetFields(fields)
		}
		attributes.Schema = &schema
	}

	req = datadogV2.NewPatchTableRequestWithDefaults()
	req.Data = datadogV2.NewPatchTableRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
