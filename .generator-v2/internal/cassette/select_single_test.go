package cassette

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestNamedSingleExampleResolvesAcrossSets(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		namedResponse, fallback bool
	}{
		{name: "named request and response", namedResponse: true},
		{name: "named request and singular response"},
		{name: "named single takes precedence over schema fallback", namedResponse: true, fallback: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := &model.ExampleCandidate{Name: "single", SourceKind: model.ExampleSourceMedia, Value: map[string]any{"name": "chosen"}}
			response := &model.ExampleCandidate{SourceKind: model.ExampleSourceMedia, Value: map[string]any{"name": "chosen"}}
			requestSet := model.ExampleSet{Named: map[string]*model.ExampleCandidate{"single": request}}
			responseSet := model.ExampleSet{Single: response}
			if tc.namedResponse {
				response.Name = "single"
				responseSet = model.ExampleSet{Named: map[string]*model.ExampleCandidate{"single": response}}
			}
			if tc.fallback {
				requestSet.SchemaFallback = []*model.ExampleCandidate{{SourceKind: model.ExampleSourceSchema, Value: map[string]any{"name": "fallback"}}}
			}
			operation := &model.Operation{
				OperationId:      "CreateWidget",
				RequestExamples:  &model.RequestBodyExamples{Present: true, Examples: requestSet},
				ResponseExamples: []model.ResponseExamples{{Status: "201", BodyPresent: true, Examples: responseSet}},
			}
			selection, err := Select([]*model.Operation{operation})
			if err != nil {
				t.Fatal(err)
			}
			if selection.ScenarioName != "single" || len(selection.Sets) != 2 {
				t.Fatalf("unexpected selection: %+v", selection)
			}
			if selection.Sets[0].Candidate != request || selection.Sets[1].Candidate != response {
				t.Fatal("selection did not retain the declared candidates")
			}
		})
	}
}
