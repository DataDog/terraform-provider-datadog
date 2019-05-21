package datadog

import (
	"errors"
	"log"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
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
			"datadog_downtime":              resourceDatadogDowntime(),
			"datadog_metric_metadata":       resourceDatadogMetricMetadata(),
			"datadog_monitor":               resourceDatadogMonitor(),
			"datadog_synthetics_test":       resourceDatadogSyntheticsTest(),
			"datadog_timeboard":             resourceDatadogTimeboard(),
			"datadog_screenboard":           resourceDatadogScreenboard(),
			"datadog_user":                  resourceDatadogUser(),
			"datadog_integration_gcp":       resourceDatadogIntegrationGcp(),
			"datadog_integration_aws":       resourceDatadogIntegrationAws(),
			"datadog_integration_pagerduty": resourceDatadogIntegrationPagerduty(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	client := datadog.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))
	if apiURL := d.Get("api_url").(string); apiURL != "" {
		client.SetBaseUrl(apiURL)
	}

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", c.Transport)
	client.HttpClient = c

	log.Println("[INFO] Datadog client successfully initialized, now validating...")
	ok, err := client.Validate()
	if err != nil {
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return client, err
	} else if !ok {
		err := errors.New(`No valid credential sources found for Datadog Provider. Please see https://terraform.io/docs/providers/datadog/index.html for more information on providing credentials for the Datadog Provider`)
		log.Printf("[ERROR] Datadog Client validation error: %v", err)
		return client, err
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	return client, nil
}
