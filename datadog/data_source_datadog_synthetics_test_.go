package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"regexp"
)

func dataSourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog Synthetic Test.",
		ReadContext: dataSourceDatadogSyntheticsTestRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"test_id": {
					Description: "The synthetic test id or URL to search for",
					Type:        schema.TypeString,
					Required:    true,
				},
				"name": {
					Description: "The name of the synthetic test.",
					Type:        schema.TypeString,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"tags": {
					Description: "A list of tags assigned to the synthetic test.",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"url": {
					Description: "The start URL of the synthetic test.",
					Type:        schema.TypeString,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"type": {
					Description: "The type of the synthetic test.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"subtype": {
					Description: "The subtype of the synthetic test. Only set for API tests.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"status": {
					Description: "Whether the synthetic test is started (`live`) or paused (`paused`).",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"message": {
					Description: "A message to include with notifications for this synthetic test.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"monitor_id": {
					Description: "ID of the monitor associated with the synthetic test.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"locations": {
					Description: "Array of locations used to run the synthetic test.",
					Type:        schema.TypeSet,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"device_ids": {
					Description: "Array with the different device IDs used to run the test. Only set for browser tests.",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"options_list":        dataSourceSyntheticsTestOptionsList(),
				"mobile_options_list": dataSourceSyntheticsMobileTestOptionsList(),
			}
		},
	}
}

// dataSourceSyntheticsTestOptionsList must declare every key buildTerraformTestOptions can
// emit. The read passes that function's output straight to d.Set, which fails at runtime on
// any key missing from this schema, so new options added there need a matching entry here.
func dataSourceSyntheticsTestOptionsList() *schema.Schema {
	return &schema.Schema{
		Description: "The synthetic test extra options.",
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allow_insecure": {
					Description: "Allows loading insecure content for a request in an API test or in a multistep API test step.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"follow_redirects": {
					Description: "Determines whether or not the API HTTP test should follow redirects.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"tick_every": {
					Description: "How often the test should run (in seconds).",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"accept_self_signed": {
					Description: "For SSL tests, whether or not the test should allow self signed certificates.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"min_location_failed": {
					Description: "Minimum number of locations in failure required to trigger an alert.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"min_failure_duration": {
					Description: "Minimum amount of time in failure required to trigger an alert (in seconds).",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"monitor_name": {
					Description: "The monitor name is used for the alert title as well as for all monitor dashboard widgets and SLOs.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"monitor_priority": {
					Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"restricted_roles": {
					Deprecated:  "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
					Description: "A list of role identifiers pulled from the Roles API to restrict read and write access. Included for parity with the `datadog_synthetics_test` resource.",
					Type:        schema.TypeSet,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"no_screenshot": {
					Description: "Prevents saving screenshots of the steps.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"check_certificate_revocation": {
					Description: "For SSL tests, whether or not the test should fail on revoked certificate in stapled OCSP.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"disable_aia_intermediate_fetching": {
					Description: "For SSL tests, whether or not the test should disable fetching intermediate certificates from AIA.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"ignore_certificate_validation": {
					Description: "Ignore server certificate error for SSL tests.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"ignore_server_certificate_error": {
					Description: "Ignore server certificate error for browser tests.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"disable_csp": {
					Description: "Disable Content Security Policy for browser tests.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"disable_cors": {
					Description: "Disable Cross-Origin Resource Sharing for browser tests.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"initial_navigation_timeout": {
					Description: "Timeout before declaring the initial step as failed (in seconds) for browser tests.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"blocked_request_patterns": {
					Description: "Blocked URL patterns. Requests made to URLs matching any of the patterns listed here will be blocked.",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"capture_network_payloads": {
					Description: "Capture HTTP request/response headers and bodies for Fetch/XHR calls made during browser tests.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"http_version": {
					Description: "HTTP version to use for an HTTP request in an API test or step.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"retry": {
					Description: "Object describing the retry strategy to apply to a Synthetic test.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"count": {
								Description: "Number of retries needed to consider a location as failed before sending a notification alert.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"interval": {
								Description: "Interval between a failed test and the next retry in milliseconds.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
						},
					},
				},
				"monitor_options": {
					Description: "Object containing the options for a Synthetic test as a monitor (for example, renotification).",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"renotify_interval": {
								Description: "Specify a renotification frequency in minutes.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"renotify_occurrences": {
								Description: "The number of times a monitor renotifies. It can only be set if `renotify_interval` is set.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"escalation_message": {
								Description: "A message to include with a re-notification.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"notification_preset_name": {
								Description: "The name of the preset for the notification for the monitor.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
				"scheduling": {
					Description: "Object containing timeframes and timezone used for advanced scheduling.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"timezone": {
								Description: "Timezone in which the timeframe is based.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"timeframes": {
								Description: "Array containing objects describing the scheduling pattern to apply to each day.",
								Type:        schema.TypeList,
								Computed:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"day": {
											Description: "Number representing the day of the week.",
											Type:        schema.TypeInt,
											Computed:    true,
										},
										"from": {
											Description: "The hour of the day on which scheduling starts.",
											Type:        schema.TypeString,
											Computed:    true,
										},
										"to": {
											Description: "The hour of the day on which scheduling ends.",
											Type:        schema.TypeString,
											Computed:    true,
										},
									},
								},
							},
						},
					},
				},
				"ci": {
					Description: "CI/CD options for a Synthetic test.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"execution_rule": {
								Description: "Execution rule for a Synthetics test.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
				"rum_settings": {
					Description: "The RUM data collection settings for the Synthetic browser test.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"is_enabled": {
								Description: "Determines whether RUM data is collected during test runs.",
								Type:        schema.TypeBool,
								Computed:    true,
							},
							"application_id": {
								Description: "RUM application ID used to collect RUM data for the browser test.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"client_token_id": {
								Description: "RUM application API key ID used to collect RUM data for the browser test.",
								Type:        schema.TypeInt,
								Computed:    true,
								Sensitive:   true,
							},
						},
					},
				},
			},
		},
	}
}

