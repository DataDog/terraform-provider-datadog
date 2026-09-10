package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PrometheusRemoteWriteSourceModel represents the Terraform model for the Prometheus Remote Write source
type PrometheusRemoteWriteSourceModel struct {
	AuthStrategy types.String                                 `tfsdk:"auth_strategy"`
	AddressKey   types.String                                 `tfsdk:"address_key"`
	Path         types.String                                 `tfsdk:"path"`
	UsernameKey  types.String                                 `tfsdk:"username_key"`
	PasswordKey  types.String                                 `tfsdk:"password_key"`
	ValidTokens  []PrometheusRemoteWriteSourceValidTokenModel `tfsdk:"valid_token"`
	Tls          []MtlsServerTlsModel                         `tfsdk:"tls"`
}

// PrometheusRemoteWriteSourceValidTokenModel represents an accepted token for authenticating incoming requests
type PrometheusRemoteWriteSourceValidTokenModel struct {
	TokenKey    types.String                                            `tfsdk:"token_key"`
	Enabled     types.Bool                                              `tfsdk:"enabled"`
	PathToToken []PrometheusRemoteWriteSourceValidTokenPathToTokenModel `tfsdk:"path_to_token"`
}

// PrometheusRemoteWriteSourceValidTokenPathToTokenModel specifies where the token is extracted from the incoming request
type PrometheusRemoteWriteSourceValidTokenPathToTokenModel struct {
	Location types.String `tfsdk:"location"`
	Header   types.String `tfsdk:"header"`
}

// ExpandPrometheusRemoteWriteSource converts the Terraform model to the Datadog API model
func ExpandPrometheusRemoteWriteSource(src *PrometheusRemoteWriteSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelinePrometheusRemoteWriteSourceWithDefaults()
	s.SetId(id)
	s.SetAuthStrategy(datadogV2.ObservabilityPipelinePrometheusRemoteWriteSourceAuthStrategy(src.AuthStrategy.ValueString()))

	if !src.AddressKey.IsNull() {
		s.SetAddressKey(src.AddressKey.ValueString())
	}
	if !src.Path.IsNull() {
		s.SetPath(src.Path.ValueString())
	}
	if !src.UsernameKey.IsNull() {
		s.SetUsernameKey(src.UsernameKey.ValueString())
	}
	if !src.PasswordKey.IsNull() {
		s.SetPasswordKey(src.PasswordKey.ValueString())
	}
	if len(src.ValidTokens) > 0 {
		s.SetValidTokens(expandPrometheusRemoteWriteValidTokens(src.ValidTokens))
	}
	if len(src.Tls) > 0 {
		s.Tls = ExpandMtlsServerTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelinePrometheusRemoteWriteSource: s,
	}
}

// FlattenPrometheusRemoteWriteSource converts the Datadog API model to the Terraform model
func FlattenPrometheusRemoteWriteSource(src *datadogV2.ObservabilityPipelinePrometheusRemoteWriteSource) *PrometheusRemoteWriteSourceModel {
	if src == nil {
		return nil
	}

	out := &PrometheusRemoteWriteSourceModel{
		AuthStrategy: types.StringValue(string(src.GetAuthStrategy())),
	}

	if v, ok := src.GetAddressKeyOk(); ok {
		out.AddressKey = types.StringValue(*v)
	}
	if v, ok := src.GetPathOk(); ok {
		out.Path = types.StringValue(*v)
	}
	if v, ok := src.GetUsernameKeyOk(); ok {
		out.UsernameKey = types.StringValue(*v)
	}
	if v, ok := src.GetPasswordKeyOk(); ok {
		out.PasswordKey = types.StringValue(*v)
	}
	if tokens, ok := src.GetValidTokensOk(); ok && tokens != nil {
		out.ValidTokens = flattenPrometheusRemoteWriteValidTokens(*tokens)
	}
	if src.Tls != nil {
		out.Tls = FlattenMtlsServerTls(src.Tls)
	}

	return out
}

func expandPrometheusRemoteWriteValidTokens(src []PrometheusRemoteWriteSourceValidTokenModel) []datadogV2.ObservabilityPipelinePrometheusRemoteWriteSourceValidToken {
	out := make([]datadogV2.ObservabilityPipelinePrometheusRemoteWriteSourceValidToken, 0, len(src))
	for _, t := range src {
		token := datadogV2.NewObservabilityPipelinePrometheusRemoteWriteSourceValidTokenWithDefaults()
		token.SetTokenKey(t.TokenKey.ValueString())
		if !t.Enabled.IsNull() && !t.Enabled.IsUnknown() {
			token.SetEnabled(t.Enabled.ValueBool())
		}
		if len(t.PathToToken) > 0 {
			token.SetPathToToken(expandPrometheusRemoteWriteValidTokenPathToToken(&t.PathToToken[0]))
		}
		out = append(out, *token)
	}
	return out
}

