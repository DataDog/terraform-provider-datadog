package providermod

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Root", func() {
	It("walks past the generator's own module to the provider's", func() {
		// The decisive case: .generator-v2/go.mod is the nearest go.mod to
		// every one of these specs, and it is the wrong answer — it declares
		// the generator module and requires no SDK.
		root, err := Root("")
		Expect(err).NotTo(HaveOccurred())

		goMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(goMod)).To(ContainSubstring(moduleDecl))
		Expect(string(goMod)).To(ContainSubstring(sdkModule))
		Expect(filepath.Base(root)).NotTo(Equal(".generator-v2"))
		Expect(root).To(BeADirectory())
	})

	It("falls back to the working directory when the hint resolves nothing", func() {
		// A hint is where the caller is writing output, which is routinely a
		// temp dir outside any module; that must not be fatal.
		fromHint, err := Root(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		fromCwd, err := Root("")
		Expect(err).NotTo(HaveOccurred())
		Expect(fromHint).To(Equal(fromCwd))
	})

	It("prefers a hint that does sit inside the provider module", func() {
		root, err := Root("")
		Expect(err).NotTo(HaveOccurred())
		fromInside, err := Root(filepath.Join(root, "datadog", "fwprovider"))
		Expect(err).NotTo(HaveOccurred())
		Expect(fromInside).To(Equal(root))
	})
})
