package test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

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
// as the tag configuration create endpoint.
func waitForMetricIndexed(t *testing.T, apiKey, appKey, apiURL, metricName string) {
	client := &http.Client{Timeout: 30 * time.Second}
	endpoint := fmt.Sprintf("%s/api/v2/metrics/%s/all-tags", apiURL, url.PathEscape(metricName))
	deadline := time.Now().Add(metricIndexPollTimeout)
	var lastStatus int
	for {
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Fatalf("error creating request for metric %s: %v", metricName, err)
		}
		req.Header.Set("DD-API-KEY", apiKey)
		req.Header.Set("DD-APPLICATION-KEY", appKey)
		resp, err := client.Do(req)
		if err == nil {
			lastStatus = resp.StatusCode
			resp.Body.Close()
			if lastStatus == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("metric %s not indexed after %s (last status %d, last error %v)", metricName, metricIndexPollTimeout, lastStatus, err)
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
