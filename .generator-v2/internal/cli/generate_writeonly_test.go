//go:build integration

package cli

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	testinfra "github.com/terraform-providers/terraform-provider-datadog/generator/internal/testdata"
)

type writeOnlyResourceFixture struct {
	name     string
	spec     string
	resource string
}

var writeOnlyResourceFixtures = []writeOnlyResourceFixture{
	{
		name:     "Twilio",
		spec:     "mini-datadog_integration_twilio_account.yaml",
		resource: "resource_datadog_integration_twilio_account.go",
	},
	{
		name:     "Elastic Cloud",
		spec:     "mini-datadog_integration_elastic_cloud_account.yaml",
		resource: "resource_datadog_integration_elastic_cloud.go",
	},
}

// TestGenerateWriteOnlyResources is the offline full-pipeline gate
// for the canonical generated write-only resources. It deliberately starts at
// the CLI so parser, model, SDK binding, rendering, writing and reporting are
// all covered before the generated code is compiled and its framework schema
// is validated in the provider module.
func TestGenerateWriteOnlyResources(t *testing.T) {
	tempRoot := t.TempDir()
	var firstStaged, secondStaged []testinfra.StagedFile

	for _, fixture := range writeOnlyResourceFixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			firstPath, first := generateWriteOnlyFixture(t, tempRoot, fixture, "first")
			secondPath, second := generateWriteOnlyFixture(t, tempRoot, fixture, "second")
			if !bytes.Equal(first, second) {
				t.Fatalf("two independent generations of the %s resource differ", fixture.name)
			}

			assertGeneratedWriteOnlyContract(t, first)
			firstStaged = append(firstStaged, testinfra.StagedFile{
				ProviderPath: filepath.Join("datadog", "fwprovider", fixture.resource),
				SourcePath:   firstPath,
			})
			secondStaged = append(secondStaged, testinfra.StagedFile{
				ProviderPath: filepath.Join("datadog", "fwprovider", fixture.resource),
				SourcePath:   secondPath,
			})
		})
	}

	probeDir := t.TempDir()
	probePath := filepath.Join(probeDir, "zz_tfgen_writeonly_schema_test.go")
	if err := os.WriteFile(probePath, []byte(writeOnlySchemaProbe), 0o644); err != nil {
		t.Fatalf("write generated schema probe: %v", err)
	}
	probe := testinfra.StagedFile{
		ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_writeonly_schema_test.go"),
		SourcePath:   probePath,
	}
	for _, run := range []struct {
		name   string
		staged []testinfra.StagedFile
	}{
		{name: "first", staged: firstStaged},
		{name: "second", staged: secondStaged},
	} {
		run := run
		t.Run(run.name+" generation compiles and validates", func(t *testing.T) {
			result, err := testinfra.RunTests(append(run.staged, probe), "^TestTfgenWriteOnlySchemasValidate$")
			if err != nil {
				t.Fatalf("run staged provider schema validation: %v", err)
			}
			if !result.OK() {
				t.Fatalf("generated resources do not compile or validate against the pinned provider and SDK:\n%s", result.Diagnostics)
			}
		})
	}
}

func generateWriteOnlyFixture(t *testing.T, tempRoot string, fixture writeOnlyResourceFixture, run string) (string, []byte) {
	t.Helper()
	tempProvider := filepath.Join(tempRoot, strings.ReplaceAll(fixture.name, " ", "_"), run, "provider")
	outputRoot := filepath.Join(tempProvider, "datadog", "fwprovider")
	reportPath := filepath.Join(tempProvider, "tfgen-report.json")
	specPath := filepath.Join("..", "testdata", "mini-oas", fixture.spec)

	if err := runTfgen(
		"generate",
		"--spec", specPath,
		"--output-root", outputRoot,
		"--hooks-root", filepath.Join(outputRoot, "hooks"),
		"--tests-output-root", filepath.Join(tempProvider, "datadog", "tests"),
		"--examples-output-root", filepath.Join(tempProvider, "examples", "resources"),
		"--docs-root", filepath.Join(tempProvider, "docs", "resources"),
		"--report", reportPath,
	); err != nil {
		t.Fatalf("generate %s resource (%s run): %v", fixture.name, run, err)
	}

	assertNoFailedArtifacts(t, reportPath)
	generatedPath := filepath.Join(outputRoot, fixture.resource)
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("read generated %s resource (%s run): %v", fixture.name, run, err)
	}
	return generatedPath, generated
}

func assertNoFailedArtifacts(t *testing.T, reportPath string) {
	t.Helper()
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read generation report: %v", err)
	}
	var report model.RunReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("decode generation report: %v", err)
	}
	if report.Summary == nil {
		t.Fatal("generation report has no summary")
	}
	if report.Summary.Failed != 0 {
		t.Fatalf("generation report contains %d failed artifacts: %s", report.Summary.Failed, reportBytes)
	}
	for _, artifact := range report.Artifacts {
		if artifact.Status == model.ArtifactStatusFailed {
			t.Fatalf("artifact %q failed despite zero summary: %s", artifact.Name, reportBytes)
		}
	}
}

