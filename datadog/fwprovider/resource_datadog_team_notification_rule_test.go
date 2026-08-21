package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newTeamNotificationRuleResponse(emailEnabled *bool) *datadogV2.TeamNotificationRule {
	attributes := datadogV2.NewTeamNotificationRuleAttributesWithDefaults()
	if emailEnabled != nil {
		email := datadogV2.NewTeamNotificationRuleAttributesEmailWithDefaults()
		email.SetEnabled(*emailEnabled)
		attributes.SetEmail(*email)
	}

	rule := datadogV2.NewTeamNotificationRuleWithDefaults()
	rule.SetId("rule-id")
	rule.SetAttributes(*attributes)
	return rule
}

// TestTeamNotificationRuleEmailPreservesNullVsDisabled guards against an
// inconsistent result after apply. Email is an optional block, so an omitted
// block must remain nil when the API omits email from the response. An
// explicitly disabled block must remain present because the API represents it
// with the same omission.
func TestTeamNotificationRuleEmailPreservesNullVsDisabled(t *testing.T) {
	r := &teamNotificationRuleResource{}

	t.Run("omitted block stays omitted", func(t *testing.T) {
		state := teamNotificationRuleModel{
			Email:   nil,
			MsTeams: &msTeamsModel{ConnectorName: types.StringValue("connector")},
		}

		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(nil))

		if state.Email != nil {
			t.Fatalf("expected email to stay nil, got %#v", state.Email)
		}
	})

	t.Run("explicitly disabled block stays present", func(t *testing.T) {
		state := teamNotificationRuleModel{
			Email: &emailModel{Enabled: types.BoolValue(false)},
		}

		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(nil))

		if state.Email == nil {
			t.Fatal("expected explicitly disabled email block to remain present")
		}
		if state.Email.Enabled.IsNull() || state.Email.Enabled.ValueBool() {
			t.Fatalf("expected email.enabled to be false, got %#v", state.Email.Enabled)
		}
	})

	t.Run("API enabled email is authoritative", func(t *testing.T) {
		enabled := true
		state := teamNotificationRuleModel{Email: nil}

		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(&enabled))

		if state.Email == nil || !state.Email.Enabled.ValueBool() {
			t.Fatalf("expected email.enabled to be true, got %#v", state.Email)
		}
	})
}
