package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// newRuleOptionsWithDecreaseCriticality builds rule options carrying
// decreaseCriticalityBasedOnEnv, mimicking a rule that has the option enabled
// in Datadog.
func newRuleOptionsWithDecreaseCriticality(decrease bool) datadogV2.SecurityMonitoringRuleOptions {
	options := datadogV2.NewSecurityMonitoringRuleOptions()
	options.SetDecreaseCriticalityBasedOnEnv(decrease)
	return *options
}

// TestSecurityMonitoringRuleDecreaseCriticalityReadMatchesWrite guards the fix
// for the "inconsistent result after apply" error on rules that are not
// log_detection. buildPayloadOptions only sends
// decrease_criticality_based_on_env for log_detection rules, so reading the API
// value back for any other type plans a change the provider can never apply.
func TestSecurityMonitoringRuleDecreaseCriticalityReadMatchesWrite(t *testing.T) {
	ctx := context.Background()
	logDetection := string(datadogV2.SECURITYMONITORINGRULETYPECREATE_LOG_DETECTION)
	applicationSecurity := string(datadogV2.SECURITYMONITORINGRULETYPECREATE_APPLICATION_SECURITY)

	// log_detection is the one type the provider writes the option for, so the
	// API value is authoritative.
	t.Run("log_detection follows the api", func(t *testing.T) {
		got, diags := extractTfOptions(ctx, newRuleOptionsWithDecreaseCriticality(true), nil, logDetection)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !got[0].DecreaseCriticalityBasedOnEnv.ValueBool() {
			t.Fatal("expected decrease_criticality_based_on_env to follow the API value true")
		}
	})

	// Importing an application_security rule that has the option enabled must not
	// overwrite the configured false, which is a value the provider cannot change.
	t.Run("application_security keeps configured false", func(t *testing.T) {
		prior := []ruleOptionsModel{{DecreaseCriticalityBasedOnEnv: types.BoolValue(false)}}
		got, diags := extractTfOptions(ctx, newRuleOptionsWithDecreaseCriticality(true), prior, applicationSecurity)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if got[0].DecreaseCriticalityBasedOnEnv.ValueBool() {
			t.Fatal("expected decrease_criticality_based_on_env to stay false, the API value is not manageable for this rule type")
		}
	})

	// A fresh import has no prior state, so the schema default applies.
	t.Run("application_security without prior state defaults to false", func(t *testing.T) {
		got, diags := extractTfOptions(ctx, newRuleOptionsWithDecreaseCriticality(true), nil, applicationSecurity)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if got[0].DecreaseCriticalityBasedOnEnv.ValueBool() {
			t.Fatal("expected decrease_criticality_based_on_env to default to false")
		}
	})

	// The same holds in the other direction: the API drops the option when it
	// creates an application_security rule, and the configured value must survive.
	t.Run("application_security keeps configured true", func(t *testing.T) {
		prior := []ruleOptionsModel{{DecreaseCriticalityBasedOnEnv: types.BoolValue(true)}}
		got, diags := extractTfOptions(ctx, *datadogV2.NewSecurityMonitoringRuleOptions(), prior, applicationSecurity)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !got[0].DecreaseCriticalityBasedOnEnv.ValueBool() {
			t.Fatal("expected decrease_criticality_based_on_env to stay true when the API omits it")
		}
	})
}
