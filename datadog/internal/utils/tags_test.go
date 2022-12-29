package utils

import (
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
