package fwprovider

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func routingRuleModelTarget(provider, model string) aiGatewayRoutingRuleTarget {
	return aiGatewayRoutingRuleTarget{
		Provider:       types.StringValue(provider),
		Model:          types.StringValue(model),
		RoutingRule:    types.StringNull(),
		Weight:         types.Int64Null(),
		ReasoningLevel: types.StringNull(),
		Rubric:         types.StringNull(),
		Default:        types.BoolNull(),
	}
}

func routingRuleRuleTarget(slug string) aiGatewayRoutingRuleTarget {
	target := routingRuleModelTarget("", "")
	target.Provider, target.Model, target.RoutingRule = types.StringNull(), types.StringNull(), types.StringValue(slug)
	return target
}

func routingRuleTier(ordering string, targets ...aiGatewayRoutingRuleTarget) aiGatewayRoutingRuleTier {
	return aiGatewayRoutingRuleTier{Ordering: types.StringValue(ordering), Targets: targets}
}

func TestBuildAiGatewayRoutingRuleReplaceRequest(t *testing.T) {
	weighted := routingRuleModelTarget("prod-us", "gpt-5.2-codex")
	weighted.Weight = types.Int64Value(70)
	weighted.ReasoningLevel = types.StringValue("high")

	state := aiGatewayRoutingRuleModel{
		Slug: types.StringValue("coding"),
		Tiers: []aiGatewayRoutingRuleTier{
			routingRuleTier("weighted_affinity", weighted, routingRuleModelTarget("prod-us", "gpt-4o-mini")),
			routingRuleTier("configured", routingRuleRuleTarget("fallback")),
		},
	}

	body, diags := buildAiGatewayRoutingRuleReplaceRequest(&state)
	if diags.HasError() {
		t.Fatalf("build: %v", diags)
	}
	got, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Every tier carries its ordering and an explicit null classifier, so the API doesn't keep saved ones, and no target
	// carries a saved ID, so the API saves exactly these targets.
	want := `{"data":{"attributes":{"tiers":[` +
		`{"classifier":null,"ordering":"weighted_affinity","targets":[` +
		`{"model":"gpt-5.2-codex","provider":"prod-us","reasoning_level":"high","type":"model","weight":70},` +
		`{"model":"gpt-4o-mini","provider":"prod-us","type":"model"}]},` +
		`{"classifier":null,"ordering":"configured","targets":[{"slug":"fallback","type":"routing-rule"}]}]},` +
		`"id":"coding","type":"routing-rules"}}`
	if string(got) != want {
		t.Errorf("request body\n got: %s\nwant: %s", got, want)
	}
}

func TestBuildAiGatewayRoutingRuleClassifiedTier(t *testing.T) {
	billing := routingRuleModelTarget("prod-us", "gpt-5.2")
	billing.Rubric = types.StringValue("Billing questions")
	fallback := routingRuleModelTarget("prod-us", "gpt-4o-mini")
	fallback.Default = types.BoolValue(true)
	tier := routingRuleTier("classified", billing, fallback)
	tier.Classifier = &aiGatewayRoutingRuleClassifier{
		Provider:            types.StringValue("prod-us"),
		Model:               types.StringValue("gpt-6-luna"),
		Context:             types.StringNull(),
		ConfidenceThreshold: types.Float64Value(0.5),
	}

	tiers, diags := buildAiGatewayRoutingRuleTierRequests([]aiGatewayRoutingRuleTier{tier})
	if diags.HasError() {
		t.Fatalf("build: %v", diags)
	}
	got, err := json.Marshal(tiers)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `[{"classifier":{"confidence_threshold":0.5,"model":"gpt-6-luna","provider":"prod-us"},"ordering":"classified","targets":[` +
		`{"model":"gpt-5.2","provider":"prod-us","rubric":"Billing questions","type":"model"},` +
		`{"default":true,"model":"gpt-4o-mini","provider":"prod-us","type":"model"}]}]`
	if string(got) != want {
		t.Errorf("request body\n got: %s\nwant: %s", got, want)
	}
}

func routingRuleResponseTarget(id int64, provider, model string) datadogV2.AiGatewayRoutingRuleTarget {
	target := datadogV2.NewAiGatewayRoutingRuleTarget(datadogV2.AIGATEWAYROUTINGRULETARGETTYPE_MODEL)
	target.SetId(id)
	target.SetProvider(provider)
	target.SetModel(model)
	return *target
}

func routingRuleResponse(tiers ...datadogV2.AiGatewayRoutingRuleTier) datadogV2.AiGatewayRoutingRuleResource {
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	return *datadogV2.NewAiGatewayRoutingRuleResource(
		*datadogV2.NewAiGatewayRoutingRuleAttributes(now, 123, tiers, now), "coding", datadogV2.AIGATEWAYROUTINGRULETYPE_ROUTING_RULES)
}

