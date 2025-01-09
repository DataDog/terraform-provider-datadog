// For more info about writing custom provider: https://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"bytes"
	"compress/zlib"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	_nethttp "net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

/*
 * Resource
 */

func resourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.",
		CreateContext: resourceDatadogSyntheticsTestCreate,
		ReadContext:   resourceDatadogSyntheticsTestRead,
		UpdateContext: resourceDatadogSyntheticsTestUpdate,
		DeleteContext: resourceDatadogSyntheticsTestDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"type": {
					Description:      "Synthetics test type.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestDetailsTypeFromValue),
				},
				"subtype": {
					Description: "The subtype of the Synthetic API test. Defaults to `http`.",
					Type:        schema.TypeString,
					Optional:    true,
					DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
						if d.Get("type") == "api" && old == "http" && new == "" {
							// defaults to http if type is api for retro-compatibility
							return true
						}
						return old == new
					},
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestDetailsSubTypeFromValue),
				},
				"request_definition": {
					Description: "Required if `type = \"api\"`. The synthetics test request.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem:        syntheticsTestRequest(),
				},
				"request_headers":            syntheticsTestRequestHeaders(),
				"request_query":              syntheticsTestRequestQuery(),
				"request_basicauth":          syntheticsTestRequestBasicAuth(),
				"request_proxy":              syntheticsTestRequestProxy(),
				"request_client_certificate": syntheticsTestRequestClientCertificate(),
				"request_metadata":           syntheticsTestRequestMetadata(),
				"request_file":               syntheticsTestRequestFile(),
				"assertion":                  syntheticsAPIAssertion(),
				"browser_variable":           syntheticsBrowserVariable(),
				"config_variable":            syntheticsConfigVariable(),
				"config_initial_application_arguments": {
					Description: "Initial application arguments for the mobile test.",
					Type:        schema.TypeMap,
					Optional:    true,
				},
				"variables_from_script": {
					Description: "Variables defined from JavaScript code for API HTTP tests.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"device_ids": {
					Description: "Required if `type = \"browser\"`. Array with the different device IDs used to run the test.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
					},
				},
				"locations": {
					Description: "Array of locations used to run the test. Refer to [the Datadog Synthetics location data source](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/data-sources/synthetics_locations) to retrieve the list of locations.",
					Type:        schema.TypeSet,
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"options_list":        syntheticsTestOptionsList(),
				"mobile_options_list": syntheticsMobileTestOptionsList(),
				"name": {
					Description: "Name of Datadog synthetics test.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"message": {
					Description: "A message to include with notifications for this synthetics test. Email notifications can be sent to specific users by using the same `@username` notation as events.",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "",
				},
				"tags": {
					Description: "A list of tags to associate with your synthetics test. This can help you categorize and filter tests in the manage synthetics page of the UI. Default is an empty list (`[]`).",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
					},
				},
				"status": {
					Description:      "Define whether you want to start (`live`) or pause (`paused`) a Synthetic test.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestPauseStatusFromValue),
				},
				"monitor_id": {
					Description: "ID of the monitor associated with the Datadog synthetics test.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"browser_step": syntheticsTestBrowserStep(),
				"api_step":     syntheticsTestAPIStep(),
				"mobile_step":  syntheticsTestMobileStep(),
				"set_cookie": {
					Description: "Cookies to be used for a browser test request, using the [Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie) syntax.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"force_delete_dependencies": {
					Description: "A boolean indicating whether this synthetics test can be deleted even if it's referenced by other resources (for example, SLOs and composite monitors).",
					Type:        schema.TypeBool,
					Optional:    true,
				},
			}
		},
	}
}

/*
 * Schemas
 */

func syntheticsTestRequest() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"method": {
				Description: "Either the HTTP method/verb to use or a gRPC method available on the service set in the `service` field. Required if `subtype` is `HTTP` or if `subtype` is `grpc` and `callType` is `unary`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"url": {
				Description: "The URL to send the request to.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"body": {
				Description: "The request body.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"body_type": {
				Description:      "Type of the request body.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestRequestBodyTypeFromValue),
			},
			"timeout": {
				Description: "Timeout in seconds for the test.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"host": {
				Description: "Host name to perform the test with.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"port": {
				Description: "Port to use when performing the test.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"dns_server": {
				Description: "DNS server to use for DNS tests (`subtype = \"dns\"`).",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"dns_server_port": {
				Description: "DNS server port to use for DNS tests.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"no_saving_response_body": {
				Description: "Determines whether or not to save the response body.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"number_of_packets": {
				Description:  "Number of pings to use per test for ICMP tests (`subtype = \"icmp\"`) between 0 and 10.",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"should_track_hops": {
				Description: "This will turn on a traceroute probe to discover all gateways along the path to the host destination. For ICMP tests (`subtype = \"icmp\"`).",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"servername": {
				Description: "For SSL tests, it specifies on which server you want to initiate the TLS handshake, allowing the server to present one of multiple possible certificates on the same IP address and TCP port number.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"message": {
				Description: "For UDP and websocket tests, message to send with the request.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"call_type": {
				Description:      "The type of gRPC call to perform.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestCallTypeFromValue),
			},
			"service": {
				Description: "The gRPC service on which you want to perform the gRPC call.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"certificate_domains": {
				Description: "By default, the client certificate is applied on the domain of the starting URL for browser tests. If you want your client certificate to be applied on other domains instead, add them in `certificate_domains`.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"persist_cookies": {
				Description: "Persist cookies across redirects.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"proto_json_descriptor": {
				Description: "A protobuf JSON descriptor.",
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "Use `plain_proto_file` instead.",
			},
			"plain_proto_file": {
				Description: "The content of a proto file as a string.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"http_version": {
				Description: "HTTP version to use for an HTTP request in an API test or step.",
				Deprecated:  "Use `http_version` in the `options_list` field instead.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func syntheticsTestRequestHeaders() *schema.Schema {
	return &schema.Schema{
		Description:  "Header name and value map.",
		Type:         schema.TypeMap,
		Optional:     true,
		ValidateFunc: validators.ValidateHttpRequestHeader,
	}
}

func syntheticsTestRequestQuery() *schema.Schema {
	return &schema.Schema{
		Description: "Query arguments name and value map.",
		Type:        schema.TypeMap,
		Optional:    true,
	}
}

func syntheticsTestRequestBasicAuth() *schema.Schema {
	return &schema.Schema{
		Description: "The HTTP basic authentication credentials. Exactly one nested block is allowed with the structure below.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Description:  "Type of basic authentication to use when performing the test.",
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "web",
					ValidateFunc: validation.StringInSlice([]string{"web", "sigv4", "ntlm", "oauth-client", "oauth-rop", "digest"}, false),
				},
				"username": {
					Description: "Username for authentication.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"password": {
					Description: "Password for authentication.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
				},
				"access_key": {
					Type:        schema.TypeString,
					Description: "Access key for `SIGV4` authentication.",
					Optional:    true,
					Sensitive:   true,
				},
				"secret_key": {
					Type:        schema.TypeString,
					Description: "Secret key for `SIGV4` authentication.",
					Optional:    true,
					Sensitive:   true,
				},
				"region": {
					Type:        schema.TypeString,
					Description: "Region for `SIGV4` authentication.",
					Optional:    true,
				},
				"service_name": {
					Type:        schema.TypeString,
					Description: "Service name for `SIGV4` authentication.",
					Optional:    true,
				},
				"session_token": {
					Type:        schema.TypeString,
					Description: "Session token for `SIGV4` authentication.",
					Optional:    true,
				},
				"domain": {
					Type:        schema.TypeString,
					Description: "Domain for `ntlm` authentication.",
					Optional:    true,
				},
				"workstation": {
					Type:        schema.TypeString,
					Description: "Workstation for `ntlm` authentication.",
					Optional:    true,
				},
				"access_token_url": {
					Type:        schema.TypeString,
					Description: "Access token url for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
				},
				"audience": {
					Type:        schema.TypeString,
					Description: "Audience for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Default:     "",
				},
				"resource": {
					Type:        schema.TypeString,
					Description: "Resource for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Default:     "",
				},
				"scope": {
					Type:        schema.TypeString,
					Description: "Scope for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Default:     "",
				},
				"token_api_authentication": {
					Type:             schema.TypeString,
					Description:      "Token API Authentication for `oauth-client` or `oauth-rop` authentication.",
					Optional:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsBasicAuthOauthTokenApiAuthenticationFromValue),
				},
				"client_id": {
					Type:        schema.TypeString,
					Description: "Client ID for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
				},
				"client_secret": {
					Type:        schema.TypeString,
					Description: "Client secret for `oauth-client` or `oauth-rop` authentication.",
					Optional:    true,
					Sensitive:   true,
				},
			},
		},
	}
}

func syntheticsTestRequestProxy() *schema.Schema {
	return &schema.Schema{
		Description: "The proxy to perform the test.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Type:        schema.TypeString,
					Description: "URL of the proxy to perform the test.",
					Required:    true,
				},
				"headers": syntheticsTestRequestHeaders(),
			},
		},
	}
}

func syntheticsTestRequestClientCertificate() *schema.Schema {
	return &schema.Schema{
		Description: "Client certificate to use when performing the test request. Exactly one nested block is allowed with the structure below.",
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cert": syntheticsTestRequestClientCertificateItem(),
				"key":  syntheticsTestRequestClientCertificateItem(),
			},
		},
	}
}

func syntheticsTestRequestClientCertificateItem() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content": {
					Description: "Content of the certificate.",
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					StateFunc: func(val interface{}) string {
						return utils.ConvertToSha256(val.(string))
					},
				},
				"filename": {
					Description: "File name for the certificate.",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "Provided in Terraform config",
				},
			},
		},
	}
}

func syntheticsTestRequestMetadata() *schema.Schema {
	return &schema.Schema{
		Description: "Metadata to include when performing the gRPC request.",
		Type:        schema.TypeMap,
		Optional:    true,
	}
}

func syntheticsAPIAssertion() *schema.Schema {
	return &schema.Schema{
		Description: "Assertions used for the test. Multiple `assertion` blocks are allowed with the structure below.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Description:      "Type of assertion. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
					Type:             schema.TypeString,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsAssertionTypeFromValue, datadogV1.NewSyntheticsAssertionBodyHashTypeFromValue, datadogV1.NewSyntheticsAssertionJavascriptTypeFromValue),
					Required:         true,
				},
				"operator": {
					Description:  "Assertion operator. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateSyntheticsAssertionOperator,
				},
				"property": {
					Description: "If assertion type is `header`, this is the header name.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"target": {
					Description: "Expected value. Depends on the assertion type, refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test) for details.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"code": {
					Description: "If assertion type is `javascript`, this is the JavaScript code that performs the assertions.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"targetjsonschema": {
					Description: "Expected structure if `operator` is `validatesJSONSchema`. Exactly one nested block is allowed with the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"jsonschema": {
								Description: "The JSON Schema to validate the body against.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"metaschema": {
								Description: "The meta schema to use for the JSON Schema.",
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "draft-07",
							},
						},
					},
				},
				"targetjsonpath": {
					Description: "Expected structure if `operator` is `validatesJSONPath`. Exactly one nested block is allowed with the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"elementsoperator": {
								Description: "The element from the list of results to assert on. Select from `firstElementMatches` (the first element in the list), `everyElementMatches` (every element in the list), `atLeastOneElementMatches` (at least one element in the list), or `serializationMatches` (the serialized value of the list). Defaults to `firstElementMatches`.",
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "firstElementMatches",
							},
							"operator": {
								Description: "The specific operator to use on the path.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"jsonpath": {
								Description: "The JSON path to assert.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"targetvalue": {
								Description: "Expected matching value.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
				"targetxpath": {
					Description: "Expected structure if `operator` is `validatesXPath`. Exactly one nested block is allowed with the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"operator": {
								Description: "The specific operator to use on the path.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"xpath": {
								Description: "The xpath to assert.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"targetvalue": {
								Description: "Expected matching value.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
				"timings_scope": {
					Description:      "Timings scope for response time assertions.",
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsAssertionTimingsScopeFromValue),
				},
			},
		},
	}
}

func syntheticsTestOptionsRetry() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"count": {
					Description: "Number of retries needed to consider a location as failed before sending a notification alert. Maximum value: `5`.",
					Type:        schema.TypeInt,
					Default:     0,
					Optional:    true,
				},
				"interval": {
					Description: "Interval between a failed test and the next retry in milliseconds. Maximum value: `5000`.",
					Type:        schema.TypeInt,
					Default:     300,
					Optional:    true,
				},
			},
		},
	}
}

func syntheticsTestAdvancedSchedulingTimeframes() *schema.Schema {
	return &schema.Schema{
		Description: "Array containing objects describing the scheduling pattern to apply to each day.",
		Type:        schema.TypeSet,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"day": {
					Description:  "Number representing the day of the week",
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(1, 7),
				},
				"from": {
					Description: "The hour of the day on which scheduling starts.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"to": {
					Description: "The hour of the day on which scheduling ends.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
	}
}

func syntheticsTestAdvancedScheduling() *schema.Schema {
	return &schema.Schema{
		Description: "Object containing timeframes and timezone used for advanced scheduling.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"timeframes": syntheticsTestAdvancedSchedulingTimeframes(),
				"timezone": {
					Description: "Timezone in which the timeframe is based.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
	}
}

func syntheticsTestOptionsList() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allow_insecure":   syntheticsAllowInsecureOption(),
				"follow_redirects": syntheticsFollowRedirectsOption(),
				"tick_every": {
					Description:  "How often the test should run (in seconds).",
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(30, 604800),
				},
				"scheduling": syntheticsTestAdvancedScheduling(),
				"accept_self_signed": {
					Description: "For SSL test, whether or not the test should allow self signed certificates.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"min_location_failed": {
					Description: "Minimum number of locations in failure required to trigger an alert.",
					Type:        schema.TypeInt,
					Default:     1,
					Optional:    true,
				},
				"min_failure_duration": {
					Description: "Minimum amount of time in failure required to trigger an alert (in seconds). Default is `0`.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"monitor_name": {
					Description: "The monitor name is used for the alert title as well as for all monitor dashboard widgets and SLOs.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"monitor_options": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"renotify_interval": {
								Description: "Specify a renotification frequency in minutes. Values available by default are `0`, `10`, `20`, `30`, `40`, `50`, `60`, `90`, `120`, `180`, `240`, `300`, `360`, `720`, `1440`.",
								Type:        schema.TypeInt,
								Default:     0,
								Optional:    true,
							},
						},
					},
				},
				"monitor_priority": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(1, 5),
				},
				"restricted_roles": {
					Description: "A list of role identifiers pulled from the Roles API to restrict read and write access.",
					Type:        schema.TypeSet,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"retry": syntheticsTestOptionsRetry(),
				"no_screenshot": {
					Description: "Prevents saving screenshots of the steps.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"check_certificate_revocation": {
					Description: "For SSL test, whether or not the test should fail on revoked certificate in stapled OCSP.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"ci": {
					Description: "CI/CD options for a Synthetic test.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"execution_rule": {
								Type:             schema.TypeString,
								Description:      "Execution rule for a Synthetics test.",
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestExecutionRuleFromValue),
								Optional:         true,
							},
						},
					},
				},
				"rum_settings": {
					Description: "The RUM data collection settings for the Synthetic browser test.",
					Type:        schema.TypeList,
					MaxItems:    1,
					DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
						if strings.Contains(key, "is_enabled") {
							if new != "true" && old != "true" {
								return true
							}
						} else {
							if rum_settings, ok := d.GetOk("options_list.0.rum_settings.0"); ok {
								settings := rum_settings.(map[string]interface{})
								isEnabled := settings["is_enabled"]

								if !isEnabled.(bool) {
									return true
								}
							}
						}
						return false
					},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"is_enabled": {
								Type:        schema.TypeBool,
								Description: "Determines whether RUM data is collected during test runs.",
								Required:    true,
							},
							"application_id": {
								Type:        schema.TypeString,
								Description: "RUM application ID used to collect RUM data for the browser test.",
								Optional:    true,
							},
							"client_token_id": {
								Type:        schema.TypeInt,
								Description: "RUM application API key ID used to collect RUM data for the browser test.",
								Sensitive:   true,
								Optional:    true,
							},
						},
					},
					Optional: true,
				},
				"ignore_server_certificate_error": {
					Description: "Ignore server certificate error for browser tests.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"disable_csp": {
					Description: "Disable Content Security Policy for browser tests.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"disable_cors": {
					Description: "Disable Cross-Origin Resource Sharing for browser tests.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"initial_navigation_timeout": {
					Description: "Timeout before declaring the initial step as failed (in seconds) for browser tests.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"http_version": syntheticsHttpVersionOption(),
			},
		},
	}
}

func syntheticsMobileTestOptionsList() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"min_failure_duration": {
					Description: "Minimum amount of time in failure required to trigger an alert (in seconds). Default is `0`.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"retry": syntheticsTestOptionsRetry(),
				"tick_every": {
					Description:  "How often the test should run (in seconds).",
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(300, 604800),
				},
				"scheduling": syntheticsTestAdvancedScheduling(),
				"monitor_name": {
					Description: "The monitor name is used for the alert title as well as for all monitor dashboard widgets and SLOs.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"monitor_options": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"renotify_interval": {
								Description: "Specify a renotification frequency in minutes. Values available by default are `0`, `10`, `20`, `30`, `40`, `50`, `60`, `90`, `120`, `180`, `240`, `300`, `360`, `720`, `1440`.",
								Type:        schema.TypeInt,
								Default:     0,
								Optional:    true,
							},
							"escalation_message": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"renotify_occurrences": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"notification_preset_name": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestOptionsMonitorOptionsNotificationPresetNameFromValue),
							},
						},
					},
				},
				"monitor_priority": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(1, 5),
				},
				"restricted_roles": {
					Description: "A list of role identifiers pulled from the Roles API to restrict read and write access.",
					Type:        schema.TypeSet,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"bindings": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"principals": {
								Type:     schema.TypeList,
								Optional: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"relation": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestRestrictionPolicyBindingRelationFromValue),
							},
						},
					},
				},
				"ci": {
					Description: "CI/CD options for a Synthetic test.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"execution_rule": {
								Type:             schema.TypeString,
								Description:      "Execution rule for a Synthetics test.",
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestExecutionRuleFromValue),
							},
						},
					},
				},
				"default_step_timeout": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(1, 300),
				},
				"device_ids": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
					},
				},
				"no_screenshot": {
					Description: "Prevents saving screenshots of the steps.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"verbosity": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(0, 5),
				},
				"allow_application_crash": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"disable_auto_accept_alert": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"mobile_application": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"application_id": {
								Type:     schema.TypeString,
								Required: true,
							},
							"reference_id": {
								Type:     schema.TypeString,
								Required: true,
							},
							"reference_type": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsMobileTestsMobileApplicationReferenceTypeFromValue),
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsTestAPIStep() *schema.Schema {
	requestElemSchema := syntheticsTestRequest()
	// In test `options_list` for single API tests, but in `api_step.request_definition` for API steps.
	requestElemSchema.Schema["allow_insecure"] = syntheticsAllowInsecureOption()
	requestElemSchema.Schema["follow_redirects"] = syntheticsFollowRedirectsOption()
	requestElemSchema.Schema["http_version"] = syntheticsHttpVersionOption()

	return &schema.Schema{
		Description: "Steps for multi-step api tests",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the step.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"subtype": {
					Description:      "The subtype of the Synthetic multi-step API test step.",
					Type:             schema.TypeString,
					Optional:         true,
					Default:          "http",
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsAPITestStepSubtypeFromValue, datadogV1.NewSyntheticsAPIWaitStepSubtypeFromValue),
				},
				"exit_if_succeed": {
					Description: "Determines whether or not to exit the test if the step succeeds.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"extracted_value": {
					Description: "Values to parse and save as variables from the response.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:     schema.TypeString,
								Required: true,
							},
							"type": {
								Description:      "Property of the Synthetics Test Response to use for the variable.",
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsLocalVariableParsingOptionsTypeFromValue),
							},
							"field": {
								Description: "When type is `http_header` or `grpc_metadata`, name of the header or metadatum to extract.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"parser": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"type": {
											Description:      "Type of parser for a Synthetics global variable from a synthetics test.",
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsGlobalVariableParserTypeFromValue),
										},
										"value": {
											Type:        schema.TypeString,
											Description: "Regex or JSON path used for the parser. Not used with type `raw`.",
											Optional:    true,
										},
									},
								},
							},
							"secure": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Determines whether or not the extracted value will be obfuscated.",
							},
						},
					},
				},
				"request_definition": {
					Description: "The request for the api step.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem:        requestElemSchema,
				},
				"request_headers":            syntheticsTestRequestHeaders(),
				"request_query":              syntheticsTestRequestQuery(),
				"request_basicauth":          syntheticsTestRequestBasicAuth(),
				"request_proxy":              syntheticsTestRequestProxy(),
				"request_client_certificate": syntheticsTestRequestClientCertificate(),
				"request_file":               syntheticsTestRequestFile(),
				"request_metadata":           syntheticsTestRequestMetadata(),
				"assertion":                  syntheticsAPIAssertion(),
				"allow_failure": {
					Description: "Determines whether or not to continue with test if this step fails.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"is_critical": {
					Description: "Determines whether or not to consider the entire test as failed if this step fails. Can be used only if `allow_failure` is `true`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"retry": syntheticsTestOptionsRetry(),
				"value": {
					Description: "The time to wait in seconds. Minimum value: 0. Maximum value: 180.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
			},
		},
	}
}

