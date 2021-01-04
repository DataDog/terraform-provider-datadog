// For more info about writing custom provider: https://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"encoding/json"
	"fmt"
	"log"
	_nethttp "net/http"
	"regexp"
	"strconv"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceDatadogSyntheticsTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogSyntheticsTestCreate,
		Read:   resourceDatadogSyntheticsTestRead,
		Update: resourceDatadogSyntheticsTestUpdate,
		Delete: resourceDatadogSyntheticsTestDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsTestDetailsTypeFromValue),
			},
			"subtype": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
					if d.Get("type") == "api" && old == "http" && new == "" {
						// defaults to http if type is api for retro-compatibility
						return true
					}
					return old == new
				},
				ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsTestDetailsSubTypeFromValue),
			},
			"request": syntheticsTestRequest(),
			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"request_query": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"request_basicauth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"request_client_certificate": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cert": syntheticsTestRequestClientCertificateItem(),
						"key":  syntheticsTestRequestClientCertificateItem(),
					},
				},
			},
			"assertions": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"assertion"},
				Deprecated:    "Use assertion instead",
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"assertion": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"assertions"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsAssertionTypeFromValue),
							Required:     true,
						},
						"operator": {
							Type:         schema.TypeString,
							ValidateFunc: validateSyntheticsAssertionOperator,
							Required:     true,
						},
						"property": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"targetjsonpath": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"operator": {
										Type:     schema.TypeString,
										Required: true,
									},
									"jsonpath": {
										Type:     schema.TypeString,
										Required: true,
									},
									"targetvalue": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"variable":         syntheticsBrowserVariableLegacy(),
			"browser_variable": syntheticsBrowserVariable(),
			"device_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsDeviceIDFromValue),
				},
			},
			"locations": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"options":      syntheticsTestOptions(),
			"options_list": syntheticsTestOptionsList(),
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:         schema.TypeString,
				ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsTestPauseStatusFromValue),
				Required:     true,
			},
			"monitor_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"step": syntheticsTestStep(),
		},
	}
}

func syntheticsTestRequest() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"method": {
					Type:         schema.TypeString,
					ValidateFunc: validateEnumValue(datadogV1.NewHTTPMethodFromValue),
					Optional:     true,
				},
				"url": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"body": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"timeout": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  60,
				},
				"host": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"port": {
					Type:     schema.TypeInt,
					Optional: true,
					Default:  60,
				},
				"dns_server": {
					Type:     schema.TypeString,
					Optional: true,
				},
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
					Type:      schema.TypeString,
					Required:  true,
					Sensitive: true,
				},
				"filename": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "Provided in Terraform config",
				},
			},
		},
	}
}

func syntheticsTestOptions() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeMap,
		ConflictsWith: []string{"options_list"},
		Deprecated:    "This parameter is deprecated, please use `options_list`",
		DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
			if key == "options.follow_redirects" || key == "options.accept_self_signed" || key == "options.allow_insecure" {
				// TF nested schemas is limited to string values only
				// follow_redirects, accept_self_signed and allow_insecure being booleans in Datadog json api
				// we need a sane way to convert from boolean to string
				// and from string to boolean
				oldValue, err1 := strconv.ParseBool(old)
				newValue, err2 := strconv.ParseBool(new)
				if err1 != nil || err2 != nil {
					return false
				}
				return oldValue == newValue
			}
			return old == new
		},
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			followRedirectsRaw, ok := val.(map[string]interface{})["follow_redirects"]
			if ok {
				followRedirectsStr := convertToString(followRedirectsRaw)
				switch followRedirectsStr {
				case "0", "1":
					warns = append(warns, fmt.Sprintf("%q.follow_redirects must be either true or false, got: %s (please change 1 => true, 0 => false)", key, followRedirectsStr))
				case "true", "false":
					break
				default:
					errs = append(errs, fmt.Errorf("%q.follow_redirects must be either true or false, got: %s", key, followRedirectsStr))
				}
			}
			acceptSelfSignedRaw, ok := val.(map[string]interface{})["accept_self_signed"]
			if ok {
				acceptSelfSignedStr := convertToString(acceptSelfSignedRaw)
				switch acceptSelfSignedStr {
				case "true", "false":
					break
				default:
					errs = append(errs, fmt.Errorf("%q.accept_self_signed must be either true or false, got: %s", key, acceptSelfSignedStr))
				}
			}
			allowInsecureRaw, ok := val.(map[string]interface{})["allow_insecure"]
			if ok {
				allowInsecureStr := convertToString(allowInsecureRaw)
				switch allowInsecureStr {
				case "true", "false":
					break
				default:
					errs = append(errs, fmt.Errorf("%q.allow_insecure must be either true or false, got: %s", key, allowInsecureStr))
				}
			}
			return
		},
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"follow_redirects": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"min_failure_duration": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"min_location_failed": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"tick_every": {
					Type:         schema.TypeInt,
					ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsTickIntervalFromValue),
					Required:     true,
				},
				"accept_self_signed": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"allow_insecure": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"retry_count": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"retry_interval": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			},
		},
	}
}

