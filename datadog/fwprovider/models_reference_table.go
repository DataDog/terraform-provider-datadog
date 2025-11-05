package fwprovider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Shared model definitions for reference table resource and data sources

type schemaModel struct {
	PrimaryKeys types.List  `tfsdk:"primary_keys"`
	Fields      []*fieldsModel `tfsdk:"fields"`
}

type fieldsModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
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
	AzureTenantId          types.String `tfsdk:"azure_tenant_id"`
	AzureClientId          types.String `tfsdk:"azure_client_id"`
	AzureStorageAccountName types.String `tfsdk:"azure_storage_account_name"`
	AzureContainerName     types.String `tfsdk:"azure_container_name"`
	FilePath               types.String `tfsdk:"file_path"`
}

type gcpDetailModel struct {
	GcpProjectId            types.String `tfsdk:"gcp_project_id"`
	GcpBucketName           types.String `tfsdk:"gcp_bucket_name"`
	FilePath                types.String `tfsdk:"file_path"`
	GcpServiceAccountEmail  types.String `tfsdk:"gcp_service_account_email"`
}

