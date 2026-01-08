package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AwsAuthModel represents AWS authentication credentials
type AwsAuthModel struct {
	AssumeRole  types.String `tfsdk:"assume_role"`
	ExternalId  types.String `tfsdk:"external_id"`
	SessionName types.String `tfsdk:"session_name"`
}

// ExpandAwsAuth converts the Terraform AWS auth model to the Datadog API model
func ExpandAwsAuth(authTF AwsAuthModel) datadogV2.ObservabilityPipelineAwsAuth {
	auth := datadogV2.ObservabilityPipelineAwsAuth{}
	if !authTF.AssumeRole.IsNull() {
		auth.AssumeRole = authTF.AssumeRole.ValueStringPointer()
	}
	if !authTF.ExternalId.IsNull() {
		auth.ExternalId = authTF.ExternalId.ValueStringPointer()
	}
	if !authTF.SessionName.IsNull() {
		auth.SessionName = authTF.SessionName.ValueStringPointer()
	}
	return auth
}

// FlattenAwsAuth converts the Datadog API AWS auth model to the Terraform model
func FlattenAwsAuth(src *datadogV2.ObservabilityPipelineAwsAuth) []AwsAuthModel {
	if src == nil {
		return nil
	}
	return []AwsAuthModel{{
		AssumeRole:  types.StringPointerValue(src.AssumeRole),
		ExternalId:  types.StringPointerValue(src.ExternalId),
		SessionName: types.StringPointerValue(src.SessionName),
	}}
}

// AwsAuthSchema returns the schema for AWS authentication configuration
func AwsAuthSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "AWS authentication credentials used for accessing AWS services. If omitted, the system's default credentials are used (for example, the IAM role and environment variables).",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"assume_role": schema.StringAttribute{
					Optional:    true,
					Description: "The Amazon Resource Name (ARN) of the role to assume.",
				},
				"external_id": schema.StringAttribute{
					Optional:    true,
					Description: "A unique identifier for cross-account role assumption.",
				},
				"session_name": schema.StringAttribute{
					Optional:    true,
					Description: "A session identifier used for logging and tracing the assumed role session.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
