//go:build integration

package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testinfra "github.com/terraform-providers/terraform-provider-datadog/generator/internal/testdata"
)

func TestGenerateCallerAssignedIdentity(t *testing.T) {
	root := t.TempDir()
	fixture := writeOnlyResourceFixture{
		name:     "TFGen identity fixture",
		spec:     writeIdentityFixtureSpec(t, root),
		resource: "resource_datadog_tfgen_identity_fixture.go",
	}
	firstPath, first := generateWriteOnlyFixture(t, root, fixture, "first")
	_, second := generateWriteOnlyFixture(t, root, fixture, "second")
	if !bytes.Equal(first, second) {
		t.Fatal("generation is not stable")
	}
	probe := filepath.Join("..", "testdata", "fixtures", "resource_identity", "identity_test.go.txt")
	if _, err := os.Stat(probe); err != nil {
		t.Fatal(err)
	}
	result, err := testinfra.RunTests([]testinfra.StagedFile{
		// The synthetic resource is additive: no production resource is replaced.
		{ProviderPath: filepath.Join("datadog", "fwprovider", fixture.resource), SourcePath: firstPath},
		{ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_identity_test.go"), SourcePath: probe},
	}, "^TestTFGenCallerAssignedIdentity$")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK() {
		t.Fatal(result.Diagnostics)
	}

	t.Run("rejects computed ID in create request", func(t *testing.T) {
		// Reintroduce the identity bug: the API ID is unknown at creation,
		// so Create must send the configured name instead. Keep the code
		// compilable to prove that the runtime probe detects the regression.
		correct := []byte("bodyData.SetId(state.Name.ValueString())")
		if count := bytes.Count(first, correct); count != 1 {
			t.Fatalf("expected exactly one caller-assigned create ID setter, got %d", count)
		}
		broken := bytes.Replace(first, correct, []byte("bodyData.SetId(state.ID.ValueString())"), 1)
		brokenPath := filepath.Join(t.TempDir(), fixture.resource)
		if err := os.WriteFile(brokenPath, broken, 0o600); err != nil {
			t.Fatal(err)
		}
		// One direct CRUD case is enough to expose the bad POST ID. Skip
		// the second canonical ID and the Terraform import/adoption tests.
		result, err := testinfra.RunTests([]testinfra.StagedFile{
			{ProviderPath: filepath.Join("datadog", "fwprovider", fixture.resource), SourcePath: brokenPath},
			{ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_identity_test.go"), SourcePath: probe},
		}, "^TestTFGenCallerAssignedIdentity$/^assessment[.]metric$/^crud$")
		if err != nil {
			t.Fatal(err)
		}
		if result.OK() {
			t.Fatal("identity probe accepted a create request using the unknown computed ID")
		}
		// Compilation or environment failures must not count as a successful
		// negative test: require the HTTP request assertion's exact diagnostic.
		if !strings.Contains(result.Diagnostics, `POST data.id = ""`) {
			t.Fatalf("expected the caller-assigned create ID assertion to fail:\n%s", result.Diagnostics)
		}
	})
}

// Give the Terraform resource its own identity before generation, while retaining
// real SDK bindings and the required caller-assigned API ID contract. The probe
// routes the SDK to a local HTTP server; no Datadog credentials are needed.
func writeIdentityFixtureSpec(t *testing.T, root string) string {
	t.Helper()
	source := filepath.Join("..", "testdata", "fixtures", "resource_derived_metric", "openapi.yaml")
	contents, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]json.RawMessage
	if err := json.Unmarshal(contents, &spec); err != nil {
		t.Fatal(err)
	}
	var tags []map[string]json.RawMessage
	if err := json.Unmarshal(spec["tags"], &tags); err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected one API tag, got %d", len(tags))
	}
	var resources map[string]json.RawMessage
	if err := json.Unmarshal(tags[0]["x-datadog-tf-generator"], &resources); err != nil {
		t.Fatal(err)
	}
	definition, ok := resources["derived_metric"]
	if !ok || len(resources) != 1 {
		t.Fatal("expected only the derived_metric resource in the source fixture")
	}
	delete(resources, "derived_metric")
	resources["tfgen_identity_fixture"] = definition
	encode := func(value any) json.RawMessage {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	tags[0]["x-datadog-tf-generator"] = encode(resources)
	spec["tags"] = encode(tags)
	path := filepath.Join(root, "identity-fixture.json")
	if err := os.WriteFile(path, encode(spec), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
