package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/cassette"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Cassette orchestration
//
// Runs one artifact through the whole chain — select, materialize, build and
// validate the scenario, render the test and the fixture, write the bundle —
// and reports the outcome.
//
// Nothing here fails the run. A target whose examples cannot form a coherent
// fixture is reported ineligible with the reason attached, because one
// incomplete description must not stop every other artifact from generating.
// ----------------------------------------------------------------------------

// cassetteRequest is everything generating one artifact's bundle needs. The
// run-level fields are filled once when the flag is read and the per-artifact
// ones per artifact, so a nil *cassetteRequest is what makes "the feature is
// off" a single nil check rather than a flag threaded through every path.
//
// The view is taken rather than rebuilt because emit.BuildResourceView depends
// on SDK bindings having been resolved; cassette generation therefore has to
// run after that step, not alongside artifact building.
type cassetteRequest struct {
	Operation       *model.Operation
	View            emit.ResourceView
	APIPaths        map[string]string
	ServerURL       string
	TestsOutputRoot string
	// Check suppresses the write, so a read-only run reports what it would
	// have done without touching the tree.
	Check bool
}

// generateCassette runs one artifact through the chain and reports the result.
func generateCassette(request cassetteRequest) model.CassetteResult {
	op := request.Operation
	result := model.CassetteResult{
		Name:        op.Tracking.ArtifactName,
		Kind:        model.ArtifactKindResource,
		Status:      model.CassetteStatusIneligible,
		WriteAction: model.CassetteWriteNone,
	}

	group := op.ResolvedGroup
	if group == nil || group.Create == nil || group.Read == nil || group.Delete == nil {
		return ineligible(result, model.DiagnosticCategoryEligibility, operationLocation(op),
			"a resource cassette needs a create, read and delete operation; "+
				"the tracking group resolves fewer than that")
	}

	selection, err := cassette.Select(lifecycleOperations(group))
	if err != nil {
		return ineligible(result, model.DiagnosticCategorySelection, operationLocation(op), err.Error())
	}
	result.SelectedExample = selection.ScenarioName

	scenario, err := cassette.BuildResourceScenario(cassette.ResourceTarget{
		ArtifactName: op.Tracking.ArtifactName,
		Create:       group.Create,
		Read:         group.Read,
		Update:       group.Update,
		Delete:       group.Delete,
		Selection:    selection,
	})
	if err != nil {
		// Scenario building materializes and validates, so the stage is read
		// off the error rather than assumed: a conformance failure anchors at
		// the set that failed, which is more use than the artifact.
		category, location := scenarioFailureAnchor(err, group, op)
		return ineligible(result, category, location, err.Error())
	}
	result.TestName = scenario.TestFuncName

	source, err := emit.RenderResourceExampleTest(scenario, request.View, request.APIPaths)
	if err != nil {
		return ineligible(result, model.DiagnosticCategoryRender, operationLocation(op), err.Error())
	}
	result.TestPath = filepath.Join(request.TestsOutputRoot, scenario.TestFilePath)

	// Only the test is written. The cassette and its freeze companion come from
	// a recording run: a fixture built from the description asserts what the
	// description claims, which is not evidence the API behaved that way, and a
	// synthesized one asserts values nobody wrote down at all. So the generator
	// produces the test and the configuration, and RECORD=true produces the
	// fixture — go-vcr writes both files itself, under the names the harness
	// derives from the test's own name.
	status, err := emit.WriteArtifactSource(result.TestPath, source, request.Check, false)
	if err != nil {
		return ineligible(result, model.DiagnosticCategoryWrite, operationLocation(op), err.Error())
	}
	result.WriteAction = testWriteAction(status)
	result.Status = model.CassetteStatusGenerated
	return result
}

