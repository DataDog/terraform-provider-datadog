package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datadogCommunity "github.com/zorkian/go-datadog-api"
	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type compositeProviderStruct struct {
	sdkV2Provider     *schema.Provider
	frameworkProvider *fwprovider.FrameworkProvider
}

func buildFrameworkDatadogClient(ctx context.Context, httpClient *http.Client) *common.APIClient {
	//Datadog API config.HTTPClient
	config := common.NewConfiguration()

	config.SetUnstableOperationEnabled("v2.CreateOpenAPI", true)
	config.SetUnstableOperationEnabled("v2.UpdateOpenAPI", true)
	config.SetUnstableOperationEnabled("v2.GetOpenAPI", true)
	config.SetUnstableOperationEnabled("v2.DeleteOpenAPI", true)
	config.SetUnstableOperationEnabled("v2.ListAWSLogsServices", true)
	config.SetUnstableOperationEnabled("v2.ListAWSNamespaces", true)
	config.SetUnstableOperationEnabled("v2.CreateAWSAccount", true)
	config.SetUnstableOperationEnabled("v2.UpdateAWSAccount", true)
	config.SetUnstableOperationEnabled("v2.DeleteAWSAccount", true)
	config.SetUnstableOperationEnabled("v2.GetAWSAccount", true)
	config.SetUnstableOperationEnabled("v2.CreateNewAWSExternalID", true)
	config.SetUnstableOperationEnabled("v2.GetDataset", true)
	config.SetUnstableOperationEnabled("v2.CreateDataset", true)
	config.SetUnstableOperationEnabled("v2.UpdateDataset", true)
	config.SetUnstableOperationEnabled("v2.DeleteDataset", true)

	// Enable Observability Pipelines
	config.SetUnstableOperationEnabled("v2.CreatePipeline", true)
	config.SetUnstableOperationEnabled("v2.GetPipeline", true)
	config.SetUnstableOperationEnabled("v2.UpdatePipeline", true)
	config.SetUnstableOperationEnabled("v2.DeletePipeline", true)

	// Enable MonitorNotificationRule
	config.SetUnstableOperationEnabled("v2.CreateMonitorNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.GetMonitorNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.DeleteMonitorNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.UpdateMonitorNotificationRule", true)

	// Enable IncidentType
	config.SetUnstableOperationEnabled("v2.CreateIncidentType", true)
	config.SetUnstableOperationEnabled("v2.GetIncidentType", true)
	config.SetUnstableOperationEnabled("v2.UpdateIncidentType", true)
	config.SetUnstableOperationEnabled("v2.DeleteIncidentType", true)

	// Enable IncidentNotificationTemplate
	config.SetUnstableOperationEnabled("v2.CreateIncidentNotificationTemplate", true)
	config.SetUnstableOperationEnabled("v2.GetIncidentNotificationTemplate", true)
	config.SetUnstableOperationEnabled("v2.UpdateIncidentNotificationTemplate", true)
	config.SetUnstableOperationEnabled("v2.DeleteIncidentNotificationTemplate", true)
	config.SetUnstableOperationEnabled("v2.ListIncidentNotificationTemplates", true)

	// Enable IncidentNotificationRule
	config.SetUnstableOperationEnabled("v2.CreateIncidentNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.GetIncidentNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.UpdateIncidentNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.DeleteIncidentNotificationRule", true)
	config.SetUnstableOperationEnabled("v2.ListIncidentNotificationRules", true)

	if ctx.Value("http_retry_enable") == true {
		config.RetryConfiguration.EnableRetry = true
	}
	config.Debug = isDebug()
	config.HTTPClient = httpClient
	return common.NewAPIClient(config)
}

