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
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// GeneratedRegistration records how one generated data source is wired into the
// framework provider: the generated constructor to register, and the
// hand-written constructor it overwrites (removed from the Datasources slice,
// empty when the data source is purely additive).
type GeneratedRegistration struct {
	Constructor string
	Overwrites  string
	// TestFileKey is the testFiles2EndpointTags key for the generated test, or
	// empty when no test was emitted (so the endpoint-tag map is left untouched).
	TestFileKey string
	// EndpointTag is the normalized OpenAPI tag registered as the test's map
	// value; empty exactly when TestFileKey is.
	EndpointTag string
	// UnstableOperations are the x-unstable SDK keys this artifact calls, which
	// the provider must enable for any of its calls to reach the API.
	UnstableOperations []string
}

// DatasourceConstructor returns the exported constructor a generated data source
// declares for artifact name. It matches the New<title GoName>DataSource the
// data-source template emits, GoName being the Datadog-prefixed dsGoName base.
func DatasourceConstructor(name string) string {
	return "New" + upperFirst(dsGoName(name)) + "DataSource"
}

// RegistrationRetirementName derives a filesystem-safe artifact name from a
// constructor, for a retirement whose generated files are already gone so the
// real artifact name can no longer be read back. It strips the New/DataSource
// affixes and snake-cases the rest, e.g. NewDatadogTeamDataSource ->
// datadog_team. The name only labels output; the constructor stays canonical.
func RegistrationRetirementName(constructor string) string {
	base := strings.TrimSuffix(strings.TrimPrefix(constructor, "New"), "DataSource")
	return model.SnakeCase(base)
}

// datasourceConstructorRe matches a New<...>DataSource constructor identifier.
// datasources_generated.go holds nothing else that fits the pattern, so it
// safely recovers the already-registered set from the file's current contents.
var datasourceConstructorRe = regexp.MustCompile(`New[A-Za-z0-9_]+DataSource`)

// GeneratedDatasourceRegistered reports whether constructor already appears in
// the generatedDatasources file at path (a missing file reports false).
func GeneratedDatasourceRegistered(path, constructor string) (bool, error) {
	return generatedRegistered(path, datasourceConstructorRe, constructor)
}

// generatedRegistered reports whether constructor appears among the tokens re
// matches in the generated slice file at path (a missing file reports false).
func generatedRegistered(path string, re *regexp.Regexp, constructor string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return slices.Contains(re.FindAllString(string(data), -1), constructor), nil
}

// RegisteredGeneratedDatasources returns the constructor identifiers currently
// registered in the generatedDatasources file at path, sorted and
// de-duplicated. A missing file yields an empty slice.
func RegisteredGeneratedDatasources(path string) ([]string, error) {
	set, err := registeredSet(path)
	if err != nil {
		return nil, err
	}
	return sortedKeys(set), nil
}

// generatedDatasourcesHeader is everything in datasources_generated.go up to and
// including the slice literal's opening brace. SyncGeneratedDatasources appends
// the sorted constructors and the closing brace, then gofmt canonicalizes it.
const generatedDatasourcesHeader = `package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/datasource"

// generatedDatasources holds the data sources produced by the generator-v2 emit
// pipeline. tfgen owns this file: each run merges the constructors it produced
// into the existing set (union, sorted) so a scoped --include run never drops
// data sources it did not regenerate; reconcile prunes entries whose annotation
// is gone. Do not edit by hand.
//
// FrameworkProvider.DataSources registers this slice alongside the hand-written
// Datasources.
var generatedDatasources = []func() datasource.DataSource{`

// SyncGeneratedDatasources rewrites path's generatedDatasources slice to hold
// the union of the constructors already registered there and the ones passed
// in, sorted and de-duplicated. Merging rather than replacing keeps a partial
// run from dropping data sources it did not regenerate. Honors check mode.
func SyncGeneratedDatasources(path string, constructors []string, check bool) (model.ArtifactStatus, error) {
	set, err := registeredSet(path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}
	for _, c := range constructors {
		set[c] = struct{}{}
	}
	return writeGeneratedDatasources(path, set, check)
}

// RemoveGeneratedDatasource deletes constructor from path's
// generatedDatasources slice, the set-difference inverse of
// SyncGeneratedDatasources's union, leaving every other entry intact.
// Idempotent: an already-absent constructor (or a missing file) reports
// Unchanged. Honors check mode.
func RemoveGeneratedDatasource(path, constructor string, check bool) (model.ArtifactStatus, error) {
	set, err := registeredSet(path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}
	if _, ok := set[constructor]; !ok {
		return model.ArtifactStatusUnchanged, nil
	}
	delete(set, constructor)
	return writeGeneratedDatasources(path, set, check)
}

