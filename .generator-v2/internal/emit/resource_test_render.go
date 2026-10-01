package emit

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Resource-specific example-test rendering
//
// A resource scenario ends with a delete and a read that expects a 404, so the
// generated test needs a CheckDestroy that actually calls the API — that read
// is what the destroy-verification interaction exists for. It is the one check
// in a generated example-backed test that is not state-only: everything else
// is already pinned by the cassette, but proving the object is gone requires
// asking.
// ----------------------------------------------------------------------------

// resourceTestView is the render context for a resource's example-backed test:
// the common parts plus how its destroy check reaches the API.
type resourceTestView struct {
	exampleTestView

	// DestroyCheckFunc is the generated CheckDestroy helper's name.
	DestroyCheckFunc string
	// APIAccessor is the DatadogApiInstances accessor, e.g.
	// "GetIncidentsApiV2". Empty when the SDK client must be constructed
	// directly, in which case SDKPackage and APIConstructor are set.
	APIAccessor string
	// SDKPackage and APIConstructor construct the client when no accessor
	// exists, mirroring what the generated resource's Configure does.
	SDKPackage     string
	APIConstructor string
	// ReadMethod is the SDK read call, e.g. "GetIncidentType".
	ReadMethod string
	// UpdateStepIndex is the one-based index of the update step, or zero when
	// the scenario has none. Templates use it to label the step.
	UpdateStepIndex int
}

// BuildResourceTestView derives the render context for a resource's
// example-backed test.
//
// It refuses a read whose SDK call needs more than the resource id, because a
// CheckDestroy only has the id to work from: Terraform state carries
// resource.Primary.ID and nothing else the call could supply. Rendering a test
// without the destroy check instead would leave the scenario's 404 interaction
// unconsumed, so the target is reported ineligible rather than producing a
// bundle whose cassette and test disagree.
func BuildResourceTestView(
	scenario *model.GeneratedTestScenario,
	view ResourceView,
) (resourceTestView, error) {
	common, err := BuildExampleTestView(scenario, view.Schema)
	if err != nil {
		return resourceTestView{}, err
	}
	if view.Read.Method == "" {
		return resourceTestView{}, fmt.Errorf(
			"resource %q: no SDK read call, so its destroy check cannot confirm the object is gone",
			scenario.ArtifactName)
	}
	if count := len(view.Read.Arguments); count != 1 {
		return resourceTestView{}, fmt.Errorf(
			"resource %q: its read takes %d arguments besides auth, but a destroy check has only the "+
				"resource id from Terraform state to supply",
			scenario.ArtifactName, count)
	}

	out := resourceTestView{
		exampleTestView:  common,
		DestroyCheckFunc: testHelperName(scenario.TestFuncName, "Destroy"),
		APIAccessor:      view.APIAccessor,
		SDKPackage:       view.SDKPackage,
		APIConstructor:   view.APIConstructor,
		ReadMethod:       view.Read.Method,
	}
	if scenario.HasUpdateStep() {
		// The scenario adds at most one update step, and it is always the last.
		out.UpdateStepIndex = len(scenario.Steps)
	}
	return out, nil
}