func syntheticsTestOptionsList() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		MaxItems:      1,
		ConflictsWith: []string{"options"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allow_insecure": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"follow_redirects": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"tick_every": {
					Type:         schema.TypeInt,
					ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsTickIntervalFromValue),
					Optional:     true,
				},
				"accept_self_signed": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"min_location_failed": {
					Type:     schema.TypeInt,
					Default:  1,
					Optional: true,
				},
				"min_failure_duration": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"monitor_options": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"renotify_interval": {
								Type:     schema.TypeInt,
								Default:  0,
								Optional: true,
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
								Type:     schema.TypeInt,
								Default:  0,
								Optional: true,
							},
							"interval": {
								Type:     schema.TypeInt,
								Default:  300,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func syntheticsTestStep() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"type": {
					Type:         schema.TypeString,
					ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsStepTypeFromValue),
					Required:     true,
				},
				"allow_failure": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"timeout": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"params": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func syntheticsBrowserVariableLegacy() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"browser_variable"},
		Deprecated:    "This parameter is deprecated, please use `browser_variable`",
		Elem:          syntheticsBrowserVariableElem(),
	}
}

func syntheticsBrowserVariable() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Optional:      true,
		ConflictsWith: []string{"variable"},
		Elem:          syntheticsBrowserVariableElem(),
	}
}

