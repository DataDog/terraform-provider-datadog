package test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/transport"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jonboulle/clockwork"
	datadogCommunity "github.com/zorkian/go-datadog-api"
	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	ddtesting "gopkg.in/DataDog/dd-trace-go.v1/contrib/testing"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type clockContextKey string

const ddTestOrg = "fasjyydbcgwwc2uc"
const testAPIKeyEnvName = "DD_TEST_CLIENT_API_KEY"
const testAPPKeyEnvName = "DD_TEST_CLIENT_APP_KEY"
const testOrgEnvName = "DD_TEST_ORG"

var isTestOrgC *bool

var testFiles2EndpointTags = map[string]string{
	"tests/data_source_datadog_dashboard_test":                         "dashboard",
	"tests/data_source_datadog_dashboard_list_test":                    "dashboard-lists",
	"tests/data_source_datadog_ip_ranges_test":                         "ip-ranges",
	"tests/data_source_datadog_monitor_test":                           "monitors",
	"tests/data_source_datadog_monitors_test":                          "monitors",
	"tests/data_source_datadog_permissions_test":                       "permissions",
	"tests/data_source_datadog_role_test":                              "roles",
	"tests/data_source_datadog_security_monitoring_rules_test":         "security-monitoring",
	"tests/data_source_datadog_security_monitoring_filters_test":         "security-monitoring",
	"tests/data_source_datadog_service_level_objective_test":           "service-level-objectives",
	"tests/data_source_datadog_service_level_objectives_test":          "service-level-objectives",
	"tests/data_source_datadog_synthetics_locations_test":              "synthetics",
	"tests/import_datadog_downtime_test":                               "downtimes",
	"tests/import_datadog_integration_pagerduty_test":                  "integration-pagerduty",
	"tests/import_datadog_logs_pipeline_test":                          "logs-pipelines",
	"tests/import_datadog_monitor_test":                                "monitors",
	"tests/import_datadog_user_test":                                   "users",
	"tests/provider_test":                                              "terraform",
	"tests/resource_datadog_dashboard_alert_graph_test":                "dashboards",
	"tests/resource_datadog_dashboard_alert_value_test":                "dashboards",
	"tests/resource_datadog_dashboard_change_test":                     "dashboards",
	"tests/resource_datadog_dashboard_check_status_test":               "dashboards",
	"tests/resource_datadog_dashboard_distribution_test":               "dashboards",
	"tests/resource_datadog_dashboard_event_stream_test":               "dashboards",
	"tests/resource_datadog_dashboard_event_timeline_test":             "dashboards",
	"tests/resource_datadog_dashboard_free_text_test":                  "dashboards",
	"tests/resource_datadog_dashboard_heatmap_test":                    "dashboards",
	"tests/resource_datadog_dashboard_hostmap_test":                    "dashboards",
	"tests/resource_datadog_dashboard_iframe_test":                     "dashboards",
	"tests/resource_datadog_dashboard_image_test":                      "dashboards",
	"tests/resource_datadog_dashboard_list_test":                       "dashboard-lists",
	"tests/resource_datadog_dashboard_log_stream_test":                 "dashboards",
	"tests/resource_datadog_dashboard_manage_status_test":              "dashboards",
	"tests/resource_datadog_dashboard_note_test":                       "dashboards",
	"tests/resource_datadog_dashboard_query_table_test":                "dashboards",
	"tests/resource_datadog_dashboard_query_value_test":                "dashboards",
	"tests/resource_datadog_dashboard_scatterplot_test":                "dashboards",
	"tests/resource_datadog_dashboard_service_map_test":                "dashboards",
	"tests/resource_datadog_dashboard_slo_test":                        "dashboards",
	"tests/resource_datadog_dashboard_test":                            "dashboards",
	"tests/resource_datadog_dashboard_timeseries_test":                 "dashboards",
	"tests/resource_datadog_dashboard_top_list_test":                   "dashboards",
	"tests/resource_datadog_dashboard_trace_service_test":              "dashboards",
	"tests/resource_datadog_dashboard_json_test":                       "dashboards-json",
	"tests/resource_datadog_downtime_test":                             "downtimes",
	"tests/resource_datadog_dashboard_geomap_test":                     "dashboards",
	"tests/resource_datadog_integration_aws_lambda_arn_test":           "integration-aws",
	"tests/resource_datadog_integration_aws_log_collection_test":       "integration-aws",
	"tests/resource_datadog_integration_aws_tag_filter_test":           "integration-aws",
	"tests/resource_datadog_integration_aws_test":                      "integration-aws",
	"tests/resource_datadog_integration_azure_test":                    "integration-azure",
	"tests/resource_datadog_integration_gcp_test":                      "integration-gcp",
	"tests/resource_datadog_integration_pagerduty_service_object_test": "integration-pagerduty",
	"tests/resource_datadog_integration_pagerduty_test":                "integration-pagerduty",
	"tests/resource_datadog_integration_slack_channel_test":            "integration-slack-channel",
	"tests/resource_datadog_logs_archive_test":                         "logs-archive",
	"tests/resource_datadog_logs_archive_order_test":                   "logs-archive-order",
	"tests/resource_datadog_logs_custom_pipeline_test":                 "logs-pipelines",
	"tests/resource_datadog_logs_metric_test":                          "logs-metric",
	"tests/resource_datadog_metric_metadata_test":                      "metrics",
	"tests/resource_datadog_metric_tag_configuration_test":             "metrics",
	"tests/resource_datadog_monitor_test":                              "monitors",
	"tests/resource_datadog_role_test":                                 "roles",
	"tests/resource_datadog_screenboard_test":                          "dashboards",
	"tests/resource_datadog_security_monitoring_default_rule_test":     "security-monitoring",
	"tests/resource_datadog_security_monitoring_rule_test":             "security-monitoring",
	"tests/resource_datadog_security_monitoring_filter_test":           "security-monitoring",
	"tests/resource_datadog_service_level_objective_test":              "service-level-objectives",
	"tests/resource_datadog_slo_correction_test":                       "slo_correction",
	"tests/resource_datadog_synthetics_test_test":                      "synthetics",
	"tests/resource_datadog_synthetics_global_variable_test":           "synthetics",
	"tests/resource_datadog_synthetics_private_location_test":          "synthetics",
	"tests/resource_datadog_timeboard_test":                            "dashboards",
	"tests/resource_datadog_user_test":                                 "users",
}

