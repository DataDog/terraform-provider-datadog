package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/terraform-providers/terraform-provider-datadog/version"
)

// DDAPPKeyEnvName name of env var for APP key
const DDAPPKeyEnvName = "DD_APP_KEY"

// DDAPIKeyEnvName name of env var for API key
const DDAPIKeyEnvName = "DD_API_KEY"

// DDAPIUrlEnvName name of env var for API key
const DDAPIUrlEnvName = "DD_HOST"

// DatadogAPPKeyEnvName name of env var for APP key
const DatadogAPPKeyEnvName = "DATADOG_APP_KEY"

// DatadogAPIKeyEnvName name of env var for API key
const DatadogAPIKeyEnvName = "DATADOG_API_KEY"

// DatadogAPIUrlEnvName name of env var for API key
const DatadogAPIUrlEnvName = "DATADOG_HOST"

// DDHTTPRetryEnabled name of env var for retry enabled
const DDHTTPRetryEnabled = "DD_HTTP_CLIENT_RETRY_ENABLED"

// DDHTTPRetryTimeout name of env var for retry timeout
const DDHTTPRetryTimeout = "DD_HTTP_CLIENT_RETRY_TIMEOUT"

// DDHTTPRetryBackoffMultiplier name of env var for retry backoff multiplier
const DDHTTPRetryBackoffMultiplier = "DD_HTTP_CLIENT_RETRY_BACKOFF_MULTIPLIER"

// DDHTTPRetryBackoffBase name of env var for retry backoff base
const DDHTTPRetryBackoffBase = "DD_HTTP_CLIENT_RETRY_BACKOFF_BASE"

// DDHTTPRetryMaxRetries name of env var for max retries
const DDHTTPRetryMaxRetries = "DD_HTTP_CLIENT_RETRY_MAX_RETRIES"

// BaseIPRangesSubdomain ip ranges subdomain
const BaseIPRangesSubdomain = "ip-ranges"

// APPKeyEnvVars names of env var for APP key
var APPKeyEnvVars = []string{DDAPPKeyEnvName, DatadogAPPKeyEnvName}

// APIKeyEnvVars names of env var for API key
var APIKeyEnvVars = []string{DDAPIKeyEnvName, DatadogAPIKeyEnvName}

// APIUrlEnvVars names of env var for API key
var APIUrlEnvVars = []string{DDAPIUrlEnvName, DatadogAPIUrlEnvName}

// DatadogProvider holds a reference to the provider
var DatadogProvider *schema.Provider

// Resource minimal interface common to ResourceData and ResourceDiff
type Resource interface {
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
}

// FrameworkErrorDiag return error diag
func FrameworkErrorDiag(err error, msg string) frameworkDiag.ErrorDiagnostic {
	var summary string

	switch v := err.(type) {
	case CustomRequestAPIError:
		summary = fmt.Sprintf("%v: %s", err, v.Body())
	case datadog.GenericOpenAPIError:
		summary = fmt.Sprintf("%v: %s", err, v.Body())
	case *url.Error:
		summary = fmt.Sprintf("url.Error: %s ", v.Error())
	default:
		summary = v.Error()
	}

	return frameworkDiag.NewErrorDiagnostic(msg, summary)
}

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
	if apiErr, ok := err.(datadog.GenericOpenAPIError); ok {
		return fmt.Errorf(msg+": %v: %s", err, apiErr.Body())
	}
	if errURL, ok := err.(*url.Error); ok {
		return fmt.Errorf(msg+" (url.Error): %s", errURL)
	}

	return fmt.Errorf(msg+": %s", err.Error())
}

// CheckForUnparsed takes in a API response object and returns an error if it contains an unparsed element
func CheckForUnparsed(resp interface{}) error {
	if unparsed, invalidPart := datadog.ContainsUnparsedObject(resp); unparsed {
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

// GetUserAgentFramework augments the default user agent with provider details for framework provider
func GetUserAgentFramework(clientUserAgent, tfCLIVersion string) string {
	return fmt.Sprintf("terraform-provider-datadog/%s (terraform-cli %s) %s",
		version.ProviderVersion,
		tfCLIVersion,
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

// AccountIDAndResourceIDFromID returns confluent resource account_id and resource_id from the ID
func AccountIDAndResourceIDFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account_id and resource_id from id: %s", id)
	}
	return result[0], result[1], nil
}

// AccountIDAndServiceIDFromID returns fastly service resource account_id and service_id from the ID
func AccountIDAndServiceIDFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account_id and service_id from id: %s", id)
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

// DeleteKeyInMap deletes key (in dot notation) in map
func DeleteKeyInMap(mapObject map[string]interface{}, keyList []string) {
	if len(keyList) == 1 {
		delete(mapObject, keyList[0])
	} else if m, ok := mapObject[keyList[0]].(map[string]interface{}); ok {
		DeleteKeyInMap(m, keyList[1:])
	}

	return
}

// GetStringSlice returns string slice for the given key if present, otherwise returns an empty slice
func GetStringSlice(d Resource, key string) []string {
	if v, ok := d.GetOk(key); ok {
		values := v.([]interface{})
		stringValues := make([]string, len(values))
		for i, value := range values {
			stringValues[i] = value.(string)
		}
		return stringValues
	}
	return []string{}
}

// GetMultiEnvVar returns first matching env var
func GetMultiEnvVar(envVars ...string) (string, error) {
	for _, value := range envVars {
		if v := os.Getenv(value); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("unable to retrieve any env vars from list: %v", envVars)
}

func ResourceIDAttribute() frameworkSchema.StringAttribute {
	return frameworkSchema.StringAttribute{
		Description: "The ID of this resource.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}
