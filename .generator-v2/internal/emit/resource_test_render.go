package emit

import (
	"bytes"
	"fmt"
	"go/format"

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
	apiPaths map[string]string,
) (resourceTestView, error) {
	common, err := BuildExampleTestView(scenario, view.Schema, apiPaths)
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
	return out, nil
}

// RenderResourceExampleTest renders a resource's example-backed acceptance
// test. The file is gofmt'd here so a syntax error surfaces as a generation
// failure naming the artifact, rather than as an unbuildable test package.
func RenderResourceExampleTest(
	scenario *model.GeneratedTestScenario,
	view ResourceView,
) ([]byte, error) {
	rendered, err := BuildResourceTestView(scenario, view)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, "resource_example_test", rendered); err != nil {
		return nil, fmt.Errorf("emit: executing example test template for %q: %w",
			scenario.ArtifactName, err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("emit: gofmt of generated example test %q: %w\n--- raw output ---\n%s",
			scenario.ArtifactName, err, buf.String())
	}
	return dropBlankLineAfterBrace(formatted), nil
}
