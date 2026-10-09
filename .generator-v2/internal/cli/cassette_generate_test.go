package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// twilioRequest loads the coherent fixture and pairs it with the minimum view
// the orchestrator needs: the schema attributes carrying materialized values,
// and the SDK read call the destroy check uses. A fuller view is emit's
// concern; this exercises the chain.
func twilioRequest(t *testing.T, dir string) cassetteRequest {
	t.Helper()
	spec, err := parser.LoadSpec(filepath.Join("..", "testdata", "mini-oas",
		"mini-datadog_integration_twilio_account.yaml"))
	if err != nil {
		t.Fatalf("loading fixture: %v", err)
	}
	var tracked *model.Operation
	for _, op := range spec.Operations {
		if op.Tracking != nil && op.Tracking.ArtifactName == "integration_twilio_account" {
			tracked = op
		}
	}
	if tracked == nil {
		t.Fatal("fixture has no tracked twilio resource")
	}

	const attrs = "data.attributes"
	return cassetteRequest{
		// The correspondence the real pipeline records from the artifact.
		APIPaths: map[string]string{
			"name":                 attrs + ".name",
			"settings.account_sid": attrs + ".settings.account_sid",
		},
		Operation: tracked,
		ServerURL: spec.ServerURL,
		View: emit.ResourceView{
			TypeName:    "datadog_integration_twilio_account",
			GoName:      "IntegrationTwilioAccount",
			SDKPackage:  "datadogV2",
			APIAccessor: "GetTwilioIntegrationApiV2",
			Read: emit.CRUDCallView{
				Method:    "GetTwilioIntegrationAccount",
				Arguments: []emit.SDKArgumentView{{Expression: "state.Id.ValueString()", TFName: "id"}},
			},
			Schema: emit.SchemaView{Attributes: []emit.AttrView{
				{TFName: "name",
					TFType: "schema.StringAttribute", Required: true},
				{TFName: "settings", IsBlock: true, Optional: true,
					Attributes: []emit.AttrView{
						{TFName: "account_sid",
							TFType: "schema.StringAttribute", Required: true},
					}},
			}},
		},
		TestsOutputRoot: dir,
	}
}

// The chain produces the test, and only the test. A fixture built from the
// description would assert what the description claims rather than what the
// API does, so it is recorded, not generated.
func TestGenerateCassetteWritesOnlyTheTest(t *testing.T) {
	dir := t.TempDir()
	result := generateCassette(twilioRequest(t, dir))
	if result.Status != model.CassetteStatusGenerated {
		t.Fatalf("status = %q; diagnostics: %v", result.Status, result.Diagnostics)
	}
	if _, err := os.Stat(result.TestPath); err != nil {
		t.Fatalf("the generated test is absent: %v", err)
	}
	// Nothing here counts interactions, because nothing here writes any. A
	// recording of this fixture contains seven: create, its post-apply read and
	// the next step's pre-apply refresh, update, that step's one read, delete,
	// and the destroy verification's 404.
	if _, err := os.Stat(filepath.Join(dir, "tests", "cassettes")); !os.IsNotExist(err) {
		t.Error("a fixture was generated; cassettes must come from a recording")
	}
}

// A rerun over tfgen's own test file is unchanged.
func TestGenerateCassetteRerunIsUnchanged(t *testing.T) {
	dir := t.TempDir()
	if first := generateCassette(twilioRequest(t, dir)); first.Status != model.CassetteStatusGenerated {
		t.Fatalf("first run: %q", first.Status)
	}
	second := generateCassette(twilioRequest(t, dir))
	if second.WriteAction != model.CassetteWriteUnchanged {
		t.Errorf("write action = %q, want unchanged", second.WriteAction)
	}
}

