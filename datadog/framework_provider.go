package datadog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datadogCommunity "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/transport"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ provider.Provider = &FrameworkProvider{}
)

type FrameworkProvider struct {
	CommunityClient     *datadogCommunity.Client
	DatadogApiInstances *utils.ApiInstances
	Auth                context.Context

	Now func() time.Time
}

// Provider schema struct
type providerSchema struct {
	ApiKey                 types.String `tfsdk:"api_key"`
	AppKey                 types.String `tfsdk:"app_key"`
	ApiUrl                 types.String `tfsdk:"api_url"`
	Validate               types.Bool   `tfsdk:"validate"`
	HttpClientRetryEnabled types.Bool   `tfsdk:"http_client_retry_enabled"`
	HttpClientRetryTimeout types.Int64  `tfsdk:"http_client_retry_timeout"`
}

func New() provider.Provider {
	return &FrameworkProvider{}
}

func (p *FrameworkProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "datadog_"
}

func (p *FrameworkProvider) MetaSchema(ctx context.Context, request provider.MetaSchemaRequest, response *provider.MetaSchemaResponse) {
}

func (p *FrameworkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
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

func (p *FrameworkProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	if p.Auth != nil && p.CommunityClient != nil && p.DatadogApiInstances != nil {
		response.DataSourceData = p
		response.ResourceData = p
		return
	}

	var config providerSchema
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	if config.ApiKey.IsNull() {
		apiKey, err := utils.GetMultiEnvVar(APIKeyEnvVars[:]...)
		if err == nil {
			config.ApiKey = types.StringValue(apiKey)
		}
	}

	if config.AppKey.IsNull() {
		appKey, err := utils.GetMultiEnvVar(APPKeyEnvVars[:]...)
		if err == nil {
			config.AppKey = types.StringValue(appKey)
		}
	}

	if config.ApiUrl.IsNull() {
		apiUrl, err := utils.GetMultiEnvVar(APIUrlEnvVars[:]...)
		if err == nil {
			config.ApiUrl = types.StringValue(apiUrl)
		}
	}

	if config.Validate.ValueBool() && (config.ApiKey.ValueString() == "" || config.AppKey.ValueString() == "") {
		response.Diagnostics.AddError("api_key and app_key must be set unless validate = false", "")
		return
	}

	if config.HttpClientRetryEnabled.IsNull() {
		retryEnabled, err := utils.GetMultiEnvVar("DD_HTTP_CLIENT_RETRY_ENABLED")
		if err == nil {
			v, _ := strconv.ParseBool(retryEnabled)
			config.HttpClientRetryEnabled = types.BoolValue(v)
		}
	}

	if config.HttpClientRetryTimeout.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar("DD_HTTP_CLIENT_RETRY_TIMEOUT")
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryTimeout = types.Int64Value(int64(v))
		}
	}

	// Initialize the community client
	p.CommunityClient = datadogCommunity.NewClient(config.ApiKey.ValueString(), config.AppKey.ValueString())
	if !config.ApiUrl.IsNull() {
		p.CommunityClient.SetBaseUrl(config.ApiUrl.ValueString())
	}
	c := cleanhttp.DefaultClient()
	p.CommunityClient.ExtraHeader["User-Agent"] = utils.GetUserAgent(fmt.Sprintf(
		"datadog-api-client-go/%s (go %s; os %s; arch %s)",
		"go-datadog-api",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	))
	p.CommunityClient.HttpClient = c

	if config.Validate.ValueBool() {
		log.Println("[INFO] Datadog client successfully initialized, now validating...")
		ok, err := p.CommunityClient.Validate()
		if err != nil {
			response.Diagnostics.AddError("[ERROR] Datadog Client validation error", err.Error())
			return
		} else if !ok {
			err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`)
			response.Diagnostics.AddError("[ERROR] Datadog Client validation error", err.Error())
			return
		}
	} else {
		log.Println("[INFO] Skipping key validation (validate = false)")
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	// Initialize http.Client for the Datadog API Clients
	httpClient := http.DefaultClient
	if config.HttpClientRetryEnabled.ValueBool() {
		ctOptions := transport.CustomTransportOptions{}
		if !config.HttpClientRetryTimeout.IsNull() {
			timeout := time.Duration(config.HttpClientRetryTimeout.ValueInt64()) * time.Second
			ctOptions.Timeout = &timeout
		}
		customTransport := transport.NewCustomTransport(httpClient.Transport, ctOptions)
		httpClient.Transport = customTransport
	}

	// Initialize the official Datadog V1 API client
	auth := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: config.ApiKey.ValueString(),
			},
			"appKeyAuth": {
				Key: config.AppKey.ValueString(),
			},
		},
	)
	ddClientConfig := datadog.NewConfiguration()
	ddClientConfig.HTTPClient = httpClient
	ddClientConfig.UserAgent = utils.GetUserAgent(ddClientConfig.UserAgent)

	if !config.ApiUrl.IsNull() {
		parsedAPIURL, parseErr := url.Parse(config.ApiUrl.ValueString())
		if parseErr != nil {
			response.Diagnostics.AddError("invalid API URL", parseErr.Error())
			return
		}
		if parsedAPIURL.Host == "" || parsedAPIURL.Scheme == "" {
			response.Diagnostics.AddError("missing protocol or host", parseErr.Error())
			return
		}
		// If api url is passed, set and use the api name and protocol on ServerIndex{1}
		auth = context.WithValue(auth, datadog.ContextServerIndex, 1)
		auth = context.WithValue(auth, datadog.ContextServerVariables, map[string]string{
			"name":     parsedAPIURL.Host,
			"protocol": parsedAPIURL.Scheme,
		})

		// Configure URL's per operation
		// IPRangesApiService.GetIPRanges
		ipRangesDNSNameArr := strings.Split(parsedAPIURL.Hostname(), ".")
		// Parse out subdomain if it exists
		if len(ipRangesDNSNameArr) > 2 {
			ipRangesDNSNameArr = ipRangesDNSNameArr[1:]
		}
		ipRangesDNSNameArr = append([]string{baseIPRangesSubdomain}, ipRangesDNSNameArr...)

		auth = context.WithValue(auth, datadog.ContextOperationServerIndices, map[string]int{
			"v1.IPRangesApi.GetIPRanges": 1,
		})
		auth = context.WithValue(auth, datadog.ContextOperationServerVariables, map[string]map[string]string{
			"v1.IPRangesApi.GetIPRanges": {
				"name": strings.Join(ipRangesDNSNameArr, "."),
			},
		})
	}

	datadogClient := datadog.NewAPIClient(ddClientConfig)

	p.DatadogApiInstances = &utils.ApiInstances{HttpClient: datadogClient}
	p.Auth = auth

	// Make config available for data sources and resources
	response.DataSourceData = p
	response.ResourceData = p
}

func (p *FrameworkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		//func() resource.Resource {
		//	return nil
		//},
	}
}

func (p *FrameworkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIPRangesDataSource,
	}
}
