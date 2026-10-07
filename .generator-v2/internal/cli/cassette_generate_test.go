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

// declaredSecret is the write-only password the Twilio fixture declares. It is
// the value a diagnostic must never carry, and it is spelled here so the
// assertion fails loudly if the fixture's example is ever changed.
const declaredSecret = "twilio-basic-auth-secret"

// attributeOf walks the fixture's JSON:API request schema to one attribute, so
// a case can contradict exactly one leaf.
func attributeOf(t *testing.T, op *model.Operation, names ...string) *model.Schema {
	t.Helper()
	if op == nil || op.RequestExamples == nil || op.RequestExamples.Schema == nil {
		t.Fatal("operation carries no request schema")
	}
	current := op.RequestExamples.Schema
	for _, name := range append([]string{"data", "attributes"}, names...) {
		next, ok := current.Properties[name]
		if !ok {
			t.Fatalf("schema has no %q under the walked path", name)
		}
		current = next
	}
	return current
}

// Every stage that can reject a target must name the operation or artifact it
// rejected, because the remedy is always an edit to that operation's
// description and a reader has to know which one.
func TestGenerateCassetteDiagnosticsNameTheirTarget(t *testing.T) {
	cases := []struct {
		name    string
		corrupt func(t *testing.T, r *cassetteRequest)
		wants   []string
	}{
		{
			name: "a group missing its delete",
			corrupt: func(t *testing.T, r *cassetteRequest) {
				r.Operation.ResolvedGroup.Delete = nil
			},
			wants: []string{"create, read and delete"},
		},
		{
			name: "a request example its schema rejects",
			corrupt: func(t *testing.T, r *cassetteRequest) {
				attributeOf(t, r.Operation.ResolvedGroup.Create, "name").
					Enum = []string{"a-name-the-example-does-not-use"}
			},
			wants: []string{"CreateTwilioIntegrationAccount", "data.attributes.name"},
		},
		{
			name: "a node the normalizer could not represent",
			corrupt: func(t *testing.T, r *cassetteRequest) {
				name := attributeOf(t, r.Operation.ResolvedGroup.Create, "name")
				name.Kind = model.SchemaKindUnsupported
				name.UnsupportedReason = "a reason the reader needs"
			},
			wants: []string{"CreateTwilioIntegrationAccount", "a reason the reader needs"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			request := twilioRequest(t, dir)
			c.corrupt(t, &request)

			result := generateCassette(request)
			if result.Status != model.CassetteStatusIneligible {
				t.Fatalf("status = %q, want ineligible", result.Status)
			}
			if len(result.Diagnostics) == 0 {
				t.Fatal("no diagnostic explains the rejection")
			}
			message := result.Diagnostics[0].Message
			for _, want := range c.wants {
				if !strings.Contains(message, want) {
					t.Errorf("diagnostic does not name %q: %s", want, message)
				}
			}
			if result.Diagnostics[0].Severity != model.SeverityWarning {
				t.Errorf("severity = %q, want warning", result.Diagnostics[0].Severity)
			}
		})
	}
}

// A declared secret must not reach a diagnostic by any route.
//
// This holds for a stronger reason than the assertion implies, and the reason
// is worth recording: the materializer replaces a write-only secret with
// RedactedPlaceholder before validation or the CLI ever sees the value, so the
// declared secret does not exist downstream of materialization. Verified by
// removing both of ValidateSet's sensitive skips and making its message quote
// the value — what leaks is "[redacted]", never the password.
//
// So this test cannot fail while replacement holds, and it is kept as a guard
// on that property rather than on anything these layers do. Note also that
// model.Redact, documented as the last line of defense before a message
// reaches the run report, is never called by anything.
func TestGenerateCassetteDiagnosticsNeverCarryADeclaredSecret(t *testing.T) {
	corruptions := map[string]func(t *testing.T, r *cassetteRequest){
		"a group missing its delete": func(t *testing.T, r *cassetteRequest) {
			r.Operation.ResolvedGroup.Delete = nil
		},
		"a sibling attribute its schema rejects": func(t *testing.T, r *cassetteRequest) {
			attributeOf(t, r.Operation.ResolvedGroup.Create, "name").
				Enum = []string{"a-name-the-example-does-not-use"}
		},
	}

	for name, corrupt := range corruptions {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			request := twilioRequest(t, dir)
			corrupt(t, &request)

			result := generateCassette(request)
			for _, d := range result.Diagnostics {
				if strings.Contains(d.Message, declaredSecret) {
					t.Errorf("diagnostic carries the declared secret: %s", d.Message)
				}
			}
		})
	}
}

// A write-only secret cannot itself produce a conformance violation, which is
// what keeps its value out of a diagnostic rather than relying on redaction
// after the fact: the materializer replaces it, validation skips the
// replacement, so no message is ever built from it.
//
// The fixture declares authentication as a oneOf, and the password sits
// directly on a variant — the twilio_integration_account_basic_auth name in
// the generated HCL is the emitter's, not the schema's.
func TestGenerateCassetteSecretCannotProduceAViolation(t *testing.T) {
	dir := t.TempDir()
	request := twilioRequest(t, dir)

	auth := attributeOf(t, request.Operation.ResolvedGroup.Create, "authentication")
	if auth.OneOf == nil || len(auth.OneOf.Variants) == 0 {
		t.Fatalf("authentication is %q with no oneOf variants; fixture shape changed", auth.Kind)
	}
	var password *model.Schema
	for _, variant := range auth.OneOf.Variants {
		if variant.Schema == nil {
			continue
		}
		if candidate, ok := variant.Schema.Properties["password"]; ok {
			password = candidate
			break
		}
	}
	if password == nil {
		t.Fatal("no oneOf variant declares a password; fixture shape changed")
	}
	if !password.WriteOnlySecret {
		t.Fatalf("the password is not classified write-only, so this test is not exercising a secret")
	}

	// Contradict the secret's own schema. Were sensitive leaves checked, this
	// would be a violation naming the path whose value is the password.
	password.Enum = []string{"a-password-the-example-does-not-use"}

	result := generateCassette(request)
	if result.Status != model.CassetteStatusGenerated {
		t.Fatalf("a contradicted secret made the target ineligible: %q, %v",
			result.Status, result.Diagnostics)
	}
	for _, d := range result.Diagnostics {
		if strings.Contains(d.Message, declaredSecret) {
			t.Errorf("diagnostic carries the declared secret: %s", d.Message)
		}
	}
}

