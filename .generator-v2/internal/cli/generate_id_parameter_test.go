//go:build integration

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/providermod"
	testinfra "github.com/terraform-providers/terraform-provider-datadog/generator/internal/testdata"
)

// Exercise the maintained fixture through parsing, SDK binding, lifecycle and
// schema construction, source emission, compilation and real local HTTP calls.
func TestGenerateFeatureFlagIDParameter(t *testing.T) {
	providerRoot, err := providermod.Root("")
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join("..", "testdata", "fixtures", "resource_feature_flag")
	first := generateFixture(t, providerRoot, filepath.Join(fixture, "openapi.yaml"))
	second := generateFixture(t, providerRoot, filepath.Join(fixture, "openapi.yaml"))
	name := "resource_datadog_tfgen_feature_flag_fixture.go"
	if !bytes.Equal(mustReadBytes(t, filepath.Join(first, name)), mustReadBytes(t, filepath.Join(second, name))) {
		t.Fatal("feature flag generation is not stable")
	}
	result, err := testinfra.RunTests([]testinfra.StagedFile{
		{ProviderPath: filepath.Join("datadog", "fwprovider", name), SourcePath: filepath.Join(first, name)},
		{ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_feature_flag_id_test.go"), SourcePath: filepath.Join(fixture, "id_parameter_test.go.txt")},
	}, "^TestTFGenFeatureFlagIDParameter$")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK() {
		t.Fatal(result.Diagnostics)
	}
	t.Run("mock rejects an Archive argument unrelated to Create ID", func(t *testing.T) {
		generated := mustReadBytes(t, filepath.Join(first, name))
		correct := []byte("r.Api.ArchiveFeatureFlag(r.Auth, parsedId)")
		if bytes.Count(generated, correct) != 1 {
			t.Fatal("expected exactly one Archive call using the parsed Terraform ID")
		}
		broken := bytes.Replace(generated, correct, []byte("r.Api.ArchiveFeatureFlag(r.Auth, uuid.UUID{parsedId[0]})"), 1)
		brokenPath := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(brokenPath, broken, 0o600); err != nil {
			t.Fatal(err)
		}
		result, err := testinfra.RunTests([]testinfra.StagedFile{
			{ProviderPath: filepath.Join("datadog", "fwprovider", name), SourcePath: brokenPath},
			{ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_feature_flag_id_test.go"), SourcePath: filepath.Join(fixture, "id_parameter_test.go.txt")},
		}, "^TestTFGenFeatureFlagIDParameter$")
		if err != nil {
			t.Fatal(err)
		}
		if result.OK() || !strings.Contains(result.Diagnostics, "from Create response data.id") {
			t.Fatalf("expected the mock's Archive UUID assertion to fail:\n%s", result.Diagnostics)
		}
	})
}
