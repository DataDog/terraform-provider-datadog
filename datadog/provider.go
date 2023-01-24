package datadog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	datadogCommunity "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/transport"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	baseIPRangesSubdomain = "ip-ranges"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	//schema.DescriptionKind = configschema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		//if s.Default != nil {
		//	desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		//}
		if s.ValidateDiagFunc != nil {
			defer func() {
				recover()
			}()
			// Call the validate func with the EnumChecker type. Only supposed to have an effect with enum validate funcs, recover from any panic caused by calling this
			diags := s.ValidateDiagFunc(validators.EnumChecker{}, cty.Path{})
			if len(diags) == 1 && diags[0].Summary == "Allowed values" {
				desc = fmt.Sprintf("%s Valid values are %s.", desc, diags[0].Detail)
			}
		} else if s.Elem != nil {
			defer func() {
				recover()
			}()
			if inner, ok := s.Elem.(*schema.Schema); ok && inner.ValidateDiagFunc != nil {
				diags := inner.ValidateDiagFunc(validators.EnumChecker{}, cty.Path{})
				if len(diags) == 1 && diags[0].Summary == "Allowed values" {
					desc = fmt.Sprintf("%s Valid values are %s.", desc, diags[0].Detail)
				}
			}
		}
		if s.Deprecated != "" {
			desc = fmt.Sprintf("%s **Deprecated.** %s", desc, s.Deprecated)
		}
		return strings.TrimSpace(desc)
	}
}

// DDAPPKeyEnvName name of env var for APP key
const DDAPPKeyEnvName = "DD_APP_KEY"

// DDAPIKeyEnvName name of env var for API key
const DDAPIKeyEnvName = "DD_API_KEY"

// DatadogAPPKeyEnvName name of env var for APP key
const DatadogAPPKeyEnvName = "DATADOG_APP_KEY"

// DatadogAPIKeyEnvName name of env var for API key
const DatadogAPIKeyEnvName = "DATADOG_API_KEY"

// APPKeyEnvVars names of env var for APP key
var APPKeyEnvVars = []string{DDAPPKeyEnvName, DatadogAPPKeyEnvName}

// APIKeyEnvVars names of env var for API key
var APIKeyEnvVars = []string{DDAPIKeyEnvName, DatadogAPIKeyEnvName}

