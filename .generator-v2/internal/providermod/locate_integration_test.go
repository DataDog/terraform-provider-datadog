//go:build integration

// SDKPackageDir shells out to `go list` in the provider module, which resolves
// (and on a cold cache downloads) the pinned SDK — not work the default unit
// run should do, and the same reason sdkbind/corroborate_test.go is tagged.
// Root's own specs only read go.mod files, so they stay untagged.

package providermod

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SDKPackageDir", func() {
	It("resolves the pinned datadogV2 package through the provider module", func() {
		// Resolved through the provider precisely because this module does not
		// require the SDK: a `go list` run here would fail (FR-005a's
		// dependency placement).
		dir, err := SDKPackageDir("")
		Expect(err).NotTo(HaveOccurred())
		Expect(dir).To(BeADirectory())
		Expect(filepath.Base(dir)).To(Equal("datadogV2"))
		Expect(filepath.Join(dir, "model_incident_type_attributes.go")).To(BeAnExistingFile())
	})
})
