package observability_pipeline

import (
	"regexp"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// MtlsServerTlsWithClientIdentityModel represents mTLS server TLS configuration for sources that support client identity checks
type MtlsServerTlsWithClientIdentityModel struct {
	CrtFile           types.String          `tfsdk:"crt_file"`
	CaFile            types.String          `tfsdk:"ca_file"`
	KeyFile           types.String          `tfsdk:"key_file"`
	KeyPassKey        types.String          `tfsdk:"key_pass_key"`
	VerifyCertificate types.Bool            `tfsdk:"verify_certificate"`
	ClientIdentity    []ClientIdentityModel `tfsdk:"client_identity"`
}

type ClientIdentityModel struct {
	Mode          types.String                   `tfsdk:"mode"`
	AllowlistFile types.String                   `tfsdk:"allowlist_file"`
	Allowlist     []ClientIdentityAllowlistModel `tfsdk:"allowlist"`
}

type ClientIdentityAllowlistModel struct {
	Rules []ClientIdentityRuleModel `tfsdk:"rule"`
}

type ClientIdentityRuleModel struct {
	Name        types.String   `tfsdk:"name"`
	CommonNames []types.String `tfsdk:"common_names"`
	DnsSans     []types.String `tfsdk:"dns_sans"`
	UriSans     []types.String `tfsdk:"uri_sans"`
}

// ExpandMtlsServerTlsWithClientIdentity converts the Terraform model to the Datadog API model
func ExpandMtlsServerTlsWithClientIdentity(tlsTF []MtlsServerTlsWithClientIdentityModel) *datadogV2.ObservabilityPipelineMtlsServerTlsWithClientIdentity {
	if len(tlsTF) == 0 {
		return nil
	}
	item := tlsTF[0]
	tls := datadogV2.NewObservabilityPipelineMtlsServerTlsWithClientIdentity(item.CrtFile.ValueString())
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
	if len(item.ClientIdentity) > 0 {
		tls.SetClientIdentity(expandClientIdentity(item.ClientIdentity[0]))
	}
	return tls
}

func expandClientIdentity(src ClientIdentityModel) datadogV2.ObservabilityPipelineClientIdentity {
	identity := datadogV2.NewObservabilityPipelineClientIdentity()
	if !src.Mode.IsNull() && !src.Mode.IsUnknown() {
		identity.SetMode(datadogV2.ObservabilityPipelineClientIdentityMode(src.Mode.ValueString()))
	}
	if !src.AllowlistFile.IsNull() {
		identity.SetAllowlistFile(src.AllowlistFile.ValueString())
	}
	if len(src.Allowlist) > 0 {
		rules := make([]datadogV2.ObservabilityPipelineClientIdentityRule, 0, len(src.Allowlist[0].Rules))
		for _, ruleTF := range src.Allowlist[0].Rules {
			rule := datadogV2.NewObservabilityPipelineClientIdentityRule(ruleTF.Name.ValueString())
			if len(ruleTF.CommonNames) > 0 {
				rule.SetCommonNames(expandStringValues(ruleTF.CommonNames))
			}
			if len(ruleTF.DnsSans) > 0 {
				rule.SetDnsSans(expandStringValues(ruleTF.DnsSans))
			}
			if len(ruleTF.UriSans) > 0 {
				rule.SetUriSans(expandStringValues(ruleTF.UriSans))
			}
			rules = append(rules, *rule)
		}
		identity.SetAllowlist(*datadogV2.NewObservabilityPipelineClientIdentityAllowlist(rules))
	}
	return *identity
}

// FlattenMtlsServerTlsWithClientIdentity converts the Datadog API model to the Terraform model
func FlattenMtlsServerTlsWithClientIdentity(src *datadogV2.ObservabilityPipelineMtlsServerTlsWithClientIdentity) []MtlsServerTlsWithClientIdentityModel {
	if src == nil {
		return []MtlsServerTlsWithClientIdentityModel{}
	}
	out := MtlsServerTlsWithClientIdentityModel{
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
	if identity, ok := src.GetClientIdentityOk(); ok {
		out.ClientIdentity = []ClientIdentityModel{flattenClientIdentity(identity)}
	}
	return []MtlsServerTlsWithClientIdentityModel{out}
}

func flattenClientIdentity(src *datadogV2.ObservabilityPipelineClientIdentity) ClientIdentityModel {
	// Mode is defaulted in the schema, so an omitted mode must read back as the API default.
	out := ClientIdentityModel{
		Mode: types.StringValue(string(datadogV2.OBSERVABILITYPIPELINECLIENTIDENTITYMODE_ENFORCE)),
	}
	if v, ok := src.GetModeOk(); ok {
		out.Mode = types.StringValue(string(*v))
	}
	if v, ok := src.GetAllowlistFileOk(); ok {
		out.AllowlistFile = types.StringValue(*v)
	}
	if allowlist, ok := src.GetAllowlistOk(); ok {
		rules := make([]ClientIdentityRuleModel, 0, len(allowlist.Rules))
		for _, rule := range allowlist.Rules {
			rules = append(rules, ClientIdentityRuleModel{
				Name:        types.StringValue(rule.Name),
				CommonNames: flattenStringValues(rule.CommonNames),
				DnsSans:     flattenStringValues(rule.DnsSans),
				UriSans:     flattenStringValues(rule.UriSans),
			})
		}
		out.Allowlist = []ClientIdentityAllowlistModel{{Rules: rules}}
	}
	return out
}

func expandStringValues(values []types.String) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.ValueString())
	}
	return out
}