// getEndpointTagValue traverses callstack frames to find the test function that invoked this call;
// it then matches the file defining this function against testFiles2EndpointTags to figure out
// the tag value to set on span
func getEndpointTagValue(t *testing.T) (string, error) {
	var pcs [512]uintptr
	var frame runtime.Frame
	more := true
	n := runtime.Callers(1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	functionFile := ""
	testName := t.Name()
	for more {
		frame, more = frames.Next()
		// nested test functions like `TestAuthenticationValidate/200_Valid` will have frame.Function ending with
		// ".funcX", `e.g. datadog.TestAuthenticationValidate.func1`, so trim everything after last "/" in test name
		// and everything after last "." in frame function name
		frameFunction := frame.Function
		if strings.Contains(testName, "/") {
			testName = testName[:strings.LastIndex(testName, "/")]
			frameFunction = frameFunction[:strings.LastIndex(frameFunction, ".")]
		}
		if strings.HasSuffix(frameFunction, "."+testName) {
			functionFile = frame.File
			// when we find the frame with the current test function, match it against testFiles2EndpointTags
			for file, tag := range testFiles2EndpointTags {
				if strings.HasSuffix(functionFile, fmt.Sprintf("datadog/%s.go", file)) {
					return tag, nil
				}
			}
		}
	}
	return "", fmt.Errorf(
		"Endpoint tag for test file %s not found in datadog/provider_test.go, please add it to `testFiles2EndpointTags`",
		functionFile)
}

func isRecording() bool {
	return os.Getenv("RECORD") == "true"
}

func isReplaying() bool {
	return os.Getenv("RECORD") == "false"
}

func isDebug() bool {
	return os.Getenv("DEBUG") == "true"
}

func isAPIKeySet() bool {
	if os.Getenv(testAPIKeyEnvName) != "" {
		return true
	}
	return false
}

func isAPPKeySet() bool {
	if os.Getenv(testAPPKeyEnvName) != "" {
		return true
	}
	return false
}

func isTestOrg() bool {
	if isTestOrgC != nil {
		return *isTestOrgC
	}
	// If keys belong to test org, then this get will succeed, otherwise it will fail with 400
	publicID := ddTestOrg
	if v := os.Getenv(testOrgEnvName); v != "" {
		publicID = v
	}
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.datadoghq.com/api/v1/org/"+publicID, nil)
	req.Header.Add("DD-API-KEY", os.Getenv(testAPIKeyEnvName))
	req.Header.Add("DD-APPLICATION-KEY", os.Getenv(testAPPKeyEnvName))
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		r := false
		isTestOrgC = &r
		return r
	}
	r := true
	isTestOrgC = &r
	return r
}

