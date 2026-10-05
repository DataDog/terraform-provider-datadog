package cassette

import (
	"errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// bundleIn builds a renderable bundle rooted at dir.
func bundleIn(dir string) *model.CassetteBundle {
	return &model.CassetteBundle{
		TestPath:    filepath.Join(dir, "resource_datadog_widget_openapi_example_test.go"),
		TestContent: []byte("// " + model.GeneratedMarker + "\npackage test\n"),
		CassettePath: filepath.Join(dir, "cassettes",
			"TestAccDatadogWidgetOpenAPIExample.yaml"),
		CassetteContent: []byte("---\n# " + model.GeneratedMarker + "\nversion: 2\n"),
		FreezePath: filepath.Join(dir, "cassettes",
			"TestAccDatadogWidgetOpenAPIExample.freeze"),
		FreezeContent: []byte("2026-06-25T08:30:50Z"),
	}
}

func collision(err error) *BundleCollisionError {
	var target *BundleCollisionError
	ExpectWithOffset(1, errors.As(err, &target)).To(BeTrue(), "want a *BundleCollisionError, got %v", err)
	return target
}

var _ = Describe("WriteBundle", func() {
	Describe("a wholly absent bundle", func() {
		It("writes all three members and reports them created", func() {
			bundle := bundleIn(GinkgoT().TempDir())
			action, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(Succeed())
			Expect(action).To(Equal(model.CassetteWriteCreated))
			Expect(bundle.Ownership).To(Equal(model.CassetteOwnershipMissing))

			for _, path := range bundle.Paths() {
				Expect(path).To(BeAnExistingFile())
			}
			written, err := os.ReadFile(bundle.FreezePath)
			Expect(err).To(Succeed())
			Expect(string(written)).To(Equal("2026-06-25T08:30:50Z"))
		})

		It("creates the cassette subdirectory it needs", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			Expect(filepath.Join(dir, "cassettes")).NotTo(BeADirectory())
			_, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(Succeed())
			Expect(filepath.Join(dir, "cassettes")).To(BeADirectory())
		})

		It("writes files tooling and humans can read, not a temp file's mode", func() {
			bundle := bundleIn(GinkgoT().TempDir())
			_, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(Succeed())
			info, err := os.Stat(bundle.TestPath)
			Expect(err).To(Succeed())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o644)))
		})

		It("leaves no temporary files behind", func() {
			dir := GinkgoT().TempDir()
			_, err := WriteBundle(bundleIn(dir), WritePolicyMissing)
			Expect(err).To(Succeed())
			for _, root := range []string{dir, filepath.Join(dir, "cassettes")} {
				entries, err := os.ReadDir(root)
				Expect(err).To(Succeed())
				for _, entry := range entries {
					Expect(entry.Name()).NotTo(ContainSubstring(".tfgen-"))
				}
			}
		})

		It("records a content hash per member", func() {
			bundle := bundleIn(GinkgoT().TempDir())
			_, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(Succeed())
			Expect(bundle.ContentHashes).To(HaveLen(3))
			for _, path := range bundle.Paths() {
				Expect(bundle.ContentHashes).To(HaveKey(path))
				Expect(bundle.ContentHashes[path]).To(HaveLen(64))
			}
		})
	})

	Describe("ownership preflight", func() {
		// Regeneration over tfgen's own output is routine, but it is a
		// separate authorization so an unreviewed fixture change cannot ride
		// along with an unrelated run.
		It("declines a bundle it generated earlier", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			_, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(Succeed())

			again := bundleIn(dir)
			action, err := WriteBundle(again, WritePolicyMissing)
			Expect(action).To(Equal(model.CassetteWriteNone))
			Expect(collision(err).Ownership).To(Equal(model.CassetteOwnershipGenerated))
			Expect(err.Error()).To(ContainSubstring("authorized separately"))
		})

		// A recording is evidence the API behaved a certain way; a generated
		// fixture is not. Replacing one is a decision about someone's work.
		It("declines a hand-recorded bundle, naming why", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			Expect(os.MkdirAll(filepath.Dir(bundle.CassettePath), 0o755)).To(Succeed())
			Expect(os.WriteFile(bundle.TestPath, []byte("package test // hand written\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(bundle.CassettePath, []byte("---\nversion: 2\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(bundle.FreezePath, []byte("2020-01-01T00:00:00Z"), 0o644)).To(Succeed())

			action, err := WriteBundle(bundleIn(dir), WritePolicyMissing)
			Expect(action).To(Equal(model.CassetteWriteNone))
			Expect(collision(err).Ownership).To(Equal(model.CassetteOwnershipHandwritten))
			Expect(err.Error()).To(ContainSubstring("recorded against a real org"))

			// The hand-written content must survive untouched.
			kept, err := os.ReadFile(bundle.TestPath)
			Expect(err).To(Succeed())
			Expect(string(kept)).To(ContainSubstring("hand written"))
		})

		// Neither tfgen's ownership nor a maintainer's intent can be read from
		// a half-present bundle, so it is never resolved by guessing.
		It("declines a partial bundle", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			Expect(os.MkdirAll(filepath.Dir(bundle.CassettePath), 0o755)).To(Succeed())
			Expect(os.WriteFile(bundle.CassettePath, bundle.CassetteContent, 0o644)).To(Succeed())

			_, err := WriteBundle(bundleIn(dir), WritePolicyMissing)
			Expect(collision(err).Ownership).To(Equal(model.CassetteOwnershipIncomplete))
			Expect(collision(err).Existing).To(ConsistOf(bundle.CassettePath))
			Expect(err.Error()).To(ContainSubstring("partial"))
		})

		It("declines a bundle whose members disagree about ownership", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			Expect(os.MkdirAll(filepath.Dir(bundle.CassettePath), 0o755)).To(Succeed())
			// Marked test, unmarked cassette: neither a recording nor ours.
			Expect(os.WriteFile(bundle.TestPath, bundle.TestContent, 0o644)).To(Succeed())
			Expect(os.WriteFile(bundle.CassettePath, []byte("---\nversion: 2\n"), 0o644)).To(Succeed())
			Expect(os.WriteFile(bundle.FreezePath, bundle.FreezeContent, 0o644)).To(Succeed())

			_, err := WriteBundle(bundleIn(dir), WritePolicyMissing)
			Expect(collision(err).Ownership).To(Equal(model.CassetteOwnershipIncomplete))
		})

		// The freeze companion is a bare timestamp with no marker of its own,
		// so it must not drag the bundle into Handwritten.
		It("reads the freeze companion's ownership from the cassette beside it", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			_, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(Succeed())

			ownership, _, err := classifyBundle(bundleIn(dir))
			Expect(err).To(Succeed())
			Expect(ownership).To(Equal(model.CassetteOwnershipGenerated))
		})
	})

	Describe("rejected input", func() {
		// A test without its cassette fails as a confusing replay error rather
		// than an obvious absence, so a bundle that cannot be completed is
		// never partially committed.
		It("refuses a bundle that was not fully rendered", func() {
			bundle := bundleIn(GinkgoT().TempDir())
			bundle.CassetteContent = nil
			action, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(action).To(Equal(model.CassetteWriteNone))
			Expect(err).To(MatchError(ContainSubstring("incomplete cassette bundle")))
			Expect(bundle.TestPath).NotTo(BeAnExistingFile())
		})

		// A partially written bundle would classify as Incomplete next run and
		// block regeneration until someone cleaned it by hand, so a failed
		// commit removes whatever landed.
		It("rolls back the members that landed when a later one cannot be written", func() {
			dir := GinkgoT().TempDir()
			bundle := bundleIn(dir)
			// Occupy the freeze path with a directory, so its write fails
			// after the test and cassette have already landed.
			Expect(os.MkdirAll(bundle.FreezePath, 0o755)).To(Succeed())

			_, err := WriteBundle(bundle, WritePolicyMissing)
			Expect(err).To(HaveOccurred())
			Expect(bundle.TestPath).NotTo(BeAnExistingFile())
			Expect(bundle.CassettePath).NotTo(BeAnExistingFile())
		})
	})

	// The generated test is provider source as far as the rest of the
	// generator is concerned, so tfgen's existing ownership predicate has to
	// recognize it rather than the fixtures keeping a private notion.
	It("marks the generated test so emit recognizes it as tfgen's own", func() {
		bundle := bundleIn(GinkgoT().TempDir())
		Expect(emit.IsGeneratedSource(bundle.TestContent)).To(BeTrue())
	})
})
