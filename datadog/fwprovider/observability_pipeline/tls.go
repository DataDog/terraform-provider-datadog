package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TlsModel represents TLS configuration
type TlsModel struct {
	CrtFile    types.String `tfsdk:"crt_file"`
	CaFile     types.String `tfsdk:"ca_file"`
	KeyFile    types.String `tfsdk:"key_file"`
	KeyPassKey types.String `tfsdk:"key_pass_key"`
}

// ExpandTls converts the Terraform TLS model to the Datadog API model
func ExpandTls(tlsTF []TlsModel) *datadogV2.ObservabilityPipelineTls {
	if len(tlsTF) == 0 {
		return nil
	}

	tlsItem := tlsTF[0]
	tls := &datadogV2.ObservabilityPipelineTls{
		CrtFile: tlsItem.CrtFile.ValueString(),
	}
	if !tlsItem.CaFile.IsNull() {
		tls.SetCaFile(tlsItem.CaFile.ValueString())
	}
	if !tlsItem.KeyFile.IsNull() {
		tls.SetKeyFile(tlsItem.KeyFile.ValueString())
	}
	if !tlsItem.KeyPassKey.IsNull() {
		tls.SetKeyPassKey(tlsItem.KeyPassKey.ValueString())
	}
	return tls
}

// FlattenTls converts the Datadog API TLS model to the Terraform model
func FlattenTls(src *datadogV2.ObservabilityPipelineTls) []TlsModel {
	if src == nil {
		return []TlsModel{}
	}
	out := TlsModel{
		CrtFile: types.StringValue(src.CrtFile),
		CaFile:  types.StringPointerValue(src.CaFile),
		KeyFile: types.StringPointerValue(src.KeyFile),
	}
	if v, ok := src.GetKeyPassKeyOk(); ok {
		out.KeyPassKey = types.StringValue(*v)
	}
	return []TlsModel{out}
}

// ClientTlsModel represents TLS configuration for outgoing (client) connections, with SNI override support
type ClientTlsModel struct {
	CrtFile    types.String `tfsdk:"crt_file"`
	CaFile     types.String `tfsdk:"ca_file"`
	KeyFile    types.String `tfsdk:"key_file"`
	KeyPassKey types.String `tfsdk:"key_pass_key"`
	ServerName types.String `tfsdk:"server_name"`
}

// ExpandClientTls converts the Terraform client TLS model to the Datadog API model
func ExpandClientTls(tlsTF []ClientTlsModel) *datadogV2.ObservabilityPipelineClientTls {
	if len(tlsTF) == 0 {
		return nil
	}

	tlsItem := tlsTF[0]
	tls := &datadogV2.ObservabilityPipelineClientTls{
		CrtFile: tlsItem.CrtFile.ValueString(),
	}
	if !tlsItem.CaFile.IsNull() {
		tls.SetCaFile(tlsItem.CaFile.ValueString())
	}
	if !tlsItem.KeyFile.IsNull() {
		tls.SetKeyFile(tlsItem.KeyFile.ValueString())
	}
	if !tlsItem.KeyPassKey.IsNull() {
		tls.SetKeyPassKey(tlsItem.KeyPassKey.ValueString())
	}
	if !tlsItem.ServerName.IsNull() {
		tls.SetServerName(tlsItem.ServerName.ValueString())
	}
	return tls
}

// FlattenClientTls converts the Datadog API client TLS model to the Terraform model
func FlattenClientTls(src *datadogV2.ObservabilityPipelineClientTls) []ClientTlsModel {
	if src == nil {
		return []ClientTlsModel{}
	}
	out := ClientTlsModel{
		CrtFile: types.StringValue(src.CrtFile),
		CaFile:  types.StringPointerValue(src.CaFile),
		KeyFile: types.StringPointerValue(src.KeyFile),
	}
	if v, ok := src.GetKeyPassKeyOk(); ok {
		out.KeyPassKey = types.StringValue(*v)
	}
	if v, ok := src.GetServerNameOk(); ok {
		out.ServerName = types.StringValue(*v)
	}
	return []ClientTlsModel{out}
}