// A target that fails any stage must leave nothing behind. The write is last,
// so this is really asserting that no earlier stage writes as a side effect.
func TestGenerateCassetteWritesNothingForAnyRejection(t *testing.T) {
	corruptions := map[string]func(t *testing.T, r *cassetteRequest){
		"a group missing its delete": func(t *testing.T, r *cassetteRequest) {
			r.Operation.ResolvedGroup.Delete = nil
		},
		"a request example its schema rejects": func(t *testing.T, r *cassetteRequest) {
			attributeOf(t, r.Operation.ResolvedGroup.Create, "name").
				Enum = []string{"a-name-the-example-does-not-use"}
		},
		"an unrepresentable node": func(t *testing.T, r *cassetteRequest) {
			name := attributeOf(t, r.Operation.ResolvedGroup.Create, "name")
			name.Kind = model.SchemaKindUnsupported
			name.UnsupportedReason = "unrepresentable"
		},
	}

	for name, corrupt := range corruptions {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			request := twilioRequest(t, dir)
			corrupt(t, &request)

			result := generateCassette(request)
			if result.WriteAction != model.CassetteWriteNone {
				t.Errorf("write action = %q, want none", result.WriteAction)
			}
			if result.TestPath != "" {
				if _, err := os.Stat(result.TestPath); err == nil {
					t.Errorf("a rejected target wrote %s", result.TestPath)
				}
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatalf("reading the output root: %v", err)
			}
			if len(entries) > 0 {
				var names []string
				for _, e := range entries {
					names = append(names, e.Name())
				}
				t.Errorf("a rejected target left %v behind", names)
			}
		})
	}
}

// Each rejection stage must say which stage it was and where to look. The
// category distinguishes a description that cannot be selected from one whose
// values its own schema rejects, which call for different edits; the location
// is the anchor a reader opens.
func TestGenerateCassetteDiagnosticsCarryStageAndAnchor(t *testing.T) {
	cases := []struct {
		name         string
		corrupt      func(t *testing.T, r *cassetteRequest)
		wantCategory model.DiagnosticCategory
		wantAnchor   []string
	}{
		{
			name: "a group missing its delete",
			corrupt: func(t *testing.T, r *cassetteRequest) {
				r.Operation.ResolvedGroup.Delete = nil
			},
			wantCategory: model.DiagnosticCategoryEligibility,
			wantAnchor:   []string{"spec:", "/api/v2/integration-interfaces/twilio/accounts"},
		},
		{
			name: "a request example its schema rejects",
			corrupt: func(t *testing.T, r *cassetteRequest) {
				attributeOf(t, r.Operation.ResolvedGroup.Create, "name").
					Enum = []string{"a-name-the-example-does-not-use"}
			},
			wantCategory: model.DiagnosticCategoryValidation,
			// Anchored at the set that failed, not the tracked operation.
			wantAnchor: []string{"spec:", ".requestBody"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			request := twilioRequest(t, dir)
			c.corrupt(t, &request)

			result := generateCassette(request)
			if len(result.Diagnostics) == 0 {
				t.Fatal("no diagnostic explains the rejection")
			}
			d := result.Diagnostics[0]
			if d.Category != c.wantCategory {
				t.Errorf("category = %q, want %q", d.Category, c.wantCategory)
			}
			if d.Location == "" {
				t.Fatal("diagnostic carries no location anchor")
			}
			for _, want := range c.wantAnchor {
				if !strings.Contains(d.Location, want) {
					t.Errorf("location %q does not contain %q", d.Location, want)
				}
			}
			// An anchor is a spec coordinate and must never be a value.
			if strings.Contains(d.Location, declaredSecret) {
				t.Errorf("location carries the declared secret: %s", d.Location)
			}
		})
	}
}

// A validation failure anchors at the operation that actually failed, which
// need not be the tracked one: the update's request is a different operation
// from the create the annotation hangs on.
func TestGenerateCassetteAnchorsAtTheFailingOperation(t *testing.T) {
	dir := t.TempDir()
	request := twilioRequest(t, dir)
	attributeOf(t, request.Operation.ResolvedGroup.Update, "name").
		Enum = []string{"a-name-the-update-example-does-not-use"}

	result := generateCassette(request)
	if result.Status != model.CassetteStatusIneligible {
		t.Fatalf("status = %q, want ineligible", result.Status)
	}
	d := result.Diagnostics[0]
	if !strings.Contains(d.Message, "UpdateTwilioIntegrationAccount") {
		t.Errorf("message does not name the failing operation: %s", d.Message)
	}
	if !strings.Contains(d.Location, "{account_id}") {
		t.Errorf("location %q is not the update operation's path", d.Location)
	}
	if d.Category != model.DiagnosticCategoryValidation {
		t.Errorf("category = %q, want validation", d.Category)
	}
}
