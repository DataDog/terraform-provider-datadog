package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceDataKey(t *testing.T) {
	// filling this with actual data is pretty much impossible outside of Terraform,
	// so we won't test any actual data access with Get/GetOk/GetWith/GetWithOk
	d := schema.ResourceData{}

	k := NewResourceDataKey(&d, "")
	assertKeyS(t, k, "")

	k.Add("foo")
	k.Add("baz.spam")
	k.Add(1)
	assertKeyS(t, k, "foo.baz.spam.1")

	k.Remove(1)
	k.Remove("spam")
	assertKeyS(t, k, "foo.baz")
	assertKeyWith(t, k, "with.more", "foo.baz.with.more")
	assertKeyWith(t, k, 0, "foo.baz.0")
	assertKeyS(t, k, "foo.baz")

	k.Pop(2)
	assertKeyS(t, k, "")
}

func assertKeyS(t *testing.T, k *ResourceDataKey, s string) {
	result := k.S()
	if result != s {
		t.Errorf("Expected k.S() to be \"%s\", got \"%s\"", s, result)
	}
}

func assertKeyWith(t *testing.T, k *ResourceDataKey, additionalParts interface{}, s string) {
	result := k.With(additionalParts)
	if result != s {
		t.Errorf("Expected k.With(\"%s\") to be \"%s\", got \"%s\"", additionalParts, s, result)
	}
}
