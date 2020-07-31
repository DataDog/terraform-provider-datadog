package datadog

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/jonboulle/clockwork"
	datadogCommunity "github.com/zorkian/go-datadog-api"
	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var testFiles2EndpointTags = map[string]string{
	"data_source_datadog_ip_ranges_test":                         "ip-ranges",
	"data_source_datadog_monitor_test":                           "monitors",
	"data_source_datadog_synthetics_locations_test":              "synthetics",
	"import_datadog_downtime_test":                               "downtimes",
	"import_datadog_integration_pagerduty_test":                  "integration-pagerduty",
	"import_datadog_logs_pipeline_test":                          "logs-pipelines",
	"import_datadog_monitor_test":                                "monitors",
	"import_datadog_user_test":                                   "users",
	"provider_test":                                              "terraform",
	"resource_datadog_dashboard_alert_graph_test":                "dashboards",
	"resource_datadog_dashboard_alert_value_test":                "dashboards",
	"resource_datadog_dashboard_change_test":                     "dashboards",
	"resource_datadog_dashboard_check_status_test":               "dashboards",
	"resource_datadog_dashboard_distribution_test":               "dashboards",
	"resource_datadog_dashboard_event_stream_test":               "dashboards",
	"resource_datadog_dashboard_event_timeline_test":             "dashboards",
	"resource_datadog_dashboard_free_text_test":                  "dashboards",
	"resource_datadog_dashboard_heatmap_test":                    "dashboards",
	"resource_datadog_dashboard_hostmap_test":                    "dashboards",
	"resource_datadog_dashboard_iframe_test":                     "dashboards",
	"resource_datadog_dashboard_image_test":                      "dashboards",
	"resource_datadog_dashboard_list_test":                       "dashboard-lists",
	"resource_datadog_dashboard_log_stream_test":                 "dashboards",
	"resource_datadog_dashboard_manage_status_test":              "dashboards",
	"resource_datadog_dashboard_note_test":                       "dashboards",
	"resource_datadog_dashboard_query_table_test":                "dashboards",
	"resource_datadog_dashboard_query_value_test":                "dashboards",
	"resource_datadog_dashboard_scatterplot_test":                "dashboards",
	"resource_datadog_dashboard_service_map_test":                "dashboards",
	"resource_datadog_dashboard_slo_test":                        "dashboards",
	"resource_datadog_dashboard_test":                            "dashboards",
	"resource_datadog_dashboard_timeseries_test":                 "dashboards",
	"resource_datadog_dashboard_top_list_test":                   "dashboards",
	"resource_datadog_dashboard_trace_service_test":              "dashboards",
	"resource_datadog_downtime_test":                             "downtimes",
	"resource_datadog_integration_aws_lambda_arn_test":           "integration-aws",
	"resource_datadog_integration_aws_log_collection_test":       "integration-aws",
	"resource_datadog_integration_aws_test":                      "integration-aws",
	"resource_datadog_integration_azure_test":                    "integration-azure",
	"resource_datadog_integration_gcp_test":                      "integration-gcp",
	"resource_datadog_integration_pagerduty_service_object_test": "integration-pagerduty",
	"resource_datadog_integration_pagerduty_test":                "integration-pagerduty",
	"resource_datadog_logs_archive_test":                         "logs-archive",
	"resource_datadog_logs_custom_pipeline_test":                 "logs-pipelines",
	"resource_datadog_metric_metadata_test":                      "metrics",
	"resource_datadog_monitor_test":                              "monitors",
	"resource_datadog_screenboard_test":                          "dashboards",
	"resource_datadog_service_level_objective_test":              "service-level-objectives",
	"resource_datadog_synthetics_test_test":                      "synthetics",
	"resource_datadog_timeboard_test":                            "dashboards",
	"resource_datadog_user_test":                                 "users",
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
	for more {
		frame, more = frames.Next()
		// nested test functions like `TestAuthenticationValidate/200_Valid` will have frame.Function ending with
		// ".funcX", `e.g. datadog.TestAuthenticationValidate.func1`, so trim everything after last "/" in test name
		// and everything after last "." in frame function name
		frameFunction := frame.Function
		testName := t.Name()
		if strings.Contains(testName, "/") {
			testName = testName[:strings.LastIndex(testName, "/")]
			frameFunction = frameFunction[:strings.LastIndex(frameFunction, ".")]
		}
		if strings.HasSuffix(frameFunction, "."+testName) {
			functionFile = frame.File
			// when we find the frame with the current test function, match it against testFiles2EndpointTags
			for file, tag := range testFiles2EndpointTags {
				if strings.HasSuffix(frame.File, fmt.Sprintf("datadog/%s.go", file)) {
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
	if os.Getenv("DATADOG_API_KEY") != "" {
		return true
	}
	if os.Getenv("DD_API_KEY") != "" {
		return true
	}
	return false
}

func isAPPKeySet() bool {
	if os.Getenv("DATADOG_APP_KEY") != "" {
		return true
	}
	if os.Getenv("DD_APP_KEY") != "" {
		return true
	}
	return false
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

	rec.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		return r.Method == i.Method && removeURLSecrets(r.URL).String() == i.URL
	})

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

func testSpan(ctx context.Context, t *testing.T) (context.Context, func()) {
	t.Helper()
	tag, err := getEndpointTagValue(t)
	if err != nil {
		t.Fatal(err.Error())
	}
	span, ctx := tracer.StartSpanFromContext(
		ctx,
		"test",
		tracer.SpanType("test"),
		tracer.ResourceName(t.Name()),
		tracer.Tag(ext.AnalyticsEvent, true),
		tracer.Measured(),
	)
	span.SetTag("version", tag)
	return ctx, func() {
		span.SetTag(ext.Error, t.Failed())
		span.Finish()
	}
}

func initAccProvider(ctx context.Context, t *testing.T, httpClient *http.Client) *schema.Provider {
	ctx, finish := testSpan(context.Background(), t)
	defer finish()

	p := Provider().(*schema.Provider)
	p.ConfigureFunc = testProviderConfigure(ctx, httpClient)

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
	configV1.Debug = isDebug()
	configV1.HTTPClient = httpClient
	configV1.UserAgent = getUserAgent(configV1.UserAgent)
	return datadogV1.NewAPIClient(configV1)
}

func buildDatadogClientV2(httpClient *http.Client) *datadogV2.APIClient {
	//Datadog V2 API config.HTTPClient
	configV2 := datadogV2.NewConfiguration()
	configV2.Debug = isDebug()
	configV2.HTTPClient = httpClient
	configV2.UserAgent = getUserAgent(configV2.UserAgent)
	return datadogV2.NewAPIClient(configV2)
}

func testProviderConfigure(ctx context.Context, httpClient *http.Client) schema.ConfigureFunc {
	return func(d *schema.ResourceData) (interface{}, error) {
		communityClient := datadogCommunity.NewClient(d.Get("api_key").(string), d.Get("app_key").(string))
		if apiURL := d.Get("api_url").(string); apiURL != "" {
			communityClient.SetBaseUrl(apiURL)
		}

		c := ddhttp.WrapClient(httpClient)

		communityClient.HttpClient = c
		communityClient.ExtraHeader["User-Agent"] = getUserAgent(fmt.Sprintf(
			"datadog-api-client-go/%s (go %s; os %s; arch %s)",
			"go-datadog-api",
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		))

		ctx, err := buildContext(ctx, d.Get("api_key").(string), d.Get("app_key").(string), d.Get("api_url").(string))
		if err != nil {
			return nil, err
		}

		return &ProviderConfiguration{
			CommunityClient: communityClient,
			DatadogClientV1: buildDatadogClientV1(c),
			DatadogClientV2: buildDatadogClientV2(c),
			AuthV1:          ctx,
			AuthV2:          ctx,
		}, nil
	}
}

func testAccProvidersWithHTTPClient(t *testing.T, httpClient *http.Client) (map[string]terraform.ResourceProvider, func()) {
	ctx, finish := testSpan(context.Background(), t)

	provider := initAccProvider(ctx, t, httpClient)
	return map[string]terraform.ResourceProvider{
		"datadog": provider,
	}, finish
}

func testAccProviders(t *testing.T, rec *recorder.Recorder) (map[string]terraform.ResourceProvider, clockwork.FakeClock, func(t *testing.T)) {
	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", rec)
	p, finish := testAccProvidersWithHTTPClient(t, c)
	return p, testClock(t), func(t *testing.T) {
		rec.Stop()
		finish()
	}
}

func testAccProvider(t *testing.T, accProviders map[string]terraform.ResourceProvider) *schema.Provider {
	accProvider, ok := accProviders["datadog"]
	if !ok {
		t.Fatal("could not find datadog provider")
	}
	return accProvider.(*schema.Provider)
}

func TestProvider(t *testing.T) {
	ctx, finish := testSpan(context.Background(), t)
	defer finish()

	rec := initRecorder(t)
	defer rec.Stop()

	c := cleanhttp.DefaultClient()
	c.Transport = logging.NewTransport("Datadog", rec)
	accProvider := initAccProvider(ctx, t, c)

	if err := accProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if isReplaying() {
		return
	}
	if !isAPIKeySet() {
		t.Fatal("DD_API_KEY must be set for acceptance tests")
	}
	if !isAPPKeySet() {
		t.Fatal("DD_APP_KEY must be set for acceptance tests")
	}
}

func testCheckResourceAttrs(name string, checkExists resource.TestCheckFunc, assertions []string) []resource.TestCheckFunc {
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
		funcs = append(funcs, resource.TestCheckResourceAttr(name, key, value))
		// Use utility method below, instead of the above one, to print out all state keys/values during test debugging
		//funcs = append(funcs, CheckResourceAttr(name, key, value))
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
			fmt.Println(fmt.Sprintf("%v = %v", k, val))
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
