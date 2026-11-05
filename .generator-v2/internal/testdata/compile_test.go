//go:build integration

// The compile gate shells out to `go build` in the *provider* module, so it
// needs that module's dependency graph in the module cache — which the unit run
// does not guarantee: .github/workflows/tfgen.yml caches only
// .generator-v2/go.sum and never runs `go mod download` in the provider root,
// and Compile sets GOPROXY=off for hermeticity (FR-027). Behind the same tag as
// sdkbind/corroborate_test.go, which is tagged for exactly this reason, and
// wired up by T080's `go test -tags=integration` step.

package testdata

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs exercise the gate itself, with Go source the spec controls rather
// than generator output. That is deliberate: the emit package's resource
// fixtures are synthetic — incidentTypeResourceOperation invents an
// internal_note field the real IncidentTypeAttributes has no setter for — so
// compiling one would assert the fixture's fidelity, not the gate's behaviour.
// Compiling real generated artifacts is T042's, once the fixture catalogue
// (T075/T078) supplies slices of the actual spec.
//
// stagedName is a scratch basename inside the package under test. It is
// additive — no hand-written provider file is named this — so the gate proves
// the staged file's own validity rather than a replacement's side effects.
const (
	stagedPackage = "datadog/fwprovider"
	stagedName    = "zz_tfgen_compile_gate_probe.go"
)

func stage(content string) []StagedFile {
	GinkgoHelper()
	source := filepath.Join(GinkgoT().TempDir(), stagedName)
	Expect(os.WriteFile(source, []byte(content), 0o644)).To(Succeed())
	return []StagedFile{{
		ProviderPath: filepath.Join(stagedPackage, stagedName),
		SourcePath:   source,
	}}
}

var _ = Describe("Compile", func() {
	It("accepts a file that compiles against the provider module's own dependencies", func() {
		// The imports are the point: the SDK and the plugin framework resolve
		// only through the provider's go.mod, which the generator module does
		// not require, and utils resolves only because the file is compiled as
		// a member of package fwprovider (FR-019).
		result, err := Compile(stage(`package fwprovider

import (
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func zzTfgenCompileGateProbe() types.String {
	body := datadogV2.NewIncidentTypeCreateRequestWithDefaults()
	_ = body
	_ = utils.ResourceIDAttribute()
	return types.StringValue("probe")
}
`))

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Diagnostics).To(BeEmpty())
		Expect(result.OK()).To(BeTrue())
	})

	It("reports the compiler's diagnostic instead of an error when the staged file is broken", func() {
		// A missing SDK setter is exactly the shape a wrong derivation produces
		// (FR-005a's detection point, and the T134 failure this gate would have
		// caught): the name is plausible and only the compiler knows better.
		result, err := Compile(stage(`package fwprovider

import "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

func zzTfgenCompileGateProbe() {
	attributes := datadogV2.NewIncidentTypeUpdateAttributesWithDefaults()
	attributes.SetNoSuchField("x")
}
`))

		By("a build the staged file loses is a verdict, not a gate failure")
		Expect(err).NotTo(HaveOccurred())
		Expect(result.OK()).To(BeFalse())
		By("and the diagnostic is returned verbatim, so a caller can assert on it (FR-021)")
		Expect(result.Diagnostics).To(ContainSubstring("SetNoSuchField undefined"))
		Expect(result.Diagnostics).To(ContainSubstring(stagedName))
	})

	It("fails the gate, rather than blaming the artifact, when a staged source is missing", func() {
		_, err := Compile([]StagedFile{{
			ProviderPath: filepath.Join(stagedPackage, stagedName),
			SourcePath:   filepath.Join(GinkgoT().TempDir(), "never-written.go"),
		}})
		Expect(err).To(MatchError(ContainSubstring("never-written.go")))
	})

	It("rejects a staged path that is not relative to the provider module root", func() {
		_, err := Compile([]StagedFile{{ProviderPath: "/etc/passwd.go", SourcePath: "x.go"}})
		Expect(err).To(MatchError(ContainSubstring("must be relative to the provider module root")))
	})

	It("fails the gate when nothing is staged", func() {
		_, err := Compile(nil)
		Expect(err).To(MatchError(ContainSubstring("no files staged")))
	})
})