// isCIRun returns true if the CI environment variable is set to "true"
func isCIRun() bool {
	return os.Getenv("CI") == "true"
}

func setClock(t *testing.T) clockwork.FakeClock {
	os.MkdirAll("cassettes", 0755)
	f, err := os.Create(fmt.Sprintf("cassettes/%s.freeze", t.Name()))
	if err != nil {
		t.Fatalf("Could not set clock: %v", err)
	}
	defer f.Close()
	now := clockwork.NewRealClock().Now()
	f.WriteString(now.Format(time.RFC3339Nano))
	return clockwork.NewFakeClockAt(now)
}

func restoreClock(t *testing.T) clockwork.FakeClock {
	data, err := ioutil.ReadFile(fmt.Sprintf("cassettes/%s.freeze", t.Name()))
	if err != nil {
		t.Logf("Could not load clock: %v", err)
		return setClock(t)
	}
	now, err := time.Parse(time.RFC3339Nano, string(data))
	if err != nil {
		t.Fatalf("Could not parse clock date: %v", err)
	}
	return clockwork.NewFakeClockAt(now)
}

func testClock(t *testing.T) clockwork.FakeClock {
	if isRecording() {
		return setClock(t)
	} else if isReplaying() {
		return restoreClock(t)
	}
	// do not set or restore frozen time
	return clockwork.NewFakeClockAt(clockwork.NewRealClock().Now())
}

func clockFromContext(ctx context.Context) clockwork.FakeClock {
	return ctx.Value(clockContextKey("clock")).(clockwork.FakeClock)
}

// uniqueEntityName will return a unique string that can be used as a title/description/summary/...
// of an API entity. When used in Azure Pipelines and RECORD=true or RECORD=none, it will include
// BuildId to enable mapping resources that weren't deleted to builds.
func uniqueEntityName(ctx context.Context, t *testing.T) string {
	name := withUniqueSurrounding(clockFromContext(ctx), t.Name())
	return name
}

// SecurePath replaces all dangerous characters in the path.
func SecurePath(path string) string {
	badChars := []string{"\\", "?", "%", "*", ":", "|", `"`, "<", ">"}
	for _, c := range badChars {
		path = strings.ReplaceAll(path, c, "_")
	}
	return filepath.Clean(path)
}

// withUniqueSurrounding will wrap a string that can be used as a title/description/summary/...
// of an API entity. When used in Azure Pipelines and RECORD=true or RECORD=none, it will include
// BuildId to enable mapping resources that weren't deleted to builds.
func withUniqueSurrounding(clock clockwork.FakeClock, name string) string {
	buildID, present := os.LookupEnv("BUILD_BUILDID")
	if !present || !isCIRun() || isReplaying() {
		buildID = "local"
	}

	// NOTE: some endpoints have limits on certain fields (e.g. Roles V2 names can only be 55 chars long),
	// so we need to keep this short
	result := fmt.Sprintf("tf-%s-%s-%d", SecurePath(name), buildID, clock.Now().Unix())
	// In case this is used in URL, make sure we replace the slash that is added by subtests
	result = strings.ReplaceAll(result, "/", "-")
	return result
}

// uniqueAWSAccountID takes uniqueEntityName result, hashes it to get a unique string
// and then returns first 12 characters (numerical only), so that the value can be used
// as AWS account ID and is still as unique as possible, it changes in CI, but is stable locally
func uniqueAWSAccountID(ctx context.Context, t *testing.T) string {
	uniq := uniqueEntityName(ctx, t)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniq)))
	result := ""
	for _, r := range hash {
		result = fmt.Sprintf("%s%s", result, strconv.Itoa(int(r)))
	}
	return result[:12]
}

// uniqueAWSAccessKeyID takes uniqueEntityName result, hashes it to get a unique string
// and then returns first 16 characters (numerical only), so that the value can be used
// as AWS account ID and is still as unique as possible, it changes in CI, but is stable locally
func uniqueAWSAccessKeyID(ctx context.Context, t *testing.T) string {
	uniq := uniqueEntityName(ctx, t)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniq)))
	result := ""
	for _, r := range hash {
		result = fmt.Sprintf("%s%s", result, strconv.Itoa(int(r)))
	}
	return result[:16]
}

