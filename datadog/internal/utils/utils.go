package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

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

// DDOrgUUIDEnvName name of env var for Org UUID
const DDOrgUUIDEnvName = "DD_ORG_UUID"

// DatadogAPPKeyEnvName name of env var for APP key
const DatadogAPPKeyEnvName = "DATADOG_APP_KEY"

// DatadogAPIKeyEnvName name of env var for API key
const DatadogAPIKeyEnvName = "DATADOG_API_KEY"

// DatadogOrgUUIDEnvName name of env var for Org UUID
const DatadogOrgUUIDEnvName = "DATADOG_ORG_UUID"

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

// AWSAccessKeyId name of env var for AWS Access Key Id
const AWSAccessKeyId = "AWS_ACCESS_KEY_ID"

// AWSSecretAccessKey name of env var for AWS Secret Access Key
const AWSSecretAccessKey = "AWS_SECRET_ACCESS_KEY"

// AWSSessionToken name of env var for AWS Session Token
const AWSSessionToken = "AWS_SESSION_TOKEN"

// BaseIPRangesSubdomain ip ranges subdomain
const BaseIPRangesSubdomain = "ip-ranges"

// APPKeyEnvVars names of env var for APP key
var APPKeyEnvVars = []string{DDAPPKeyEnvName, DatadogAPPKeyEnvName}

// APIKeyEnvVars names of env var for API key
var APIKeyEnvVars = []string{DDAPIKeyEnvName, DatadogAPIKeyEnvName}

// OrgUUIDEnvVars names of env var for Org UUID
var OrgUUIDEnvVars = []string{DDOrgUUIDEnvName, DatadogOrgUUIDEnvName}

// APIUrlEnvVars names of env var for API key
var APIUrlEnvVars = []string{DDAPIUrlEnvName, DatadogAPIUrlEnvName}

// DatadogProvider holds a reference to the provider
var DatadogProvider *schema.Provider

// IntegrationAwsMutex mutex for AWS Integration resources
var IntegrationAwsMutex = sync.Mutex{}

// Resource minimal interface common to ResourceData and ResourceDiff
type Resource interface {
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
}

// NewTransport returns new transport with default values borrowed from http.DefaultTransport
func NewTransport() *http.Transport {
	return &http.Transport{
		// Default values copied from http.DefaultTransport
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       45 * time.Second, // Reduced idle connection timeout from default of 90s
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewHTTPClient returns new http.Client
func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
	}
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
		msg = fmt.Sprintf("%s from %s", msg, httpresp.Request.URL.EscapedPath())
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

// CheckForAdditionalProperties takes in a API object and returns an error if it contains an additional property
func CheckForAdditionalProperties(resp interface{}) error {
	if unparsed, invalidPart := ContainsAdditionalProperties(resp); unparsed {
		return fmt.Errorf("object contains additional property: %+v", invalidPart)
	}
	return nil
}

// ContainsAdditionalProperties returns true if the given data contains an additional properties.
func ContainsAdditionalProperties(i interface{}) (bool, interface{}) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if n, m := ContainsAdditionalProperties(v.Index(i).Interface()); n {
				return n, m
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if n, m := ContainsAdditionalProperties(v.MapIndex(k).Interface()); n {
				return n, m
			}
		}
	case reflect.Struct:
		if u := v.FieldByName("AdditionalProperties"); u.IsValid() && !u.IsNil() {
			return true, u.Interface()
		}
		for i := 0; i < v.NumField(); i++ {
			if fn := v.Type().Field(i).Name; string(fn[0]) == strings.ToUpper(string(fn[0])) && fn != "AdditionalProperties" {
				if n, m := ContainsAdditionalProperties(v.Field(i).Interface()); n {
					return n, m
				}
			} else if fn == "value" { // Special case for Nullables
				if get := v.MethodByName("Get"); get.IsValid() {
					if n, m := ContainsAdditionalProperties(get.Call([]reflect.Value{})[0].Interface()); n {
						return n, m
					}
				}
			}
		}
	case reflect.Interface, reflect.Ptr:
		if !v.IsNil() {
			return ContainsAdditionalProperties(v.Elem().Interface())
		}
	default:
		if v.IsValid() {
			if m := v.MethodByName("IsValid"); m.IsValid() {
				if !m.Call([]reflect.Value{})[0].Bool() {
					return true, v.Interface()
				}
			}
		}
	}
	return false, nil
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

func NormalizeIPAddress(ipAddress string) string {
	_, ipNet, err := net.ParseCIDR(ipAddress)
	if err != nil {
		ip := net.ParseIP(ipAddress)
		if ip == nil {
			return ""
		}
		// ipAddress is a single IP address
		// if it is ipv4, the prefix is 32. if ipv6, it is 128
		prefix := "32"
		if ip.DefaultMask() == nil {
			prefix = "128"
		}
		return fmt.Sprintf("%v/%v", ip, prefix)
	}
	return ipNet.String()
}

func StringSliceDifference(slice1, slice2 []string) []string {
	elements := make(map[string]bool)
	for _, val := range slice2 {
		elements[val] = true
	}

	var diff []string
	for _, val := range slice1 {
		if !elements[val] {
			diff = append(diff, val)
		}
	}
	return diff
}

