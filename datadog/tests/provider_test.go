package test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"math/rand"
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
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	ddtesting "github.com/DataDog/dd-sdk-go-testing"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jonboulle/clockwork"
	datadogCommunity "github.com/zorkian/go-datadog-api"
	ddhttp "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

type clockContextKey string

const (
	ddTestOrg         = "fasjyydbcgwwc2uc"
	testAPIKeyEnvName = "DD_TEST_CLIENT_API_KEY"
	testAPPKeyEnvName = "DD_TEST_CLIENT_APP_KEY"
	testAPIUrlEnvName = "DD_TEST_SITE_URL"
	testOrgEnvName    = "DD_TEST_ORG"
)

var isTestOrgC *bool

var allowedHeaders = map[string]string{"Accept": "", "Content-Type": ""}

var testFiles2EndpointTags = map[string]string{
	"tests/data_source_datadog_api_key_test":                                 "api_keys",
	"tests/data_source_datadog_apm_retention_filters_order_test":             "apm_retention_filters_order",
	"tests/data_source_datadog_application_key_test":                         "application_keys",
	"tests/data_source_datadog_cloud_workload_security_agent_rules_test":     "cloud-workload-security",
	"tests/data_source_datadog_csm_threats_agent_rules_test":                 "cloud-workload-security",
	"tests/data_source_datadog_dashboard_list_test":                          "dashboard-lists",
	"tests/data_source_datadog_dashboard_test":                               "dashboard",
	"tests/data_source_datadog_hosts_test":                                   "hosts",
	"tests/data_source_datadog_integration_aws_logs_services_test":           "integration-aws",
	"tests/data_source_datadog_integration_aws_namespace_rules_test":         "integration-aws",
	"tests/data_source_datadog_ip_ranges_test":                               "ip-ranges",
	"tests/data_source_datadog_logs_archives_order_test":                     "logs-archive",
	"tests/data_source_datadog_logs_indexes_order_test":                      "logs-index",
	"tests/data_source_datadog_logs_indexes_test":                            "logs-index",
	"tests/data_source_datadog_logs_pipelines_test":                          "logs-pipelines",
	"tests/data_source_datadog_monitor_config_policies_test":                 "monitor-config-policies",
	"tests/data_source_datadog_monitor_config_policy_test":                   "monitor-config-policies",
	"tests/data_source_datadog_monitor_test":                                 "monitors",
	"tests/data_source_datadog_monitors_test":                                "monitors",
	"tests/data_source_datadog_permissions_test":                             "permissions",
	"tests/data_source_datadog_powerpack_test":                               "powerpacks",
	"tests/data_source_datadog_restriction_policy_test":                      "restriction-policy",
	"tests/data_source_datadog_role_test":                                    "roles",
	"tests/data_source_datadog_role_users_test":                              "roles",
	"tests/data_source_datadog_roles_test":                                   "roles",
	"tests/data_source_datadog_rum_application_test":                         "rum-application",
	"tests/data_source_datadog_security_monitoring_filters_test":             "security-monitoring",
	"tests/data_source_datadog_security_monitoring_rules_test":               "security-monitoring",
	"tests/data_source_datadog_security_monitoring_suppressions_test":        "security-monitoring",
	"tests/data_source_datadog_sensitive_data_scanner_group_order_test":      "sensitive-data-scanner",
	"tests/data_source_datadog_sensitive_data_scanner_standard_pattern_test": "sensitive-data-scanner",
	"tests/data_source_datadog_service_account_test":                         "users",
	"tests/data_source_datadog_service_level_objective_test":                 "service-level-objectives",
	"tests/data_source_datadog_service_level_objectives_test":                "service-level-objectives",
	"tests/data_source_datadog_synthetics_global_variable_test":              "synthetics",
	"tests/data_source_datadog_synthetics_locations_test":                    "synthetics",
	"tests/data_source_datadog_synthetics_test_test":                         "synthetics",
	"tests/data_source_datadog_team_memberships_test":                        "team",
	"tests/data_source_datadog_team_test":                                    "team",
	"tests/data_source_datadog_user_test":                                    "users",
	"tests/data_source_datadog_users_test":                                   "users",
	"tests/import_datadog_downtime_test":                                     "downtimes",
	"tests/import_datadog_integration_pagerduty_test":                        "integration-pagerduty",
	"tests/import_datadog_logs_pipeline_test":                                "logs-pipelines",
	"tests/import_datadog_monitor_test":                                      "monitors",
	"tests/import_datadog_user_test":                                         "users",
	"tests/provider_test":                                                    "terraform",
	"tests/resource_datadog_api_key_test":                                    "api_keys",
	"tests/resource_datadog_apm_retention_filter_test":                       "apm_retention_filter",
	"tests/resource_datadog_apm_retention_filter_order_test":                 "apm_retention_filter_order",
	"tests/resource_datadog_application_key_test":                            "application_keys",
	"tests/resource_datadog_authn_mapping_test":                              "authn_mapping",
	"tests/resource_datadog_child_organization_test":                         "organization",
	"tests/resource_datadog_cloud_configuration_rule_test":                   "security-monitoring",
	"tests/resource_datadog_cloud_workload_security_agent_rule_test":         "cloud_workload_security",
	"tests/resource_datadog_csm_threats_agent_rule_test":                     "cloud-workload-security",
	"tests/resource_datadog_dashboard_alert_graph_test":                      "dashboards",
	"tests/resource_datadog_dashboard_alert_value_test":                      "dashboards",
	"tests/resource_datadog_dashboard_change_test":                           "dashboards",
	"tests/resource_datadog_dashboard_check_status_test":                     "dashboards",
	"tests/resource_datadog_dashboard_cross_org_test":                        "dashboards",
	"tests/resource_datadog_dashboard_distribution_test":                     "dashboards",
	"tests/resource_datadog_dashboard_event_stream_test":                     "dashboards",
	"tests/resource_datadog_dashboard_event_timeline_test":                   "dashboards",
	"tests/resource_datadog_dashboard_free_text_test":                        "dashboards",
	"tests/resource_datadog_dashboard_geomap_test":                           "dashboards",
	"tests/resource_datadog_dashboard_heatmap_test":                          "dashboards",
	"tests/resource_datadog_dashboard_hostmap_test":                          "dashboards",
	"tests/resource_datadog_dashboard_iframe_test":                           "dashboards",
	"tests/resource_datadog_dashboard_image_test":                            "dashboards",
	"tests/resource_datadog_dashboard_json_test":                             "dashboards-json",
	"tests/resource_datadog_dashboard_list_stream_storage_test":              "dashboards",
	"tests/resource_datadog_dashboard_list_stream_test":                      "dashboards",
	"tests/resource_datadog_dashboard_list_test":                             "dashboard-lists",
	"tests/resource_datadog_dashboard_log_stream_test":                       "dashboards",
	"tests/resource_datadog_dashboard_manage_status_test":                    "dashboards",
	"tests/resource_datadog_dashboard_note_test":                             "dashboards",
	"tests/resource_datadog_dashboard_powerpack_test":                        "dashboards",
	"tests/resource_datadog_dashboard_query_table_test":                      "dashboards",
	"tests/resource_datadog_dashboard_query_value_test":                      "dashboards",
	"tests/resource_datadog_dashboard_run_workflow_test":                     "dashboards",
	"tests/resource_datadog_dashboard_scatterplot_test":                      "dashboards",
	"tests/resource_datadog_dashboard_service_map_test":                      "dashboards",
	"tests/resource_datadog_dashboard_slo_list_test":                         "dashboards",
	"tests/resource_datadog_dashboard_slo_test":                              "dashboards",
	"tests/resource_datadog_dashboard_style_test":                            "dashboards",
	"tests/resource_datadog_dashboard_split_graph_test":                      "dashboards",
	"tests/resource_datadog_dashboard_sunburst_test":                         "dashboards",
	"tests/resource_datadog_dashboard_test":                                  "dashboards",
	"tests/resource_datadog_dashboard_timeseries_test":                       "dashboards",
	"tests/resource_datadog_dashboard_top_list_test":                         "dashboards",
	"tests/resource_datadog_dashboard_topology_map_test":                     "dashboards",
	"tests/resource_datadog_dashboard_trace_service_test":                    "dashboards",
	"tests/resource_datadog_dashboard_treemap_test":                          "dashboards",
	"tests/resource_datadog_openapi_api_test":                                "apimanagement",
	"tests/resource_datadog_powerpack_test":                                  "powerpacks",
	"tests/resource_datadog_powerpack_alert_graph_test":                      "powerpacks",
	"tests/resource_datadog_powerpack_alert_value_test":                      "powerpacks",
	"tests/resource_datadog_powerpack_change_test":                           "powerpacks",
	"tests/resource_datadog_powerpack_check_status_test":                     "powerpacks",
	"tests/resource_datadog_powerpack_distribution_test":                     "powerpacks",
	"tests/resource_datadog_powerpack_event_stream_test":                     "powerpacks",
	"tests/resource_datadog_powerpack_event_timeline_test":                   "powerpacks",
	"tests/resource_datadog_powerpack_geomap_test":                           "powerpacks",
	"tests/resource_datadog_powerpack_iframe_test":                           "powerpacks",
	"tests/resource_datadog_powerpack_image_test":                            "powerpacks",
	"tests/resource_datadog_powerpack_free_text_test":                        "powerpacks",
	"tests/resource_datadog_powerpack_heatmap_test":                          "powerpacks",
	"tests/resource_datadog_powerpack_hostmap_test":                          "powerpacks",
	"tests/resource_datadog_powerpack_list_stream_test":                      "powerpacks",
	"tests/resource_datadog_powerpack_log_stream_test":                       "powerpacks",
	"tests/resource_datadog_powerpack_manage_status_test":                    "powerpacks",
	"tests/resource_datadog_powerpack_note_test":                             "powerpacks",
	"tests/resource_datadog_powerpack_query_table_test":                      "powerpacks",
	"tests/resource_datadog_powerpack_query_value_test":                      "powerpacks",
	"tests/resource_datadog_powerpack_run_workflow_test":                     "powerpacks",
	"tests/resource_datadog_powerpack_scatterplot_test":                      "powerpacks",
	"tests/resource_datadog_powerpack_servicemap_test":                       "powerpacks",
	"tests/resource_datadog_powerpack_slo_test":                              "powerpacks",
	"tests/resource_datadog_powerpack_slo_list_test":                         "powerpacks",
	"tests/resource_datadog_powerpack_sunburst_test":                         "powerpacks",
	"tests/resource_datadog_powerpack_timeseries_test":                       "powerpacks",
	"tests/resource_datadog_powerpack_toplist_test":                          "powerpacks",
	"tests/resource_datadog_powerpack_topology_map_test":                     "powerpacks",
	"tests/resource_datadog_powerpack_trace_service_test":                    "powerpacks",
	"tests/resource_datadog_powerpack_treemap_test":                          "powerpacks",
	"tests/resource_datadog_downtime_test":                                   "downtimes",
	"tests/resource_datadog_downtime_schedule_test":                          "downtimes",
	"tests/resource_datadog_integration_aws_lambda_arn_test":                 "integration-aws",
	"tests/resource_datadog_integration_aws_log_collection_test":             "integration-aws",
	"tests/resource_datadog_integration_aws_tag_filter_test":                 "integration-aws",
	"tests/resource_datadog_integration_aws_test":                            "integration-aws",
	"tests/resource_datadog_aws_account_v2_test":                             "integration-aws",
	"tests/resource_datadog_integration_aws_event_bridge_test":               "integration-aws",
	"tests/resource_datadog_integration_azure_test":                          "integration-azure",
	"tests/resource_datadog_integration_cloudflare_account_test":             "integration-cloudflare",
	"tests/resource_datadog_integration_confluent_account_test":              "integration-confluend-account",
	"tests/resource_datadog_integration_confluent_resource_test":             "integration-confluend-resource",
	"tests/resource_datadog_integration_fastly_account_test":                 "integration-fastly-account",
	"tests/resource_datadog_integration_gcp_sts_test":                        "integration-gcp",
	"tests/resource_datadog_integration_gcp_test":                            "integration-gcp",
	"tests/resource_datadog_integration_opsgenie_service_object_test":        "integration-opsgenie-service",
	"tests/resource_datadog_integration_pagerduty_service_object_test":       "integration-pagerduty",
	"tests/resource_datadog_integration_pagerduty_test":                      "integration-pagerduty",
	"tests/resource_datadog_integration_slack_channel_test":                  "integration-slack-channel",
	"tests/resource_datadog_ip_allowlist_test":                               "ip_allowlist",
	"tests/resource_datadog_logs_archive_order_test":                         "logs-archive-order",
	"tests/resource_datadog_logs_archive_test":                               "logs-archive",
	"tests/resource_datadog_logs_custom_destination_test":                    "logs-custom-destination",
	"tests/resource_datadog_logs_custom_pipeline_test":                       "logs-pipelines",
	"tests/resource_datadog_logs_index_test":                                 "logs-index",
	"tests/resource_datadog_logs_metric_test":                                "logs-metric",
	"tests/resource_datadog_metric_metadata_test":                            "metrics",
	"tests/resource_datadog_metric_tag_configuration_test":                   "metrics",
	"tests/resource_datadog_monitor_config_policy_test":                      "monitor-config-policies",
	"tests/resource_datadog_monitor_json_test":                               "monitors-json",
	"tests/resource_datadog_monitor_test":                                    "monitors",
	"tests/resource_datadog_organization_settings_test":                      "organization",
	"tests/resource_datadog_restriction_policy_test":                         "restriction-policy",
	"tests/resource_datadog_role_test":                                       "roles",
	"tests/resource_datadog_rum_application_test":                            "rum-application",
	"tests/resource_datadog_screenboard_test":                                "dashboards",
	"tests/resource_datadog_security_monitoring_default_rule_test":           "security-monitoring",
	"tests/resource_datadog_security_monitoring_filter_test":                 "security-monitoring",
	"tests/resource_datadog_security_monitoring_rule_test":                   "security-monitoring",
	"tests/resource_datadog_security_monitoring_suppression_test":            "security-monitoring",
	"tests/resource_datadog_sensitive_data_scanner_group_order_test":         "sensitive-data-scanner",
	"tests/resource_datadog_sensitive_data_scanner_group_test":               "sensitive-data-scanner",
	"tests/resource_datadog_sensitive_data_scanner_rule_test":                "sensitive-data-scanner",
	"tests/resource_datadog_service_account_application_key_test":            "users",
	"tests/resource_datadog_service_account_test":                            "users",
	"tests/resource_datadog_service_definition_yaml_test":                    "service-definition",
	"tests/resource_datadog_service_level_objective_test":                    "service-level-objectives",
	"tests/resource_datadog_slo_correction_test":                             "slo_correction",
	"tests/resource_datadog_software_catalog_test":                           "software-catalog",
	"tests/resource_datadog_spans_metric_test":                               "spans-metric",
	"tests/resource_datadog_synthetics_concurrency_cap_test":                 "synthetics",
	"tests/resource_datadog_synthetics_global_variable_test":                 "synthetics",
	"tests/resource_datadog_synthetics_private_location_test":                "synthetics",
	"tests/resource_datadog_synthetics_test_test":                            "synthetics",
	"tests/resource_datadog_team_link_test":                                  "team",
	"tests/resource_datadog_team_membership_test":                            "team",
	"tests/resource_datadog_team_permission_setting_test":                    "team",
	"tests/resource_datadog_team_test":                                       "team",
	"tests/resource_datadog_timeboard_test":                                  "dashboards",
	"tests/resource_datadog_user_test":                                       "users",
	"tests/resource_datadog_user_role_test":                                  "roles",
	"tests/resource_datadog_webhook_custom_variable_test":                    "webhook_custom_variable",
	"tests/resource_datadog_webhook_test":                                    "webhook",
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

	var apiURL string
	if apiURL = os.Getenv(testAPIUrlEnvName); apiURL == "" {
		apiURL = "https://api.datadoghq.com"
	}

	// If keys belong to test org, then this get will succeed, otherwise it will fail with 400
	publicID := ddTestOrg
	if v := os.Getenv(testOrgEnvName); v != "" {
		publicID = v
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s", strings.TrimRight(apiURL, "/"), "api/v1/org", publicID), nil)
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
	data, err := os.ReadFile(fmt.Sprintf("cassettes/%s.freeze", t.Name()))
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

// uniqueAgentRuleName takes the current/frozen time and uses it to generate a unique agent
// rule name that changes in CI, but is stable locally.
func uniqueAgentRuleName(ctx context.Context) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(clockFromContext(ctx).Now().Unix()))
	charset := "abcdefghijklmnopqrstuvwxyz"
	nameLength := 10
	var buf bytes.Buffer
	buf.Grow(nameLength)
	for i := 0; i < nameLength; i++ {
		buf.WriteString(string(charset[seededRand.Intn(len(charset))]))
	}
	return buf.String()
}