func removeURLSecrets(u *url.URL) *url.URL {
	query := u.Query()
	query.Del("api_key")
	query.Del("application_key")
	u.RawQuery = query.Encode()
	return u
}

func initRecorder(t *testing.T) *recorder.Recorder {
	var mode recorder.Mode
	if isRecording() {
		mode = recorder.ModeRecording
	} else if isReplaying() {
		mode = recorder.ModeReplaying
	} else {
		mode = recorder.ModeDisabled
	}

	rec, err := recorder.NewAsMode(fmt.Sprintf("cassettes/%s", t.Name()), mode, nil)
	if err != nil {
		log.Fatal(err)
	}

	rec.SetMatcher(matchInteraction)

	rec.AddFilter(func(i *cassette.Interaction) error {
		u, err := url.Parse(i.URL)
		if err != nil {
			return err
		}
		i.URL = removeURLSecrets(u).String()
		i.Request.Headers.Del("Dd-Api-Key")
		i.Request.Headers.Del("Dd-Application-Key")
		return nil
	})
	return rec
}

// matchInteraction checks if the request matches a store request in the given cassette.
func matchInteraction(r *http.Request, i cassette.Request) bool {
	// Default matching on method and URL without secrets
	if !(r.Method == i.Method && removeURLSecrets(r.URL).String() == i.URL) {
		log.Printf("HTTP method: %s != %s; URL: %s != %s", r.Method, i.Method, removeURLSecrets(r.URL), i.URL)
		return false
	}

	// Request does not contain body (e.g. `GET`)
	if r.Body == nil {
		log.Printf("request body is empty and cassette body is: %s", i.Body)
		return i.Body == ""
	}

	// Load request body
	var b bytes.Buffer
	if _, err := b.ReadFrom(r.Body); err != nil {
		log.Printf("could not read request body: %v\n", err)
		return false
	}
	r.Body = ioutil.NopCloser(&b)

	matched := b.String() == "" || b.String() == i.Body

	// Ignore boundary differences for multipart/form-data content
	if !matched && strings.HasPrefix(r.Header["Content-Type"][0], "multipart/form-data") {
		rl := strings.Split(strings.TrimSpace(b.String()), "\n")
		cl := strings.Split(strings.TrimSpace(i.Body), "\n")
		if len(rl) > 1 && len(cl) > 1 {
			rs := strings.Join(rl[1:len(rl)-1], "\n")
			cs := strings.Join(cl[1:len(cl)-1], "\n")
			if rs == cs {
				matched = true
			}
		}
	}

	if !matched {
		log.Printf("%s != %s", b.String(), i.Body)
		log.Printf("full cassette info: %v", i)
		log.Printf("full request info: %v", *r)
	}
	return matched
}

func testSpan(ctx context.Context, t *testing.T) context.Context {
	t.Helper()
	tag, err := getEndpointTagValue(t)
	if err != nil {
		t.Fatal(err.Error())
	}
	ctx, finish := ddtesting.StartSpanWithFinish(ctx, t, ddtesting.WithSkipFrames(3), ddtesting.WithSpanOptions(
		// We need to make the tag be something that is then searchable in monitors
		// https://docs.datadoghq.com/tracing/guide/metrics_namespace/#errors
		// "version" is really the only one we can use here
		// NOTE: version is treated in slightly different way, because it's a special tag;
		// if we set it in StartSpanFromContext, it would get overwritten
		tracer.Tag(ext.Version, tag),
	))
	t.Cleanup(finish)
	return ctx
}

func initAccProvider(ctx context.Context, t *testing.T, httpClient *http.Client) *schema.Provider {
	p := datadog.Provider()
	p.ConfigureContextFunc = testProviderConfigure(ctx, httpClient, testClock(t))

	return p
}

