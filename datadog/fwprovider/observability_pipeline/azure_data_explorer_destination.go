package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AzureDataExplorerDestinationModel represents the Terraform model for the Azure Data Explorer destination.
type AzureDataExplorerDestinationModel struct {
	IngestionEndpointKey types.String                        `tfsdk:"ingestion_endpoint_key"`
	Database             types.String                        `tfsdk:"database"`
	Table                types.String                        `tfsdk:"table"`
	MappingReference     types.String                        `tfsdk:"mapping_reference"`
	TokenScope           types.String                        `tfsdk:"token_scope"`
	Auth                 []AzureDataExplorerAuthModel        `tfsdk:"auth"`
	Compression          []AzureDataExplorerCompressionModel `tfsdk:"compression"`
	Batch                []AzureDataExplorerBatchModel       `tfsdk:"batch"`
	Buffer               []BufferOptionsModel                `tfsdk:"buffer"`
}

// AzureDataExplorerAuthModel represents the authentication configuration for the Azure Data Explorer destination.
// Exactly one credential kind block must be set.
type AzureDataExplorerAuthModel struct {
	AzureCli                       []AzureDataExplorerAuthAzureCliModel                       `tfsdk:"azure_cli"`
	ClientSecret                   []AzureDataExplorerAuthClientSecretModel                   `tfsdk:"client_secret"`
	ClientCertificate              []AzureDataExplorerAuthClientCertificateModel              `tfsdk:"client_certificate"`
	ManagedIdentity                []AzureDataExplorerAuthManagedIdentityModel                `tfsdk:"managed_identity"`
	ManagedIdentityClientAssertion []AzureDataExplorerAuthManagedIdentityClientAssertionModel `tfsdk:"managed_identity_client_assertion"`
	WorkloadIdentity               []AzureDataExplorerAuthWorkloadIdentityModel               `tfsdk:"workload_identity"`
}

// AzureDataExplorerAuthAzureCliModel represents Azure CLI authentication. It has no configurable fields.
type AzureDataExplorerAuthAzureCliModel struct{}

// AzureDataExplorerAuthClientSecretModel represents Microsoft Entra client secret authentication.
type AzureDataExplorerAuthClientSecretModel struct {
	AzureTenantId        types.String `tfsdk:"azure_tenant_id"`
	AzureClientId        types.String `tfsdk:"azure_client_id"`
	AzureClientSecretKey types.String `tfsdk:"azure_client_secret_key"`
}

// AzureDataExplorerAuthClientCertificateModel represents Microsoft Entra client certificate authentication.
type AzureDataExplorerAuthClientCertificateModel struct {
	AzureTenantId          types.String `tfsdk:"azure_tenant_id"`
	AzureClientId          types.String `tfsdk:"azure_client_id"`
	CertificatePath        types.String `tfsdk:"certificate_path"`
	CertificatePasswordKey types.String `tfsdk:"certificate_password_key"`
}

// AzureDataExplorerAuthManagedIdentityModel represents Azure managed identity authentication.
type AzureDataExplorerAuthManagedIdentityModel struct {
	UserAssignedManagedIdentityId     types.String `tfsdk:"user_assigned_managed_identity_id"`
	UserAssignedManagedIdentityIdType types.String `tfsdk:"user_assigned_managed_identity_id_type"`
}

// AzureDataExplorerAuthManagedIdentityClientAssertionModel represents managed identity client assertion authentication.
type AzureDataExplorerAuthManagedIdentityClientAssertionModel struct {
	ClientAssertionClientId           types.String `tfsdk:"client_assertion_client_id"`
	ClientAssertionTenantId           types.String `tfsdk:"client_assertion_tenant_id"`
	UserAssignedManagedIdentityId     types.String `tfsdk:"user_assigned_managed_identity_id"`
	UserAssignedManagedIdentityIdType types.String `tfsdk:"user_assigned_managed_identity_id_type"`
}

// AzureDataExplorerAuthWorkloadIdentityModel represents Azure Workload Identity authentication.
type AzureDataExplorerAuthWorkloadIdentityModel struct {
	ClientId      types.String `tfsdk:"client_id"`
	TenantId      types.String `tfsdk:"tenant_id"`
	TokenFilePath types.String `tfsdk:"token_file_path"`
}

