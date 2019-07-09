// For more info about writing custom provider: shttps://www.terraform.io/docs/extend/writing-custom-providers.html

package datadog

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	datadog "github.com/zorkian/go-datadog-api"
)

var syntheticsTypes = []string{"api", "browser"}

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
					Required: true,
				},
				"url": {
					Type:     schema.TypeString,
					Required: true,
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
			},
		},
	}
}

func syntheticsTestOptions() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
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
			},
		},
	}
}

func resourceDatadogSyntheticsTestCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

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
	client := meta.(*datadog.Client)

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
	client := meta.(*datadog.Client)

	syntheticsTest := newSyntheticsTestFromLocalState(d)
	if _, err := client.UpdateSyntheticsTest(d.Id(), syntheticsTest); err != nil {
		// If the Update callback returns with or without an error, the full state is saved.
		return err
	}

	// Return the read function to ensure the state is reflected in the terraform.state file
	return resourceDatadogSyntheticsTestRead(d, meta)
}

func resourceDatadogSyntheticsTestDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteSyntheticsTests([]string{d.Id()}); err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return err
	}

	// The resource is assumed to be destroyed, and all state is removed.
	return nil
}

func isTargetOfTypeInt(assertionType string) bool {
	for _, intTargetAssertionType := range []string{"responseTime", "statusCode"} {
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
		// TF nested schemas is limited to string values only
		// follow_redirects being a boolean in Datadog json api
		// we need a sane way to convert from boolean to string
		// and from string to boolean
		followRedirects := attr.(string) == "true"
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

	return &syntheticsTest
}

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadog.SyntheticsTest) {
	d.Set("type", syntheticsTest.GetType())

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
		if v {
			return "true"
		}
		return "false"
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
