package datadog

import (
	"testing"

	"github.com/zorkian/go-datadog-api"
)

func TestBuildDatadogHeader(t *testing.T) {
	cases := map[string]struct {
		terraformHeaders       map[string]string
		expectedDatadogHeaders string
	}{
		"no headers": {
			terraformHeaders:       map[string]string{},
			expectedDatadogHeaders: "",
		},
		"single header": {
			terraformHeaders: map[string]string{
				"header1": "val1",
			},
			expectedDatadogHeaders: "header1: val1",
		},
		"multiple header": {
			terraformHeaders: map[string]string{
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
		expectedTerraformHeaders map[string]string
		expectedError            error
	}{
		"no headers": {
			datadogHeaders:           "",
			expectedTerraformHeaders: map[string]string{},
			expectedError:            nil,
		},
		"no headers with whitespace": {
			datadogHeaders:           "  \t",
			expectedTerraformHeaders: map[string]string{},
			expectedError:            nil,
		},
		"single header": {
			datadogHeaders: "header1: val1",
			expectedTerraformHeaders: map[string]string{
				"header1": "val1",
			},
			expectedError: nil,
		},
		"multiple headers": {
			datadogHeaders: "header1: val1\nheader2: val2\nheader3: val3",
			expectedTerraformHeaders: map[string]string{
				"header1": "val1",
				"header2": "val2",
				"header3": "val3",
			},
			expectedError: nil,
		},
	}
	for name, tc := range cases {
		actualTerraformHeaders, err := buildTerraformHeader(&tc.datadogHeaders)

		if err != tc.expectedError {
			if err != nil {
				t.Errorf("%s: Unexpected error occured '%s', expected '%s'", name, err.Error(), tc.expectedError.Error())
			} else {
				t.Errorf("%s: Expected error '%s', but none was thrown", name, tc.expectedError.Error())
			}
			return
		}

		if actualTerraformHeaders == nil && tc.expectedError == nil {
			t.Errorf("%s: Did not expect an error and terraform webhook is nil", name)
		} else if tc.expectedError != nil {
			t.Errorf("%s: Expected error '%s', but got non-nil terraform webhook", name, tc.expectedError.Error())
		} else if actualTerraformHeaders != nil {
			compareTerraformHeaders(tc.expectedTerraformHeaders, *actualTerraformHeaders, name, t)
		}
	}
}

func compareTerraformHeaders(expectedTerraformHeaders map[string]string, actualTerraformHeaders map[string]string, name string, t *testing.T) {
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
				"headers": map[string]string{
					"header1": "val1",
				},
			},
			expectedDatadogWebhook: datadog.Webhook{
				Name: datadog.String("my_webhook"),
				URL:  datadog.String("http://example.com"),
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
				"headers": map[string]string{
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
	}
	for name, tc := range cases {
		actualTerraformWebhooks, err := buildTerraformWebhooks(tc.datadogWebhooks)

		if err != tc.expectedError {
			if err != nil {
				t.Errorf("%s: Unexpected error occured '%s', expected '%s'", name, err.Error(), tc.expectedError.Error())
			} else {
				t.Errorf("%s: Expected error '%s', but none was thrown", name, tc.expectedError.Error())
			}
			return
		}

		if actualTerraformWebhooks == nil && tc.expectedError == nil {
			t.Errorf("%s: Did not expect an error and terraform webhook is nil", name)
		} else if tc.expectedError != nil {
			t.Errorf("%s: Expected error '%s', but got non-nil terraform webhook", name, tc.expectedError.Error())
		} else if actualTerraformWebhooks != nil {
			compareTerraformWebhooks(tc.expectedTerraformWebhooks, *actualTerraformWebhooks, name, t)
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
	} else if expectedCustomPayloadOk && expectedCustomPayloadVal.(string) != *actualCustomPayloadVal.(*string) {
		t.Errorf("%s: Expected ustom payload to be '%s', but got '%s'", name, expectedCustomPayloadVal.(string), *actualCustomPayloadVal.(*string))
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
		compareTerraformHeaders(expectedHeadersVal.(map[string]string), *actualHeadersVal.(*map[string]string), name, t)
	}
}

func isPresentToString(isPresent bool) string {
	if isPresent {
		return "present"
	}

	return "not present"
}
