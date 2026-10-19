package fwprovider

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// newIncidentNotificationRuleResponse builds a minimal API response whose
// renotify_on attribute is set to renotifyOn (which may be an empty slice to
// mimic the API echoing back "renotify_on": []).
func newIncidentNotificationRuleResponse(renotifyOn []string) *datadogV2.IncidentNotificationRule {
	attrs := datadogV2.NewIncidentNotificationRuleAttributesWithDefaults()
	attrs.SetTrigger("incident_created_trigger")
	attrs.SetRenotifyOn(renotifyOn)

	data := datadogV2.NewIncidentNotificationRuleResponseData(
		uuid.Nil,
		datadogV2.INCIDENTNOTIFICATIONRULETYPE_INCIDENT_NOTIFICATION_RULES,
	)
	data.SetAttributes(*attrs)

	return datadogV2.NewIncidentNotificationRule(*data)
}

// TestIncidentNotificationRuleRenotifyOnPreservesNullVsEmpty guards the fix for
// the "inconsistent result after apply" error: renotify_on is Optional (not
// Computed), so the flatten must not turn the API's empty list into an empty
// value when the practitioner left the attribute null.
func TestIncidentNotificationRuleRenotifyOnPreservesNullVsEmpty(t *testing.T) {
	r := &incidentNotificationRuleResource{}

	// Attribute omitted in config -> incoming state is nil -> must stay nil (null),
	// even though the API echoes back an empty list.
	t.Run("null stays null", func(t *testing.T) {
		state := incidentNotificationRuleModel{RenotifyOn: nil}
		r.updateStateFromResponse(&state, newIncidentNotificationRuleResponse([]string{}))
		if state.RenotifyOn != nil {
			t.Fatalf("expected renotify_on to stay nil (null), got %#v", state.RenotifyOn)
		}
	})

	// Attribute set to [] in config -> incoming state is a non-nil empty slice ->
	// must stay an empty (non-null) list.
	t.Run("explicit empty stays empty", func(t *testing.T) {
		state := incidentNotificationRuleModel{RenotifyOn: []types.String{}}
		r.updateStateFromResponse(&state, newIncidentNotificationRuleResponse([]string{}))
		if state.RenotifyOn == nil || len(state.RenotifyOn) != 0 {
			t.Fatalf("expected renotify_on to stay empty (non-null), got %#v", state.RenotifyOn)
		}
	})

	// A non-empty API list is authoritative and overwrites the incoming value.
	t.Run("non-empty overwrites", func(t *testing.T) {
		state := incidentNotificationRuleModel{RenotifyOn: nil}
		r.updateStateFromResponse(&state, newIncidentNotificationRuleResponse([]string{"status", "severity"}))
		if len(state.RenotifyOn) != 2 || state.RenotifyOn[0].ValueString() != "status" || state.RenotifyOn[1].ValueString() != "severity" {
			t.Fatalf("expected renotify_on = [status severity], got %#v", state.RenotifyOn)
		}
	})
}