// uniqueAWSAccessKeyID takes uniqueEntityName result, hashes it to get a unique string
// and then returns first 16 characters (numerical only), so that the value can be used
// as AWS account ID and is still as unique as possible, it changes in CI, but is stable locally
func uniqueAWSAccessKeyID(ctx context.Context, t *testing.T) string {
	uniq := uniqueEntityName(ctx, t)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(uniq)))
	result := "AKIA"
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
		mode = recorder.ModeRecordOnly
	} else if isReplaying() {
		mode = recorder.ModeReplayOnly
	} else {
		mode = recorder.ModePassthrough
	}

	opts := &recorder.Options{
		CassetteName:       fmt.Sprintf("cassettes/%s", t.Name()),
		Mode:               mode,
		SkipRequestLatency: true,
		RealTransport:      http.DefaultTransport,
	}

	rec, err := recorder.NewWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}

	rec.SetMatcher(matchInteraction)

	redactHook := func(i *cassette.Interaction) error {
		u, err := url.Parse(i.Request.URL)
		if err != nil {
			return err
		}
		i.Request.URL = removeURLSecrets(u).String()

		filterHeaders(i)
		return nil
	}
	rec.AddHook(redactHook, recorder.AfterCaptureHook)

	return rec
}

// filterHeaders filter out headers
func filterHeaders(i *cassette.Interaction) {
	requestHeadersCopy := i.Request.Headers.Clone()
	responseHeadersCopy := i.Response.Headers.Clone()

	for k := range requestHeadersCopy {
		if _, ok := allowedHeaders[k]; !ok {
			i.Request.Headers.Del(k)
		}
	}
	for k := range responseHeadersCopy {
		if _, ok := allowedHeaders[k]; !ok {
			i.Response.Headers.Del(k)
		}
	}
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
	r.Body = io.NopCloser(&b)

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

	ctx, finish := ddtesting.StartTestWithContext(ctx, t, ddtesting.WithSkipFrames(3), ddtesting.WithSpanOptions(
		// Set resource name to TestName
		tracer.ResourceName(t.Name()),

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
		common.ContextAPIKeys,
		map[string]common.APIKey{
			"apiKeyAuth": {
				Key: apiKey,
			},
			"appKeyAuth": {
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
		ctx = context.WithValue(ctx, common.ContextServerIndex, 1)

		serverVariables := map[string]string{
			"name":     parsedAPIURL.Host,
			"protocol": parsedAPIURL.Scheme,
		}
		ctx = context.WithValue(ctx, common.ContextServerVariables, serverVariables)
	}
	return ctx, nil
}

func buildDatadogClient(ctx context.Context, httpClient *http.Client) *common.APIClient {
	// Datadog API config.HTTPClient
	config := common.NewConfiguration()
	if ctx.Value("http_retry_enable") == true {
		config.RetryConfiguration.EnableRetry = true
	}
	config.Debug = isDebug()
	config.HTTPClient = httpClient
	config.UserAgent = utils.GetUserAgent(config.UserAgent)
	return common.NewAPIClient(config)
}

func testProviderConfigure(ctx context.Context, httpClient *http.Client, clock clockwork.FakeClock) schema.ConfigureContextFunc {
	return func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		apiKey := d.Get("api_key").(string)
		if apiKey == "" {
			apiKey, _ = utils.GetMultiEnvVar(utils.APIKeyEnvVars[:]...)
		}

		appKey := d.Get("app_key").(string)
		if appKey == "" {
			appKey, _ = utils.GetMultiEnvVar(utils.APPKeyEnvVars[:]...)
		}

		apiURL := d.Get("api_url").(string)
		if apiURL == "" {
			apiURL, _ = utils.GetMultiEnvVar(utils.APIUrlEnvVars[:]...)
		}

		communityClient := datadogCommunity.NewClient(apiKey, appKey)
		if apiURL != "" {
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

		ctx, err := buildContext(ctx, apiKey, appKey, apiURL)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return &datadog.ProviderConfiguration{
			CommunityClient:     communityClient,
			DatadogApiInstances: &utils.ApiInstances{HttpClient: buildDatadogClient(ctx, c)},
			Auth:                ctx,

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
	c.Transport = rec
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

func withDefaultTags(providerFactory func() (*schema.Provider, error), defaultTags map[string]interface{}) func() (*schema.Provider, error) {
	provider, err := providerFactory()
	newProvider := *provider
	return func() (*schema.Provider, error) {
		configureFunc := func(lctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			config, diags := provider.ConfigureContextFunc(lctx, d)
			if config != nil {
				config.(*datadog.ProviderConfiguration).DefaultTags = defaultTags
			}
			return config, diags
		}
		newProvider.ConfigureContextFunc = configureFunc
		return &newProvider, err
	}
}

func TestProvider(t *testing.T) {
	rec := initRecorder(t)
	defer rec.Stop()

	c := cleanhttp.DefaultClient()
	c.Transport = rec
	accProvider := initAccProvider(context.Background(), t, c)

	if err := accProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	_ = datadog.Provider()
}

func testAccPreCheck(t *testing.T) {
	// Unset all regular env to avoid mistakenly running tests against wrong org
	var envVars []string
	envVars = append(envVars, utils.APPKeyEnvVars...)
	envVars = append(envVars, utils.APIKeyEnvVars...)
	envVars = append(envVars, utils.APIUrlEnvVars...)

	for _, v := range envVars {
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

	if err := os.Setenv(utils.DDAPIKeyEnvName, os.Getenv(testAPIKeyEnvName)); err != nil {
		t.Fatalf("Error setting API key: %v", err)
	}

	if err := os.Setenv(utils.DDAPPKeyEnvName, os.Getenv(testAPPKeyEnvName)); err != nil {
		t.Fatalf("Error setting API key: %v", err)
	}
	if err := os.Setenv(utils.DDAPIUrlEnvName, os.Getenv(testAPIUrlEnvName)); err != nil {
		t.Fatalf("Error setting API url: %v", err)
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
			// funcs = append(funcs, CheckResourceAttr(name, key, value))
		}
	}
	return funcs
}

/*
Utility method for Debugging purpose. This method helps list assertions as well
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