// AzureDataExplorerCompressionModel represents the compression configuration for the Azure Data Explorer destination.
type AzureDataExplorerCompressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	Level     types.Int64  `tfsdk:"level"`
}

// AzureDataExplorerBatchModel represents the batch configuration for the Azure Data Explorer destination.
type AzureDataExplorerBatchModel struct {
	MaxBytes    types.Int64 `tfsdk:"max_bytes"`
	MaxEvents   types.Int64 `tfsdk:"max_events"`
	TimeoutSecs types.Int64 `tfsdk:"timeout_secs"`
}

// ExpandAzureDataExplorerDestination converts the Terraform model to the Datadog API model.
func ExpandAzureDataExplorerDestination(ctx context.Context, id string, inputs types.List, src *AzureDataExplorerDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	dest.SetDatabase(src.Database.ValueString())
	dest.SetTable(src.Table.ValueString())

	if !src.IngestionEndpointKey.IsNull() && !src.IngestionEndpointKey.IsUnknown() {
		dest.SetIngestionEndpointKey(src.IngestionEndpointKey.ValueString())
	}
	if !src.MappingReference.IsNull() && !src.MappingReference.IsUnknown() {
		dest.SetMappingReference(src.MappingReference.ValueString())
	}
	if !src.TokenScope.IsNull() && !src.TokenScope.IsUnknown() {
		dest.SetTokenScope(src.TokenScope.ValueString())
	}

	if len(src.Auth) > 0 {
		if auth := expandAzureDataExplorerAuth(src.Auth[0]); auth != nil {
			dest.SetAuth(*auth)
		}
	}

	if len(src.Compression) > 0 {
		c := src.Compression[0]
		compression := datadogV2.NewObservabilityPipelineAzureStorageDestinationCompressionGzipWithDefaults()
		compression.SetLevel(c.Level.ValueInt64())
		dest.SetCompression(*compression)
	}

	if len(src.Batch) > 0 {
		b := src.Batch[0]
		batch := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationBatchWithDefaults()
		if !b.MaxBytes.IsNull() && !b.MaxBytes.IsUnknown() {
			batch.SetMaxBytes(b.MaxBytes.ValueInt64())
		}
		if !b.MaxEvents.IsNull() && !b.MaxEvents.IsUnknown() {
			batch.SetMaxEvents(b.MaxEvents.ValueInt64())
		}
		if !b.TimeoutSecs.IsNull() && !b.TimeoutSecs.IsUnknown() {
			batch.SetTimeoutSecs(b.TimeoutSecs.ValueInt64())
		}
		dest.SetBatch(*batch)
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAzureDataExplorerDestination: dest,
	}
}

