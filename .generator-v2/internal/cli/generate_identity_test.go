//go:build integration

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	testinfra "github.com/terraform-providers/terraform-provider-datadog/generator/internal/testdata"
)

func TestGenerateCallerAssignedIdentity(t *testing.T) {
	fixture := writeOnlyResourceFixture{name: "Derived metric", spec: "../fixtures/resource_derived_metric/openapi.yaml", resource: "resource_datadog_derived_metric.go"}
	root := t.TempDir()
	firstPath, first := generateWriteOnlyFixture(t, root, fixture, "first")
	_, second := generateWriteOnlyFixture(t, root, fixture, "second")
	if !bytes.Equal(first, second) {
		t.Fatal("generation is not stable")
	}
	probe := filepath.Join("..", "testdata", "fixtures", "resource_derived_metric", "identity_test.go.txt")
	if _, err := os.Stat(probe); err != nil {
		t.Fatal(err)
	}
	result, err := testinfra.RunTests([]testinfra.StagedFile{
		{ProviderPath: filepath.Join("datadog", "fwprovider", fixture.resource), SourcePath: firstPath},
		{ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_identity_test.go"), SourcePath: probe},
	}, "^TestDerivedMetricCallerAssignedIdentity$")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK() {
		t.Fatal(result.Diagnostics)
	}
}
