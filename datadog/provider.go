package datadog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"runtime"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/meta"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/version"
	datadogCommunity "github.com/zorkian/go-datadog-api"
)

var datadogProvider *schema.Provider

func Provider() terraform.ResourceProvider {
	datadogProvider = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DATADOG_API_KEY", nil),
			},
			"app_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("DATADOG_APP_KEY", nil),
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DATADOG_HOST", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"datadog_dashboard":                            resourceDatadogDashboard(),
			"datadog_dashboard_list":                       resourceDatadogDashboardList(),
			"datadog_downtime":                             resourceDatadogDowntime(),
			"datadog_integration_gcp":                      resourceDatadogIntegrationGcp(),
			"datadog_integration_aws":                      resourceDatadogIntegrationAws(),
			"datadog_integration_pagerduty":                resourceDatadogIntegrationPagerduty(),
			"datadog_integration_pagerduty_service_object": resourceDatadogIntegrationPagerdutySO(),
			"datadog_logs_custom_pipeline":                 resourceDatadogLogsCustomPipeline(),
			"datadog_logs_index":                           resourceDatadogLogsIndex(),
			"datadog_logs_index_order":                     resourceDatadogLogsIndexOrder(),
			"datadog_logs_integration_pipeline":            resourceDatadogLogsIntegrationPipeline(),
			"datadog_logs_pipeline_order":                  resourceDatadogLogsPipelineOrder(),
			"datadog_metric_metadata":                      resourceDatadogMetricMetadata(),
			"datadog_monitor":                              resourceDatadogMonitor(),
			"datadog_screenboard":                          resourceDatadogScreenboard(),
			"datadog_service_level_objective":              resourceDatadogServiceLevelObjective(),
			"datadog_synthetics_test":                      resourceDatadogSyntheticsTest(),
			"datadog_timeboard":                            resourceDatadogTimeboard(),
			"datadog_user":                                 resourceDatadogUser(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"datadog_ip_ranges": dataSourceDatadogIpRanges(),
		},

		ConfigureFunc: providerConfigure,
	}

	return datadogProvider
}

//ProviderConfiguration contains the initialized API clients to communicate with the Datadog API
type ProviderConfiguration struct {
	CommunityClient *datadogCommunity.Client
	DatadogClientV1 *datadogV1.APIClient
	DatadogClientV2 *datadogV2.APIClient
	AuthV1          context.Context
	AuthV2          context.Context
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// Initialize the community client
	communityClient := datadogCommunity.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))

	if apiURL := d.Get("api_url").(string); apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", c.Transport)
	communityClient.ExtraHeader["User-Agent"] = fmt.Sprintf("terraform-provider-datadog/%s (go %s; terraform %s; terraform-cli %s)", version.ProviderVersion, runtime.Version(), meta.SDKVersionString(), datadogProvider.TerraformVersion)
	communityClient.HttpClient = c

	log.Println("[INFO] Datadog client successfully initialized, now validating...")
	ok, err := communityClient.Validate()
	if err != nil {
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return nil, err
	} else if !ok {
		err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`)
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return nil, err
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	// Initialize the official Datadog V1 API client
	authV1 := context.WithValue(
		context.Background(),
		datadogV1.ContextAPIKeys,
		map[string]datadogV1.APIKey{
			"apiKeyAuth": {
				Key: d.Get("api_key").(string),
			},
			"appKeyAuth": {
				Key: d.Get("app_key").(string),
			},
		},
	)
	configV1 := datadogV1.NewConfiguration()
	if apiURL := d.Get("api_url").(string); apiURL != "" {
		parsedApiUrl, parseErr := url.Parse(apiURL)
		if parseErr != nil {
			return nil, fmt.Errorf(`invalid API Url : %v`, parseErr)
		}
		if parsedApiUrl.Host == "" || parsedApiUrl.Scheme == "" {
			return nil, fmt.Errorf(`missing protocol or host : %v`, apiURL)
		}
		// If api url is passed, set and use the api name and protocol on ServerIndex{1}
		authV1 = context.WithValue(authV1, datadogV1.ContextServerIndex, 1)
		authV1 = context.WithValue(authV1, datadogV1.ContextServerVariables, map[string]string{
			"name":     parsedApiUrl.Host,
			"protocol": parsedApiUrl.Scheme,
		})
	}
	datadogClientV1 := datadogV1.NewAPIClient(configV1)

	// Initialize the official Datadog V2 API client
	authV2 := context.WithValue(
		context.Background(),
		datadogV2.ContextAPIKeys,
		map[string]datadogV2.APIKey{
			"apiKeyAuth": {
				Key: d.Get("api_key").(string),
			},
			"appKeyAuth": {
				Key: d.Get("app_key").(string),
			},
		},
	)
	configV2 := datadogV2.NewConfiguration()
	if apiURL := d.Get("api_url").(string); apiURL != "" {
		parsedApiUrl, parseErr := url.Parse(apiURL)
		if parseErr != nil {
			return nil, fmt.Errorf(`invalid API Url : %v`, parseErr)
		}
		if parsedApiUrl.Host == "" || parsedApiUrl.Scheme == "" {
			return nil, fmt.Errorf(`missing protocol or host : %v`, apiURL)
		}
		// If api url is passed, set and use the api name and protocol on ServerIndex{1}
		authV2 = context.WithValue(authV2, datadogV2.ContextServerIndex, 1)
		authV2 = context.WithValue(authV2, datadogV2.ContextServerVariables, map[string]string{
			"name":     parsedApiUrl.Host,
			"protocol": parsedApiUrl.Scheme,
		})
	}
	datadogClientV2 := datadogV2.NewAPIClient(configV2)

	return &ProviderConfiguration{
		CommunityClient: communityClient,
		DatadogClientV1: datadogClientV1,
		DatadogClientV2: datadogClientV2,
		AuthV1:          authV1,
		AuthV2:          authV2,
	}, nil
}

func translateClientError(err error, msg string) error {
	if msg == "" {
		msg = "an error occurred"
	}

	if apiErr, ok := err.(datadogV1.GenericOpenAPIError); ok {
		return fmt.Errorf(msg+": %v: %s", err, apiErr.Body())
	}
	if errUrl, ok := err.(*url.Error); ok {
		return fmt.Errorf(msg+" (url.Error): %s", errUrl)
	}

	return fmt.Errorf(msg+": %s", err.Error())
}

func translateClientErrorV2(err error, msg string) error {
	if msg == "" {
		msg = "an error occurred"
	}

	if apiErr, ok := err.(datadogV2.GenericOpenAPIError); ok {
		fmt.Errorf(msg+": %v: %s", err, apiErr.Body())
	}
	if errUrl, ok := err.(*url.Error); ok {
		return fmt.Errorf(msg+" (url.Error): %s", errUrl)
	}

	return fmt.Errorf(msg+": %s", err.Error())
}
