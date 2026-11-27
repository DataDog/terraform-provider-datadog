package main

import (
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/contracts"
)

func TestRegistryReferenceIncludesValidationRules(t *testing.T) {
	doc, err := newSchemaDoc(contracts.FamilyRegistrySchema)
	if err != nil {
		t.Fatal(err)
	}
	var body strings.Builder
	doc.renderBody(&body)
	text := body.String()
	for _, want := range []string{
		"## Registry map", "at least 1 entry", "^[xX]-", "^[a-z][a-z0-9_]*$",
		"max length 64", "header.location", "Default: `data.id`", "`skip` = `true`",
		"[artifact_name](#artifact_name)", "| `id_strategy` | [id_strategy](#id_strategy) | no | `data.id` |",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("reference missing %q", want)
		}
	}
	if count := strings.Count(text, "Constraint: at least one of `resource`"); count != 1 {
		t.Errorf("family constraint rendered %d times, want 1", count)
	}
}
