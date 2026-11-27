package cli

import (
	"errors"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/sdkbinding"
)

func TestExplicitIDBindingErrorAddsOnlySelectedPathContext(t *testing.T) {
	op := &model.Operation{
		Tracking: &model.TrackingFieldMetadata{ArtifactName: "thing", IdParameter: "thing_id"},
		ResolvedGroup: &model.ResolvedGroup{
			Create: &model.Operation{OperationId: "CreateThing"},
			Read:   &model.Operation{OperationId: "GetThing"},
			Update: &model.Operation{OperationId: "UpdateThing"},
			Delete: &model.Operation{OperationId: "ArchiveThing"},
		},
	}
	for _, role := range []model.GroupRole{model.GroupRoleCreate, model.GroupRoleRead, model.GroupRoleUpdate, model.GroupRoleDelete} {
		err := &sdkbinding.ParameterError{
			OperationId: op.ResolvedGroup.Op(role).OperationId,
			Parameter:   "thing_id", Location: "path", Cause: errors.New("unsupported schema"),
		}
		var contextual *model.IDParameterError
		if got := explicitIDBindingError(op, err); !errors.As(got, &contextual) {
			t.Fatalf("%s: selected path parameter error lacks ID context: %v", role, got)
		}
		if contextual.Artifact != "thing" || contextual.Role != role || contextual.OperationId != err.OperationId || contextual.Parameter != "thing_id" {
			t.Errorf("%s: incorrect explicit binding context: %#v", role, contextual)
		}
	}
	for _, unrelated := range []error{
		errors.New("independent failure"),
		&sdkbinding.ParameterError{OperationId: "GetThing", Parameter: "parent_id", Location: "path", Cause: errors.New("bad parent")},
		&sdkbinding.ParameterError{OperationId: "GetThing", Parameter: "thing_id", Location: "query", Cause: errors.New("bad query")},
		&sdkbinding.ParameterError{OperationId: "Unrelated", Parameter: "thing_id", Location: "path", Cause: errors.New("bad unrelated operation")},
	} {
		if got := explicitIDBindingError(op, unrelated); got != unrelated {
			t.Errorf("unrelated error was changed: %v -> %v", unrelated, got)
		}
	}
	op.Tracking.IdParameter = ""
	legacy := &sdkbinding.ParameterError{OperationId: "GetThing", Parameter: "thing_id", Location: "path", Cause: errors.New("bad schema")}
	if got := explicitIDBindingError(op, legacy); got != legacy {
		t.Errorf("omitting id_parameter changed existing SDK error: %v", got)
	}
}