func flattenPrometheusRemoteWriteValidTokens(src []datadogV2.ObservabilityPipelinePrometheusRemoteWriteSourceValidToken) []PrometheusRemoteWriteSourceValidTokenModel {
	out := make([]PrometheusRemoteWriteSourceValidTokenModel, 0, len(src))
	for _, t := range src {
		model := PrometheusRemoteWriteSourceValidTokenModel{
			TokenKey: types.StringValue(t.GetTokenKey()),
			Enabled:  types.BoolValue(t.GetEnabled()),
		}
		if pt, ok := t.GetPathToTokenOk(); ok && pt != nil {
			if mapped := flattenPrometheusRemoteWriteValidTokenPathToToken(pt); mapped != nil {
				model.PathToToken = []PrometheusRemoteWriteSourceValidTokenPathToTokenModel{*mapped}
			}
		}
		out = append(out, model)
	}
	return out
}

// expandPrometheusRemoteWriteValidTokenPathToToken maps into the shared HTTP server valid-token path-to-token
// union type, which the prometheus_remote_write source's valid_tokens[].path_to_token also uses per the API spec.
func expandPrometheusRemoteWriteValidTokenPathToToken(src *PrometheusRemoteWriteSourceValidTokenPathToTokenModel) datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToToken {
	if !src.Header.IsNull() && src.Header.ValueString() != "" {
		header := datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToTokenHeader{
			Header: src.Header.ValueString(),
		}
		return datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToTokenHeaderAsObservabilityPipelineHttpServerSourceValidTokenPathToToken(&header)
	}
	location := datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToTokenLocation(src.Location.ValueString())
	return datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToTokenLocationAsObservabilityPipelineHttpServerSourceValidTokenPathToToken(&location)
}

func flattenPrometheusRemoteWriteValidTokenPathToToken(src *datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToToken) *PrometheusRemoteWriteSourceValidTokenPathToTokenModel {
	switch v := src.GetActualInstance().(type) {
	case *datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToTokenLocation:
		return &PrometheusRemoteWriteSourceValidTokenPathToTokenModel{
			Location: types.StringValue(string(*v)),
		}
	case *datadogV2.ObservabilityPipelineHttpServerSourceValidTokenPathToTokenHeader:
		return &PrometheusRemoteWriteSourceValidTokenPathToTokenModel{
			Header: types.StringValue(v.Header),
		}
	}
	return nil
}

// PrometheusRemoteWriteSourceSchema returns the schema for the Prometheus Remote Write source
func PrometheusRemoteWriteSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `prometheus_remote_write` source ingests metrics pushed over the Prometheus Remote Write protocol.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"auth_strategy": schema.StringAttribute{
					Required:    true,
					Description: "HTTP authentication method.",
					Validators: []validator.String{
						stringvalidator.OneOf("none", "plain"),
					},
				},
				"address_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the listen address for the Prometheus Remote Write endpoint.",
				},
				"path": schema.StringAttribute{
					Optional:    true,
					Description: "The HTTP path on which the source listens for incoming Prometheus Remote Write requests. Defaults to `/api/v1/write`.",
				},
				"username_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the username. Used when `auth_strategy` is `plain`.",
				},
				"password_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the password. Used when `auth_strategy` is `plain`.",
				},
			},
			Blocks: map[string]schema.Block{
				"tls": MtlsServerTlsSchema(),
				"valid_token": schema.ListNestedBlock{
					Description: "A token accepted for authenticating incoming Prometheus Remote Write requests. When set, the source rejects any request whose token does not match an enabled entry in this list.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1000),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"token_key": schema.StringAttribute{
								Required:    true,
								Description: "Name of the environment variable or secret that holds the expected token value.",
							},
							"enabled": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Whether this token is currently accepted. Defaults to `true`.",
							},
						},
						Blocks: map[string]schema.Block{
							"path_to_token": schema.ListNestedBlock{
								Description: "Specifies where the worker extracts the token from the incoming HTTP request. Set either `location` for a built-in source or `header` to read it from a request header.",
								Validators: []validator.List{
									listvalidator.SizeAtMost(1),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"location": schema.StringAttribute{
											Optional:    true,
											Description: "Built-in token location on the incoming HTTP request. One of `path`, `address`.",
											Validators: []validator.String{
												stringvalidator.OneOf("path", "address"),
											},
										},
										"header": schema.StringAttribute{
											Optional:    true,
											Description: "The name of the HTTP header that carries the token.",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
