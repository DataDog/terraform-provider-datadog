package fwprovider

import (
	"context"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
// inconsistent result after apply. Email is an optional block, but the API
// returns email.enabled=false even when the request omitted email. The provider
// must preserve whether the practitioner configured the block.
func TestTeamNotificationRuleEmailPreservesNullVsDisabled(t *testing.T) {
	r := &teamNotificationRuleResource{}

	t.Run("omitted block stays omitted", func(t *testing.T) {
		state := teamNotificationRuleModel{
			Email:   nil,
			MsTeams: &msTeamsModel{ConnectorName: types.StringValue("connector")},
		}

		disabled := false
		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(&disabled))

		if state.Email != nil {
			t.Fatalf("expected email to stay nil, got %#v", state.Email)
		}
	})

	t.Run("explicitly disabled block stays present", func(t *testing.T) {
		state := teamNotificationRuleModel{
			Email: &emailModel{Enabled: types.BoolValue(false)},
		}

		disabled := false
		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(&disabled))

		if state.Email == nil {
			t.Fatal("expected explicitly disabled email block to remain present")
		}
		if state.Email.Enabled.IsNull() || state.Email.Enabled.ValueBool() {
			t.Fatalf("expected email.enabled to be false, got %#v", state.Email.Enabled)
		}
	})

	t.Run("API enabled email is retained", func(t *testing.T) {
		enabled := true
		state := teamNotificationRuleModel{Email: nil}

		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(&enabled))

		if state.Email == nil || !state.Email.Enabled.ValueBool() {
			t.Fatalf("expected email.enabled to be true, got %#v", state.Email)
		}
	})
}

func TestTeamNotificationRuleRecipientEmailState(t *testing.T) {
	r := &teamNotificationRuleResource{}

	newResponse := func(recipient string) *datadogV2.TeamNotificationRule {
		rule := newTeamNotificationRuleResponse(datadog.PtrBool(true))
		email := rule.Attributes.GetEmail()
		email.SetRecipientEmail(recipient)
		rule.Attributes.SetEmail(email)
		return rule
	}

	t.Run("recipient is read into state", func(t *testing.T) {
		state := teamNotificationRuleModel{}
		r.updateState(context.Background(), &state, newResponse(" Team-Alerts@Example.com "))

		if state.Email == nil {
			t.Fatal("expected email block to be present")
		}
		// The API stores the address exactly as sent, so state must match it verbatim.
		if got := state.Email.RecipientEmail.ValueString(); got != " Team-Alerts@Example.com " {
			t.Fatalf("expected recipient to be kept verbatim, got %q", got)
		}
	})

	t.Run("missing recipient is null", func(t *testing.T) {
		state := teamNotificationRuleModel{}
		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(datadog.PtrBool(true)))

		if state.Email == nil || !state.Email.RecipientEmail.IsNull() {
			t.Fatalf("expected recipient_email to be null, got %#v", state.Email)
		}
	})

	t.Run("empty recipient is null", func(t *testing.T) {
		state := teamNotificationRuleModel{}
		r.updateState(context.Background(), &state, newResponse(""))

		if state.Email == nil || !state.Email.RecipientEmail.IsNull() {
			t.Fatalf("expected recipient_email to be null, got %#v", state.Email)
		}
	})
}

func TestTeamNotificationRuleServicenowState(t *testing.T) {
	r := &teamNotificationRuleResource{}

	t.Run("templates are read in order with duplicates", func(t *testing.T) {
		rule := newTeamNotificationRuleResponse(datadog.PtrBool(false))
		servicenow := datadogV2.NewTeamNotificationRuleAttributesServiceNowWithDefaults()
		servicenow.SetTemplates([]string{"b", "a", "b"})
		rule.Attributes.SetServicenow(*servicenow)

		state := teamNotificationRuleModel{}
		r.updateState(context.Background(), &state, rule)

		if state.Servicenow == nil {
			t.Fatal("expected servicenow block to be present")
		}
		var templates []string
		state.Servicenow.Templates.ElementsAs(context.Background(), &templates, false)
		if strings.Join(templates, ",") != "b,a,b" {
			t.Fatalf("expected templates [b a b], got %v", templates)
		}
		// email.enabled=false is an API default and was not configured.
		if state.Email != nil {
			t.Fatalf("expected email to stay nil, got %#v", state.Email)
		}
	})

	t.Run("missing servicenow stays nil", func(t *testing.T) {
		state := teamNotificationRuleModel{}
		r.updateState(context.Background(), &state, newTeamNotificationRuleResponse(datadog.PtrBool(true)))

		if state.Servicenow != nil {
			t.Fatalf("expected servicenow to be nil, got %#v", state.Servicenow)
		}
	})
}

