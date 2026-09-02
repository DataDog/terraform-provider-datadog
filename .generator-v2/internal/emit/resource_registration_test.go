package emit

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("ResourceConstructor", func() {
	It("matches the NewDatadog<SdkName>Resource the resource template emits", func() {
		Expect(ResourceConstructor("thing")).To(Equal("NewDatadogThingResource"))
		Expect(ResourceConstructor("incident_type")).To(Equal("NewDatadogIncidentTypeResource"))
	})
})

var _ = Describe("SyncGeneratedResources", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "resource-registration-sync-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "resources_generated.go")
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("creates the file with the constructors sorted and de-duplicated", func() {
		status, err := SyncGeneratedResources(path, []string{"NewThingResource", "NewAbcResource", "NewThingResource"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		Expect(read()).To(ContainSubstring("\tNewAbcResource,\n\tNewThingResource,\n"))
	})

	It("merges with already-registered constructors instead of replacing them", func() {
		_, err := SyncGeneratedResources(path, []string{"NewThingResource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncGeneratedResources(path, []string{"NewAbcResource"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("NewAbcResource"))
		Expect(read()).To(ContainSubstring("NewThingResource"))
	})

	It("is idempotent: a second identical sync reports Unchanged", func() {
		_, err := SyncGeneratedResources(path, []string{"NewThingResource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncGeneratedResources(path, []string{"NewThingResource"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("renders a canonical empty slice when nothing is registered", func() {
		status, err := SyncGeneratedResources(path, nil, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		Expect(read()).To(ContainSubstring("var generatedResources = []func() resource.Resource{}"))
	})

	It("reports the change in check mode without writing the file", func() {
		status, err := SyncGeneratedResources(path, []string{"NewThingResource"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		_, statErr := os.Stat(path)
		Expect(os.IsNotExist(statErr)).To(BeTrue())
	})
})

var _ = Describe("GeneratedResourceRegistered", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "resource-registration-check-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "resources_generated.go")
	})

	It("reports true for a registered constructor and false for an absent one", func() {
		_, err := SyncGeneratedResources(path, []string{"NewThingResource"}, false)
		Expect(err).NotTo(HaveOccurred())

		registered, err := GeneratedResourceRegistered(path, "NewThingResource")
		Expect(err).NotTo(HaveOccurred())
		Expect(registered).To(BeTrue())

		registered, err = GeneratedResourceRegistered(path, "NewAbsentResource")
		Expect(err).NotTo(HaveOccurred())
		Expect(registered).To(BeFalse())
	})

	It("reports false for a missing file", func() {
		registered, err := GeneratedResourceRegistered(path, "NewThingResource")
		Expect(err).NotTo(HaveOccurred())
		Expect(registered).To(BeFalse())
	})
})

var _ = Describe("RemoveHandwrittenResource", func() {
	const provider = `package fwprovider

var Resources = []func() resource.Resource{
	NewAPIKeyResource,
	NewDatadogThingResource,
	NewHostsResource,
}

var Datasources = []func() datasource.DataSource{
	NewTeamDataSource,
}
`
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "resource-registration-remove-handwritten-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "framework_provider.go")
		Expect(os.WriteFile(path, []byte(provider), 0o644)).To(Succeed())
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("removes the named constructor from the Resources slice", func() {
		status, err := RemoveHandwrittenResource(path, "NewDatadogThingResource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).NotTo(ContainSubstring("NewDatadogThingResource"))
		Expect(read()).To(ContainSubstring("NewAPIKeyResource"))
		Expect(read()).To(ContainSubstring("NewHostsResource"))
	})

	It("is idempotent: removing an already-absent constructor reports Unchanged", func() {
		status, err := RemoveHandwrittenResource(path, "NewNotPresentResource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("leaves a like-named entry outside the Resources block untouched", func() {
		status, err := RemoveHandwrittenResource(path, "NewTeamDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
		Expect(read()).To(ContainSubstring("NewTeamDataSource"))
	})

	It("reports the change in check mode without writing", func() {
		status, err := RemoveHandwrittenResource(path, "NewDatadogThingResource", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("NewDatadogThingResource"))
	})

	It("errors when the file has no Resources slice", func() {
		other := filepath.Join(filepath.Dir(path), "other.go")
		Expect(os.WriteFile(other, []byte("package fwprovider\n"), 0o644)).To(Succeed())
		_, err := RemoveHandwrittenResource(other, "NewDatadogThingResource", false)
		Expect(err).To(HaveOccurred())
	})
})
