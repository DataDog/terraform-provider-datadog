package cassette

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Scenario construction
//
// A scenario is what one example-backed target's generated test is rendered
// from: its ordered Terraform steps and the materialized values each step's
// configuration carries.
//
// It describes no HTTP. The cassette the test replays comes from a recording
// run, so there is no trace to keep in agreement with the request the test
// sends — the recording is the agreement.
// ----------------------------------------------------------------------------

// ResourceTarget is everything scenario construction needs about one resource.
type ResourceTarget struct {
	// ArtifactName is the Terraform artifact name without the datadog_ prefix.
	ArtifactName string
	// Create, Read, Update and Delete are the lifecycle operations. Update may
	// be nil; the others are required.
	Create, Read, Update, Delete *model.Operation
	// Selection is the chosen scenario these steps are built from.
	Selection TargetSelection
}

// BuildResourceScenario assembles the Terraform steps for one resource target.
//
// It produces the initial apply, plus an update step only when the examples
// describe a distinct updated state. An update whose response equals the create
// response asserts nothing, so a scenario is not given a step it cannot check.
//
// Read and Delete gate eligibility but contribute no step. The refreshes, the
// delete and the 404 that verifies it are interactions of the recorded
// cassette, provoked by the lifecycle the generated test already runs — this
// builds only what the configuration has to say.
func BuildResourceScenario(target ResourceTarget) (*model.GeneratedTestScenario, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}

	createRequest, createResponse, err := target.materializePair(target.Create)
	if err != nil {
		return nil, err
	}

	scenario := &model.GeneratedTestScenario{
		ArtifactName:     target.ArtifactName,
		ArtifactKind:     model.ArtifactKindResource,
		TestFilePath:     resourceTestFilePath(target.ArtifactName),
		TestFuncName:     resourceTestFuncName(target.ArtifactName),
		TerraformAddress: fmt.Sprintf("datadog_%s.foo", target.ArtifactName),
		Selection:        target.Selection.Model(),
	}

	// The create step's configuration is the create request's own values.
	scenario.Steps = append(scenario.Steps, model.ScenarioStep{State: createRequest.configuration()})

	// A second step only when the examples describe a distinct state to assert.
	// The response is materialized to make that comparison, not to record it.
	updated, hasUpdate, err := target.updateStep(createRequest, createResponse)
	if err != nil {
		return nil, err
	}
	if hasUpdate {
		scenario.Steps = append(scenario.Steps, model.ScenarioStep{State: updated.request.configuration()})
	}

	if err := scenario.Validate(); err != nil {
		return nil, err
	}
	return scenario, nil
}

// validate reports a target that cannot describe a resource lifecycle.
func (t ResourceTarget) validate() error {
	switch {
	case t.ArtifactName == "":
		return fmt.Errorf("scenario: no artifact name")
	case t.Create == nil:
		return fmt.Errorf("scenario %q: no create operation", t.ArtifactName)
	case t.Read == nil:
		return fmt.Errorf("scenario %q: no read operation", t.ArtifactName)
	case t.Delete == nil:
		return fmt.Errorf("scenario %q: no delete operation", t.ArtifactName)
	}
	return nil
}

// materializePair materializes one operation's request and success response.
func (t ResourceTarget) materializePair(op *model.Operation) (MaterializedSet, MaterializedSet, error) {
	request, err := t.materializeRequest(op)
	if err != nil {
		return MaterializedSet{}, MaterializedSet{}, err
	}
	response, err := t.materializeResponse(op)
	if err != nil {
		return MaterializedSet{}, MaterializedSet{}, err
	}
	return request, response, nil
}

func (t ResourceTarget) materializeRequest(op *model.Operation) (MaterializedSet, error) {
	// An operation that declares no request body sends none. Unusual for a
	// create, but legitimate when every input travels in the path or query,
	// and a hard failure here would reject a shape the API permits.
	if op.RequestExamples == nil || !op.RequestExamples.Present {
		return MaterializedSet{}, nil
	}
	key := SetKey{OperationId: op.OperationId, Role: SetRoleRequest}
	set, ok := t.Selection.Set(key)
	if !ok {
		return MaterializedSet{}, fmt.Errorf("scenario %q: no selected request example for %s",
			t.ArtifactName, op.OperationId)
	}
	var schema *model.Schema
	if op.RequestExamples != nil {
		schema = op.RequestExamples.Schema
	}
	return materializeAndValidate(key, set, schema)
}