// registeredSet reads the constructor identifiers currently in the
// generatedDatasources file at path into a set; a missing file yields an empty
// set.
func registeredSet(path string) (map[string]struct{}, error) {
	return registeredSetMatching(path, datasourceConstructorRe, identity)
}

// writeGeneratedDatasources renders the generatedDatasources file from a set of
// constructors (sorted, gofmt-canonicalized) and writes it through WriteFile.
func writeGeneratedDatasources(path string, set map[string]struct{}, check bool) (model.ArtifactStatus, error) {
	return writeGeneratedSet(path, generatedDatasourcesHeader, renderIdentifier, set, check)
}

// identity returns s unchanged; it is the unwrap function for a
// registeredSetMatching call whose regex already captures the bare token.
func identity(s string) string { return s }

// renderIdentifier formats a bare Go identifier as one line of a generated
// slice literal.
func renderIdentifier(s string) string {
	return "\t" + s + ",\n"
}

// registeredSetMatching reads the tokens re matches in the file at path,
// unwraps each one, and returns them as a set; a missing file yields an empty
// set. It is the extract half of the extract/union/rewrite round trip the Sync*
// functions perform on a generated slice file.
func registeredSetMatching(path string, re *regexp.Regexp, unwrap func(string) string) (map[string]struct{}, error) {
	set := map[string]struct{}{}
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, m := range re.FindAllString(string(existing), -1) {
		set[unwrap(m)] = struct{}{}
	}
	return set, nil
}

// writeGeneratedSet renders header, one render(key) line per sorted key in
// set, and a closing brace, gofmt-canonicalizes the result, and writes it
// through WriteFile.
func writeGeneratedSet(path, header string, render func(string) string, set map[string]struct{}, check bool) (model.ArtifactStatus, error) {
	var buf bytes.Buffer
	buf.WriteString(header)
	buf.WriteByte('\n')
	for _, k := range sortedKeys(set) {
		buf.WriteString(render(k))
	}
	buf.WriteString("}\n")

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return model.ArtifactStatusFailed, fmt.Errorf("emit: gofmt of %s: %w", path, err)
	}
	return WriteFile(path, src, check)
}

// sortedKeys returns a set's keys as a sorted slice.
func sortedKeys(set map[string]struct{}) []string {
	names := make([]string, 0, len(set))
	for c := range set {
		names = append(names, c)
	}
	sort.Strings(names)
	return names
}

// datasourcesSliceHeader is the line opening the hand-written Datasources slice
// in framework_provider.go. Line removal is scoped to this block so a
// like-named entry in another slice is never touched.
const datasourcesSliceHeader = "var Datasources = []func() datasource.DataSource{"

// RemoveHandwrittenDatasource deletes constructor from the hand-written
// Datasources slice in framework_provider.go (the file at path). The removal is
// scoped to the Datasources block so a like-named Resources entry is never
// touched, and it is idempotent: an already-absent constructor reports
// Unchanged. Check mode does not write.
func RemoveHandwrittenDatasource(path, constructor string, check bool) (model.ArtifactStatus, error) {
	return removeFromSliceBlock(path, datasourcesSliceHeader, "Datasources slice", constructor, check)
}

// removeFromSliceBlock deletes the line holding constructor from the slice
// literal that header opens in the file at path, scoped to that block so a
// like-named entry in another slice is never touched. Idempotent: an
// already-absent constructor reports Unchanged. label names the block in error
// messages; check mode does not write.
func removeFromSliceBlock(path, header, label, constructor string, check bool) (model.ArtifactStatus, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}

	lines := strings.Split(string(original), "\n")
	start, end, err := literalBlockBounds(lines, path, header, label)
	if err != nil {
		return model.ArtifactStatusFailed, err
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

// literalBlockBounds returns the line indices of the slice or map literal that
// header opens and of its closing brace. Entries hold no braces of their own,
// so the first "}" after the header terminates the block. label names the block
// in error messages.
func literalBlockBounds(lines []string, path, header, label string) (start, end int, err error) {
	start = -1
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			break
		}
	}
	if start == -1 {
		return 0, 0, fmt.Errorf("emit: %s: %s not found", path, label)
	}
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "}" {
			return start, i, nil
		}
	}
	return 0, 0, fmt.Errorf("emit: %s: %s is not terminated", path, label)
}

// endpointTagsMapHeader is the line opening the hand-written
// testFiles2EndpointTags map in provider_test.go. Edits are scoped to this
// block so a like-named key in another map is never touched.
const endpointTagsMapHeader = "var testFiles2EndpointTags = map[string]string{"

