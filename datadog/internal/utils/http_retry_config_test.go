package utils

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

var retryEnvVars = []string{
	DDHTTPRetryEnabled, DDHTTPRetryTimeout, DDHTTPRetryBackoffMultiplier,
	DDHTTPRetryBackoffBase, DDHTTPRetryMaxRetries, DDHTTPRetryJitter,
}

func TestApplyHTTPRetryConfig(t *testing.T) {
	defaults := datadog.NewConfiguration().RetryConfiguration
	cases := []struct {
		name    string
		env     map[string]string
		force   bool
		wantErr bool
		check   func(t *testing.T, rc datadog.RetryConfiguration)
	}{
		{name: "unset enables retry with client defaults", check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if !rc.EnableRetry || rc.MaxRetries != defaults.MaxRetries || rc.HTTPRetryTimeout != defaults.HTTPRetryTimeout {
				t.Errorf("unexpected config %+v", rc)
			}
		}},
		{name: "true", env: map[string]string{DDHTTPRetryEnabled: "true"}, check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if !rc.EnableRetry {
				t.Error("retry disabled")
			}
		}},
		{name: "false with invalid max retries", env: map[string]string{DDHTTPRetryEnabled: "false", DDHTTPRetryMaxRetries: "9"}, wantErr: true},
		{name: "false with non-integer base", env: map[string]string{DDHTTPRetryEnabled: "false", DDHTTPRetryBackoffBase: "x"}, wantErr: true},
		{name: "negative timeout accepted", env: map[string]string{DDHTTPRetryTimeout: "-5"}, check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if rc.HTTPRetryTimeout != -5*time.Second {
				t.Errorf("unexpected timeout %v", rc.HTTPRetryTimeout)
			}
		}},
		{name: "non-integer jitter", env: map[string]string{DDHTTPRetryJitter: "2s"}, wantErr: true},
		{name: "false", env: map[string]string{DDHTTPRetryEnabled: "false", DDHTTPRetryMaxRetries: "5"}, check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if rc.EnableRetry {
				t.Error("retry enabled")
			}
		}},
		{name: "context flag overrides false", env: map[string]string{DDHTTPRetryEnabled: "false"}, force: true, check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if !rc.EnableRetry {
				t.Error("retry disabled")
			}
		}},
		{name: "numeric settings", env: map[string]string{
			DDHTTPRetryTimeout: "30", DDHTTPRetryBackoffMultiplier: "3", DDHTTPRetryBackoffBase: "4",
			DDHTTPRetryMaxRetries: "5", DDHTTPRetryJitter: "2",
		}, check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if rc.HTTPRetryTimeout != 30*time.Second || rc.BackOffMultiplier != 3 || rc.BackOffBase != 4 ||
				rc.MaxRetries != 5 || rc.RetryJitter != 2*time.Second {
				t.Errorf("unexpected config %+v", rc)
			}
		}},
		{name: "jitter zero allowed", env: map[string]string{DDHTTPRetryJitter: "0"}, check: func(t *testing.T, rc datadog.RetryConfiguration) {
			if rc.RetryJitter != 0 {
				t.Error("jitter not zero")
			}
		}},
		{name: "invalid enabled", env: map[string]string{DDHTTPRetryEnabled: "yes"}, wantErr: true},
		{name: "non-integer timeout", env: map[string]string{DDHTTPRetryTimeout: "abc"}, wantErr: true},
		{name: "multiplier below 1", env: map[string]string{DDHTTPRetryBackoffMultiplier: "0"}, wantErr: true},
		{name: "base below 1", env: map[string]string{DDHTTPRetryBackoffBase: "0"}, wantErr: true},
		{name: "max retries above 5", env: map[string]string{DDHTTPRetryMaxRetries: "6"}, wantErr: true},
		{name: "max retries below 1", env: map[string]string{DDHTTPRetryMaxRetries: "0"}, wantErr: true},
		{name: "negative jitter", env: map[string]string{DDHTTPRetryJitter: "-1"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, k := range retryEnvVars {
				t.Setenv(k, "")
			}
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			config := datadog.NewConfiguration()
			err := ApplyHTTPRetryConfigFromEnv(tc.force, config)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, config.RetryConfiguration)
		})
	}
}