// expandAzureDataExplorerAuth converts the Terraform auth model to the API oneOf.
func expandAzureDataExplorerAuth(src AzureDataExplorerAuthModel) *datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuth {
	switch {
	case len(src.AzureCli) > 0:
		a := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationAuthAzureCliWithDefaults()
		auth := datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuthAzureCliAsObservabilityPipelineAzureDataExplorerDestinationAuth(a)
		return &auth
	case len(src.ClientSecret) > 0:
		m := src.ClientSecret[0]
		a := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationAuthClientSecretWithDefaults()
		a.SetAzureTenantId(m.AzureTenantId.ValueString())
		a.SetAzureClientId(m.AzureClientId.ValueString())
		a.SetAzureClientSecretKey(m.AzureClientSecretKey.ValueString())
		auth := datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuthClientSecretAsObservabilityPipelineAzureDataExplorerDestinationAuth(a)
		return &auth
	case len(src.ClientCertificate) > 0:
		m := src.ClientCertificate[0]
		a := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationAuthClientCertificateWithDefaults()
		a.SetAzureTenantId(m.AzureTenantId.ValueString())
		a.SetAzureClientId(m.AzureClientId.ValueString())
		a.SetCertificatePath(m.CertificatePath.ValueString())
		if !m.CertificatePasswordKey.IsNull() && !m.CertificatePasswordKey.IsUnknown() {
			a.SetCertificatePasswordKey(m.CertificatePasswordKey.ValueString())
		}
		auth := datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuthClientCertificateAsObservabilityPipelineAzureDataExplorerDestinationAuth(a)
		return &auth
	case len(src.ManagedIdentity) > 0:
		m := src.ManagedIdentity[0]
		a := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentityWithDefaults()
		if !m.UserAssignedManagedIdentityId.IsNull() && !m.UserAssignedManagedIdentityId.IsUnknown() {
			a.SetUserAssignedManagedIdentityId(m.UserAssignedManagedIdentityId.ValueString())
		}
		if !m.UserAssignedManagedIdentityIdType.IsNull() && !m.UserAssignedManagedIdentityIdType.IsUnknown() {
			a.SetUserAssignedManagedIdentityIdType(datadogV2.ObservabilityPipelineAzureDataExplorerDestinationManagedIdentityIdType(m.UserAssignedManagedIdentityIdType.ValueString()))
		}
		auth := datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentityAsObservabilityPipelineAzureDataExplorerDestinationAuth(a)
		return &auth
	case len(src.ManagedIdentityClientAssertion) > 0:
		m := src.ManagedIdentityClientAssertion[0]
		a := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentityClientAssertionWithDefaults()
		a.SetClientAssertionClientId(m.ClientAssertionClientId.ValueString())
		a.SetClientAssertionTenantId(m.ClientAssertionTenantId.ValueString())
		if !m.UserAssignedManagedIdentityId.IsNull() && !m.UserAssignedManagedIdentityId.IsUnknown() {
			a.SetUserAssignedManagedIdentityId(m.UserAssignedManagedIdentityId.ValueString())
		}
		if !m.UserAssignedManagedIdentityIdType.IsNull() && !m.UserAssignedManagedIdentityIdType.IsUnknown() {
			a.SetUserAssignedManagedIdentityIdType(datadogV2.ObservabilityPipelineAzureDataExplorerDestinationManagedIdentityIdType(m.UserAssignedManagedIdentityIdType.ValueString()))
		}
		auth := datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentityClientAssertionAsObservabilityPipelineAzureDataExplorerDestinationAuth(a)
		return &auth
	case len(src.WorkloadIdentity) > 0:
		m := src.WorkloadIdentity[0]
		a := datadogV2.NewObservabilityPipelineAzureDataExplorerDestinationAuthWorkloadIdentityWithDefaults()
		if !m.ClientId.IsNull() && !m.ClientId.IsUnknown() {
			a.SetClientId(m.ClientId.ValueString())
		}
		if !m.TenantId.IsNull() && !m.TenantId.IsUnknown() {
			a.SetTenantId(m.TenantId.ValueString())
		}
		if !m.TokenFilePath.IsNull() && !m.TokenFilePath.IsUnknown() {
			a.SetTokenFilePath(m.TokenFilePath.ValueString())
		}
		auth := datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuthWorkloadIdentityAsObservabilityPipelineAzureDataExplorerDestinationAuth(a)
		return &auth
	default:
		return nil
	}
}

