//go:build integration

package emit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	testinfra "github.com/terraform-providers/terraform-provider-datadog/generator/internal/testdata"
)

var _ = ginkgo.Describe("generated nested schemas", func() {
	ginkgo.It("loads protocol schemas for plural results at every depth", func() {
		// Exercise the production schema templates and builder output against the
		// provider's pinned framework, without requiring synthetic SDK models.
		var source bytes.Buffer
		source.WriteString(`package fwprovider
import (
 "context"
 "testing"
 "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
 "github.com/hashicorp/terraform-plugin-framework/datasource"
 "github.com/hashicorp/terraform-plugin-framework/provider"
 "github.com/hashicorp/terraform-plugin-framework/providerserver"
 "github.com/hashicorp/terraform-plugin-framework/resource"
 "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)
type tfgenPresenceProvider struct { schema schema.Schema }
func (p *tfgenPresenceProvider) Metadata(_ context.Context, _ provider.MetadataRequest, r *provider.MetadataResponse) { r.TypeName = "tfgen" }
func (p *tfgenPresenceProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {}
func (p *tfgenPresenceProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {}
func (p *tfgenPresenceProvider) Resources(context.Context) []func() resource.Resource { return nil }
func (p *tfgenPresenceProvider) DataSources(context.Context) []func() datasource.DataSource {
 return []func() datasource.DataSource{func() datasource.DataSource { return &tfgenPresenceDataSource{schema: p.schema} }}
}
type tfgenPresenceDataSource struct { schema schema.Schema }
func (d *tfgenPresenceDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, r *datasource.MetadataResponse) { r.TypeName = "tfgen_presence" }
func (d *tfgenPresenceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, r *datasource.SchemaResponse) { r.Schema = d.schema }
func (d *tfgenPresenceDataSource) Read(context.Context, datasource.ReadRequest, *datasource.ReadResponse) {}
func tfgenPresenceErrors(t *testing.T, s schema.Schema) []*tfprotov6.Diagnostic {
 t.Helper()
 response, err := providerserver.NewProtocol6(&tfgenPresenceProvider{schema: s})().GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
 if err != nil { t.Fatal(err) }
 var errors []*tfprotov6.Diagnostic
 for _, d := range response.Diagnostics { if d.Severity == tfprotov6.DiagnosticSeverityError { errors = append(errors, d) } }
 return errors
}
func TestTfgenPluralPresence(t *testing.T) {
 // ValidateImplementation does not check presence; protocol conversion must.
 bad := schema.Schema{Attributes: map[string]schema.Attribute{"results": schema.ListNestedAttribute{
  NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{"value": schema.StringAttribute{Computed: true}}},
 }}}
 if d := tfgenPresenceErrors(t, bad); len(d) == 0 { t.Fatal("missing presence negative control unexpectedly loaded") }

`)
		for _, view := range []DataSourceView{pluralNestedView(), pluralObjectView(), mustView(pluralDeepPresenceOperation()), datastoresView()} {
			fmt.Fprintf(&source, "t.Run(%q, func(t *testing.T) {\ns := schema.Schema{Attributes: map[string]schema.Attribute{\n", view.TypeName)
			for _, a := range view.Schema.Attributes {
				gomega.Expect(templates.ExecuteTemplate(&source, "schemaAttribute", a)).To(gomega.Succeed())
			}
			for _, a := range view.Schema.Blocks {
				gomega.Expect(templates.ExecuteTemplate(&source, "schemaNestedAttribute", a)).To(gomega.Succeed())
			}
			source.WriteString("}}\nif d := tfgenPresenceErrors(t, s); len(d) != 0 { t.Fatalf(\"invalid generated schema: %v\", d) }\n})\n")
		}
		source.WriteString("}\n")
		path := filepath.Join(ginkgo.GinkgoT().TempDir(), "zz_tfgen_presence_test.go")
		gomega.Expect(os.WriteFile(path, source.Bytes(), 0o644)).To(gomega.Succeed())
		result, err := testinfra.RunTests([]testinfra.StagedFile{{ProviderPath: "datadog/fwprovider/zz_tfgen_presence_test.go", SourcePath: path}}, "^TestTfgenPluralPresence$")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(result.OK()).To(gomega.BeTrue(), result.Diagnostics)
	})
})
