package datadog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	datadogCommunity "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
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

// Provider returns the built datadog provider object
func Provider() *schema.Provider {
	utils.DatadogProvider = &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "(Required unless validate is false) Datadog API key. This can also be set via the DD_API_KEY environment variable.",
			},
			"app_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "(Required unless validate is false) Datadog APP key. This can also be set via the DD_APP_KEY environment variable.",
			},
			"api_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The API URL. This can also be set via the DD_HOST environment variable. Note that this URL must not end with the `/api/` path. For example, `https://api.datadoghq.com/` is a correct value, while `https://api.datadoghq.com/api/` is not. And if you're working with \"EU\" version of Datadog, use `https://api.datadoghq.eu/`. Other Datadog region examples: `https://api.us5.datadoghq.com/`, `https://api.us3.datadoghq.com/` and `https://api.ddog-gov.com/`. See https://docs.datadoghq.com/getting_started/site/ for all available regions.",
			},
			"validate": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Enables validation of the provided API key during provider initialization. Valid values are [`true`, `false`]. Default is true. When false, api_key won't be checked.",
				ValidateFunc: validation.StringInSlice([]string{"true", "false"}, true),
			},
			"http_client_retry_enabled": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Enables request retries on HTTP status codes 429 and 5xx. Valid values are [`true`, `false`]. Defaults to `true`.",
				ValidateFunc: validation.StringInSlice([]string{"true", "false"}, true),
			},
			"http_client_retry_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The HTTP request retry timeout period. Defaults to 60 seconds.",
			},
			"http_client_retry_backoff_multiplier": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The HTTP request retry back off multiplier. Defaults to 2.",
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					value, ok := v.(int)
					var diags diag.Diagnostics
					if ok && value <= 0 {
						return append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Backoff multiplier must be greater than 0.",
						})
					}
					return diags
				},
			},
			"http_client_retry_backoff_base": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The HTTP request retry back off base. Defaults to 2.",
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					value, ok := v.(int)
					var diags diag.Diagnostics
					if ok && value <= 0 {
						return append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Backoff base must be greater than 0.",
						})
					}
					return diags
				},
			},
			"http_client_retry_max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The HTTP request maximum retry number. Defaults to 3.",
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					value, ok := v.(int)
					var diags diag.Diagnostics
					if ok && (value <= 0 || value > 5) {
						return append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Max retries must be between 0 and 5",
						})
					}
					return diags
				},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"datadog_application_key":                      resourceDatadogApplicationKey(),
			"datadog_authn_mapping":                        resourceDatadogAuthnMapping(),
			"datadog_child_organization":                   resourceDatadogChildOrganization(),
			"datadog_cloud_configuration_rule":             resourceDatadogCloudConfigurationRule(),
			"datadog_cloud_workload_security_agent_rule":   resourceDatadogCloudWorkloadSecurityAgentRule(),
			"datadog_dashboard":                            resourceDatadogDashboard(),
			"datadog_dashboard_json":                       resourceDatadogDashboardJSON(),
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
			"datadog_ip_allowlist":                         resourceDatadogIPAllowlist(),
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
			"datadog_monitor_config_policy":                resourceDatadogMonitorConfigPolicy(),
			"datadog_monitor_json":                         resourceDatadogMonitorJSON(),
			"datadog_organization_settings":                resourceDatadogOrganizationSettings(),
			"datadog_role":                                 resourceDatadogRole(),
			"datadog_rum_application":                      resourceDatadogRUMApplication(),
			"datadog_service_account":                      resourceDatadogServiceAccount(),
			"datadog_security_monitoring_default_rule":     resourceDatadogSecurityMonitoringDefaultRule(),
			"datadog_security_monitoring_rule":             resourceDatadogSecurityMonitoringRule(),
			"datadog_security_monitoring_filter":           resourceDatadogSecurityMonitoringFilter(),
			"datadog_sensitive_data_scanner_group":         resourceDatadogSensitiveDataScannerGroup(),
			"datadog_sensitive_data_scanner_rule":          resourceDatadogSensitiveDataScannerRule(),
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
			"datadog_application_key":                         dataSourceDatadogApplicationKey(),
			"datadog_cloud_workload_security_agent_rules":     dataSourceDatadogCloudWorkloadSecurityAgentRules(),
			"datadog_dashboard":                               dataSourceDatadogDashboard(),
			"datadog_integration_aws_logs_services":           dataSourceDatadogIntegrationAWSLogsServices(),
			"datadog_logs_archives_order":                     dataSourceDatadogLogsArchivesOrder(),
			"datadog_logs_indexes":                            dataSourceDatadogLogsIndexes(),
			"datadog_logs_indexes_order":                      dataSourceDatadogLogsIndexesOrder(),
			"datadog_logs_pipelines":                          dataSourceDatadogLogsPipelines(),
			"datadog_monitor":                                 dataSourceDatadogMonitor(),
			"datadog_monitors":                                dataSourceDatadogMonitors(),
			"datadog_monitor_config_policies":                 dataSourceDatadogMonitorConfigPolicies(),
			"datadog_permissions":                             dataSourceDatadogPermissions(),
			"datadog_role":                                    dataSourceDatadogRole(),
			"datadog_roles":                                   dataSourceDatadogRoles(),
			"datadog_rum_application":                         dataSourceDatadogRUMApplication(),
			"datadog_security_monitoring_rules":               dataSourceDatadogSecurityMonitoringRules(),
			"datadog_security_monitoring_filters":             dataSourceDatadogSecurityMonitoringFilters(),
			"datadog_sensitive_data_scanner_standard_pattern": dataSourceDatadogSensitiveDataScannerStandardPattern(),
			"datadog_service_level_objective":                 dataSourceDatadogServiceLevelObjective(),
			"datadog_service_level_objectives":                dataSourceDatadogServiceLevelObjectives(),
			"datadog_synthetics_locations":                    dataSourceDatadogSyntheticsLocations(),
			"datadog_synthetics_global_variable":              dataSourceDatadogSyntheticsGlobalVariable(),
			"datadog_synthetics_test":                         dataSourceDatadogSyntheticsTest(),
			"datadog_user":                                    dataSourceDatadogUser(),
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
	if apiKey == "" {
		apiKey, _ = utils.GetMultiEnvVar(utils.APIKeyEnvVars[:]...)
	}

	appKey := d.Get("app_key").(string)
	if appKey == "" {
		appKey, _ = utils.GetMultiEnvVar(utils.APPKeyEnvVars[:]...)
	}

	apiURL := d.Get("api_url").(string)
	if apiURL == "" {
		apiURL, _ = utils.GetMultiEnvVar(utils.APIUrlEnvVars[:]...)
	}

	httpRetryEnabled := true
	httpRetryEnabledStr := d.Get("http_client_retry_enabled").(string)
	if httpRetryEnabledStr == "" {
		envVal, err := utils.GetMultiEnvVar(utils.DDHTTPRetryEnabled)
		if err == nil {
			httpRetryEnabled, _ = strconv.ParseBool(envVal)
		}
	} else {
		httpRetryEnabled, _ = strconv.ParseBool(httpRetryEnabledStr)
	}

	validate := true
	if v := d.Get("validate").(string); v != "" {
		validate, _ = strconv.ParseBool(v)
	}

	if validate && (apiKey == "" || appKey == "") {
		return nil, diag.FromErr(errors.New("api_key and app_key must be set unless validate = false"))
	}

	// Initialize the community client
	communityClient := datadogCommunity.NewClient(apiKey, appKey)

	if apiURL != "" {
		communityClient.SetBaseUrl(apiURL)
	}

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewLoggingHTTPTransport(c.Transport)
	communityClient.ExtraHeader["User-Agent"] = utils.GetUserAgent(fmt.Sprintf(
		"datadog-api-client-go/%s (go %s; os %s; arch %s)",
		"go-datadog-api",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	))
	communityClient.HttpClient = c

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
	config.RetryConfiguration.EnableRetry = httpRetryEnabled

	if timeoutInterface, ok := d.GetOk("http_client_retry_timeout"); ok {
		timeout := time.Duration(int64(timeoutInterface.(int))) * time.Second
		config.RetryConfiguration.HTTPRetryTimeout = timeout
	} else {
		envVal, err := utils.GetMultiEnvVar(utils.DDHTTPRetryTimeout)
		if err == nil {
			vInt, _ := strconv.Atoi(envVal)
			timeout := time.Duration(int64(vInt)) * time.Second
			config.RetryConfiguration.HTTPRetryTimeout = timeout
		}
	}

	if backoffMultiplierInterface, ok := d.GetOk("http_client_retry_backoff_multiplier"); ok {
		backOffMultiplier := float64(backoffMultiplierInterface.(int))
		config.RetryConfiguration.BackOffMultiplier = backOffMultiplier
	} else {
		envVal, err := utils.GetMultiEnvVar(utils.DDHTTPRetryBackoffMultiplier)
		if err == nil {
			fVal, _ := strconv.ParseFloat(envVal, 64)
			config.RetryConfiguration.BackOffMultiplier = fVal
		}
	}

	if retryBackoffBaseInterface, ok := d.GetOk("http_client_retry_backoff_base"); ok {
		retryBackoffBase := float64(retryBackoffBaseInterface.(int))
		config.RetryConfiguration.BackOffBase = retryBackoffBase
	} else {
		envVal, err := utils.GetMultiEnvVar(utils.DDHTTPRetryBackoffBase)
		if err == nil {
			fVal, _ := strconv.ParseFloat(envVal, 64)
			config.RetryConfiguration.BackOffBase = fVal
		}
	}

	if maxRetryInterface, ok := d.GetOk("http_client_retry_max_retries"); ok {
		config.RetryConfiguration.MaxRetries = maxRetryInterface.(int)
	} else {
		envVal, err := utils.GetMultiEnvVar(utils.DDHTTPRetryMaxRetries)
		if err == nil {
			fVal, _ := strconv.Atoi(envVal)
			config.RetryConfiguration.MaxRetries = fVal
		}
	}

	config.UserAgent = utils.GetUserAgent(config.UserAgent)
	config.Debug = logging.IsDebugOrHigher()
	if apiURL != "" {
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

	datadogClient := datadog.NewAPIClient(config)
	apiInstances := &utils.ApiInstances{HttpClient: datadogClient}
	if validate {
		log.Println("[INFO] Datadog client successfully initialized, now validating...")
		resp, _, err := apiInstances.GetAuthenticationApiV1().Validate(auth)
		if err != nil {
			log.Printf("[ERROR] Datadog Client validation error: %v", err)
			return nil, diag.FromErr(err)
		}
		valid, ok := resp.GetValidOk()
		if (ok && !*valid) || !ok {
			err := errors.New(`Invalid or missing credentials provided to the Datadog Provider. Please confirm your API and APP keys are valid and are for the correct region, see https://www.terraform.io/docs/providers/datadog/ for more information on providing credentials for the Datadog Provider`)
			log.Printf("[ERROR] Datadog Client validation error: %v", err)
			return nil, diag.FromErr(err)
		}
	} else {
		log.Println("[INFO] Skipping key validation (validate = false)")
	}
	log.Printf("[INFO] Datadog Client successfully validated.")

	return &ProviderConfiguration{
		CommunityClient:     communityClient,
		DatadogApiInstances: apiInstances,
		Auth:                auth,

		Now: time.Now,
	}, nil
}