// ClientTlsSchema returns the schema for outgoing (client) TLS configuration, with SNI override support
func ClientTlsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for enabling TLS encryption between the pipeline component and external services, with support for overriding the server name used for the TLS handshake.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"crt_file": schema.StringAttribute{
					Required:    true,
					Description: "Path to the TLS client certificate file used to authenticate the pipeline component with upstream or downstream services.",
				},
				"ca_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to the Certificate Authority (CA) file used to validate the server's TLS certificate.",
				},
				"key_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to the private key file associated with the TLS client certificate. Used for mutual TLS authentication.",
				},
				"key_pass_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the passphrase for the private key file.",
				},
				"server_name": schema.StringAttribute{
					Optional:    true,
					Description: "Server name to use for Server Name Indication (SNI) and to verify against the certificate presented by the remote host. Use this when the address you connect to doesn't match the certificate's Common Name or Subject Alternative Name.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// MtlsServerTlsModel represents mTLS server TLS configuration for sources
type MtlsServerTlsModel struct {
	CrtFile           types.String `tfsdk:"crt_file"`
	CaFile            types.String `tfsdk:"ca_file"`
	KeyFile           types.String `tfsdk:"key_file"`
	KeyPassKey        types.String `tfsdk:"key_pass_key"`
	VerifyCertificate types.Bool   `tfsdk:"verify_certificate"`
}

// ExpandMtlsServerTls converts the Terraform mTLS server TLS model to the Datadog API model
func ExpandMtlsServerTls(tlsTF []MtlsServerTlsModel) *datadogV2.ObservabilityPipelineMtlsServerTls {
	if len(tlsTF) == 0 {
		return nil
	}
	item := tlsTF[0]
	tls := datadogV2.NewObservabilityPipelineMtlsServerTls(item.CrtFile.ValueString())
	if !item.CaFile.IsNull() {
		tls.SetCaFile(item.CaFile.ValueString())
	}
	if !item.KeyFile.IsNull() {
		tls.SetKeyFile(item.KeyFile.ValueString())
	}
	if !item.KeyPassKey.IsNull() {
		tls.SetKeyPassKey(item.KeyPassKey.ValueString())
	}
	if !item.VerifyCertificate.IsNull() {
		tls.SetVerifyCertificate(item.VerifyCertificate.ValueBool())
	}
	return tls
}

// FlattenMtlsServerTls converts the Datadog API mTLS server TLS model to the Terraform model
func FlattenMtlsServerTls(src *datadogV2.ObservabilityPipelineMtlsServerTls) []MtlsServerTlsModel {
	if src == nil {
		return []MtlsServerTlsModel{}
	}
	out := MtlsServerTlsModel{
		CrtFile: types.StringValue(src.CrtFile),
		CaFile:  types.StringPointerValue(src.CaFile),
		KeyFile: types.StringPointerValue(src.KeyFile),
	}
	if v, ok := src.GetKeyPassKeyOk(); ok {
		out.KeyPassKey = types.StringValue(*v)
	}
	if v, ok := src.GetVerifyCertificateOk(); ok {
		out.VerifyCertificate = types.BoolValue(*v)
	}
	return []MtlsServerTlsModel{out}
}

// MtlsServerTlsSchema returns the schema for mTLS server TLS configuration
func MtlsServerTlsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for enabling TLS encryption between the pipeline component and external connecting clients.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"crt_file": schema.StringAttribute{
					Required:    true,
					Description: "Path to the TLS server certificate file used to identify the pipeline component to connecting clients.",
				},
				"ca_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to the Certificate Authority (CA) file used to validate connecting clients' TLS certificates.",
				},
				"key_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to the private key file associated with the TLS server certificate.",
				},
				"key_pass_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the passphrase for the private key file.",
				},
				"verify_certificate": schema.BoolAttribute{
					Optional:    true,
					Description: "When `true`, requires client connections to present a valid certificate, enabling mutual TLS authentication.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// TlsSchema returns the schema for TLS configuration
func TlsSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Configuration for enabling TLS encryption between the pipeline component and external services.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"crt_file": schema.StringAttribute{
					Required:    true,
					Description: "Path to the TLS client certificate file used to authenticate the pipeline component with upstream or downstream services.",
				},
				"ca_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to the Certificate Authority (CA) file used to validate the server's TLS certificate.",
				},
				"key_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to the private key file associated with the TLS client certificate. Used for mutual TLS authentication.",
				},
				"key_pass_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the passphrase for the private key file.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