func initAccTestApiClients(ctx context.Context, t *testing.T, httpClient *http.Client) (context.Context, *utils.ApiInstances, *datadogCommunity.Client) {
	// This logic was previously done by the PreCheck func (e.g PreCheck: func() { testAccPreCheck(t) })
	// Since we no longer configure the providers with a callback function, this
	// step must occur prior to test running.
	testAccPreCheck(t)

	apiKey, _ := utils.GetMultiEnvVar(utils.APIKeyEnvVars[:]...)
	appKey, _ := utils.GetMultiEnvVar(utils.APPKeyEnvVars[:]...)
	apiURL, _ := utils.GetMultiEnvVar(utils.APIUrlEnvVars[:]...)

	communityClient := datadogCommunity.NewClient(apiKey, appKey)
	if apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}
	c := ddhttp.WrapClient(httpClient)
	communityClient.HttpClient = c

	ctx, _ = buildContext(ctx, apiKey, appKey, apiURL)
	apiInstances := &utils.ApiInstances{HttpClient: buildFrameworkDatadogClient(ctx, httpClient)}

	return ctx, apiInstances, communityClient
}

func testAccFrameworkMuxProvidersServer(ctx context.Context, sdkV2Provider *schema.Provider, frameworkProvider *fwprovider.FrameworkProvider) map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		"datadog": func() (tfprotov5.ProviderServer, error) {
			muxServer, err := tf5muxserver.NewMuxServer(ctx, providerserver.NewProtocol5(frameworkProvider), sdkV2Provider.GRPCProvider)
			return muxServer, err
		},
	}
}

func testAccFrameworkMuxProviders(ctx context.Context, t *testing.T) (context.Context, *compositeProviderStruct, map[string]func() (tfprotov5.ProviderServer, error)) {
	ctx, httpClient := initHttpClient(ctx, t)
	ctx, apiInstances, communityClient := initAccTestApiClients(ctx, t, httpClient)

	// Init sdkV2 provider
	sdkV2Provider := datadog.Provider()
	sdkV2Provider.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &datadog.ProviderConfiguration{
			Auth:                ctx,
			CommunityClient:     communityClient,
			DatadogApiInstances: apiInstances,

			Now: clockFromContext(ctx).Now,
		}, nil
	}

	// Init framework provider
	frameworkProvider := &fwprovider.FrameworkProvider{
		Auth:                ctx,
		CommunityClient:     communityClient,
		DatadogApiInstances: apiInstances,

		Now: clockFromContext(ctx).Now,
		ConfigureCallbackFunc: func(p *fwprovider.FrameworkProvider, request *provider.ConfigureRequest, config *fwprovider.ProviderSchema) frameworkDiag.Diagnostics {
			return nil
		},
	}

	// The provider must be initialized prior to setting User-Agent headers
	// Hence we add the headers here after initialization.
	communityClient.ExtraHeader["User-Agent"] = utils.GetUserAgent(fmt.Sprintf(
		"datadog-api-client-go/%s (go %s; os %s; arch %s)",
		"go-datadog-api",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	))
	apiInstances.HttpClient.Cfg.UserAgent = utils.GetUserAgent(apiInstances.HttpClient.Cfg.UserAgent)

	// Init mux servers
	muxServer := testAccFrameworkMuxProvidersServer(ctx, sdkV2Provider, frameworkProvider)

	providers := &compositeProviderStruct{
		sdkV2Provider:     sdkV2Provider,
		frameworkProvider: frameworkProvider,
	}

	return ctx, providers, muxServer
}

func initHttpClient(ctx context.Context, t *testing.T) (context.Context, *http.Client) {
	ctx = context.WithValue(ctx, clockContextKey("clock"), testClock(t))
	ctx = testSpan(ctx, t)
	rec := initRecorder(t)
	httpClient := cleanhttp.DefaultClient()
	httpClient.Transport = rec
	t.Cleanup(func() {
		rec.Stop()
	})

	return ctx, httpClient
}

func withDefaultTagsFw(ctx context.Context, providers *compositeProviderStruct, defaultTags map[string]string) func() (tfprotov5.ProviderServer, error) {
	return func() (tfprotov5.ProviderServer, error) {
		providers.frameworkProvider.DefaultTags = defaultTags
		muxServer, err := tf5muxserver.NewMuxServer(ctx,
			providerserver.NewProtocol5(providers.frameworkProvider), providers.sdkV2Provider.GRPCProvider)
		return muxServer, err
	}
}