// FlattenAzureDataExplorerDestination converts the Datadog API model to the Terraform model.
func FlattenAzureDataExplorerDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineAzureDataExplorerDestination) *AzureDataExplorerDestinationModel {
	if src == nil {
		return nil
	}

	model := &AzureDataExplorerDestinationModel{
		Database: types.StringValue(src.GetDatabase()),
		Table:    types.StringValue(src.GetTable()),
	}

	if v, ok := src.GetIngestionEndpointKeyOk(); ok {
		model.IngestionEndpointKey = types.StringValue(*v)
	} else {
		model.IngestionEndpointKey = types.StringNull()
	}

	if v, ok := src.GetMappingReferenceOk(); ok && v != nil {
		model.MappingReference = types.StringValue(*v)
	} else {
		model.MappingReference = types.StringNull()
	}

	if v, ok := src.GetTokenScopeOk(); ok {
		model.TokenScope = types.StringValue(*v)
	} else {
		model.TokenScope = types.StringNull()
	}

	if auth, ok := src.GetAuthOk(); ok {
		if authModel := flattenAzureDataExplorerAuth(auth); authModel != nil {
			model.Auth = []AzureDataExplorerAuthModel{*authModel}
		}
	}

	if compression, ok := src.GetCompressionOk(); ok {
		model.Compression = []AzureDataExplorerCompressionModel{{
			Algorithm: types.StringValue(string(compression.GetAlgorithm())),
			Level:     types.Int64Value(compression.GetLevel()),
		}}
	}

	if batch, ok := src.GetBatchOk(); ok {
		bm := AzureDataExplorerBatchModel{}
		if v, ok2 := batch.GetMaxBytesOk(); ok2 {
			bm.MaxBytes = types.Int64Value(*v)
		} else {
			bm.MaxBytes = types.Int64Null()
		}
		if v, ok2 := batch.GetMaxEventsOk(); ok2 {
			bm.MaxEvents = types.Int64Value(*v)
		} else {
			bm.MaxEvents = types.Int64Null()
		}
		if v, ok2 := batch.GetTimeoutSecsOk(); ok2 {
			bm.TimeoutSecs = types.Int64Value(*v)
		} else {
			bm.TimeoutSecs = types.Int64Null()
		}
		model.Batch = []AzureDataExplorerBatchModel{bm}
	}

	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			model.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return model
}

// flattenAzureDataExplorerAuth converts the API auth oneOf to the Terraform model.
func flattenAzureDataExplorerAuth(src *datadogV2.ObservabilityPipelineAzureDataExplorerDestinationAuth) *AzureDataExplorerAuthModel {
	if src == nil {
		return nil
	}

	out := &AzureDataExplorerAuthModel{}
	switch {
	case src.ObservabilityPipelineAzureDataExplorerDestinationAuthAzureCli != nil:
		out.AzureCli = []AzureDataExplorerAuthAzureCliModel{{}}
	case src.ObservabilityPipelineAzureDataExplorerDestinationAuthClientSecret != nil:
		a := src.ObservabilityPipelineAzureDataExplorerDestinationAuthClientSecret
		out.ClientSecret = []AzureDataExplorerAuthClientSecretModel{{
			AzureTenantId:        types.StringValue(a.GetAzureTenantId()),
			AzureClientId:        types.StringValue(a.GetAzureClientId()),
			AzureClientSecretKey: types.StringValue(a.GetAzureClientSecretKey()),
		}}
	case src.ObservabilityPipelineAzureDataExplorerDestinationAuthClientCertificate != nil:
		a := src.ObservabilityPipelineAzureDataExplorerDestinationAuthClientCertificate
		out.ClientCertificate = []AzureDataExplorerAuthClientCertificateModel{{
			AzureTenantId:          types.StringValue(a.GetAzureTenantId()),
			AzureClientId:          types.StringValue(a.GetAzureClientId()),
			CertificatePath:        types.StringValue(a.GetCertificatePath()),
			CertificatePasswordKey: nullableStringValue(a.GetCertificatePasswordKeyOk()),
		}}
	case src.ObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentity != nil:
		a := src.ObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentity
		m := AzureDataExplorerAuthManagedIdentityModel{
			UserAssignedManagedIdentityId:     nullableStringValue(a.GetUserAssignedManagedIdentityIdOk()),
			UserAssignedManagedIdentityIdType: types.StringNull(),
		}
		if v, ok := a.GetUserAssignedManagedIdentityIdTypeOk(); ok {
			m.UserAssignedManagedIdentityIdType = types.StringValue(string(*v))
		}
		out.ManagedIdentity = []AzureDataExplorerAuthManagedIdentityModel{m}
	case src.ObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentityClientAssertion != nil:
		a := src.ObservabilityPipelineAzureDataExplorerDestinationAuthManagedIdentityClientAssertion
		m := AzureDataExplorerAuthManagedIdentityClientAssertionModel{
			ClientAssertionClientId:           types.StringValue(a.GetClientAssertionClientId()),
			ClientAssertionTenantId:           types.StringValue(a.GetClientAssertionTenantId()),
			UserAssignedManagedIdentityId:     nullableStringValue(a.GetUserAssignedManagedIdentityIdOk()),
			UserAssignedManagedIdentityIdType: types.StringNull(),
		}
		if v, ok := a.GetUserAssignedManagedIdentityIdTypeOk(); ok {
			m.UserAssignedManagedIdentityIdType = types.StringValue(string(*v))
		}
		out.ManagedIdentityClientAssertion = []AzureDataExplorerAuthManagedIdentityClientAssertionModel{m}
	case src.ObservabilityPipelineAzureDataExplorerDestinationAuthWorkloadIdentity != nil:
		a := src.ObservabilityPipelineAzureDataExplorerDestinationAuthWorkloadIdentity
		out.WorkloadIdentity = []AzureDataExplorerAuthWorkloadIdentityModel{{
			ClientId:      nullableStringValue(a.GetClientIdOk()),
			TenantId:      nullableStringValue(a.GetTenantIdOk()),
			TokenFilePath: nullableStringValue(a.GetTokenFilePathOk()),
		}}
	default:
		return nil
	}

	return out
}

