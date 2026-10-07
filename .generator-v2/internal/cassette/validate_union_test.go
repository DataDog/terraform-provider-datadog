package cassette

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestValidateUsesSelectedUnionSchema(t *testing.T) {
	schema := exampleUnionSchema()
	key := SetKey{Role: SetRoleRequest}
	set, err := MaterializeSet(SelectedSet{Key: key, Candidate: &model.ExampleCandidate{Value: map[string]any{"mode": "token", "secret": "private-example"}}}, schema)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateSet(key, set, schema); err != nil {
		t.Fatalf("selected token branch validated against basic branch: %v", err)
	}
}
