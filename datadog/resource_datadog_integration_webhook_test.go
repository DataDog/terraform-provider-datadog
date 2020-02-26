package datadog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestBuildDatadogHeader(t *testing.T) {
	cases := map[string]struct {
		terraformHeaders       map[string]interface{}
		expectedDatadogHeaders string
	}{
		"no headers": {
			terraformHeaders:       map[string]interface{}{},
			expectedDatadogHeaders: "",
		},
		"single header": {
			terraformHeaders: map[string]interface{}{
				"header1": "val1",
			},
			expectedDatadogHeaders: "header1: val1",
		},
		"multiple header": {
			terraformHeaders: map[string]interface{}{
				"header1": "val1",
				"header2": "val2",
				"header3": "val3",
			},
			expectedDatadogHeaders: "header1: val1\nheader2: val2\nheader3: val3",
		},
	}
	for name, tc := range cases {
		actualDatadogHeaders := buildDatadogHeader(tc.terraformHeaders)

		if tc.expectedDatadogHeaders != actualDatadogHeaders {
			t.Errorf("%s: Expected '%s', but got '%s'", name, tc.expectedDatadogHeaders, actualDatadogHeaders)
		}
	}
}

func TestBuildTerraformHeader(t *testing.T) {
	cases := map[string]struct {
		datadogHeaders           string
		expectedTerraformHeaders map[string]interface{}
		expectedError            error
	}{
		"no headers": {
			datadogHeaders:           "",
			expectedTerraformHeaders: map[string]interface{}{},
			expectedError:            nil,
		},
		"no headers with whitespace": {
			datadogHeaders:           "  \t",
			expectedTerraformHeaders: map[string]interface{}{},
			expectedError:            nil,
		},
		"single header": {
			datadogHeaders: "header1: val1",
			expectedTerraformHeaders: map[string]interface{}{
				"header1": "val1",
			},
			expectedError: nil,
		},
		"multiple headers": {
			datadogHeaders: "header1: val1\nheader2: val2\nheader3: val3",
			expectedTerraformHeaders: map[string]interface{}{
				"header1": "val1",
				"header2": "val2",
				"header3": "val3",
			},
			expectedError: nil,
		},
		"colon in value": {
			datadogHeaders: "header1: val:with:colon",
			expectedTerraformHeaders: map[string]interface{}{
				"header1": "val:with:colon",
			},
			expectedError: nil,
		},
		"no colon between header and value": {
			datadogHeaders:           "header1 val",
			expectedTerraformHeaders: nil,
			expectedError:            fmt.Errorf("header not correctly formatted, expected ':' in 'header1 val'"),
		},
	}
	for name, tc := range cases {
		actualTerraformHeaders, err := buildTerraformHeader(tc.datadogHeaders)

		if err != nil && tc.expectedError == nil {
			t.Errorf("%s: Unexpected error occured '%s", name, err.Error())
		} else if err == nil && tc.expectedError != nil {
			t.Errorf("%s: Expected error '%s', but it was not thrown", name, tc.expectedError.Error())
		} else if err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
			t.Errorf("%s: Unexpected error occured '%s', expected '%s'", name, err.Error(), tc.expectedError.Error())
		} else if actualTerraformHeaders == nil && tc.expectedError == nil {
			t.Errorf("%s: Did not expect an error and terraform webhook is nil", name)
		} else if actualTerraformHeaders != nil && tc.expectedError != nil {
			t.Errorf("%s: Expected error '%s', but got non-nil terraform webhook", name, tc.expectedError.Error())
		} else if actualTerraformHeaders != nil {
			compareTerraformHeaders(tc.expectedTerraformHeaders, *actualTerraformHeaders, name, t)
		}
	}
}

