package test

import (
	"context"
	"os"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// newSweepAPIClient creates an authenticated API client for sweep operations.
// Returns a nil client if credentials are not set (caller should skip).
func newSweepAPIClient(t *testing.T) (context.Context, *common.APIClient) {
	t.Helper()

	apiKey := os.Getenv(testAPIKeyEnvName)
	appKey := os.Getenv(testAPPKeyEnvName)
	apiURL := os.Getenv(testAPIUrlEnvName)

	if apiKey == "" || appKey == "" {
		t.Log("sweep: DD_TEST_CLIENT_API_KEY or DD_TEST_CLIENT_APP_KEY not set, skipping")
		return nil, nil
	}

	ctx, err := buildContext(context.Background(), apiKey, appKey, apiURL)
	if err != nil {
		t.Logf("sweep: failed to build API context: %v", err)
		return nil, nil
	}

	cfg := common.NewConfiguration()
	client := common.NewAPIClient(cfg)
	return ctx, client
}
