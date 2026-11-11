package fwprovider

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func policyTarget(provider, model string) aiGatewayPolicyTarget {
	target := aiGatewayPolicyTarget{Provider: types.StringNull(), Model: types.StringValue(model)}
	if provider != "" {
		target.Provider = types.StringValue(provider)
	}
	return target
}

func TestBuildAiGatewayPolicyRequestAttributes(t *testing.T) {
	state := aiGatewayPolicyModel{
		Inbound:  types.StringValue("coding-agent"),
		UserID:   types.StringNull(),
		Priority: types.Int64Value(3),
		Tiers: []aiGatewayPolicyTier{
			{Targets: []aiGatewayPolicyTarget{policyTarget("prod-us", "gpt-5.2-codex")}},
			{Targets: []aiGatewayPolicyTarget{policyTarget("", "claude-sonnet-5*"), policyTarget("*", "gpt-5*")}},
		},
	}

	body, err := json.Marshal(buildAiGatewayPolicyRequestAttributes(&state))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"inbound":"coding-agent","priority":3,"tiers":[` +
		`{"priority":0,"targets":[{"model":"gpt-5.2-codex","provider":"prod-us"}]},` +
		`{"priority":1,"targets":[{"model":"claude-sonnet-5*"},{"model":"gpt-5*","provider":"*"}]}]}`
	if string(body) != want {
		t.Errorf("request body\n got: %s\nwant: %s", body, want)
	}
}

func policyResponse(inbound string, tiers ...[][2]string) datadogV2.AiGatewayPolicyResource {
	apiTiers := make([]datadogV2.AiGatewayPolicyTier, 0, len(tiers))
	for _, tier := range tiers {
		targets := make([]datadogV2.AiGatewayPolicyTarget, 0, len(tier))
		for i, pm := range tier {
			target := datadogV2.NewAiGatewayPolicyTarget(int64(i+1), 3)
			if pm[0] != "" {
				target.Provider = *datadog.NewNullableString(datadog.PtrString(pm[0]))
			}
			target.SetModel(pm[1])
			targets = append(targets, *target)
		}
		apiTiers = append(apiTiers, *datadogV2.NewAiGatewayPolicyTier(targets))
	}
	return *datadogV2.NewAiGatewayPolicyResource(*datadogV2.NewAiGatewayPolicyAttributes(3, apiTiers), inbound, datadogV2.AIGATEWAYPOLICYTYPE_ROUTING_DOCUMENTS)
}

func TestUpdateAiGatewayPolicyState(t *testing.T) {
	configured := []aiGatewayPolicyTier{
		{Targets: []aiGatewayPolicyTarget{policyTarget("prod-us", "gpt-5.2-codex")}},
		{Targets: []aiGatewayPolicyTarget{policyTarget("", "claude-sonnet-5*"), policyTarget("*", "gpt-5*")}},
	}

	t.Run("same tiers in another order keep the configured order", func(t *testing.T) {
		state := aiGatewayPolicyModel{Inbound: types.StringValue("coding-agent"), UserID: types.StringNull(), Tiers: configured}
		updateAiGatewayPolicyState(&state, policyResponse("coding-agent",
			[][2]string{{"*", "gpt-5*"}, {"", "claude-sonnet-5*"}},
			[][2]string{{"prod-us", "gpt-5.2-codex"}},
		))
		if got := state.Tiers[0].Targets[0].Model.ValueString(); got != "gpt-5.2-codex" {
			t.Errorf("first tier model = %q, want the configured first tier", got)
		}
		if state.ID.ValueString() != "coding-agent" || state.Priority.ValueInt64() != 3 {
			t.Errorf("id = %q, priority = %d", state.ID.ValueString(), state.Priority.ValueInt64())
		}
	})

	t.Run("changed tiers come from the response", func(t *testing.T) {
		state := aiGatewayPolicyModel{Inbound: types.StringValue("coding-agent"), UserID: types.StringNull(), Tiers: configured}
		updateAiGatewayPolicyState(&state, policyResponse("coding-agent",
			[][2]string{{"prod-eu", "gpt-5.2-codex"}},
			[][2]string{{"*", "gpt-5*"}, {"", "claude-sonnet-5*"}},
		))
		if got := state.Tiers[0].Targets[0].Provider.ValueString(); got != "prod-eu" {
			t.Errorf("first tier provider = %q, want prod-eu from the response", got)
		}
		if !state.Tiers[1].Targets[1].Provider.IsNull() {
			t.Errorf("a target without a provider must read back as null")
		}
	})

	t.Run("a user-scoped policy gets a user-scoped ID", func(t *testing.T) {
		const userID = "55555555-5555-5555-5555-555555555555"
		state := aiGatewayPolicyModel{Inbound: types.StringValue("coding-agent"), UserID: types.StringValue(userID)}
		updateAiGatewayPolicyState(&state, policyResponse("coding-agent", [][2]string{{"prod-us", "gpt-5.2-codex"}}))
		if got := state.ID.ValueString(); got != userID+":coding-agent" {
			t.Errorf("id = %q", got)
		}
	})
}

func TestParseAiGatewayPolicyID(t *testing.T) {
	const userID = "55555555-5555-5555-5555-555555555555"
	tests := []struct {
		id, wantUser, wantInbound string
	}{
		{id: "coding-agent", wantInbound: "coding-agent"},
		{id: "gpt-5*", wantInbound: "gpt-5*"},
		{id: userID + ":coding-agent", wantUser: userID, wantInbound: "coding-agent"},
		{id: "team:coding", wantInbound: "team:coding"},
	}
	for _, tt := range tests {
		user, inbound := parseAiGatewayPolicyID(tt.id)
		gotUser := ""
		if user != nil {
			gotUser = user.String()
		}
		if gotUser != tt.wantUser || inbound != tt.wantInbound {
			t.Errorf("parse(%q) = (%q, %q), want (%q, %q)", tt.id, gotUser, inbound, tt.wantUser, tt.wantInbound)
		}
	}
}
