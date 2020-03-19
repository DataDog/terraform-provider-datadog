package datadog

import (
	"errors"
	"fmt"
	"log"
	"runtime"

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
		return communityClient, err
	} else if !ok {
		err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and see https://terraform.io/docs/providers/datadog/index.html for more information on providing credentials for the Datadog Provider`)
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return communityClient, err
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	return &ProviderConfiguration{
		CommunityClient: communityClient,
	}, nil
}

func translateClientError(err error, msg string) error {
	if msg == "" {
		msg = "an error occurred"
	}

	return fmt.Errorf(msg+": %s", err.Error())
}
