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
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"paused": {
				Type:     schema.TypeBool,
				Optional: true,
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
					Default:  0,
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
				"tick_every": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"min_failure_duration": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"min_location_failed": {
					Type:     schema.TypeInt,
					Optional: true,
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
		followRedirects, _ := attr.(bool)
		options.SetFollowRedirects(followRedirects)
	}
	if attr, ok := d.GetOk("options.min_failure_duration"); ok {
		minFailureDuration, _ := strconv.Atoi(attr.(string))
		options.SetMinFailureDuration(minFailureDuration)
	}
	if attr, ok := d.GetOk("options.min_location_failed"); ok {
		minLocationFailed, _ := strconv.Atoi(attr.(string))
		options.SetMinFailureDuration(minLocationFailed)
	}
	if attr, ok := d.GetOk("device_ids"); ok {
		deviceIds := []string{}
		for _, s := range attr.([]interface{}) {
			deviceIds = append(deviceIds, s.(string))
		}
		options.DeviceIds = deviceIds
	}

	status := "live"
	if d.Get("paused").(bool) == true {
		status = "paused"
	}

	syntheticsTest := datadog.SyntheticsTest{
		Name:    datadog.String(d.Get("name").(string)),
		Type:    datadog.String(d.Get("type").(string)),
		Config:  &config,
		Options: &options,
		Message: datadog.String(d.Get("message").(string)),
		Status:  datadog.String(status),
	}

	if attr, ok := d.GetOk("locations"); ok {
		locations := []string{}
		for _, s := range attr.([]interface{}) {
			locations = append(locations, s.(string))
		}
		syntheticsTest.Locations = locations
	}

	if attr, ok := d.GetOk("tags"); ok {
		tags := []string{}
		for _, s := range attr.([]interface{}) {
			tags = append(tags, s.(string))
		}
		syntheticsTest.Tags = tags
	}

	return &syntheticsTest
}

func updateSyntheticsTestLocalState(d *schema.ResourceData, syntheticsTest *datadog.SyntheticsTest) {
	d.Set("type", syntheticsTest.GetType())

	actualRequest := syntheticsTest.GetConfig().Request
	localRequest := newLocalMap(map[string]interface{}{
		"method":  actualRequest.GetMethod(),
		"url":     actualRequest.GetUrl(),
		"body":    actualRequest.GetBody(),
		"timeout": actualRequest.GetTimeout(),
	})
	d.Set("request", localRequest)
	d.Set("request_headers", actualRequest.Headers)

	actualAssertions := syntheticsTest.GetConfig().Assertions
	localAssertions := []map[string]string{}
	for _, assertion := range actualAssertions {
		localAssertion := newLocalMap(map[string]interface{}{
			"type":     assertion.GetType(),
			"property": assertion.GetProperty(),
			"operator": assertion.GetOperator(),
			"target":   assertion.Target,
		})
		localAssertions = append(localAssertions, localAssertion)
	}
	d.Set("assertions", localAssertions)

	d.Set("device_ids", syntheticsTest.GetOptions().DeviceIds)

	d.Set("locations", syntheticsTest.Locations)

	actualOptions := syntheticsTest.GetOptions()
	localOptions := newLocalMap(map[string]interface{}{
		"tick_every":           actualOptions.GetTickEvery(),
		"follow_redirects":     actualOptions.GetFollowRedirects(),
		"min_failure_duration": actualOptions.GetMinFailureDuration(),
		"min_location_failed":  actualOptions.GetMinLocationFailed(),
	})
	d.Set("options", localOptions)

	d.Set("name", syntheticsTest.GetName())
	d.Set("message", syntheticsTest.GetMessage())
	d.Set("tags", syntheticsTest.Tags)
	d.Set("paused", *syntheticsTest.Status == "paused")
}

func newLocalMap(actualMap map[string]interface{}) map[string]string {
	localMap := make(map[string]string)
	for k, i := range actualMap {
		var valStr string
		switch v := i.(type) {
		case bool:
			if v {
				valStr = "1"
			} else {
				valStr = "0"
			}
		case int:
			valStr = strconv.Itoa(v)
		case float64:
			valStr = strconv.FormatFloat(v, 'f', -1, 64)
		case string:
			valStr = v
		default:
			// TODO: manage target for JSON body assertions
			valStrr, err := json.Marshal(v)
			if err == nil {
				valStr = string(valStrr)
			}
		}
		if valStr != "" {
			localMap[k] = valStr
		}
	}
	return localMap
}
