package utils

import (
	"slices"
	"testing"
)

func TestTagNormalization(t *testing.T) {
	cases := map[string]string{
		"foo":                "foo",
		"FOO":                "foo",
		"1foo":               "foo",
		"foo_":               "foo",
		":foo":               ":foo",
		"foo_bar":            "foo_bar",
		"foo__bar":           "foo_bar",
		"foo123":             "foo123",
		"f!@#$%^&*(),./-=_+": "f_./-",
	}
	for tag, expected_tag := range cases {
		normalized := NormalizeTag(tag)
		if normalized != expected_tag {
			t.Errorf("Expected tag '%s' normalized to '%s', got '%s' instead.", tag, expected_tag, normalized)
		}
	}
}

func TestStripIgnoredTagsNormalization(t *testing.T) {
	cases := []struct {
		name       string
		planTags   []string
		stateTags  []string
		ignoreKeys []string
		expected   []string
	}{
		{
			name:       "empty ignoreKeys returns plan unchanged",
			planTags:   []string{"env:prod", "team:platform"},
			stateTags:  []string{"env:prod", "team:platform"},
			ignoreKeys: []string{},
			expected:   []string{"env:prod", "team:platform"},
		},
		{
			name:       "empty stateTags returns plan unchanged",
			planTags:   []string{"env:prod", "team:platform"},
			stateTags:  []string{},
			ignoreKeys: []string{"team"},
			expected:   []string{"env:prod", "team:platform"},
		},
		{
			name:       "single-valued ignored key pulls state value over plan value",
			planTags:   []string{"env:prod", "domain:web"},
			stateTags:  []string{"env:prod", "domain:eng"},
			ignoreKeys: []string{"domain"},
			expected:   []string{"domain:eng", "env:prod"},
		},
		{
			name:       "multi-valued ignored key preserves every state entry",
			planTags:   []string{"env:prod", "team:web"},
			stateTags:  []string{"env:prod", "team:a", "team:b"},
			ignoreKeys: []string{"team"},
			expected:   []string{"env:prod", "team:a", "team:b"},
		},
		{
			name:       "bare-key tag with no colon",
			planTags:   []string{"production", "env:prod"},
			stateTags:  []string{"staging", "env:prod"},
			ignoreKeys: []string{"production"},
			expected:   []string{"env:prod"},
		},
		{
			name:       "mixed-case ignore key matches lowercase state value",
			planTags:   []string{"domain:web", "env:prod"},
			stateTags:  []string{"domain:eng", "env:prod"},
			ignoreKeys: []string{"Domain"},
			expected:   []string{"domain:eng", "env:prod"},
		},
		{
			name:       "ignored key absent in state drops from output",
			planTags:   []string{"domain:web", "env:prod"},
			stateTags:  []string{"env:prod"},
			ignoreKeys: []string{"domain"},
			expected:   []string{"env:prod"},
		},
		{
			name:       "framework-style raw casing on plan and state still matches normalized ignore key",
			planTags:   []string{"Domain:Web", "Env:Prod"},
			stateTags:  []string{"Domain:Eng", "Env:Prod"},
			ignoreKeys: []string{"domain"},
			expected:   []string{"Domain:Eng", "Env:Prod"},
		},
		{
			name:       "ignore key with special chars normalizes to underscore and matches state",
			planTags:   []string{"foo_bar:wrong", "env:prod"},
			stateTags:  []string{"foo_bar:right", "env:prod"},
			ignoreKeys: []string{"foo!bar"},
			expected:   []string{"env:prod", "foo_bar:right"},
		},
		{
			name:       "ignore key with leading digit is stripped before matching",
			planTags:   []string{"team:wrong", "env:prod"},
			stateTags:  []string{"team:right", "env:prod"},
			ignoreKeys: []string{"1team"},
			expected:   []string{"env:prod", "team:right"},
		},
		{
			name:       "ignore key passed as full key:value pair has value stripped",
			planTags:   []string{"team:wrong", "env:prod"},
			stateTags:  []string{"team:right", "env:prod"},
			ignoreKeys: []string{"team:engineering"},
			expected:   []string{"env:prod", "team:right"},
		},
		{
			name:       "ignored entries preserve state casing not plan casing",
			planTags:   []string{"Domain:Web", "env:prod"},
			stateTags:  []string{"domain:eng", "env:prod"},
			ignoreKeys: []string{"domain"},
			expected:   []string{"domain:eng", "env:prod"},
		},
		{
			name:       "non-ignored key with differing plan and state values keeps plan value so drift still surfaces",
			planTags:   []string{"env:prod", "team:platform"},
			stateTags:  []string{"env:dev", "team:platform"},
			ignoreKeys: []string{"team"},
			expected:   []string{"env:prod", "team:platform"},
		},
		{
			name:       "every plan tag is an ignored key returns the state's ignored entries only",
			planTags:   []string{"team:wrong", "domain:wrong"},
			stateTags:  []string{"team:right", "domain:right", "env:prod"},
			ignoreKeys: []string{"team", "domain"},
			expected:   []string{"domain:right", "team:right"},
		},
		{
			name:       "underscore-collapsing normalization matches across both sides",
			planTags:   []string{"foo_bar:wrong"},
			stateTags:  []string{"foo_bar:right"},
			ignoreKeys: []string{"foo__bar"},
			expected:   []string{"foo_bar:right"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StripIgnoredTags(tc.planTags, tc.stateTags, tc.ignoreKeys)
			if !slices.Equal(got, tc.expected) {
				t.Errorf("Expected %v, got %v instead.", tc.expected, got)
			}
		})
	}
}
