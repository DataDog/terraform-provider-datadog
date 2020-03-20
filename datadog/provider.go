package datadog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strings"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/version"
	datadogCommunity "github.com/zorkian/go-datadog-api"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
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
}

//ProviderConfiguration contains the initialized API clients to communicate with the Datadog API
type ProviderConfiguration struct {
	CommunityClient *datadogCommunity.Client
	DatadogClientV1 *datadog.APIClient
	Auth            context.Context
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// Initialize the community client
	communityClient := datadogCommunity.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))

	if apiURL := d.Get("api_url").(string); apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", c.Transport)
	communityClient.HttpClient = c
	communityClient.ExtraHeader["User-Agent"] = fmt.Sprintf("Datadog/%s/terraform (%s)", version.ProviderVersion, runtime.Version())

	log.Println("[INFO] Datadog client successfully initialized, now validating...")
	ok, err := communityClient.Validate()
	if err != nil {
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return nil, err
	} else if !ok {
		err := errors.New(`invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and see https://terraform.io/docs/providers/datadog/index.html for more information on providing credentials for the Datadog Provider`)
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return nil, err
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	// Initialize the official Datadog client
	auth := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: d.Get("api_key").(string),
			},
			"appKeyAuth": {
				Key: d.Get("app_key").(string),
			},
		},
	)
	config := datadog.NewConfiguration()
	if apiURL := d.Get("api_url").(string); apiURL != "" {
		if strings.Contains(apiURL, "datadoghq.eu") {
			auth = context.WithValue(auth, datadog.ContextServerVariables, map[string]string{
				"site": "datadoghq.eu",
			})
		}
	}
	datadogClient := datadog.NewAPIClient(config)
	return &ProviderConfiguration{
		CommunityClient: communityClient,
		DatadogClientV1: datadogClient,
		Auth:            auth,
	}, nil
}

func translateClientError(err error, msg string) error {
	if msg == "" {
		msg = "an error occurred"
	}

	if errBody, ok := err.(datadog.GenericOpenAPIError); ok {
		return fmt.Errorf(msg+" (datadog.GenericOpenAPIError): %s", string(errBody.Body()))
	}
	if errUrl, ok := err.(*url.Error); ok {
		return fmt.Errorf(msg+" (url.Error): %s", errUrl)
	}

	return fmt.Errorf(msg+": %s", err.Error())
}