// dataSourceSyntheticsMobileTestOptionsList must declare every key
// buildTerraformMobileTestOptions can emit, for the same reason as
// dataSourceSyntheticsTestOptionsList above.
func dataSourceSyntheticsMobileTestOptionsList() *schema.Schema {
	return &schema.Schema{
		Description: "The mobile synthetic test extra options.",
		Type:        schema.TypeList,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"tick_every": {
					Description: "How often the test should run (in seconds).",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"min_failure_duration": {
					Description: "Minimum amount of time in failure required to trigger an alert (in seconds).",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"monitor_name": {
					Description: "The monitor name is used for the alert title as well as for all monitor dashboard widgets and SLOs.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"monitor_priority": {
					Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"device_ids": {
					Description: "Array with the different device IDs used to run the test.",
					Type:        schema.TypeList,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"default_step_timeout": {
					Description: "Default timeout for steps in the test (in seconds).",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"no_screenshot": {
					Description: "Prevents saving screenshots of the steps.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"allow_application_crash": {
					Description: "Whether the application crashing is considered a failure.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"disable_auto_accept_alert": {
					Description: "Whether to disable automatically accepting alerts during the test.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"restricted_roles": {
					Deprecated:  "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
					Description: "A list of role identifiers pulled from the Roles API to restrict read and write access. Included for parity with the `datadog_synthetics_test` resource.",
					Type:        schema.TypeSet,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
				},
				"retry": {
					Description: "Object describing the retry strategy to apply to a Synthetic test.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"count": {
								Description: "Number of retries needed to consider a location as failed before sending a notification alert.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"interval": {
								Description: "Interval between a failed test and the next retry in milliseconds.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
						},
					},
				},
				"monitor_options": {
					Description: "Object containing the options for a Synthetic test as a monitor (for example, renotification).",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"renotify_interval": {
								Description: "Specify a renotification frequency in minutes.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"renotify_occurrences": {
								Description: "The number of times a monitor renotifies. It can only be set if `renotify_interval` is set.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"escalation_message": {
								Description: "A message to include with a re-notification.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"notification_preset_name": {
								Description: "The name of the preset for the notification for the monitor.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
				"scheduling": {
					Description: "Object containing timeframes and timezone used for advanced scheduling.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"timezone": {
								Description: "Timezone in which the timeframe is based.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"timeframes": {
								Description: "Array containing objects describing the scheduling pattern to apply to each day.",
								Type:        schema.TypeList,
								Computed:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"day": {
											Description: "Number representing the day of the week.",
											Type:        schema.TypeInt,
											Computed:    true,
										},
										"from": {
											Description: "The hour of the day on which scheduling starts.",
											Type:        schema.TypeString,
											Computed:    true,
										},
										"to": {
											Description: "The hour of the day on which scheduling ends.",
											Type:        schema.TypeString,
											Computed:    true,
										},
									},
								},
							},
						},
					},
				},
				"ci": {
					Description: "CI/CD options for a Synthetic test.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"execution_rule": {
								Description: "Execution rule for a Synthetics test.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
				"bindings": {
					Description: "Restriction policy bindings for the Synthetic mobile test.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"principals": {
								Description: "List of principals for the binding.",
								Type:        schema.TypeList,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Computed:    true,
							},
							"relation": {
								Description: "The relation restriction for the binding.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
				"mobile_application": {
					Description: "Mobile application to run the test against.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"application_id": {
								Description: "The ID of the mobile application.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"reference_id": {
								Description: "The reference ID of the mobile application.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"reference_type": {
								Description: "The reference type of the mobile application.",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}

// syntheticsTestDataSourceCommon covers the SyntheticsAPITest / SyntheticsBrowserTest
// accessors shared by the data source read. GetType and GetSubtype are excluded: their
// concrete return types differ per test kind, so those two attributes are set in the
// caller's per-branch code instead. Mobile and network tests don't satisfy this interface
// and are handled separately.
type syntheticsTestDataSourceCommon interface {
	GetName() string
	GetTags() []string
	GetMessage() string
	GetMonitorId() int64
	GetLocations() []string
	GetStatus() datadogV1.SyntheticsTestPauseStatus
	GetOptions() datadogV1.SyntheticsTestOptions
}

func dataSourceDatadogSyntheticsTestSetCommonState(d *schema.ResourceData, test syntheticsTestDataSourceCommon) diag.Diagnostics {
	if err := d.Set("name", test.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", test.GetTags()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", test.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_id", test.GetMonitorId()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("locations", test.GetLocations()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", string(test.GetStatus())); err != nil {
		return diag.FromErr(err)
	}

	options := test.GetOptions()
	if err := d.Set("device_ids", options.DeviceIds); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("options_list", buildTerraformTestOptions(options)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func dataSourceDatadogSyntheticsTestSetAPITestState(d *schema.ResourceData, test *datadogV1.SyntheticsAPITest) diag.Diagnostics {
	d.SetId(test.GetPublicId())
	if err := d.Set("type", string(test.GetType())); err != nil {
		return diag.FromErr(err)
	}
	if test.HasSubtype() {
		if err := d.Set("subtype", string(test.GetSubtype())); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("url", test.Config.Request.GetUrl()); err != nil {
		return diag.FromErr(err)
	}
	return dataSourceDatadogSyntheticsTestSetCommonState(d, test)
}

func dataSourceDatadogSyntheticsTestSetBrowserTestState(d *schema.ResourceData, test *datadogV1.SyntheticsBrowserTest) diag.Diagnostics {
	d.SetId(test.GetPublicId())
	if err := d.Set("type", string(test.GetType())); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", test.Config.Request.GetUrl()); err != nil {
		return diag.FromErr(err)
	}
	return dataSourceDatadogSyntheticsTestSetCommonState(d, test)
}

// dataSourceDatadogSyntheticsTestSetMobileTestState mirrors updateSyntheticsMobileTestLocalState.
// Mobile tests carry no locations, subtype, or request URL, and their options have their own
// shape, so they populate `mobile_options_list` rather than `options_list`.
func dataSourceDatadogSyntheticsTestSetMobileTestState(d *schema.ResourceData, test *datadogV1.SyntheticsMobileTest) diag.Diagnostics {
	d.SetId(test.GetPublicId())
	if err := d.Set("type", string(test.GetType())); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", test.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", test.GetTags()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", test.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_id", test.GetMonitorId()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", string(test.GetStatus())); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("mobile_options_list", buildTerraformMobileTestOptions(test.GetOptions())); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// dataSourceDatadogSyntheticsTestSetNetworkTestState mirrors updateSyntheticsNetworkTestLocalState.
// Network tests come from the v2 API; their options are a subset of the v1 options, so they are
// converted and reuse `options_list`.
func dataSourceDatadogSyntheticsTestSetNetworkTestState(d *schema.ResourceData, response *datadogV2.SyntheticsNetworkTestResponse) diag.Diagnostics {
	test := response.Data.GetAttributes()

	d.SetId(response.Data.GetId())
	if err := d.Set("type", string(datadogV1.SYNTHETICSTESTDETAILSTYPE_NETWORK)); err != nil {
		return diag.FromErr(err)
	}
	if test.HasSubtype() {
		if err := d.Set("subtype", string(test.GetSubtype())); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("name", test.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", test.GetTags()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", test.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if test.HasMonitorId() {
		if err := d.Set("monitor_id", test.GetMonitorId()); err != nil {
			return diag.FromErr(err)
		}
	}
	if test.HasStatus() {
		if err := d.Set("status", string(test.GetStatus())); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("locations", test.GetLocations()); err != nil {
		return diag.FromErr(err)
	}

	optionsV2 := test.GetOptions()
	optionsJSON, err := optionsV2.MarshalJSON()
	if err != nil {
		return diag.FromErr(err)
	}
	optionsV1 := datadogV1.NewSyntheticsTestOptions()
	if err := optionsV1.UnmarshalJSON(optionsJSON); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("options_list", buildTerraformTestOptions(*optionsV1)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func dataSourceDatadogSyntheticsTestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	urlRegex := regexp.MustCompile(`https:\/\/(.*)\.(datadoghq|ddog-gov)\.(com|eu)\/synthetics\/details\/`)
	searchedId := urlRegex.ReplaceAllString(d.Get("test_id").(string), "")

	// Fetch the generic test first to detect its type, then call the type-specific endpoint,
	// the same way resourceDatadogSyntheticsTestRead does. The type-specific endpoints can't
	// be used to probe for the type: given a test of another type they still return HTTP 200,
	// leaving the type enum at its zero value and flagging the object unparsed rather than
	// returning an error.
	syntheticsTest, httpresp, err := apiInstances.GetSyntheticsApiV1().GetTest(auth, searchedId)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics test")
	}
	if err := utils.CheckForUnparsed(syntheticsTest); err != nil {
		return diag.FromErr(err)
	}

	switch syntheticsTest.GetType() {
	case datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER:
		test, httpresp, err := apiInstances.GetSyntheticsApiV1().GetBrowserTest(auth, searchedId)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics browser test")
		}
		if err := utils.CheckForUnparsed(test); err != nil {
			return diag.FromErr(err)
		}
		return dataSourceDatadogSyntheticsTestSetBrowserTestState(d, &test)

	case datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE:
		test, httpresp, err := apiInstances.GetSyntheticsApiV1().GetMobileTest(auth, searchedId)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics mobile test")
		}
		if err := utils.CheckForUnparsed(test); err != nil {
			return diag.FromErr(err)
		}
		return dataSourceDatadogSyntheticsTestSetMobileTestState(d, &test)

	case datadogV1.SYNTHETICSTESTDETAILSTYPE_NETWORK:
		response, httpresp, err := apiInstances.GetSyntheticsApiV2().GetSyntheticsNetworkTest(auth, searchedId)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics network test")
		}
		if err := utils.CheckForUnparsed(response); err != nil {
			return diag.FromErr(err)
		}
		return dataSourceDatadogSyntheticsTestSetNetworkTestState(d, &response)

	default:
		test, httpresp, err := apiInstances.GetSyntheticsApiV1().GetAPITest(auth, searchedId)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics api test")
		}
		if err := utils.CheckForUnparsed(test); err != nil {
			return diag.FromErr(err)
		}
		return dataSourceDatadogSyntheticsTestSetAPITestState(d, &test)
	}
}