// nullableStringValue converts a nullable API string getter result to a Terraform string.
func nullableStringValue(v *string, ok bool) types.String {
	if !ok || v == nil {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

// azureDataExplorerManagedIdentityAttributes returns the user-assigned managed identity attributes shared by
// the `managed_identity` and `managed_identity_client_assertion` auth blocks.
func azureDataExplorerManagedIdentityAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"user_assigned_managed_identity_id": schema.StringAttribute{
			Optional:    true,
			Description: "The ID of the user-assigned managed identity. If omitted, the system-assigned managed identity is used.",
		},
		"user_assigned_managed_identity_id_type": schema.StringAttribute{
			Optional:    true,
			Description: "The type of the user-assigned managed identity ID. Valid values are `client_id`, `object_id`, and `resource_id`.",
			Validators: []validator.String{
				stringvalidator.OneOf("client_id", "object_id", "resource_id"),
			},
		},
	}
}

// AzureDataExplorerDestinationSchema returns the schema for the Azure Data Explorer destination.
func AzureDataExplorerDestinationSchema() schema.ListNestedBlock {
	managedIdentityClientAssertionAttributes := azureDataExplorerManagedIdentityAttributes()
	managedIdentityClientAssertionAttributes["client_assertion_client_id"] = schema.StringAttribute{
		Required:    true,
		Description: "The client ID of the Microsoft Entra application that trusts the managed identity.",
	}
	managedIdentityClientAssertionAttributes["client_assertion_tenant_id"] = schema.StringAttribute{
		Required:    true,
		Description: "The tenant ID of the Microsoft Entra application that trusts the managed identity.",
	}

	return schema.ListNestedBlock{
		Description: "The `azure_data_explorer` destination sends log events to an Azure Data Explorer table.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"ingestion_endpoint_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the Azure Data Explorer ingestion endpoint URL. Defaults to `DESTINATION_AZURE_DATA_EXPLORER_INGESTION_ENDPOINT`.",
				},
				"database": schema.StringAttribute{
					Required:    true,
					Description: "The name of the Azure Data Explorer database to ingest into. Supports template syntax.",
				},
				"table": schema.StringAttribute{
					Required:    true,
					Description: "The name of the Azure Data Explorer table to ingest into. Supports template syntax.",
				},
				"mapping_reference": schema.StringAttribute{
					Optional:    true,
					Description: "The name of a pre-created ingestion mapping on the table used to map incoming events to columns. Supports template syntax.",
				},
				"token_scope": schema.StringAttribute{
					Optional:    true,
					Description: "The OAuth scope requested when acquiring an access token for Azure Data Explorer. Defaults to `https://kusto.kusto.windows.net/.default`.",
				},
			},
			Blocks: map[string]schema.Block{
				"auth": schema.ListNestedBlock{
					Description: "Authentication configuration for Azure Data Explorer. Exactly one of `azure_cli`, `client_secret`, `client_certificate`, `managed_identity`, `managed_identity_client_assertion`, or `workload_identity` must be specified.",
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"azure_cli": schema.ListNestedBlock{
								Description: "Authenticate using the Azure CLI credentials available in the environment.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{},
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
									listvalidator.ExactlyOneOf(
										frameworkPath.MatchRelative().AtParent().AtName("client_secret"),
										frameworkPath.MatchRelative().AtParent().AtName("client_certificate"),
										frameworkPath.MatchRelative().AtParent().AtName("managed_identity"),
										frameworkPath.MatchRelative().AtParent().AtName("managed_identity_client_assertion"),
										frameworkPath.MatchRelative().AtParent().AtName("workload_identity"),
									),
								},
							},
							"client_secret": schema.ListNestedBlock{
								Description: "Authenticate using a Microsoft Entra application client secret.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"azure_tenant_id": schema.StringAttribute{
											Required:    true,
											Description: "The Microsoft Entra tenant ID.",
										},
										"azure_client_id": schema.StringAttribute{
											Required:    true,
											Description: "The Microsoft Entra application (client) ID.",
										},
										"azure_client_secret_key": schema.StringAttribute{
											Required:    true,
											Description: "Name of the environment variable or secret that holds the Microsoft Entra application client secret.",
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
							},
							"client_certificate": schema.ListNestedBlock{
								Description: "Authenticate using a Microsoft Entra application client certificate.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"azure_tenant_id": schema.StringAttribute{
											Required:    true,
											Description: "The Microsoft Entra tenant ID.",
										},
										"azure_client_id": schema.StringAttribute{
											Required:    true,
											Description: "The Microsoft Entra application (client) ID.",
										},
										"certificate_path": schema.StringAttribute{
											Required:    true,
											Description: "Path to the `.pfx` client certificate file on the Worker.",
										},
										"certificate_password_key": schema.StringAttribute{
											Optional:    true,
											Description: "Name of the environment variable or secret that holds the password for the client certificate.",
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
							},
							"managed_identity": schema.ListNestedBlock{
								Description: "Authenticate using an Azure managed identity.",
								NestedObject: schema.NestedBlockObject{
									Attributes: azureDataExplorerManagedIdentityAttributes(),
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
							},
							"managed_identity_client_assertion": schema.ListNestedBlock{
								Description: "Authenticate using a managed identity as a client assertion for a Microsoft Entra application.",
								NestedObject: schema.NestedBlockObject{
									Attributes: managedIdentityClientAssertionAttributes,
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
							},
							"workload_identity": schema.ListNestedBlock{
								Description: "Authenticate using Azure Workload Identity (for example, on Kubernetes).",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"client_id": schema.StringAttribute{
											Optional:    true,
											Description: "The client ID of the Microsoft Entra application. If omitted, it is read from the environment.",
										},
										"tenant_id": schema.StringAttribute{
											Optional:    true,
											Description: "The Microsoft Entra tenant ID. If omitted, it is read from the environment.",
										},
										"token_file_path": schema.StringAttribute{
											Optional:    true,
											Description: "Path to the federated token file. If omitted, it is read from the environment.",
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(1),
					},
				},
				"compression": schema.ListNestedBlock{
					Description: "Compression configuration for ingested data. Only `gzip` is supported.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"algorithm": schema.StringAttribute{
								Required:    true,
								Description: "Compression algorithm. Must be `gzip`.",
								Validators: []validator.String{
									stringvalidator.OneOf("gzip"),
								},
							},
							"level": schema.Int64Attribute{
								Required:    true,
								Description: "Gzip compression level (1-9).",
								Validators: []validator.Int64{
									int64validator.Between(1, 9),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"batch": schema.ListNestedBlock{
					Description: "Event batching settings for Azure Data Explorer ingestion.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"max_bytes": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum batch size in bytes.",
								Validators: []validator.Int64{
									int64validator.Between(1, 4000000),
								},
							},
							"max_events": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum number of events per batch before it is flushed.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
							"timeout_secs": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum number of seconds to wait before flushing a partial batch.",
								Validators: []validator.Int64{
									int64validator.AtLeast(1),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