func syntheticsTestRequestFile() *schema.Schema {
	requestFilesSchema := schema.Schema{
		Description: "Files to be used as part of the request in the test.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"content": {
					Type:         schema.TypeString,
					Description:  "Content of the file.",
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 3145728),
				},
				"bucket_key": {
					Type:        schema.TypeString,
					Description: "Bucket key of the file.",
					Computed:    true,
				},
				"name": {
					Type:         schema.TypeString,
					Description:  "Name of the file.",
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 1500),
				},
				"original_file_name": {
					Type:         schema.TypeString,
					Description:  "Original name of the file.",
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 1500),
				},
				"size": {
					Type:         schema.TypeInt,
					Description:  "Size of the file.",
					Required:     true,
					ValidateFunc: validation.IntBetween(1, 3145728),
				},
				"type": {
					Type:         schema.TypeString,
					Description:  "Type of the file.",
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 1500),
				},
			},
		},
	}

	return &requestFilesSchema
}

func syntheticsTestBrowserStep() *schema.Schema {
	paramsSchema := syntheticsBrowserStepParams()
	browserStepSchema := schema.Schema{
		Description: "Steps for browser tests.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "Name of the step.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"local_key": {
					Description: "A unique identifier used to track steps after reordering.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"public_id": {
					Description: "The identifier of the step on the backend.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"type": {
					Description:      "Type of the step.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsStepTypeFromValue),
				},
				"allow_failure": {
					Description: "Determines if the step should be allowed to fail.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"always_execute": {
					Description: "Determines whether or not to always execute this step even if the previous step failed or was skipped.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"exit_if_succeed": {
					Description: "Determines whether or not to exit the test if the step succeeds.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"is_critical": {
					Description: "Determines whether or not to consider the entire test as failed if this step fails. Can be used only if `allow_failure` is `true`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"timeout": {
					Description: "Used to override the default timeout of a step.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"params": &paramsSchema,
				"force_element_update": {
					Description: `Force update of the "element" parameter for the step`,
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"no_screenshot": {
					Description: "Prevents saving screenshots of the step.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
			},
		},
	}

	return &browserStepSchema
}

func syntheticsBrowserStepParams() schema.Schema {
	return schema.Schema{
		Description: "Parameters for the step.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"attribute": {
					Description: `Name of the attribute to use for an "assert attribute" step.`,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"check": {
					Description:      "Check type to use for an assertion step.",
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsCheckTypeFromValue),
				},
				"click_type": {
					Description:  `Type of click to use for a "click" step.`,
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice([]string{"contextual", "double", "primary"}, false),
				},
				"code": {
					Description: "Javascript code to use for the step.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"delay": {
					Description: `Delay between each key stroke for a "type test" step.`,
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"element": {
					Description: "Element to use for the step, JSON encoded string.",
					Type:        schema.TypeString,
					Optional:    true,
					DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
						// if there is no old value we let TF handle this
						if old == "" {
							return old == new
						}

						forceElementUpdateKey := strings.Replace(key, "params.0.element", "force_element_update", 1)

						// if the field force_element_update is present we force the update
						// of the step params
						if attr, ok := d.GetOk(forceElementUpdateKey); ok && attr.(bool) {
							return false
						}

						// by default we ignore the diff for step parameters because some of them
						// are updated by the backend.
						return true
					},
				},
				"element_user_locator": {
					Description: "Custom user selector to use for the step.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"fail_test_on_cannot_locate": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
							"value": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"type": {
											Type:         schema.TypeString,
											Optional:     true,
											Default:      "css",
											ValidateFunc: validation.StringInSlice([]string{"css", "xpath"}, false),
										},
										"value": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
						},
					},
				},
				"email": {
					Description: `Details of the email for an "assert email" step, JSON encoded string.`,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"file": {
					Description: `JSON encoded string used for an "assert download" step. Refer to the examples for a usage example showing the schema.`,
					Type:        schema.TypeString,
					Optional:    true,
					DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
						return strings.TrimSpace(old) == strings.TrimSpace(new)
					},
				},
				"files": {
					Description: `Details of the files for an "upload files" step, JSON encoded string.`,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"modifiers": {
					Description: `Modifier to use for a "press key" step.`,
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice([]string{"Alt", "Control", "meta", "Shift"}, false),
					},
				},
				"playing_tab_id": {
					Description: "ID of the tab to play the subtest.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"request": {
					Description: "Request for an API step.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"subtest_public_id": {
					Description: "ID of the Synthetics test to use as subtest.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"value": {
					Description: "Value of the step.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"variable": {
					Description: "Details of the variable to extract.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Description: "Name of the extracted variable.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"example": {
								Description: "Example of the extracted variable.",
								Default:     "",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
					Optional: true,
				},
				"with_click": {
					Description: `For "file upload" steps.`,
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"x": {
					Description: `X coordinates for a "scroll step".`,
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"y": {
					Description: `Y coordinates for a "scroll step".`,
					Type:        schema.TypeInt,
					Optional:    true,
				},
			},
		},
	}
}

func syntheticsTestMobileStep() *schema.Schema {
	paramsSchema := syntheticsMobileStepParams()
	return &schema.Schema{
		Description: "Steps for mobile tests",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allow_failure": {
					Description: "A boolean set to allow this step to fail.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"has_new_step_element": {
					Description: "A boolean set to determine if the step has a new step element.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"is_critical": {
					Description: "A boolean to use in addition to `allowFailure` to determine if the test should be marked as failed when the step fails.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"name": {
					Description: "The name of the step.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"no_screenshot": {
					Description: "A boolean set to not take a screenshot for the step.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"params": &paramsSchema,
				"public_id": {
					Description: "The public ID of the step.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"timeout": {
					Description: "The time before declaring a step failed.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"type": {
					Description:      "The type of the step.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsMobileStepTypeFromValue),
				},
			},
		},
	}
}

func syntheticsMobileStepParams() schema.Schema {
	return schema.Schema{
		Description: "Parameters for the step.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"value": {
					Description: "Value of the step.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"check": {
					Description:      "Check type to use for an assertion step.",
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsCheckTypeFromValue),
				},
				"element": {
					Description: "Element to use for the step, JSON encoded string.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"multi_locator": {
								Type:     schema.TypeMap,
								Optional: true,
							},
							"context": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"context_type": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsMobileStepParamsElementContextTypeFromValue),
							},
							"user_locator": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"fail_test_on_cannot_locate": {
											Type:     schema.TypeBool,
											Optional: true,
										},
										"values": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 5,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"type": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsMobileStepParamsElementUserLocatorValuesItemsTypeFromValue),
													},
													"value": {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"element_description": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"relative_position": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"x": {
											Type:     schema.TypeFloat,
											Optional: true,
										},
										"y": {
											Type:     schema.TypeFloat,
											Optional: true,
										},
									},
								},
							},
							"text_content": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"view_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"variable": {
					Description: "Details of the variable to extract.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Description: "Name of the extracted variable.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"example": {
								Description: "Example of the extracted variable.",
								Default:     "",
								Type:        schema.TypeString,
								// Required:    true, // TODO SYNTH-17172 - fix for steps, the tests don't like this being required for some reason
								Optional: true,
							},
						},
					},
				},
				"positions": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"x": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"y": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
						},
					},
				},
				"subtest_public_id": {
					Description: "ID of the Synthetics test to use as subtest.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"x": {
					Description: `X coordinates for a "scroll step".`,
					Type:        schema.TypeFloat,
					Optional:    true,
				},
				"y": {
					Description: `Y coordinates for a "scroll step".`,
					Type:        schema.TypeFloat,
					Optional:    true,
				},
				"direction": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsMobileStepParamsDirectionFromValue),
				},
				"max_scrolls": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"enable": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"delay": {
					Description: `Delay between each key stroke for a "type test" step.`,
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"with_enter": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
	}
}

func syntheticsBrowserVariable() *schema.Schema {
	return &schema.Schema{
		Description: "Variables used for a browser test steps. Multiple `variable` blocks are allowed with the structure below.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem:        syntheticsBrowserVariableElem(),
	}
}

func syntheticsBrowserVariableElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"example": {
				Description: "Example for the variable.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"id": {
				Description: "ID of the global variable to use. This is actually only used (and required) in the case of using a variable of type `global`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description:  "Name of the variable.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
			},
			"pattern": {
				Description: "Pattern of the variable.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"type": {
				Description:      "Type of browser test variable.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsBrowserVariableTypeFromValue),
			},
			"secure": {
				Description: "Determines whether or not the browser test variable is obfuscated. Can only be used with a browser variable of type `text`",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func syntheticsConfigVariable() *schema.Schema {
	return &schema.Schema{
		Description: "Variables used for the test configuration. Multiple `config_variable` blocks are allowed with the structure below.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"example": {
					Description: "Example for the variable. This value is not returned by the api when `secure = true`. Avoid drift by only making updates to this value from within Terraform.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"name": {
					Description:  "Name of the variable.",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
				},
				"pattern": {
					Description: "Pattern of the variable. This value is not returned by the api when `secure = true`. Avoid drift by only making updates to this value from within Terraform.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"type": {
					Description:      "Type of test configuration variable.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsConfigVariableTypeFromValue),
				},
				"id": {
					Description: "When type = `global`, ID of the global variable to use.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"secure": {
					Description: "Whether the value of this variable will be obfuscated in test results.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
			},
		},
	}
}

func syntheticsAllowInsecureOption() *schema.Schema {
	return &schema.Schema{
		Description: "Allows loading insecure content for a request in an API test or in a multistep API test step.",
		Type:        schema.TypeBool,
		Optional:    true,
	}
}

func syntheticsFollowRedirectsOption() *schema.Schema {
	return &schema.Schema{
		Description: "Determines whether or not the API HTTP test should follow redirects.",
		Type:        schema.TypeBool,
		Optional:    true,
	}
}

func syntheticsHttpVersionOption() *schema.Schema {
	return &schema.Schema{
		Description:      "HTTP version to use for an HTTP request in an API test or step.",
		Default:          datadogV1.SYNTHETICSTESTOPTIONSHTTPVERSION_ANY,
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestOptionsHTTPVersionFromValue),
	}
}

/*
 * CRUD functions
 */

func resourceDatadogSyntheticsTestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	testType := getSyntheticsTestType(d)

	if *testType == datadogV1.SYNTHETICSTESTDETAILSTYPE_API {
		syntheticsTest := buildDatadogSyntheticsAPITest(d)
		createdSyntheticsTest, httpResponseCreate, err := apiInstances.GetSyntheticsApiV1().CreateSyntheticsAPITest(auth, *syntheticsTest)
		if err != nil {
			// Note that Id won't be set, so no state will be saved.
			return utils.TranslateClientErrorDiag(err, httpResponseCreate, "error creating synthetics API test")
		}
		if err := utils.CheckForUnparsed(createdSyntheticsTest); err != nil {
			return diag.FromErr(err)
		}

		var getSyntheticsApiTestResponse datadogV1.SyntheticsAPITest
		var httpResponseGet *_nethttp.Response
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			getSyntheticsApiTestResponse, httpResponseGet, err = apiInstances.GetSyntheticsApiV1().GetAPITest(auth, createdSyntheticsTest.GetPublicId())
			if err != nil {
				if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
					return retry.RetryableError(fmt.Errorf("synthetics api test not created yet"))
				}

				return retry.NonRetryableError(err)
			}
			if err := utils.CheckForUnparsed(getSyntheticsApiTestResponse); err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getSyntheticsApiTestResponse.GetPublicId())

		return updateSyntheticsAPITestLocalState(d, &getSyntheticsApiTestResponse)
	} else if *testType == datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER {
		syntheticsTest := buildDatadogSyntheticsBrowserTest(d)
		createdSyntheticsTest, httpResponse, err := apiInstances.GetSyntheticsApiV1().CreateSyntheticsBrowserTest(auth, *syntheticsTest)
		if err != nil {
			// Note that Id won't be set, so no state will be saved.
			return utils.TranslateClientErrorDiag(err, httpResponse, "error creating synthetics browser test")
		}
		if err := utils.CheckForUnparsed(createdSyntheticsTest); err != nil {
			return diag.FromErr(err)
		}

		var getSyntheticsBrowserTestResponse datadogV1.SyntheticsBrowserTest
		var httpResponseGet *_nethttp.Response
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			getSyntheticsBrowserTestResponse, httpResponseGet, err = apiInstances.GetSyntheticsApiV1().GetBrowserTest(auth, createdSyntheticsTest.GetPublicId())
			if err != nil {
				if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
					return retry.RetryableError(fmt.Errorf("synthetics browser test not created yet"))
				}

				return retry.NonRetryableError(err)
			}
			if err := utils.CheckForUnparsed(getSyntheticsBrowserTestResponse); err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getSyntheticsBrowserTestResponse.GetPublicId())

		return updateSyntheticsBrowserTestLocalState(d, &getSyntheticsBrowserTestResponse)
	} else if *testType == datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE {
		syntheticsTest := buildDatadogSyntheticsMobileTest(d)
		createdSyntheticsTest, httpResponse, err := apiInstances.GetSyntheticsApiV1().CreateSyntheticsMobileTest(auth, *syntheticsTest)
		if err != nil {
			// Note that Id won't be set, so no state will be saved.
			return utils.TranslateClientErrorDiag(err, httpResponse, "error creating synthetics mobile test")
		}
		if err := utils.CheckForUnparsed(createdSyntheticsTest); err != nil {
			return diag.FromErr(err)
		}

		var getSyntheticsMobileTestResponse datadogV1.SyntheticsMobileTest
		var httpResponseGet *_nethttp.Response
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
			getSyntheticsMobileTestResponse, httpResponseGet, err = apiInstances.GetSyntheticsApiV1().GetMobileTest(auth, createdSyntheticsTest.GetPublicId())
			if err != nil {
				if httpResponseGet != nil && httpResponseGet.StatusCode == 404 {
					return retry.RetryableError(fmt.Errorf("synthetics mobile test not created yet"))
				}

				return retry.NonRetryableError(err)
			}
			if err := utils.CheckForUnparsed(getSyntheticsMobileTestResponse); err != nil {
				return retry.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(getSyntheticsMobileTestResponse.GetPublicId())

		return updateSyntheticsMobileTestLocalState(d, &getSyntheticsMobileTestResponse)

	}

	return diag.Errorf("unrecognized synthetics test type %v", testType)
}

func resourceDatadogSyntheticsTestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	var syntheticsTest datadogV1.SyntheticsTestDetails
	var syntheticsAPITest datadogV1.SyntheticsAPITest
	var syntheticsBrowserTest datadogV1.SyntheticsBrowserTest
	var syntheticsMobileTest datadogV1.SyntheticsMobileTest
	var err error
	var httpresp *_nethttp.Response

	// get the generic test to detect if it's an api or browser test
	syntheticsTest, httpresp, err = apiInstances.GetSyntheticsApiV1().GetTest(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics test")
	}
	if err := utils.CheckForUnparsed(syntheticsTest); err != nil {
		return diag.FromErr(err)
	}

	if syntheticsTest.GetType() == datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER {
		syntheticsBrowserTest, _, err = apiInstances.GetSyntheticsApiV1().GetBrowserTest(auth, d.Id())
	} else if syntheticsTest.GetType() == datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE {
		syntheticsMobileTest, _, err = apiInstances.GetSyntheticsApiV1().GetMobileTest(auth, d.Id())
	} else {
		syntheticsAPITest, _, err = apiInstances.GetSyntheticsApiV1().GetAPITest(auth, d.Id())
	}

	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting synthetics test")
	}

	if syntheticsTest.GetType() == datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER {
		if err := utils.CheckForUnparsed(syntheticsBrowserTest); err != nil {
			return diag.FromErr(err)
		}
		return updateSyntheticsBrowserTestLocalState(d, &syntheticsBrowserTest)
	}

	if syntheticsTest.GetType() == datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE {
		if err := utils.CheckForUnparsed(syntheticsMobileTest); err != nil {
			return diag.FromErr(err)
		}
		return updateSyntheticsMobileTestLocalState(d, &syntheticsMobileTest)
	}

	if err := utils.CheckForUnparsed(syntheticsAPITest); err != nil {
		return diag.FromErr(err)
	}
	return updateSyntheticsAPITestLocalState(d, &syntheticsAPITest)
}

func resourceDatadogSyntheticsTestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	testType := getSyntheticsTestType(d)

	if *testType == datadogV1.SYNTHETICSTESTDETAILSTYPE_API {
		syntheticsTest := buildDatadogSyntheticsAPITest(d)
		updatedTest, httpResponse, err := apiInstances.GetSyntheticsApiV1().UpdateAPITest(auth, d.Id(), *syntheticsTest)
		if err != nil {
			// If the Update callback returns with or without an error, the full state is saved.
			return utils.TranslateClientErrorDiag(err, httpResponse, "error updating synthetics API test")
		}
		if err := utils.CheckForUnparsed(updatedTest); err != nil {
			return diag.FromErr(err)
		}
		return updateSyntheticsAPITestLocalState(d, &updatedTest)
	} else if *testType == datadogV1.SYNTHETICSTESTDETAILSTYPE_BROWSER {
		syntheticsTest := buildDatadogSyntheticsBrowserTest(d)
		updatedTest, httpResponse, err := apiInstances.GetSyntheticsApiV1().UpdateBrowserTest(auth, d.Id(), *syntheticsTest)
		if err != nil {
			// If the Update callback returns with or without an error, the full state is saved.
			return utils.TranslateClientErrorDiag(err, httpResponse, "error updating synthetics browser test")
		}
		if err := utils.CheckForUnparsed(updatedTest); err != nil {
			return diag.FromErr(err)
		}
		return updateSyntheticsBrowserTestLocalState(d, &updatedTest)
	} else if *testType == datadogV1.SYNTHETICSTESTDETAILSTYPE_MOBILE {
		syntheticsTest := buildDatadogSyntheticsMobileTest(d)
		updatedTest, httpResponse, err := apiInstances.GetSyntheticsApiV1().UpdateMobileTest(auth, d.Id(), *syntheticsTest)
		if err != nil {
			// If the Update callback returns with or without an error, the full state is saved.
			return utils.TranslateClientErrorDiag(err, httpResponse, "error updating synthetics browser test")
		}
		if err := utils.CheckForUnparsed(updatedTest); err != nil {
			return diag.FromErr(err)
		}

		return updateSyntheticsMobileTestLocalState(d, &updatedTest)
	}

	return diag.Errorf("unrecognized synthetics test type %v", testType)
}

func resourceDatadogSyntheticsTestDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	syntheticsDeleteTestsPayload := datadogV1.SyntheticsDeleteTestsPayload{PublicIds: []string{d.Id()}}
	if d.Get("force_delete_dependencies").(bool) {
		syntheticsDeleteTestsPayload.SetForceDeleteDependencies(true)
	}

	if _, httpResponse, err := apiInstances.GetSyntheticsApiV1().DeleteTests(auth, syntheticsDeleteTestsPayload); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting synthetics test")
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func updateSyntheticsBrowserTestLocalState(d *schema.ResourceData, syntheticsTest *datadogV1.SyntheticsBrowserTest) diag.Diagnostics {
	if err := d.Set("type", syntheticsTest.GetType()); err != nil {
		return diag.FromErr(err)
	}

	config := syntheticsTest.GetConfig()
	actualRequest := config.GetRequest()
	localRequest := buildTerraformTestRequest(actualRequest)

	if config.HasSetCookie() {
		if err := d.Set("set_cookie", config.GetSetCookie()); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("request_definition", []map[string]interface{}{localRequest}); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("request_headers", actualRequest.Headers); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("request_query", actualRequest.GetQuery()); err != nil {
		return diag.FromErr(err)
	}

	if basicAuth, ok := actualRequest.GetBasicAuthOk(); ok && basicAuth.SyntheticsBasicAuthWeb != nil {
		localAuth := buildTerraformBasicAuth(basicAuth)

		if err := d.Set("request_basicauth", []map[string]string{localAuth}); err != nil {
			return diag.FromErr(err)
		}
	}

	if clientCertificate, ok := actualRequest.GetCertificateOk(); ok {
		oldCertificates := d.Get("request_client_certificate").([]interface{})
		localCertificate := buildTerraformRequestCertificates(*clientCertificate, oldCertificates)

		if err := d.Set("request_client_certificate", []map[string][]map[string]string{localCertificate}); err != nil {
			return diag.FromErr(err)
		}
	}

	if proxy, ok := actualRequest.GetProxyOk(); ok {
		localProxy := buildTerraformTestRequestProxy(*proxy)
		if err := d.Set("request_proxy", []map[string]interface{}{localProxy}); err != nil {
			return diag.FromErr(err)
		}
	}

	// assertions are required but not used for browser tests
	localAssertions := make([]map[string]interface{}, 0)

	if err := d.Set("assertion", localAssertions); err != nil {
		return diag.FromErr(err)
	}

	configVariables := config.GetConfigVariables()
	oldConfigVariables := d.Get("config_variable").([]interface{})
	if err := d.Set("config_variable", buildTerraformConfigVariables(configVariables, oldConfigVariables)); err != nil {
		return diag.FromErr(err)
	}

	actualVariables := config.GetVariables()
	localBrowserVariables := make([]map[string]interface{}, len(actualVariables))
	for i, variable := range actualVariables {
		localVariable := make(map[string]interface{})
		if v, ok := variable.GetTypeOk(); ok {
			localVariable["type"] = *v
		}
		if v, ok := variable.GetNameOk(); ok {
			localVariable["name"] = *v
		}
		if v, ok := variable.GetIdOk(); ok {
			localVariable["id"] = *v
		}
		if v, ok := variable.GetSecureOk(); ok {
			localVariable["secure"] = *v
		}
		if v, ok := variable.GetExampleOk(); ok {
			localVariable["example"] = *v
		} else if v, ok := localVariable["secure"].(bool); ok && v {
			localVariable["example"] = d.Get(fmt.Sprintf("browser_variable.%d.example", i))
		}
		if v, ok := variable.GetPatternOk(); ok {
			localVariable["pattern"] = *v
		} else if v, ok := localVariable["secure"].(bool); ok && v {
			localVariable["pattern"] = d.Get(fmt.Sprintf("browser_variable.%d.pattern", i))
		}
		localBrowserVariables[i] = localVariable
	}

	if err := d.Set("browser_variable", localBrowserVariables); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("device_ids", syntheticsTest.GetOptions().DeviceIds); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("locations", syntheticsTest.Locations); err != nil {
		return diag.FromErr(err)
	}

	localOptionsLists := buildTerraformTestOptions(syntheticsTest.GetOptions())

	if err := d.Set("options_list", localOptionsLists); err != nil {
		return diag.FromErr(err)
	}

	steps := syntheticsTest.GetSteps()
	var localSteps []map[string]interface{}

	for stepIndex, step := range steps {
		localStep := make(map[string]interface{})
		localStep["name"] = step.GetName()
		localStep["public_id"] = step.GetPublicId()
		localStep["type"] = string(step.GetType())
		localStep["timeout"] = step.GetTimeout()

		if allowFailure, ok := step.GetAllowFailureOk(); ok {
			localStep["allow_failure"] = allowFailure
		}
		if alwaysExecute, ok := step.GetAlwaysExecuteOk(); ok {
			localStep["always_execute"] = alwaysExecute
		}
		if exitIfSucceed, ok := step.GetExitIfSucceedOk(); ok {
			localStep["exit_if_succeed"] = exitIfSucceed
		}
		if isCritical, ok := step.GetIsCriticalOk(); ok {
			localStep["is_critical"] = isCritical
		}
		if hasNoScreenshot, ok := step.GetNoScreenshotOk(); ok {
			localStep["no_screenshot"] = hasNoScreenshot
		}

		localParams := make(map[string]interface{})

		forceElementUpdate, ok := d.GetOk(fmt.Sprintf("browser_step.%d.force_element_update", stepIndex))
		if ok {
			localStep["force_element_update"] = forceElementUpdate
		}

		localKey, ok := d.GetOk(fmt.Sprintf("browser_step.%d.local_key", stepIndex))
		if ok {
			localStep["local_key"] = localKey
		}
		publicId, ok := d.GetOk(fmt.Sprintf("browser_step.%d.public_id", stepIndex))
		if ok {
			localStep["public_id"] = publicId
		}

		params := step.GetParams()
		paramsMap := params.(map[string]interface{})

		for key, value := range paramsMap {
			if key == "element" && forceElementUpdate == true {
				// prevent overriding `element` in the local state with the one received from the backend, and
				// keep the element from the local state instead
				element := d.Get(fmt.Sprintf("browser_step.%d.params.0.element", stepIndex))
				localParams["element"] = element
			} else if key == "files" {
				// prevent overriding `files` in the local state with the one received from the backend, and
				// keep the files from the local state instead
				files := d.Get(fmt.Sprintf("browser_step.%d.params.0.files", stepIndex))
				localParams["files"] = files
			} else {
				localParams[convertStepParamsKey(key)] = convertStepParamsValueForState(convertStepParamsKey(key), value)
			}
		}

		// If received an element from the backend, extract the user locator part to update the local state
		if elementParams, ok := paramsMap["element"]; ok {
			serializedElementParams := convertStepParamsValueForState("element", elementParams)
			var stepElement interface{}
			utils.GetMetadataFromJSON([]byte(serializedElementParams.(string)), &stepElement)
			if elementUserLocator, ok := stepElement.(map[string]interface{})["userLocator"]; ok {
				userLocator := elementUserLocator.(map[string]interface{})
				values := userLocator["values"]
				value := values.([]interface{})[0]

				localElementUserLocator := map[string]interface{}{
					"fail_test_on_cannot_locate": userLocator["failTestOnCannotLocate"],
					"value": []map[string]interface{}{
						value.(map[string]interface{}),
					},
				}

				localParams["element_user_locator"] = []map[string]interface{}{localElementUserLocator}
			}
		}

		localStep["params"] = []interface{}{localParams}

		localSteps = append(localSteps, localStep)
	}

	if err := d.Set("browser_step", localSteps); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", syntheticsTest.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", syntheticsTest.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", syntheticsTest.GetStatus()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", syntheticsTest.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_id", syntheticsTest.MonitorId); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateSyntheticsAPITestLocalState(d *schema.ResourceData, syntheticsTest *datadogV1.SyntheticsAPITest) diag.Diagnostics {
	if err := d.Set("type", syntheticsTest.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if syntheticsTest.HasSubtype() {
		if err := d.Set("subtype", syntheticsTest.GetSubtype()); err != nil {
			return diag.FromErr(err)
		}
	}

	config := syntheticsTest.GetConfig()
	actualRequest := config.GetRequest()
	localRequest := buildTerraformTestRequest(actualRequest)

	if syntheticsTest.GetSubtype() != "multi" {
		if err := d.Set("request_definition", []map[string]interface{}{localRequest}); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("request_headers", actualRequest.Headers); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("request_query", actualRequest.GetQuery()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("request_metadata", actualRequest.GetMetadata()); err != nil {
		return diag.FromErr(err)
	}

	if basicAuth, ok := actualRequest.GetBasicAuthOk(); ok {
		localAuth := buildTerraformBasicAuth(basicAuth)

		if err := d.Set("request_basicauth", []map[string]string{localAuth}); err != nil {
			return diag.FromErr(err)
		}
	}

	if clientCertificate, ok := actualRequest.GetCertificateOk(); ok {
		oldCertificates := d.Get("request_client_certificate").([]interface{})
		localCertificate := buildTerraformRequestCertificates(*clientCertificate, oldCertificates)

		if err := d.Set("request_client_certificate", []map[string][]map[string]string{localCertificate}); err != nil {
			return diag.FromErr(err)
		}
	}

	if proxy, ok := actualRequest.GetProxyOk(); ok {
		localProxy := buildTerraformTestRequestProxy(*proxy)
		if err := d.Set("request_proxy", []map[string]interface{}{localProxy}); err != nil {
			return diag.FromErr(err)
		}
	}

	if files, ok := actualRequest.GetFilesOk(); ok && files != nil && len(*files) > 0 {
		oldLocalFilesCount := d.Get("request_file.#").(int)
		oldLocalFiles := make([]map[string]interface{}, oldLocalFilesCount)
		for i := 0; i < oldLocalFilesCount; i++ {
			oldLocalFile := d.Get(fmt.Sprintf("request_file.%d", i)).(map[string]interface{})
			oldLocalFiles[i] = oldLocalFile
		}

		localFiles := buildTerraformBodyFiles(files, oldLocalFiles)
		if err := d.Set("request_file", localFiles); err != nil {
			return diag.FromErr(err)
		}
	}

	actualAssertions := config.GetAssertions()
	localAssertions, err := buildTerraformAssertions(actualAssertions)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("assertion", localAssertions); err != nil {
		return diag.FromErr(err)
	}

	configVariables := config.GetConfigVariables()
	oldConfigVariables := d.Get("config_variable").([]interface{})
	if err := d.Set("config_variable", buildTerraformConfigVariables(configVariables, oldConfigVariables)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("variables_from_script", config.GetVariablesFromScript()); err != nil {
		return diag.FromErr(err)
	}

	if steps, ok := config.GetStepsOk(); ok {
		localSteps := make([]interface{}, len(*steps))

		for i, step := range *steps {
			localStep := make(map[string]interface{})

			if step.SyntheticsAPITestStep != nil {
				localStep["name"] = step.SyntheticsAPITestStep.GetName()
				localStep["subtype"] = step.SyntheticsAPITestStep.GetSubtype()

				localAssertions, err := buildTerraformAssertions(step.SyntheticsAPITestStep.GetAssertions())
				if err != nil {
					return diag.FromErr(err)
				}
				localStep["assertion"] = localAssertions
				localStep["extracted_value"] = buildTerraformExtractedValues(step.SyntheticsAPITestStep.GetExtractedValues())

				stepRequest := step.SyntheticsAPITestStep.GetRequest()
				localRequest := buildTerraformTestRequest(stepRequest)
				localRequest["allow_insecure"] = stepRequest.GetAllowInsecure()
				localRequest["follow_redirects"] = stepRequest.GetFollowRedirects()
				if step.SyntheticsAPITestStep.GetSubtype() == "grpc" {
					// the schema defines a default value of `http_version` for any kind of step,
					// but it's not supported for `grpc` - so we save `any` in the local state to avoid diffs
					localRequest["http_version"] = datadogV1.SYNTHETICSTESTOPTIONSHTTPVERSION_ANY
				}
				localStep["request_definition"] = []map[string]interface{}{localRequest}
				localStep["request_headers"] = stepRequest.GetHeaders()
				localStep["request_query"] = stepRequest.GetQuery()
				localStep["request_metadata"] = stepRequest.GetMetadata()

				if basicAuth, ok := stepRequest.GetBasicAuthOk(); ok {
					localAuth := buildTerraformBasicAuth(basicAuth)
					localStep["request_basicauth"] = []map[string]string{localAuth}
				}

				if clientCertificate, ok := stepRequest.GetCertificateOk(); ok {
					oldCertificates := d.Get(fmt.Sprintf("api_step.%d.request_client_certificate", i)).([]interface{})
					localCertificate := buildTerraformRequestCertificates(*clientCertificate, oldCertificates)

					localStep["request_client_certificate"] = []map[string][]map[string]string{localCertificate}
				}

				if proxy, ok := stepRequest.GetProxyOk(); ok {
					localProxy := buildTerraformTestRequestProxy(*proxy)
					localStep["request_proxy"] = []map[string]interface{}{localProxy}
				}

				if files, ok := stepRequest.GetFilesOk(); ok && files != nil && len(*files) > 0 {
					oldLocalFilesCount := d.Get(fmt.Sprintf("api_step.%d.request_file.#", i)).(int)
					oldLocalFiles := make([]map[string]interface{}, oldLocalFilesCount)
					for j := 0; j < oldLocalFilesCount; j++ {
						oldLocalFile := d.Get(fmt.Sprintf("api_step.%d.request_file.%d", i, j)).(map[string]interface{})
						oldLocalFiles[j] = oldLocalFile
					}

					localFiles := buildTerraformBodyFiles(files, oldLocalFiles)
					localStep["request_file"] = localFiles
				}

				localStep["allow_failure"] = step.SyntheticsAPITestStep.GetAllowFailure()
				localStep["exit_if_succeed"] = step.SyntheticsAPITestStep.GetExitIfSucceed()
				localStep["is_critical"] = step.SyntheticsAPITestStep.GetIsCritical()

				if retry, ok := step.SyntheticsAPITestStep.GetRetryOk(); ok {
					localRetry := make(map[string]interface{})
					if count, ok := retry.GetCountOk(); ok {
						localRetry["count"] = *count
					}
					if interval, ok := retry.GetIntervalOk(); ok {
						localRetry["interval"] = *interval
					}
					localStep["retry"] = []map[string]interface{}{localRetry}
				}
			} else if step.SyntheticsAPIWaitStep != nil {
				localStep["name"] = step.SyntheticsAPIWaitStep.GetName()
				localStep["subtype"] = step.SyntheticsAPIWaitStep.GetSubtype()
				localStep["value"] = step.SyntheticsAPIWaitStep.GetValue()
			}

			localSteps[i] = localStep
		}

		if err := d.Set("api_step", localSteps); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("device_ids", syntheticsTest.GetOptions().DeviceIds); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("locations", syntheticsTest.Locations); err != nil {
		return diag.FromErr(err)
	}

	localOptionsLists := buildTerraformTestOptions(syntheticsTest.GetOptions())

	if err := d.Set("options_list", localOptionsLists); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", syntheticsTest.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", syntheticsTest.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", syntheticsTest.GetStatus()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", syntheticsTest.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_id", syntheticsTest.MonitorId); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func updateSyntheticsMobileTestLocalState(d *schema.ResourceData, syntheticsTest *datadogV1.SyntheticsMobileTest) diag.Diagnostics {
	// There two fields that you might think should be here but are not:
	// - `device_ids` at the root of the request can be set by the user, but is not part of the response
	// - `locations` can not be set by the user as mobile tests only run on one location, but it's part of the response
	if err := d.Set("type", syntheticsTest.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", syntheticsTest.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", syntheticsTest.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", syntheticsTest.GetStatus()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", syntheticsTest.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_id", syntheticsTest.MonitorId); err != nil {
		return diag.FromErr(err)
	}

	config := syntheticsTest.GetConfig()

	actualVariables := config.GetVariables()
	localMobileVariables := make([]map[string]interface{}, len(actualVariables))

	for i, variable := range actualVariables {
		localVariable := make(map[string]interface{})
		if v, ok := variable.GetTypeOk(); ok {
			localVariable["type"] = *v
		}
		if v, ok := variable.GetNameOk(); ok {
			localVariable["name"] = *v
		}
		if v, ok := variable.GetIdOk(); ok {
			localVariable["id"] = *v
		}
		if v, ok := variable.GetSecureOk(); ok {
			localVariable["secure"] = *v
		}
		if v, ok := variable.GetExampleOk(); ok {
			localVariable["example"] = *v
		} else if v, ok := localVariable["secure"].(bool); ok && v {
			localVariable["example"] = d.Get(fmt.Sprintf("mobile_variable.%d.example", i))
		}
		if v, ok := variable.GetPatternOk(); ok {
			localVariable["pattern"] = *v
		} else if v, ok := localVariable["secure"].(bool); ok && v {
			localVariable["pattern"] = d.Get(fmt.Sprintf("mobile_variable.%d.pattern", i))
		}
		localMobileVariables[i] = localVariable
	}
	if err := d.Set("config_variable", localMobileVariables); err != nil {
		return diag.FromErr(err)
	}

	if config.HasInitialApplicationArguments() {
		if err := d.Set("config_initial_application_arguments", config.GetInitialApplicationArguments()); err != nil {
			return diag.FromErr(err)
		}
	}

	localOptionsLists := buildTerraformMobileTestOptions(syntheticsTest.GetOptions())

	if err := d.Set("mobile_options_list", localOptionsLists); err != nil {
		return diag.FromErr(err)
	}

	steps := syntheticsTest.GetSteps()
	localSteps := buildTerraformMobileTestSteps(steps)

	if err := d.Set("mobile_step", localSteps); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

/*
 * transformer functions between datadog and terraform
 */

func buildDatadogSyntheticsAPITest(d *schema.ResourceData) *datadogV1.SyntheticsAPITest {
	syntheticsTest := datadogV1.NewSyntheticsAPITestWithDefaults()
	syntheticsTest.SetName(d.Get("name").(string))

	if attr, ok := d.GetOk("subtype"); ok {
		syntheticsTest.SetSubtype(datadogV1.SyntheticsTestDetailsSubType(attr.(string)))
	} else {
		syntheticsTest.SetSubtype(datadogV1.SYNTHETICSTESTDETAILSSUBTYPE_HTTP)
	}

	request := datadogV1.SyntheticsTestRequest{}
	method, methodOk := d.GetOk("request_definition.0.method")
	if methodOk {
		request.SetMethod(method.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.url"); ok {
		request.SetUrl(attr.(string))
	}
	// Only set the body if the request method allows it
	body, bodyOk := d.GetOk("request_definition.0.body")
	httpVersion, httpVersionOk := d.GetOk("options_list.0.http_version")
	if bodyOk && body != "" {
		if methodOk && (method == "GET" || method == "HEAD" || method == "DELETE") && (!httpVersionOk || httpVersion != "http1") {
			log.Printf("[WARN] body is not valid for %s requests. It'll be ignored.", method)
		} else {
			request.SetBody(body.(string))
		}
	}
	if attr, ok := d.GetOk("request_definition.0.body_type"); ok {
		request.SetBodyType(datadogV1.SyntheticsTestRequestBodyType(attr.(string)))
	}
	if attr, ok := d.GetOk("request_file"); ok && attr != nil && len(attr.([]interface{})) > 0 {
		request.SetFiles(buildDatadogBodyFiles(attr.([]interface{})))
	}
	if attr, ok := d.GetOk("request_definition.0.timeout"); ok {
		request.SetTimeout(float64(attr.(int)))
	}
	if attr, ok := d.GetOk("request_definition.0.host"); ok {
		request.SetHost(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.port"); ok {
		port := attr.(string)
		request.SetPort(datadogV1.SyntheticsTestRequestPort{
			SyntheticsTestRequestVariablePort: &port,
		})
	}
	if attr, ok := d.GetOk("request_definition.0.dns_server"); ok {
		request.SetDnsServer(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.dns_server_port"); ok {
		request.SetDnsServerPort(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.no_saving_response_body"); ok {
		request.SetNoSavingResponseBody(attr.(bool))
	}
	if attr, ok := d.GetOk("request_definition.0.number_of_packets"); ok {
		request.SetNumberOfPackets(int32(attr.(int)))
	}
	if attr, ok := d.GetOk("request_definition.0.should_track_hops"); ok {
		request.SetShouldTrackHops(attr.(bool))
	}
	if attr, ok := d.GetOk("request_definition.0.servername"); ok {
		request.SetServername(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.message"); ok {
		request.SetMessage(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.call_type"); ok {
		request.SetCallType(datadogV1.SyntheticsTestCallType(attr.(string)))
	}
	if syntheticsTest.GetSubtype() == "grpc" {
		if attr, ok := d.GetOk("request_definition.0.service"); ok {
			request.SetService(attr.(string))
		} else {
			request.SetService("")
		}
	}
	if attr, ok := d.GetOk("request_definition.0.persist_cookies"); ok {
		request.SetPersistCookies(attr.(bool))
	}
	if attr, ok := d.GetOk("request_definition.0.proto_json_descriptor"); ok {
		request.SetCompressedJsonDescriptor(compressAndEncodeValue(attr.(string)))
	}
	if attr, ok := d.GetOk("request_definition.0.plain_proto_file"); ok {
		request.SetCompressedProtoFile(compressAndEncodeValue(attr.(string)))
	}

	if attr, ok := d.GetOk("request_client_certificate"); ok {
		if requestClientCertificates, ok := attr.([]interface{}); ok && len(requestClientCertificates) > 0 {
			if requestClientCertificate, ok := requestClientCertificates[0].(map[string]interface{}); ok {
				clientCert, clientKey := getCertAndKeyFromMap(requestClientCertificate)
				request.SetCertificate(buildDatadogRequestCertificates(clientCert["content"].(string), clientCert["filename"].(string), clientKey["content"].(string), clientKey["filename"].(string)))
			}
		}
	}

	request = *completeSyntheticsTestRequest(request, d.Get("request_headers").(map[string]interface{}), d.Get("request_query").(map[string]interface{}), d.Get("request_basicauth").([]interface{}), d.Get("request_proxy").([]interface{}), d.Get("request_metadata").(map[string]interface{}))

	config := datadogV1.NewSyntheticsAPITestConfigWithDefaults()

	if syntheticsTest.GetSubtype() != "multi" {
		config.SetRequest(request)
	}

	config.Assertions = []datadogV1.SyntheticsAssertion{}
	if attr, ok := d.GetOk("assertion"); ok && attr != nil {
		assertions := buildDatadogAssertions(attr.([]interface{}))
		config.Assertions = assertions
	}

	requestConfigVariables := d.Get("config_variable").([]interface{})
	config.SetConfigVariables(buildDatadogConfigVariables(requestConfigVariables))

	if attr, ok := d.GetOk("variables_from_script"); ok && attr != nil {
		config.SetVariablesFromScript(attr.(string))
	}

	if attr, ok := d.GetOk("api_step"); ok && syntheticsTest.GetSubtype() == "multi" {
		steps := []datadogV1.SyntheticsAPIStep{}

		for i, s := range attr.([]interface{}) {
			step := datadogV1.SyntheticsAPIStep{}
			stepMap := s.(map[string]interface{})

			stepSubtype := stepMap["subtype"].(string)

			if stepSubtype == "" || stepSubtype == "http" || stepSubtype == "grpc" {
				step.SyntheticsAPITestStep = datadogV1.NewSyntheticsAPITestStepWithDefaults()
				step.SyntheticsAPITestStep.SetName(stepMap["name"].(string))
				step.SyntheticsAPITestStep.SetSubtype(datadogV1.SyntheticsAPITestStepSubtype(stepMap["subtype"].(string)))

				extractedValues := buildDatadogExtractedValues(stepMap["extracted_value"].([]interface{}))
				step.SyntheticsAPITestStep.SetExtractedValues(extractedValues)

				assertions := stepMap["assertion"].([]interface{})
				step.SyntheticsAPITestStep.SetAssertions(buildDatadogAssertions(assertions))

				request := datadogV1.SyntheticsTestRequest{}
				requests := stepMap["request_definition"].([]interface{})
				if len(requests) > 0 && requests[0] != nil {
					requestMap := requests[0].(map[string]interface{})
					method := requestMap["method"].(string)
					request.SetMethod(method)
					request.SetTimeout(float64(requestMap["timeout"].(int)))
					request.SetAllowInsecure(requestMap["allow_insecure"].(bool))
					if step.SyntheticsAPITestStep.GetSubtype() == "grpc" {
						request.SetHost(requestMap["host"].(string))
						port := requestMap["port"].(string)
						request.SetPort(datadogV1.SyntheticsTestRequestPort{
							SyntheticsTestRequestVariablePort: &port,
						})
						request.SetService(requestMap["service"].(string))
						request.SetMessage(requestMap["message"].(string))
						if v, ok := requestMap["call_type"].(string); ok && v != "" {
							request.SetCallType(datadogV1.SyntheticsTestCallType(v))
						}
						if v, ok := requestMap["plain_proto_file"].(string); ok && v != "" {
							request.SetCompressedProtoFile(compressAndEncodeValue(v))
						}
					} else if step.SyntheticsAPITestStep.GetSubtype() == "http" {
						request.SetUrl(requestMap["url"].(string))
						httpVersion, httpVersionOk := requestMap["http_version"].(string)
						if httpVersionOk && httpVersion != "" {
							request.SetHttpVersion(datadogV1.SyntheticsTestOptionsHTTPVersion(httpVersion))
						}
						// Only set the body if the request method allows it
						body := requestMap["body"].(string)
						if body != "" {
							if (method == "GET" || method == "HEAD" || method == "DELETE") && httpVersion != "http1" {
								log.Printf("[WARN] body is not valid for %s requests. It'll be ignored.", method)
							} else {
								request.SetBody(body)
							}
						}
						request.SetFollowRedirects(requestMap["follow_redirects"].(bool))
						request.SetPersistCookies(requestMap["persist_cookies"].(bool))
						request.SetNoSavingResponseBody(requestMap["no_saving_response_body"].(bool))
						if v, ok := requestMap["body_type"].(string); ok && v != "" {
							request.SetBodyType(datadogV1.SyntheticsTestRequestBodyType(v))
						}

						if attr, ok := stepMap["request_file"]; ok && attr != nil && len(attr.([]interface{})) > 0 {
							request.SetFiles(buildDatadogBodyFiles(attr.([]interface{})))
						}
					}
				}
				// Override the request client certificate with the one from the config
				configCertContent, configKeyContent := getConfigCertAndKeyContent(d, i)

				if requestClientCertificates, ok := stepMap["request_client_certificate"].([]interface{}); ok && len(requestClientCertificates) > 0 {
					if requestClientCertificate, ok := requestClientCertificates[0].(map[string]interface{}); ok {
						clientCert, clientKey := getCertAndKeyFromMap(requestClientCertificate)
						if configCertContent != nil || configKeyContent != nil {
							request.SetCertificate(buildDatadogRequestCertificates(*configCertContent, clientCert["filename"].(string), *configKeyContent, clientKey["filename"].(string)))
						}
					}
				}

				request = *completeSyntheticsTestRequest(request, stepMap["request_headers"].(map[string]interface{}), stepMap["request_query"].(map[string]interface{}), stepMap["request_basicauth"].([]interface{}), stepMap["request_proxy"].([]interface{}), stepMap["request_metadata"].(map[string]interface{}))

				step.SyntheticsAPITestStep.SetRequest(request)

				step.SyntheticsAPITestStep.SetAllowFailure(stepMap["allow_failure"].(bool))
				step.SyntheticsAPITestStep.SetExitIfSucceed(stepMap["exit_if_succeed"].(bool))
				step.SyntheticsAPITestStep.SetIsCritical(stepMap["is_critical"].(bool))

				optionsRetry := datadogV1.SyntheticsTestOptionsRetry{}
				retries := stepMap["retry"].([]interface{})
				if len(retries) > 0 && retries[0] != nil {
					retry := retries[0]

					if count, ok := retry.(map[string]interface{})["count"]; ok {
						optionsRetry.SetCount(int64(count.(int)))
					}
					if interval, ok := retry.(map[string]interface{})["interval"]; ok {
						optionsRetry.SetInterval(float64(interval.(int)))
					}
					step.SyntheticsAPITestStep.SetRetry(optionsRetry)
				}
			} else if stepSubtype == "wait" {
				step.SyntheticsAPIWaitStep = datadogV1.NewSyntheticsAPIWaitStepWithDefaults()
				step.SyntheticsAPIWaitStep.SetName(stepMap["name"].(string))
				step.SyntheticsAPIWaitStep.SetSubtype(datadogV1.SyntheticsAPIWaitStepSubtype(stepMap["subtype"].(string)))
				step.SyntheticsAPIWaitStep.SetValue(int32(stepMap["value"].(int)))
			}

			steps = append(steps, step)
		}

		config.SetSteps(steps)
	}

	options := buildDatadogTestOptions(d)

	syntheticsTest.SetConfig(*config)
	syntheticsTest.SetOptions(*options)
	syntheticsTest.SetMessage(d.Get("message").(string))
	syntheticsTest.SetStatus(datadogV1.SyntheticsTestPauseStatus(d.Get("status").(string)))

	if attr, ok := d.GetOk("locations"); ok {
		var locations []string
		for _, s := range attr.(*schema.Set).List() {
			locations = append(locations, s.(string))
		}
		syntheticsTest.SetLocations(locations)
	}

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
	}
	syntheticsTest.SetTags(tags)

	return syntheticsTest
}

func buildDatadogSyntheticsBrowserTest(d *schema.ResourceData) *datadogV1.SyntheticsBrowserTest {
	request := datadogV1.SyntheticsTestRequest{}
	if attr, ok := d.GetOk("request_definition.0.method"); ok {
		request.SetMethod(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.url"); ok {
		request.SetUrl(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.body"); ok {
		request.SetBody(attr.(string))
	}
	if attr, ok := d.GetOk("request_definition.0.body_type"); ok {
		request.SetBodyType(datadogV1.SyntheticsTestRequestBodyType(attr.(string)))
	}
	if attr, ok := d.GetOk("request_definition.0.timeout"); ok {
		request.SetTimeout(float64(attr.(int)))
	}
	if attr, ok := d.GetOk("request_definition.0.certificate_domains"); ok {
		var certificateDomains []string

		for _, s := range attr.([]interface{}) {
			certificateDomains = append(certificateDomains, s.(string))
		}
		request.SetCertificateDomains(certificateDomains)
	}

	if attr, ok := d.GetOk("request_query"); ok {
		query := attr.(map[string]interface{})
		if len(query) > 0 {
			request.SetQuery(query)
		}
	}
	if username, ok := d.GetOk("request_basicauth.0.username"); ok {
		if password, ok := d.GetOk("request_basicauth.0.password"); ok {
			basicAuth := datadogV1.NewSyntheticsBasicAuthWebWithDefaults()
			basicAuth.SetPassword(password.(string))
			basicAuth.SetUsername(username.(string))
			// Works for Web Basic Auth, NTLM and Digest as they all use `username` + `password`
			request.SetBasicAuth(datadogV1.SyntheticsBasicAuthWebAsSyntheticsBasicAuth(basicAuth))
		}
	}

	if attr, ok := d.GetOk("request_headers"); ok {
		headers := attr.(map[string]interface{})
		if len(headers) > 0 {
			request.SetHeaders(make(map[string]string))
		}
		for k, v := range headers {
			request.GetHeaders()[k] = v.(string)
		}
	}

	if attr, ok := d.GetOk("request_client_certificate"); ok {
		requestClientCertificate := attr.(map[string]interface{})
		clientCert, clientKey := getCertAndKeyFromMap(requestClientCertificate)
		request.SetCertificate(buildDatadogRequestCertificates(clientCert["content"].(string), clientCert["filename"].(string), clientKey["content"].(string), clientKey["filename"].(string)))
	}

	if attr, ok := d.GetOk("request_proxy"); ok {
		requestProxies := attr.([]interface{})
		if requestProxy, ok := requestProxies[0].(map[string]interface{}); ok {
			request.SetProxy(buildDatadogTestRequestProxy(requestProxy))
		}
	}

	config := datadogV1.SyntheticsBrowserTestConfig{}
	config.SetAssertions([]datadogV1.SyntheticsAssertion{})
	config.SetRequest(request)
	config.SetVariables([]datadogV1.SyntheticsBrowserVariable{})

	var browserVariables []interface{}

	if attr, ok := d.GetOk("browser_variable"); ok && attr != nil {
		browserVariables = attr.([]interface{})
	}

	for _, variable := range browserVariables {
		variableMap := variable.(map[string]interface{})
		if v, ok := variableMap["type"]; ok {
			variableType := datadogV1.SyntheticsBrowserVariableType(v.(string))
			if v, ok := variableMap["name"]; ok {
				variableName := v.(string)
				newVariable := datadogV1.NewSyntheticsBrowserVariable(variableName, variableType)
				if v, ok := variableMap["example"]; ok {
					newVariable.SetExample(v.(string))
				}
				if v, ok := variableMap["id"]; ok && v.(string) != "" {
					newVariable.SetId(v.(string))
				}
				if v, ok := variableMap["pattern"]; ok {
					newVariable.SetPattern(v.(string))
				}
				if v, ok := variableMap["secure"]; ok && variableType == datadogV1.SYNTHETICSBROWSERVARIABLETYPE_TEXT {
					newVariable.SetSecure(v.(bool))
				}

				config.SetVariables(append(config.GetVariables(), *newVariable))
			}
		}
	}

	requestConfigVariables := d.Get("config_variable").([]interface{})
	config.SetConfigVariables(buildDatadogConfigVariables(requestConfigVariables))

	if attr, ok := d.GetOk("set_cookie"); ok {
		config.SetSetCookie(attr.(string))
	}

	options := buildDatadogTestOptions(d)

	syntheticsTest := datadogV1.NewSyntheticsBrowserTestWithDefaults()
	syntheticsTest.SetMessage(d.Get("message").(string))
	syntheticsTest.SetName(d.Get("name").(string))
	syntheticsTest.SetConfig(config)
	syntheticsTest.SetOptions(*options)
	syntheticsTest.SetStatus(datadogV1.SyntheticsTestPauseStatus(d.Get("status").(string)))

	if attr, ok := d.GetOk("locations"); ok {
		var locations []string
		for _, s := range attr.(*schema.Set).List() {
			locations = append(locations, s.(string))
		}
		syntheticsTest.SetLocations(locations)
	}

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			if tag, ok := s.(string); ok {
				tags = append(tags, tag)
			}
		}
	}
	syntheticsTest.SetTags(tags)

	if attr, ok := d.GetOk("browser_step"); ok {
		steps := []datadogV1.SyntheticsStep{}

		for _, s := range attr.([]interface{}) {
			step := datadogV1.SyntheticsStep{}
			stepMap := s.(map[string]interface{})

			step.SetName(stepMap["name"].(string))
			step.SetType(datadogV1.SyntheticsStepType(stepMap["type"].(string)))
			step.SetAllowFailure(stepMap["allow_failure"].(bool))
			step.SetAlwaysExecute(stepMap["always_execute"].(bool))
			step.SetExitIfSucceed(stepMap["exit_if_succeed"].(bool))
			step.SetIsCritical(stepMap["is_critical"].(bool))
			step.SetTimeout(int64(stepMap["timeout"].(int)))
			step.SetNoScreenshot(stepMap["no_screenshot"].(bool))

			params := getStepParams(stepMap, d)
			step.SetParams(params)

			steps = append(steps, step)
		}

		syntheticsTest.SetSteps(steps)
	}

	return syntheticsTest
}

func buildDatadogSyntheticsMobileTest(d *schema.ResourceData) *datadogV1.SyntheticsMobileTest {
	syntheticsTest := datadogV1.NewSyntheticsMobileTestWithDefaults()

	// There three fields that you might think should be here but are not:
	// - `device_ids` at the root of the request can be set by the user, but is not part of the response, so we're not going to set it as not to mess with the local state of Terraform
	// - `locations` can not be set by the user as mobile tests only run on one location, but it's part of the response
	// - `monitor_id` is not set by the user, but returned by the response
	if attr, ok := d.GetOk("message"); ok {
		syntheticsTest.SetMessage(attr.(string))
	}
	if attr, ok := d.GetOk("name"); ok {
		syntheticsTest.SetName(attr.(string))
	}
	if attr, ok := d.GetOk("status"); ok {
		syntheticsTest.SetStatus(datadogV1.SyntheticsTestPauseStatus(attr.(string)))
	}
	if attr, ok := d.GetOk("type"); ok {
		syntheticsTest.SetType(datadogV1.SyntheticsMobileTestType(attr.(string)))
	}
	config := datadogV1.SyntheticsMobileTestConfig{}
	config.SetVariables([]datadogV1.SyntheticsConfigVariable{})

	requestConfigVariables := d.Get("config_variable").([]interface{})
	config.SetVariables(buildDatadogConfigVariables(requestConfigVariables))

	if attr, ok := d.GetOk("config_initial_application_arguments"); ok {
		initialApplicationArguments := attr.(map[string]interface{})
		if len(initialApplicationArguments) > 0 {
			config.SetInitialApplicationArguments(make(map[string]string))
		}
		for k, v := range initialApplicationArguments {
			config.GetInitialApplicationArguments()[k] = v.(string)
		}
	}

	syntheticsTest.SetConfig(config)

	options := buildDatadogMobileTestOptions(d)
	syntheticsTest.SetOptions(*options)

	if attr, ok := d.GetOk("mobile_step"); ok {
		steps := []datadogV1.SyntheticsMobileStep{}

		for _, s := range attr.([]interface{}) {
			step := datadogV1.SyntheticsMobileStep{}
			stepMap := s.(map[string]interface{})

			step.SetAllowFailure(stepMap["allow_failure"].(bool))
			step.SetHasNewStepElement(stepMap["has_new_step_element"].(bool))
			step.SetIsCritical(stepMap["is_critical"].(bool))
			step.SetNoScreenshot(stepMap["no_screenshot"].(bool))

			if stepMap["name"] != "" {
				step.SetName(stepMap["name"].(string))
			}
			if stepMap["public_id"] != "" {
				step.SetPublicId(stepMap["public_id"].(string))
			}
			if stepMap["timeout"] != 0 {
				step.SetTimeout(int64(stepMap["timeout"].(int)))
			}
			if stepMap["type"] != "" {
				step.SetType(datadogV1.SyntheticsMobileStepType(stepMap["type"].(string)))
			}

			params := datadogV1.SyntheticsMobileStepParams{}
			stepParams := stepMap["params"].([]interface{})[0]
			params = buildDatadogParamsForMobileStep(step.GetType(), stepParams.(map[string]interface{}))
			step.SetParams(params)
			steps = append(steps, step)
		}

		syntheticsTest.SetSteps(steps)
	}

	if attr, ok := d.GetOk("tags"); ok {
		tags := make([]string, 0)
		for _, s := range attr.([]interface{}) {
			if tag, ok := s.(string); ok {
				tags = append(tags, tag)
			}
		}
		syntheticsTest.SetTags(tags)
	}

	return syntheticsTest
}

func buildDatadogAssertions(attr []interface{}) []datadogV1.SyntheticsAssertion {
	assertions := make([]datadogV1.SyntheticsAssertion, 0)

	for _, assertion := range attr {
		assertionMap := assertion.(map[string]interface{})
		if v, ok := assertionMap["type"]; ok {
			assertionType := v.(string)
			if assertionType == string(datadogV1.SYNTHETICSASSERTIONJAVASCRIPTTYPE_JAVASCRIPT) {
				// Handling the case for javascript assertion that does not contains any `operator`
				assertionJavascript := datadogV1.NewSyntheticsAssertionJavascriptWithDefaults()
				assertionJavascript.SetType(datadogV1.SYNTHETICSASSERTIONJAVASCRIPTTYPE_JAVASCRIPT)
				if v, ok := assertionMap["code"]; ok {
					assertionCode := v.(string)
					assertionJavascript.SetCode((assertionCode))
				}
				assertions = append(assertions, datadogV1.SyntheticsAssertionJavascriptAsSyntheticsAssertion(assertionJavascript))
			} else if v, ok := assertionMap["operator"]; ok {
				assertionOperator := v.(string)
				if assertionOperator == string(datadogV1.SYNTHETICSASSERTIONJSONSCHEMAOPERATOR_VALIDATES_JSON_SCHEMA) {
					assertionJSONSchemaTarget := datadogV1.NewSyntheticsAssertionJSONSchemaTarget(datadogV1.SyntheticsAssertionJSONSchemaOperator(assertionOperator), datadogV1.SyntheticsAssertionType(assertionType))
					if v, ok := assertionMap["targetjsonschema"].([]interface{}); ok && len(v) > 0 {
						subTarget := datadogV1.NewSyntheticsAssertionJSONSchemaTargetTarget()
						targetMap := v[0].(map[string]interface{})
						if v, ok := targetMap["jsonschema"]; ok {
							subTarget.SetJsonSchema(v.(string))
						}
						if v, ok := targetMap["metaschema"]; ok {
							if metaSchema, err := datadogV1.NewSyntheticsAssertionJSONSchemaMetaSchemaFromValue(v.(string)); err == nil {
								subTarget.SetMetaSchema(*metaSchema)
							} else {
								log.Printf("[ERROR] Error converting json schema meta schema: %v", err)
							}
						}
						assertionJSONSchemaTarget.SetTarget(*subTarget)
					}
					if _, ok := assertionMap["target"]; ok {
						log.Printf("[WARN] target shouldn't be specified for validateJSONSchema operator, only targetJSONSchema")
					}
					assertions = append(assertions, datadogV1.SyntheticsAssertionJSONSchemaTargetAsSyntheticsAssertion(assertionJSONSchemaTarget))
				} else if assertionOperator == string(datadogV1.SYNTHETICSASSERTIONJSONPATHOPERATOR_VALIDATES_JSON_PATH) {
					assertionJSONPathTarget := datadogV1.NewSyntheticsAssertionJSONPathTarget(datadogV1.SyntheticsAssertionJSONPathOperator(assertionOperator), datadogV1.SyntheticsAssertionType(assertionType))
					if v, ok := assertionMap["property"].(string); ok && len(v) > 0 {
						assertionJSONPathTarget.SetProperty(v)
					}
					if v, ok := assertionMap["targetjsonpath"].([]interface{}); ok && len(v) > 0 {
						subTarget := datadogV1.NewSyntheticsAssertionJSONPathTargetTarget()
						targetMap := v[0].(map[string]interface{})
						if v, ok := targetMap["jsonpath"]; ok {
							subTarget.SetJsonPath(v.(string))
						}

						operator, ok := targetMap["operator"]
						if ok {
							subTarget.SetOperator(operator.(string))
						}
						if v, ok := targetMap["targetvalue"]; ok {
							switch datadogV1.SyntheticsAssertionOperator(operator.(string)) {
							case datadogV1.SYNTHETICSASSERTIONOPERATOR_IS_UNDEFINED:
								// no target value must be set for isUndefined operator
							case
								datadogV1.SYNTHETICSASSERTIONOPERATOR_LESS_THAN,
								datadogV1.SYNTHETICSASSERTIONOPERATOR_MORE_THAN:
								if match, _ := regexp.MatchString("{{\\s*([^{}]*?)\\s*}}", v.(string)); match {
									subTarget.SetTargetValue(v)
								} else {
									if floatValue, err := strconv.ParseFloat(v.(string), 64); err == nil {
										subTarget.SetTargetValue(floatValue)
									}
								}
							default:
								subTarget.SetTargetValue(v)
							}
						}
						if v, ok := targetMap["elementsoperator"]; ok {
							subTarget.SetElementsOperator(v.(string))
						}
						assertionJSONPathTarget.SetTarget(*subTarget)
					}
					if _, ok := assertionMap["target"]; ok {
						log.Printf("[WARN] target shouldn't be specified for validatesJSONPath operator, only targetjsonpath")
					}
					assertions = append(assertions, datadogV1.SyntheticsAssertionJSONPathTargetAsSyntheticsAssertion(assertionJSONPathTarget))
				} else if assertionOperator == string(datadogV1.SYNTHETICSASSERTIONXPATHOPERATOR_VALIDATES_X_PATH) {
					assertionXPathTarget := datadogV1.NewSyntheticsAssertionXPathTarget(datadogV1.SyntheticsAssertionXPathOperator(assertionOperator), datadogV1.SyntheticsAssertionType(assertionType))
					if v, ok := assertionMap["property"].(string); ok && len(v) > 0 {
						assertionXPathTarget.SetProperty(v)
					}
					if v, ok := assertionMap["targetxpath"].([]interface{}); ok && len(v) > 0 {
						subTarget := datadogV1.NewSyntheticsAssertionXPathTargetTarget()
						targetMap := v[0].(map[string]interface{})
						if v, ok := targetMap["xpath"]; ok {
							subTarget.SetXPath(v.(string))
						}
						operator, ok := targetMap["operator"]
						if ok {
							subTarget.SetOperator(operator.(string))
						}
						if v, ok := targetMap["targetvalue"]; ok {
							switch datadogV1.SyntheticsAssertionOperator(operator.(string)) {
							case
								datadogV1.SYNTHETICSASSERTIONOPERATOR_LESS_THAN,
								datadogV1.SYNTHETICSASSERTIONOPERATOR_MORE_THAN:
								if match, _ := regexp.MatchString("{{\\s*([^{}]*?)\\s*}}", v.(string)); match {
									subTarget.SetTargetValue(v)
								} else {
									if floatValue, err := strconv.ParseFloat(v.(string), 64); err == nil {
										subTarget.SetTargetValue(floatValue)
									}
								}
							default:
								subTarget.SetTargetValue(v)
							}
						}
						assertionXPathTarget.SetTarget(*subTarget)
					}
					if _, ok := assertionMap["target"]; ok {
						log.Printf("[WARN] target shouldn't be specified for validateXPath operator, only targetxpath")
					}
					assertions = append(assertions, datadogV1.SyntheticsAssertionXPathTargetAsSyntheticsAssertion(assertionXPathTarget))
				} else {
					assertionTarget := datadogV1.NewSyntheticsAssertionTargetWithDefaults()
					assertionTarget.SetOperator(datadogV1.SyntheticsAssertionOperator(assertionOperator))
					assertionTarget.SetType(datadogV1.SyntheticsAssertionType(assertionType))
					if v, ok := assertionMap["property"].(string); ok && len(v) > 0 {
						assertionTarget.SetProperty(v)
					}
					if v, ok := assertionMap["target"]; ok {
						if isTargetOfTypeInt(assertionTarget.GetType(), assertionTarget.GetOperator()) {
							assertionTargetInt, _ := strconv.Atoi(v.(string))
							assertionTarget.SetTarget(assertionTargetInt)
						} else if assertionTarget.GetType() == datadogV1.SYNTHETICSASSERTIONTYPE_PACKET_LOSS_PERCENTAGE {
							assertionTargetFloat, _ := strconv.ParseFloat(v.(string), 64)
							assertionTarget.SetTarget(assertionTargetFloat)
						} else {
							assertionTarget.SetTarget(v.(string))
						}
					}
					if v, ok := assertionMap["timings_scope"].(string); ok && len(v) > 0 {
						assertionTarget.SetTimingsScope(datadogV1.SyntheticsAssertionTimingsScope(v))
					}
					if v, ok := assertionMap["targetjsonschema"].([]interface{}); ok && len(v) > 0 {
						log.Printf("[WARN] targetjsonschema shouldn't be specified for non-validatesJSONSchema operator, only target")
					}
					if v, ok := assertionMap["targetjsonpath"].([]interface{}); ok && len(v) > 0 {
						log.Printf("[WARN] targetjsonpath shouldn't be specified for non-validatesJSONPath operator, only target")
					}
					if v, ok := assertionMap["targetxpath"].([]interface{}); ok && len(v) > 0 {
						log.Printf("[WARN] targetxpath shouldn't be specified for non-validatesXPath operator, only target")
					}
					assertions = append(assertions, datadogV1.SyntheticsAssertionTargetAsSyntheticsAssertion(assertionTarget))
				}
			}
		}
	}

	return assertions
}

func buildTerraformAssertions(actualAssertions []datadogV1.SyntheticsAssertion) (localAssertions []map[string]interface{}, err error) {
	localAssertions = make([]map[string]interface{}, len(actualAssertions))
	for i, assertion := range actualAssertions {
		localAssertion := make(map[string]interface{})
		if assertion.SyntheticsAssertionTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion["operator"] = string(*v)
			}
			if assertionTarget.HasProperty() {
				localAssertion["property"] = assertionTarget.GetProperty()
			}
			if target := assertionTarget.GetTarget(); target != nil {
				localAssertion["target"] = convertToString(target)
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}
			if assertionTarget.HasTimingsScope() {
				localAssertion["timings_scope"] = assertionTarget.GetTimingsScope()
			}
		} else if assertion.SyntheticsAssertionJSONSchemaTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionJSONSchemaTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion["operator"] = string(*v)
			}
			if target, ok := assertionTarget.GetTargetOk(); ok {
				localTarget := make(map[string]string)
				if v, ok := target.GetJsonSchemaOk(); ok {
					localTarget["jsonschema"] = string(*v)
				}
				if v, ok := target.GetMetaSchemaOk(); ok {
					localTarget["metaschema"] = string(*v)
				}
				localAssertion["targetjsonschema"] = []map[string]string{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}
		} else if assertion.SyntheticsAssertionJSONPathTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionJSONPathTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion["operator"] = string(*v)
			}
			if assertionTarget.HasProperty() {
				localAssertion["property"] = assertionTarget.GetProperty()
			}
			if target, ok := assertionTarget.GetTargetOk(); ok {
				localTarget := make(map[string]string)
				if v, ok := target.GetJsonPathOk(); ok {
					localTarget["jsonpath"] = string(*v)
				}
				if v, ok := target.GetOperatorOk(); ok {
					localTarget["operator"] = string(*v)
				}
				if v, ok := target.GetTargetValueOk(); ok {
					val := (*v).(interface{})
					if vAsString, ok := val.(string); ok {
						localTarget["targetvalue"] = vAsString
					} else if vAsFloat, ok := val.(float64); ok {
						localTarget["targetvalue"] = strconv.FormatFloat(vAsFloat, 'f', -1, 64)
					} else {
						return localAssertions, fmt.Errorf("unrecognized targetvalue type %v", v)
					}
				}
				if v, ok := target.GetElementsOperatorOk(); ok {
					localTarget["elementsoperator"] = string(*v)
				}
				localAssertion["targetjsonpath"] = []map[string]string{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}
		} else if assertion.SyntheticsAssertionXPathTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionXPathTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion["operator"] = string(*v)
			}
			if assertionTarget.HasProperty() {
				localAssertion["property"] = assertionTarget.GetProperty()
			}
			if target, ok := assertionTarget.GetTargetOk(); ok {
				localTarget := make(map[string]string)
				if v, ok := target.GetXPathOk(); ok {
					localTarget["xpath"] = string(*v)
				}
				if v, ok := target.GetOperatorOk(); ok {
					localTarget["operator"] = string(*v)
				}
				if v, ok := target.GetTargetValueOk(); ok {
					val := (*v).(interface{})
					if vAsString, ok := val.(string); ok {
						localTarget["targetvalue"] = vAsString
					} else if vAsFloat, ok := val.(float64); ok {
						localTarget["targetvalue"] = strconv.FormatFloat(vAsFloat, 'f', -1, 64)
					} else {
						return localAssertions, fmt.Errorf("unrecognized targetvalue type %v", v)
					}
				}
				localAssertion["targetxpath"] = []map[string]string{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}
		} else if assertion.SyntheticsAssertionBodyHashTarget != nil {
			assertionTarget := assertion.SyntheticsAssertionBodyHashTarget
			if v, ok := assertionTarget.GetOperatorOk(); ok {
				localAssertion["operator"] = string(*v)
			}
			if target := assertionTarget.GetTarget(); target != nil {
				localAssertion["target"] = convertToString(target)
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}
		} else if assertion.SyntheticsAssertionJavascript != nil {
			assertionTarget := assertion.SyntheticsAssertionJavascript

			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}

			if v, ok := assertionTarget.GetCodeOk(); ok {
				localAssertion["code"] = v
			}
		}
		localAssertions[i] = localAssertion
	}

	return localAssertions, nil
}

func buildDatadogBasicAuth(requestBasicAuth map[string]interface{}) datadogV1.SyntheticsBasicAuth {
	if requestBasicAuth["type"] == "web" && requestBasicAuth["username"] != "" {
		basicAuth := datadogV1.NewSyntheticsBasicAuthWebWithDefaults()
		basicAuth.SetPassword(requestBasicAuth["password"].(string))
		basicAuth.SetUsername(requestBasicAuth["username"].(string))
		return datadogV1.SyntheticsBasicAuthWebAsSyntheticsBasicAuth(basicAuth)
	}

	if requestBasicAuth["type"] == "sigv4" && requestBasicAuth["access_key"] != "" && requestBasicAuth["secret_key"] != "" {
		basicAuth := datadogV1.NewSyntheticsBasicAuthSigv4(requestBasicAuth["access_key"].(string), requestBasicAuth["secret_key"].(string), datadogV1.SYNTHETICSBASICAUTHSIGV4TYPE_SIGV4)

		basicAuth.SetRegion(requestBasicAuth["region"].(string))
		basicAuth.SetServiceName(requestBasicAuth["service_name"].(string))
		basicAuth.SetSessionToken(requestBasicAuth["session_token"].(string))

		return datadogV1.SyntheticsBasicAuthSigv4AsSyntheticsBasicAuth(basicAuth)
	}

	if requestBasicAuth["type"] == "ntlm" {
		basicAuth := datadogV1.NewSyntheticsBasicAuthNTLM(datadogV1.SYNTHETICSBASICAUTHNTLMTYPE_NTLM)

		basicAuth.SetUsername(requestBasicAuth["username"].(string))
		basicAuth.SetPassword(requestBasicAuth["password"].(string))
		basicAuth.SetDomain(requestBasicAuth["domain"].(string))
		basicAuth.SetWorkstation(requestBasicAuth["workstation"].(string))

		return datadogV1.SyntheticsBasicAuthNTLMAsSyntheticsBasicAuth(basicAuth)
	}

	if requestBasicAuth["type"] == "oauth-client" {
		tokenApiAuthentication, err := datadogV1.NewSyntheticsBasicAuthOauthTokenApiAuthenticationFromValue(requestBasicAuth["token_api_authentication"].(string))
		var tokenApiAuthenticationValue datadogV1.SyntheticsBasicAuthOauthTokenApiAuthentication
		if err == nil {
			tokenApiAuthenticationValue = *tokenApiAuthentication
		}
		basicAuth := datadogV1.NewSyntheticsBasicAuthOauthClient(
			requestBasicAuth["access_token_url"].(string),
			requestBasicAuth["client_id"].(string),
			requestBasicAuth["client_secret"].(string),
			tokenApiAuthenticationValue,
			datadogV1.SYNTHETICSBASICAUTHOAUTHCLIENTTYPE_OAUTH_CLIENT,
		)

		// optional fields for oauth must not be included if they have no value, or the authentication will fail
		if v, ok := requestBasicAuth["audience"].(string); ok && v != "" {
			basicAuth.SetAudience(v)
		}
		if v, ok := requestBasicAuth["resource"].(string); ok && v != "" {
			basicAuth.SetResource(v)
		}
		if v, ok := requestBasicAuth["scope"].(string); ok && v != "" {
			basicAuth.SetScope(v)
		}

		return datadogV1.SyntheticsBasicAuthOauthClientAsSyntheticsBasicAuth(basicAuth)
	}

	if requestBasicAuth["type"] == "oauth-rop" {
		tokenApiAuthentication, err := datadogV1.NewSyntheticsBasicAuthOauthTokenApiAuthenticationFromValue(requestBasicAuth["token_api_authentication"].(string))
		var tokenApiAuthenticationValue datadogV1.SyntheticsBasicAuthOauthTokenApiAuthentication
		if err == nil {
			tokenApiAuthenticationValue = *tokenApiAuthentication
		}
		basicAuth := datadogV1.NewSyntheticsBasicAuthOauthROP(
			requestBasicAuth["access_token_url"].(string),
			requestBasicAuth["password"].(string),
			tokenApiAuthenticationValue,
			datadogV1.SYNTHETICSBASICAUTHOAUTHROPTYPE_OAUTH_ROP,
			requestBasicAuth["username"].(string))

		// optional fields for oauth must not be included if they have no value, or the authentication will fail
		if v, ok := requestBasicAuth["audience"].(string); ok && v != "" {
			basicAuth.SetAudience(v)
		}
		if v, ok := requestBasicAuth["resource"].(string); ok && v != "" {
			basicAuth.SetResource(v)
		}
		if v, ok := requestBasicAuth["scope"].(string); ok && v != "" {
			basicAuth.SetScope(v)
		}
		basicAuth.SetClientId(requestBasicAuth["client_id"].(string))
		basicAuth.SetClientSecret(requestBasicAuth["client_secret"].(string))

		return datadogV1.SyntheticsBasicAuthOauthROPAsSyntheticsBasicAuth(basicAuth)
	}

	if requestBasicAuth["type"] == "digest" {
		basicAuth := datadogV1.NewSyntheticsBasicAuthDigest(
			requestBasicAuth["password"].(string),
			datadogV1.SYNTHETICSBASICAUTHDIGESTTYPE_DIGEST,
			requestBasicAuth["username"].(string),
		)
		return datadogV1.SyntheticsBasicAuthDigestAsSyntheticsBasicAuth(basicAuth)
	}

	log.Printf("[WARN] unrecognized basic auth type %s", requestBasicAuth["type"].(string))
	return datadogV1.SyntheticsBasicAuth{}
}

func buildTerraformBasicAuth(basicAuth *datadogV1.SyntheticsBasicAuth) map[string]string {
	localAuth := make(map[string]string)

	if basicAuth.SyntheticsBasicAuthWeb != nil {
		basicAuthWeb := basicAuth.SyntheticsBasicAuthWeb

		localAuth["username"] = basicAuthWeb.Username
		localAuth["password"] = basicAuthWeb.Password
		localAuth["type"] = "web"
	}

	if basicAuth.SyntheticsBasicAuthSigv4 != nil {
		basicAuthSigv4 := basicAuth.SyntheticsBasicAuthSigv4
		localAuth["access_key"] = basicAuthSigv4.AccessKey
		localAuth["secret_key"] = basicAuthSigv4.SecretKey
		if v, ok := basicAuthSigv4.GetRegionOk(); ok {
			localAuth["region"] = *v
		}
		if v, ok := basicAuthSigv4.GetSessionTokenOk(); ok {
			localAuth["session_token"] = *v
		}
		if v, ok := basicAuthSigv4.GetServiceNameOk(); ok {
			localAuth["service_name"] = *v
		}
		localAuth["type"] = "sigv4"
	}

	if basicAuth.SyntheticsBasicAuthNTLM != nil {
		basicAuthNtlm := basicAuth.SyntheticsBasicAuthNTLM
		if v, ok := basicAuthNtlm.GetUsernameOk(); ok {
			localAuth["username"] = *v
		}
		if v, ok := basicAuthNtlm.GetPasswordOk(); ok {
			localAuth["password"] = *v
		}
		if v, ok := basicAuthNtlm.GetDomainOk(); ok {
			localAuth["domain"] = *v
		}
		if v, ok := basicAuthNtlm.GetWorkstationOk(); ok {
			localAuth["workstation"] = *v
		}
		localAuth["type"] = "ntlm"
	}

	if basicAuth.SyntheticsBasicAuthOauthClient != nil {
		basicAuthOauthClient := basicAuth.SyntheticsBasicAuthOauthClient
		localAuth["access_token_url"] = basicAuthOauthClient.AccessTokenUrl
		localAuth["client_id"] = basicAuthOauthClient.ClientId
		localAuth["client_secret"] = basicAuthOauthClient.ClientSecret
		localAuth["token_api_authentication"] = string(basicAuthOauthClient.TokenApiAuthentication)
		if v, ok := basicAuthOauthClient.GetAudienceOk(); ok {
			localAuth["audience"] = *v
		}
		if v, ok := basicAuthOauthClient.GetScopeOk(); ok {
			localAuth["scope"] = *v
		}
		if v, ok := basicAuthOauthClient.GetResourceOk(); ok {
			localAuth["resource"] = *v
		}
		localAuth["type"] = "oauth-client"
	}
	if basicAuth.SyntheticsBasicAuthOauthROP != nil {
		basicAuthOauthROP := basicAuth.SyntheticsBasicAuthOauthROP
		localAuth["access_token_url"] = basicAuthOauthROP.AccessTokenUrl
		if v, ok := basicAuthOauthROP.GetClientIdOk(); ok {
			localAuth["client_id"] = *v
		}
		if v, ok := basicAuthOauthROP.GetClientSecretOk(); ok {
			localAuth["client_secret"] = *v
		}
		localAuth["token_api_authentication"] = string(basicAuthOauthROP.TokenApiAuthentication)
		if v, ok := basicAuthOauthROP.GetAudienceOk(); ok {
			localAuth["audience"] = *v
		}
		if v, ok := basicAuthOauthROP.GetScopeOk(); ok {
			localAuth["scope"] = *v
		}
		if v, ok := basicAuthOauthROP.GetResourceOk(); ok {
			localAuth["resource"] = *v
		}
		localAuth["username"] = basicAuthOauthROP.Username
		localAuth["password"] = basicAuthOauthROP.Password

		localAuth["type"] = "oauth-rop"
	}

	if basicAuth.SyntheticsBasicAuthDigest != nil {
		basicAuthDigest := basicAuth.SyntheticsBasicAuthDigest
		localAuth["username"] = basicAuthDigest.Username
		localAuth["password"] = basicAuthDigest.Password

		localAuth["type"] = "digest"
	}

	return localAuth
}

func buildDatadogBodyFiles(attr []interface{}) []datadogV1.SyntheticsTestRequestBodyFile {
	files := []datadogV1.SyntheticsTestRequestBodyFile{}
	for _, f := range attr {
		fileMap := f.(map[string]interface{})
		file := datadogV1.SyntheticsTestRequestBodyFile{}

		file.SetName(fileMap["name"].(string))
		file.SetOriginalFileName(fileMap["original_file_name"].(string))
		file.SetType(fileMap["type"].(string))
		file.SetSize(int64(fileMap["size"].(int)))

		if content, ok := fileMap["content"]; ok && content != "" {
			file.SetContent(content.(string))
		}

		// We aren't sure yet how to let the provider check if the file content was updated to upload it again.
		// Hence, the provider is uploading the file every time the resource is modified.
		// Always adding the bucket key to the request would prevent updating the file content.
		// Always omitting the existing bucket key from the request update the file every time the resource is updated.
		// We purposely choose the latter.
		// if bucketKey, ok := fileMap["bucket_key"]; ok && bucketKey != "" {
		// 	file.SetBucketKey(bucketKey.(string))
		// }

		files = append(files, file)
	}

	return files
}

func buildTerraformBodyFiles(actualBodyFiles *[]datadogV1.SyntheticsTestRequestBodyFile, oldLocalBodyFiles []map[string]interface{}) (localBodyFiles []map[string]interface{}) {
	localBodyFiles = make([]map[string]interface{}, len(*actualBodyFiles))
	for i, file := range *actualBodyFiles {
		localFile := make(map[string]interface{})
		if i < len(oldLocalBodyFiles) && oldLocalBodyFiles[i] != nil {
			// The file content is kept from the existing localFile from the state,
			// as the response from the backend contains the bucket key rather than the content.
			localFile = oldLocalBodyFiles[i]
		}
		localFile["name"] = file.GetName()
		localFile["original_file_name"] = file.GetOriginalFileName()
		localFile["type"] = file.GetType()
		localFile["size"] = file.GetSize()

		if bucket_key, ok := file.GetBucketKeyOk(); ok {
			localFile["bucket_key"] = bucket_key
		}
		localBodyFiles[i] = localFile
	}
	return localBodyFiles
}

func buildDatadogConfigVariables(requestConfigVariables []interface{}) []datadogV1.SyntheticsConfigVariable {
	configVariables := make([]datadogV1.SyntheticsConfigVariable, 0)

	for _, v := range requestConfigVariables {
		variableMap := v.(map[string]interface{})
		variable := datadogV1.SyntheticsConfigVariable{}

		variable.SetName(variableMap["name"].(string))
		variable.SetType(datadogV1.SyntheticsConfigVariableType(variableMap["type"].(string)))

		if variable.GetType() != "global" {
			variable.SetPattern(variableMap["pattern"].(string))
			variable.SetExample(variableMap["example"].(string))
			variable.SetSecure(variableMap["secure"].(bool))
		}

		if variableMap["id"] != "" {
			variable.SetId(variableMap["id"].(string))
		}
		configVariables = append(configVariables, variable)
	}

	return configVariables
}

func buildTerraformConfigVariables(configVariables []datadogV1.SyntheticsConfigVariable, oldConfigVariables []interface{}) []map[string]interface{} {
	localConfigVariables := make([]map[string]interface{}, len(configVariables))
	for i, configVariable := range configVariables {
		localVariable := make(map[string]interface{})
		if v, ok := configVariable.GetTypeOk(); ok {
			localVariable["type"] = *v
		}
		if v, ok := configVariable.GetNameOk(); ok {
			localVariable["name"] = *v
		}
		if v, ok := configVariable.GetSecureOk(); ok {
			localVariable["secure"] = *v
		}

		if configVariable.GetType() != "global" {
			// If the variable is secure, the example and pattern are not returned by the API,
			// so we need to keep the values from the terraform config.
			if v, ok := localVariable["secure"].(bool); ok && v {
				// There is no previous state to fallback on during import
				if i < len(oldConfigVariables) && oldConfigVariables[i] != nil {
					localVariable["example"] = oldConfigVariables[i].(map[string]interface{})["example"].(string)
					localVariable["pattern"] = oldConfigVariables[i].(map[string]interface{})["pattern"].(string)
				}
			} else {
				if v, ok := configVariable.GetExampleOk(); ok {
					localVariable["example"] = *v
				}
				if v, ok := configVariable.GetPatternOk(); ok {
					localVariable["pattern"] = *v
				}
			}
		}
		if v, ok := configVariable.GetIdOk(); ok {
			localVariable["id"] = *v
		}
		localConfigVariables[i] = localVariable
	}
	return localConfigVariables
}

func buildDatadogExtractedValues(stepExtractedValues []interface{}) []datadogV1.SyntheticsParsingOptions {
	values := make([]datadogV1.SyntheticsParsingOptions, len(stepExtractedValues))

	for i, extractedValue := range stepExtractedValues {
		extractedValueMap := extractedValue.(map[string]interface{})
		value := datadogV1.SyntheticsParsingOptions{}

		value.SetName(extractedValueMap["name"].(string))
		value.SetType(datadogV1.SyntheticsLocalVariableParsingOptionsType(extractedValueMap["type"].(string)))
		if extractedValueMap["field"] != "" {
			value.SetField(extractedValueMap["field"].(string))
		}

		valueParsers := extractedValueMap["parser"].([]interface{})
		valueParser := valueParsers[0].(map[string]interface{})

		parser := datadogV1.SyntheticsVariableParser{}
		parser.SetType(datadogV1.SyntheticsGlobalVariableParserType(valueParser["type"].(string)))
		if valueParser["value"] != "" {
			parser.SetValue(valueParser["value"].(string))
		}

		value.SetParser(parser)

		if secure, ok := extractedValueMap["secure"].(bool); ok {
			value.SetSecure(secure)
		}

		values[i] = value
	}

	return values
}

func buildTerraformExtractedValues(extractedValues []datadogV1.SyntheticsParsingOptions) []map[string]interface{} {
	localExtractedValues := make([]map[string]interface{}, len(extractedValues))

	for i, extractedValue := range extractedValues {
		localExtractedValue := make(map[string]interface{})
		localExtractedValue["name"] = extractedValue.GetName()
		localExtractedValue["type"] = string(extractedValue.GetType())
		localExtractedValue["field"] = extractedValue.GetField()
		localExtractedValue["secure"] = extractedValue.GetSecure()

		parser := extractedValue.GetParser()
		localParser := make(map[string]interface{})
		localParser["type"] = string(parser.GetType())
		localParser["value"] = parser.GetValue()
		localExtractedValue["parser"] = []map[string]interface{}{localParser}

		localExtractedValues[i] = localExtractedValue
	}

	return localExtractedValues
}

func buildDatadogRequestCertificates(clientCertContent string, clientCertFilename string, clientKeyContent string, clientKeyFilename string) datadogV1.SyntheticsTestRequestCertificate {
	cert := datadogV1.SyntheticsTestRequestCertificateItem{}
	key := datadogV1.SyntheticsTestRequestCertificateItem{}

	if clientCertContent != "" {
		// only set the certificate content if it is not an already hashed string
		// this is needed for the update function that receives the data from the state
		// and not from the config. So we get a hash of the certificate and not it's real
		// value.
		if isHash := isCertHash(clientCertContent); !isHash {
			cert.SetContent(clientCertContent)
		}
	}
	if clientCertFilename != "" {
		cert.SetFilename(clientCertFilename)
	}

	if clientKeyContent != "" {
		// only set the key content if it is not an already hashed string
		if isHash := isCertHash(clientKeyContent); !isHash {
			key.SetContent(clientKeyContent)
		}
	}
	if clientKeyFilename != "" {
		key.SetFilename(clientKeyFilename)
	}

	return datadogV1.SyntheticsTestRequestCertificate{
		Cert: &cert,
		Key:  &key,
	}
}

func buildTerraformRequestCertificates(clientCertificate datadogV1.SyntheticsTestRequestCertificate, oldClientCertificates []interface{}) map[string][]map[string]string {
	localCertificate := make(map[string][]map[string]string)
	localCertificate["cert"] = make([]map[string]string, 1)
	localCertificate["cert"][0] = make(map[string]string)
	localCertificate["key"] = make([]map[string]string, 1)
	localCertificate["key"][0] = make(map[string]string)

	cert := clientCertificate.GetCert()
	localCertificate["cert"][0]["filename"] = cert.GetFilename()

	key := clientCertificate.GetKey()
	localCertificate["key"][0]["filename"] = key.GetFilename()

	// the content of client certificate is write-only so it will not be returned by the API.
	// To avoid useless diff but also prevent storing the value in clear in the state
	// we store a hash of the value.
	if len(oldClientCertificates) > 0 {
		if configCertificateContent, ok := oldClientCertificates[0].(map[string]interface{})["cert"].([]interface{})[0].(map[string]interface{})["content"].(string); ok {
			localCertificate["cert"][0]["content"] = getCertificateStateValue(configCertificateContent)
		}
		if configKeyContent, ok := oldClientCertificates[0].(map[string]interface{})["key"].([]interface{})[0].(map[string]interface{})["content"].(string); ok {
			localCertificate["key"][0]["content"] = getCertificateStateValue(configKeyContent)
		}
	}

	return localCertificate
}

func buildDatadogTestOptions(d *schema.ResourceData) *datadogV1.SyntheticsTestOptions {
	options := datadogV1.NewSyntheticsTestOptions()

	if attr, ok := d.GetOk("options_list"); ok && attr != nil {
		// common browser and API tests options
		if attr, ok := d.GetOk("options_list.0.tick_every"); ok {
			options.SetTickEvery(int64(attr.(int)))
		}
		if attr, ok := d.GetOk("options_list.0.http_version"); ok {
			options.SetHttpVersion(datadogV1.SyntheticsTestOptionsHTTPVersion(attr.(string)))
		}
		if attr, ok := d.GetOk("options_list.0.accept_self_signed"); ok {
			options.SetAcceptSelfSigned(attr.(bool))
		}
		if attr, ok := d.GetOk("options_list.0.check_certificate_revocation"); ok {
			options.SetCheckCertificateRevocation(attr.(bool))
		}
		if attr, ok := d.GetOk("options_list.0.min_location_failed"); ok {
			options.SetMinLocationFailed(int64(attr.(int)))
		}
		if attr, ok := d.GetOk("options_list.0.min_failure_duration"); ok {
			options.SetMinFailureDuration(int64(attr.(int)))
		}
		if attr, ok := d.GetOk("options_list.0.follow_redirects"); ok {
			options.SetFollowRedirects(attr.(bool))
		}
		if attr, ok := d.GetOk("options_list.0.allow_insecure"); ok {
			options.SetAllowInsecure(attr.(bool))
		}

		if rawScheduling, ok := d.GetOk("options_list.0.scheduling"); ok {
			optionsScheduling := datadogV1.SyntheticsTestOptionsScheduling{}
			scheduling := rawScheduling.([]interface{})[0]
			if rawTimeframes, ok := scheduling.(map[string]interface{})["timeframes"]; ok {
				var timeFrames []datadogV1.SyntheticsTestOptionsSchedulingTimeframe
				for _, tf := range rawTimeframes.(*schema.Set).List() {
					timeframe := datadogV1.NewSyntheticsTestOptionsSchedulingTimeframe(
						int32(tf.(map[string]interface{})["day"].(int)),
						tf.(map[string]interface{})["from"].(string),
						tf.(map[string]interface{})["to"].(string),
					)
					timeFrames = append(timeFrames, *timeframe)
				}
				optionsScheduling.SetTimeframes(timeFrames)
			}
			if timezone, ok := scheduling.(map[string]interface{})["timezone"]; ok {
				optionsScheduling.SetTimezone(timezone.(string))
			}
			options.SetScheduling(optionsScheduling)
		}

		if retryRaw, ok := d.GetOk("options_list.0.retry"); ok {
			optionsRetry := datadogV1.SyntheticsTestOptionsRetry{}
			retry := retryRaw.([]interface{})[0]

			if count, ok := retry.(map[string]interface{})["count"]; ok {
				optionsRetry.SetCount(int64(count.(int)))
			}
			if interval, ok := retry.(map[string]interface{})["interval"]; ok {
				optionsRetry.SetInterval(float64(interval.(int)))
			}

			options.SetRetry(optionsRetry)
		}

		if monitorOptionsRaw, ok := d.GetOk("options_list.0.monitor_options"); ok {
			monitorOptions := monitorOptionsRaw.([]interface{})[0]
			optionsMonitorOptions := datadogV1.SyntheticsTestOptionsMonitorOptions{}

			if renotifyInterval, ok := monitorOptions.(map[string]interface{})["renotify_interval"]; ok {
				optionsMonitorOptions.SetRenotifyInterval(int64(renotifyInterval.(int)))
			}

			options.SetMonitorOptions(optionsMonitorOptions)
		}

		if monitorName, ok := d.GetOk("options_list.0.monitor_name"); ok {
			options.SetMonitorName(monitorName.(string))
		}

		if monitorPriority, ok := d.GetOk("options_list.0.monitor_priority"); ok {
			options.SetMonitorPriority(int32(monitorPriority.(int)))
		}

		if restricted_roles, ok := d.GetOk("options_list.0.restricted_roles"); ok {
			roles := []string{}
			for _, role := range restricted_roles.(*schema.Set).List() {
				roles = append(roles, role.(string))
			}
			options.SetRestrictedRoles(roles)
		}

		if ciRaw, ok := d.GetOk("options_list.0.ci"); ok {
			ci := ciRaw.([]interface{})[0]
			if testCiOptions, ok := ci.(map[string]interface{}); ok {
				ciOptions := datadogV1.SyntheticsTestCiOptions{}
				ciOptions.SetExecutionRule(datadogV1.SyntheticsTestExecutionRule(testCiOptions["execution_rule"].(string)))
				options.SetCi(ciOptions)
			}
		}

		if ignoreServerCertificateError, ok := d.GetOk("options_list.0.ignore_server_certificate_error"); ok {
			options.SetIgnoreServerCertificateError(ignoreServerCertificateError.(bool))
		}

		// browser tests specific options
		if attr, ok := d.GetOk("options_list.0.no_screenshot"); ok {
			options.SetNoScreenshot(attr.(bool))
		}

		if rum_settings, ok := d.GetOk("options_list.0.rum_settings.0"); ok {
			settings := rum_settings.(map[string]interface{})
			isEnabled := settings["is_enabled"]

			rumSettings := datadogV1.SyntheticsBrowserTestRumSettings{}

			if isEnabled == true {
				rumSettings.SetIsEnabled(true)

				if applicationId, ok := settings["application_id"]; ok {
					if len(applicationId.(string)) > 0 {
						rumSettings.SetApplicationId(applicationId.(string))
					}
				}

				if clientTokenId, ok := settings["client_token_id"]; ok {
					if clientTokenId.(int) != 0 {
						rumSettings.SetClientTokenId(int64(clientTokenId.(int)))
					}
				}
			} else {
				rumSettings.SetIsEnabled(false)
			}

			options.SetRumSettings(rumSettings)
		}

		if disableCsp, ok := d.GetOk("options_list.0.disable_csp"); ok {
			options.SetDisableCsp(disableCsp.(bool))
		}

		if disableCors, ok := d.GetOk("options_list.0.disable_cors"); ok {
			options.SetDisableCors(disableCors.(bool))
		}

		if initialNavigationTimeout, ok := d.GetOk("options_list.0.initial_navigation_timeout"); ok {
			options.SetInitialNavigationTimeout(int64(initialNavigationTimeout.(int)))
		}

		if attr, ok := d.GetOk("device_ids"); ok {
			var deviceIds []string
			for _, s := range attr.([]interface{}) {
				deviceIds = append(deviceIds, s.(string))
			}
			options.DeviceIds = deviceIds
		}
	}

	return options
}

func buildTerraformTestOptions(actualOptions datadogV1.SyntheticsTestOptions) []map[string]interface{} {
	localOptionsList := make(map[string]interface{})

	if actualOptions.HasFollowRedirects() {
		localOptionsList["follow_redirects"] = actualOptions.GetFollowRedirects()
	}
	if actualOptions.HasMinFailureDuration() {
		localOptionsList["min_failure_duration"] = actualOptions.GetMinFailureDuration()
	}
	if actualOptions.HasMinLocationFailed() {
		localOptionsList["min_location_failed"] = actualOptions.GetMinLocationFailed()
	}
	if actualOptions.HasTickEvery() {
		localOptionsList["tick_every"] = actualOptions.GetTickEvery()
	}
	if actualOptions.HasHttpVersion() {
		localOptionsList["http_version"] = actualOptions.GetHttpVersion()
	}
	if actualOptions.HasAcceptSelfSigned() {
		localOptionsList["accept_self_signed"] = actualOptions.GetAcceptSelfSigned()
	}
	if actualOptions.HasCheckCertificateRevocation() {
		localOptionsList["check_certificate_revocation"] = actualOptions.GetCheckCertificateRevocation()
	}
	if actualOptions.HasAllowInsecure() {
		localOptionsList["allow_insecure"] = actualOptions.GetAllowInsecure()
	}

	if actualOptions.HasScheduling() {
		scheduling := actualOptions.GetScheduling()
		timeFrames := scheduling.GetTimeframes()
		optionsListScheduling := make(map[string]interface{})
		optionsListSchedulingTimeframes := make([]map[string]interface{}, 0, len(timeFrames))
		for _, tf := range timeFrames {
			timeframe := make(map[string]interface{})
			timeframe["from"] = tf.GetFrom()
			timeframe["day"] = tf.GetDay()
			timeframe["to"] = tf.GetTo()
			optionsListSchedulingTimeframes = append(optionsListSchedulingTimeframes, timeframe)
		}
		optionsListScheduling["timeframes"] = optionsListSchedulingTimeframes
		optionsListScheduling["timezone"] = scheduling.GetTimezone()
		optionsListSchedulingList := []map[string]interface{}{optionsListScheduling}
		localOptionsList["scheduling"] = optionsListSchedulingList
	}

	if actualOptions.HasRetry() {
		retry := actualOptions.GetRetry()
		optionsListRetry := make(map[string]interface{})
		optionsListRetry["count"] = retry.GetCount()

		if interval, ok := retry.GetIntervalOk(); ok {
			optionsListRetry["interval"] = interval
		}

		localOptionsList["retry"] = []map[string]interface{}{optionsListRetry}
	}
	if actualOptions.HasMonitorOptions() {
		actualMonitorOptions := actualOptions.GetMonitorOptions()
		optionsListMonitorOptions := make(map[string]int64)
		shouldUpdate := false

		if actualMonitorOptions.HasRenotifyInterval() {
			optionsListMonitorOptions["renotify_interval"] = actualMonitorOptions.GetRenotifyInterval()
			shouldUpdate = true
		}

		if shouldUpdate {
			localOptionsList["monitor_options"] = []map[string]int64{optionsListMonitorOptions}
		}
	}
	if actualOptions.HasNoScreenshot() {
		localOptionsList["no_screenshot"] = actualOptions.GetNoScreenshot()
	}
	if actualOptions.HasMonitorName() {
		localOptionsList["monitor_name"] = actualOptions.GetMonitorName()
	}
	if actualOptions.HasMonitorPriority() {
		localOptionsList["monitor_priority"] = actualOptions.GetMonitorPriority()
	}
	if actualOptions.HasRestrictedRoles() {
		localOptionsList["restricted_roles"] = actualOptions.GetRestrictedRoles()
	}
	if actualOptions.HasCi() {
		actualCi := actualOptions.GetCi()
		ciOptions := make(map[string]interface{})
		ciOptions["execution_rule"] = actualCi.GetExecutionRule()

		localOptionsList["ci"] = []map[string]interface{}{ciOptions}
	}

	if rumSettings, ok := actualOptions.GetRumSettingsOk(); ok {
		localRumSettings := make(map[string]interface{})
		localRumSettings["is_enabled"] = rumSettings.GetIsEnabled()

		if rumSettings.HasApplicationId() {
			localRumSettings["application_id"] = rumSettings.GetApplicationId()
		}

		if rumSettings.HasClientTokenId() {
			localRumSettings["client_token_id"] = rumSettings.GetClientTokenId()
		}

		localOptionsList["rum_settings"] = []map[string]interface{}{localRumSettings}
	}
	if actualOptions.HasIgnoreServerCertificateError() {
		localOptionsList["ignore_server_certificate_error"] = actualOptions.GetIgnoreServerCertificateError()
	}
	if actualOptions.HasDisableCsp() {
		localOptionsList["disable_csp"] = actualOptions.GetDisableCsp()
	}
	if actualOptions.HasDisableCors() {
		localOptionsList["disable_cors"] = actualOptions.GetDisableCors()
	}
	if actualOptions.HasInitialNavigationTimeout() {
		localOptionsList["initial_navigation_timeout"] = actualOptions.GetInitialNavigationTimeout()
	}

	localOptionsLists := make([]map[string]interface{}, 1)
	localOptionsLists[0] = localOptionsList

	return localOptionsLists
}

func buildDatadogMobileTestOptions(d *schema.ResourceData) *datadogV1.SyntheticsMobileTestOptions {
	options := datadogV1.SyntheticsMobileTestOptions{}

	if mobile_options_list_attr, ok := d.GetOk("mobile_options_list"); ok && mobile_options_list_attr != nil {
		// Verbosity is also part of the options but it can not be set by users so we're not setting it here
		if attr, ok := d.GetOk("mobile_options_list.0.min_failure_duration"); ok {
			options.SetMinFailureDuration(int64(attr.(int)))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.tick_every"); ok {
			options.SetTickEvery(int64(attr.(int)))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.monitor_name"); ok {
			options.SetMonitorName(attr.(string))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.monitor_priority"); ok {
			options.SetMonitorPriority(int32(attr.(int)))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.default_step_timeout"); ok {
			options.SetDefaultStepTimeout(int32(attr.(int)))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.no_screenshot"); ok {
			options.SetNoScreenshot(attr.(bool))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.allow_application_crash"); ok {
			options.SetAllowApplicationCrash(attr.(bool))
		}
		if attr, ok := d.GetOk("mobile_options_list.0.disable_auto_accept_alert"); ok {
			options.SetDisableAutoAcceptAlert(attr.(bool))
		}

		if retryRaw, ok := d.GetOk("mobile_options_list.0.retry"); ok {
			optionsRetry := datadogV1.SyntheticsTestOptionsRetry{}
			retry := retryRaw.([]interface{})[0]
			if count, ok := retry.(map[string]interface{})["count"]; ok {
				optionsRetry.SetCount(int64(count.(int)))
			}
			if interval, ok := retry.(map[string]interface{})["interval"]; ok {
				optionsRetry.SetInterval(float64(interval.(int)))
			}
			options.SetRetry(optionsRetry)
		}

		if rawScheduling, ok := d.GetOk("mobile_options_list.0.scheduling"); ok {
			optionsScheduling := datadogV1.SyntheticsTestOptionsScheduling{}
			scheduling := rawScheduling.([]interface{})[0]
			if tfs, ok := scheduling.(map[string]interface{})["timeframes"]; ok {
				timeFrames := []datadogV1.SyntheticsTestOptionsSchedulingTimeframe{}
				for _, tf := range tfs.(*schema.Set).List() {
					timeframe := datadogV1.SyntheticsTestOptionsSchedulingTimeframe{}
					timeframe.SetDay(int32(tf.(map[string]interface{})["day"].(int)))
					timeframe.SetFrom(string(tf.(map[string]interface{})["from"].(string)))
					timeframe.SetTo(string(tf.(map[string]interface{})["to"].(string)))
					timeFrames = append(timeFrames, timeframe)
				}
				optionsScheduling.SetTimeframes(timeFrames)
			}
			if tz, ok := scheduling.(map[string]interface{})["timezone"]; ok {
				optionsScheduling.SetTimezone(tz.(string))
			}

			options.SetScheduling(optionsScheduling)
		}

		if monitorOptionsRaw, ok := d.GetOk("mobile_options_list.0.monitor_options"); ok {

			monitorOptions := monitorOptionsRaw.([]interface{})[0]

			optionsMonitorOptions := datadogV1.SyntheticsTestOptionsMonitorOptions{}
			if renotifyInterval, ok := monitorOptions.(map[string]interface{})["renotify_interval"]; ok {
				optionsMonitorOptions.SetRenotifyInterval(int64(renotifyInterval.(int)))
			}
			if escalationMessage, ok := monitorOptions.(map[string]interface{})["escalation_message"]; ok {
				optionsMonitorOptions.SetEscalationMessage(escalationMessage.(string))
			}
			if renotifyOccurrences, ok := monitorOptions.(map[string]interface{})["renotify_occurrences"]; ok {
				optionsMonitorOptions.SetRenotifyOccurrences(int64(renotifyOccurrences.(int)))
			}
			if notificationPresetName, ok := monitorOptions.(map[string]interface{})["notification_preset_name"]; ok {
				optionsMonitorOptions.SetNotificationPresetName(datadogV1.SyntheticsTestOptionsMonitorOptionsNotificationPresetName(notificationPresetName.(string)))
			}
			options.SetMonitorOptions(optionsMonitorOptions)
		}

		if restricted_roles, ok := d.GetOk("mobile_options_list.0.restricted_roles"); ok {

			roles := []string{}
			for _, role := range restricted_roles.(*schema.Set).List() {
				roles = append(roles, role.(string))
			}
			options.SetRestrictedRoles(roles)
		}

		if bindings, ok := d.GetOk("mobile_options_list.0.bindings"); ok {
			optionsBindings := []datadogV1.SyntheticsTestRestrictionPolicyBinding{}
			for _, b := range bindings.([]interface{}) {
				binding := datadogV1.NewSyntheticsTestRestrictionPolicyBinding()
				if ps, ok := b.(map[string]interface{})["principals"]; ok {
					principals := []string{}
					for _, p := range ps.([]interface{}) {
						principals = append(principals, p.(string))
					}
					binding.SetPrincipals(principals)
				}
				if r, ok := b.(map[string]interface{})["relation"]; ok {
					binding.SetRelation(datadogV1.SyntheticsTestRestrictionPolicyBindingRelation(r.(string)))
				}
				optionsBindings = append(optionsBindings, *binding)
			}
			options.SetBindings(optionsBindings)
		}

		if rawCi, ok := d.GetOk("mobile_options_list.0.ci"); ok {
			ci := rawCi.([]interface{})[0]
			if testCiOptions, ok := ci.(map[string]interface{}); ok {
				ciOptions := datadogV1.SyntheticsTestCiOptions{}
				ciOptions.SetExecutionRule(datadogV1.SyntheticsTestExecutionRule(testCiOptions["execution_rule"].(string)))
				options.SetCi(ciOptions)
			}
		}

		if deviceIds, ok := d.GetOk("mobile_options_list.0.device_ids"); ok {
			optionsDeviceIds := []string{}
			for _, s := range deviceIds.([]interface{}) {
				optionsDeviceIds = append(optionsDeviceIds, s.(string))
			}
			options.SetDeviceIds(optionsDeviceIds)
		}

		if rawMobileApplication, ok := d.GetOk("mobile_options_list.0.mobile_application"); ok {
			mobileApplication := rawMobileApplication.([]interface{})[0]
			optionsMobileApplication := datadogV1.SyntheticsMobileTestsMobileApplication{}

			if s, ok := mobileApplication.(map[string]interface{})["application_id"]; ok {
				optionsMobileApplication.SetApplicationId(s.(string))
			}
			if s, ok := mobileApplication.(map[string]interface{})["reference_id"]; ok {
				optionsMobileApplication.SetReferenceId(s.(string))
			}
			if s, ok := mobileApplication.(map[string]interface{})["reference_type"]; ok {
				optionsMobileApplication.SetReferenceType(datadogV1.SyntheticsMobileTestsMobileApplicationReferenceType(s.(string)))
			}
			options.SetMobileApplication(optionsMobileApplication)
		}
	}

	return &options
}

func buildTerraformMobileTestOptions(actualOptions datadogV1.SyntheticsMobileTestOptions) []map[string]interface{} {
	localOptionsList := make(map[string]interface{})

	if actualOptions.HasMinFailureDuration() {
		localOptionsList["min_failure_duration"] = actualOptions.GetMinFailureDuration()
	}
	if attr, ok := actualOptions.GetTickEveryOk(); ok {
		localOptionsList["tick_every"] = attr
	}
	if actualOptions.HasMonitorName() {
		localOptionsList["monitor_name"] = actualOptions.GetMonitorName()
	}
	if actualOptions.HasMonitorPriority() {
		localOptionsList["monitor_priority"] = actualOptions.GetMonitorPriority()
	}
	if actualOptions.HasRestrictedRoles() {
		localOptionsList["restricted_roles"] = actualOptions.GetRestrictedRoles()
	}
	if actualOptions.HasDefaultStepTimeout() {
		localOptionsList["default_step_timeout"] = actualOptions.GetDefaultStepTimeout()
	}
	if actualOptions.HasNoScreenshot() {
		localOptionsList["no_screenshot"] = actualOptions.GetNoScreenshot()
	}
	if actualOptions.HasAllowApplicationCrash() {
		localOptionsList["allow_application_crash"] = actualOptions.GetAllowApplicationCrash()
	}
	if actualOptions.HasDisableAutoAcceptAlert() {
		localOptionsList["disable_auto_accept_alert"] = actualOptions.GetDisableAutoAcceptAlert()
	}

	if actualOptions.HasRetry() {
		retry := actualOptions.GetRetry()
		optionsListRetry := make(map[string]interface{})
		optionsListRetry["count"] = retry.GetCount()

		if interval, ok := retry.GetIntervalOk(); ok {
			optionsListRetry["interval"] = interval
		}

		localOptionsList["retry"] = []map[string]interface{}{optionsListRetry}
	}

	if actualOptions.HasScheduling() {
		scheduling := actualOptions.GetScheduling()
		timeFrames := scheduling.GetTimeframes()
		optionsListScheduling := make(map[string]interface{})
		optionsListSchedulingTimeframes := make([]map[string]interface{}, 0, len(timeFrames))
		for _, tf := range timeFrames {
			timeframe := make(map[string]interface{})
			timeframe["from"] = tf.GetFrom()
			timeframe["day"] = tf.GetDay()
			timeframe["to"] = tf.GetTo()
			optionsListSchedulingTimeframes = append(optionsListSchedulingTimeframes, timeframe)
		}
		optionsListScheduling["timeframes"] = optionsListSchedulingTimeframes
		optionsListScheduling["timezone"] = scheduling.GetTimezone()
		optionsListSchedulingList := []map[string]interface{}{optionsListScheduling}
		localOptionsList["scheduling"] = optionsListSchedulingList
	}

	if actualOptions.HasMonitorOptions() {
		actualMonitorOptions := actualOptions.GetMonitorOptions()
		optionsListMonitorOptions := make(map[string]interface{})
		shouldUpdate := false

		if actualMonitorOptions.HasRenotifyInterval() {
			optionsListMonitorOptions["renotify_interval"] = actualMonitorOptions.GetRenotifyInterval()
			shouldUpdate = true
		}
		if actualMonitorOptions.HasEscalationMessage() {
			optionsListMonitorOptions["escalation_message"] = actualMonitorOptions.GetEscalationMessage()
		}
		if actualMonitorOptions.HasRenotifyOccurrences() {
			optionsListMonitorOptions["renotify_occurrences"] = actualMonitorOptions.GetRenotifyOccurrences()
		}
		if actualMonitorOptions.HasNotificationPresetName() {
			optionsListMonitorOptions["notification_preset_name"] = actualMonitorOptions.GetNotificationPresetName()
		}

		if shouldUpdate {
			localOptionsList["monitor_options"] = []map[string]interface{}{optionsListMonitorOptions}
		}
	}

	if actualOptions.HasBindings() {
		actualBindings := actualOptions.GetBindings()
		optionsListBindings := make([]map[string]interface{}, 0, len(actualBindings))
		for _, binding := range actualBindings {
			optionsListBindingsItem := make(map[string]interface{})

			if binding.HasPrincipals() {
				actualBindingsItemsPrincipals := binding.GetPrincipals()
				optionsListBindingsItemsPrincipals := make([]string, 0, len(actualBindingsItemsPrincipals))
				for _, principals := range actualBindingsItemsPrincipals {
					optionsListBindingsItemsPrincipals = append(optionsListBindingsItemsPrincipals, principals)
				}
				optionsListBindingsItem["principals"] = optionsListBindingsItemsPrincipals
			}

			if binding.HasRelation() {
				optionsListBindingsItem["relation"] = binding.GetRelation()
			}

			optionsListBindings = append(optionsListBindings, optionsListBindingsItem)
		}
		localOptionsList["bindings"] = optionsListBindings
	}

	if actualOptions.HasCi() {
		actualCi := actualOptions.GetCi()
		ciOptions := make(map[string]interface{})
		ciOptions["execution_rule"] = actualCi.GetExecutionRule()

		localOptionsList["ci"] = []map[string]interface{}{ciOptions}
	}

	if _, ok := actualOptions.GetDeviceIdsOk(); ok {
		actualDeviceIds := actualOptions.GetDeviceIds()
		optionsListDeviceIds := make([]string, 0, len(actualDeviceIds))
		for _, device_id := range actualDeviceIds {
			optionsListDeviceIds = append(optionsListDeviceIds, string(device_id))
		}
		localOptionsList["device_ids"] = optionsListDeviceIds
	}

	if _, ok := actualOptions.GetMobileApplicationOk(); ok {
		actualMobileApplication := actualOptions.GetMobileApplication()
		optionsListMobileApplication := make(map[string]interface{})

		if _, ok := actualMobileApplication.GetApplicationIdOk(); ok {
			optionsListMobileApplication["application_id"] = actualMobileApplication.GetApplicationId()
		}
		if _, ok := actualMobileApplication.GetReferenceIdOk(); ok {
			optionsListMobileApplication["reference_id"] = actualMobileApplication.GetReferenceId()
		}
		if _, ok := actualMobileApplication.GetReferenceTypeOk(); ok {
			optionsListMobileApplication["reference_type"] = actualMobileApplication.GetReferenceType()
		}

		localOptionsList["mobile_application"] = []map[string]interface{}{optionsListMobileApplication}
	}

	localOptionsLists := make([]map[string]interface{}, 1)
	localOptionsLists[0] = localOptionsList

	return localOptionsLists
}

func buildTerraformMobileTestSteps(steps []datadogV1.SyntheticsMobileStep) []map[string]interface{} { // TODO SYNTH-17172 make sure everything is working in this function
	var localSteps []map[string]interface{}

	for _, step := range steps {

		localStep := make(map[string]interface{})

		// These two and params are required fields
		localStep["name"] = step.GetName()
		localStep["type"] = string(step.GetType())

		if allowFailure, ok := step.GetAllowFailureOk(); ok {
			localStep["allow_failure"] = allowFailure
		}
		if isCritical, ok := step.GetIsCriticalOk(); ok {
			localStep["is_critical"] = isCritical
		}
		if hasNoScreenshot, ok := step.GetNoScreenshotOk(); ok {
			localStep["no_screenshot"] = hasNoScreenshot
		}
		if HasNewStepElement, ok := step.GetHasNewStepElementOk(); ok {
			localStep["has_new_step_element"] = HasNewStepElement
		}
		if publicId, ok := step.GetPublicIdOk(); ok {
			localStep["public_id"] = publicId
		}
		if timeout, ok := step.GetTimeoutOk(); ok {
			localStep["timeout"] = timeout
		}

		localParams := make(map[string]interface{})
		params := step.GetParams()

		if params.HasCheck() {
			localParams["check"] = params.GetCheck()
		}
		if params.HasDelay() {
			localParams["delay"] = params.GetDelay()
		}
		if params.HasDirection() {
			localParams["direction"] = params.GetDirection()
		}
		if params.HasElement() {
			element := params.GetElement()
			localElement := make([]map[string]interface{}, 1)
			localElement[0] = make(map[string]interface{})
			if element.HasContext() {
				localElement[0]["context"] = element.GetContext()
			}
			if element.HasContextType() {
				localElement[0]["context_type"] = element.GetContextType()
			}
			if element.HasElementDescription() {
				localElement[0]["element_description"] = element.GetElementDescription()
			}
			if element.HasMultiLocator() {
				localElement[0]["multi_locator"] = element.GetMultiLocator()
			}
			if element.HasRelativePosition() {
				relativePosition := element.GetRelativePosition()
				localRelativePosition := make([]map[string]interface{}, 1)
				localRelativePosition[0] = make(map[string]interface{})
				if relativePosition.HasX() {
					localRelativePosition[0]["x"] = relativePosition.GetX()
				}
				if relativePosition.HasY() {
					localRelativePosition[0]["y"] = relativePosition.GetY()
				}
				localElement[0]["relative_position"] = localRelativePosition
			}
			if element.HasTextContent() {
				localElement[0]["text_content"] = element.GetTextContent()
			}
			if element.HasUserLocator() {
				userLocator := element.GetUserLocator()
				localUserLocator := make([]map[string]interface{}, 1)
				localUserLocator[0] = make(map[string]interface{})
				if userLocator.HasFailTestOnCannotLocate() {
					localUserLocator[0]["fail_test_on_cannot_locate"] = userLocator.GetFailTestOnCannotLocate()
				}
				if userLocator.HasValues() {
					values := userLocator.GetValues()
					localValues := make([]map[string]interface{}, len(values))
					for i, valuesItem := range values {
						localValuesItem := make(map[string]interface{})
						if valuesItem.HasValue() {
							localValuesItem["value"] = valuesItem.GetValue()
						}
						if valuesItem.HasType() {
							localValuesItem["type"] = valuesItem.GetType()
						}
						localValues[i] = localValuesItem
					}

					localUserLocator[0]["values"] = localValues
				}
				localElement[0]["user_locator"] = localUserLocator
			}
			if element.HasViewName() {
				localElement[0]["view_name"] = element.GetViewName()
			}
			localParams["element"] = localElement
		}
		if params.HasEnabled() {
			localParams["enabled"] = params.GetEnabled()
		}
		if params.HasMaxScrolls() {
			localParams["maxScrolls"] = params.GetMaxScrolls()
		}
		if params.HasPositions() {
			positions := params.GetPositions()
			for i, positionsItem := range positions {
				localPositionsItem := make(map[string]interface{})
				if positionsItem.HasX() {
					localPositionsItem["x"] = positionsItem.GetX()
				}
				if positionsItem.HasY() {
					localPositionsItem["y"] = positionsItem.GetY()
				}
				positions[i] = datadogV1.SyntheticsMobileStepParamsPositionsItems{
					X: localPositionsItem["x"].(*float64),
					Y: localPositionsItem["y"].(*float64),
				}
			}
			localParams["positions"] = positions
		}
		if params.HasSubtestPublicId() {
			localParams["subtestPublicId"] = params.GetSubtestPublicId()
		}
		if params.HasValue() {
			value := params.GetValue()
			actualValue := value.GetActualInstance()
			localParams["value"] = actualValue
		}
		if params.HasVariable() {
			localParams["variable"] = params.GetVariable()
		}
		if params.HasWithEnter() {
			localParams["withEnter"] = params.GetWithEnter()
		}
		if params.HasX() {
			localParams["x"] = params.GetX()
		}
		if params.HasY() {
			localParams["y"] = params.GetY()
		}

		localStep["params"] = []interface{}{localParams}

		localSteps = append(localSteps, localStep)
	}

	return localSteps
}

func completeSyntheticsTestRequest(request datadogV1.SyntheticsTestRequest, requestHeaders map[string]interface{}, requestQuery map[string]interface{}, requestBasicAuths []interface{}, requestProxies []interface{}, requestMetadata map[string]interface{}) *datadogV1.SyntheticsTestRequest {
	if len(requestHeaders) > 0 {
		headers := make(map[string]string, len(requestHeaders))

		for k, v := range requestHeaders {
			headers[k] = v.(string)
		}

		request.SetHeaders(headers)
	}

	if len(requestQuery) > 0 {
		request.SetQuery(requestQuery)
	}

	if len(requestBasicAuths) > 0 {
		if requestBasicAuth, ok := requestBasicAuths[0].(map[string]interface{}); ok {
			request.SetBasicAuth(buildDatadogBasicAuth(requestBasicAuth))
		}
	}

	if len(requestProxies) > 0 {
		if requestProxy, ok := requestProxies[0].(map[string]interface{}); ok {
			request.SetProxy(buildDatadogTestRequestProxy(requestProxy))
		}
	}

	if len(requestMetadata) > 0 {
		metadata := make(map[string]string, len(requestMetadata))

		for k, v := range requestMetadata {
			metadata[k] = v.(string)
		}

		request.SetMetadata(metadata)
	}

	return &request
}

func buildTerraformTestRequest(request datadogV1.SyntheticsTestRequest) map[string]interface{} {
	localRequest := make(map[string]interface{})
	if request.HasBody() {
		localRequest["body"] = request.GetBody()
	}
	if request.HasBodyType() {
		localRequest["body_type"] = request.GetBodyType()
	}
	if request.HasMethod() {
		localRequest["method"] = convertToString(request.GetMethod())
	}
	if request.HasTimeout() {
		localRequest["timeout"] = request.GetTimeout()
	}
	if request.HasUrl() {
		localRequest["url"] = request.GetUrl()
	}
	if request.HasHost() {
		localRequest["host"] = request.GetHost()
	}
	if request.HasPort() {
		var port = request.GetPort()
		if port.SyntheticsTestRequestNumericalPort != nil {
			localRequest["port"] = strconv.FormatInt(*port.SyntheticsTestRequestNumericalPort, 10)
		} else if port.SyntheticsTestRequestVariablePort != nil {
			localRequest["port"] = *port.SyntheticsTestRequestVariablePort
		}
	}
	if request.HasDnsServer() {
		localRequest["dns_server"] = convertToString(request.GetDnsServer())
	}
	if request.HasDnsServerPort() {
		localRequest["dns_server_port"] = request.GetDnsServerPort()
	}
	if request.HasNoSavingResponseBody() {
		localRequest["no_saving_response_body"] = request.GetNoSavingResponseBody()
	}
	if request.HasNumberOfPackets() {
		localRequest["number_of_packets"] = request.GetNumberOfPackets()
	}
	if request.HasShouldTrackHops() {
		localRequest["should_track_hops"] = request.GetShouldTrackHops()
	}
	if request.HasServername() {
		localRequest["servername"] = request.GetServername()
	}
	if request.HasMessage() {
		localRequest["message"] = request.GetMessage()
	}
	if request.HasCallType() {
		localRequest["call_type"] = request.GetCallType()
	}
	if request.HasService() {
		localRequest["service"] = request.GetService()
	}
	if request.HasCertificateDomains() {
		localRequest["certificate_domains"] = request.GetCertificateDomains()
	}
	if request.HasPersistCookies() {
		localRequest["persist_cookies"] = request.GetPersistCookies()
	}
	if request.HasHttpVersion() {
		localRequest["http_version"] = request.GetHttpVersion()
	}
	if request.HasCompressedJsonDescriptor() {
		localRequest["proto_json_descriptor"] = decompressAndDecodeValue(request.GetCompressedJsonDescriptor())
	}

	if request.HasCompressedProtoFile() {
		localRequest["plain_proto_file"] = decompressAndDecodeValue(request.GetCompressedProtoFile())
	}

	return localRequest
}

func buildDatadogTestRequestProxy(requestProxy map[string]interface{}) datadogV1.SyntheticsTestRequestProxy {
	testRequestProxy := datadogV1.SyntheticsTestRequestProxy{}
	testRequestProxy.SetUrl(requestProxy["url"].(string))

	proxyHeaders := make(map[string]string, len(requestProxy["headers"].(map[string]interface{})))

	for k, v := range requestProxy["headers"].(map[string]interface{}) {
		proxyHeaders[k] = v.(string)
	}

	testRequestProxy.SetHeaders(proxyHeaders)

	return testRequestProxy
}

func buildTerraformTestRequestProxy(proxy datadogV1.SyntheticsTestRequestProxy) map[string]interface{} {
	localProxy := make(map[string]interface{})
	localProxy["url"] = proxy.GetUrl()
	localProxy["headers"] = proxy.GetHeaders()

	return localProxy
}

/*
 * Utils
 */

func compressAndEncodeValue(value string) string {
	var compressedValue bytes.Buffer
	zl := zlib.NewWriter(&compressedValue)
	zl.Write([]byte(value))
	zl.Close()
	encodedCompressedValue := b64.StdEncoding.EncodeToString(compressedValue.Bytes())
	return encodedCompressedValue
}

func decompressAndDecodeValue(value string) string {
	decodedValue, _ := b64.StdEncoding.DecodeString(value)
	decodedBytes := bytes.NewReader(decodedValue)
	zl, _ := zlib.NewReader(decodedBytes)
	defer zl.Close()
	compressedProtoFile, _ := io.ReadAll(zl)
	return string(compressedProtoFile)
}

func convertStepParamsValueForConfig(stepType interface{}, key string, value interface{}) interface{} {
	switch key {
	case "element", "email", "file", "files", "request":
		var result interface{}
		if err := utils.GetMetadataFromJSON([]byte(value.(string)), &result); err != nil {
			log.Printf("[ERROR] Error converting step param %s: %v", key, err)
		}
		return result

	case "playing_tab_id":
		result, _ := strconv.Atoi(value.(string))
		return result

	case "value":
		if stepType == datadogV1.SYNTHETICSSTEPTYPE_WAIT {
			result, _ := strconv.Atoi(value.(string))
			return result
		}

		return value

	case "variable":
		return value.([]interface{})[0]
	}

	return value
}

func convertStepParamsValueForState(key string, value interface{}) interface{} {
	switch key {
	case "element", "email", "file", "files", "request":
		result, _ := json.Marshal(value)
		return string(result)

	case "playing_tab_id", "value":
		return convertToString(value)

	case "variable":
		return []interface{}{value}
	}

	return value
}

func convertStepParamsKey(key string) string {
	switch key {
	case "click_type":
		return "clickType"

	case "clickType":
		return "click_type"

	case "playing_tab_id":
		return "playingTabId"

	case "playingTabId":
		return "playing_tab_id"

	case "subtest_public_id":
		return "subtestPublicId"

	case "subtestPublicId":
		return "subtest_public_id"

	case "with_click":
		return "withClick"

	case "withClick":
		return "with_click"
	}

	return key
}

func convertToString(i interface{}) string {
	switch v := i.(type) {
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		// TODO: manage target for JSON body assertions
		valStrr, err := json.Marshal(v)
		if err == nil {
			return string(valStrr)
		}
		return ""
	}
}

// get the sha256 of a client certificate content
// in some case where Terraform compares the state value
// we already get the hashed value so we don't need to
// hash it again
func getCertificateStateValue(content string) string {
	if isHash := isCertHash(content); isHash {
		return content
	}

	return utils.ConvertToSha256(content)
}

func getStepParams(stepMap map[string]interface{}, d *schema.ResourceData) map[string]interface{} {
	stepType := datadogV1.SyntheticsStepType(stepMap["type"].(string))

	params := make(map[string]interface{})
	stepParams := stepMap["params"].([]interface{})[0]
	stepTypeParams := getParamsKeysForStepType(stepType)

	includeElement := false
	for _, key := range stepTypeParams {
		if stepMap, ok := stepParams.(map[string]interface{}); ok && stepMap[key] != "" {
			convertedValue := convertStepParamsValueForConfig(stepType, key, stepMap[key])
			params[convertStepParamsKey(key)] = convertedValue
		}

		if key == "element" {
			includeElement = true
		}
	}

	stepElement := make(map[string]interface{})
	if stepParamsMap, ok := stepParams.(map[string]interface{}); ok {

		// Initialize the element with the values from the state
		if stepParamsElement, ok := stepParamsMap["element"]; ok {
			utils.GetMetadataFromJSON([]byte(stepParamsElement.(string)), &stepElement)
		}

		// When conciliating the config and the state, the provider is not updating the ML in the state as
		// a side effect of the diffSuppressFunc, but it nonetheless updates the other fields.
		// So after reordering the steps in the config, the state contains steps with mixed up MLs.
		// This propagates to the crafted request to update the test on the backend, and eventually mess up
		// the remote test.
		//
		// To fix this issue, the user can provide a local key for each step to track steps when reordering.
		// The provider can use the local key to reconcile the right ML into the right step.
		// To retrieve the right ML, this function needs to look for the step which has the same localKey
		// than the current step in the state, then in the config.
		// The right ML could be in the state when the user didn't provide it in the config, but the provider
		// keep it there anyway to keep track of it. Or it could be in the config when the user provided
		// it directly.
		//
		// In the following,
		// - GetRawState is used to retrieve the state of the resource before the reconciliation.
		//   It contains the ML when the user didn't provide it in the config.
		// - GetRawConfig is used to retrieve the config of the resource as written by the user.
		//   It contains the ML when the user provided it in the config.

		// Update the ML from the state, if found
		rawState := d.GetRawState()
		stateStepCount := 0
		stateSteps := cty.ListValEmpty(cty.DynamicPseudoType)
		if !rawState.IsNull() {
			stateSteps = rawState.GetAttr("browser_step")
			stateStepCount = stateSteps.LengthInt()
		}

		if stateStepCount > 0 {
			for i := range stateStepCount {
				stateStep := stateSteps.Index(cty.NumberIntVal(int64(i)))
				localKeyValue := stateStep.GetAttr("local_key")
				if localKeyValue.IsNull() {
					continue
				}

				localKey := localKeyValue.AsString()
				if localKey == stepMap["local_key"] {
					stepParamsValue := stateStep.GetAttr("params")
					if stepParamsValue.IsNull() {
						continue
					}

					stepParams := stepParamsValue.Index(cty.NumberIntVal(0))
					elementValue := stepParams.GetAttr("element")
					if elementValue.IsNull() {
						continue
					}
					element := elementValue.AsString()
					stateStepElement := make(map[string]interface{})
					utils.GetMetadataFromJSON([]byte(element), &stateStepElement)

					for key, value := range stateStepElement {
						stepElement[key] = value
					}
				}
			}
		}

		// Update the ML from the config, if found
		rawConfig := d.GetRawConfig()
		configStepCount := 0
		configSteps := cty.ListValEmpty(cty.DynamicPseudoType)
		if !rawConfig.IsNull() {
			configSteps = rawConfig.GetAttr("browser_step")
			configStepCount = configSteps.LengthInt()
		}

		if configStepCount > 0 {
			for i := range configStepCount {
				configStep := configSteps.Index(cty.NumberIntVal(int64(i)))
				localKeyValue := configStep.GetAttr("local_key")
				if localKeyValue.IsNull() {
					continue
				}

				localKey := localKeyValue.AsString()
				if localKey == stepMap["local_key"] {
					stepParamsValue := configStep.GetAttr("params")
					if stepParamsValue.IsNull() {
						continue
					}

					stepParams := stepParamsValue.Index(cty.NumberIntVal(0))
					elementValue := stepParams.GetAttr("element")
					if elementValue.IsNull() {
						continue
					}
					element := elementValue.AsString()
					configStepElement := make(map[string]interface{})
					utils.GetMetadataFromJSON([]byte(element), &configStepElement)

					for key, value := range configStepElement {
						stepElement[key] = value
					}
				}
			}
		}

		// If the step has a user locator in the config, set it in the stepElement as well
		if stepParamsMap["element_user_locator"] != "" {
			userLocatorsParams := stepParamsMap["element_user_locator"].([]interface{})

			if len(userLocatorsParams) != 0 {
				userLocatorParams := userLocatorsParams[0].(map[string]interface{})
				values := userLocatorParams["value"].([]interface{})
				userLocator := map[string]interface{}{
					"failTestOnCannotLocate": userLocatorParams["fail_test_on_cannot_locate"],
					"values":                 []map[string]interface{}{values[0].(map[string]interface{})},
				}

				stepElement["userLocator"] = userLocator
			}
		}

	}

	// If the step should contain an element, and it's not empty, add it to the params.
	// This is to avoid sending an empty element to the backend, as some steps have an optional element.
	if includeElement && len(stepElement) > 0 {
		params["element"] = stepElement
	}

	return params
}

func getParamsKeysForStepType(stepType datadogV1.SyntheticsStepType) []string {
	switch stepType {
	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_CURRENT_URL:
		return []string{"check", "value"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_ELEMENT_ATTRIBUTE:
		return []string{"attribute", "check", "element", "value"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_ELEMENT_CONTENT:
		return []string{"check", "element", "value"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_ELEMENT_PRESENT:
		return []string{"element"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_EMAIL:
		return []string{"email"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_FILE_DOWNLOAD:
		return []string{"file"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_FROM_JAVASCRIPT:
		return []string{"code", "element"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_PAGE_CONTAINS:
		return []string{"value"}

	case datadogV1.SYNTHETICSSTEPTYPE_ASSERT_PAGE_LACKS:
		return []string{"value"}

	case datadogV1.SYNTHETICSSTEPTYPE_CLICK:
		return []string{"click_type", "element"}

	case datadogV1.SYNTHETICSSTEPTYPE_EXTRACT_FROM_JAVASCRIPT:
		return []string{"code", "element", "variable"}

	case datadogV1.SYNTHETICSSTEPTYPE_EXTRACT_VARIABLE:
		return []string{"element", "variable"}

	case datadogV1.SYNTHETICSSTEPTYPE_GO_TO_EMAIL_LINK:
		return []string{"value"}

	case datadogV1.SYNTHETICSSTEPTYPE_GO_TO_URL:
		return []string{"value"}

	case datadogV1.SYNTHETICSSTEPTYPE_HOVER:
		return []string{"element"}

	case datadogV1.SYNTHETICSSTEPTYPE_PLAY_SUB_TEST:
		return []string{"playing_tab_id", "subtest_public_id"}

	case datadogV1.SYNTHETICSSTEPTYPE_PRESS_KEY:
		return []string{"modifiers", "value"}

	case datadogV1.SYNTHETICSSTEPTYPE_REFRESH:
		return []string{}

	case datadogV1.SYNTHETICSSTEPTYPE_RUN_API_TEST:
		return []string{"request"}

	case datadogV1.SYNTHETICSSTEPTYPE_SCROLL:
		return []string{"element", "x", "y"}

	case datadogV1.SYNTHETICSSTEPTYPE_SELECT_OPTION:
		return []string{"element", "value"}

	case datadogV1.SYNTHETICSSTEPTYPE_TYPE_TEXT:
		return []string{"delay", "element", "value"}

	case datadogV1.SYNTHETICSSTEPTYPE_UPLOAD_FILES:
		return []string{"element", "files", "with_click"}

	case datadogV1.SYNTHETICSSTEPTYPE_WAIT:
		return []string{"value"}
	}

	return []string{}
}

func getParamsKeysForMobileStepType(stepType datadogV1.SyntheticsMobileStepType) []string {
	switch stepType {
	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ASSERTELEMENTCONTENT:
		return []string{"check", "element", "value"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ASSERTSCREENCONTAINS:
		return []string{"value"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ASSERTSCREENLACKS:
		return []string{"value"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_DOUBLETAP:
		return []string{"element"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_EXTRACTVARIABLE:
		return []string{"element", "variable"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_FLICK:
		return []string{"position"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_OPENDEEPLINK:
		return []string{"value"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_PLAYSUBTEST:
		return []string{"subtest_public_id"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_PRESSBACK:
		return []string{}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_RESTARTAPPLICATION:
		return []string{}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ROTATE:
		return []string{"value"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_SCROLL:
		return []string{"element", "x", "y"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_SCROLLTOELEMENT:
		return []string{"element", "direction", "max_scrolls"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_TAP:
		return []string{"element"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_TOGGLEWIFI:
		return []string{"enabled"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_TYPETEXT:
		return []string{"value", "element", "delay", "with_enter"}

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_WAIT:
		return []string{"value"}
	}

	return []string{}
}

func buildDatadogParamsForMobileStep(stepType datadogV1.SyntheticsMobileStepType, stepParams map[string]interface{}) datadogV1.SyntheticsMobileStepParams {
	params := datadogV1.SyntheticsMobileStepParams{}
	switch stepType {
	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ASSERTELEMENTCONTENT:
		if stepParams["check"] != "" {
			params.SetCheck(datadogV1.SyntheticsCheckType(stepParams["check"].(string)))
		}
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(string)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueStringAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ASSERTSCREENCONTAINS:
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(string)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueStringAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ASSERTSCREENLACKS:
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(string)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueStringAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_DOUBLETAP:
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_EXTRACTVARIABLE:
		if len(stepParams["variable"].(map[string]interface{})) != 0 {
			paramsVarible := datadogV1.SyntheticsMobileStepParamsVariable{}
			paramsVarible.SetName(stepParams["variable"].(map[string]interface{})["name"].(string))
			paramsVarible.SetExample(stepParams["variable"].(map[string]interface{})["example"].(string))
			params.SetVariable(paramsVarible)
		}
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_FLICK:
		if len(stepParams["position"].(map[string]interface{})) != 0 {
			positions := []datadogV1.SyntheticsMobileStepParamsPositionsItems{}
			for _, position := range stepParams["position"].([]interface{}) {
				positionItem := datadogV1.SyntheticsMobileStepParamsPositionsItems{}
				positionItem.SetX(position.(map[string]interface{})["x"].(float64))
				positionItem.SetY(position.(map[string]interface{})["y"].(float64))

				positions = append(positions, positionItem)
			}

			params.SetPositions(positions)
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_OPENDEEPLINK:
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(string)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueStringAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_PLAYSUBTEST:
		if stepParams["subtest_public_id"] != "" {
			params.SetSubtestPublicId(stepParams["subtest_public_id"].(string))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_PRESSBACK:
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_RESTARTAPPLICATION:
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_ROTATE:
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(string)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueStringAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_SCROLL:
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		if stepParams["x"] != "" {
			params.SetX(stepParams["x"].(float64))
		}
		if stepParams["y"] != "" {
			params.SetY(stepParams["y"].(float64))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_SCROLLTOELEMENT:
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		if stepParams["direction"] != "" {
			params.SetDirection(datadogV1.SyntheticsMobileStepParamsDirection(stepParams["direction"].(string)))
		}
		if stepParams["max_scrolls"] != "" {
			params.SetMaxScrolls(stepParams["max_scrolls"].(int64))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_TAP:
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_TOGGLEWIFI:
		if stepParams["enabled"] != "" {
			params.SetEnabled(stepParams["enabled"].(bool))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_TYPETEXT:
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(string)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueStringAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		stepParam := stepParams["element"].([]interface{})[0].(map[string]interface{})
		if len(stepParam) != 0 {
			params.SetElement(buildDatadogParamsElementForMobileStep(stepParam))
		}
		if stepParams["delay"] != "" {
			params.SetDelay(stepParams["delay"].(int64))
		}
		if stepParams["with_enter"] != "" {
			params.SetWithEnter(stepParams["with_enter"].(bool))
		}
		return params

	case datadogV1.SYNTHETICSMOBILESTEPTYPE_WAIT:
		if stepParams["value"] != "" {
			stepParamsValue := stepParams["value"].(int64)
			params.SetValue(datadogV1.SyntheticsMobileStepParamsValueNumberAsSyntheticsMobileStepParamsValue(&stepParamsValue))
		}
		return params
	}

	return params
}

func buildDatadogParamsElementForMobileStep(stepParamsElements map[string]interface{}) datadogV1.SyntheticsMobileStepParamsElement {
	elements := datadogV1.SyntheticsMobileStepParamsElement{}

	if len(stepParamsElements["multi_locator"].(map[string]interface{})) != 0 {
		elements.SetMultiLocator(stepParamsElements["multi_locator"].(string))
	}
	if stepParamsElements["context"].(string) != "" {
		elements.SetContext(stepParamsElements["context"].(string))
	}
	if stepParamsElements["context_type"].(string) != "" {
		elements.SetContextType(datadogV1.SyntheticsMobileStepParamsElementContextType(stepParamsElements["context_type"].(string)))
	}
	stepParamsElement := stepParamsElements["user_locator"].([]interface{})[0].(map[string]interface{})
	if len(stepParamsElement) != 0 {

		userLocator := datadogV1.SyntheticsMobileStepParamsElementUserLocator{}
		userLocatorValues := []datadogV1.SyntheticsMobileStepParamsElementUserLocatorValuesItems{}

		userLocator.SetFailTestOnCannotLocate(stepParamsElement["fail_test_on_cannot_locate"].(bool))

		for _, value := range stepParamsElement["values"].([]interface{}) {
			userLocatorValue := datadogV1.SyntheticsMobileStepParamsElementUserLocatorValuesItems{}
			userLocatorValue.SetType(datadogV1.SyntheticsMobileStepParamsElementUserLocatorValuesItemsType(value.(map[string]interface{})["type"].(string)))
			userLocatorValue.SetValue(value.(map[string]interface{})["value"].(string))

			userLocatorValues = append(userLocatorValues, userLocatorValue)
		}

		userLocator.SetValues(userLocatorValues)
		elements.SetUserLocator(userLocator)
	}
	if stepParamsElements["element_description"].(string) != "" {
		elements.SetElementDescription(stepParamsElements["element_description"].(string))
	}
	elementRelativePosition := stepParamsElements["relative_position"].([]interface{})[0].(map[string]interface{})
	if len(elementRelativePosition) != 0 {
		relativePosition := datadogV1.SyntheticsMobileStepParamsElementRelativePosition{}
		relativePosition.SetX(elementRelativePosition["x"].(float64))
		relativePosition.SetY(elementRelativePosition["y"].(float64))

		elements.SetRelativePosition(relativePosition)
	}
	if stepParamsElements["text_content"].(string) != "" {
		elements.SetTextContent(stepParamsElements["text_content"].(string))
	}
	if stepParamsElements["view_name"].(string) != "" {
		elements.SetViewName(stepParamsElements["view_name"].(string))
	}

	return elements
}

func getSyntheticsTestType(d *schema.ResourceData) *datadogV1.SyntheticsTestDetailsType {
	v := datadogV1.SyntheticsTestDetailsType(d.Get("type").(string))
	return &v
}

func isCertHash(content string) bool {
	// a sha256 hash consists of 64 hexadecimal characters
	isHash, _ := regexp.MatchString("^[A-Fa-f0-9]{64}$", content)

	return isHash
}

func isTargetOfTypeInt(assertionType datadogV1.SyntheticsAssertionType, assertionOperator datadogV1.SyntheticsAssertionOperator) bool {
	for _, intTargetAssertionType := range []datadogV1.SyntheticsAssertionType{
		datadogV1.SYNTHETICSASSERTIONTYPE_RESPONSE_TIME,
		datadogV1.SYNTHETICSASSERTIONTYPE_CERTIFICATE,
		datadogV1.SYNTHETICSASSERTIONTYPE_LATENCY,
		datadogV1.SYNTHETICSASSERTIONTYPE_PACKETS_RECEIVED,
		datadogV1.SYNTHETICSASSERTIONTYPE_NETWORK_HOP,
		datadogV1.SYNTHETICSASSERTIONTYPE_GRPC_HEALTHCHECK_STATUS,
	} {
		if assertionType == intTargetAssertionType {
			return true
		}
	}
	if assertionType == datadogV1.SYNTHETICSASSERTIONTYPE_STATUS_CODE &&
		(assertionOperator == datadogV1.SYNTHETICSASSERTIONOPERATOR_IS || assertionOperator == datadogV1.SYNTHETICSASSERTIONOPERATOR_IS_NOT) {
		return true
	}
	return false
}

func validateSyntheticsAssertionOperator(val interface{}, key string) (warns []string, errs []error) {
	_, err := datadogV1.NewSyntheticsAssertionOperatorFromValue(val.(string))
	if err != nil {
		_, err2 := datadogV1.NewSyntheticsAssertionJSONPathOperatorFromValue(val.(string))
		_, err3 := datadogV1.NewSyntheticsAssertionJSONSchemaOperatorFromValue(val.(string))
		_, err4 := datadogV1.NewSyntheticsAssertionXPathOperatorFromValue(val.(string))
		_, err5 := datadogV1.NewSyntheticsAssertionBodyHashOperatorFromValue(val.(string))

		if err2 == nil || err3 == nil || err4 == nil || err5 == nil {
			return
		} else {
			errs = append(errs, err, err2, err3, err4, err5)
		}
	}
	return
}

func getConfigCertAndKeyContent(d *schema.ResourceData, stepIndex int) (*string, *string) {
	// For security reasons, the certificate and keys can't be stored in the terraform state. It needs to stay in clear only in the config. This function retrieve the certificate from the terraform config, rather than the state.
	// To retrieve the certificate and key, we first need to build the paths to the cert and key content, and then apply these paths to the rawConfig.

	rawConfig := d.GetRawConfig()
	basePath := cty.GetAttrPath("api_step").
		Index(cty.NumberIntVal(int64(stepIndex))).
		GetAttr("request_client_certificate").
		Index(cty.NumberIntVal(0))

	// Get the certificate
	certContentPath := basePath.
		GetAttr("cert").
		Index(cty.NumberIntVal(0)).
		GetAttr("content")
	certContent, err := certContentPath.Apply(rawConfig)
	if err != nil || !certContent.IsKnown() || certContent.IsNull() {
		return nil, nil
	}
	certContentString := certContent.AsString()

	// Get the key
	keyContentPath := basePath.
		GetAttr("key").
		Index(cty.NumberIntVal(0)).
		GetAttr("content")
	keyContent, err := keyContentPath.Apply(rawConfig)
	if err != nil || !keyContent.IsKnown() || keyContent.IsNull() {
		return nil, nil
	}
	keyContentString := keyContent.AsString()

	return &certContentString, &keyContentString
}

func getCertAndKeyFromMap(certAndKey map[string]interface{}) (map[string]interface{}, map[string]interface{}) {

	clientCerts := certAndKey["cert"].([]interface{})
	clientKeys := certAndKey["key"].([]interface{})
	clientCert := clientCerts[0].(map[string]interface{})
	clientKey := clientKeys[0].(map[string]interface{})

	return clientCert, clientKey
}
