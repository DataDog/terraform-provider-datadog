package fwutils

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

func TestApplyIgnoreTagKeys(t *testing.T) {
	cases := map[string]struct {
		planTags    []string
		priorTags   []string
		ignoreKeys  []string // nil means the attribute is unset (null)
		ignoreUnset bool
		priorUnset  bool // true means a zero-value (typeless) state Set, as on create
		expected    []string
	}{
		"unset ignore_tag_keys is a passthrough": {
			planTags:    []string{"a:1", "b:2"},
			priorTags:   []string{"a:1", "b:9"},
			ignoreUnset: true,
			expected:    []string{"a:1", "b:2"},
		},
		"re-injects the prior value of an ignored key": {
			planTags:   []string{"a:1", "test:wrong"},
			priorTags:  []string{"a:1", "test:right"},
			ignoreKeys: []string{"test"},
			expected:   []string{"a:1", "test:right"},
		},
		"create has no prior value to re-inject": {
			planTags:   []string{"a:1"},
			priorTags:  []string{},
			ignoreKeys: []string{"test"},
			expected:   []string{"a:1"},
		},
		"zero-value state Set on create does not panic": {
			planTags:   []string{"a:1", "test:set"},
			priorUnset: true,
			ignoreKeys: []string{"test"},
			expected:   []string{"a:1", "test:set"},
		},
		"non-ignored keys keep their planned values": {
			planTags:   []string{"a:1", "b:new", "test:wrong"},
			priorTags:  []string{"a:1", "b:old", "test:right"},
			ignoreKeys: []string{"test"},
			expected:   []string{"a:1", "b:new", "test:right"},
		},
	}
	for name, tc := range cases {
		ctx := context.Background()
		planTags, _ := types.SetValueFrom(ctx, types.StringType, tc.planTags)
		priorTags := types.Set{} // zero value, as a never-set state field is on create
		if !tc.priorUnset {
			priorTags, _ = types.SetValueFrom(ctx, types.StringType, tc.priorTags)
		}
		ignoreKeys := types.SetNull(types.StringType)
		if !tc.ignoreUnset {
			ignoreKeys, _ = types.SetValueFrom(ctx, types.StringType, tc.ignoreKeys)
		}
		expected, _ := types.SetValueFrom(ctx, types.StringType, tc.expected)
		result, diags := ApplyIgnoreTagKeys(ctx, planTags, priorTags, ignoreKeys)
		if diags.HasError() {
			t.Errorf("%s: unexpected diagnostics: %v", name, diags)
		}
		if !result.Equal(expected) {
			t.Errorf("%s: expected '%s', got '%s' instead.", name, expected, result)
		}
	}
}
