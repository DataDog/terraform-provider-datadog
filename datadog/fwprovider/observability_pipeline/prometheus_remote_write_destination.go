package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// PrometheusRemoteWriteDestinationModel represents the Terraform model for the Prometheus Remote Write destination
type PrometheusRemoteWriteDestinationModel struct {
	EndpointUrlKey   types.String         `tfsdk:"endpoint_url_key"`
	DefaultNamespace types.String         `tfsdk:"default_namespace"`
	TenantId         types.String         `tfsdk:"tenant_id"`
	AuthStrategy     types.String         `tfsdk:"auth_strategy"`
	UsernameKey      types.String         `tfsdk:"username_key"`
	PasswordKey      types.String         `tfsdk:"password_key"`
	TokenKey         types.String         `tfsdk:"token_key"`
	Tls              []ClientTlsModel     `tfsdk:"tls"`
	Buffer           []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandPrometheusRemoteWriteDestination converts the Terraform model to the Datadog API model
func ExpandPrometheusRemoteWriteDestination(ctx context.Context, id string, inputs types.List, src *PrometheusRemoteWriteDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelinePrometheusRemoteWriteDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	if !src.EndpointUrlKey.IsNull() {
		dest.SetEndpointUrlKey(src.EndpointUrlKey.ValueString())
	}
	if !src.DefaultNamespace.IsNull() {
		dest.SetDefaultNamespace(src.DefaultNamespace.ValueString())
	}
	if !src.TenantId.IsNull() {
		dest.SetTenantId(src.TenantId.ValueString())
	}
	if !src.AuthStrategy.IsNull() {
		dest.SetAuthStrategy(datadogV2.ObservabilityPipelinePrometheusRemoteWriteDestinationAuthStrategy(src.AuthStrategy.ValueString()))
	}
	if !src.UsernameKey.IsNull() {
		dest.SetUsernameKey(src.UsernameKey.ValueString())
	}
	if !src.PasswordKey.IsNull() {
		dest.SetPasswordKey(src.PasswordKey.ValueString())
	}
	if !src.TokenKey.IsNull() {
		dest.SetTokenKey(src.TokenKey.ValueString())
	}
	if len(src.Tls) > 0 {
		dest.Tls = ExpandClientTls(src.Tls)
	}
	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelinePrometheusRemoteWriteDestination: dest,
	}
}

// FlattenPrometheusRemoteWriteDestination converts the Datadog API model to the Terraform model
func FlattenPrometheusRemoteWriteDestination(src *datadogV2.ObservabilityPipelinePrometheusRemoteWriteDestination) *PrometheusRemoteWriteDestinationModel {
	if src == nil {
		return nil
	}

	out := &PrometheusRemoteWriteDestinationModel{}

	if v, ok := src.GetEndpointUrlKeyOk(); ok {
		out.EndpointUrlKey = types.StringValue(*v)
	}
	if v, ok := src.GetDefaultNamespaceOk(); ok {
		out.DefaultNamespace = types.StringValue(*v)
	}
	if v, ok := src.GetTenantIdOk(); ok {
		out.TenantId = types.StringValue(*v)
	}
	if v, ok := src.GetAuthStrategyOk(); ok {
		out.AuthStrategy = types.StringValue(string(*v))
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
	if src.Tls != nil {
		out.Tls = FlattenClientTls(src.Tls)
	}
	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return out
}

// PrometheusRemoteWriteDestinationSchema returns the schema for the Prometheus Remote Write destination
func PrometheusRemoteWriteDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `prometheus_remote_write` destination sends metrics to a Prometheus Remote Write compatible endpoint.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"endpoint_url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the Prometheus Remote Write endpoint URL.",
				},
				"default_namespace": schema.StringAttribute{
					Optional:    true,
					Description: "The default namespace to prefix onto metric names that don't already have one.",
				},
				"tenant_id": schema.StringAttribute{
					Optional:    true,
					Description: "The tenant ID to include with outgoing requests. Used by multi-tenant Prometheus Remote Write receivers.",
				},
				"auth_strategy": schema.StringAttribute{
					Optional:    true,
					Description: "The authentication strategy to use for outgoing Prometheus Remote Write requests.",
					Validators: []validator.String{
						stringvalidator.OneOf("none", "basic", "bearer"),
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
			},
			Blocks: map[string]schema.Block{
				"tls":    ClientTlsSchema(),
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
