package sdkbinding

import (
	"errors"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestParameterErrorRetainsSourceAndExistingDiagnostic(t *testing.T) {
	for _, optional := range []bool{false, true} {
		op := &model.Operation{OperationId: "GetThing"}
		parameter := model.QueryParam{Name: "thing_id"} // Missing schema cannot derive an SDK type.
		wantLocation := "path"
		wantMessage := "sdk binding: derive GetThing argument thing_id: parameter has no schema"
		if optional {
			op.QueryParams = []model.QueryParam{parameter}
			wantLocation = "query"
			wantMessage = "sdk binding: derive GetThing optional argument thing_id: parameter has no schema"
		} else {
			op.PathParams = []model.QueryParam{parameter}
		}
		_, err := Bind(op, nil)
		var parameterErr *ParameterError
		if !errors.As(err, &parameterErr) {
			t.Fatalf("Bind error = %v, want ParameterError", err)
		}
		if parameterErr.OperationId != "GetThing" || parameterErr.Parameter != "thing_id" || parameterErr.Location != wantLocation || parameterErr.Optional != optional {
			t.Errorf("incorrect SDK parameter context: %#v", parameterErr)
		}
		if err.Error() != wantMessage {
			t.Errorf("existing diagnostic changed: got %q, want %q", err, wantMessage)
		}
		if !errors.Is(err, parameterErr.Cause) {
			t.Error("typed parameter error lost its underlying cause")
		}
	}
}
