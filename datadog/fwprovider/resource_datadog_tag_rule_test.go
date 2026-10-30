package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tagRuleConfigValue builds a tag rule config with the given patterns and
// flags. A nil flag is left out of the config, as when the user omits it.
func tagRuleConfigValue(t *testing.T, patterns tftypes.Value, required, negated interface{}) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	NewTagRuleResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError(), "schema should not produce diagnostics")

	objectType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attrType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["name"] = tftypes.NewValue(tftypes.String, "env must be set")
	values["source"] = tftypes.NewValue(tftypes.String, "logs")
	values["scope"] = tftypes.NewValue(tftypes.String, "org")
	values["tag_key"] = tftypes.NewValue(tftypes.String, "env")
	values["rule_type"] = tftypes.NewValue(tftypes.String, "surfacing")
	values["tag_value_patterns"] = patterns
	values["required"] = tftypes.NewValue(tftypes.Bool, required)
	values["negated"] = tftypes.NewValue(tftypes.Bool, negated)

	return tfsdk.Config{
		Raw:    tftypes.NewValue(objectType, values),
		Schema: schemaResp.Schema,
	}
}

func tagRulePatterns(patterns ...string) tftypes.Value {
	elements := make([]tftypes.Value, len(patterns))
	for i, pattern := range patterns {
		elements[i] = tftypes.NewValue(tftypes.String, pattern)
	}
	return tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, elements)
}

func validateTagRuleConfig(t *testing.T, config tfsdk.Config) *resource.ValidateConfigResponse {
	t.Helper()
	resp := &resource.ValidateConfigResponse{}
	(&tagRuleResource{}).ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: config}, resp)
	return resp
}

func TestTagRuleValidateConfig_AcceptsWildcardCombinationsTheAPIAccepts(t *testing.T) {
	cases := map[string]struct {
		patterns          tftypes.Value
		required, negated interface{}
	}{
		"required tag with any value":          {tagRulePatterns("*"), true, false},
		"required tag, negated omitted":        {tagRulePatterns("*"), true, nil},
		"tag must be absent":                   {tagRulePatterns("*"), false, true},
		"tag must be absent, required omitted": {tagRulePatterns("*"), nil, true},
		"specific values, both flags false":    {tagRulePatterns("prod", "staging"), false, false},
		"prefix wildcard, both flags false":    {tagRulePatterns("web-*"), false, false},
		"specific values, both flags true":     {tagRulePatterns("prod"), true, true},
		"unknown patterns":                     {tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue), false, false},
		"unknown required":                     {tagRulePatterns("*"), tftypes.UnknownValue, false},
		"unknown pattern element": {
			tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}),
			false, false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resp := validateTagRuleConfig(t, tagRuleConfigValue(t, tc.patterns, tc.required, tc.negated))
			assert.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestTagRuleValidateConfig_RejectsWildcardCombinationsTheAPIRejects(t *testing.T) {
	cases := map[string]struct {
		patterns          tftypes.Value
		required, negated interface{}
		summary           string
	}{
		"both flags false":                {tagRulePatterns("*"), false, false, "Tag rule has no effect"},
		"both flags omitted":              {tagRulePatterns("*"), nil, nil, "Tag rule has no effect"},
		"wildcard alongside other values": {tagRulePatterns("prod", "*"), false, false, "Tag rule has no effect"},
		"repeated wildcard, both false":   {tagRulePatterns("**"), false, false, "Tag rule has no effect"},
		"both flags true":                 {tagRulePatterns("*"), true, true, "Tag rule cannot be satisfied"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resp := validateTagRuleConfig(t, tagRuleConfigValue(t, tc.patterns, tc.required, tc.negated))
			require.True(t, resp.Diagnostics.HasError(), "expected an error diagnostic")
			assert.Equal(t, tc.summary, resp.Diagnostics.Errors()[0].Summary())
		})
	}
}

func TestTagRuleSchema_TagValuePatternsCharsetStillEnforced(t *testing.T) {
	assert.True(t, tagValuePatternCharsetRegex.MatchString("*"), "a bare wildcard is a valid pattern on its own")
	assert.True(t, tagValuePatternCharsetRegex.MatchString("web-*"))
	assert.False(t, tagValuePatternCharsetRegex.MatchString("prod|staging"), "regex syntax is not allowed")
	assert.False(t, tagValuePatternCharsetRegex.MatchString(""), "an empty pattern is not allowed")
}