// TestFrameworkProviderConfigure_CloudAuthOnly tests that cloud auth works when only cloud_provider_type is set
func TestFrameworkProviderConfigure_CloudAuthOnly(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DD_API_KEY")
	os.Unsetenv("DD_APP_KEY")
	os.Unsetenv("DATADOG_API_KEY")
	os.Unsetenv("DATADOG_APP_KEY")

	p := fwprovider.New().(*fwprovider.FrameworkProvider)
	config := &fwprovider.ProviderSchema{
		OrgUuid:                types.StringValue("test-org-uuid"),
		CloudProviderType:      types.StringValue("aws"),
		CloudProviderRegion:    types.StringValue("us-east-1"),
		ApiUrl:                 types.StringValue("https://api.datad0g.com"),
		Validate:               types.StringValue("false"),
		HttpClientRetryEnabled: types.StringValue("false"),
	}

	request := &provider.ConfigureRequest{}
	diags := p.ConfigureCallbackFunc(p, request, config)

	// Should not have errors
	if diags.HasError() {
		t.Errorf("framework provider configure should not error with cloud auth only, got: %v", diags)
	}
	if p.DatadogApiInstances == nil {
		t.Fatal("DatadogApiInstances should be set")
	}

	// Verify DelegatedTokenConfig is set (cloud auth is enabled)
	if p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig == nil {
		t.Error("DelegatedTokenConfig should be set when using cloud auth only")
	}
}

// TestFrameworkProviderConfigure_APIKeyOnly tests that API key auth works when only api_key/app_key are set
func TestFrameworkProviderConfigure_APIKeyOnly(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DD_API_KEY")
	os.Unsetenv("DD_APP_KEY")
	os.Unsetenv("DATADOG_API_KEY")
	os.Unsetenv("DATADOG_APP_KEY")

	p := fwprovider.New().(*fwprovider.FrameworkProvider)
	config := &fwprovider.ProviderSchema{
		ApiKey:                 types.StringValue("test_api_key"),
		AppKey:                 types.StringValue("test_app_key"),
		ApiUrl:                 types.StringValue("https://api.datad0g.com"),
		Validate:               types.StringValue("false"),
		HttpClientRetryEnabled: types.StringValue("false"),
	}

	request := &provider.ConfigureRequest{}
	diags := p.ConfigureCallbackFunc(p, request, config)

	// Should not have errors
	if diags.HasError() {
		t.Errorf("framework provider configure should not error with API key auth only, got: %v", diags)
	}
	if p.DatadogApiInstances == nil {
		t.Fatal("DatadogApiInstances should be set")
	}

	// Verify DelegatedTokenConfig is NOT set (API key auth is used)
	if p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig != nil {
		t.Errorf("DelegatedTokenConfig should NOT be set when using API key auth, got: %+v",
			p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig)
	}
}

// TestFrameworkProviderConfigure_CloudAuthWithAPIKey tests the bug scenario:
// When both cloud_provider_type and api_key are set, API key auth should take precedence
func TestFrameworkProviderConfigure_CloudAuthWithAPIKey(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DD_API_KEY")
	os.Unsetenv("DD_APP_KEY")
	os.Unsetenv("DATADOG_API_KEY")
	os.Unsetenv("DATADOG_APP_KEY")

	p := fwprovider.New().(*fwprovider.FrameworkProvider)
	config := &fwprovider.ProviderSchema{
		ApiKey:                 types.StringValue("test_api_key"),
		AppKey:                 types.StringValue("test_app_key"),
		OrgUuid:                types.StringValue("test-org-uuid"),
		CloudProviderType:      types.StringValue("aws"),
		CloudProviderRegion:    types.StringValue("us-east-1"),
		ApiUrl:                 types.StringValue("https://api.datad0g.com"),
		Validate:               types.StringValue("false"),
		HttpClientRetryEnabled: types.StringValue("false"),
	}

	request := &provider.ConfigureRequest{}
	diags := p.ConfigureCallbackFunc(p, request, config)

	// Should not have errors (this was the bug - it would cause "DelegatedTokenCredentials not found in context")
	if diags.HasError() {
		t.Errorf("framework provider configure should not error when both cloud auth and API keys are set, got: %v", diags)
	}
	if p.DatadogApiInstances == nil {
		t.Fatal("DatadogApiInstances should be set")
	}

	// Verify DelegatedTokenConfig is NOT set (API key auth takes precedence)
	if p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig != nil {
		t.Errorf("DelegatedTokenConfig should NOT be set when API keys are present (API key auth takes precedence), got: %+v",
			p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig)
	}
}

