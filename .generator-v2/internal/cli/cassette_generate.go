package cli

import (
	"fmt"
	"path/filepath"
	"slices"
	"time"

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

// cassetteFreezeTime is the instant every generated freeze companion carries.
//
// It is a constant rather than the wall clock because regeneration has to be
// byte-identical: a timestamp taken at generation time would rewrite every
// fixture on every run. Its value barely matters — the replay harness only
// needs a parseable instant to restore, and a generated test derives its
// configuration from the description rather than from the clock — so a fixed,
// obviously-synthetic date is clearer than a plausible one.
var cassetteFreezeTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

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
		return ineligible(result, "a resource cassette needs a create, read and delete operation; "+
			"the tracking group resolves fewer than that")
	}

	selection, err := cassette.Select(lifecycleOperations(group))
	if err != nil {
		return ineligible(result, err.Error())
	}
	result.SelectedExample = selection.ScenarioName

	scenario, err := cassette.BuildResourceScenario(cassette.ResourceTarget{
		ArtifactName: op.Tracking.ArtifactName,
		ServerURL:    request.ServerURL,
		IdStrategy:   op.Tracking.IdStrategy,
		FreezeTime:   cassetteFreezeTime,
		Create:       group.Create,
		Read:         group.Read,
		Update:       group.Update,
		Delete:       group.Delete,
		Selection:    selection,
	})
	if err != nil {
		return ineligible(result, err.Error())
	}
	result.TestName = scenario.TestFuncName
	result.Interactions = interactionSummaries(scenario)

	source, err := emit.RenderResourceExampleTest(scenario, request.View, request.APIPaths)
	if err != nil {
		return ineligible(result, err.Error())
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
		return ineligible(result, err.Error())
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

// interactionSummaries renders the safe, reviewer-facing view of a trace.
// Bodies are deliberately absent: one may hold a replaced secret, and the run
// report is committed.
func interactionSummaries(scenario *model.GeneratedTestScenario) []model.InteractionSummary {
	out := make([]model.InteractionSummary, 0, len(scenario.Interactions))
	for _, interaction := range scenario.Interactions {
		summary := model.InteractionSummary{
			Index:       interaction.Index,
			Role:        string(interaction.Role),
			OperationId: interaction.OperationId,
			Method:      interaction.Request.Method,
			URL:         interaction.Request.URL,
			Status:      interaction.Response.StatusCode,
		}
		for _, provenance := range interaction.SourceExamples {
			summary.SourcePaths = append(summary.SourcePaths, provenance.Location.String())
		}
		out = append(out, summary)
	}
	return out
}

// ineligible attaches the reason a target could not produce a fixture. The
// message is what a description author has to act on, so it is carried rather
// than reduced to a status.
func ineligible(result model.CassetteResult, reason string) model.CassetteResult {
	result.Status = model.CassetteStatusIneligible
	result.WriteAction = model.CassetteWriteNone
	result.Diagnostics = append(result.Diagnostics, model.Diagnostic{
		Severity: model.SeverityWarning,
		Message:  fmt.Sprintf("no cassette generated: %s", reason),
	})
	return result
}
