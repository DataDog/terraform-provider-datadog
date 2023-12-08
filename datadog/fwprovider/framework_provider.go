package fwprovider

import (
	"context"
	"fmt"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	datadogCommunity "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ provider.Provider = &FrameworkProvider{}
)

var Resources = []func() resource.Resource{
	NewAPIKeyResource,
	NewDashboardListResource,
	NewApmRetentionFilterResource,
	NewApmRetentionFiltersOrderResource,
	NewDowntimeScheduleResource,
	NewIntegrationAzureResource,
	NewIntegrationAwsEventBridgeResource,
	NewIntegrationCloudflareAccountResource,
	NewIntegrationConfluentAccountResource,
	NewIntegrationConfluentResourceResource,
	NewIntegrationFastlyAccountResource,
	NewIntegrationFastlyServiceResource,
	NewIntegrationGcpStsResource,
	NewRestrictionPolicyResource,
	NewSensitiveDataScannerGroupOrder,
	NewServiceAccountApplicationKeyResource,
	NewSpansMetricResource,
	NewSyntheticsConcurrencyCapResource,
	NewTeamLinkResource,
	NewTeamMembershipResource,
	NewTeamPermissionSettingResource,
	NewTeamResource,
}

var Datasources = []func() datasource.DataSource{
	NewAPIKeyDataSource,
	NewDatadogDashboardListDataSource,
	NewDatadogApmRetentionFiltersOrderDataSource,
	NewDatadogIntegrationAWSNamespaceRulesDatasource,
	NewDatadogServiceAccountDatasource,
	NewDatadogTeamDataSource,
	NewDatadogTeamMembershipsDataSource,
	NewHostsDataSource,
	NewIPRangesDataSource,
	NewSensitiveDataScannerGroupOrderDatasource,
}

// FrameworkProvider struct
type FrameworkProvider struct {
	CommunityClient     *datadogCommunity.Client
	DatadogApiInstances *utils.ApiInstances
	Auth                context.Context

	ConfigureCallbackFunc func(p *FrameworkProvider, request *provider.ConfigureRequest, config *ProviderSchema) diag.Diagnostics
	Now                   func() time.Time
}

// ProviderSchema struct
type ProviderSchema struct {
	ApiKey                           types.String `tfsdk:"api_key"`
	AppKey                           types.String `tfsdk:"app_key"`
	ApiUrl                           types.String `tfsdk:"api_url"`
	Validate                         types.String `tfsdk:"validate"`
	HttpClientRetryEnabled           types.String `tfsdk:"http_client_retry_enabled"`
	HttpClientRetryTimeout           types.Int64  `tfsdk:"http_client_retry_timeout"`
	HttpClientRetryBackoffMultiplier types.Int64  `tfsdk:"http_client_retry_backoff_multiplier"`
	HttpClientRetryBackoffBase       types.Int64  `tfsdk:"http_client_retry_backoff_base"`
	HttpClientRetryMaxRetries        types.Int64  `tfsdk:"http_client_retry_max_retries"`
}

func New() provider.Provider {
	return &FrameworkProvider{
		ConfigureCallbackFunc: defaultConfigureFunc,
	}
}

func (p *FrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	var wrappedResources []func() resource.Resource
	for _, f := range Resources {
		r := f()
		wrappedResources = append(wrappedResources, func() resource.Resource { return NewFrameworkResourceWrapper(&r) })
	}

	return wrappedResources
}

func (p *FrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	var wrappedDatasources []func() datasource.DataSource
	for _, f := range Datasources {
		r := f()
		wrappedDatasources = append(wrappedDatasources, func() datasource.DataSource { return NewFrameworkDatasourceWrapper(&r) })
	}

	return wrappedDatasources
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
				Description: "Enables validation of the provided API key during provider initialization. Valid values are [`true`, `false`]. Default is true. When false, api_key won't be checked.",
			},
			"http_client_retry_enabled": schema.StringAttribute{
				Optional:    true,
				Description: "Enables request retries on HTTP status codes 429 and 5xx. Valid values are [`true`, `false`]. Defaults to `true`.",
			},
			"http_client_retry_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry timeout period. Defaults to 60 seconds.",
			},
			"http_client_retry_backoff_multiplier": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry back off multiplier. Defaults to 2.",
			},
			"http_client_retry_backoff_base": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request retry back off base. Defaults to 2.",
			},
			"http_client_retry_max_retries": schema.Int64Attribute{
				Optional:    true,
				Description: "The HTTP request maximum retry number. Defaults to 3.",
			},
		},
	}
}

