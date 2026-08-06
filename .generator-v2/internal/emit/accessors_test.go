package emit

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveAPIAccessors", func() {
	// A helper stub carrying the shapes that matter: a plain accessor, two whose
	// names diverge from the struct (RUM, Observability Pipelines), a V1 accessor,
	// and a non-accessor method.
	const fixture = `package utils

func (i *ApiInstances) GetTeamsApiV2() *datadogV2.TeamsApi { return nil }
func (i *ApiInstances) GetRumApiV2() *datadogV2.RUMApi { return nil }
func (i *ApiInstances) GetObsPipelinesV2() *datadogV2.ObservabilityPipelinesApi { return nil }
func (i *ApiInstances) GetUsersApiV1() *datadogV1.UsersApi { return nil }
func (i *ApiInstances) HttpClient() {}
`

	var path string
	BeforeEach(func() {
		path = filepath.Join(GinkgoT().TempDir(), "api_instances_helper.go")
		Expect(os.WriteFile(path, []byte(fixture), 0o644)).To(Succeed())
	})

	It("maps each V2 API struct to its accessor, including diverging acronym and alias names", func() {
		m, err := ResolveAPIAccessors(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(m).To(HaveKeyWithValue("TeamsApi", "GetTeamsApiV2"))
		Expect(m).To(HaveKeyWithValue("RUMApi", "GetRumApiV2"))
		Expect(m).To(HaveKeyWithValue("ObservabilityPipelinesApi", "GetObsPipelinesV2"))
	})

	It("ignores V1 accessors and non-accessor methods", func() {
		m, err := ResolveAPIAccessors(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(m).NotTo(HaveKey("UsersApi"))
		Expect(m).To(HaveLen(3))
	})

	It("returns an error on an unparseable file", func() {
		bad := filepath.Join(GinkgoT().TempDir(), "bad.go")
		Expect(os.WriteFile(bad, []byte("package x\nfunc ("), 0o644)).To(Succeed())
		_, err := ResolveAPIAccessors(bad)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("ResolveAPIConstructors", func() {
	const fixture = `package datadogV2

func NewCaseManagementApi(client *datadog.APIClient) *CaseManagementApi { return nil }
func NewRUMApi(client *datadog.APIClient) *RUMApi { return nil }
func NewWrongClient(client *other.APIClient) *WrongClientApi { return nil }
func NewWrongResult(client *datadog.APIClient) WrongResultApi { return WrongResultApi{} }
func NewModel(client *datadog.APIClient) *Model { return nil }
`

	var dir string
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "api_fixture.go"), []byte(fixture), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "model_fixture.go"), []byte("package datadogV2\nfunc NewIgnoredApi(client *datadog.APIClient) *IgnoredApi { return nil }\n"), 0o644)).To(Succeed())
	})

	It("maps API structs to verified shared-client constructors and preserves SDK acronym spelling", func() {
		constructors, err := ResolveAPIConstructors(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(constructors).To(Equal(map[string]string{
			"CaseManagementApi": "NewCaseManagementApi",
			"RUMApi":            "NewRUMApi",
		}))
	})

	It("returns an error when the SDK package cannot be read", func() {
		_, err := ResolveAPIConstructors(filepath.Join(dir, "missing"))
		Expect(err).To(MatchError(ContainSubstring("read SDK API package")))
	})
})

var _ = Describe("ApplyAPIAccessor", func() {
	accessors := map[string]string{"RUMApi": "GetRumApiV2"}
	constructors := map[string]string{
		"RUMApi":            "NewRUMApi",
		"CaseManagementApi": "NewCaseManagementApi",
	}

	It("prefers a discovered accessor when the provider names it differently", func() {
		view := &DataSourceView{APIStruct: "RUMApi", APIAccessor: "GetRUMApiV2"}
		Expect(ApplyAPIAccessor(view, accessors, constructors)).To(Succeed())
		Expect(view.APIAccessor).To(Equal("GetRumApiV2"))
		Expect(view.APIConstructor).To(BeEmpty())
	})

	It("uses a discovered SDK constructor when no helper accessor exists", func() {
		view := &DataSourceView{
			SDKPackage:  "datadogV2",
			APIStruct:   "CaseManagementApi",
			APIAccessor: "GetCaseManagementApiV2",
		}
		Expect(ApplyAPIAccessor(view, accessors, constructors)).To(Succeed())
		Expect(view.APIAccessor).To(BeEmpty())
		Expect(view.APIConstructor).To(Equal("NewCaseManagementApi"))
	})

	It("fails instead of retaining a guessed accessor when neither path resolves", func() {
		view := &DataSourceView{
			SDKPackage:  "datadogV2",
			APIStruct:   "MissingApi",
			APIAccessor: "GetMissingApiV2",
		}
		err := ApplyAPIAccessor(view, accessors, constructors)
		Expect(err).To(MatchError(And(
			ContainSubstring("datadogV2.MissingApi"),
			ContainSubstring("no ApiInstances accessor or SDK constructor"),
		)))
		Expect(view.APIAccessor).To(BeEmpty())
		Expect(view.APIConstructor).To(BeEmpty())
	})
})