// Check mode reports what the chain produced without touching the tree.
func TestGenerateCassetteCheckModeWritesNothing(t *testing.T) {
	dir := t.TempDir()
	request := twilioRequest(t, dir)
	request.Check = true

	result := generateCassette(request)
	if result.Status != model.CassetteStatusGenerated {
		t.Fatalf("status = %q; diagnostics: %v", result.Status, result.Diagnostics)
	}
	if result.WriteAction != model.CassetteWriteCreated {
		t.Errorf("write action = %q, want created in check mode", result.WriteAction)
	}
	if _, err := os.Stat(result.TestPath); !os.IsNotExist(err) {
		t.Error("check mode wrote the test file")
	}
}

// One incomplete description must not stop other artifacts generating, so
// every failure is a reported ineligibility rather than a run failure.
func TestGenerateCassetteReportsIneligibility(t *testing.T) {
	dir := t.TempDir()

	t.Run("a group missing a delete operation", func(t *testing.T) {
		// twilioRequest reloads the spec, so each subtest gets its own
		// operation tree and this mutation cannot leak.
		request := twilioRequest(t, dir)
		request.Operation.ResolvedGroup.Delete = nil

		result := generateCassette(request)
		if result.Status != model.CassetteStatusIneligible {
			t.Fatalf("status = %q, want ineligible", result.Status)
		}
		if len(result.Diagnostics) == 0 ||
			!strings.Contains(result.Diagnostics[0].Message, "create, read and delete") {
			t.Errorf("diagnostics do not name the missing role: %v", result.Diagnostics)
		}
		if result.Diagnostics[0].Severity != model.SeverityWarning {
			t.Errorf("severity = %q, want warning", result.Diagnostics[0].Severity)
		}
	})
}

func TestLifecycleOperationsDeduplicates(t *testing.T) {
	// A minimal annotation may name one operationId for several roles, and
	// selection must not see it twice.
	shared := &model.Operation{OperationId: "GetThing"}
	other := &model.Operation{OperationId: "DeleteThing"}
	got := lifecycleOperations(&model.ResolvedGroup{
		Create: shared, Read: shared, Update: nil, Delete: other,
	})
	if len(got) != 2 || got[0] != shared || got[1] != other {
		t.Fatalf("got %d operations, want the shared one once then the delete", len(got))
	}
}

// A validation failure must reach the report as an ineligibility and leave the
// tree alone: the write is the last stage, and a target that fails any earlier
// one never reaches it.
func TestGenerateCassetteSuppressesTheWriteOnAViolation(t *testing.T) {
	dir := t.TempDir()
	request := twilioRequest(t, dir)

	// Narrow one attribute's enum so the example the fixture already declares
	// no longer satisfies it, exactly as a self-contradicting description
	// would.
	create := request.Operation.ResolvedGroup.Create
	if create == nil || create.RequestExamples == nil || create.RequestExamples.Schema == nil {
		t.Fatal("fixture create operation carries no request schema")
	}
	data, ok := create.RequestExamples.Schema.Properties["data"]
	if !ok {
		t.Fatal("fixture request schema has no data member")
	}
	attributes, ok := data.Properties["attributes"]
	if !ok {
		t.Fatal("fixture request schema has no data.attributes")
	}
	name, ok := attributes.Properties["name"]
	if !ok {
		t.Fatal("fixture declares no name attribute")
	}
	name.Enum = []string{"a-name-the-example-does-not-use"}

	result := generateCassette(request)

	if result.Status != model.CassetteStatusIneligible {
		t.Fatalf("status = %q, want ineligible", result.Status)
	}
	if result.WriteAction != model.CassetteWriteNone {
		t.Errorf("write action = %q, want none", result.WriteAction)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("no diagnostic explains the violation")
	}
	message := result.Diagnostics[0].Message
	for _, want := range []string{"CreateTwilioIntegrationAccount", "data.attributes.name"} {
		if !strings.Contains(message, want) {
			t.Errorf("diagnostic does not name %q: %s", want, message)
		}
	}
	// The offending value could as easily have been a credential.
	if strings.Contains(message, "twilio-prod") {
		t.Errorf("diagnostic quotes the offending value: %s", message)
	}
	// Nothing was written, including no test file.
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("a violating target wrote %v", names)
	}
}