func syntheticsBrowserVariableElem() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"example": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Z][A-Z0-9_]+[A-Z0-9]$`), "must be all uppercase with underscores"),
			},
			"pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateEnumValue(datadogV1.NewSyntheticsBrowserVariableTypeFromValue),
			},
		},
	}
}

func resourceDatadogSyntheticsTestCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsTest := buildSyntheticsTestStruct(d)
	createdSyntheticsTest, _, err := datadogClientV1.SyntheticsApi.CreateTest(authV1).Body(*syntheticsTest).Execute()
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return translateClientError(err, "error creating synthetics test")
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsTest.GetPublicId())

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestRead(d *schema.ResourceData, meta interface{}) error {
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
		return translateClientError(err, "error getting synthetics test")
	}

	return updateSyntheticsTestLocalState(d, &syntheticsTest)
}

func resourceDatadogSyntheticsTestUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsTest := buildSyntheticsTestStruct(d)
	if _, _, err := datadogClientV1.SyntheticsApi.UpdateTest(authV1, d.Id()).Body(*syntheticsTest).Execute(); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return translateClientError(err, "error updating synthetics test")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsDeleteTestsPayload := datadogV1.SyntheticsDeleteTestsPayload{PublicIds: &[]string{d.Id()}}
	if _, _, err := datadogClientV1.SyntheticsApi.DeleteTests(authV1).Body(syntheticsDeleteTestsPayload).Execute(); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return translateClientError(err, "error deleting synthetics test")
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
	if attr, ok := d.GetOk("request.method"); ok {
		request.SetMethod(datadogV1.HTTPMethod(attr.(string)))
	}
	if attr, ok := d.GetOk("request.url"); ok {
		request.SetUrl(attr.(string))
	}
	if attr, ok := d.GetOk("request.body"); ok {
		request.SetBody(attr.(string))
	}
	if attr, ok := d.GetOk("request.timeout"); ok {
		timeoutInt, _ := strconv.Atoi(attr.(string))
		request.SetTimeout(float64(timeoutInt))
	}
	if attr, ok := d.GetOk("request.host"); ok {
		request.SetHost(attr.(string))
	}
	if attr, ok := d.GetOk("request.port"); ok {
		portInt, _ := strconv.Atoi(attr.(string))
		request.SetPort(int64(portInt))
	}
	if attr, ok := d.GetOk("request.dns_server"); ok {
		request.SetDnsServer(attr.(string))
	}
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
			cert.SetContent(attr.(string))
		}

		if attr, ok := d.GetOk("request_client_certificate.0.key.0.filename"); ok {
			key.SetFilename(attr.(string))
		}
		if attr, ok := d.GetOk("request_client_certificate.0.key.0.content"); ok {
			key.SetContent(attr.(string))
		}

		clientCertificate := datadogV1.SyntheticsTestRequestCertificate{
			Cert: &cert,
			Key:  &key,
		}

		request.SetCertificate(clientCertificate)
	}

	config := datadogV1.NewSyntheticsTestConfig([]datadogV1.SyntheticsAssertion{}, *request)
	config.SetVariables([]datadogV1.SyntheticsBrowserVariable{})

	// Deprecated path, the assertions field is replaced with assertion
	if attr, ok := d.GetOk("assertions"); ok && attr != nil {
		for _, assertion := range attr.([]interface{}) {
			assertionMap := assertion.(map[string]interface{})
			if v, ok := assertionMap["type"]; ok {
				assertionType := v.(string)
				if v, ok := assertionMap["operator"]; ok {
					assertionOperator := v.(string)
					assertionTarget := datadogV1.NewSyntheticsAssertionTarget(datadogV1.SyntheticsAssertionOperator(assertionOperator), datadogV1.SyntheticsAssertionType(assertionType))
					if v, ok := assertionMap["property"]; ok {
						assertionProperty := v.(string)
						assertionTarget.SetProperty(assertionProperty)
					}
					if v, ok := assertionMap["target"]; ok {
						if isTargetOfTypeInt(assertionTarget.GetType(), assertionTarget.GetOperator()) {
							assertionTargetInt, _ := strconv.Atoi(v.(string))
							assertionTarget.SetTarget(assertionTargetInt)
						} else {
							assertionTarget.SetTarget(v.(string))
						}
					}
					config.Assertions = append(config.Assertions, datadogV1.SyntheticsAssertionTargetAsSyntheticsAssertion(assertionTarget))
				}
			}
		}
	}

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
	} else if attr, ok := d.GetOk("variable"); ok && attr != nil {
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
		for _, s := range attr.([]interface{}) {
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
			step.SetTimeout(float32(stepMap["timeout"].(int)))
			params := make(map[string]interface{})
			getMetadataFromJSON([]byte(stepMap["params"].(string)), &params)
			step.SetParams(params)

			steps = append(steps, step)
		}

		syntheticsTest.SetSteps(steps)
	}

	return syntheticsTest
}

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadogV1.SyntheticsTestDetails) error {
	d.Set("type", syntheticsTest.GetType())
	if syntheticsTest.HasSubtype() {
		d.Set("subtype", syntheticsTest.GetSubtype())
	}

	config := syntheticsTest.GetConfig()
	actualRequest := config.GetRequest()
	localRequest := make(map[string]string)
	if actualRequest.HasBody() {
		localRequest["body"] = actualRequest.GetBody()
	}
	if actualRequest.HasMethod() {
		localRequest["method"] = convertToString(actualRequest.GetMethod())
	}
	if actualRequest.HasTimeout() {
		localRequest["timeout"] = convertToString(actualRequest.GetTimeout())
	}
	if actualRequest.HasUrl() {
		localRequest["url"] = actualRequest.GetUrl()
	}
	if actualRequest.HasHost() {
		localRequest["host"] = actualRequest.GetHost()
	}
	if actualRequest.HasPort() {
		localRequest["port"] = convertToString(actualRequest.GetPort())
	}
	if actualRequest.HasDnsServer() {
		localRequest["dns_server"] = convertToString(actualRequest.GetDnsServer())
	}
	d.Set("request", localRequest)
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

		// the content of the certificate and the key are write-only
		// so we need to get them from the config since they will
		// not be in the api response
		if configCertificateContent, ok := d.GetOk("request_client_certificate.0.cert.0.content"); ok {
			localCertificate["cert"][0]["content"] = configCertificateContent.(string)
		}
		if configKeyContent, ok := d.GetOk("request_client_certificate.0.key.0.content"); ok {
			localCertificate["key"][0]["content"] = configKeyContent.(string)
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
						return fmt.Errorf("Unrecognized targetvalue type %v", v)
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
	// If the existing state still uses assertions, keep using that in the state to not generate useless diffs
	if attr, ok := d.GetOk("assertions"); ok && attr != nil && len(attr.([]interface{})) > 0 {
		if err := d.Set("assertions", localAssertions); err != nil {
			return err
		}
	} else {
		if err := d.Set("assertion", localAssertions); err != nil {
			return err
		}
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

	// If the existing state still uses variables, keep using that in the state to not generate useless diffs
	if attr, ok := d.GetOk("variable"); ok && attr != nil && len(attr.([]interface{})) > 0 {
		if err := d.Set("variable", localBrowserVariables); err != nil {
			return err
		}
	} else {
		if err := d.Set("browser_variable", localBrowserVariables); err != nil {
			return err
		}
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

	// If the existing state still uses options, keep using that in the state to not generate useless diffs
	if attr, ok := d.GetOk("options"); ok && attr != nil && len(attr.(map[string]interface{})) > 0 {
		if err := d.Set("options", localOption); err != nil {
			return err
		}
	} else {
		localOptionsLists := make([]map[string]interface{}, 1)
		localOptionsLists[0] = localOptionsList
		if err := d.Set("options_list", localOptionsLists); err != nil {
			return err
		}
	}

	if syntheticsTest.GetType() == "browser" {
		steps := syntheticsTest.GetSteps()
		var localSteps []map[string]interface{}

		for _, step := range steps {
			localStep := make(map[string]interface{})
			localStep["name"] = step.GetName()
			localStep["type"] = string(step.GetType())
			localStep["timeout"] = step.GetTimeout()

			if allowFailure, ok := step.GetAllowFailureOk(); ok {
				localStep["allow_failure"] = allowFailure
			}

			localStep["params"] = convertToString(step.GetParams().(interface{}))

			localSteps = append(localSteps, localStep)
		}

		d.Set("step", localSteps)
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
