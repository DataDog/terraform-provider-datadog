// For more info about writing custom provider: https://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	_nethttp "net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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
		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Synthetics test type (`api` or `browser`).",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestDetailsTypeFromValue),
			},
			"subtype": {
				Description: "When `type` is `api`, choose from `http`, `ssl`, `tcp` or `dns`. Defaults to `http`.",
				Type:        schema.TypeString,
				Optional:    true,
				DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
					if d.Get("type") == "api" && old == "http" && new == "" {
						// defaults to http if type is api for retro-compatibility
						return true
					}
					return old == new
				},
				ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestDetailsSubTypeFromValue),
			},
			"request_definition": {
				Description: "The synthetics test request. Required if `type = \"api\"`.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem:        syntheticsTestRequest(),
			},
			"request_headers": {
				Description: "Header name and value map.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"request_query": {
				Description: "Query arguments name and value map.",
				Type:        schema.TypeMap,
				Optional:    true,
			},
			"request_basicauth": {
				Description: "The HTTP basic authentication credentials. Exactly one nested block is allowed with the structure below.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Description: "Username for authentication.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"password": {
							Description: "Password for authentication.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"request_client_certificate": {
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
			},
			"assertion": {
				Description: "Assertions used for the test. Multiple `assertion` blocks are allowed with the structure below.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description:  "Type of assertion. Choose from `body`, `header`, `responseTime`, `statusCode`. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsAssertionTypeFromValue),
							Required:     true,
						},
						"operator": {
							Description:  "Assertion operator. **Note** Only some combinations of `type` and `operator` are valid (please refer to [Datadog documentation](https://docs.datadoghq.com/api/latest/synthetics/#create-a-test)).",
							Type:         schema.TypeString,
							Required:     true,
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
						"targetjsonpath": {
							Description: "Expected structure if `operator` is `validatesJSONPath`. Exactly one nested block is allowed with the structure below.",
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
									"jsonpath": {
										Description: "The JSON path to assert.",
										Type:        schema.TypeString,
										Required:    true,
									},
									"targetvalue": {
										Description: "Expected matching value.",
										Type:        schema.TypeString,
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
			"browser_variable": syntheticsBrowserVariable(),
			"config_variable":  syntheticsConfigVariable(),
			"device_ids": {
				Description: "Array with the different device IDs used to run the test. Allowed enum values: `laptop_large`, `tablet`, `mobile_small` (only available for `browser` tests).",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsDeviceIDFromValue),
				},
			},
			"locations": {
				Description: "Array of locations used to run the test. Refer to [Datadog documentation](https://docs.datadoghq.com/synthetics/api_test/#request) for available locations (e.g. `aws:eu-central-1`).",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"options_list": syntheticsTestOptionsList(),
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
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Description:  "Define whether you want to start (`live`) or pause (`paused`) a Synthetic test. Allowed enum values: `live`, `paused`",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTestPauseStatusFromValue),
			},
			"monitor_id": {
				Description: "ID of the monitor associated with the Datadog synthetics test.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"browser_step": syntheticsTestBrowserStep(),
		},
	}
}

