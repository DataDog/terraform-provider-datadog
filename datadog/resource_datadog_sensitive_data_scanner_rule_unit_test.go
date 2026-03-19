package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestShouldSaveMatchFlattenViaProperField verifies that should_save_match is read
// from the ShouldSaveMatch struct field (via GetShouldSaveMatchOk), not from
// AdditionalProperties. This is a regression test for a bug in provider v3.88
// where the value was read from AdditionalProperties["should_save_match"], but the
// API client (v2.54.1+) moved should_save_match to a proper struct field and stopped
// populating AdditionalProperties for it, causing the state to always read false.
func TestShouldSaveMatchFlattenViaProperField(t *testing.T) {
	r := resourceDatadogSensitiveDataScannerRule()
	d := schema.TestResourceDataRaw(t, r.SchemaFunc(), map[string]interface{}{})
	d.SetId("test-id")

	shouldSaveMatchTrue := true
	replacementType := datadogV2.SENSITIVEDATASCANNERTEXTREPLACEMENTTYPE_REPLACEMENT_STRING
	replacementString := "REDACTED"

	// Simulate the API client behavior: should_save_match is in ShouldSaveMatch field,
	// NOT in AdditionalProperties (the client deletes known fields from AdditionalProperties).
	attrs := datadogV2.NewSensitiveDataScannerRuleAttributes()
	attrs.SetTextReplacement(datadogV2.SensitiveDataScannerTextReplacement{
		Type:              &replacementType,
		ReplacementString: &replacementString,
		ShouldSaveMatch:   &shouldSaveMatchTrue,
		// AdditionalProperties intentionally empty — should_save_match is NOT here
		AdditionalProperties: map[string]interface{}{},
	})

	diags := updateSensitiveDataScannerRuleState(d, attrs)
	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	tr := d.Get("text_replacement").([]interface{})[0].(map[string]interface{})
	if got := tr["should_save_match"]; got != true {
		t.Errorf("should_save_match = %v (%T), want true\n"+
			"Hint: the value must be read via GetShouldSaveMatchOk(), not AdditionalProperties.\n"+
			"In provider v3.88, AdditionalProperties was used, but api-client v2.54.1+ moves\n"+
			"should_save_match to a proper struct field and removes it from AdditionalProperties.", got, got)
	}
}