func TestBuildDatadogWebhook(t *testing.T) {
	cases := map[string]struct {
		terraformWebhook       map[string]interface{}
		expectedDatadogWebhook datadog.Webhook
	}{
		"required fields only": {
			terraformWebhook: map[string]interface{}{
				"name": "my_webhook",
				"url":  "http://example.com",
			},
			expectedDatadogWebhook: datadog.Webhook{
				Name: datadog.String("my_webhook"),
				URL:  datadog.String("http://example.com"),
			},
		},
		"custom headers": {
			terraformWebhook: map[string]interface{}{
				"name": "my_webhook",
				"url":  "http://example.com",
				"headers": map[string]interface{}{
					"header1": "val1",
				},
			},
			expectedDatadogWebhook: datadog.Webhook{
				Name: datadog.String("my_webhook"),
				URL:  datadog.String("http://example.com"),
				Headers: datadog.String("header1: val),
			},
		},
		"custom payload": {
			terraformWebhook: map[string]interface{}{
				"name":               "my_webhook",
				"url":                "http://example.com",
				"use_custom_payload": true,
				"custom_payload":     "field: value",
			},
			expectedDatadogWebhook: datadog.Webhook{
				Name:             datadog.String("my_webhook"),
				URL:              datadog.String("http://example.com"),
				UseCustomPayload: datadog.String("true"),
				CustomPayload:    datadog.String("field: value"),
			},
		},
		"custom payload form encoded": {
			terraformWebhook: map[string]interface{}{
				"name":               "my_webhook",
				"url":                "http://example.com",
				"use_custom_payload": true,
				"custom_payload":     "field: value",
				"encode_as_form":     true,
			},
			expectedDatadogWebhook: datadog.Webhook{
				Name:             datadog.String("my_webhook"),
				URL:              datadog.String("http://example.com"),
				UseCustomPayload: datadog.String("true"),
				CustomPayload:    datadog.String("field: value"),
				EncodeAsForm:     datadog.String("true"),
			},
		},
	}
	for name, tc := range cases {
		actualDatadogWebhook := buildDatadogWebhook(tc.terraformWebhook)

		if actualDatadogWebhook.Name == nil {
			t.Errorf("%s: Name is nil", name)
		}

		if *tc.expectedDatadogWebhook.Name != *actualDatadogWebhook.Name {
			t.Errorf("%s: Expected name to be '%s', but got '%s'", name, *tc.expectedDatadogWebhook.Name, *actualDatadogWebhook.Name)
		}

		if actualDatadogWebhook.URL == nil {
			t.Errorf("%s: URL is nil", name)
		}

		if *tc.expectedDatadogWebhook.URL != *actualDatadogWebhook.URL {
			t.Errorf("%s: Expected URL to be '%s', but got '%s'", name, *tc.expectedDatadogWebhook.URL, *actualDatadogWebhook.URL)
		}

		if tc.expectedDatadogWebhook.UseCustomPayload != nil {
			if actualDatadogWebhook.UseCustomPayload == nil {
				t.Errorf("%s: use custom payload field is nil", name)
			}

			if *tc.expectedDatadogWebhook.UseCustomPayload != *actualDatadogWebhook.UseCustomPayload {
				t.Errorf("%s: Expected use custom payload field to be '%s', but got '%s'", name, *tc.expectedDatadogWebhook.UseCustomPayload, *actualDatadogWebhook.UseCustomPayload)
			}
		}

		if tc.expectedDatadogWebhook.CustomPayload != nil {
			if actualDatadogWebhook.CustomPayload == nil {
				t.Errorf("%s: custom payload is nil", name)
			}

			if *tc.expectedDatadogWebhook.CustomPayload != *actualDatadogWebhook.CustomPayload {
				t.Errorf("%s: Expected custom payload to be '%s', but got '%s'", name, *tc.expectedDatadogWebhook.CustomPayload, *actualDatadogWebhook.CustomPayload)
			}
		}

		if tc.expectedDatadogWebhook.EncodeAsForm != nil {
			if actualDatadogWebhook.EncodeAsForm == nil {
				t.Errorf("%s: encode as form field is nil", name)
			}

			if *tc.expectedDatadogWebhook.EncodeAsForm != *actualDatadogWebhook.EncodeAsForm {
				t.Errorf("%s: Expected encode as form field to be '%s', but got '%s'", name, *tc.expectedDatadogWebhook.EncodeAsForm, *actualDatadogWebhook.EncodeAsForm)
			}
		}

		if tc.expectedDatadogWebhook.Headers != nil {
			if actualDatadogWebhook.Headers == nil {
				t.Errorf("%s: headers is nil", name)
			}

			if *tc.expectedDatadogWebhook.Headers != *actualDatadogWebhook.Headers {
				t.Errorf("%s: Expected headers to be '%s', but got '%s'", name, *tc.expectedDatadogWebhook.Headers, *actualDatadogWebhook.Headers)
			}
		}
	}
}

func TestBuildTerraformWebhook(t *testing.T) {
	cases := map[string]struct {
		datadogWebhooks           []datadog.Webhook
		expectedTerraformWebhooks []map[string]interface{}
		expectedError             error
	}{
		"required fields only": {
			datadogWebhooks: []datadog.Webhook{{
				Name: datadog.String("my_webhook"),
				URL:  datadog.String("http://example.com"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name": "my_webhook",
				"url":  "http://example.com",
			}},
			expectedError: nil,
		},
		"custom headers": {
			datadogWebhooks: []datadog.Webhook{{
				Name:    datadog.String("my_webhook"),
				URL:     datadog.String("http://example.com"),
				Headers: datadog.String("header1: val1"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name": "my_webhook",
				"url":  "http://example.com",
				"headers": map[string]interface{}{
					"header1": "val1",
				},
			}},
			expectedError: nil,
		},
		"custom payload": {
			datadogWebhooks: []datadog.Webhook{{
				Name:             datadog.String("my_webhook"),
				URL:              datadog.String("http://example.com"),
				UseCustomPayload: datadog.String("true"),
				CustomPayload:    datadog.String("field: value"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name":               "my_webhook",
				"url":                "http://example.com",
				"use_custom_payload": true,
				"custom_payload":     "field: value",
			}},
			expectedError: nil,
		},
		"custom payload form encoded": {
			datadogWebhooks: []datadog.Webhook{{
				Name:             datadog.String("my_webhook"),
				URL:              datadog.String("http://example.com"),
				UseCustomPayload: datadog.String("true"),
				CustomPayload:    datadog.String("field: value"),
				EncodeAsForm:     datadog.String("true"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name":               "my_webhook",
				"url":                "http://example.com",
				"use_custom_payload": true,
				"custom_payload":     "field: value",
				"encode_as_form":     true,
			}},
			expectedError: nil,
		},
		"non-bool useCustomHeader": {
			datadogWebhooks: []datadog.Webhook{{
				Name:             datadog.String("my_webhook"),
				URL:              datadog.String("http://example.com"),
				UseCustomPayload: datadog.String("I'm not a boolean"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name": "my_webhook",
				"url":  "http://example.com",
			}},
			expectedError: fmt.Errorf("strconv.ParseBool: parsing \"I'm not a boolean\": invalid syntax"),
		},
		"non-bool encodeAsForm": {
			datadogWebhooks: []datadog.Webhook{{
				Name:         datadog.String("my_webhook"),
				URL:          datadog.String("http://example.com"),
				EncodeAsForm: datadog.String("I'm also not a boolean"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name": "my_webhook",
				"url":  "http://example.com",
			}},
			expectedError: fmt.Errorf("strconv.ParseBool: parsing \"I'm also not a boolean\": invalid syntax"),
		},
		"no colon between header and value": {
			datadogWebhooks: []datadog.Webhook{{
				Name:    datadog.String("my_webhook"),
				URL:     datadog.String("http://example.com"),
				Headers: datadog.String("header1 val1"),
			}},
			expectedTerraformWebhooks: []map[string]interface{}{{
				"name": "my_webhook",
				"url":  "http://example.com",
			}},
			expectedError: fmt.Errorf("header not correctly formatted, expected ':' in 'header1 val1'"),
		},
	}
	for name, tc := range cases {
		actualTerraformWebhooks, err := buildTerraformWebhooks(tc.datadogWebhooks)

		if err != nil && tc.expectedError == nil {
			t.Errorf("%s: Unexpected error occured '%s", name, err.Error())
		} else if err == nil && tc.expectedError != nil {
			t.Errorf("%s: Expected error '%s', but it was not thrown", name, tc.expectedError.Error())
		} else if err != nil && tc.expectedError != nil && err.Error() != tc.expectedError.Error() {
			t.Errorf("%s: Unexpected error occured '%s', expected '%s'", name, err.Error(), tc.expectedError.Error())
		} else if actualTerraformWebhooks == nil && tc.expectedError == nil {
			t.Errorf("%s: Did not expect an error and terraform webhook is nil", name)
		} else if actualTerraformWebhooks != nil && tc.expectedError != nil {
			t.Errorf("%s: Expected error '%s', but got non-nil terraform webhook", name, tc.expectedError.Error())
		} else if actualTerraformWebhooks != nil {
			compareTerraformWebhooks(tc.expectedTerraformWebhooks, *actualTerraformWebhooks, name, t)
		}
	}
}

// Helpers
func compareTerraformHeaders(expectedTerraformHeaders map[string]interface{}, actualTerraformHeaders map[string]interface{}, name string, t *testing.T) {
	for header, actualVal := range actualTerraformHeaders {
		if expectedVal, ok := expectedTerraformHeaders[header]; ok {
			if expectedVal != actualVal {
				t.Errorf("%s: Expectd '%s', but got '%s' for header %s", name, expectedVal, actualVal, header)
			}
		} else {
			t.Errorf("%s: found header '%s' that was not expected", name, header)
		}
	}
}

func compareTerraformWebhooks(expectedTerraformWebhooks []map[string]interface{}, actualTerraformWebhooks []map[string]interface{}, name string, t *testing.T) {
	if len(expectedTerraformWebhooks) != len(actualTerraformWebhooks) {
		t.Errorf("%s: Expected %d webhooks, got %d", name, len(expectedTerraformWebhooks), len(actualTerraformWebhooks))
	}

	for _, actualTerraformWebhook := range actualTerraformWebhooks {
		found := false
		for _, expectedTerraformWebhook := range expectedTerraformWebhooks {
			if expectedTerraformWebhook["name"].(string) == *actualTerraformWebhook["name"].(*string) {
				found = true
				compareTerraformWebhook(expectedTerraformWebhook, actualTerraformWebhook, name, t)
			}
		}

		if !found {
			t.Errorf("%s: Did not find any expected webhooks with the name '%s'", name, *actualTerraformWebhook["name"].(*string))
		}
	}
}

func compareTerraformWebhook(expectedTerraformWebhook map[string]interface{}, actualTerraformWebhook map[string]interface{}, name string, t *testing.T) {
	expectedUrlVal, expectedUrlOk := expectedTerraformWebhook["url"]
	actualUrlVal, actualUrlOK := actualTerraformWebhook["url"]

	if expectedUrlOk != actualUrlOK {
		t.Errorf("%s: Excpected URL to be %s and was %s", name, isPresentToString(expectedUrlOk), isPresentToString(actualUrlOK))
	} else if expectedUrlOk && expectedUrlVal.(string) != *actualUrlVal.(*string) {
		t.Errorf("%s: Expected URL to be '%s', but got '%s'", name, expectedUrlVal.(string), *actualUrlVal.(*string))
	}

	expectedUseCustomPayloadVal, expectedUseCustomPayloadOk := expectedTerraformWebhook["use_custom_payload"]
	actualUseCustomPayloadVal, actualUseCustomPayloadOK := actualTerraformWebhook["use_custom_payload"]

	if expectedUseCustomPayloadOk != actualUseCustomPayloadOK {
		t.Errorf("%s: Excpected use custom payload field to be %s and was %s", name, isPresentToString(expectedUseCustomPayloadOk), isPresentToString(actualUseCustomPayloadOK))
	} else if expectedUseCustomPayloadOk && expectedUseCustomPayloadVal.(bool) != actualUseCustomPayloadVal.(bool) {
		t.Errorf("%s: Expected use custom payload field to be '%t', but got '%t'", name, expectedUseCustomPayloadVal.(bool), actualUseCustomPayloadVal.(bool))
	}

	expectedCustomPayloadVal, expectedCustomPayloadOk := expectedTerraformWebhook["custom_payload"]
	actualCustomPayloadVal, actualCustomPayloadOK := actualTerraformWebhook["custom_payload"]

	if expectedUseCustomPayloadOk != actualUseCustomPayloadOK {
		t.Errorf("%s: Excpected custom payload to be %s and was %s", name, isPresentToString(expectedCustomPayloadOk), isPresentToString(actualCustomPayloadOK))
	} else if expectedCustomPayloadOk && expectedCustomPayloadVal.(string) != actualCustomPayloadVal.(string) {
		t.Errorf("%s: Expected ustom payload to be '%s', but got '%s'", name, expectedCustomPayloadVal.(string), actualCustomPayloadVal.(string))
	}

	expectedEncodeAsFormVal, expectedEncodeAsFormOk := expectedTerraformWebhook["encode_as_form"]
	actualEncodeAsFormVal, actualEncodeAsFormOK := actualTerraformWebhook["encode_as_form"]

	if expectedEncodeAsFormOk != actualEncodeAsFormOK {
		t.Errorf("%s: Excpected encode as form field to be %s and was %s", name, isPresentToString(expectedEncodeAsFormOk), isPresentToString(actualEncodeAsFormOK))
	} else if expectedEncodeAsFormOk && expectedEncodeAsFormVal.(bool) != actualEncodeAsFormVal.(bool) {
		t.Errorf("%s: Expected use encode as form field to be '%t', but got '%t'", name, expectedEncodeAsFormVal.(bool), actualEncodeAsFormVal.(bool))
	}

	expectedHeadersVal, expectedHeadersOk := expectedTerraformWebhook["headers"]
	actualHeadersVal, actualHeadersOK := actualTerraformWebhook["headers"]

	if expectedHeadersOk != actualHeadersOK {
		t.Errorf("%s: Excpected URL to be %s and was %s", name, isPresentToString(expectedHeadersOk), isPresentToString(actualHeadersOK))
	} else if expectedHeadersOk {
		compareTerraformHeaders(expectedHeadersVal.(map[string]interface{}), actualHeadersVal.(map[string]interface{}), name, t)
	}
}

func isPresentToString(isPresent bool) string {
	if isPresent {
		return "present"
	}

	return "not present"
}

// Acceptance tests
func TestAccIntegrationWebhook_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIntegrationWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationWebhookConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationWebhookExistsAndValid("test_1", "http://example.com", false, "", false, ""),
				),
			},
		},
	})
}

func TestAccIntegrationWebhook_AllParams(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIntegrationWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationWebhookConfigAllParams,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationWebhookExistsAndValid("test_1", "http://example.com", true, "TITLE = $EVENT_TITLE", true, "X-DataDog-Event-Id: $ID"),
				),
			},
		},
	})
}

