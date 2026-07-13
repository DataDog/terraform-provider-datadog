package emit

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("DatasourceConstructor", func() {
	It("matches the NewDatadog<SdkName>DataSource the data-source template emits", func() {
		Expect(DatasourceConstructor("team")).To(Equal("NewDatadogTeamDataSource"))
		Expect(DatasourceConstructor("datastore_item")).To(Equal("NewDatadogDatastoreItemDataSource"))
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

var _ = Describe("RemoveGeneratedDatasource", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "registration-remove-generated-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "datasources_generated.go")
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("drops one constructor and leaves the rest of the slice intact", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewAbcDataSource", "NewTeamDataSource", "NewZooDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := RemoveGeneratedDatasource(path, "NewTeamDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).NotTo(ContainSubstring("NewTeamDataSource"))
		Expect(read()).To(ContainSubstring("NewAbcDataSource"))
		Expect(read()).To(ContainSubstring("NewZooDataSource"))
	})

	It("renders the canonical empty slice when the last constructor is removed", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := RemoveGeneratedDatasource(path, "NewTeamDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("var generatedDatasources = []func() datasource.DataSource{}"))
	})

	It("is idempotent: removing an already-absent constructor reports Unchanged", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := RemoveGeneratedDatasource(path, "NewAbsentDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("reports Unchanged when the file does not exist", func() {
		status, err := RemoveGeneratedDatasource(path, "NewTeamDataSource", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("reports the change in check mode without writing", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewAbcDataSource", "NewTeamDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		status, err := RemoveGeneratedDatasource(path, "NewTeamDataSource", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring("NewTeamDataSource"))
	})
})

var _ = Describe("RegisteredGeneratedDatasources", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "registration-registered-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "datasources_generated.go")
	})

	It("returns the registered constructors sorted", func() {
		_, err := SyncGeneratedDatasources(path, []string{"NewZooDataSource", "NewAbcDataSource"}, false)
		Expect(err).NotTo(HaveOccurred())

		got, err := RegisteredGeneratedDatasources(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{"NewAbcDataSource", "NewZooDataSource"}))
	})

	It("returns an empty slice for a missing file", func() {
		got, err := RegisteredGeneratedDatasources(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
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

var _ = Describe("EndpointTagTestKey", func() {
	It("builds the tests/-prefixed, no-.go key getEndpointTagValue matches on", func() {
		Expect(EndpointTagTestKey("gcp_uc_configs")).To(Equal("tests/data_source_datadog_gcp_uc_configs_test"))
	})
})

var _ = Describe("NormalizeEndpointTag", func() {
	It("lowercases and turns spaces into hyphens like the existing map values", func() {
		Expect(NormalizeEndpointTag("Cloud Workload Security")).To(Equal("cloud-workload-security"))
		Expect(NormalizeEndpointTag("integration-aws")).To(Equal("integration-aws"))
		Expect(NormalizeEndpointTag("")).To(Equal(""))
	})
})

// endpointTagsFixture is a provider_test.go-shaped file: the target map plus a
// decoy map so the scoped edit can be shown not to touch a like-named key elsewhere.
const endpointTagsFixture = `package test

var otherMap = map[string]string{
	"tests/data_source_datadog_decoy_test": "decoy",
}

var testFiles2EndpointTags = map[string]string{
	"tests/data_source_datadog_team_test":             "team",
	"tests/data_source_datadog_team_memberships_test": "team",
}
`

var _ = Describe("InsertEndpointTag", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "registration-insert-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "provider_test.go")
		Expect(os.WriteFile(path, []byte(endpointTagsFixture), 0o644)).To(Succeed())
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("appends a new entry into the testFiles2EndpointTags map", func() {
		status, err := InsertEndpointTag(path, "tests/data_source_datadog_gcp_uc_configs_test", "cloud-cost", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		// gofmt pads the value column, so match key and value without fixed spacing.
		Expect(read()).To(MatchRegexp(`"tests/data_source_datadog_gcp_uc_configs_test":\s+"cloud-cost",`))
	})

	It("is idempotent: re-inserting the same entry reports Unchanged", func() {
		_, err := InsertEndpointTag(path, "tests/data_source_datadog_gcp_uc_configs_test", "cloud-cost", false)
		Expect(err).NotTo(HaveOccurred())
		status, err := InsertEndpointTag(path, "tests/data_source_datadog_gcp_uc_configs_test", "cloud-cost", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("rewrites an existing key's value in place without duplicating it", func() {
		status, err := InsertEndpointTag(path, "tests/data_source_datadog_team_test", "teams", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(MatchRegexp(`"tests/data_source_datadog_team_test":\s+"teams",`))
		Expect(strings.Count(read(), `"tests/data_source_datadog_team_test":`)).To(Equal(1))
	})

	It("leaves a like-named key in another map untouched", func() {
		_, err := InsertEndpointTag(path, "tests/data_source_datadog_new_test", "svc", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(read()).To(ContainSubstring(`"tests/data_source_datadog_decoy_test": "decoy",`))
	})

	It("reports the change in check mode without writing", func() {
		status, err := InsertEndpointTag(path, "tests/data_source_datadog_new_test", "svc", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).NotTo(ContainSubstring("data_source_datadog_new_test"))
	})

	It("errors when the file has no testFiles2EndpointTags map", func() {
		other := filepath.Join(filepath.Dir(path), "other.go")
		Expect(os.WriteFile(other, []byte("package test\n"), 0o644)).To(Succeed())
		_, err := InsertEndpointTag(other, "tests/data_source_datadog_new_test", "svc", false)
		Expect(err).To(HaveOccurred())
	})

	It("errors when the file is missing", func() {
		_, err := InsertEndpointTag(filepath.Join(filepath.Dir(path), "absent.go"), "tests/data_source_datadog_new_test", "svc", false)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("RemoveEndpointTag", func() {
	var path string

	BeforeEach(func() {
		dir, err := os.MkdirTemp("", "registration-remove-tag-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		path = filepath.Join(dir, "provider_test.go")
		Expect(os.WriteFile(path, []byte(endpointTagsFixture), 0o644)).To(Succeed())
	})

	read := func() string {
		content, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		return string(content)
	}

	It("removes the named entry from the map", func() {
		status, err := RemoveEndpointTag(path, "tests/data_source_datadog_team_test", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).NotTo(ContainSubstring(`"tests/data_source_datadog_team_test":`))
		Expect(read()).To(ContainSubstring(`"tests/data_source_datadog_team_memberships_test":`))
	})

	It("is idempotent: removing an already-absent key reports Unchanged", func() {
		status, err := RemoveEndpointTag(path, "tests/data_source_datadog_absent_test", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("tolerates a missing file, reporting Unchanged", func() {
		status, err := RemoveEndpointTag(filepath.Join(filepath.Dir(path), "absent.go"), "tests/data_source_datadog_team_test", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUnchanged))
	})

	It("reports the change in check mode without writing", func() {
		status, err := RemoveEndpointTag(path, "tests/data_source_datadog_team_test", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal(model.ArtifactStatusUpdated))
		Expect(read()).To(ContainSubstring(`"tests/data_source_datadog_team_test":`))
	})
})