// Provider returns the built datadog provider object
func Provider() *schema.Provider {
	utils.DatadogProvider = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc(APIKeyEnvVars, nil),
				Description: "(Required unless validate is false) Datadog API key. This can also be set via the DD_API_KEY environment variable.",
			},
			"app_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc(APPKeyEnvVars, nil),
				Description: "(Required unless validate is false) Datadog APP key. This can also be set via the DD_APP_KEY environment variable.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"DATADOG_HOST", "DD_HOST"}, nil),
				Description: "The API URL. This can also be set via the DD_HOST environment variable. Note that this URL must not end with the `/api/` path. For example, `https://api.datadoghq.com/` is a correct value, while `https://api.datadoghq.com/api/` is not. And if you're working with \"EU\" version of Datadog, use `https://api.datadoghq.eu/`. Other Datadog region examples: `https://api.us5.datadoghq.com/`, `https://api.us3.datadoghq.com/` and `https://api.ddog-gov.com/`. See https://docs.datadoghq.com/getting_started/site/ for all available regions.",
			},
			"validate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enables validation of the provided API and APP keys during provider initialization. Default is true. When false, api_key and app_key won't be checked.",
			},
			"http_client_retry_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DD_HTTP_CLIENT_RETRY_ENABLED", true),
				Description: "Enables request retries on HTTP status codes 429 and 5xx. Defaults to `true`.",
			},
			"http_client_retry_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("DD_HTTP_CLIENT_RETRY_TIMEOUT", nil),
				Description: "The HTTP request retry timeout period. Defaults to 60 seconds.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"datadog_api_key":                              resourceDatadogApiKey(),
			"datadog_application_key":                      resourceDatadogApplicationKey(),
			"datadog_authn_mapping":                        resourceDatadogAuthnMapping(),
			"datadog_child_organization":                   resourceDatadogChildOrganization(),
			"datadog_cloud_configuration_rule":             resourceDatadogCloudConfigurationRule(),
			"datadog_cloud_workload_security_agent_rule":   resourceDatadogCloudWorkloadSecurityAgentRule(),
			"datadog_dashboard":                            resourceDatadogDashboard(),
			"datadog_dashboard_json":                       resourceDatadogDashboardJSON(),
			"datadog_dashboard_list":                       resourceDatadogDashboardList(),
			"datadog_downtime":                             resourceDatadogDowntime(),
			"datadog_integration_aws":                      resourceDatadogIntegrationAws(),
			"datadog_integration_aws_tag_filter":           resourceDatadogIntegrationAwsTagFilter(),
			"datadog_integration_aws_lambda_arn":           resourceDatadogIntegrationAwsLambdaArn(),
			"datadog_integration_aws_log_collection":       resourceDatadogIntegrationAwsLogCollection(),
			"datadog_integration_azure":                    resourceDatadogIntegrationAzure(),
			"datadog_integration_gcp":                      resourceDatadogIntegrationGcp(),
			"datadog_integration_opsgenie_service_object":  resourceDatadogIntegrationOpsgenieService(),
			"datadog_integration_pagerduty":                resourceDatadogIntegrationPagerduty(),
			"datadog_integration_pagerduty_service_object": resourceDatadogIntegrationPagerdutySO(),
			"datadog_integration_slack_channel":            resourceDatadogIntegrationSlackChannel(),
			"datadog_logs_archive":                         resourceDatadogLogsArchive(),
			"datadog_logs_archive_order":                   resourceDatadogLogsArchiveOrder(),
			"datadog_logs_custom_pipeline":                 resourceDatadogLogsCustomPipeline(),
			"datadog_logs_index":                           resourceDatadogLogsIndex(),
			"datadog_logs_index_order":                     resourceDatadogLogsIndexOrder(),
			"datadog_logs_integration_pipeline":            resourceDatadogLogsIntegrationPipeline(),
			"datadog_logs_metric":                          resourceDatadogLogsMetric(),
			"datadog_logs_pipeline_order":                  resourceDatadogLogsPipelineOrder(),
			"datadog_metric_metadata":                      resourceDatadogMetricMetadata(),
			"datadog_metric_tag_configuration":             resourceDatadogMetricTagConfiguration(),
			"datadog_monitor":                              resourceDatadogMonitor(),
			"datadog_monitor_json":                         resourceDatadogMonitorJSON(),
			"datadog_organization_settings":                resourceDatadogOrganizationSettings(),
			"datadog_role":                                 resourceDatadogRole(),
			"datadog_rum_application":                      resourceDatadogRUMApplication(),
			"datadog_service_account":                      resourceDatadogServiceAccount(),
			"datadog_security_monitoring_default_rule":     resourceDatadogSecurityMonitoringDefaultRule(),
			"datadog_security_monitoring_rule":             resourceDatadogSecurityMonitoringRule(),
			"datadog_security_monitoring_filter":           resourceDatadogSecurityMonitoringFilter(),
			"datadog_sensitive_data_scanner_group":         resourceDatadogSensitiveDataScannerGroup(),
			"datadog_service_level_objective":              resourceDatadogServiceLevelObjective(),
			"datadog_service_definition_yaml":              resourceDatadogServiceDefinitionYAML(),
			"datadog_slo_correction":                       resourceDatadogSloCorrection(),
			"datadog_synthetics_test":                      resourceDatadogSyntheticsTest(),
			"datadog_synthetics_global_variable":           resourceDatadogSyntheticsGlobalVariable(),
			"datadog_synthetics_private_location":          resourceDatadogSyntheticsPrivateLocation(),
			"datadog_user":                                 resourceDatadogUser(),
			"datadog_webhook":                              resourceDatadogWebhook(),
			"datadog_webhook_custom_variable":              resourceDatadogWebhookCustomVariable(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"datadog_api_key":                             dataSourceDatadogApiKey(),
			"datadog_application_key":                     dataSourceDatadogApplicationKey(),
			"datadog_cloud_workload_security_agent_rules": dataSourceDatadogCloudWorkloadSecurityAgentRules(),
			"datadog_dashboard":                           dataSourceDatadogDashboard(),
			"datadog_dashboard_list":                      dataSourceDatadogDashboardList(),
			"datadog_integration_aws_logs_services":       dataSourceDatadogIntegrationAWSLogsServices(),
			"datadog_ip_ranges":                           dataSourceDatadogIPRanges(),
			"datadog_logs_archives_order":                 dataSourceDatadogLogsArchivesOrder(),
			"datadog_logs_indexes":                        dataSourceDatadogLogsIndexes(),
			"datadog_logs_indexes_order":                  dataSourceDatadogLogsIndexesOrder(),
			"datadog_logs_pipelines":                      dataSourceDatadogLogsPipelines(),
			"datadog_monitor":                             dataSourceDatadogMonitor(),
			"datadog_monitors":                            dataSourceDatadogMonitors(),
			"datadog_permissions":                         dataSourceDatadogPermissions(),
			"datadog_role":                                dataSourceDatadogRole(),
			"datadog_roles":                               dataSourceDatadogRoles(),
			"datadog_rum_application":                     dataSourceDatadogRUMApplication(),
			"datadog_security_monitoring_rules":           dataSourceDatadogSecurityMonitoringRules(),
			"datadog_security_monitoring_filters":         dataSourceDatadogSecurityMonitoringFilters(),
			"datadog_sensitive_data_scanner_group":        dataSourceDatadogSensitiveDataScannerGroup(),
			"datadog_service_level_objective":             dataSourceDatadogServiceLevelObjective(),
			"datadog_service_level_objectives":            dataSourceDatadogServiceLevelObjectives(),
			"datadog_synthetics_locations":                dataSourceDatadogSyntheticsLocations(),
			"datadog_synthetics_global_variable":          dataSourceDatadogSyntheticsGlobalVariable(),
			"datadog_synthetics_test":                     dataSourceDatadogSyntheticsTest(),
			"datadog_user":                                dataSourceDatadogUser(),
		},

		ConfigureContextFunc: providerConfigure,
	}

	return utils.DatadogProvider
}

