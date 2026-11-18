package datadog

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func legacyAgentRuleData(t *testing.T) *schema.ResourceData {
	d := schema.TestResourceDataRaw(t, cloudWorkloadSecurityAgentRuleSchema(), map[string]interface{}{
		"name":       "my_rule",
		"expression": `exec.file.name == "java"`,
	})
	d.SetId("abc-123")
	return d
}

func TestCloudWorkloadSecurityAgentRuleWritesPointToCSMThreats(t *testing.T) {
	ctx := context.Background()
	for name, write := range map[string]func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics{
		"create": resourceDatadogCloudWorkloadSecurityAgentRuleCreate,
		"update": resourceDatadogCloudWorkloadSecurityAgentRuleUpdate,
	} {
		t.Run(name, func(t *testing.T) {
			diags := write(ctx, legacyAgentRuleData(t), nil)
			if assert.True(t, diags.HasError()) {
				assert.Contains(t, diags[0].Summary+diags[0].Detail, "datadog_csm_threats_agent_rule")
			}
		})
	}
}

func TestCloudWorkloadSecurityAgentRuleDeleteOnlyForgetsTheRule(t *testing.T) {
	d := legacyAgentRuleData(t)

	diags := resourceDatadogCloudWorkloadSecurityAgentRuleDelete(context.Background(), d, nil)

	assert.False(t, diags.HasError(), "removing the resource from a config must not fail")
	if assert.Len(t, diags, 1) {
		assert.Equal(t, diag.Warning, diags[0].Severity)
		assert.Contains(t, diags[0].Detail, "datadog_csm_threats_agent_rule")
	}
	assert.Empty(t, d.Id())
}
