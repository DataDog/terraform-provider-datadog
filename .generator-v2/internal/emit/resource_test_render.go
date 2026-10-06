package emit

import (
	"fmt"
	"strings"

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
	ReadMethod                      string
	ReadArgument                    SDKArgumentView
	UsesUUID, UsesStrconv, UsesTime bool
	UsesSDK                         bool
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
	argument := view.Read.Arguments[0]
	if argument.TFName != "id" {
		return resourceTestView{}, fmt.Errorf("resource %q: its destroy check cannot supply read argument %q from the resource id", scenario.ArtifactName, argument.TFName)
	}
	// Reuse the SDK binding's conversion, replacing only its Terraform-state
	// source with the test harness's string ID.
	if argument.GoType == "" || argument.GoType == "string" {
		argument.Expression = "r.Primary.ID"
	} else {
		idSource := "state." + goFieldName("id") + ".ValueString()"
		argument.Expression = strings.ReplaceAll(argument.Expression, idSource, "r.Primary.ID")
		argument.ParseCall = strings.ReplaceAll(argument.ParseCall, idSource, "r.Primary.ID")
	}
	usesUUID, usesStrconv, usesTime := parseCallImports(argument.ParseCall)

	out := resourceTestView{
		exampleTestView:  common,
		DestroyCheckFunc: testHelperName(scenario.TestFuncName, "Destroy"),
		APIAccessor:      view.APIAccessor,
		SDKPackage:       view.SDKPackage,
		APIConstructor:   view.APIConstructor,
		ReadMethod:       view.Read.Method,
		ReadArgument:     argument,
		UsesUUID:         usesUUID,
		UsesStrconv:      usesStrconv,
		UsesTime:         usesTime,
		UsesSDK:          view.APIAccessor == "" || strings.Contains(argument.ParseCall, view.SDKPackage+".") || strings.Contains(argument.Expression, view.SDKPackage+"."),
	}
	return out, nil
}

// RenderResourceExampleTest renders a resource's example-backed acceptance
// test. The file is gofmt'd here so a syntax error surfaces as a generation
// failure naming the artifact, rather than as an unbuildable test package.
func RenderResourceExampleTest(
	scenario *model.GeneratedTestScenario,
	view ResourceView,
	apiPaths map[string]string,
) ([]byte, error) {
	rendered, err := BuildResourceTestView(scenario, view, apiPaths)
	if err != nil {
		return nil, err
	}
	return renderGoTemplate("resource_example_test", "example test", scenario.ArtifactName, rendered)
}
