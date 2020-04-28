// For more info about writing custom provider: shttps://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var syntheticsTypes = []string{"api", "browser"}
var syntheticsSubTypes = []string{"http", "ssl"}

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
				ValidateFunc: validation.StringInSlice(syntheticsTypes, false),
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
				ValidateFunc: validation.StringInSlice(syntheticsSubTypes, false),
			},
			"request": syntheticsTestRequest(),
			"request_headers": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"assertions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
			"device_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"locations": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"options": syntheticsTestOptions(),
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
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Required: true,
			},
			"monitor_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
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
					Type:     schema.TypeString,
					Optional: true,
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
			},
		},
	}
}

func syntheticsTestOptions() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeMap,
		DiffSuppressFunc: func(key, old, new string, d *schema.ResourceData) bool {
			if key == "options.follow_redirects" || key == "options.accept_self_signed" {
				// TF nested schemas is limited to string values only
				// follow_redirects and accept_self_signed being booleans in Datadog json api
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
					Type:     schema.TypeInt,
					Required: true,
				},
				"accept_self_signed": {
					Type:     schema.TypeBool,
					Optional: true,
				},
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

	syntheticsTest, _, err := datadogClientV1.SyntheticsApi.GetTest(authV1, d.Id()).Execute()
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting synthetics test")
	}

	updateSyntheticsTestLocalState(d, &syntheticsTest)

	return nil
}

func resourceDatadogSyntheticsTestUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	syntheticsTest := buildSyntheticsTestStruct(d)
	if _, _, err := datadogClientV1.SyntheticsApi.UpdateTest(authV1, d.Id()).Body(*syntheticsTest).Execute(); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		translateClientError(err, "error updating synthetics test")
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if _, _, err := datadogClientV1.SyntheticsApi.DeleteTests(authV1).Body(datadogV1.SyntheticsDeleteTestsPayload{PublicIds: &[]string{d.Id()}}).Execute(); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return translateClientError(err, "error deleting synthetics test")
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func isTargetOfTypeInt(assertionType datadogV1.SyntheticsAssertionType) bool {
	for _, intTargetAssertionType := range []datadogV1.SyntheticsAssertionType{datadogV1.SYNTHETICSASSERTIONTYPE_RESPONSE_TIME, datadogV1.SYNTHETICSASSERTIONTYPE_STATUS_CODE, datadogV1.SYNTHETICSASSERTIONTYPE_CERTIFICATE} {
		if assertionType == intTargetAssertionType {
			return true
		}
	}
	return false
}

