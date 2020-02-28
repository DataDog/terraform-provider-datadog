// For more info about writing custom provider: shttps://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	datadog "github.com/zorkian/go-datadog-api"
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
	client := providerConf.CommunityClient

	syntheticsTest := newSyntheticsTestFromLocalState(d)
	createdSyntheticsTest, err := client.CreateSyntheticsTest(syntheticsTest)
	if err != nil {
		// Note that Id won't be set, so no state will be saved.
		return fmt.Errorf("error creating synthetics test: %s", err.Error())
	}

	// If the Create callback returns with or without an error without an ID set using SetId,
	// the resource is assumed to not be created, and no state is saved.
	d.SetId(createdSyntheticsTest.GetPublicId())

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	syntheticsTest, err := client.GetSyntheticsTest(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return err
	}

	updateSyntheticsTestLocalState(d, syntheticsTest)

	return nil
}

func resourceDatadogSyntheticsTestUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	syntheticsTest := newSyntheticsTestFromLocalState(d)
	if _, err := client.UpdateSyntheticsTest(d.Id(), syntheticsTest); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return err
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	if err := client.DeleteSyntheticsTests([]string{d.Id()}); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return err
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func isTargetOfTypeInt(assertionType string) bool {
	for _, intTargetAssertionType := range []string{"responseTime", "statusCode", "certificate"} {
		if assertionType == intTargetAssertionType {
			return true
		}
	}
	return false
}

func newSyntheticsTestFromLocalState(d *schema.ResourceData) *datadog.SyntheticsTest {
	request := datadog.SyntheticsRequest{}
	if attr, ok := d.GetOk("request.method"); ok {
		request.SetMethod(attr.(string))
	}
	if attr, ok := d.GetOk("request.url"); ok {
		request.SetUrl(attr.(string))
	}
	if attr, ok := d.GetOk("request.body"); ok {
		request.SetBody(attr.(string))
	}
	if attr, ok := d.GetOk("request.timeout"); ok {
		timeoutInt, _ := strconv.Atoi(attr.(string))
		request.SetTimeout(timeoutInt)
	}
	if attr, ok := d.GetOk("request.host"); ok {
		request.SetHost(attr.(string))
	}
	if attr, ok := d.GetOk("request.port"); ok {
		portInt, _ := strconv.Atoi(attr.(string))
		request.SetPort(portInt)
	}
	if attr, ok := d.GetOk("request_headers"); ok {
		headers := attr.(map[string]interface{})
		if len(headers) > 0 {
			request.Headers = make(map[string]string)
		}
		for k, v := range headers {
			request.Headers[k] = v.(string)
		}
	}

	config := datadog.SyntheticsConfig{
		Request:   &request,
		Variables: []interface{}{},
	}

	if attr, ok := d.GetOk("assertions"); ok {
		for _, attr := range attr.([]interface{}) {
			assertion := datadog.SyntheticsAssertion{}
			assertionMap := attr.(map[string]interface{})
			if v, ok := assertionMap["type"]; ok {
				assertionType := v.(string)
				assertion.Type = &assertionType
			}
			if v, ok := assertionMap["property"]; ok {
				assertionProperty := v.(string)
				assertion.Property = &assertionProperty
			}
			if v, ok := assertionMap["operator"]; ok {
				assertionOperator := v.(string)
				assertion.Operator = &assertionOperator
			}
			if v, ok := assertionMap["target"]; ok {
				if isTargetOfTypeInt(*assertion.Type) {
					assertionTargetInt, _ := strconv.Atoi(v.(string))
					assertion.Target = assertionTargetInt
				} else if *assertion.Operator == "validates" {
					assertion.Target = json.RawMessage(v.(string))
				} else {
					assertion.Target = v.(string)
				}
			}
			config.Assertions = append(config.Assertions, assertion)
		}
	}

	options := datadog.SyntheticsOptions{}
	if attr, ok := d.GetOk("options.tick_every"); ok {
		tickEvery, _ := strconv.Atoi(attr.(string))
		options.SetTickEvery(tickEvery)
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
		options.SetMinFailureDuration(minFailureDuration)
	}
	if attr, ok := d.GetOk("options.min_location_failed"); ok {
		minLocationFailed, _ := strconv.Atoi(attr.(string))
		options.SetMinLocationFailed(minLocationFailed)
	}
	if attr, ok := d.GetOk("options.accept_self_signed"); ok {
		// for some reason, attr is equal to "1" or "0" in TF 0.11
		// so ParseBool is required for retro-compatibility
		acceptSelfSigned, _ := strconv.ParseBool(attr.(string))
		options.SetAcceptSelfSigned(acceptSelfSigned)
	}
	if attr, ok := d.GetOk("device_ids"); ok {
		deviceIds := []string{}
		for _, s := range attr.([]interface{}) {
			deviceIds = append(deviceIds, s.(string))
		}
		options.DeviceIds = deviceIds
	}

	syntheticsTest := datadog.SyntheticsTest{
		Name:    datadog.String(d.Get("name").(string)),
		Type:    datadog.String(d.Get("type").(string)),
		Config:  &config,
		Options: &options,
		Message: datadog.String(d.Get("message").(string)),
		Status:  datadog.String(d.Get("status").(string)),
	}

	if attr, ok := d.GetOk("locations"); ok {
		locations := []string{}
		for _, s := range attr.([]interface{}) {
			locations = append(locations, s.(string))
		}
		syntheticsTest.Locations = locations
	}

	tags := []string{}
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
	}
	syntheticsTest.Tags = tags

	if attr, ok := d.GetOk("subtype"); ok {
		syntheticsTest.Subtype = datadog.String(attr.(string))
	} else {
		if *syntheticsTest.Type == "api" {
			// we want to default to "http" subtype when type is "api"
			syntheticsTest.Subtype = datadog.String("http")
		}
	}

	return &syntheticsTest
}

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadog.SyntheticsTest) {
	d.Set("type", syntheticsTest.GetType())
	if syntheticsTest.HasSubtype() {
		d.Set("subtype", syntheticsTest.GetSubtype())
	}

	actualRequest := syntheticsTest.GetConfig().Request
	localRequest := make(map[string]string)
	if actualRequest.HasBody() {
		localRequest["body"] = actualRequest.GetBody()
	}
	if actualRequest.HasMethod() {
		localRequest["method"] = actualRequest.GetMethod()
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
	d.Set("request", localRequest)
	d.Set("request_headers", actualRequest.Headers)

	actualAssertions := syntheticsTest.GetConfig().Assertions
	localAssertions := []map[string]string{}
	for _, assertion := range actualAssertions {
		localAssertion := make(map[string]string)
		if assertion.HasOperator() {
			localAssertion["operator"] = assertion.GetOperator()
		}
		if assertion.HasProperty() {
			localAssertion["property"] = assertion.GetProperty()
		}
		if target := assertion.Target; target != nil {
			localAssertion["target"] = convertToString(target)
		}
		if assertion.HasType() {
			localAssertion["type"] = assertion.GetType()
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
	if actualOptions.HasMinFailureDuration() {
		localOptions["min_failure_duration"] = convertToString(actualOptions.GetMinFailureDuration())
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
	d.Set("monitor_id", syntheticsTest.MonitorId)
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