func buildContext(ctx context.Context, apiKey string, appKey string, apiURL string) (context.Context, error) {
	ctx = context.WithValue(
		ctx,
		datadogV1.ContextAPIKeys,
		map[string]datadogV1.APIKey{
			"apiKeyAuth": datadogV1.APIKey{
				Key: apiKey,
			},
			"appKeyAuth": datadogV1.APIKey{
				Key: appKey,
			},
		},
	)
	ctx = context.WithValue(
		ctx,
		datadogV2.ContextAPIKeys,
		map[string]datadogV2.APIKey{
			"apiKeyAuth": datadogV2.APIKey{
				Key: apiKey,
			},
			"appKeyAuth": datadogV2.APIKey{
				Key: appKey,
			},
		},
	)

	if apiURL != "" {
		parsedAPIURL, parseErr := url.Parse(apiURL)
		if parseErr != nil {
			return nil, fmt.Errorf(`invalid API Url : %v`, parseErr)
		}
		if parsedAPIURL.Host == "" || parsedAPIURL.Scheme == "" {
			return nil, fmt.Errorf(`missing protocol or host : %v`, apiURL)
		}
		// If api url is passed, set and use the api name and protocol on ServerIndex{1}
		ctx = context.WithValue(ctx, datadogV1.ContextServerIndex, 1)
		ctx = context.WithValue(ctx, datadogV2.ContextServerIndex, 1)

		serverVariables := map[string]string{
			"name":     parsedAPIURL.Host,
			"protocol": parsedAPIURL.Scheme,
		}
		ctx = context.WithValue(ctx, datadogV1.ContextServerVariables, serverVariables)
		ctx = context.WithValue(ctx, datadogV2.ContextServerVariables, serverVariables)
	}
	return ctx, nil
}

func buildDatadogClientV1(httpClient *http.Client) *datadogV1.APIClient {
	//Datadog V1 API config.HTTPClient
	configV1 := datadogV1.NewConfiguration()
	configV1.SetUnstableOperationEnabled("CreateSLOCorrection", true)
	configV1.SetUnstableOperationEnabled("GetSLOCorrection", true)
	configV1.SetUnstableOperationEnabled("UpdateSLOCorrection", true)
	configV1.SetUnstableOperationEnabled("DeleteSLOCorrection", true)
	configV1.Debug = isDebug()
	configV1.HTTPClient = httpClient
	configV1.UserAgent = utils.GetUserAgent(configV1.UserAgent)
	return datadogV1.NewAPIClient(configV1)
}

func buildDatadogClientV2(httpClient *http.Client) *datadogV2.APIClient {
	//Datadog V2 API config.HTTPClient
	configV2 := datadogV2.NewConfiguration()
	configV2.SetUnstableOperationEnabled("CreateTagConfiguration", true)
	configV2.SetUnstableOperationEnabled("DeleteTagConfiguration", true)
	configV2.SetUnstableOperationEnabled("ListTagConfigurationByName", true)
	configV2.SetUnstableOperationEnabled("UpdateTagConfiguration", true)
	configV2.Debug = isDebug()
	configV2.HTTPClient = httpClient
	configV2.UserAgent = utils.GetUserAgent(configV2.UserAgent)
	return datadogV2.NewAPIClient(configV2)
}

func testProviderConfigure(ctx context.Context, httpClient *http.Client, clock clockwork.FakeClock) schema.ConfigureContextFunc {
	return func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		communityClient := datadogCommunity.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))
		if apiURL := d.Get("api_url").(string); apiURL != "" {
			communityClient.SetBaseUrl(apiURL)
		}

		c := ddhttp.WrapClient(httpClient)

		communityClient.HttpClient = c
		communityClient.ExtraHeader["User-Agent"] = utils.GetUserAgent(fmt.Sprintf(
			"datadog-api-client-go/%s (go %s; os %s; arch %s)",
			"go-datadog-api",
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		))

		ctx, err := buildContext(ctx, d.Get("api_key").(string), d.Get("app_key").(string), d.Get("api_url").(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return &datadog.ProviderConfiguration{
			CommunityClient: communityClient,
			DatadogClientV1: buildDatadogClientV1(c),
			DatadogClientV2: buildDatadogClientV2(c),
			AuthV1:          ctx,
			AuthV2:          ctx,

			Now: clock.Now,
		}, nil
	}
}

func testAccProvidersWithHTTPClient(ctx context.Context, t *testing.T, httpClient *http.Client) map[string]func() (*schema.Provider, error) {
	provider := initAccProvider(ctx, t, httpClient)
	return map[string]func() (*schema.Provider, error){
		"datadog": func() (*schema.Provider, error) {
			return provider, nil
		},
	}
}

