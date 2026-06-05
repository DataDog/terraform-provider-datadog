package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func fixedTime() time.Time { return time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC) }

// sampleReport has one artifact in each ArtifactStatus, plus a diagnostic and a
// skipped operation, so every summary bucket and every nested shape is covered.
func sampleReport() *model.RunReport {
	ts := fixedTime()
	return &model.RunReport{
		RunId:            "11111111-1111-1111-1111-111111111111",
		GeneratorVersion: "v1.2.3",
		SpecHash:         "abc123",
		StartedAt:        ts,
		FinishedAt:       ts,
		Artifacts: []model.ArtifactReportEntry{
			{Name: "pet", Kind: model.ArtifactKindDataSource, Status: model.ArtifactStatusCreated, Path: "datadog/fwprovider/data_source_datadog_pet.go"},
			{Name: "team", Kind: model.ArtifactKindDataSource, Status: model.ArtifactStatusUnchanged, Path: "p2"},
			{Name: "monitor", Kind: model.ArtifactKindResource, Status: model.ArtifactStatusUpdated, Path: "p3"},
			{Name: "slo", Kind: model.ArtifactKindResource, Status: model.ArtifactStatusSkipped, Path: "p4"},
			{
				Name: "incident_type", Kind: model.ArtifactKindResource, Status: model.ArtifactStatusFailed, Path: "p5",
				Diagnostics: []model.Diagnostic{{Severity: model.SeverityError, Message: "boom", Location: "spec:x"}},
			},
		},
		SkippedOperations: []model.SkippedOperation{
			{OperationId: "ListThings", Path: "/things", Method: "GET", Reason: model.SkipReasonTrackingFieldAbsent},
		},
	}
}

func writeToMap(t *testing.T, r *model.RunReport) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, r); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	return m
}

func TestWriteJSONHasRequiredTopLevelFields(t *testing.T) {
	m := writeToMap(t, sampleReport())

	// Required  fields
	for _, k := range []string{"run_id", "generator_version", "spec_hash", "started_at", "finished_at", "artifacts", "summary"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing required key %q", k)
		}
	}

	for key, want := range map[string]any{
		"run_id":            "11111111-1111-1111-1111-111111111111",
		"generator_version": "v1.2.3",
		"spec_hash":         "abc123",
		"started_at":        "2026-06-03T12:00:00Z",
		"finished_at":       "2026-06-03T12:00:00Z",
	} {
		if m[key] != want {
			t.Errorf("%s = %v, want %v", key, m[key], want)
		}
	}
}

func TestWriteJSONSummaryCounts(t *testing.T) {
	m := writeToMap(t, sampleReport())

	sum, ok := m["summary"].(map[string]any)
	if !ok {
		t.Fatalf("summary missing or wrong type: %T", m["summary"])
	}
	want := map[string]float64{"created": 1, "updated": 1, "unchanged": 1, "skipped": 1, "failed": 1}
	for k, v := range want {
		got, ok := sum[k].(float64)
		if !ok || got != v {
			t.Errorf("summary[%q] = %v, want %v", k, sum[k], v)
		}
	}
}

// TestWriteJSONDerivesSummary proves the writer computes the summary from the
// artifacts (ignoring a stale caller value) and does not mutate the caller's report.
func TestWriteJSONDerivesSummary(t *testing.T) {
	r := sampleReport()
	r.Summary = &model.RunSummary{Created: 99} // deliberately wrong

	sum := writeToMap(t, r)["summary"].(map[string]any)
	if sum["created"].(float64) != 1 {
		t.Errorf("writer should derive summary from artifacts, got created=%v", sum["created"])
	}
	if r.Summary == nil || r.Summary.Created != 99 {
		t.Errorf("WriteJSON must not mutate the caller's report, Summary=%+v", r.Summary)
	}
}

func TestWriteJSONArtifactAndSkippedShape(t *testing.T) {
	m := writeToMap(t, sampleReport())

	arts, ok := m["artifacts"].([]any)
	if !ok || len(arts) != 5 {
		t.Fatalf("artifacts = %v, want 5 entries", m["artifacts"])
	}

	first := arts[0].(map[string]any)
	for _, k := range []string{"name", "kind", "status", "path"} {
		if _, ok := first[k]; !ok {
			t.Errorf("artifact entry missing required key %q", k)
		}
	}
	if first["kind"] != "data_source" || first["status"] != "created" {
		t.Errorf("artifact[0] kind/status = %v/%v, want data_source/created", first["kind"], first["status"])
	}

	// The failed artifact carries a diagnostic with severity + message + location.
	diags := arts[4].(map[string]any)["diagnostics"].([]any)
	d := diags[0].(map[string]any)
	if d["severity"] != "error" || d["message"] != "boom" || d["location"] != "spec:x" {
		t.Errorf("diagnostic = %v", d)
	}

	skipped, ok := m["skipped_operations"].([]any)
	if !ok || len(skipped) != 1 {
		t.Fatalf("skipped_operations = %v, want 1 entry", m["skipped_operations"])
	}
	so := skipped[0].(map[string]any)
	for _, k := range []string{"operation_id", "path", "method", "reason"} {
		if _, ok := so[k]; !ok {
			t.Errorf("skipped operation missing required key %q", k)
		}
	}
	if so["reason"] != "tracking_field_absent" || so["method"] != "GET" {
		t.Errorf("skipped operation = %v", so)
	}
}

func TestWriteJSONOmitsEmptyOptionalFields(t *testing.T) {
	ts := fixedTime()
	r := &model.RunReport{
		RunId: "x", GeneratorVersion: "v", SpecHash: "h", StartedAt: ts, FinishedAt: ts,
		Artifacts: []model.ArtifactReportEntry{
			{Name: "a", Kind: model.ArtifactKindDataSource, Status: model.ArtifactStatusUnchanged, Path: "p"},
		},
	}
	m := writeToMap(t, r)

	if _, ok := m["skipped_operations"]; ok {
		t.Error("skipped_operations should be omitted when empty")
	}
	art := m["artifacts"].([]any)[0].(map[string]any)
	if _, ok := art["diagnostics"]; ok {
		t.Error("diagnostics should be omitted when empty")
	}
	if _, ok := art["orphaned_hooks"]; ok {
		t.Error("orphaned_hooks should be omitted when empty")
	}
	if _, ok := m["summary"]; !ok {
		t.Error("summary should always be present, even with no skipped/failed artifacts")
	}
}

func TestWriteJSONDeterministic(t *testing.T) {
	r := sampleReport()
	var a, b bytes.Buffer
	if err := WriteJSON(&a, r); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if err := WriteJSON(&b, r); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if a.String() != b.String() {
		t.Errorf("non-deterministic output:\n--- run 1 ---\n%s\n--- run 2 ---\n%s", a.String(), b.String())
	}
}

func TestWriteJSONNilReport(t *testing.T) {
	if err := WriteJSON(io.Discard, nil); err == nil {
		t.Error("expected an error for a nil report")
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestWriteJSONPropagatesWriterError(t *testing.T) {
	if err := WriteJSON(errWriter{}, sampleReport()); err == nil {
		t.Error("expected WriteJSON to propagate the writer error")
	}
}
