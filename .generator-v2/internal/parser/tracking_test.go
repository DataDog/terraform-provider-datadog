package parser

import (
	"errors"
	"strings"
	"testing"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	yaml "go.yaml.in/yaml/v4"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

const (
	trackPath   = "/api/v2/things"
	trackMethod = "POST"
)

// opWithExt builds a *v3.Operation carrying a single vendor extension under
// key, whose value is the YAML mapping in body — mirroring how libopenapi
// stores an extension (the mapping node, i.e. the parsed document's Content[0]).
func opWithExt(t *testing.T, operationId, key, body string) *v3.Operation {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatalf("unmarshal extension body: %v", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		t.Fatalf("expected a YAML document node with content, got kind %d", doc.Kind)
	}
	ext := orderedmap.ToOrderedMap(map[string]*yaml.Node{key: doc.Content[0]})
	return &v3.Operation{OperationId: operationId, Extensions: ext}
}

func TestDecodeTrackingOutOfScopeReturnsNil(t *testing.T) {
	t.Run("nil operation", func(t *testing.T) {
		got, err := DecodeTracking(nil, trackPath, trackMethod, DefaultTrackingFieldName)
		if got != nil || err != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})
	t.Run("no extensions", func(t *testing.T) {
		got, err := DecodeTracking(&v3.Operation{OperationId: "Op"}, trackPath, trackMethod, DefaultTrackingFieldName)
		if got != nil || err != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})
	t.Run("different extension key present", func(t *testing.T) {
		op := opWithExt(t, "Op", "x-some-other-extension", "foo: bar\n")
		got, err := DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
		if got != nil || err != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})
	// skip:true short-circuits BEFORE validation — a skipped operation is out of
	// scope, so even an otherwise-malformed extension produces no error (which
	// also subsumes a well-formed annotation carrying skip:true).
	t.Run("skip true bypasses validation", func(t *testing.T) {
		op := opWithExt(t, "Op", DefaultTrackingFieldName,
			"artifact_kind: widget\nbogus: 1\nskip: true\n") // bad kind, unknown prop, missing name
		got, err := DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
		if got != nil || err != nil {
			t.Fatalf("got (%v, %v), want (nil, nil) — skip must bypass validation", got, err)
		}
	})
}

func TestDecodeTrackingValidResource(t *testing.T) {
	op := opWithExt(t, "CreateIncidentType", DefaultTrackingFieldName, `
artifact_kind: resource
artifact_name: incident_type
group:
  create: CreateIncidentType
  read: GetIncidentType
  update: UpdateIncidentType
  delete: DeleteIncidentType
id_strategy: data.attributes.id
`)
	got, err := DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
	if err != nil {
		t.Fatalf("DecodeTracking: %v", err)
	}
	if got == nil {
		t.Fatal("expected metadata, got nil")
	}
	if got.ArtifactKind != model.ArtifactKindResource {
		t.Errorf("ArtifactKind = %q, want %q", got.ArtifactKind, model.ArtifactKindResource)
	}
	if got.ArtifactName != "incident_type" {
		t.Errorf("ArtifactName = %q, want %q", got.ArtifactName, "incident_type")
	}
	if got.IdStrategy != model.IdStrategyDataAttributesID {
		t.Errorf("IdStrategy = %q, want %q", got.IdStrategy, model.IdStrategyDataAttributesID)
	}
	if got.Group == nil || got.Group.Create != "CreateIncidentType" || got.Group.Delete != "DeleteIncidentType" {
		t.Errorf("Group = %+v, want full CRUD", got.Group)
	}
}

func TestDecodeTrackingValidDataSource(t *testing.T) {
	op := opWithExt(t, "GetTeam", DefaultTrackingFieldName, `
artifact_kind: data_source
artifact_name: team
group:
  read: GetTeam
`)
	got, err := DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
	if err != nil {
		t.Fatalf("DecodeTracking: %v", err)
	}
	if got == nil || got.ArtifactKind != model.ArtifactKindDataSource {
		t.Fatalf("got %+v, want data_source kind", got)
	}
	if got.Group == nil || got.Group.Read != "GetTeam" {
		t.Errorf("Group = %+v, want Read=GetTeam", got.Group)
	}
}

// TestDecodeTrackingOptionalFields covers each non-required field both present
// and omitted. The required pair (artifact_kind + artifact_name) alone is a
// schema-valid extension, so it is the base every case builds on.
func TestDecodeTrackingOptionalFields(t *testing.T) {
	const required = "artifact_kind: data_source\nartifact_name: team\n"

	decode := func(t *testing.T, body string) *model.TrackingFieldMetadata {
		t.Helper()
		op := opWithExt(t, "Op", DefaultTrackingFieldName, body)
		got, err := DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
		if err != nil {
			t.Fatalf("DecodeTracking: %v", err)
		}
		if got == nil {
			t.Fatal("expected metadata, got nil")
		}
		return got
	}

	// All four optional fields omitted: group nil, id_strategy defaulted,
	// sensitive/skip false. Also confirms a groupless extension is schema-valid.
	t.Run("all omitted", func(t *testing.T) {
		got := decode(t, required)
		if got.Group != nil {
			t.Errorf("Group = %+v, want nil when omitted", got.Group)
		}
		if got.IdStrategy != model.IdStrategyDataID {
			t.Errorf("IdStrategy = %q, want default %q", got.IdStrategy, model.IdStrategyDataID)
		}
		if got.Sensitive {
			t.Error("Sensitive = true, want false when omitted")
		}
		if got.Skip {
			t.Error("Skip = true, want false when omitted")
		}
	})

	t.Run("group present", func(t *testing.T) {
		got := decode(t, required+"group:\n  create: C\n  read: R\n  update: U\n  delete: D\n")
		if got.Group == nil {
			t.Fatal("Group = nil, want populated")
		}
		if got.Group.Create != "C" || got.Group.Read != "R" || got.Group.Update != "U" || got.Group.Delete != "D" {
			t.Errorf("Group = %+v, want C/R/U/D", got.Group)
		}
	})

	t.Run("id_strategy present", func(t *testing.T) {
		got := decode(t, required+"id_strategy: data.attributes.uuid\n")
		if got.IdStrategy != model.IdStrategyDataAttributesUID {
			t.Errorf("IdStrategy = %q, want %q", got.IdStrategy, model.IdStrategyDataAttributesUID)
		}
	})

	// sensitive present also proves additionalProperties:false does not reject
	// the declared optional property.
	t.Run("sensitive present", func(t *testing.T) {
		got := decode(t, required+"sensitive: true\n")
		if !got.Sensitive {
			t.Error("Sensitive = false, want true")
		}
	})

	// skip:true is the one optional field whose presence changes control flow —
	// it short-circuits to (nil, nil) before validation (covered in
	// TestDecodeTrackingOutOfScopeReturnsNil). An explicit skip:false decodes
	// normally.
	t.Run("skip false present", func(t *testing.T) {
		got := decode(t, required+"skip: false\n")
		if got.Skip {
			t.Error("Skip = true, want false")
		}
	})
}

func TestDecodeTrackingMalformedReturnsTrackingError(t *testing.T) {
	cases := map[string]string{
		"missing artifact_name": "artifact_kind: resource\n",
		"missing artifact_kind": "artifact_name: thing\n",
		"unknown artifact_kind": "artifact_kind: widget\nartifact_name: thing\n",
		"unknown property":      "artifact_kind: resource\nartifact_name: thing\nbogus: 1\n",
		"bad name pattern":      "artifact_kind: resource\nartifact_name: NotSnake\n",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			op := opWithExt(t, "Op", DefaultTrackingFieldName, body)
			got, err := DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
			if got != nil {
				t.Errorf("expected nil metadata on error, got %+v", got)
			}
			var te *TrackingError
			if !errors.As(err, &te) {
				t.Fatalf("error %v (%T) is not a *TrackingError", err, err)
			}
		})
	}
}

