package sdkbinding

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestBindResolvesScalarArgumentsInSDKOrder(t *testing.T) {
	dir := t.TempDir()
	source := `package datadogV2

type ThingsApi struct{}
type GetThingOptionalParameters struct{}

func (a *ThingsApi) GetThing(ctx context.Context, accountId string, thingId uuid.UUID, o ...GetThingOptionalParameters) {}
func (o *GetThingOptionalParameters) WithInclude(include ThingInclude) *GetThingOptionalParameters { return o }
`
	if err := os.WriteFile(filepath.Join(dir, "api_things.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	inv, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	op := &model.Operation{
		OperationId: "GetThing",
		PathParams: []model.QueryParam{
			{Name: "thing_id", Required: true, Schema: scalar("string", "uuid")},
			{Name: "account_id", Required: true, Schema: scalar("string", "")},
		},
		QueryParams: []model.QueryParam{
			{Name: "include", Schema: scalar("string", "")},
		},
	}
	if err := inv.Bind(op); err != nil {
		t.Fatal(err)
	}

	if got, want := len(op.SDKBinding.Required), 2; got != want {
		t.Fatalf("required argument count = %d, want %d", got, want)
	}
	first, second := op.SDKBinding.Required[0], op.SDKBinding.Required[1]
	if first.Name != "account_id" || first.GoName != "accountId" || first.GoType != "string" || first.Location != "path" {
		t.Fatalf("first required argument = %#v", first)
	}
	if second.Name != "thing_id" || second.GoName != "thingId" || second.GoType != "uuid.UUID" || second.Location != "path" {
		t.Fatalf("second required argument = %#v", second)
	}
	if got, want := op.SDKBinding.OptionalParamsType, "GetThingOptionalParameters"; got != want {
		t.Fatalf("optional parameters type = %q, want %q", got, want)
	}
	if got, want := len(op.SDKBinding.Optional), 1; got != want {
		t.Fatalf("optional argument count = %d, want %d", got, want)
	}
	optional := op.SDKBinding.Optional[0]
	if optional.Name != "include" || optional.GoType != "ThingInclude" || optional.Setter != "WithInclude" || optional.Location != "query" {
		t.Fatalf("optional argument = %#v", optional)
	}
}

func TestBindSupportsSingletonMethod(t *testing.T) {
	dir := t.TempDir()
	source := `package datadogV2
type UsersApi struct{}
func (a *UsersApi) GetCurrentUser(ctx context.Context) {}
`
	if err := os.WriteFile(filepath.Join(dir, "api_users.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	inv, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	op := &model.Operation{OperationId: "GetCurrentUser"}
	if err := inv.Bind(op); err != nil {
		t.Fatal(err)
	}
	if op.SDKBinding == nil || len(op.SDKBinding.Required) != 0 {
		t.Fatalf("singleton binding = %#v", op.SDKBinding)
	}
}

func TestBindRejectsUnmatchedSDKArgument(t *testing.T) {
	dir := t.TempDir()
	source := `package datadogV2
type ThingsApi struct{}
func (a *ThingsApi) GetThing(ctx context.Context, thingId string) {}
`
	if err := os.WriteFile(filepath.Join(dir, "api_things.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	inv, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := inv.Bind(&model.Operation{OperationId: "GetThing"}); err == nil {
		t.Fatal("expected unmatched SDK argument error")
	}
}

func scalar(typ, format string) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindPrimitive, Type: typ, Format: format}
}
