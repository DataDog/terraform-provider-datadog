package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &datadogFrameworkProvider{}
)

type datadogFrameworkProvider struct{}

// Provider schema struct
type datadogProviderSchema struct {
	ApiKey                 types.String `tfsdk:"api_key"`
	AppKey                 types.String `tfsdk:"app_key"`
	ApiUrl                 types.String `tfsdk:"api_url"`
	Validate               types.Bool   `tfsdk:"validate"`
	HttpClientRetryEnabled types.Bool   `tfsdk:"http_client_retry_enabled"`
	HttpClientRetryTimeout types.Int64  `tfsdk:"http_client_retry_timeout"`
}

func New() provider.Provider {
	return &datadogFrameworkProvider{}
}

func (p *datadogFrameworkProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "datadog"
}

func (p *datadogFrameworkProvider) MetaSchema(ctx context.Context, request provider.MetaSchemaRequest, response *provider.MetaSchemaResponse) {
}

func (p *datadogFrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "(Required unless validate is false) Datadog API key. This can also be set via the DD_API_KEY environment variable.",
			},
			"app_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "(Required unless validate is false) Datadog APP key. This can also be set via the DD_APP_KEY environment variable.",
			},
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "The API URL. This can also be set via the DD_HOST environment variable. Note that this URL must not end with the `/api/` path. For example, `https://api.datadoghq.com/` is a correct value, while `https://api.datadoghq.com/api/` is not. And if you're working with \"EU\" version of Datadog, use `https://api.datadoghq.eu/`. Other Datadog region examples: `https://api.us5.datadoghq.com/`, `https://api.us3.datadoghq.com/` and `https://api.ddog-gov.com/`. See https://docs.datadoghq.com/getting_started/site/ for all available regions.",
			},
			"validate": schema.BoolAttribute{
				Optional:    true,
				Description: "Enables validation of the provided API and APP keys during provider initialization. Default is true. When false, api_key and app_key won't be checked.",
			},
			"http_client_retry_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enables request retries on HTTP status codes 429 and 5xx. Defaults to `true`.",
			},
			"http_client_retry_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry timeout period. Defaults to 60 seconds.",
			},
		},
	}
}

func (p *datadogFrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config datadogProviderSchema
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (p *datadogFrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		//func() resource.Resource {
		//	return nil
		//},
	}
}

func (p *datadogFrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//func() datasource.DataSource {
		//	return nil
		//},
	}
}