func TestAccIntegrationWebhook_BasicMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIntegrationWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationWebhookConfigBasicMultiple,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationWebhookExistsAndValid("test_1", "http://example.com", false, "", false, ""),
					testAccCheckDatadogIntegrationWebhookExistsAndValid("test_2", "https://another.example.com", false, "", false, ""),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationWebhookExistsAndValid(hookName string, url string, useCustomPayload bool, customPayload string, encodeAsForm bool, headers string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		slackIntegration, err := client.GetIntegrationWebhook()
		if err != nil {
			return fmt.Errorf("received an error retrieving integration webhook %s", err)
		}
		for _, hook := range slackIntegration.Webhooks {
			if hook.GetName() == hookName {

				if !hook.HasURL() && url != "" {
					return fmt.Errorf("expected encode as form field to be '%s', but was not set", url)
				} else if actual := hook.GetURL(); actual != url {
					return fmt.Errorf("expected URL to be '%s', but was '%s'", url, actual)
				}

				if !hook.HasUseCustomPayload() && useCustomPayload {
					return fmt.Errorf("expected use custom payload field to be '%t', but was not set", useCustomPayload)
				} else {
					actual, err := strconv.ParseBool(hook.GetUseCustomPayload())

					if err != nil {
						return fmt.Errorf("unexpected error occured: %s", err)
					}

					if actual != useCustomPayload {
						return fmt.Errorf("expected use custom payload field to be '%t', but was '%t'", useCustomPayload, actual)
					}
				}

				if !hook.HasCustomPayload() && customPayload != "" {
					return fmt.Errorf("expected encode as form field to be '%s', but was not set", customPayload)
				} else if actual := hook.GetCustomPayload(); actual != customPayload {
					return fmt.Errorf("expected custom payload to be '%s', but was '%s'", customPayload, actual)
				}

				if !hook.HasEncodeAsForm() && encodeAsForm {
					return fmt.Errorf("expected encode as form field to be '%t', but was not set", encodeAsForm)
				} else {
					actual, err := strconv.ParseBool(hook.GetEncodeAsForm())

					if err != nil {
						return fmt.Errorf("unexpected error occured: %s", err)
					}

					if actual != encodeAsForm {
						return fmt.Errorf("expected encode as form field to be '%t', but was '%t'", encodeAsForm, actual)
					}
				}

				if !hook.HasHeaders() && headers != "" {
					return fmt.Errorf("expected encode as form field to be '%s', but was not set", headers)
				} else if actual := hook.GetHeaders(); actual != headers {
					return fmt.Errorf("expected custom payload to be '%s', but was '%s'", actual, headers)
				}

				return nil
			}
		}
		return fmt.Errorf("didn't find hook (%s) in integration webhook", hookName)
	}
}

func testAccCheckIntegrationWebhookDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	_, err := client.GetIntegrationWebhook()
	if err != nil {
		if strings.Contains(err.Error(), "webhooks not found") {
			return nil
		}

		return fmt.Errorf("received an error retrieving integration webhook %s", err)
	}

	return fmt.Errorf("integration webhook is not properly destroyed")
}

const testAccIntegrationWebhookConfigBasic = `
resource "datadog_integration_webhook" "test" {
   hook {
	   name = "test_1"
	   url = "http://example.com"       
   }
}`

const testAccIntegrationWebhookConfigAllParams = `
resource "datadog_integration_webhook" "test" {
   hook {
	   name = "test_1"
	   url = "http://example.com"       
       use_custom_payload = true
       custom_payload = "TITLE = $EVENT_TITLE"
       encode_as_form = true
       headers = {
		   "X-DataDog-Event-Id": "$ID"
       }
   }
}`

const testAccIntegrationWebhookConfigBasicMultiple = `
resource "datadog_integration_webhook" "test" {
   hook {
	   name = "test_1"
	   url = "http://example.com"       
   }

	hook {
	   name = "test_2"
	   url = "https://another.example.com"       
   }
}`
