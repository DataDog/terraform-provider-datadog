package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WebsocketSourceTlsValidator enforces mode-specific field rules for the TLS block:
//   - mode = "enabled": crt_file, ca_file, key_file, key_pass_key must NOT be set.
//   - mode = "with_client_cert": crt_file must be set (non-empty).
type WebsocketSourceTlsValidator struct{}

func (v WebsocketSourceTlsValidator) Description(_ context.Context) string {
	return "validates mode-specific TLS field requirements"
}

func (v WebsocketSourceTlsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v WebsocketSourceTlsValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()

	modeAttr, ok := attrs["mode"]
	if !ok || modeAttr.IsNull() || modeAttr.IsUnknown() {
		return
	}
	mode := modeAttr.(types.String).ValueString()

	certFields := []string{"crt_file", "ca_file", "key_file", "key_pass_key"}

	// isKnownAndSet returns true only when the attribute is known (not null, not unknown)
	// and contains a non-empty string value. Unknown values are skipped so that
	// plan-time validation does not fail when a field is derived from another resource.
	isKnownAndSet := func(name string) bool {
		attr, exists := attrs[name]
		if !exists || attr.IsNull() || attr.IsUnknown() {
			return false
		}
		s, ok := attr.(types.String)
		return ok && s.ValueString() != ""
	}

	// isKnownAndMissing returns true only when the attribute is known (not unknown)
	// and is either null or an empty string. Unknown values are skipped so that
	// plan-time validation does not fail when a field will be set at apply time.
	isKnownAndMissing := func(name string) bool {
		attr, exists := attrs[name]
		if !exists {
			return true
		}
		if attr.IsUnknown() {
			return false
		}
		if attr.IsNull() {
			return true
		}
		s, ok := attr.(types.String)
		return !ok || s.ValueString() == ""
	}

	switch mode {
	case "enabled":
		for _, field := range certFields {
			if isKnownAndSet(field) {
				resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
					req.Path,
					"Invalid TLS Configuration",
					"When 'mode' is 'enabled', '"+field+"' must not be set. Certificate fields are only valid with mode 'with_client_cert'.",
				))
			}
		}
	case "with_client_cert":
		if isKnownAndMissing("crt_file") {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
				req.Path,
				"Missing Required Field",
				"'crt_file' is required when 'mode' is 'with_client_cert'.",
			))
		}
	}
}

// WebsocketSourceTlsModel represents the TLS configuration for the websocket source.
// The tls block uses a mode discriminator: "enabled" or "with_client_cert".
// When mode is "with_client_cert", crt_file is required and ca_file/key_file/key_pass_key
// are optional. This mirrors the OpenAPI oneOf shape:
//
//	ObservabilityPipelineWebsocketSourceTls:
//	  oneOf:
//	    - ObservabilityPipelineWebsocketSourceTlsEnabled     (mode: "enabled")
//	    - ObservabilityPipelineWebsocketSourceTlsWithClientCert (mode: "with_client_cert",
//	                                                              crt_file required)
type WebsocketSourceTlsModel struct {
	Mode       types.String `tfsdk:"mode"`
	CrtFile    types.String `tfsdk:"crt_file"`
	CaFile     types.String `tfsdk:"ca_file"`
	KeyFile    types.String `tfsdk:"key_file"`
	KeyPassKey types.String `tfsdk:"key_pass_key"`
}

// WebsocketSourceModel represents the Terraform model for websocket source configuration.
type WebsocketSourceModel struct {
	UriKey       types.String              `tfsdk:"uri_key"`
	Decoding     types.String              `tfsdk:"decoding"`
	AuthStrategy types.String              `tfsdk:"auth_strategy"`
	UsernameKey  types.String              `tfsdk:"username_key"`
	PasswordKey  types.String              `tfsdk:"password_key"`
	TokenKey     types.String              `tfsdk:"token_key"`
	CustomKey    types.String              `tfsdk:"custom_key"`
	Tls          []WebsocketSourceTlsModel `tfsdk:"tls"`
}

// ExpandWebsocketSource converts the Terraform model to the Datadog API model.
func ExpandWebsocketSource(src *WebsocketSourceModel, id string) (datadogV2.ObservabilityPipelineConfigSourceItem, diag.Diagnostics) {
	var diags diag.Diagnostics

	s := datadogV2.NewObservabilityPipelineWebsocketSourceWithDefaults()
	s.SetId(id)
	s.SetDecoding(datadogV2.ObservabilityPipelineDecoding(src.Decoding.ValueString()))
	s.SetAuthStrategy(datadogV2.ObservabilityPipelineWebsocketSourceAuthStrategy(src.AuthStrategy.ValueString()))

	if !src.UriKey.IsNull() {
		s.SetUriKey(src.UriKey.ValueString())
	}
	if !src.UsernameKey.IsNull() {
		s.SetUsernameKey(src.UsernameKey.ValueString())
	}
	if !src.PasswordKey.IsNull() {
		s.SetPasswordKey(src.PasswordKey.ValueString())
	}
	if !src.TokenKey.IsNull() {
		s.SetTokenKey(src.TokenKey.ValueString())
	}
	if !src.CustomKey.IsNull() {
		s.SetCustomKey(src.CustomKey.ValueString())
	}

	if len(src.Tls) > 0 {
		tlsItem := src.Tls[0]
		switch tlsItem.Mode.ValueString() {
		case "enabled":
			s.Tls = &datadogV2.ObservabilityPipelineWebsocketSourceTls{
				ObservabilityPipelineWebsocketSourceTlsEnabled: &datadogV2.ObservabilityPipelineWebsocketSourceTlsEnabled{
					Mode: datadogV2.OBSERVABILITYPIPELINEWEBSOCKETSOURCETLSENABLEDMODE_ENABLED,
				},
			}
		case "with_client_cert":
			withCert := &datadogV2.ObservabilityPipelineWebsocketSourceTlsWithClientCert{
				Mode:    datadogV2.OBSERVABILITYPIPELINEWEBSOCKETSOURCETLSWITHCLIENTCERTMODE_WITH_CLIENT_CERT,
				CrtFile: tlsItem.CrtFile.ValueString(),
			}
			if !tlsItem.CaFile.IsNull() {
				withCert.SetCaFile(tlsItem.CaFile.ValueString())
			}
			if !tlsItem.KeyFile.IsNull() {
				withCert.SetKeyFile(tlsItem.KeyFile.ValueString())
			}
			if !tlsItem.KeyPassKey.IsNull() {
				withCert.SetKeyPassKey(tlsItem.KeyPassKey.ValueString())
			}
			s.Tls = &datadogV2.ObservabilityPipelineWebsocketSourceTls{
				ObservabilityPipelineWebsocketSourceTlsWithClientCert: withCert,
			}
		}
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineWebsocketSource: s,
	}, diags
}

