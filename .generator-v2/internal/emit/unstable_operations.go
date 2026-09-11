package emit

import (
	"regexp"
	"strconv"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// unstableOperationRe matches one quoted "v<n>.<OperationId>" entry. The
// generated file holds nothing else shaped like it, so the already-registered
// set can be recovered from the file's current contents.
var unstableOperationRe = regexp.MustCompile(`"v[0-9]+\.[A-Za-z0-9_]+"`)

// generatedUnstableOperationsHeader is everything in
// unstable_operations_generated.go up to the slice literal's opening brace.
// SyncUnstableOperations appends the sorted keys and the closing brace, then
// gofmt canonicalizes it.
const generatedUnstableOperationsHeader = `package fwprovider

// generatedUnstableOperations lists the x-unstable operations the generated
// artifacts call. The pinned SDK defaults every one of them to disabled, so an
// artifact naming an operation absent from this slice compiles and registers
// and then fails every call at runtime. tfgen owns this file: each run merges
// the operations it produced into the existing set (union, sorted) so a scoped
// --include run never drops entries another artifact needs. Do not edit by hand.
//
// enrichFrameworkProviderConfig enables each of these on the client config.
var generatedUnstableOperations = []string{`

// SyncUnstableOperations rewrites path's generatedUnstableOperations slice to
// hold the union of the keys already registered there and the ones passed in,
// sorted and de-duplicated. Merging rather than replacing keeps a partial run
// from disabling operations it did not regenerate. Honors check mode.
func SyncUnstableOperations(path string, keys []string, check bool) (model.ArtifactStatus, error) {
	set, err := registeredSetMatching(path, unstableOperationRe, unquote)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}
	for _, k := range keys {
		set[k] = struct{}{}
	}
	return writeGeneratedSet(path, generatedUnstableOperationsHeader, renderQuoted, set, check)
}

// unquote strips the surrounding quotes unstableOperationRe always matches.
func unquote(s string) string { return s[1 : len(s)-1] }

// renderQuoted formats a quoted string as one line of a generated slice
// literal.
func renderQuoted(s string) string {
	return "\t" + strconv.Quote(s) + ",\n"
}