func flattenStringValues(values []string) []types.String {
	if values == nil {
		return nil
	}
	out := make([]types.String, 0, len(values))
	for _, value := range values {
		out = append(out, types.StringValue(value))
	}
	return out
}

// MtlsServerTlsWithClientIdentitySchema returns the schema for mTLS server TLS configuration with client identity checks
func MtlsServerTlsWithClientIdentitySchema() schema.ListNestedBlock {
	tlsSchema := MtlsServerTlsSchema()
	tlsSchema.Description = "Configuration for enabling TLS encryption between the pipeline component and external connecting clients, or mutual TLS when `verify_certificate` is `true`. Mutual TLS also enables an optional allowlist of client certificate identities."
	tlsSchema.NestedObject.Blocks = map[string]schema.Block{
		"client_identity": clientIdentitySchema(),
	}
	return tlsSchema
}

func clientIdentitySchema() schema.ListNestedBlock {
	patternsAttribute := func(description string) schema.ListAttribute {
		return schema.ListAttribute{
			Optional:    true,
			ElementType: types.StringType,
			Description: description,
			Validators: []validator.List{
				listvalidator.SizeBetween(1, 1000),
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		}
	}
	return schema.ListNestedBlock{
		Description: "Restricts which clients may connect, based on the identities in their TLS certificate. Checks run during the TLS handshake, after the client certificate chain is validated against `ca_file`, and never relax that validation. Requires `verify_certificate = true` and `ca_file`. Set exactly one of `allowlist_file` or `allowlist`. When set, connections are closed after a maximum age of 300 seconds so that clients removed from the allowlist are disconnected. Requires Observability Pipelines Worker 2.27 or later.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"mode": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Default:     stringdefault.StaticString("enforce"),
					Description: "What happens to a client that no rule allows. `enforce` rejects the TLS handshake. `audit` accepts the client and reports that it would have been rejected.",
					Validators: []validator.String{
						stringvalidator.OneOf("enforce", "audit"),
					},
				},
				"allowlist_file": schema.StringAttribute{
					Optional:    true,
					Description: "Path to an allowlist file, relative to the Worker's data directory. The Worker reloads the file when it changes. Must not contain parent directory (`..`) references.",
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"allowlist": schema.ListNestedBlock{
					Description: "An allowlist of client identities. A client is allowed when any rule matches.",
					NestedObject: schema.NestedBlockObject{
						Blocks: map[string]schema.Block{
							"rule": schema.ListNestedBlock{
								Description: "A named rule, evaluated in order. The first matching rule is the one reported in metrics. It matches when every constraint it defines matches at least one identity in the client's certificate. It must define at least one of `common_names`, `dns_sans`, or `uri_sans`.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required:    true,
											Description: "Unique rule name, reported in metrics.",
											Validators: []validator.String{
												stringvalidator.LengthBetween(1, 64),
												stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9._-]+$`), "must contain only letters, digits, `.`, `_`, or `-`"),
											},
										},
										"common_names": patternsAttribute("Patterns for the certificate subject common name. Matching is case-sensitive, and `*` matches any characters, including dots."),
										"dns_sans":     patternsAttribute("Patterns for DNS subject alternative names. Matching is case-insensitive, and `*` matches any characters within one DNS label."),
										"uri_sans":     patternsAttribute("Patterns for URI subject alternative names. Matching is case-sensitive. `*` matches any characters within one path segment, and a trailing `/**` matches the path and everything below it. Wildcards are only supported in the path."),
									},
								},
								Validators: []validator.List{
									listvalidator.SizeBetween(1, 1000),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("allowlist_file")),
					},
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
