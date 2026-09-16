package emit

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("SyncUnstableOperations", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "unstable-operations-sync-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "unstable_operations_generated.go")
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("creates the file with the keys sorted and de-duplicated", func() {
		status, err := SyncUnstableOperations(path,
			[]string{"v2.GetThing", "v2.CreateThing", "v2.GetThing"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		Expect(read()).To(ContainSubstring("\t\"v2.CreateThing\",\n\t\"v2.GetThing\",\n"))
	})

	// The merge is the whole point: the SDK disables an operation absent from the
	// slice, so a scoped --include run that replaced rather than merged would
	// silently break every artifact it did not regenerate.
	It("merges with already-registered keys instead of replacing them", func() {
		_, err := SyncUnstableOperations(path, []string{"v2.GetThing"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncUnstableOperations(path, []string{"v2.CreateOther"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("v2.GetThing"))
		Expect(read()).To(ContainSubstring("v2.CreateOther"))
	})

	It("is idempotent: a second identical sync reports Unchanged", func() {
		_, err := SyncUnstableOperations(path, []string{"v2.GetThing"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncUnstableOperations(path, []string{"v2.GetThing"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("reports the change in check mode without writing the file", func() {
		status, err := SyncUnstableOperations(path, []string{"v2.GetThing"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		Expect(path).NotTo(BeAnExistingFile())
	})

	// A version other than v2 has to survive the round trip, since the key is
	// what SetUnstableOperationEnabled matches on.
	It("recovers keys of any API version from an existing file", func() {
		_, err := SyncUnstableOperations(path, []string{"v1.GetLegacy", "v2.GetThing"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncUnstableOperations(path, []string{"v1.GetLegacy", "v2.GetThing"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})
})