func syntheticsTestRequest() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"method": {
				Description:  "The HTTP method. One of `DELETE`, `GET`, `HEAD`, `OPTIONS`, `PATCH`, `POST`, `PUT`.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validators.ValidateEnumValue(datadogV1.NewHTTPMethodFromValue),
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
			"timeout": {
				Description: "Timeout in seconds for the test. Defaults to `60`.",
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
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"dns_server": {
				Description: "DNS server to use for DNS tests (`subtype = \"dns\"`).",
				Type:        schema.TypeString,
				Optional:    true,
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

func syntheticsTestOptionsList() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allow_insecure": {
					Description: "Allows loading insecure content for an HTTP test.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"follow_redirects": {
					Description: "For API HTTP test, whether or not the test should follow redirects.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"tick_every": {
					Description:  "How often the test should run (in seconds). Current possible values are `900`, `1800`, `3600`, `21600`, `43200`, `86400`, `604800` plus `60` for API tests or `300` for browser tests.",
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsTickIntervalFromValue),
				},
				"accept_self_signed": {
					Description: "For SSL test, whether or not the test should allow self signed certificates.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"min_location_failed": {
					Description: "Minimum number of locations in failure required to trigger an alert. Default is `1`.",
					Type:        schema.TypeInt,
					Default:     1,
					Optional:    true,
				},
				"min_failure_duration": {
					Description: "Minimum amount of time in failure required to trigger an alert. Default is `0`.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"monitor_options": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"renotify_interval": {
								Description: "Specify a renotification frequency.",
								Type:        schema.TypeInt,
								Default:     0,
								Optional:    true,
							},
						},
					},
				},
				"retry": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"count": {
								Description: "Number of retries needed to consider a location as failed before sending a notification alert.",
								Type:        schema.TypeInt,
								Default:     0,
								Optional:    true,
							},
							"interval": {
								Description: "Interval between a failed test and the next retry in milliseconds.",
								Type:        schema.TypeInt,
								Default:     300,
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsTestBrowserStep() *schema.Schema {
	var paramsSchema schema.Schema
	paramsSchema = syntheticsBrowserStepParams()

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
				"type": {
					Description:  "Type of the step. Refer to [Datadog documentation](https://docs.datadoghq.com/api/v1/synthetics/#create-a-test) for the complete list of available types.",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsStepTypeFromValue),
				},
				"allow_failure": {
					Description: "Determines if the step should be allowed to fail.",
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
					Description:  "Check type to use for an assertion step.",
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsCheckTypeFromValue),
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
					Description: "Element to use for the step, json encoded string.",
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
				"email": {
					Description: `Details of the email for an "assert email" step.`,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"file": {
					Description: `For an "assert download" step.`,
					Type:        schema.TypeString,
					Optional:    true,
				},
				"files": {
					Description: `Details of the files for an "upload files" step, json encoded string.`,
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
								Default:  "Example of the extracted variable.",
								Type:     schema.TypeString,
								Optional: true,
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
			},
			"type": {
				Description:  "Type of browser test variable. Allowed enum values: `element`, `email`, `global`, `javascript`, `text`.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsBrowserVariableTypeFromValue),
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
					Description: "Example for the variable.",
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
				},
				"type": {
					Description:  "Type of test configuration variable. Allowed enum values: `text`.",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSyntheticsConfigVariableTypeFromValue),
				},
			},
		},
	}
}

func resourceDatadogSyntheticsTestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsTest := buildSyntheticsTestStruct(d)
	createdSyntheticsTest, _, err := datadogClientV1.SyntheticsApi.CreateTest(authV1).Body(*syntheticsTest).Execute()
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return utils.TranslateClientErrorDiag(err, "error creating synthetics test")
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsTest.GetPublicId())

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(ctx, d, meta)
}

func resourceDatadogSyntheticsTestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	var syntheticsTest datadogV1.SyntheticsTestDetails
	var err error
	var httpresp *_nethttp.Response

	if d.Get("type") == "browser" {
		syntheticsTest, httpresp, err = datadogClientV1.SyntheticsApi.GetBrowserTest(authV1, d.Id()).Execute()
	} else {
		syntheticsTest, httpresp, err = datadogClientV1.SyntheticsApi.GetTest(authV1, d.Id()).Execute()

		// re-fetch test if it was actually a browser but we didn't have the info earlier
		if syntheticsTest.GetType() == "browser" {
			syntheticsTest, _, err = datadogClientV1.SyntheticsApi.GetBrowserTest(authV1, d.Id()).Execute()
		}
	}

	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, "error getting synthetics test")
	}

	return updateSyntheticsTestLocalState(d, &syntheticsTest)
}

func resourceDatadogSyntheticsTestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsTest := buildSyntheticsTestStruct(d)
	if _, _, err := datadogClientV1.SyntheticsApi.UpdateTest(authV1, d.Id()).Body(*syntheticsTest).Execute(); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return utils.TranslateClientErrorDiag(err, "error updating synthetics test")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(ctx, d, meta)
}

func resourceDatadogSyntheticsTestDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsDeleteTestsPayload := datadogV1.SyntheticsDeleteTestsPayload{PublicIds: &[]string{d.Id()}}
	if _, _, err := datadogClientV1.SyntheticsApi.DeleteTests(authV1).Body(syntheticsDeleteTestsPayload).Execute(); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, "error deleting synthetics test")
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func isTargetOfTypeInt(assertionType datadogV1.SyntheticsAssertionType, assertionOperator datadogV1.SyntheticsAssertionOperator) bool {
	for _, intTargetAssertionType := range []datadogV1.SyntheticsAssertionType{datadogV1.SYNTHETICSASSERTIONTYPE_RESPONSE_TIME, datadogV1.SYNTHETICSASSERTIONTYPE_CERTIFICATE} {
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

func buildSyntheticsTestStruct(d *schema.ResourceData) *datadogV1.SyntheticsTestDetails {
	request := datadogV1.NewSyntheticsTestRequest()
	k := utils.NewResourceDataKey(d, "")
	parts := ""
	parts = "request_definition.0"
	k.Add(parts)
	if attr, ok := k.GetOkWith("method"); ok {
		request.SetMethod(datadogV1.HTTPMethod(attr.(string)))
	}
	if attr, ok := k.GetOkWith("url"); ok {
		request.SetUrl(attr.(string))
	}
	if attr, ok := k.GetOkWith("body"); ok {
		request.SetBody(attr.(string))
	}
	if attr, ok := k.GetOkWith("timeout"); ok {
		var timeoutInt int
		// first try to convert to int if we're getting from the new TypeList field
		timeoutInt, ok = attr.(int)
		if !ok {
			// if it's not an int, means we are in the deprecated TypeMap field, so convert from string
			timeoutInt, _ = strconv.Atoi(attr.(string))
		}
		request.SetTimeout(float64(timeoutInt))
	}
	if attr, ok := k.GetOkWith("host"); ok {
		request.SetHost(attr.(string))
	}
	if attr, ok := k.GetOkWith("port"); ok {
		var portInt int
		// first try to convert to int if we're getting from the new TypeList field
		portInt, ok = attr.(int)
		if !ok {
			// if it's not an int, means we are in the deprecated TypeMap field, so convert from string
			portInt, _ = strconv.Atoi(attr.(string))
		}
		request.SetPort(int64(portInt))
	}
	if attr, ok := k.GetOkWith("dns_server"); ok {
		request.SetDnsServer(attr.(string))
	}
	k.Remove(parts)
	if attr, ok := d.GetOk("request_query"); ok {
		query := attr.(map[string]interface{})
		if len(query) > 0 {
			request.SetQuery(query)
		}
	}
	if username, ok := d.GetOk("request_basicauth.0.username"); ok {
		if password, ok := d.GetOk("request_basicauth.0.password"); ok {
			basicAuth := datadogV1.NewSyntheticsBasicAuth(password.(string), username.(string))
			request.SetBasicAuth(*basicAuth)
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

	if _, ok := d.GetOk("request_client_certificate"); ok {
		cert := datadogV1.SyntheticsTestRequestCertificateItem{}
		key := datadogV1.SyntheticsTestRequestCertificateItem{}

		if attr, ok := d.GetOk("request_client_certificate.0.cert.0.filename"); ok {
			cert.SetFilename(attr.(string))
		}
		if attr, ok := d.GetOk("request_client_certificate.0.cert.0.content"); ok {
			// only set the certificate content if it is not an already hashed string
			// this is needed for the update function that receives the data from the state
			// and not from the config. So we get a hash of the certificate and not it's real
			// value.
			if isHash := isCertHash(attr.(string)); isHash == false {
				cert.SetContent(attr.(string))
			}
		}

		if attr, ok := d.GetOk("request_client_certificate.0.key.0.filename"); ok {
			key.SetFilename(attr.(string))
		}
		if attr, ok := d.GetOk("request_client_certificate.0.key.0.content"); ok {
			// only set the key content if it is not an already hashed string
			if isHash := isCertHash(attr.(string)); isHash == false {
				key.SetContent(attr.(string))
			}
		}

		clientCertificate := datadogV1.SyntheticsTestRequestCertificate{
			Cert: &cert,
			Key:  &key,
		}

		request.SetCertificate(clientCertificate)
	}

	config := datadogV1.NewSyntheticsTestConfig([]datadogV1.SyntheticsAssertion{}, *request)
	config.SetVariables([]datadogV1.SyntheticsBrowserVariable{})

	if attr, ok := d.GetOk("assertion"); ok && attr != nil {
		for _, assertion := range attr.([]interface{}) {
			assertionMap := assertion.(map[string]interface{})
			if v, ok := assertionMap["type"]; ok {
				assertionType := v.(string)
				if v, ok := assertionMap["operator"]; ok {
					assertionOperator := v.(string)
					if assertionOperator == string(datadogV1.SYNTHETICSASSERTIONJSONPATHOPERATOR_VALIDATES_JSON_PATH) {
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
								case
									datadogV1.SYNTHETICSASSERTIONOPERATOR_LESS_THAN,
									datadogV1.SYNTHETICSASSERTIONOPERATOR_MORE_THAN:
									setFloatTargetValue(subTarget, v.(string))
								default:
									subTarget.SetTargetValue(v)
								}
							}
							assertionJSONPathTarget.SetTarget(*subTarget)
						}
						if _, ok := assertionMap["target"]; ok {
							log.Printf("[WARN] target shouldn't be specified for validateJSONPath operator, only targetjsonpath")
						}
						config.Assertions = append(config.Assertions, datadogV1.SyntheticsAssertionJSONPathTargetAsSyntheticsAssertion(assertionJSONPathTarget))
					} else {
						assertionTarget := datadogV1.NewSyntheticsAssertionTarget(datadogV1.SyntheticsAssertionOperator(assertionOperator), datadogV1.SyntheticsAssertionType(assertionType))
						if v, ok := assertionMap["property"].(string); ok && len(v) > 0 {
							assertionTarget.SetProperty(v)
						}
						if v, ok := assertionMap["target"]; ok {
							if isTargetOfTypeInt(assertionTarget.GetType(), assertionTarget.GetOperator()) {
								assertionTargetInt, _ := strconv.Atoi(v.(string))
								assertionTarget.SetTarget(assertionTargetInt)
							} else {
								assertionTarget.SetTarget(v.(string))
							}
						}
						if v, ok := assertionMap["targetjsonpath"].([]interface{}); ok && len(v) > 0 {
							log.Printf("[WARN] targetjsonpath shouldn't be specified for non-validateJSONPath operator, only target")
						}
						config.Assertions = append(config.Assertions, datadogV1.SyntheticsAssertionTargetAsSyntheticsAssertion(assertionTarget))
					}
				}
			}
		}
	}

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
				if v, ok := variableMap["example"]; ok && v.(string) != "" {
					newVariable.SetExample(v.(string))
				}
				if v, ok := variableMap["id"]; ok && v.(string) != "" {
					newVariable.SetId(v.(string))
				}
				if v, ok := variableMap["pattern"]; ok && v.(string) != "" {
					newVariable.SetPattern(v.(string))
				}
				config.SetVariables(append(config.GetVariables(), *newVariable))
			}
		}
	}

	configVariables := make([]datadogV1.SyntheticsConfigVariable, 0)

	if attr, ok := d.GetOk("config_variable"); ok && attr != nil {
		for _, v := range attr.([]interface{}) {
			variableMap := v.(map[string]interface{})
			variable := datadogV1.SyntheticsConfigVariable{}

			variable.SetType(datadogV1.SyntheticsConfigVariableType(variableMap["type"].(string)))
			variable.SetName(variableMap["name"].(string))
			variable.SetPattern(variableMap["pattern"].(string))
			variable.SetExample(variableMap["example"].(string))

			configVariables = append(configVariables, variable)
		}
	}

	config.SetConfigVariables(configVariables)

	options := datadogV1.NewSyntheticsTestOptions()

	// use new options_list first, then fallback to legacy options
	if attr, ok := d.GetOk("options_list"); ok && attr != nil {
		if attr, ok := d.GetOk("options_list.0.tick_every"); ok {
			options.SetTickEvery(datadogV1.SyntheticsTickInterval(attr.(int)))
		}
		if attr, ok := d.GetOk("options_list.0.accept_self_signed"); ok {
			options.SetAcceptSelfSigned(attr.(bool))
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
	} else {
		if attr, ok := d.GetOk("options.tick_every"); ok {
			tickEvery, _ := strconv.Atoi(attr.(string))
			options.SetTickEvery(datadogV1.SyntheticsTickInterval(tickEvery))
		}
		if attr, ok := d.GetOk("options.follow_redirects"); ok {
			// follow_redirects is a string ("true" or "false") in TF state
			// it used to be "1" and "0" but it does not play well with the API
			// we support both for retro-compatibility
			followRedirects, _ := strconv.ParseBool(attr.(string))
			options.SetFollowRedirects(followRedirects)
		}
		if attr, ok := d.GetOk("options.min_failure_duration"); ok {
			minFailureDuration, _ := strconv.Atoi(attr.(string))
			options.SetMinFailureDuration(int64(minFailureDuration))
		}
		if attr, ok := d.GetOk("options.min_location_failed"); ok {
			minLocationFailed, _ := strconv.Atoi(attr.(string))
			options.SetMinLocationFailed(int64(minLocationFailed))
		}
		if attr, ok := d.GetOk("options.accept_self_signed"); ok {
			// for some reason, attr is equal to "1" or "0" in TF 0.11
			// so ParseBool is required for retro-compatibility
			acceptSelfSigned, _ := strconv.ParseBool(attr.(string))
			options.SetAcceptSelfSigned(acceptSelfSigned)
		}
		if attr, ok := d.GetOk("options.allow_insecure"); ok {
			// for some reason, attr is equal to "1" or "0" in TF 0.11
			// so ParseBool is required for retro-compatibility
			allowInsecure, _ := strconv.ParseBool(attr.(string))
			options.SetAllowInsecure(allowInsecure)
		}
		if attr, ok := d.GetOk("options.retry_count"); ok {
			retryCount, _ := strconv.Atoi(attr.(string))
			retry := datadogV1.SyntheticsTestOptionsRetry{}
			retry.SetCount(int64(retryCount))

			if retryIntervalRaw, ok := d.GetOk("options.retry_interval"); ok {
				retryInterval, _ := strconv.Atoi(retryIntervalRaw.(string))
				retry.SetInterval(float64(retryInterval))
			}

			options.SetRetry(retry)
		}
	}

	if attr, ok := d.GetOk("device_ids"); ok {
		var deviceIds []datadogV1.SyntheticsDeviceID
		for _, s := range attr.([]interface{}) {
			deviceIds = append(deviceIds, datadogV1.SyntheticsDeviceID(s.(string)))
		}
		options.DeviceIds = &deviceIds
	}

	syntheticsTest := datadogV1.NewSyntheticsTestDetails()
	syntheticsTest.SetName(d.Get("name").(string))
	syntheticsTest.SetType(datadogV1.SyntheticsTestDetailsType(d.Get("type").(string)))
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

	if attr, ok := d.GetOk("subtype"); ok {
		syntheticsTest.SetSubtype(datadogV1.SyntheticsTestDetailsSubType(attr.(string)))
	} else {
		if syntheticsTest.GetType() == "api" {
			// we want to default to "http" subtype when type is "api"
			syntheticsTest.SetSubtype(datadogV1.SYNTHETICSTESTDETAILSSUBTYPE_HTTP)
		}
	}

	if attr, ok := d.GetOk("step"); ok && syntheticsTest.GetType() == "browser" {
		steps := []datadogV1.SyntheticsStep{}

		for _, s := range attr.([]interface{}) {
			step := datadogV1.SyntheticsStep{}
			stepMap := s.(map[string]interface{})

			step.SetName(stepMap["name"].(string))
			step.SetType(datadogV1.SyntheticsStepType(stepMap["type"].(string)))
			step.SetAllowFailure(stepMap["allow_failure"].(bool))
			step.SetTimeout(int64(stepMap["timeout"].(int)))
			params := make(map[string]interface{})
			utils.GetMetadataFromJSON([]byte(stepMap["params"].(string)), &params)
			step.SetParams(params)

			steps = append(steps, step)
		}

		syntheticsTest.SetSteps(steps)
	}

	if attr, ok := d.GetOk("browser_step"); ok && syntheticsTest.GetType() == "browser" {
		steps := []datadogV1.SyntheticsStep{}

		for _, s := range attr.([]interface{}) {
			step := datadogV1.SyntheticsStep{}
			stepMap := s.(map[string]interface{})

			step.SetName(stepMap["name"].(string))
			step.SetType(datadogV1.SyntheticsStepType(stepMap["type"].(string)))
			step.SetAllowFailure(stepMap["allow_failure"].(bool))
			step.SetTimeout(int64(stepMap["timeout"].(int)))

			params := make(map[string]interface{})
			stepParams := stepMap["params"].([]interface{})[0]
			stepTypeParams := getParamsKeysForStepType(step.GetType())

			for _, key := range stepTypeParams {
				if stepParams.(map[string]interface{})[key] != "" {
					convertedValue := convertStepParamsValueForConfig(step.GetType(), key, stepParams.(map[string]interface{})[key])
					params[convertStepParamsKey(key)] = convertedValue
				}
			}

			step.SetParams(params)

			steps = append(steps, step)
		}

		syntheticsTest.SetSteps(steps)
	}

	return syntheticsTest
}

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadogV1.SyntheticsTestDetails) diag.Diagnostics {
	d.Set("type", syntheticsTest.GetType())
	if syntheticsTest.HasSubtype() {
		d.Set("subtype", syntheticsTest.GetSubtype())
	}

	config := syntheticsTest.GetConfig()
	actualRequest := config.GetRequest()
	localRequest := make(map[string]interface{})
	if actualRequest.HasBody() {
		localRequest["body"] = actualRequest.GetBody()
	}
	if actualRequest.HasMethod() {
		localRequest["method"] = convertToString(actualRequest.GetMethod())
	}
	if actualRequest.HasTimeout() {
		localRequest["timeout"] = actualRequest.GetTimeout()
	}
	if actualRequest.HasUrl() {
		localRequest["url"] = actualRequest.GetUrl()
	}
	if actualRequest.HasHost() {
		localRequest["host"] = actualRequest.GetHost()
	}
	if actualRequest.HasPort() {
		localRequest["port"] = actualRequest.GetPort()
	}
	if actualRequest.HasDnsServer() {
		localRequest["dns_server"] = convertToString(actualRequest.GetDnsServer())
	}
	if err := d.Set("request_definition", []map[string]interface{}{localRequest}); err != nil {
		return diag.FromErr(err)
	}
	d.Set("request_headers", actualRequest.Headers)
	d.Set("request_query", actualRequest.GetQuery())
	if basicAuth, ok := actualRequest.GetBasicAuthOk(); ok {
		localAuth := make(map[string]string)
		localAuth["username"] = basicAuth.Username
		localAuth["password"] = basicAuth.Password
		d.Set("request_basicauth", []map[string]string{localAuth})
	}

	if clientCertificate, ok := actualRequest.GetCertificateOk(); ok {
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
		if configCertificateContent, ok := d.GetOk("request_client_certificate.0.cert.0.content"); ok {
			localCertificate["cert"][0]["content"] = getCertificateStateValue(configCertificateContent.(string))
		}
		if configKeyContent, ok := d.GetOk("request_client_certificate.0.key.0.content"); ok {
			localCertificate["key"][0]["content"] = getCertificateStateValue(configKeyContent.(string))
		}

		d.Set("request_client_certificate", []map[string][]map[string]string{localCertificate})
	}

	actualAssertions := config.GetAssertions()
	localAssertions := make([]map[string]interface{}, len(actualAssertions))
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
					if vAsString, ok := (*v).(string); ok {
						localTarget["targetvalue"] = vAsString
					} else if vAsFloat, ok := (*v).(float64); ok {
						localTarget["targetvalue"] = strconv.FormatFloat(vAsFloat, 'f', -1, 64)
					} else {
						return diag.Errorf("Unrecognized targetvalue type %v", v)
					}
				}
				localAssertion["targetjsonpath"] = []map[string]string{localTarget}
			}
			if v, ok := assertionTarget.GetTypeOk(); ok {
				localAssertion["type"] = string(*v)
			}
		}
		localAssertions[i] = localAssertion
	}

	if err := d.Set("assertion", localAssertions); err != nil {
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
		if v, ok := variable.GetExampleOk(); ok {
			localVariable["example"] = *v
		}
		if v, ok := variable.GetIdOk(); ok {
			localVariable["id"] = *v
		}
		if v, ok := variable.GetPatternOk(); ok {
			localVariable["pattern"] = *v
		}
		localBrowserVariables[i] = localVariable
	}

	if err := d.Set("browser_variable", localBrowserVariables); err != nil {
		return diag.FromErr(err)
	}

	configVariables := config.GetConfigVariables()
	localConfigVariables := make([]map[string]interface{}, len(configVariables))
	for i, configVariable := range configVariables {
		localVariable := make(map[string]interface{})
		if v, ok := configVariable.GetTypeOk(); ok {
			localVariable["type"] = *v
		}
		if v, ok := configVariable.GetNameOk(); ok {
			localVariable["name"] = *v
		}
		if v, ok := configVariable.GetExampleOk(); ok {
			localVariable["example"] = *v
		}
		if v, ok := configVariable.GetPatternOk(); ok {
			localVariable["pattern"] = *v
		}
		localConfigVariables[i] = localVariable
	}

	if err := d.Set("config_variable", localConfigVariables); err != nil {
		return diag.FromErr(err)
	}

	d.Set("device_ids", syntheticsTest.GetOptions().DeviceIds)

	d.Set("locations", syntheticsTest.Locations)

	actualOptions := syntheticsTest.GetOptions()
	localOptionsList := make(map[string]interface{})
	localOption := make(map[string]string)
	if actualOptions.HasFollowRedirects() {
		localOption["follow_redirects"] = convertToString(actualOptions.GetFollowRedirects())
		localOptionsList["follow_redirects"] = actualOptions.GetFollowRedirects()
	}
	if actualOptions.HasMinFailureDuration() {
		localOption["min_failure_duration"] = convertToString(actualOptions.GetMinFailureDuration())
		localOptionsList["min_failure_duration"] = actualOptions.GetMinFailureDuration()
	}
	if actualOptions.HasMinLocationFailed() {
		localOption["min_location_failed"] = convertToString(actualOptions.GetMinLocationFailed())
		localOptionsList["min_location_failed"] = actualOptions.GetMinLocationFailed()
	}
	if actualOptions.HasTickEvery() {
		localOption["tick_every"] = convertToString(actualOptions.GetTickEvery())
		localOptionsList["tick_every"] = actualOptions.GetTickEvery()
	}
	if actualOptions.HasAcceptSelfSigned() {
		localOption["accept_self_signed"] = convertToString(actualOptions.GetAcceptSelfSigned())
		localOptionsList["accept_self_signed"] = actualOptions.GetAcceptSelfSigned()
	}
	if actualOptions.HasAllowInsecure() {
		localOption["allow_insecure"] = convertToString(actualOptions.GetAllowInsecure())
		localOptionsList["allow_insecure"] = actualOptions.GetAllowInsecure()
	}
	if actualOptions.HasRetry() {
		retry := actualOptions.GetRetry()
		optionsListRetry := make(map[string]interface{})
		localOption["retry_count"] = convertToString(retry.GetCount())
		optionsListRetry["count"] = retry.GetCount()

		if interval, ok := retry.GetIntervalOk(); ok {
			localOption["retry_interval"] = convertToString(interval)
			optionsListRetry["interval"] = interval
		}

		localOptionsList["retry"] = []map[string]interface{}{optionsListRetry}
	}
	if actualOptions.HasMonitorOptions() {
		actualMonitorOptions := actualOptions.GetMonitorOptions()
		renotifyInterval := actualMonitorOptions.GetRenotifyInterval()

		optionsListMonitorOptions := make(map[string]int64)
		optionsListMonitorOptions["renotify_interval"] = renotifyInterval
		localOptionsList["monitor_options"] = []map[string]int64{optionsListMonitorOptions}
	}

	localOptionsLists := make([]map[string]interface{}, 1)
	localOptionsLists[0] = localOptionsList
	if err := d.Set("options_list", localOptionsLists); err != nil {
		return diag.FromErr(err)
	}

	if syntheticsTest.GetType() == "browser" {
		steps := syntheticsTest.GetSteps()
		var localSteps []map[string]interface{}

		useLegacyStep := false

		if attr, ok := d.GetOk("step"); ok && attr != nil && len(attr.([]interface{})) > 0 {
			useLegacyStep = true
		}

		for stepIndex, step := range steps {
			localStep := make(map[string]interface{})
			localStep["name"] = step.GetName()
			localStep["type"] = string(step.GetType())
			localStep["timeout"] = step.GetTimeout()

			if allowFailure, ok := step.GetAllowFailureOk(); ok {
				localStep["allow_failure"] = allowFailure
			}

			if useLegacyStep {
				localStep["params"] = convertToString(step.GetParams().(interface{}))
			} else {
				localParams := make(map[string]interface{})
				params := step.GetParams()
				paramsMap := params.(map[string]interface{})

				for key, value := range paramsMap {
					localParams[convertStepParamsKey(key)] = convertStepParamsValueForState(convertStepParamsKey(key), value)
				}

				localStep["params"] = []interface{}{localParams}

				if forceElementUpdate, ok := d.GetOk(fmt.Sprintf("browser_step.%d.force_element_update", stepIndex)); ok {
					localStep["force_element_update"] = forceElementUpdate
				}
			}

			localSteps = append(localSteps, localStep)
		}

		// If the existing state still uses step, keep using that in the state to not generate useless diffs
		if useLegacyStep {
			if err := d.Set("step", localSteps); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("browser_step", localSteps); err != nil {
				return diag.FromErr(err)
			}
		}

	}

	d.Set("name", syntheticsTest.GetName())
	d.Set("message", syntheticsTest.GetMessage())
	d.Set("status", syntheticsTest.GetStatus())
	d.Set("tags", syntheticsTest.Tags)
	d.Set("monitor_id", syntheticsTest.MonitorId)
	return nil
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
	case datadogV1.HTTPMethod:
		return string(v)
	default:
		// TODO: manage target for JSON body assertions
		valStrr, err := json.Marshal(v)
		if err == nil {
			return string(valStrr)
		}
		return ""
	}
}

func setFloatTargetValue(subTarget *datadogV1.SyntheticsAssertionJSONPathTargetTarget, value string) {
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		subTarget.SetTargetValue(floatValue)
	}
}

func validateSyntheticsAssertionOperator(val interface{}, key string) (warns []string, errs []error) {
	_, err := datadogV1.NewSyntheticsAssertionOperatorFromValue(val.(string))
	if err != nil {
		_, err2 := datadogV1.NewSyntheticsAssertionJSONPathOperatorFromValue(val.(string))
		if err2 != nil {
			errs = append(errs, err, err2)
		}
	}
	return
}

func isCertHash(content string) bool {
	// a sha256 hash consists of 64 hexadecimal characters
	isHash, _ := regexp.MatchString("^[A-Fa-f0-9]{64}$", content)

	return isHash
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

func convertStepParamsValueForConfig(stepType datadogV1.SyntheticsStepType, key string, value interface{}) interface{} {
	switch key {
	case "element", "email", "file", "request":
		result := make(map[string]interface{})
		utils.GetMetadataFromJSON([]byte(value.(string)), &result)
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
	case "element", "email", "file", "request":
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