// TestFrameworkProviderConfigure_CloudAuthWithEnvVarAPIKey tests the bug scenario with environment variables:
// When both cloud_provider_type and DD_API_KEY/DD_APP_KEY env vars are set, API key auth should take precedence
func TestFrameworkProviderConfigure_CloudAuthWithEnvVarAPIKey(t *testing.T) {
	// Set environment variables
	os.Setenv("DD_API_KEY", "test_api_key_from_env")
	os.Setenv("DD_APP_KEY", "test_app_key_from_env")
	defer func() {
		os.Unsetenv("DD_API_KEY")
		os.Unsetenv("DD_APP_KEY")
	}()

	p := fwprovider.New().(*fwprovider.FrameworkProvider)
	config := &fwprovider.ProviderSchema{
		OrgUuid:                types.StringValue("test-org-uuid"),
		CloudProviderType:      types.StringValue("aws"),
		CloudProviderRegion:    types.StringValue("us-east-1"),
		ApiUrl:                 types.StringValue("https://api.datad0g.com"),
		Validate:               types.StringValue("false"),
		HttpClientRetryEnabled: types.StringValue("false"),
	}

	// Populate config from environment variables
	ctx := context.Background()
	diags := p.ConfigureConfigDefaults(ctx, config)
	if diags.HasError() {
		t.Fatalf("ConfigureConfigDefaults failed: %v", diags)
	}

	request := &provider.ConfigureRequest{}
	diags = p.ConfigureCallbackFunc(p, request, config)

	// Should not have errors
	if diags.HasError() {
		t.Errorf("framework provider configure should not error when cloud auth is set but env vars have API keys, got: %v", diags)
	}
	if p.DatadogApiInstances == nil {
		t.Fatal("DatadogApiInstances should be set")
	}

	// Verify DelegatedTokenConfig is NOT set (API key from env takes precedence)
	if p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig != nil {
		t.Errorf("DelegatedTokenConfig should NOT be set when API keys from env vars are present, got: %+v",
			p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig)
	}
}

// TestFrameworkProviderConfigure_CloudAuthWithOnlyAppKey tests that even with only app_key set (no api_key),
// API key auth still takes precedence over cloud auth
func TestFrameworkProviderConfigure_CloudAuthWithOnlyAppKey(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DD_API_KEY")
	os.Unsetenv("DD_APP_KEY")
	os.Unsetenv("DATADOG_API_KEY")
	os.Unsetenv("DATADOG_APP_KEY")

	p := fwprovider.New().(*fwprovider.FrameworkProvider)
	config := &fwprovider.ProviderSchema{
		AppKey:                 types.StringValue("test_app_key"),
		OrgUuid:                types.StringValue("test-org-uuid"),
		CloudProviderType:      types.StringValue("aws"),
		CloudProviderRegion:    types.StringValue("us-east-1"),
		ApiUrl:                 types.StringValue("https://api.datad0g.com"),
		Validate:               types.StringValue("false"),
		HttpClientRetryEnabled: types.StringValue("false"),
	}

	request := &provider.ConfigureRequest{}
	diags := p.ConfigureCallbackFunc(p, request, config)

	// Should not have errors
	if diags.HasError() {
		t.Errorf("framework provider configure should not error when both cloud auth and app_key are set, got: %v", diags)
	}
	if p.DatadogApiInstances == nil {
		t.Fatal("DatadogApiInstances should be set")
	}

	// Verify DelegatedTokenConfig is NOT set (app_key alone triggers API key auth)
	if p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig != nil {
		t.Errorf("DelegatedTokenConfig should NOT be set when app_key is present (even without api_key), got: %+v",
			p.DatadogApiInstances.HttpClient.GetConfig().DelegatedTokenConfig)
	}
}