func TestTeamNotificationRuleRequestBody(t *testing.T) {
	r := &teamNotificationRuleResource{}
	ctx := context.Background()

	t.Run("recipient and templates are sent", func(t *testing.T) {
		state := teamNotificationRuleModel{
			Email: &emailModel{
				Enabled:        types.BoolValue(true),
				RecipientEmail: types.StringValue("team-alerts@example.com"),
			},
			Servicenow: &servicenowModel{
				Templates: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("b"), types.StringValue("a")}),
			},
		}

		req, diags := r.buildTeamNotificationRuleRequestBody(ctx, &state)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if got := req.Data.Attributes.Email.GetRecipientEmail(); got != "team-alerts@example.com" {
			t.Fatalf("expected recipient_email to be sent, got %q", got)
		}
		if got := strings.Join(req.Data.Attributes.Servicenow.GetTemplates(), ","); got != "b,a" {
			t.Fatalf("expected templates [b a], got %v", got)
		}
	})

	t.Run("null recipient and servicenow are omitted", func(t *testing.T) {
		state := teamNotificationRuleModel{
			Email: &emailModel{Enabled: types.BoolValue(true), RecipientEmail: types.StringNull()},
		}

		req, diags := r.buildTeamNotificationRuleRequestBody(ctx, &state)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if _, ok := req.Data.Attributes.Email.GetRecipientEmailOk(); ok {
			t.Fatal("expected recipient_email to be omitted")
		}
		if req.Data.Attributes.Servicenow != nil {
			t.Fatalf("expected servicenow to be omitted, got %#v", req.Data.Attributes.Servicenow)
		}
	})
}

// validateTeamNotificationRuleConfig runs the resource's config validator
// against a config built from the given model.
func validateTeamNotificationRuleConfig(t *testing.T, model teamNotificationRuleModel) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	(&teamNotificationRuleResource{}).Schema(ctx, resource.SchemaRequest{}, schemaResp)

	// tfsdk.State is used only to encode the model into a raw config value.
	state := tfsdk.State{Schema: schemaResp.Schema}
	if diags := state.Set(ctx, &model); diags.HasError() {
		t.Fatalf("failed to build config: %v", diags)
	}

	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: state.Raw}}
	resp := &resource.ValidateConfigResponse{}
	(&teamNotificationRuleValidator{}).ValidateResource(ctx, req, resp)
	return resp.Diagnostics
}

func TestTeamNotificationRuleValidator(t *testing.T) {
	templates := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("incident-template")})

	cases := []struct {
		name      string
		model     teamNotificationRuleModel
		wantError string
	}{
		{
			name: "recipient with email enabled",
			model: teamNotificationRuleModel{Email: &emailModel{
				Enabled: types.BoolValue(true), RecipientEmail: types.StringValue("team-alerts@example.com"),
			}},
		},
		{
			name: "recipient with email disabled",
			model: teamNotificationRuleModel{
				Email:   &emailModel{Enabled: types.BoolValue(false), RecipientEmail: types.StringValue("team-alerts@example.com")},
				MsTeams: &msTeamsModel{ConnectorName: types.StringValue("connector")},
			},
			wantError: "Invalid Email Configuration",
		},
		{
			name: "recipient without enabled",
			model: teamNotificationRuleModel{
				Email:   &emailModel{Enabled: types.BoolNull(), RecipientEmail: types.StringValue("team-alerts@example.com")},
				MsTeams: &msTeamsModel{ConnectorName: types.StringValue("connector")},
			},
			wantError: "Invalid Email Configuration",
		},
		{
			name: "unknown recipient is not validated",
			model: teamNotificationRuleModel{
				Email:   &emailModel{Enabled: types.BoolValue(false), RecipientEmail: types.StringUnknown()},
				MsTeams: &msTeamsModel{ConnectorName: types.StringValue("connector")},
			},
		},
		{
			name:  "servicenow only",
			model: teamNotificationRuleModel{Servicenow: &servicenowModel{Templates: templates}},
		},
		{
			name:      "no notification types",
			model:     teamNotificationRuleModel{},
			wantError: "Missing Notification Configuration",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateTeamNotificationRuleConfig(t, tc.model)

			if tc.wantError == "" {
				if diags.HasError() {
					t.Fatalf("expected no errors, got %v", diags)
				}
				return
			}
			if !diags.HasError() || diags.Errors()[0].Summary() != tc.wantError {
				t.Fatalf("expected %q error, got %v", tc.wantError, diags)
			}
		})
	}
}
