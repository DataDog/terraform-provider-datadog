package test

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
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
	httpClient.Transport = logging.NewTransport("Datadog", rec)
	t.Cleanup(func() {
		rec.Stop()
	})

	return ctx, httpClient
}