// EndpointTagTestKey returns the testFiles2EndpointTags key for a generated
// data source's acceptance test: "tests/data_source_datadog_<name>_test". The
// tests/ prefix and missing .go suffix match what getEndpointTagValue keys on
// (it appends "datadog/" and ".go" before comparing).
func EndpointTagTestKey(name string) string {
	return "tests/data_source_datadog_" + name + "_test"
}

// NormalizeEndpointTag lowercases an OpenAPI tag and turns its spaces into
// hyphens ("Cloud Workload Security" -> "cloud-workload-security"). An empty
// tag normalizes to empty; the caller substitutes a fallback.
func NormalizeEndpointTag(tag string) string {
	return strings.ToLower(strings.ReplaceAll(tag, " ", "-"))
}

// endpointTagsBlock returns the line indices of the testFiles2EndpointTags map
// header and its closing brace.
func endpointTagsBlock(lines []string, path string) (start, end int, err error) {
	return literalBlockBounds(lines, path, endpointTagsMapHeader, "testFiles2EndpointTags map")
}

// RegisteredEndpointTags returns the testFiles2EndpointTags entries in path. A
// missing file, or a map written as an empty one-line literal, yields an empty
// map.
func RegisteredEndpointTags(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	emptyMapLine := strings.TrimSuffix(endpointTagsMapHeader, "{") + "{}"
	for _, line := range lines {
		if strings.TrimSpace(line) == emptyMapLine {
			return map[string]string{}, nil
		}
	}
	start, end, err := endpointTagsBlock(lines, path)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{}
	for _, line := range lines[start+1 : end] {
		keyText, valueText, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		key, err := strconv.Unquote(strings.TrimSpace(keyText))
		if err != nil {
			return nil, fmt.Errorf("emit: %s: invalid endpoint-tag key %q: %w", path, keyText, err)
		}
		value, err := strconv.Unquote(strings.TrimSuffix(strings.TrimSpace(valueText), ","))
		if err != nil {
			return nil, fmt.Errorf("emit: %s: invalid endpoint-tag value %q: %w", path, valueText, err)
		}
		tags[key] = value
	}
	return tags, nil
}

// InsertEndpointTag sets `"<key>": "<value>",` in the testFiles2EndpointTags
// map in provider_test.go (the file at path), scoped between
// endpointTagsMapHeader and its closing brace: an existing entry for key is
// rewritten in place, a new key appended just before the brace. The result is
// gofmt-canonicalized, so a re-run rebuilding the identical file is Unchanged.
// A missing file is a hard error.
func InsertEndpointTag(path, key, value string, check bool) (model.ArtifactStatus, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}

	lines := strings.Split(string(original), "\n")
	start, end, err := endpointTagsBlock(lines, path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}

	entry := "\t" + strconv.Quote(key) + ": " + strconv.Quote(value) + ","
	keyPrefix := strconv.Quote(key) + ":"
	out := make([]string, 0, len(lines)+1)
	replaced := false
	for i, line := range lines {
		if i > start && i < end && strings.HasPrefix(strings.TrimSpace(line), keyPrefix) {
			out = append(out, entry)
			replaced = true
			continue
		}
		if i == end && !replaced {
			out = append(out, entry)
		}
		out = append(out, line)
	}

	src, err := format.Source([]byte(strings.Join(out, "\n")))
	if err != nil {
		return model.ArtifactStatusFailed, fmt.Errorf("emit: gofmt of testFiles2EndpointTags: %w", err)
	}
	return WriteFile(path, src, check)
}

// RemoveEndpointTag deletes the entry keyed by key from the
// testFiles2EndpointTags map, scoped to the block. Idempotent: an absent key or
// a missing file reports Unchanged. It gofmt-canonicalizes and honors check
// mode.
func RemoveEndpointTag(path, key string, check bool) (model.ArtifactStatus, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return model.ArtifactStatusUnchanged, nil
		}
		return model.ArtifactStatusFailed, err
	}

	lines := strings.Split(string(original), "\n")
	start, end, err := endpointTagsBlock(lines, path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}

	keyPrefix := strconv.Quote(key) + ":"
	out := make([]string, 0, len(lines))
	removed := false
	for i, line := range lines {
		if i > start && i < end && !removed && strings.HasPrefix(strings.TrimSpace(line), keyPrefix) {
			removed = true
			continue
		}
		out = append(out, line)
	}
	if !removed {
		return model.ArtifactStatusUnchanged, nil
	}

	src, err := format.Source([]byte(strings.Join(out, "\n")))
	if err != nil {
		return model.ArtifactStatusFailed, fmt.Errorf("emit: gofmt of testFiles2EndpointTags: %w", err)
	}
	return WriteFile(path, src, check)
}
