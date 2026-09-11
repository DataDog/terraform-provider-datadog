package emit

import (
	"regexp"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ResourceConstructor returns the exported constructor a generated resource
// declares for artifact name. It matches the New<title GoName>Resource the
// resource template emits, GoName being the Datadog-prefixed dsGoName base.
func ResourceConstructor(name string) string {
	return "New" + upperFirst(dsGoName(name)) + "Resource"
}

// resourceConstructorRe matches a New<...>Resource constructor identifier.
// resources_generated.go holds nothing else that fits the pattern, so it
// safely recovers the already-registered set from the file's current contents.
var resourceConstructorRe = regexp.MustCompile(`New[A-Za-z0-9_]+Resource`)

// GeneratedResourceRegistered reports whether constructor already appears in
// the generatedResources file at path (a missing file reports false).
func GeneratedResourceRegistered(path, constructor string) (bool, error) {
	return generatedRegistered(path, resourceConstructorRe, constructor)
}

// RegisteredGeneratedResources returns the constructor identifiers currently
// registered in the generatedResources file at path, sorted and de-duplicated.
// A missing file yields an empty slice. wireGeneratedResources reads the set
// once to tell an idempotent re-run, where a prior run already retired the
// overwrites target, from a target that never existed.
func RegisteredGeneratedResources(path string) ([]string, error) {
	set, err := registeredResourceSet(path)
	if err != nil {
		return nil, err
	}
	return sortedKeys(set), nil
}

// generatedResourcesHeader is everything in resources_generated.go up to and
// including the slice literal's opening brace. SyncGeneratedResources appends
// the sorted constructors and the closing brace, then gofmt canonicalizes it.
const generatedResourcesHeader = `package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/resource"

// generatedResources holds the resources produced by the generator-v2 emit
// pipeline. tfgen owns this file: each run merges the constructors it produced
// into the existing set (union, sorted) so a scoped --include run never drops
// resources it did not regenerate. Do not edit by hand.
//
// FrameworkProvider.Resources registers this slice alongside the hand-written
// Resources.
var generatedResources = []func() resource.Resource{`

// SyncGeneratedResources rewrites path's generatedResources slice to hold the
// union of the constructors already registered there and the ones passed in,
// sorted and de-duplicated. Merging (rather than replacing) keeps a partial
// --include run from dropping resources it did not regenerate this time. It
// honors check mode through WriteFile.
func SyncGeneratedResources(path string, constructors []string, check bool) (model.ArtifactStatus, error) {
	set, err := registeredResourceSet(path)
	if err != nil {
		return model.ArtifactStatusFailed, err
	}
	for _, c := range constructors {
		set[c] = struct{}{}
	}
	return writeGeneratedResources(path, set, check)
}

// registeredResourceSet reads the constructor identifiers currently in the
// generatedResources file at path into a set; a missing file yields an empty
// set.
func registeredResourceSet(path string) (map[string]struct{}, error) {
	return registeredSetMatching(path, resourceConstructorRe, identity)
}

// writeGeneratedResources renders the generatedResources file from a set of
// constructors (sorted, gofmt-canonicalized) and writes it through WriteFile.
func writeGeneratedResources(path string, set map[string]struct{}, check bool) (model.ArtifactStatus, error) {
	return writeGeneratedSet(path, generatedResourcesHeader, renderIdentifier, set, check)
}

// resourcesSliceHeader is the line opening the hand-written Resources slice in
// framework_provider.go. RemoveHandwrittenResource scopes its line removal to
// this block so a like-named entry in another slice (e.g. Datasources) is
// never touched.
const resourcesSliceHeader = "var Resources = []func() resource.Resource{"

// RemoveHandwrittenResource deletes constructor from the hand-written
// Resources slice in framework_provider.go (the file at path) — the slice a
// generated resource supersedes when its spec sets overwrites. The removal is
// scoped to the Resources block so a like-named Datasources entry is never
// touched, and it is idempotent: an already-absent constructor reports
// Unchanged. It honors check mode by not writing.
func RemoveHandwrittenResource(path, constructor string, check bool) (model.ArtifactStatus, error) {
	return removeFromSliceBlock(path, resourcesSliceHeader, "Resources slice", constructor, check)
}
