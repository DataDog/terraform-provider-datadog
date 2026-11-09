package test

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// Builds the sweep client with placeholder keys; no request is sent.
func TestNewSweepAPIClientAppliesRetryEnv(t *testing.T) {
	t.Setenv(testAPIKeyEnvName, "placeholder")
	t.Setenv(testAPPKeyEnvName, "placeholder")
	t.Setenv(testAPIUrlEnvName, "")
	t.Setenv(utils.DDHTTPRetryEnabled, "")
	t.Setenv(utils.DDHTTPRetryMaxRetries, "5")

	_, client := newSweepAPIClient(t)
	rc := client.GetConfig().RetryConfiguration
	if !rc.EnableRetry || rc.MaxRetries != 5 {
		t.Fatalf("unexpected retry config %+v", rc)
	}

	t.Setenv(utils.DDHTTPRetryEnabled, "false")
	_, client = newSweepAPIClient(t)
	if client.GetConfig().RetryConfiguration.EnableRetry {
		t.Fatal("retry enabled with DD_HTTP_CLIENT_RETRY_ENABLED=false")
	}
}
