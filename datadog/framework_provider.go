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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	datadogCommunity "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/transport"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators/frameworkvalidators"
)

var (
	_ provider.Provider                     = &FrameworkProvider{}
	_ provider.ProviderWithConfigValidators = &FrameworkProvider{}
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
	Validate               types.String `tfsdk:"validate"`
	HttpClientRetryEnabled types.String `tfsdk:"http_client_retry_enabled"`
	HttpClientRetryTimeout types.Int64  `tfsdk:"http_client_retry_timeout"`
}

func New() provider.Provider {
	return &FrameworkProvider{}
}

func (p *FrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "datadog_"
}

func (p *FrameworkProvider) MetaSchema(_ context.Context, _ provider.MetaSchemaRequest, _ *provider.MetaSchemaResponse) {
}

func (p *FrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
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
			"validate": schema.StringAttribute{
				Optional:    true,
				Description: "Enables validation of the provided API and APP keys during provider initialization. Valid values are [`true`, `false`]. Default is true. When false, api_key and app_key won't be checked.",
			},
			"http_client_retry_enabled": schema.StringAttribute{
				Optional:    true,
				Description: "Enables request retries on HTTP status codes 429 and 5xx. Valid values are [`true`, `false`]. Defaults to `true`.",
			},
			"http_client_retry_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry timeout period. Defaults to 60 seconds.",
			},
		},
	}
}

func (p *FrameworkProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	// Provider has already been configured. This should only occur in testing scenario
	// where a custom HttpClient needs to be used
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

	p.ConfigureConfigDefaults(&config)

	validate, _ := strconv.ParseBool(config.Validate.ValueString())
	httpClientRetryEnabled, _ := strconv.ParseBool(config.Validate.ValueString())

	if validate && (config.ApiKey.ValueString() == "" || config.AppKey.ValueString() == "") {
		response.Diagnostics.AddError("api_key and app_key must be set unless validate = false", "")
		return
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

	if validate {
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
	if httpClientRetryEnabled {
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
	ddClientConfig.Debug = logging.IsDebugOrHigher()

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

func (p *FrameworkProvider) ConfigureConfigDefaults(config *providerSchema) {
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

	if config.HttpClientRetryEnabled.IsNull() {
		retryEnabled, err := utils.GetMultiEnvVar("DD_HTTP_CLIENT_RETRY_ENABLED")
		if err == nil {
			config.HttpClientRetryEnabled = types.StringValue(retryEnabled)
		}
	}

	if config.HttpClientRetryTimeout.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar("DD_HTTP_CLIENT_RETRY_TIMEOUT")
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryTimeout = types.Int64Value(int64(v))
		}
	}

	// Configure defaults for booleans.
	// Remove this once fully migrated to framework
	if config.Validate.IsNull() {
		config.Validate = types.StringValue("true")
	}
	if config.HttpClientRetryEnabled.IsNull() {
		config.HttpClientRetryEnabled = types.StringValue("true")
	}
}

func (p *FrameworkProvider) ConfigValidators(ctx context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		frameworkvalidators.NewValidateProviderStringValIn("validate", "true", "false"),
		frameworkvalidators.NewValidateProviderStringValIn("http_client_retry_enabled", "true", "false"),
	}
}

func (p *FrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIKeyResource,
	}
}

func (p *FrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIPRangesDataSource,
	}
}
