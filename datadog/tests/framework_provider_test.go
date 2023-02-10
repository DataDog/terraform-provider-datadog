package test

import (
	"context"
	"net/http"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
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

func initAccFrameworkProvider(ctx context.Context, t *testing.T, httpClient *http.Client) provider.Provider {
	apiKey, _ := utils.GetMultiEnvVar(datadog.APIKeyEnvVars[:]...)
	appKey, _ := utils.GetMultiEnvVar(datadog.APPKeyEnvVars[:]...)
	apiURL, _ := utils.GetMultiEnvVar(datadog.APIUrlEnvVars[:]...)

	communityClient := datadogCommunity.NewClient(apiKey, appKey)
	if apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}
	c := ddhttp.WrapClient(httpClient)
	communityClient.HttpClient = c

	authCtx, _ := buildContext(ctx, apiKey, appKey, apiURL)

	p := datadog.FrameworkProvider{
		Auth:                authCtx,
		CommunityClient:     communityClient,
		DatadogApiInstances: &utils.ApiInstances{HttpClient: buildFrameworkDatadogClient(c)},
	}
	return &p
}

func testAccFrameworkProvidersWithHTTPClient(ctx context.Context, t *testing.T, httpClient *http.Client) map[string]func() (tfprotov5.ProviderServer, error) {
	p := initAccFrameworkProvider(ctx, t, httpClient)
	return map[string]func() (tfprotov5.ProviderServer, error){
		"datadog": providerserver.NewProtocol5WithError(p),
	}

}

func testAccFrameworkProviders(ctx context.Context, t *testing.T) (context.Context, map[string]func() (tfprotov5.ProviderServer, error)) {
	ctx = testSpan(ctx, t)
	rec := initRecorder(t)
	ctx = context.WithValue(ctx, clockContextKey("clock"), testClock(t))
	c := cleanhttp.DefaultClient()
	loggingTransport := logging.NewTransport("Datadog", rec)
	c.Transport = transport.NewCustomTransport(loggingTransport, transport.CustomTransportOptions{})
	p := testAccFrameworkProvidersWithHTTPClient(ctx, t, c)
	t.Cleanup(func() {
		rec.Stop()
	})

	return ctx, p
}