func TestUpdateAiGatewayRoutingRuleState(t *testing.T) {
	t.Run("API defaults stay unset when the configuration left them unset", func(t *testing.T) {
		codex := routingRuleResponseTarget(10, "prod-us", "gpt-5.2-codex")
		codex.SetWeight(1)
		mini := routingRuleResponseTarget(11, "prod-us", "gpt-4o-mini")
		mini.SetWeight(1)
		state := aiGatewayRoutingRuleModel{Tiers: []aiGatewayRoutingRuleTier{
			routingRuleTier("weighted_affinity", routingRuleModelTarget("prod-us", "gpt-5.2-codex"), routingRuleModelTarget("prod-us", "gpt-4o-mini")),
		}}
		state.Tiers[0].Targets[1].Weight = types.Int64Value(1)

		updateAiGatewayRoutingRuleState(&state, routingRuleResponse(*datadogV2.NewAiGatewayRoutingRuleTier(
			datadogV2.AIGATEWAYROUTINGRULETIERORDERING_WEIGHTED_AFFINITY, []datadogV2.AiGatewayRoutingRuleTarget{codex, mini})))

		if !state.Tiers[0].Targets[0].Weight.IsNull() {
			t.Errorf("weight = %v, want null: the configuration left it unset and the API returned its default", state.Tiers[0].Targets[0].Weight)
		}
		if got := state.Tiers[0].Targets[1].Weight; got.ValueInt64() != 1 {
			t.Errorf("weight = %v, want 1: the configuration set it", got)
		}
		if state.ID.ValueString() != "coding" || state.RouteID.ValueInt64() != 123 {
			t.Errorf("id = %q, route_id = %d", state.ID.ValueString(), state.RouteID.ValueInt64())
		}
	})

	t.Run("import reads a default weight as unset and other weights as set", func(t *testing.T) {
		codex := routingRuleResponseTarget(10, "prod-us", "gpt-5.2-codex")
		codex.SetWeight(70)
		mini := routingRuleResponseTarget(11, "prod-us", "gpt-4o-mini")
		mini.SetWeight(1)
		nested := datadogV2.NewAiGatewayRoutingRuleTarget(datadogV2.AIGATEWAYROUTINGRULETARGETTYPE_ROUTING_RULE)
		nested.SetId(12)
		nested.SetSlug("fallback")
		state := aiGatewayRoutingRuleModel{}

		updateAiGatewayRoutingRuleState(&state, routingRuleResponse(
			*datadogV2.NewAiGatewayRoutingRuleTier(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_WEIGHTED_AFFINITY, []datadogV2.AiGatewayRoutingRuleTarget{codex, mini}),
			*datadogV2.NewAiGatewayRoutingRuleTier(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_CONFIGURED, []datadogV2.AiGatewayRoutingRuleTarget{*nested}),
		))

		if got := state.Tiers[0].Targets[0].Weight; got.ValueInt64() != 70 {
			t.Errorf("weight = %v, want 70", got)
		}
		if got := state.Tiers[0].Targets[1].Weight; !got.IsNull() {
			t.Errorf("weight = %v, want null for the API's default with no prior state", got)
		}
		second := state.Tiers[1].Targets[0]
		if second.RoutingRule.ValueString() != "fallback" || !second.Provider.IsNull() || !second.Model.IsNull() {
			t.Errorf("routing rule target = %+v", second)
		}
	})

	t.Run("a classified tier reads back its classifier, rubrics and default", func(t *testing.T) {
		billing := routingRuleResponseTarget(10, "prod-us", "gpt-5.2")
		billing.SetRubric("Billing questions")
		fallback := routingRuleResponseTarget(11, "prod-us", "gpt-4o-mini")
		fallback.SetDefault(true)
		tier := datadogV2.NewAiGatewayRoutingRuleTier(datadogV2.AIGATEWAYROUTINGRULETIERORDERING_CLASSIFIED, []datadogV2.AiGatewayRoutingRuleTarget{billing, fallback})
		tier.SetClassifier(*datadogV2.NewAiGatewayRoutingClassifier(0.5, "gpt-6-luna", "prod-us"))
		state := aiGatewayRoutingRuleModel{}

		updateAiGatewayRoutingRuleState(&state, routingRuleResponse(*tier))

		got := state.Tiers[0]
		if got.Classifier == nil || got.Classifier.Model.ValueString() != "gpt-6-luna" || !got.Classifier.Context.IsNull() {
			t.Fatalf("classifier = %+v", got.Classifier)
		}
		if got.Targets[0].Rubric.ValueString() != "Billing questions" || !got.Targets[0].Default.IsNull() {
			t.Errorf("first target = %+v", got.Targets[0])
		}
		if !got.Targets[1].Default.ValueBool() || !got.Targets[1].Rubric.IsNull() {
			t.Errorf("default target = %+v", got.Targets[1])
		}
	})

	t.Run("a target removed outside Terraform shows in state", func(t *testing.T) {
		state := aiGatewayRoutingRuleModel{Tiers: []aiGatewayRoutingRuleTier{
			routingRuleTier("configured", routingRuleModelTarget("prod-us", "gpt-5.2-codex"), routingRuleModelTarget("prod-us", "gpt-4o-mini")),
		}}
		updateAiGatewayRoutingRuleState(&state, routingRuleResponse(*datadogV2.NewAiGatewayRoutingRuleTier(
			datadogV2.AIGATEWAYROUTINGRULETIERORDERING_CONFIGURED, []datadogV2.AiGatewayRoutingRuleTarget{routingRuleResponseTarget(10, "prod-us", "gpt-4o-mini")})))
		if len(state.Tiers[0].Targets) != 1 || state.Tiers[0].Targets[0].Model.ValueString() != "gpt-4o-mini" {
			t.Errorf("targets = %+v, want only the target the server still holds", state.Tiers[0].Targets)
		}
	})
}
