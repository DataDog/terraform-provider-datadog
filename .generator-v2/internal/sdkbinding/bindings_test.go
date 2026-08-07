package sdkbinding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestBindDerivesSDKGeneratorNamesTypesAndOrder(t *testing.T) {
	op := &model.Operation{
		OperationId: "GetThing",
		PathParams: []model.QueryParam{
			{Name: "thing_id", Required: true, DeclarationOrder: 3, Schema: scalar("string", "uuid")},
			{Name: "account_id", Required: true, DeclarationOrder: 1, Schema: scalar("string", "")},
		},
		QueryParams: []model.QueryParam{
			{Name: "start_time", Required: true, DeclarationOrder: 2, Schema: scalar("integer", "int64")},
			{Name: "filter[type]", DeclarationOrder: 4, Schema: namedEnum("ThingFilterType")},
			{Name: "type", DeclarationOrder: 5, Schema: scalar("boolean", "")},
		},
	}
	diagnostics, err := Bind(op, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	if got, want := len(op.SDKBinding.Required), 3; got != want {
		t.Fatalf("required argument count = %d, want %d", got, want)
	}
	wantRequired := []struct {
		name, goName, goType, location string
	}{
		{"account_id", "accountId", "string", "path"},
		{"start_time", "startTime", "int64", "query"},
		{"thing_id", "thingId", "uuid.UUID", "path"},
	}
	for index, want := range wantRequired {
		got := op.SDKBinding.Required[index]
		if got.Name != want.name || got.GoName != want.goName || got.GoType != want.goType || got.Location != want.location {
			t.Fatalf("required argument %d = %#v, want %#v", index, got, want)
		}
	}
	if got, want := op.SDKBinding.OptionalParamsType, "GetThingOptionalParameters"; got != want {
		t.Fatalf("optional parameters type = %q, want %q", got, want)
	}
	if got, want := len(op.SDKBinding.Optional), 2; got != want {
		t.Fatalf("optional argument count = %d, want %d", got, want)
	}
	first, second := op.SDKBinding.Optional[0], op.SDKBinding.Optional[1]
	if first.Name != "filter[type]" || first.GoName != "filterType" || first.GoType != "ThingFilterType" || first.Setter != "WithFilterType" {
		t.Fatalf("first optional argument = %#v", first)
	}
	if second.Name != "type" || second.GoName != "typeVar" || second.GoType != "bool" || second.Setter != "WithType" {
		t.Fatalf("second optional argument = %#v", second)
	}
}

func TestBindSupportsSingletonAndMissingPinnedEndpoint(t *testing.T) {
	op := &model.Operation{OperationId: "GetCurrentUser"}
	inventory := &Inventory{methods: map[string]methodSignature{}, setters: map[string]map[string]argumentSignature{}}
	diagnostics, err := Bind(op, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("missing pinned endpoint diagnostics = %#v", diagnostics)
	}
	if op.SDKBinding == nil || len(op.SDKBinding.Required) != 0 || op.SDKBinding.OptionalParamsType != "" {
		t.Fatalf("singleton binding = %#v", op.SDKBinding)
	}
}

func TestBindCorroboratesMatchingPinnedSDK(t *testing.T) {
	inventory := loadFixtureInventory(t, `package datadogV2
type ThingsApi struct{}
type GetThingOptionalParameters struct{}
func (a *ThingsApi) GetThing(ctx context.Context, thingId uuid.UUID, o ...GetThingOptionalParameters) {}
func (o *GetThingOptionalParameters) WithInclude(include ThingInclude) *GetThingOptionalParameters { return o }
`)
	op := &model.Operation{
		OperationId: "GetThing",
		PathParams:  []model.QueryParam{{Name: "thing_id", Required: true, Schema: scalar("string", "uuid")}},
		QueryParams: []model.QueryParam{{Name: "include", Schema: namedEnum("ThingInclude")}},
	}
	diagnostics, err := Bind(op, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("matching SDK diagnostics = %#v", diagnostics)
	}
}

func TestBindWarnsAndKeepsDerivedBindingOnPinnedMismatch(t *testing.T) {
	inventory := loadFixtureInventory(t, `package datadogV2
type ThingsApi struct{}
type GetThingOptionalParameters struct{}
func (a *ThingsApi) GetThing(ctx context.Context, thingId string, o ...GetThingOptionalParameters) {}
func (o *GetThingOptionalParameters) WithInclude(include string) *GetThingOptionalParameters { return o }
`)
	op := &model.Operation{
		OperationId: "GetThing",
		PathParams:  []model.QueryParam{{Name: "thing_id", Required: true, Schema: scalar("string", "uuid")}},
		QueryParams: []model.QueryParam{{Name: "include", Schema: namedEnum("ThingInclude")}},
	}
	diagnostics, err := Bind(op, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(diagnostics), 1; got != want {
		t.Fatalf("diagnostic count = %d, want %d: %#v", got, want, diagnostics)
	}
	if diagnostics[0].Severity != model.SeverityWarning || !strings.Contains(diagnostics[0].Message, "generated with OpenAPI binding") {
		t.Fatalf("mismatch diagnostic = %#v", diagnostics[0])
	}
	if got, want := op.SDKBinding.Required[0].GoType, "uuid.UUID"; got != want {
		t.Fatalf("authoritative derived type = %q, want %q", got, want)
	}
	if got, want := op.SDKBinding.Optional[0].GoType, "ThingInclude"; got != want {
		t.Fatalf("authoritative derived setter type = %q, want %q", got, want)
	}
}

func TestBindWarnsWhenPinnedSDKHasAnExtraSetter(t *testing.T) {
	inventory := loadFixtureInventory(t, `package datadogV2
type ThingsApi struct{}
type ListThingsOptionalParameters struct{}
func (a *ThingsApi) ListThings(ctx context.Context, o ...ListThingsOptionalParameters) {}
func (o *ListThingsOptionalParameters) WithInclude(include string) *ListThingsOptionalParameters { return o }
func (o *ListThingsOptionalParameters) WithHeaderValue(headerValue string) *ListThingsOptionalParameters { return o }
`)
	op := &model.Operation{
		OperationId: "ListThings",
		QueryParams: []model.QueryParam{{Name: "include", Schema: scalar("string", "")}},
	}
	diagnostics, err := Bind(op, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(diagnostics), 1; got != want {
		t.Fatalf("diagnostic count = %d, want %d: %#v", got, want, diagnostics)
	}
	if !strings.Contains(diagnostics[0].Message, "WithHeaderValue") || !strings.Contains(diagnostics[0].Message, "absent from the OpenAPI-derived binding") {
		t.Fatalf("extra setter diagnostic = %#v", diagnostics[0])
	}
}

func TestBindRejectsAnonymousEnumTheSDKGeneratorCannotName(t *testing.T) {
	op := &model.Operation{
		OperationId: "ListThings",
		QueryParams: []model.QueryParam{{
			Name: "sort", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"name"}},
		}},
	}
	_, err := Bind(op, nil)
	if err == nil || !strings.Contains(err.Error(), "anonymous enum") {
		t.Fatalf("error = %v, want anonymous enum error", err)
	}
}

func loadFixtureInventory(t *testing.T, source string) *Inventory {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "api_things.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	inventory, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return inventory
}

func scalar(typ, format string) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindPrimitive, Type: typ, Format: format}
}

func namedEnum(name string) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"value"}, RefName: name}
}
