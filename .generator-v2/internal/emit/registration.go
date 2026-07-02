package emit

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// GeneratedRegistration records how one successfully generated data source is
// wired into the framework provider: the generated constructor to register in
// generatedDatasources, and the hand-written constructor it overwrites (removed
// from the Datasources slice), empty when the data source is purely additive.
type GeneratedRegistration struct {
	Constructor string
	Overwrites  string
}

// DatasourceConstructor returns the exported constructor a generated data source
// declares for artifact name. It matches the New<title GoName>DataSource the
// data-source template emits, GoName being the Datadog-prefixed dsGoName base.
func DatasourceConstructor(name string) string {
	return "New" + upperFirst(dsGoName(name)) + "DataSource"
}

// datasourceConstructorRe matches a New<...>DataSource constructor identifier.
// datasources_generated.go holds nothing else that fits the pattern, so it
// safely recovers the already-registered set from the file's current contents.
var datasourceConstructorRe = regexp.MustCompile(`New[A-Za-z0-9_]+DataSource`)

// GeneratedDatasourceRegistered reports whether constructor already appears in
// the generatedDatasources file at path (a missing file reports false).
// wireGeneratedDatasources uses it to tell an idempotent re-run, where a prior
// run already retired the overwrites target, from an overwrites target that
// never existed in the framework Datasources slice.
func GeneratedDatasourceRegistered(path, constructor string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return slices.Contains(datasourceConstructorRe.FindAllString(string(data), -1), constructor), nil
}

// generatedDatasourcesHeader is everything in datasources_generated.go up to and
// including the slice literal's opening brace. SyncGeneratedDatasources appends
// the sorted constructors and the closing brace, then gofmt canonicalizes it.
const generatedDatasourcesHeader = `package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/datasource"

// generatedDatasources holds the data sources produced by the generator-v2 emit
// pipeline. tfgen owns this file: every generate run rewrites it from the set of
// data sources it produced, keeping the generated registrations in one
// reviewable place without churning framework_provider.go. Do not edit by hand.
//
// FrameworkProvider.DataSources registers this slice alongside the hand-written
// Datasources.
var generatedDatasources = []func() datasource.DataSource{`

// SyncGeneratedDatasources rewrites path's generatedDatasources slice to hold the
// union of the constructors already registered there and the ones passed in,
// sorted and de-duplicated. Merging (rather than replacing) keeps a partial
// --include run from dropping data sources it did not regenerate this time. It
// honors check mode through WriteFile.
func SyncGeneratedDatasources(path string, constructors []string, check bool) (model.ArtifactStatus, error) {
	set := map[string]struct{}{}
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return model.ArtifactStatusFailed, err
	}
	for _, c := range datasourceConstructorRe.FindAllString(string(existing), -1) {
		set[c] = struct{}{}
	}
	for _, c := range constructors {
		set[c] = struct{}{}
	}

	names := make([]string, 0, len(set))
	for c := range set {
		names = append(names, c)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	buf.WriteString(generatedDatasourcesHeader)
	buf.WriteByte('\n')
	for _, c := range names {
		buf.WriteByte('\t')
		buf.WriteString(c)
		buf.WriteString(",\n")
	}
	buf.WriteString("}\n")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return model.ArtifactStatusFailed, fmt.Errorf("emit: gofmt of generatedDatasources: %w", err)
	}
	return WriteFile(path, src, check)
}

// datasourcesSliceHeader is the line opening the hand-written Datasources slice
// in framework_provider.go. RemoveHandwrittenDatasource scopes its line removal
// to this block so a like-named entry in another slice is never touched.
const datasourcesSliceHeader = "var Datasources = []func() datasource.DataSource{"

// RemoveHandwrittenDatasource deletes constructor from the hand-written
// Datasources slice in framework_provider.go (the file at path) — the slice a
// generated data source supersedes when its spec sets overwrites. The removal is
// scoped to the Datasources block so a like-named Resources entry is never
// touched, and it is idempotent: an already-absent constructor reports Unchanged.
// It honors check mode by not writing.
func RemoveHandwrittenDatasource(path, constructor string, check bool) (model.ArtifactStatus, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}

	lines := strings.Split(string(original), "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == datasourcesSliceHeader {
			start = i
			break
		}
	}
	if start == -1 {
		return model.ArtifactStatusFailed, fmt.Errorf("emit: %s: Datasources slice not found", path)
	}
	end := -1
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "}" {
			end = i
			break
		}
	}
	if end == -1 {
		return model.ArtifactStatusFailed, fmt.Errorf("emit: %s: Datasources slice is not terminated", path)
	}

	target := constructor + ","
	out := make([]string, 0, len(lines))
	removed := false
	for i, line := range lines {
		if i > start && i < end && !removed && strings.TrimSpace(line) == target {
			removed = true
			continue
		}
		out = append(out, line)
	}
	if !removed {
		return model.ArtifactStatusUnchanged, nil
	}

	return WriteFile(path, []byte(strings.Join(out, "\n")), check)
}
