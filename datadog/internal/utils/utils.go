package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/terraform-providers/terraform-provider-datadog/version"
)

// DatadogProvider holds a reference to the provider
var DatadogProvider *schema.Provider

// TranslateClientError turns an error into a message
func TranslateClientError(err error, httpresp *http.Response, msg string) error {
	if msg == "" {
		msg = "an error occurred"
	}

	if httpresp != nil && httpresp.Request != nil {
		msg = fmt.Sprintf("%s from %s", msg, httpresp.Request.URL.String())
	}

	if apiErr, ok := err.(CustomRequestAPIError); ok {
		return fmt.Errorf(msg+": %v: %s", err, apiErr.Body())
	}
	if apiErr, ok := err.(datadogV1.GenericOpenAPIError); ok {
		return fmt.Errorf(msg+": %v: %s", err, apiErr.Body())
	}
	if apiErr, ok := err.(datadogV2.GenericOpenAPIError); ok {
		return fmt.Errorf(msg+": %v: %s", err, apiErr.Body())
	}
	if errURL, ok := err.(*url.Error); ok {
		return fmt.Errorf(msg+" (url.Error): %s", errURL)
	}

	return fmt.Errorf(msg+": %s", err.Error())
}

// CheckForUnparsed takes in a API response object and returns an error if it contains an unparsed element
func CheckForUnparsed(resp interface{}) error {
	if unparsed, invalidPart := datadogV1.ContainsUnparsedObject(resp); unparsed {
		return fmt.Errorf("object contains unparsed element: %+v", invalidPart)
	}
	return nil
}

// TranslateClientErrorDiag returns client error as type diag.Diagnostics
func TranslateClientErrorDiag(err error, httpresp *http.Response, msg string) diag.Diagnostics {
	return diag.FromErr(TranslateClientError(err, httpresp, msg))
}

// GetUserAgent augments the default user agent with provider details
func GetUserAgent(clientUserAgent string) string {
	return fmt.Sprintf("terraform-provider-datadog/%s (terraform %s; terraform-cli %s) %s",
		version.ProviderVersion,
		meta.SDKVersionString(),
		DatadogProvider.TerraformVersion,
		clientUserAgent)
}

// GetMetadataFromJSON decodes passed JSON data
func GetMetadataFromJSON(jsonBytes []byte, unmarshalled interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(jsonBytes))
	// make sure we return errors on attributes that we don't expect in metadata
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(unmarshalled); err != nil {
		return fmt.Errorf("failed to unmarshal metadata_json: %s", err)
	}
	return nil
}

// ConvertToSha256 builds a SHA256 hash of the passed string
func ConvertToSha256(content string) string {
	data := []byte(content)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}

// AccountAndNamespaceFromID returns account and namespace from an ID
func AccountAndNamespaceFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account ID and namespace: %s", id)
	}
	return result[0], result[1], nil
}

// AccountAndRoleFromID returns account and role from an ID
func AccountAndRoleFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account ID and Role name from an Amazon Web Services integration id: %s", id)
	}
	return result[0], result[1], nil
}

// AccountAndLambdaArnFromID returns account and Lambda ARN from an ID
func AccountAndLambdaArnFromID(id string) (string, string, error) {
	result := strings.Split(id, " ")
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: %s", id)
	}
	return result[0], result[1], nil
}

// TenantAndClientFromID returns azure account and client from an ID
func TenantAndClientFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting tenant name and client ID from an Azure integration id: %s", id)
	}
	return result[0], result[1], nil
}

// AccountNameAndChannelNameFromID returns slack account and channel from an ID
func AccountNameAndChannelNameFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account name and channel name: %s", id)
	}
	return result[0], result[1], nil
}

// ConvertResponseByteToMap converts JSON []byte to map[string]interface{}
func ConvertResponseByteToMap(b []byte) (map[string]interface{}, error) {
	convertedMap := make(map[string]interface{})
	err := json.Unmarshal(b, &convertedMap)
	if err != nil {
		return nil, err
	}

	return convertedMap, nil
}
