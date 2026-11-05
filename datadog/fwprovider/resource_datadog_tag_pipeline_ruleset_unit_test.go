package fwprovider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestIfTagExistsFromIfNotExists(t *testing.T) {
	cases := []struct {
		name             string
		ifNotExists      bool
		fallbackForFalse string
		expected         string
	}{
		{"query true", true, ifTagExistsAppend, ifTagExistsDoNotApply},
		{"query false", false, ifTagExistsAppend, ifTagExistsAppend},
		{"reference table false", false, ifTagExistsAppend, ifTagExistsAppend},
		{"mapping true", true, ifTagExistsReplace, ifTagExistsDoNotApply},
		{"mapping false", false, ifTagExistsReplace, ifTagExistsReplace},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ifTagExistsFromIfNotExists(tc.ifNotExists, tc.fallbackForFalse))
		})
	}
}

func TestIfNotExistsFromIfTagExists(t *testing.T) {
	cases := []struct {
		ifTagExists string
		expected    bool
		ok          bool
	}{
		{ifTagExistsDoNotApply, true, true},
		{ifTagExistsAppend, false, true},
		{ifTagExistsReplace, false, true},
		{"unknown", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.ifTagExists, func(t *testing.T) {
			value, ok := ifNotExistsFromIfTagExists(tc.ifTagExists)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, value)
		})
	}
}

func TestClearUnconfiguredTagExistsFields(t *testing.T) {
	plan := tagPipelineRulesetModel{
		Rules: []ruleItem{
			{Query: &ruleQuery{IfNotExists: types.BoolValue(true), IfTagExists: types.StringValue(ifTagExistsDoNotApply)}},
			{Mapping: &ruleMapping{IfNotExists: types.BoolValue(false), IfTagExists: types.StringValue(ifTagExistsAppend)}},
			{ReferenceTable: &referenceTable{IfNotExists: types.BoolValue(false), IfTagExists: types.StringValue(ifTagExistsAppend)}},
			{Query: &ruleQuery{IfNotExists: types.BoolValue(false), IfTagExists: types.StringValue(ifTagExistsAppend)}},
		},
	}
	config := tagPipelineRulesetModel{
		Rules: []ruleItem{
			// Only the deprecated field is configured.
			{Query: &ruleQuery{IfNotExists: types.BoolValue(true), IfTagExists: types.StringNull()}},
			// Only if_tag_exists is configured.
			{Mapping: &ruleMapping{IfNotExists: types.BoolNull(), IfTagExists: types.StringValue(ifTagExistsAppend)}},
			{ReferenceTable: &referenceTable{IfNotExists: types.BoolNull(), IfTagExists: types.StringValue(ifTagExistsAppend)}},
		},
	}

	clearUnconfiguredTagExistsFields(&plan, config)

	assert.Equal(t, types.BoolValue(true), plan.Rules[0].Query.IfNotExists)
	assert.True(t, plan.Rules[0].Query.IfTagExists.IsNull())

	assert.True(t, plan.Rules[1].Mapping.IfNotExists.IsNull())
	assert.Equal(t, types.StringValue(ifTagExistsAppend), plan.Rules[1].Mapping.IfTagExists)

	assert.True(t, plan.Rules[2].ReferenceTable.IfNotExists.IsNull())
	assert.Equal(t, types.StringValue(ifTagExistsAppend), plan.Rules[2].ReferenceTable.IfTagExists)

	// Rules beyond the configuration are left alone.
	assert.Equal(t, types.BoolValue(false), plan.Rules[3].Query.IfNotExists)
	assert.Equal(t, types.StringValue(ifTagExistsAppend), plan.Rules[3].Query.IfTagExists)
}
