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
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datadogCommunity "github.com/zorkian/go-datadog-api"
	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/transport"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func buildFrameworkDatadogClient(httpClient *http.Client) *common.APIClient {
	//Datadog API config.HTTPClient
	config := common.NewConfiguration()
	config.Debug = isDebug()
	config.HTTPClient = httpClient
	return common.NewAPIClient(config)
}

func testAccFrameworkPreCheck(t *testing.T) {
	// This logic was previously done by the PreCheck func (e.g PreCheck: func() { testAccPreCheck(t) })
	// Since we no longer configure the providers with a callback function, this
	// step must occur prior to test running.
	for _, v := range append(datadog.APPKeyEnvVars, datadog.APIKeyEnvVars...) {
		_ = os.Unsetenv(v)
	}

	if isReplaying() {
		return
	}

	if !isAPIKeySet() {
		t.Fatalf("%s must be set for acceptance tests", testAPIKeyEnvName)
	}
	if !isAPPKeySet() {
		t.Fatalf("%s must be set for acceptance tests", testAPPKeyEnvName)
	}

	if !isTestOrg() {
		t.Fatalf(
			"The keys you've set potentially belong to a production environment. "+
				"Tests do all sorts of create/update/delete calls to the organisation, so only run them against a sandbox environment. "+
				"If you know what you are doing, set the `%s` environment variable to the public ID of your organization. "+
				"See https://docs.datadoghq.com/api/latest/organizations/#list-your-managed-organizations to get it.",
			testOrgEnvName,
		)
	}

	if err := os.Setenv(datadog.DDAPIKeyEnvName, os.Getenv(testAPIKeyEnvName)); err != nil {
		t.Fatalf("Error setting API key: %v", err)
	}

	if err := os.Setenv(datadog.DDAPPKeyEnvName, os.Getenv(testAPPKeyEnvName)); err != nil {
		t.Fatalf("Error setting API key: %v", err)
	}
}

func initAccTestApiClients(ctx context.Context, t *testing.T, httpClient *http.Client) (context.Context, *utils.ApiInstances, *datadogCommunity.Client) {
	testAccFrameworkPreCheck(t)

	apiKey, _ := utils.GetMultiEnvVar(datadog.APIKeyEnvVars[:]...)
	appKey, _ := utils.GetMultiEnvVar(datadog.APPKeyEnvVars[:]...)
	apiURL, _ := utils.GetMultiEnvVar(datadog.APIUrlEnvVars[:]...)

	communityClient := datadogCommunity.NewClient(apiKey, appKey)
	if apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}
	c := ddhttp.WrapClient(httpClient)
	communityClient.HttpClient = c

	ctx, _ = buildContext(ctx, apiKey, appKey, apiURL)
	apiInstances := &utils.ApiInstances{HttpClient: buildFrameworkDatadogClient(httpClient)}

	return ctx, apiInstances, communityClient
}

func testAccFrameworkMuxProvidersServer(ctx context.Context, sdkV2Provider *schema.Provider, frameworkProvider *datadog.FrameworkProvider) map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		"datadog": func() (tfprotov5.ProviderServer, error) {
			muxServer, err := tf5muxserver.NewMuxServer(ctx, providerserver.NewProtocol5(frameworkProvider), sdkV2Provider.GRPCProvider)
			return muxServer, err
		},
	}
}

func testAccFrameworkMuxProviders(ctx context.Context, t *testing.T) (context.Context, *schema.Provider, *datadog.FrameworkProvider, map[string]func() (tfprotov5.ProviderServer, error)) {
	ctx, httpClient := initHttpClient(ctx, t)
	ctx, apiInstances, communityClient := initAccTestApiClients(ctx, t, httpClient)
	tClock := testClock(t)

	// Init sdkV2 provider
	sdkV2Provider := datadog.Provider()
	sdkV2Provider.ConfigureContextFunc = testProviderConfigure(ctx, httpClient, tClock)
	sdkV2Provider.ConfigureContextFunc = func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &datadog.ProviderConfiguration{
			Auth:                ctx,
			CommunityClient:     communityClient,
			DatadogApiInstances: apiInstances,

			Now: tClock.Now,
		}, nil
	}

	// Init framework provider
	frameworkProvider := &datadog.FrameworkProvider{
		Auth:                ctx,
		CommunityClient:     communityClient,
		DatadogApiInstances: apiInstances,

		Now: tClock.Now,
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

	return ctx, sdkV2Provider, frameworkProvider, muxServer
}

func initHttpClient(ctx context.Context, t *testing.T) (context.Context, *http.Client) {
	ctx = testSpan(ctx, t)
	rec := initRecorder(t)
	ctx = context.WithValue(ctx, clockContextKey("clock"), testClock(t))
	httpClient := cleanhttp.DefaultClient()
	loggingTransport := logging.NewTransport("Datadog", rec)
	httpClient.Transport = transport.NewCustomTransport(loggingTransport, transport.CustomTransportOptions{})

	return ctx, httpClient
}
