package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Shared model definitions for reference table resource and data sources

type schemaModel struct {
	PrimaryKeys types.List `tfsdk:"primary_keys"`
	Fields      types.List `tfsdk:"fields"` // List of fieldsModel
}

type fieldsModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

// fieldsModelAttrTypes returns the attribute types for fieldsModel
func fieldsModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"type": types.StringType,
	}
}

// fieldsModelObjectType returns the object type for fieldsModel
func fieldsModelObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: fieldsModelAttrTypes()}
}

// getFieldsFromList extracts a slice of fieldsModel from a types.List
func getFieldsFromList(ctx context.Context, list types.List) ([]*fieldsModel, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}

	var fields []*fieldsModel
	diags := list.ElementsAs(ctx, &fields, false)
	return fields, diags
}

// fieldsToListValue converts a slice of fieldsModel to a types.List
func fieldsToListValue(ctx context.Context, fields []*fieldsModel) (basetypes.ListValue, diag.Diagnostics) {
	if len(fields) == 0 {
		return types.ListNull(fieldsModelObjectType()), nil
	}

	var elements []attr.Value
	for _, field := range fields {
		objVal, diags := types.ObjectValue(fieldsModelAttrTypes(), map[string]attr.Value{
			"name": field.Name,
			"type": field.Type,
		})
		if diags.HasError() {
			return types.ListNull(fieldsModelObjectType()), diags
		}
		elements = append(elements, objVal)
	}

	return types.ListValue(fieldsModelObjectType(), elements)
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
	AzureTenantId           types.String `tfsdk:"azure_tenant_id"`
	AzureClientId           types.String `tfsdk:"azure_client_id"`
	AzureStorageAccountName types.String `tfsdk:"azure_storage_account_name"`
	AzureContainerName      types.String `tfsdk:"azure_container_name"`
	FilePath                types.String `tfsdk:"file_path"`
}

type gcpDetailModel struct {
	GcpProjectId           types.String `tfsdk:"gcp_project_id"`
	GcpBucketName          types.String `tfsdk:"gcp_bucket_name"`
	FilePath               types.String `tfsdk:"file_path"`
	GcpServiceAccountEmail types.String `tfsdk:"gcp_service_account_email"`
}
