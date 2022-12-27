package utils

import (
	"testing"
)

func TestTagNormalization(t *testing.T) {
	cases := map[string]string{
		"foo":     "foo",
		"Foo":     "foo",
		"1foo":    "foo",
		"foo_bar": "foo_bar",
	}
	for tag, expected_tag := range cases {
		normalized := NormalizeTag(tag)
		if normalized != expected_tag {
			t.Errorf("Expected tag '%s' normalized to '%s', got '%s' instead.", tag, expected_tag, normalized)
		}
	}
}