func (t ResourceTarget) materializeResponse(op *model.Operation) (MaterializedSet, error) {
	success := op.SuccessResponseExample()
	if success == nil {
		return MaterializedSet{}, fmt.Errorf("scenario %q: %s declares no success response",
			t.ArtifactName, op.OperationId)
	}
	key := SetKey{OperationId: op.OperationId, Role: SetRoleResponse, Detail: success.Status}
	set, ok := t.Selection.Set(key)
	if !ok {
		return MaterializedSet{}, fmt.Errorf("scenario %q: no selected response example for %s",
			t.ArtifactName, op.OperationId)
	}
	return materializeAndValidate(key, set, success.Schema)
}

// materializeAndValidate assembles one set and then checks it against the
// schema it was assembled from, which is the only place both are in hand.
//
// Validation runs here rather than over a finished scenario so a target fails
// before any step is built, and therefore long before anything is written. A
// value the description's own schema rejects would otherwise reach a committed
// test and only surface on a recording run against a real org.
func materializeAndValidate(key SetKey, set SelectedSet, schema *model.Schema) (MaterializedSet, error) {
	materialized, err := MaterializeSet(set, schema)
	if err != nil {
		return MaterializedSet{}, err
	}
	if err := ValidateSet(key, materialized, schema); err != nil {
		return MaterializedSet{}, err
	}
	return materialized, nil
}

// updatedState is one update step's materialized request and response.
type updatedState struct {
	request  MaterializedSet
	response MaterializedSet
}

// updateStep materializes an update step, and reports whether the examples
// justify having one. They do only when the update's response differs from the
// create's: an update that changes nothing gives the step nothing to assert.
//
// The returned request is the create request overlaid with the update's own
// values. See MaterializedSet.overlaidWith for why the delta alone cannot
// match what the provider sends.
func (t ResourceTarget) updateStep(createRequest, createResponse MaterializedSet) (updatedState, bool, error) {
	if t.Update == nil {
		return updatedState{}, false, nil
	}
	request, response, err := t.materializePair(t.Update)
	if err != nil {
		// An update the description does not fully describe is not fatal: the
		// scenario simply keeps the create-only flow it can support.
		var incomplete *IncompleteError
		if errors.As(err, &incomplete) {
			return updatedState{}, false, nil
		}
		return updatedState{}, false, err
	}
	if equivalentJSON(response.Body, createResponse.Body) {
		return updatedState{}, false, nil
	}
	// An update whose every value was invented describes no state the author
	// wrote down, so it earns no step: the recorded PATCH would assert values
	// the description never gave. Synthesis exists so a target can be recorded
	// over, not to manufacture a lifecycle nobody described.
	if whollySynthesized(request) || whollySynthesized(response) {
		return updatedState{}, false, nil
	}
	// The recorded request is what the provider will send: the created
	// resource with the update's changes applied, not the example's delta.
	return updatedState{request: createRequest.overlaidWith(request), response: response}, true, nil
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// configuration turns a materialized request into the configuration a test
// step applies. Config rendering stays in the emitter, which owns the mapping
// from these paths to HCL; this carries only the values.
func (m MaterializedSet) configuration() *model.MaterializedConfiguration {
	return &model.MaterializedConfiguration{
		RequestValues:         m.Values,
		SensitiveReplacements: m.SensitiveReplacements,
	}
}

// whollySynthesized reports a set none of whose values the description
// supplied. Compared as a set rather than by count: a synthesized collection
// records both its own path and its elements', so the two lists are not
// one-to-one.
func whollySynthesized(set MaterializedSet) bool {
	if len(set.Values) == 0 {
		// Nothing described, but nothing invented either — a bodyless
		// operation is legitimate.
		return false
	}
	invented := make(map[string]bool, len(set.SynthesizedPaths))
	for _, path := range set.SynthesizedPaths {
		invented[path] = true
	}
	for _, value := range set.Values {
		if !invented[value.Path] {
			return false
		}
	}
	return true
}

// equivalentJSON reports whether two bodies carry the same values, used to tell
// a real update from one that changes nothing.
func equivalentJSON(left, right any) bool {
	return reflect.DeepEqual(left, right)
}

// resourceTestFuncName builds the generated test's name, which doubles as the
// cassette basename because the harness derives the cassette from t.Name().
func resourceTestFuncName(artifact string) string {
	return "TestAccDatadog" + camelCase(artifact) + "OpenAPIExample"
}

func resourceTestFilePath(artifact string) string {
	return fmt.Sprintf("resource_datadog_%s_openapi_example_test.go", artifact)
}

// camelCase converts a snake_case artifact name to UpperCamelCase.
func camelCase(in string) string {
	var out strings.Builder
	for _, part := range strings.Split(in, "_") {
		if part == "" {
			continue
		}
		out.WriteString(strings.ToUpper(part[:1]))
		out.WriteString(part[1:])
	}
	return out.String()
}