func (p *FrameworkProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config ProviderSchema
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	diags := p.ConfigureConfigDefaults(ctx, &config)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
	}

	response.Diagnostics.Append(p.ConfigureCallbackFunc(p, &request, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Make config available for data sources and resources
	response.DataSourceData = p
	response.ResourceData = p
}

func (p *FrameworkProvider) ConfigureConfigDefaults(ctx context.Context, config *ProviderSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	if config.ApiKey.IsNull() {
		apiKey, err := utils.GetMultiEnvVar(utils.APIKeyEnvVars[:]...)
		if err == nil {
			config.ApiKey = types.StringValue(apiKey)
		}
	}

	if config.AppKey.IsNull() {
		appKey, err := utils.GetMultiEnvVar(utils.APPKeyEnvVars[:]...)
		if err == nil {
			config.AppKey = types.StringValue(appKey)
		}
	}

	if config.ApiUrl.IsNull() {
		apiUrl, err := utils.GetMultiEnvVar(utils.APIUrlEnvVars[:]...)
		if err == nil {
			config.ApiUrl = types.StringValue(apiUrl)
		}
	}

	if config.HttpClientRetryEnabled.IsNull() {
		retryEnabled, err := utils.GetMultiEnvVar(utils.DDHTTPRetryEnabled)
		if err == nil {
			config.HttpClientRetryEnabled = types.StringValue(retryEnabled)
		}
	}

	if config.HttpClientRetryTimeout.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryTimeout)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryTimeout = types.Int64Value(int64(v))
		}
	}

	if config.HttpClientRetryBackoffMultiplier.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryBackoffMultiplier)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryBackoffMultiplier = types.Int64Value(int64(v))
		}
	}

	if config.HttpClientRetryBackoffBase.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryBackoffBase)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryBackoffBase = types.Int64Value(int64(v))
		}
	}

	if config.HttpClientRetryMaxRetries.IsNull() {
		rTimeout, err := utils.GetMultiEnvVar(utils.DDHTTPRetryMaxRetries)
		if err == nil {
			v, _ := strconv.Atoi(rTimeout)
			config.HttpClientRetryMaxRetries = types.Int64Value(int64(v))
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

	// Run validations on the provider config after defaults and values from
	// env var has been set.
	diags.Append(p.ValidateConfigValues(ctx, config)...)

	return diags
}

func (p *FrameworkProvider) ValidateConfigValues(ctx context.Context, config *ProviderSchema) diag.Diagnostics {
	var diags diag.Diagnostics
	// Init validators we need for purposes of config validation only
	oneOfStringValidator := stringvalidator.OneOf("true", "false")
	int64AtLeastValidator := int64validator.AtLeast(1)
	int64BetweenValidator := int64validator.Between(1, 5)

	if !config.Validate.IsNull() {
		res := validator.StringResponse{}
		oneOfStringValidator.ValidateString(ctx, validator.StringRequest{ConfigValue: config.Validate}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryEnabled.IsNull() {
		res := validator.StringResponse{}
		oneOfStringValidator.ValidateString(ctx, validator.StringRequest{ConfigValue: config.HttpClientRetryEnabled}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryBackoffMultiplier.IsNull() {
		res := validator.Int64Response{}
		int64AtLeastValidator.ValidateInt64(ctx, validator.Int64Request{ConfigValue: config.HttpClientRetryBackoffMultiplier}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryBackoffBase.IsNull() {
		res := validator.Int64Response{}
		int64AtLeastValidator.ValidateInt64(ctx, validator.Int64Request{ConfigValue: config.HttpClientRetryBackoffBase}, &res)
		diags.Append(res.Diagnostics...)
	}

	if !config.HttpClientRetryMaxRetries.IsNull() {
		res := validator.Int64Response{}
		int64BetweenValidator.ValidateInt64(ctx, validator.Int64Request{ConfigValue: config.HttpClientRetryMaxRetries}, &res)
		diags.Append(res.Diagnostics...)
	}

	return diags
}

// Helper method to configure the provider
func defaultConfigureFunc(p *FrameworkProvider, request *provider.ConfigureRequest, config *ProviderSchema) diag.Diagnostics {
	diags := diag.Diagnostics{}
	validate, _ := strconv.ParseBool(config.Validate.ValueString())
	httpClientRetryEnabled, _ := strconv.ParseBool(config.Validate.ValueString())

	if validate && (config.ApiKey.ValueString() == "" || config.AppKey.ValueString() == "") {
		diags.AddError("api_key and app_key must be set unless validate = false", "")
		return diags
	}

	// Initialize the community client
	p.CommunityClient = datadogCommunity.NewClient(config.ApiKey.ValueString(), config.AppKey.ValueString())
	if !config.ApiUrl.IsNull() && config.ApiUrl.ValueString() != "" {
		p.CommunityClient.SetBaseUrl(config.ApiUrl.ValueString())
	}
	c := cleanhttp.DefaultClient()
	p.CommunityClient.ExtraHeader["User-Agent"] = utils.GetUserAgentFramework(fmt.Sprintf(
		"datadog-api-client-go/%s (go %s; os %s; arch %s)",
		"go-datadog-api",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	), request.TerraformVersion)
	p.CommunityClient.HttpClient = c

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
	ddClientConfig.UserAgent = utils.GetUserAgentFramework(ddClientConfig.UserAgent, request.TerraformVersion)
	ddClientConfig.Debug = logging.IsDebugOrHigher()

	if !config.ApiUrl.IsNull() && config.ApiUrl.ValueString() != "" {
		parsedAPIURL, parseErr := url.Parse(config.ApiUrl.ValueString())
		if parseErr != nil {
			diags.AddError("invalid API URL", parseErr.Error())
			return diags
		}
		if parsedAPIURL.Host == "" || parsedAPIURL.Scheme == "" {
			diags.AddError("invalid API URL", fmt.Sprintf("API URL '%s' missing protocol or host", config.ApiUrl.ValueString()))
			return diags
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
		ipRangesDNSNameArr = append([]string{utils.BaseIPRangesSubdomain}, ipRangesDNSNameArr...)

		auth = context.WithValue(auth, datadog.ContextOperationServerIndices, map[string]int{
			"v1.IPRangesApi.GetIPRanges": 1,
		})
		auth = context.WithValue(auth, datadog.ContextOperationServerVariables, map[string]map[string]string{
			"v1.IPRangesApi.GetIPRanges": {
				"name": strings.Join(ipRangesDNSNameArr, "."),
			},
		})
	}

	if httpClientRetryEnabled {
		ddClientConfig.RetryConfiguration.EnableRetry = httpClientRetryEnabled

		if !config.HttpClientRetryBackoffMultiplier.IsNull() {
			timeout := time.Duration(config.HttpClientRetryBackoffMultiplier.ValueInt64()) * time.Second
			ddClientConfig.RetryConfiguration.HTTPRetryTimeout = timeout
		}

		if !config.HttpClientRetryBackoffBase.IsNull() {
			ddClientConfig.RetryConfiguration.BackOffBase = float64(config.HttpClientRetryBackoffMultiplier.ValueInt64())
		}

		if !config.HttpClientRetryMaxRetries.IsNull() {
			ddClientConfig.RetryConfiguration.MaxRetries = int(config.HttpClientRetryMaxRetries.ValueInt64())
		}
	}

	datadogClient := datadog.NewAPIClient(ddClientConfig)

	p.DatadogApiInstances = &utils.ApiInstances{HttpClient: datadogClient}
	p.Auth = auth

	/*  Commented out due to duplicate validation in SDK provider - remove after Framework migration is complete.
	if validate {
		log.Println("[INFO] Datadog client successfully initialized, now validating...")
		resp, _, err := p.DatadogApiInstances.GetAuthenticationApiV1().Validate(auth)
		if err != nil {
			diags.AddError("[ERROR] Datadog Client validation error", err.Error())
			return diags
		}
		valid, ok := resp.GetValidOk()
		if (ok && !*valid) || !ok {
			err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`)
			diags.AddError("[ERROR] Datadog Client validation error", err.Error())
			return diags
		}
	} else {
		log.Println("[INFO] Skipping key validation (validate = false)")
	}
	log.Printf("[INFO] Datadog Client successfully validated.")
	*/
	return nil
}

var (
	_ resource.ResourceWithConfigure        = &FrameworkResourceWrapper{}
	_ resource.ResourceWithImportState      = &FrameworkResourceWrapper{}
	_ resource.ResourceWithConfigValidators = &FrameworkResourceWrapper{}
	_ resource.ResourceWithModifyPlan       = &FrameworkResourceWrapper{}
	_ resource.ResourceWithUpgradeState     = &FrameworkResourceWrapper{}
	_ resource.ResourceWithValidateConfig   = &FrameworkResourceWrapper{}
)

func NewFrameworkResourceWrapper(i *resource.Resource) resource.Resource {
	return &FrameworkResourceWrapper{
		innerResource: i,
	}
}

type FrameworkResourceWrapper struct {
	innerResource *resource.Resource
}

func (r *FrameworkResourceWrapper) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		_, ok := req.ProviderData.(*FrameworkProvider)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Resource Configure Type", "")
			return
		}

		rCasted.Configure(ctx, req, resp)
	}
}

func (r *FrameworkResourceWrapper) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	(*r.innerResource).Metadata(ctx, req, resp)
	resp.TypeName = req.ProviderTypeName + resp.TypeName
}

func (r *FrameworkResourceWrapper) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	(*r.innerResource).Schema(ctx, req, resp)
	fwutils.EnrichFrameworkResourceSchema(&resp.Schema)
}

func (r *FrameworkResourceWrapper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	(*r.innerResource).Create(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	(*r.innerResource).Read(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	(*r.innerResource).Update(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	(*r.innerResource).Delete(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithImportState); ok {
		rCasted.ImportState(ctx, req, resp)
		return
	}

	resp.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (r *FrameworkResourceWrapper) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

func (r *FrameworkResourceWrapper) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithModifyPlan); ok {
		v.ModifyPlan(ctx, req, resp)
	}
}

func (r *FrameworkResourceWrapper) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := (*r.innerResource).(resource.ResourceWithUpgradeState); ok {
		return v.UpgradeState(ctx)
	}
	return nil
}

func (r *FrameworkResourceWrapper) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithValidateConfig); ok {
		v.ValidateConfig(ctx, req, resp)
	}
}

var (
	_ datasource.DataSourceWithConfigure        = &FrameworkDatasourceWrapper{}
	_ datasource.DataSourceWithConfigValidators = &FrameworkDatasourceWrapper{}
	_ datasource.DataSourceWithValidateConfig   = &FrameworkDatasourceWrapper{}
	_ datasource.DataSource                     = &FrameworkDatasourceWrapper{}
)

func NewFrameworkDatasourceWrapper(i *datasource.DataSource) datasource.DataSource {
	return &FrameworkDatasourceWrapper{
		innerDatasource: i,
	}
}

type FrameworkDatasourceWrapper struct {
	innerDatasource *datasource.DataSource
}

func (r *FrameworkDatasourceWrapper) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	rCasted, ok := (*r.innerDatasource).(datasource.DataSourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		_, ok := req.ProviderData.(*FrameworkProvider)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "")
			return
		}

		rCasted.Configure(ctx, req, resp)
	}
}

func (r *FrameworkDatasourceWrapper) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	if rCasted, ok := (*r.innerDatasource).(datasource.DataSourceWithValidateConfig); ok {
		rCasted.ValidateConfig(ctx, req, resp)
	}
}

func (r *FrameworkDatasourceWrapper) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	if rCasted, ok := (*r.innerDatasource).(datasource.DataSourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

func (r *FrameworkDatasourceWrapper) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	(*r.innerDatasource).Metadata(ctx, req, resp)
	resp.TypeName = req.ProviderTypeName + resp.TypeName
}

func (r *FrameworkDatasourceWrapper) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	(*r.innerDatasource).Schema(ctx, req, resp)
}

func (r *FrameworkDatasourceWrapper) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	(*r.innerDatasource).Read(ctx, req, resp)
}
