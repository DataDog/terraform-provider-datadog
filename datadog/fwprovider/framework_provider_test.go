package fwprovider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDefaultConfigureFuncRetryConfiguration(t *testing.T) {
	p := New().(*FrameworkProvider)
	config := &ProviderSchema{
		ApiUrl:                           types.StringValue("https://api.datadoghq.com"),
		Validate:                         types.StringValue("false"),
		HttpClientRetryEnabled:           types.StringValue("true"),
		HttpClientRetryTimeout:           types.Int64Value(300),
		HttpClientRetryBackoffMultiplier: types.Int64Value(7),
		HttpClientRetryBackoffBase:       types.Int64Value(4),
		HttpClientRetryMaxRetries:        types.Int64Value(5),
		HttpClientRetryJitter:            types.Int64Value(9),
	}

	diags := defaultConfigureFunc(p, &provider.ConfigureRequest{}, config)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	retryConfig := p.DatadogApiInstances.HttpClient.GetConfig().RetryConfiguration
	if !retryConfig.EnableRetry {
		t.Error("expected retries to be enabled")
	}
	if got, want := retryConfig.HTTPRetryTimeout, 300*time.Second; got != want {
		t.Errorf("HTTPRetryTimeout = %s, want %s", got, want)
	}
	if got, want := retryConfig.BackOffMultiplier, float64(7); got != want {
		t.Errorf("BackOffMultiplier = %v, want %v", got, want)
	}
	if got, want := retryConfig.BackOffBase, float64(4); got != want {
		t.Errorf("BackOffBase = %v, want %v", got, want)
	}
	if got, want := retryConfig.MaxRetries, 5; got != want {
		t.Errorf("MaxRetries = %d, want %d", got, want)
	}
	if got, want := retryConfig.RetryJitter, 9*time.Second; got != want {
		t.Errorf("RetryJitter = %s, want %s", got, want)
	}

	clientConfig := p.DatadogApiInstances.HttpClient.GetConfig()
	for _, operation := range []string{"v2.CreateFleetSchedule", "v2.UpdateFleetSchedule", "v2.DeleteFleetSchedule"} {
		if !clientConfig.IsUnstableOperationEnabled(operation) {
			t.Errorf("%s should be enabled", operation)
		}
	}
	for _, stableOperation := range []string{"v2.GetFleetScheduleV2", "v2.ListFleetSchedulesV2"} {
		if clientConfig.IsUnstableOperationEnabled(stableOperation) {
			t.Errorf("stable operation %s should not be enabled as unstable", stableOperation)
		}
	}
}