func TestDecodeTrackingErrorNamesOpenAPILocation(t *testing.T) {
	op := opWithExt(t, "CreateWidget", DefaultTrackingFieldName, "artifact_kind: widget\nartifact_name: widget\n")
	_, err := DecodeTracking(op, "/api/v2/widgets", "POST", DefaultTrackingFieldName)
	var te *TrackingError
	if !errors.As(err, &te) {
		t.Fatalf("error %v (%T) is not a *TrackingError", err, err)
	}
	if te.Path != "/api/v2/widgets" || te.Method != "POST" || te.OperationId != "CreateWidget" {
		t.Errorf("location = %s %s (%s), want POST /api/v2/widgets (CreateWidget)", te.Method, te.Path, te.OperationId)
	}
	if !strings.Contains(te.Error(), "/api/v2/widgets") {
		t.Errorf("error message %q does not contain the OpenAPI path", te.Error())
	}
}

// TestDecodeTrackingHonorsCustomExtensionName covers the --tracking-field
// override path (reserved for generator-internal fixture tests).
func TestDecodeTrackingHonorsCustomExtensionName(t *testing.T) {
	const custom = "x-tfgen-fixture"
	op := opWithExt(t, "GetTeam", custom, "artifact_kind: data_source\nartifact_name: team\ngroup:\n  read: GetTeam\n")

	got, err := DecodeTracking(op, trackPath, trackMethod, custom)
	if err != nil {
		t.Fatalf("DecodeTracking(custom): %v", err)
	}
	if got == nil || got.ArtifactName != "team" {
		t.Fatalf("got %+v, want artifact_name team", got)
	}

	// The same op under the default name resolves to out-of-scope.
	got, err = DecodeTracking(op, trackPath, trackMethod, DefaultTrackingFieldName)
	if got != nil || err != nil {
		t.Fatalf("got (%v, %v), want (nil, nil) under non-matching name", got, err)
	}
}
