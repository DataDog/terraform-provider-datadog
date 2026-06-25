package emit

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("DatasourceConstructor", func() {
	It("matches the New<SdkName>DataSource the data-source template emits", func() {
		Expect(DatasourceConstructor("team")).To(Equal("NewTeamDataSource"))
		Expect(DatasourceConstructor("datastore_item")).To(Equal("NewDatastoreItemDataSource"))
	})
})

var _ = Describe("SyncGeneratedDatasources", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "registration-sync-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "datasources_generated.go")
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("creates the file with the constructors sorted and de-duplicated", func() {
		status, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource", "NewAbcDataSource", "NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		Expect(read()).To(ContainSubstring("\tNewAbcDataSource,\n\tNewTeamDataSource,\n"))
	})

	It("merges with already-registered constructors instead of replacing them", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncGeneratedDatasources(path, []string{"NewAbcDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("NewAbcDataSource"))
		Expect(read()).To(ContainSubstring("NewTeamDataSource"))
	})

	It("is idempotent: a second identical sync reports Unchanged", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("renders a canonical empty slice when nothing is registered", func() {
		status, err := SyncGeneratedDatasources(path, nil, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		Expect(read()).To(ContainSubstring("var generatedDatasources = []func() datasource.DataSource{}"))
	})

	It("reports the change in check mode without writing the file", func() {
		status, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource"}, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusCreated))
		_, statErr := os.Stat(path)
		Expect(os.IsNotExist(statErr)).To(BeTrue())
	})
})

var _ = Describe("RemoveHandwrittenDatasource", func() {
	const provider = `package fwprovider

var Resources = []func() resource.Resource{
	NewTeamResource,
}

var Datasources = []func() datasource.DataSource{
	NewAPIKeyDataSource,
	NewDatadogTeamDataSource,
	NewHostsDataSource,
}
`
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "registration-remove-*")
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

	It("removes the named constructor from the Datasources slice", func() {
		status, err := RemoveHandwrittenDatasource(path, "NewDatadogTeamDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).NotTo(ContainSubstring("NewDatadogTeamDataSource"))
		Expect(read()).To(ContainSubstring("NewAPIKeyDataSource"))
		Expect(read()).To(ContainSubstring("NewHostsDataSource"))
	})

	It("is idempotent: removing an already-absent constructor reports Unchanged", func() {
		status, err := RemoveHandwrittenDatasource(path, "NewNotPresentDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("leaves a like-named entry outside the Datasources block untouched", func() {
		status, err := RemoveHandwrittenDatasource(path, "NewTeamResource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
		Expect(read()).To(ContainSubstring("NewTeamResource"))
	})

	It("reports the change in check mode without writing", func() {
		status, err := RemoveHandwrittenDatasource(path, "NewDatadogTeamDataSource", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("NewDatadogTeamDataSource"))
	})

	It("errors when the file has no Datasources slice", func() {
		other := filepath.Join(filepath.Dir(path), "other.go")
		Expect(os.WriteFile(other, []byte("package fwprovider\n"), 0o644)).To(Succeed())
		_, err := RemoveHandwrittenDatasource(other, "NewDatadogTeamDataSource", false)
		Expect(err).To(HaveOccurred())
	})
})