func testAccProviders(ctx context.Context, t *testing.T) (context.Context, map[string]func() (*schema.Provider, error)) {
	ctx = testSpan(ctx, t)
	rec := initRecorder(t)
	ctx = context.WithValue(ctx, clockContextKey("clock"), testClock(t))
	c := cleanhttp.DefaultClient()
	loggingTransport := logging.NewTransport("Datadog", rec)
	c.Transport = transport.NewCustomTransport(loggingTransport, transport.CustomTransportOptions{})
	p := testAccProvidersWithHTTPClient(ctx, t, c)
	t.Cleanup(func() {
		rec.Stop()
	})

	return ctx, p
}

func testAccProvider(t *testing.T, accProviders map[string]func() (*schema.Provider, error)) func() (*schema.Provider, error) {
	accProvider, ok := accProviders["datadog"]
	if !ok {
		t.Fatal("could not find datadog provider")
	}
	return accProvider
}

func TestProvider(t *testing.T) {
	rec := initRecorder(t)
	defer rec.Stop()

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", rec)
	accProvider := initAccProvider(context.Background(), t, c)

	if err := accProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = datadog.Provider()
}

func testAccPreCheck(t *testing.T) {
	// Unset all regular env to avoid mistakenly running tests against wrong org
	for _, v := range append(datadog.APPKeyEnvVars, datadog.APIKeyEnvVars...) {
		_ = os.Unsetenv(v)
	}

	if isReplaying() {
		return
	}

	if !isAPIKeySet() {
		t.Fatalf("%s must be set for acceptance tests", testAPIKeyEnvName)
	}
	if !isAPPKeySet() {
		t.Fatalf("%s must be set for acceptance tests", testAPPKeyEnvName)
	}

	if !isTestOrg() {
		t.Fatalf(
			"The keys you've set potentially belong to a production environment. "+
				"Tests do all sorts of create/update/delete calls to the organisation, so only run them against a sandbox environment. "+
				"If you know what you are doing, set the `%s` environment variable to the public ID of your organization. "+
				"See https://docs.datadoghq.com/api/latest/organizations/#list-your-managed-organizations to get it.",
			testOrgEnvName,
		)
	}

	if err := os.Setenv(datadog.DDAPIKeyEnvName, os.Getenv(testAPIKeyEnvName)); err != nil {
		t.Fatalf("Error setting API key: %v", err)
	}

	if err := os.Setenv(datadog.DDAPPKeyEnvName, os.Getenv(testAPPKeyEnvName)); err != nil {
		t.Fatalf("Error setting API key: %v", err)
	}
}

func testCheckResourceAttrs(name string, checkExists resource.TestCheckFunc, assertions []string) []resource.TestCheckFunc {
	typeSet := "TypeSet"
	funcs := []resource.TestCheckFunc{}
	funcs = append(funcs, checkExists)
	for _, assertion := range assertions {
		assertionPair := strings.Split(assertion, " = ")
		if len(assertionPair) == 1 {
			assertionPair = strings.Split(assertion, " =")
		}
		key := assertionPair[0]
		value := ""
		if len(assertionPair) > 1 {
			value = assertionPair[1]
		}

		// Handle TypeSet attributes
		if strings.Contains(key, typeSet) {
			key = strings.Replace(key, typeSet, "*", 1)
			funcs = append(funcs, resource.TestCheckTypeSetElemAttr(name, key, value))
		} else {
			funcs = append(funcs, resource.TestCheckResourceAttr(name, key, value))
			// Use utility method below, instead of the above one, to print out all state keys/values during test debugging
			//funcs = append(funcs, CheckResourceAttr(name, key, value))
		}
	}
	return funcs
}

/* Utility method for Debugging purpose. This method helps list assertions as well
It is a duplication of `resource.TestCheckResourceAttr` into which we added print statements.
*/
func CheckResourceAttr(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return nil
		}

		is := rs.Primary
		if is == nil {
			return nil
		}

		for k, val := range is.Attributes {
			fmt.Printf("%v = %v\n", k, val)
		}

		// Empty containers may be elided from the state.
		// If the intent here is to check for an empty container, allow the key to
		// also be non-existent.
		emptyCheck := value == "0" && (strings.HasSuffix(key, ".#") || strings.HasSuffix(key, ".%"))

		if v, ok := is.Attributes[key]; !ok || v != value {

			if emptyCheck && !ok {
				return nil
			}

			if !ok {
				return fmt.Errorf("%s: Attribute '%s' not found", name, key)
			}

			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v",
				name,
				key,
				value,
				v)
		}
		return nil
	}
}