// testWriteAction maps an artifact write onto the cassette result's own
// vocabulary, so the run report says what happened to the generated test.
func testWriteAction(status model.ArtifactStatus) model.CassetteWriteAction {
	switch status {
	case model.ArtifactStatusCreated:
		return model.CassetteWriteCreated
	case model.ArtifactStatusUnchanged:
		return model.CassetteWriteUnchanged
	case model.ArtifactStatusUpdated:
		return model.CassetteWriteUpdated
	default:
		return model.CassetteWriteNone
	}
}

// lifecycleOperations returns the group's operations in lifecycle order,
// deduplicated: a minimal annotation may name one operationId for several
// roles, and selection must not see it twice.
func lifecycleOperations(group *model.ResolvedGroup) []*model.Operation {
	var out []*model.Operation
	for _, op := range []*model.Operation{group.Create, group.Read, group.Update, group.Delete} {
		if op == nil {
			continue
		}
		// A description may name one operation in two roles — a single GET as
		// both create and read — and a replayed interaction is consumed once.
		if !slices.Contains(out, op) {
			out = append(out, op)
		}
	}
	return out
}

// ineligible attaches the reason a target could not produce a fixture, the
// stage that decided so, and the anchor a reader should open.
//
// The message is what a description author has to act on, so it is carried
// rather than reduced to a status. The category says which stage rejected the
// target, because "cannot be selected" and "contradicts its own schema" call
// for different edits. The location is the spec anchor; it is never a value.
func ineligible(
	result model.CassetteResult,
	category model.DiagnosticCategory,
	location model.ExampleLocation,
	reason string,
) model.CassetteResult {
	result.Status = model.CassetteStatusIneligible
	result.WriteAction = model.CassetteWriteNone
	diagnostic := model.NewCassetteDiagnostic(
		model.SeverityWarning,
		fmt.Sprintf("no cassette generated: %s", reason),
		location,
	)
	diagnostic.Category = category
	result.Diagnostics = append(result.Diagnostics, diagnostic)
	return result
}

// operationLocation anchors at the tracked operation itself, for the stages
// that fail before any one example is implicated.
func operationLocation(op *model.Operation) model.ExampleLocation {
	if op == nil {
		return model.ExampleLocation{}
	}
	return model.ExampleLocation{
		Path:        op.Path,
		Method:      op.Method,
		OperationId: op.OperationId,
	}
}

// scenarioFailureAnchor reads the stage and the anchor off a scenario-building
// failure. Materialization and validation both happen in there, and a
// conformance failure already knows which set it concerns, so that set is a
// better anchor than the artifact it belongs to.
func scenarioFailureAnchor(
	err error,
	group *model.ResolvedGroup,
	fallback *model.Operation,
) (model.DiagnosticCategory, model.ExampleLocation) {
	var conformance *cassette.ConformanceError
	if errors.As(err, &conformance) {
		return model.DiagnosticCategoryValidation, setLocation(group, conformance.Key, fallback)
	}
	var incomplete *cassette.IncompleteError
	if errors.As(err, &incomplete) {
		return model.DiagnosticCategoryMaterialization, setLocation(group, incomplete.Key, fallback)
	}
	return model.DiagnosticCategoryMaterialization, operationLocation(fallback)
}

// setLocation turns a set key into a spec anchor, resolving the operation the
// key names so the path and method are the real ones rather than the tracked
// operation's.
func setLocation(
	group *model.ResolvedGroup,
	key cassette.SetKey,
	fallback *model.Operation,
) model.ExampleLocation {
	op := fallback
	if group != nil {
		for _, candidate := range lifecycleOperations(group) {
			if candidate.OperationId == key.OperationId {
				op = candidate
				break
			}
		}
	}
	location := operationLocation(op)
	location.OperationId = key.OperationId
	switch key.Role {
	case cassette.SetRoleRequest:
		location.Component = model.ExampleComponentRequestBody
	case cassette.SetRoleResponse:
		location.Component = model.ExampleComponentResponse
		location.Detail = key.Detail
	case cassette.SetRoleParameter:
		location.Component = model.ExampleComponentParameter
		location.Detail = key.Detail
	}
	return location
}