// ProviderConfiguration contains the initialized API clients to communicate with the Datadog API
type ProviderConfiguration struct {
	CommunityClient     *datadogCommunity.Client
	DatadogApiInstances *utils.ApiInstances
	Auth                context.Context

	Now func() time.Time
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	appKey := d.Get("app_key").(string)
	validate := d.Get("validate").(bool)
	httpRetryEnabled := d.Get("http_client_retry_enabled").(bool)

	if validate && (apiKey == "" || appKey == "") {
		return nil, diag.FromErr(errors.New("api_key and app_key must be set unless validate = false"))
	}

	// Initialize the community client
	communityClient := datadogCommunity.NewClient(apiKey, appKey)

	if apiURL := d.Get("api_url").(string); apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", c.Transport)
	communityClient.ExtraHeader["User-Agent"] = utils.GetUserAgent(fmt.Sprintf(
		"datadog-api-client-go/%s (go %s; os %s; arch %s)",
		"go-datadog-api",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	))
	communityClient.HttpClient = c

	if validate {
		log.Println("[INFO] Datadog client successfully initialized, now validating...")
		ok, err := communityClient.Validate()
		if err != nil {
			log.Printf("[ERROR] Datadog Client validation error: %v", err)
			return nil, diag.FromErr(err)
		} else if !ok {
			err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`)
			log.Printf("[ERROR] Datadog Client validation error: %v", err)
			return nil, diag.FromErr(err)
		}
	} else {
		log.Println("[INFO] Skipping key validation (validate = false)")
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	// Initialize http.Client for the Datadog API Clients
	httpClient := http.DefaultClient
	if httpRetryEnabled {
		ctOptions := transport.CustomTransportOptions{}
		if v, ok := d.GetOk("http_client_retry_timeout"); ok {
			timeout := time.Duration(int64(v.(int))) * time.Second
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
				Key: apiKey,
			},
			"appKeyAuth": {
				Key: appKey,
			},
		},
	)
	config := datadog.NewConfiguration()
	config.HTTPClient = httpClient

	config.UserAgent = utils.GetUserAgent(config.UserAgent)
	config.Debug = logging.IsDebugOrHigher()
	if apiURL := d.Get("api_url").(string); apiURL != "" {
		parsedAPIURL, parseErr := url.Parse(apiURL)
		if parseErr != nil {
			return nil, diag.Errorf(`invalid API URL : %v`, parseErr)
		}
		if parsedAPIURL.Host == "" || parsedAPIURL.Scheme == "" {
			return nil, diag.Errorf(`missing protocol or host : %v`, apiURL)
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

	datadogClient := datadog.NewAPIClient(config)

	return &ProviderConfiguration{
		CommunityClient:     communityClient,
		DatadogApiInstances: &utils.ApiInstances{HttpClient: datadogClient},
		Auth:                auth,

		Now: time.Now,
	}, nil
}
