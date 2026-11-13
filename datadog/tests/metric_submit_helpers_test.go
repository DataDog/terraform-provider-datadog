package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	datadogCommunity "github.com/zorkian/go-datadog-api"
)

const (
	metricIndexPollInterval = 5 * time.Second
	metricIndexPollTimeout  = 5 * time.Minute
)

// submitTestMetricPreConfig returns a PreConfig function that submits a metric and waits
// until the metrics index returns it. Tag configuration and metric metadata reads fail
// with "metric does not exist" until intake has indexed the metric, and the indexing
// delay is not bounded by the API.
// Skipped in replay: the cassette holds only the provider's own API interactions.
func submitTestMetricPreConfig(t *testing.T, metricName string, metricType string) func() {
	return func() {
		if isReplaying() {
			return
		}
		apiKey := os.Getenv(testAPIKeyEnvName)
		appKey := os.Getenv(testAPPKeyEnvName)
		apiURL := os.Getenv(testAPIUrlEnvName)
		if apiURL == "" {
			apiURL = "https://api.datadoghq.com"
		}

		if metricType == "distribution" {
			// The community client has no distribution submission.
			submitDistributionMetric(t, apiKey, apiURL, metricName)
		} else {
			client := datadogCommunity.NewClient(apiKey, appKey)
			client.SetBaseUrl(apiURL)
			ts := float64(time.Now().Unix())
			val := float64(1)
			metric := datadogCommunity.Metric{
				Metric: datadogCommunity.String(metricName),
				Points: []datadogCommunity.DataPoint{{&ts, &val}},
				Type:   datadogCommunity.String(metricType),
				Tags:   []string{"sport:soccer", "datacenter:us-east"},
			}
			if err := client.PostMetrics([]datadogCommunity.Metric{metric}); err != nil {
				t.Fatalf("error submitting test metric %s: %v", metricName, err)
			}
		}
		waitForMetricIndexed(t, apiKey, appKey, apiURL, metricName)
	}
}

// waitForMetricIndexed polls the metric tags endpoint, which uses the same existence check
// as the tag configuration create endpoint. The client is separate from the provider client so
// that polls are not recorded in cassettes.
func waitForMetricIndexed(t *testing.T, apiKey, appKey, apiURL, metricName string) {
	cfg := datadog.NewConfiguration()
	cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	cfg.Servers = datadog.ServerConfigurations{{URL: apiURL}}
	api := datadogV2.NewMetricsApi(datadog.NewAPIClient(cfg))
	ctx := context.WithValue(context.Background(), datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: apiKey},
		"appKeyAuth": {Key: appKey},
	})
	deadline := time.Now().Add(metricIndexPollTimeout)
	for {
		_, httpResp, err := api.ListTagsByMetricName(ctx, metricName)
		if err == nil {
			return
		}
		status := 0
		if httpResp != nil {
			status = httpResp.StatusCode
		}
		if time.Now().After(deadline) {
			t.Fatalf("metric %s not indexed after %s (last status %d, last error %v)", metricName, metricIndexPollTimeout, status, err)
		}
		time.Sleep(metricIndexPollInterval)
	}
}

// submitDistributionMetric submits a distribution metric using the V1 distribution_points API.
func submitDistributionMetric(t *testing.T, apiKey, apiURL, metricName string) {
	timestamp := float64(time.Now().Unix())
	body := fmt.Sprintf(`{"series":[{"metric":"%s","points":[[%f,[1.0,2.0,3.0,4.0,5.0]]],"tags":["sport:soccer","datacenter:us-east"]}]}`,
		metricName, timestamp)
	req, err := http.NewRequest("POST", apiURL+"/api/v1/distribution_points", strings.NewReader(body))
	if err != nil {
		t.Fatalf("error creating request for distribution metric %s: %v", metricName, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", apiKey)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("error submitting distribution metric %s: %v", metricName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("error submitting distribution metric %s: status %d, body: %s", metricName, resp.StatusCode, string(b))
	}
}
