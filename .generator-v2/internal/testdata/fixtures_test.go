//go:build integration

// Behind the integration tag for the same reason as compile_test.go: the gate
// shells out to `go build` in the provider module, which the unit run does not
// guarantee is downloaded.

package testdata

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Asserts what a textual golden diff cannot see: that a committed golden is
// valid Go as a member of package fwprovider, against the SDK the provider
// pins, with no manual edits. Comparing the golden against a fresh generator
// run is the catalogue walker's job and is not duplicated here.
var _ = Describe("Fixture catalogue goldens", func() {
	It("compiles the resource_incident_type golden against the provider module", func() {
		out := filepath.Join("fixtures", "resource_incident_type", "out")

		// The resource is staged under a scratch basename, following stagedName's
		// convention, rather than at its own path. That path holds the
		// hand-written datadog_incident_type, and the fixture carries no
		// `overwrites` (see its README), so taking the path over is not what this
		// golden claims — building additively is. The registration file has no
		// hand-written counterpart and goes to its real path, which is what makes
		// it prove that the constructor it names is the one the resource file
		// declares.
		result, err := Compile([]StagedFile{
			{
				ProviderPath: filepath.Join(stagedPackage, "zz_tfgen_fixture_incident_type.go"),
				SourcePath:   filepath.Join(out, "resource_datadog_incident_type.go"),
			},
			{
				ProviderPath: filepath.Join(stagedPackage, "resources_generated.go"),
				SourcePath:   filepath.Join(out, "resources_generated.go"),
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Diagnostics).To(BeEmpty())
	})
})