func buildSyntheticsTestStruct(d *schema.ResourceData) *datadogV1.SyntheticsTestDetails {
	request := datadogV1.SyntheticsTestRequest{}
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
	if attr, ok := d.GetOk("request_headers"); ok {
		headers := attr.(map[string]interface{})
		if len(headers) > 0 {
			request.SetHeaders(make(map[string]string))
		}
		for k, v := range headers {
			tHeaders := request.GetHeaders()
			tHeaders[k] = v.(string)
			request.SetHeaders(tHeaders)
		}
	}

	config := datadogV1.SyntheticsTestConfig{
		Request:   request,
		Variables: &[]datadogV1.SyntheticsBrowserVariable{},
	}

	if attr, ok := d.GetOk("assertions"); ok {
		for _, attr := range attr.([]interface{}) {
			assertion := datadogV1.SyntheticsAssertion{}
			assertionMap := attr.(map[string]interface{})
			if v, ok := assertionMap["type"]; ok {
				assertionType := v.(string)
				assertion.SetType(datadogV1.SyntheticsAssertionType(assertionType))
			}
			if v, ok := assertionMap["property"]; ok {
				assertionProperty := v.(string)
				assertion.SetProperty(assertionProperty)
			}
			if v, ok := assertionMap["operator"]; ok {
				assertionOperator := v.(string)
				assertion.SetOperator(datadogV1.SyntheticsAssertionOperator(assertionOperator))
			}
			if v, ok := assertionMap["target"]; ok {
				if isTargetOfTypeInt(assertion.Type) {
					assertionTargetInt, _ := strconv.Atoi(v.(string))
					assertion.SetTarget(assertionTargetInt)
				} else if assertion.Operator == "validates" {
					assertion.SetTarget(v.(string))
				} else {
					assertion.SetTarget(v.(string))
				}
			}
			config.Assertions = append(config.Assertions, assertion)
		}
	}

	options := datadogV1.SyntheticsTestOptions{}
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
	//if attr, ok := d.GetOk("options.min_failure_duration"); ok {
	//	minFailureDuration, _ := strconv.Atoi(attr.(string))
	//	options.SetMinFailureDuration(minFailureDuration)
	//}
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
	if attr, ok := d.GetOk("device_ids"); ok {
		var deviceIds []datadogV1.SyntheticsDeviceID
		for _, s := range attr.([]interface{}) {
			deviceIds = append(deviceIds, datadogV1.SyntheticsDeviceID(s.(string)))
		}
		options.DeviceIds = &deviceIds
	}

	syntheticsTest := datadogV1.SyntheticsTestDetails{
		Name:    datadogV1.PtrString(d.Get("name").(string)),
		Type:    datadogV1.SyntheticsTestDetailsType(d.Get("type").(string)).Ptr(),
		Config:  &config,
		Options: &options,
		Message: datadogV1.PtrString(d.Get("message").(string)),
		Status:  datadogV1.SyntheticsTestPauseStatus(d.Get("status").(string)).Ptr(),
	}

	if attr, ok := d.GetOk("locations"); ok {
		var locations []string
		for _, s := range attr.([]interface{}) {
			locations = append(locations, s.(string))
		}
		syntheticsTest.SetLocations(locations)
	}

	var tags []string
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
	}
	syntheticsTest.SetTags(tags)

	if attr, ok := d.GetOk("subtype"); ok {
		syntheticsTest.SetSubtype(datadogV1.SyntheticsTestDetailsSubType(attr.(string)))
	} else {
		if *syntheticsTest.Type == "api" {
			// we want to default to "http" subtype when type is "api"
			syntheticsTest.SetSubtype("http")
		}
	}

	return &syntheticsTest
}

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadogV1.SyntheticsTestDetails) {
	d.Set("type", syntheticsTest.GetType())
	if syntheticsTest.HasSubtype() {
		d.Set("subtype", syntheticsTest.GetSubtype())
	}

	actualRequest := syntheticsTest.GetConfig().Request
	localRequest := make(map[string]string)
	if actualRequest.HasBody() {
		localRequest["body"] = actualRequest.GetBody()
	}
	if _, ok := actualRequest.GetMethodOk(); ok {
		localRequest["method"] = convertToString(actualRequest.GetMethod())
	}
	if actualRequest.HasTimeout() {
		localRequest["timeout"] = convertToString(actualRequest.GetTimeout())
	}
	if _, ok := actualRequest.GetUrlOk(); ok {
		localRequest["url"] = actualRequest.GetUrl()
	}
	if actualRequest.HasHost() {
		localRequest["host"] = actualRequest.GetHost()
	}
	if actualRequest.HasPort() {
		localRequest["port"] = convertToString(actualRequest.GetPort())
	}
	d.Set("request", localRequest)
	d.Set("request_headers", actualRequest.Headers)

	actualAssertions := syntheticsTest.GetConfig().Assertions
	var localAssertions []map[string]string
	for _, assertion := range actualAssertions {
		localAssertion := make(map[string]string)
		if _, ok := assertion.GetOperatorOk(); ok {
			localAssertion["operator"] = convertToString(assertion.GetOperator())
		}
		if assertion.HasProperty() {
			localAssertion["property"] = assertion.GetProperty()
		}
		if target := assertion.Target; target != nil {
			localAssertion["target"] = convertToString(target)
		}
		if _, ok := assertion.GetTypeOk(); ok {
			localAssertion["type"] = convertToString(assertion.GetType())
		}
		localAssertions = append(localAssertions, localAssertion)
	}
	d.Set("assertions", localAssertions)

	d.Set("device_ids", syntheticsTest.GetOptions().DeviceIds)

	d.Set("locations", syntheticsTest.Locations)

	actualOptions := syntheticsTest.GetOptions()
	localOptions := make(map[string]string)
	if actualOptions.HasFollowRedirects() {
		localOptions["follow_redirects"] = convertToString(actualOptions.GetFollowRedirects())
	}
	if v, ok := actualOptions.GetMinLocationFailedOk(); ok {
		localOptions["min_failure_duration"] = convertToString(v)
	}
	if actualOptions.HasMinLocationFailed() {
		localOptions["min_location_failed"] = convertToString(actualOptions.GetMinLocationFailed())
	}
	if actualOptions.HasTickEvery() {
		localOptions["tick_every"] = convertToString(actualOptions.GetTickEvery())
	}
	if actualOptions.HasAcceptSelfSigned() {
		localOptions["accept_self_signed"] = convertToString(actualOptions.GetAcceptSelfSigned())
	}

	d.Set("options", localOptions)

	d.Set("name", syntheticsTest.GetName())
	d.Set("message", syntheticsTest.GetMessage())
	d.Set("status", syntheticsTest.GetStatus())
	d.Set("tags", syntheticsTest.Tags)
	//d.Set("monitor_id", syntheticsTest.MonitorId)
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
