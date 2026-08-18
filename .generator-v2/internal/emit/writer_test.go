package emit

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("WriteFile", func() {
	var dir string

	BeforeEach(func() {
		var err error
		dir, err = os.MkdirTemp("", "writer-test-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
	})

	path := func(name string) string { return filepath.Join(dir, name) }

	Describe("write mode", func() {
		It("creates a new file and returns Created", func() {
			p := path("new.go")
			status, err := WriteFile(p, []byte("hello"), false)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusCreated))
			Expect(os.ReadFile(p)).To(Equal([]byte("hello")))
		})

		It("returns Unchanged and leaves the file alone when content matches", func() {
			p := path("same.go")
			Expect(os.WriteFile(p, []byte("hello"), 0o644)).To(Succeed())
			status, err := WriteFile(p, []byte("hello"), false)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusUnchanged))
			Expect(os.ReadFile(p)).To(Equal([]byte("hello")))
		})

		It("overwrites a file with changed content and returns Updated", func() {
			p := path("changed.go")
			Expect(os.WriteFile(p, []byte("old"), 0o644)).To(Succeed())
			status, err := WriteFile(p, []byte("new"), false)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusUpdated))
			Expect(os.ReadFile(p)).To(Equal([]byte("new")))
		})

		It("creates parent directories as needed", func() {
			p := filepath.Join(dir, "sub", "dir", "file.go")
			status, err := WriteFile(p, []byte("hello"), false)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusCreated))
			Expect(os.ReadFile(p)).To(Equal([]byte("hello")))
		})
	})

	Describe("check mode", func() {
		It("returns Created without creating the file", func() {
			p := path("check-new.go")
			status, err := WriteFile(p, []byte("hello"), true)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusCreated))
			_, statErr := os.Stat(p)
			Expect(os.IsNotExist(statErr)).To(BeTrue())
		})

		It("returns Unchanged without touching the file", func() {
			p := path("check-same.go")
			Expect(os.WriteFile(p, []byte("hello"), 0o644)).To(Succeed())
			status, err := WriteFile(p, []byte("hello"), true)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusUnchanged))
		})

		It("returns Updated without writing the file", func() {
			p := path("check-changed.go")
			Expect(os.WriteFile(p, []byte("old"), 0o644)).To(Succeed())
			status, err := WriteFile(p, []byte("new"), true)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(model.ArtifactStatusUpdated))
			Expect(os.ReadFile(p)).To(Equal([]byte("old")))
		})
	})

})