// fast isAlpha for ascii
func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// fast isAlphaNumeric for ascii
func isAlphaNum(b byte) bool {
	return isAlpha(b) || (b >= '0' && b <= '9')
}

// ValidateMetricName ensures the given metric name length is in [0, MaxMetricLen] and
// contains at least one alphabetic character whose index is returned
func ValidateMetricName(name string) (int, error) {
	var i int
	if name == "" {
		return 0, fmt.Errorf("metric name is empty")
	}

	// skip non-alphabetic characters
	for ; i < len(name) && !isAlpha(name[i]); i++ {
	}

	// if there were no alphabetic characters it wasn't valid
	if i == len(name) {
		return 0, fmt.Errorf("metric name %s is invalid. it must contain at least one alphabetic character", name)
	}

	return i, nil
}

// NormMetricNameParse normalizes metric names with a parser instead of using
// garbage-creating string replacement routines.
func NormMetricNameParse(name string) string {
	i, err := ValidateMetricName(name)
	if err != nil {
		return name
	}

	var ptr int
	res := make([]byte, 0, len(name))

	for ; i < len(name); i++ {
		switch {
		case isAlphaNum(name[i]):
			res = append(res, name[i])
			ptr++
		case name[i] == '.':
			// we skipped all non-alpha chars up front so we have seen at least one
			switch res[ptr-1] {
			// overwrite underscores that happen before periods
			case '_':
				res[ptr-1] = '.'
			default:
				res = append(res, '.')
				ptr++
			}
		default:
			// we skipped all non-alpha chars up front so we have seen at least one
			switch res[ptr-1] {
			// no double underscores, no underscores after periods
			case '.', '_':
			default:
				res = append(res, '_')
				ptr++
			}
		}
	}

	if res[ptr-1] == '_' {
		res = res[:ptr-1]
	}
	// safe because res does not escape this function
	return string(res)

}

// AnyToSlice casts a raw interface{} to a well-typed slice (useful for reading Terraform ResourceData)
func AnyToSlice[T any](raw any) []T {
	rawSlice := raw.([]interface{})
	result := make([]T, len(rawSlice))
	for i, x := range rawSlice {
		result[i] = x.(T)
	}
	return result
}

// RemoveEmptyValuesInMap removes empty arrays, maps, and null values from a map.
// This ensures we don't treat empty/null values as meaningful differences when comparing maps.
func RemoveEmptyValuesInMap(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case []any:
			if len(val) == 0 {
				delete(m, k)
			} else {
				for _, item := range val {
					if itemMap, ok := item.(map[string]any); ok {
						RemoveEmptyValuesInMap(itemMap)
					}
				}
			}
		case map[string]any:
			RemoveEmptyValuesInMap(val)
			if len(val) == 0 {
				delete(m, k)
			}
		case nil:
			delete(m, k)
		}
	}
}

// Reference: https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/v0.2.0/jsontypes/normalized_value.go
// StringSemanticEquals returns true if the given JSON string value is semantically equal to the current JSON string value. When compared,
// these JSON string values are "normalized" by marshalling them to empty Go structs. This prevents Terraform data consistency errors and
// resource drift due to inconsequential differences in the JSON strings (whitespace, property order, etc), similar to jsontypes.Normalized,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
func AppJSONStringSemanticEquals(s1 string, s2 string) (bool, frameworkDiag.Diagnostics) {
	var diags frameworkDiag.Diagnostics

	normalizedS1, err := normalizeAppBuilderAppJSONString(s1)
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return false, diags
	}

	normalizedS2, err := normalizeAppBuilderAppJSONString(s2)
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return false, diags
	}

	return normalizedS1 == normalizedS2, diags
}

func normalizeAppBuilderAppJSONString(jsonStr string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonStr))

	// This ensures the JSON decoder will not parse JSON numbers into Go's float64 type; avoiding Go
	// normalizing the JSON number representation or imposing limits on numeric range. See the unit test cases
	// of StringSemanticEquals for examples.
	dec.UseNumber()

	var temp any
	if err := dec.Decode(&temp); err != nil {
		return "", err
	}

	// feature specific to AppBuilderAppStringValue:
	// we only want to compare fields that matter to Create/Update requests when comparing JSON strings
	if jsonMap, ok := temp.(map[string]any); ok {
		// fields that would get excluded in this comparison would be "id", "favorite", "selfService", "tags", "connections", "deployment"
		// these fields are included in the App JSON but don't make a difference when calling Create/Update endpoints
		// the logic focuses on fields to keep because newer but irrelevant fields may be added to the App JSON in the future
		fieldsToKeep := []string{"components", "description", "name", "queries", "rootInstanceName"}

		newJsonMap := make(map[string]any)
		for _, field := range fieldsToKeep {
			if val, ok := jsonMap[field]; ok {
				newJsonMap[field] = val
			}
		}

		temp = newJsonMap
	}

	jsonBytes, err := json.Marshal(&temp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func UseMonitorFrameworkProvider() bool {
	return getEnv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "false") == "true"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
