package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

// ApplyHTTPRetryConfigFromEnv mirrors the provider env var handling for HTTP retries
// (fwprovider ConfigureConfigDefaults/ValidateConfigValues), so acceptance
// tests retry 429 and 5xx responses like the provider does. Retry is on when
// DD_HTTP_CLIENT_RETRY_ENABLED is unset. forceEnable (the "http_retry_enable"
// test context value)
// forces retry on.
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
	if forceEnable {
		enabled = true
	}
	if !enabled {
		config.RetryConfiguration.EnableRetry = false
		return nil
	}
	config.RetryConfiguration.EnableRetry = true

	if v, ok, err := retryEnvInt(DDHTTPRetryTimeout, 0, -1); err != nil {
		return err
	} else if ok {
		config.RetryConfiguration.HTTPRetryTimeout = time.Duration(v) * time.Second
	}
	if v, ok, err := retryEnvInt(DDHTTPRetryBackoffMultiplier, 1, -1); err != nil {
		return err
	} else if ok {
		config.RetryConfiguration.BackOffMultiplier = float64(v)
	}
	if v, ok, err := retryEnvInt(DDHTTPRetryBackoffBase, 1, -1); err != nil {
		return err
	} else if ok {
		config.RetryConfiguration.BackOffBase = float64(v)
	}
	if v, ok, err := retryEnvInt(DDHTTPRetryMaxRetries, 1, 5); err != nil {
		return err
	} else if ok {
		config.RetryConfiguration.MaxRetries = v
	}
	if v, ok, err := retryEnvInt(DDHTTPRetryJitter, 0, -1); err != nil {
		return err
	} else if ok {
		config.RetryConfiguration.RetryJitter = time.Duration(v) * time.Second
	}
	return nil
}

// retryEnvInt reads an integer env var. A max below 0 means no upper bound.
// The provider ignores Atoi errors (value becomes 0, then fails validation for
// most settings); an explicit error is stricter but gives the same accept set.
func retryEnvInt(name string, min, max int) (int, bool, error) {
	raw, err := GetMultiEnvVar(name)
	if err != nil {
		return 0, false, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, fmt.Errorf("%s must be an integer, got %q", name, raw)
	}
	if v < min || (max >= 0 && v > max) {
		return 0, false, fmt.Errorf("%s out of range: %d", name, v)
	}
	return v, true, nil
}
