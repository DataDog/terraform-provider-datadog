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

var _ = Describe("ApplyAPIAccessor", func() {
	accessors := map[string]string{"RUMApi": "GetRumApiV2"}

	It("overrides the accessor when the provider names it differently", func() {
		view := &DataSourceView{APIStruct: "RUMApi", APIAccessor: "GetRUMApiV2"}
		ApplyAPIAccessor(view, accessors)
		Expect(view.APIAccessor).To(Equal("GetRumApiV2"))
	})

	It("leaves the derived accessor when the struct name matches the accessor base", func() {
		view := &DataSourceView{APIStruct: "TeamsApi", APIAccessor: "GetTeamsApiV2"}
		ApplyAPIAccessor(view, accessors)
		Expect(view.APIAccessor).To(Equal("GetTeamsApiV2"))
	})

	It("leaves the derived accessor when no accessor map was resolved", func() {
		view := &DataSourceView{APIStruct: "RUMApi", APIAccessor: "GetRUMApiV2"}
		ApplyAPIAccessor(view, nil)
		Expect(view.APIAccessor).To(Equal("GetRUMApiV2"))
	})
})