// FlattenWebsocketSource converts the Datadog API model to the Terraform model.
func FlattenWebsocketSource(src *datadogV2.ObservabilityPipelineWebsocketSource) *WebsocketSourceModel {
	if src == nil {
		return nil
	}

	out := &WebsocketSourceModel{
		Decoding:     types.StringValue(string(src.GetDecoding())),
		AuthStrategy: types.StringValue(string(src.GetAuthStrategy())),
	}

	if v, ok := src.GetUriKeyOk(); ok {
		out.UriKey = types.StringValue(*v)
	}
	if v, ok := src.GetUsernameKeyOk(); ok {
		out.UsernameKey = types.StringValue(*v)
	}
	if v, ok := src.GetPasswordKeyOk(); ok {
		out.PasswordKey = types.StringValue(*v)
	}
	if v, ok := src.GetTokenKeyOk(); ok {
		out.TokenKey = types.StringValue(*v)
	}
	if v, ok := src.GetCustomKeyOk(); ok {
		out.CustomKey = types.StringValue(*v)
	}

	if src.Tls != nil {
		tlsModel := WebsocketSourceTlsModel{}
		if src.Tls.ObservabilityPipelineWebsocketSourceTlsEnabled != nil {
			tlsModel.Mode = types.StringValue("enabled")
			tlsModel.CrtFile = types.StringNull()
			tlsModel.CaFile = types.StringNull()
			tlsModel.KeyFile = types.StringNull()
			tlsModel.KeyPassKey = types.StringNull()
		} else if src.Tls.ObservabilityPipelineWebsocketSourceTlsWithClientCert != nil {
			cert := src.Tls.ObservabilityPipelineWebsocketSourceTlsWithClientCert
			tlsModel.Mode = types.StringValue("with_client_cert")
			tlsModel.CrtFile = types.StringValue(cert.CrtFile)
			tlsModel.CaFile = types.StringPointerValue(cert.CaFile)
			tlsModel.KeyFile = types.StringPointerValue(cert.KeyFile)
			if v, ok := cert.GetKeyPassKeyOk(); ok {
				tlsModel.KeyPassKey = types.StringValue(*v)
			} else {
				tlsModel.KeyPassKey = types.StringNull()
			}
		}
		out.Tls = []WebsocketSourceTlsModel{tlsModel}
	}

	return out
}

// WebsocketSourceSchema returns the schema for the websocket source block.
func WebsocketSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `websocket` source establishes a persistent WebSocket connection to a remote endpoint and ingests log events as they are pushed by the server.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"uri_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the WebSocket URI to connect to.",
				},
				"decoding": schema.StringAttribute{
					Required:    true,
					Description: "The decoding format used to interpret incoming log events.",
					Validators: []validator.String{
						stringvalidator.OneOf("bytes", "gelf", "json", "syslog"),
					},
				},
				"auth_strategy": schema.StringAttribute{
					Required:    true,
					Description: "The authentication strategy used when connecting to the WebSocket server.",
					Validators: []validator.String{
						stringvalidator.OneOf("none", "basic", "bearer", "custom"),
					},
				},
				"username_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the username. Used when `auth_strategy` is `basic`.",
				},
				"password_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the password. Used when `auth_strategy` is `basic`.",
				},
				"token_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the bearer token. Used when `auth_strategy` is `bearer`.",
				},
				"custom_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds a custom header value. Used when `auth_strategy` is `custom`.",
				},
			},
			Blocks: map[string]schema.Block{
				"tls": schema.ListNestedBlock{
					Description: "TLS configuration for the WebSocket connection. Set `mode` to `enabled` for server-certificate validation only, or `with_client_cert` to additionally present a client certificate.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Required:    true,
								Description: "The TLS mode. Use `enabled` for server-only TLS, or `with_client_cert` for mutual TLS with a client certificate.",
								Validators: []validator.String{
									stringvalidator.OneOf("enabled", "with_client_cert"),
								},
							},
							"crt_file": schema.StringAttribute{
								Optional:    true,
								Description: "Path to the client certificate file. Required when `mode` is `with_client_cert`.",
							},
							"ca_file": schema.StringAttribute{
								Optional:    true,
								Description: "Path to the Certificate Authority (CA) file used to validate the server's TLS certificate.",
							},
							"key_file": schema.StringAttribute{
								Optional:    true,
								Description: "Path to the private key file associated with the client certificate.",
							},
							"key_pass_key": schema.StringAttribute{
								Optional:    true,
								Description: "Name of the environment variable or secret that holds the passphrase for the private key file.",
							},
						},
						Validators: []validator.Object{
							WebsocketSourceTlsValidator{},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