func assertGeneratedWriteOnlyContract(t *testing.T, source []byte) {
	t.Helper()
	text := string(source)

	for _, want := range []string{
		`OriginalAttr: "password"`,
		`WriteOnlyAttr: "password_wo"`,
		`TriggerAttr: "password_wo_version"`,
		`ParentBlocks: []string{"authentication", "integration_account_basic_auth"}`,
		`Mode: fwutils.WriteOnlySecretModeOnly`,
		`fwutils.CreateWriteOnlySecretAttributes(`,
	} {
		if !strings.Contains(text, want) {
			t.Errorf("generated resource is missing write-only contract %q", want)
		}
	}
	if got := strings.Count(text, "`tfsdk:\"password_wo\"`"); got != 1 {
		t.Errorf("password_wo model field count = %d, want 1", got)
	}
	if got := strings.Count(text, "`tfsdk:\"password_wo_version\"`"); got != 1 {
		t.Errorf("password_wo_version model field count = %d, want 1", got)
	}
	for _, forbidden := range []string{
		"`tfsdk:\"password\"`",
		`"password": schema.StringAttribute{`,
		"Sensitive: true",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("generated resource contains forbidden plaintext/sensitive schema form %q", forbidden)
		}
	}

	for _, method := range []struct {
		name       string
		secretCall string
	}{
		{name: "Create", secretCall: "GetSecretForCreate(ctx, &request.Config)"},
		{name: "Update", secretCall: "GetSecretForUpdate(ctx, &request.Config, &request)"},
	} {
		body := generatedMethod(t, source, method.name)
		if !strings.Contains(body, method.secretCall) {
			t.Errorf("%s does not retrieve password through its write-only handler", method.name)
		}
		if got := strings.Count(body, "request.Config"); got != 1 {
			t.Errorf("%s configuration read count = %d, want only the write-only handler read", method.name, got)
		}
		setters := regexp.MustCompile(`SetPassword\(([^)]*)\)`).FindAllStringSubmatch(body, -1)
		if len(setters) != 1 || !strings.HasSuffix(setters[0][1], "PasswordSecretResult.Value") {
			t.Errorf("%s password setters = %v, want exactly one SecretResult.Value setter", method.name, setters)
		}
	}

	updateState := generatedMethod(t, source, "updateState")
	if strings.Contains(updateState, "Password") {
		t.Error("updateState must not assign or otherwise map the write-only password")
	}

	lower := strings.ToLower(text)
	for _, forbidden := range []string{
		"private.set",
		"privatestate",
		"sha256",
		"secret hash",
		"digest",
		"keepers",
	} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("generated resource contains forbidden secret-persistence mechanism %q", forbidden)
		}
	}
}

func generatedMethod(t *testing.T, source []byte, methodName string) string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated.go", source, 0)
	if err != nil {
		t.Fatalf("parse generated resource: %v", err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv == nil || function.Name.Name != methodName {
			continue
		}
		start := fset.Position(function.Pos()).Offset
		end := fset.Position(function.End()).Offset
		return string(source[start:end])
	}
	t.Fatalf("generated resource has no %s method", methodName)
	return ""
}

const writeOnlySchemaProbe = `package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestTfgenWriteOnlySchemasValidate(t *testing.T) {
	constructors := map[string]func() resource.Resource{
		"twilio": NewDatadogIntegrationTwilioAccountResource,
		"elastic_cloud": NewDatadogIntegrationElasticCloudResource,
	}
	for name, constructor := range constructors {
		t.Run(name, func(t *testing.T) {
			var response resource.SchemaResponse
			constructor().Schema(context.Background(), resource.SchemaRequest{}, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("Schema returned errors: %v", response.Diagnostics)
			}

			authentication, ok := response.Schema.Attributes["authentication"].(resourceschema.SingleNestedAttribute)
			if !ok {
				t.Fatalf("authentication is %T, want SingleNestedAttribute", response.Schema.Attributes["authentication"])
			}
			basic, ok := authentication.Attributes["integration_account_basic_auth"].(resourceschema.SingleNestedAttribute)
			if !ok {
				t.Fatalf("integration_account_basic_auth is %T, want SingleNestedAttribute", authentication.Attributes["integration_account_basic_auth"])
			}
			if _, exists := basic.Attributes["password"]; exists {
				t.Fatal("plaintext password attribute is exposed")
			}
			password, ok := basic.Attributes["password_wo"].(resourceschema.StringAttribute)
			if !ok {
				t.Fatalf("password_wo is %T, want StringAttribute", basic.Attributes["password_wo"])
			}
			if !password.WriteOnly || password.Sensitive || !password.Required || password.Optional || password.Computed {
				t.Fatalf("password_wo flags = %#v, want required write-only and not sensitive/optional/computed", password)
			}
			version, ok := basic.Attributes["password_wo_version"].(resourceschema.StringAttribute)
			if !ok {
				t.Fatalf("password_wo_version is %T, want StringAttribute", basic.Attributes["password_wo_version"])
			}
			if version.WriteOnly || version.Sensitive || !version.Required || version.Optional || version.Computed {
				t.Fatalf("password_wo_version flags = %#v, want required stateful trigger", version)
			}

			if diagnostics := response.Schema.ValidateImplementation(context.Background()); diagnostics.HasError() {
				t.Fatalf("Schema.ValidateImplementation returned errors: %v", diagnostics)
			}
		})
	}
}
`
