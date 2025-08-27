package utils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
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

func TestCombineTags(t *testing.T) {
	cases := map[string]struct {
		resourceTags []string
		defaultTags  map[string]string
		expected     []string
	}{
		"basic": {
			[]string{
				"foo:bar", "foo:new",
			}, map[string]string{
				"foo":     "hello",
				"default": "newVal",
			},
			[]string{
				"default:newVal", "foo:bar", "foo:new",
			}},
		"empty default": {
			[]string{
				"foo:bar", "foo:new",
			}, map[string]string{},
			[]string{
				"foo:bar", "foo:new",
			}},
		"empty resource": {
			[]string{}, map[string]string{
				"default": "newVal",
			},
			[]string{
				"default:newVal",
			}},
		"tag without value": {
			[]string{
				"foo",
			}, map[string]string{
				"default": "",
			},
			[]string{
				"default", "foo",
			}},
		"all empty": {
			[]string{},
			map[string]string{},
			[]string{}},
	}
	for _, tc := range cases {
		ctx := context.Background()
		input, _ := types.SetValueFrom(ctx, types.StringType, tc.resourceTags)
		expected, _ := types.SetValueFrom(ctx, types.StringType, tc.expected)
		result, _ := CombineTags(ctx, input, tc.defaultTags)
		if !result.Equal(expected) {
			t.Errorf("Expected: '%s', got '%s' instead.", tc.expected, result)
		}
	}
}
