package utils

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// ApplyHTTPRetryConfigFromEnv applies the provider's DD_HTTP_CLIENT_RETRY_* defaults and
// validation: retry is on when unset, forceEnable overrides an explicit false, and invalid
// values return an error.
func ApplyHTTPRetryConfigFromEnv(forceEnable bool, config *datadog.Configuration) error {
	enabled := true
	if v, err := GetMultiEnvVar(DDHTTPRetryEnabled); err == nil {
		switch v {
		case "true":
		case "false":
			enabled = false
		default:
			return fmt.Errorf("%s must be one of true, false, got %q", DDHTTPRetryEnabled, v)
		}
	}

	timeout, hasTimeout, err := retryEnvInt(DDHTTPRetryTimeout, math.MinInt, math.MaxInt)
	if err != nil {
		return err
	}
	multiplier, hasMultiplier, err := retryEnvInt(DDHTTPRetryBackoffMultiplier, 1, math.MaxInt)
	if err != nil {
		return err
	}
	base, hasBase, err := retryEnvInt(DDHTTPRetryBackoffBase, 1, math.MaxInt)
	if err != nil {
		return err
	}
	maxRetries, hasMaxRetries, err := retryEnvInt(DDHTTPRetryMaxRetries, 1, 5)
	if err != nil {
		return err
	}
	jitter, hasJitter, err := retryEnvInt(DDHTTPRetryJitter, 0, math.MaxInt)
	if err != nil {
		return err
	}

	config.RetryConfiguration.EnableRetry = enabled || forceEnable
	if !config.RetryConfiguration.EnableRetry {
		return nil
	}
	if hasTimeout {
		config.RetryConfiguration.HTTPRetryTimeout = time.Duration(timeout) * time.Second
	}
	if hasMultiplier {
		config.RetryConfiguration.BackOffMultiplier = float64(multiplier)
	}
	if hasBase {
		config.RetryConfiguration.BackOffBase = float64(base)
	}
	if hasMaxRetries {
		config.RetryConfiguration.MaxRetries = maxRetries
	}
	if hasJitter {
		config.RetryConfiguration.RetryJitter = time.Duration(jitter) * time.Second
	}
	return nil
}

func retryEnvInt(name string, min, max int) (int, bool, error) {
	raw, err := GetMultiEnvVar(name)
	if err != nil {
		return 0, false, nil
	}
	v, err := strconv.Atoi(raw)
	// Deviation from the provider: its Atoi error becomes 0, and a 0s timeout fails every request.
	if err != nil {
		return 0, false, fmt.Errorf("%s must be an integer, got %q", name, raw)
	}
	if v < min || v > max {
		return 0, false, fmt.Errorf("%s out of range: %d", name, v)
	}
	return v, true, nil
}
