package emit

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// component wraps an object schema with the OpenAPI component name that supplied
// it, which is what the parser records as Schema.RefName.
func component(name string, props map[string]*model.Schema) *model.Schema {
	s := obj(props)
	s.RefName = name
	return s
}

// twoTagFiltersOperation reduces the shape that made integration_aws_external_id
// uncompilable: two differently-shaped objects, each reachable under a property
// named "tag_filters", one nested a level deeper than the other. Deriving a model
// name from the property alone gave both TagFiltersModel.
func twoTagFiltersOperation() *model.Operation {
	return &model.Operation{
		Path:            "/api/v2/integration/aws/accounts/{aws_account_config_id}",
		Method:          "GET",
		OperationId:     "GetAWSAccount",
		Tag:             "AWS Integration",
		ResponseRefName: "AWSAccountResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindDataSource,
			ArtifactName: "integration_aws_account",
			IdStrategy:   model.IdStrategyDataID,
			Group:        &model.OperationGroup{Read: "GetAWSAccount"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The account's identifier."),
				"type": prim("string", "Resource type."),
				"attributes": obj(map[string]*model.Schema{
					"logs_config": component("AWSLogsConfig", map[string]*model.Schema{
						"log_source_config": component("AWSLogSourceConfig", map[string]*model.Schema{
							"tag_filters": {
								Kind:        model.SchemaKindArray,
								Description: "Log source tag filters.",
								Items: component("AWSLogSourceTagFilter", map[string]*model.Schema{
									"source": prim("string", "The log source."),
								}),
							},
						}),
					}),
					"metrics_config": component("AWSMetricsConfig", map[string]*model.Schema{
						"tag_filters": {
							Kind:        model.SchemaKindArray,
							Description: "Metrics namespace tag filters.",
							Items: component("AWSNamespaceTagFilter", map[string]*model.Schema{
								"namespace": prim("string", "The AWS namespace."),
							}),
						},
					}),
				}),
			}),
		}),
	}
}

// sharedComponentOperation reaches one component from two properties in one
// artifact. Both sites must converge on a single declaration rather than
// redeclaring the type. The properties deliberately avoid created_by/updated_by,
// which flattenEnvelope drops as server-managed audit fields.
func sharedComponentOperation() *model.Operation {
	creator := func() *model.Schema {
		return component("Creator", map[string]*model.Schema{
			"email": prim("string", "The creator's email."),
			"name":  prim("string", "The creator's name."),
		})
	}
	return &model.Operation{
		Path:            "/api/v2/widgets/{widget_id}",
		Method:          "GET",
		OperationId:     "GetWidget",
		Tag:             "Widgets",
		ResponseRefName: "WidgetResponse",
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindDataSource,
			ArtifactName: "widget",
			IdStrategy:   model.IdStrategyDataID,
			Group:        &model.OperationGroup{Read: "GetWidget"},
		},
		ResponseSchema: obj(map[string]*model.Schema{
			"data": obj(map[string]*model.Schema{
				"id":   prim("string", "The widget's identifier."),
				"type": prim("string", "Resource type."),
				"attributes": obj(map[string]*model.Schema{
					"primary_contact":   creator(),
					"secondary_contact": creator(),
				}),
			}),
		}),
	}
}

func modelNames(view DataSourceView) []string {
	names := make([]string, 0, len(view.Models))
	for _, m := range view.Models {
		names = append(names, m.Name)
	}
	return names
}

var _ = Describe("generated model naming", func() {

	Context("two objects under one property name", func() {
		It("names each after its own component instead of colliding", func() {
			view, err := BuildDataSourceView(mustArtifact(twoTagFiltersOperation()))
			Expect(err).NotTo(HaveOccurred())

			names := modelNames(view)
			Expect(names).To(ContainElement("datadogIntegrationAwsAccountAWSLogSourceTagFilterModel"))
			Expect(names).To(ContainElement("datadogIntegrationAwsAccountAWSNamespaceTagFilterModel"))
		})

		It("declares every model exactly once", func() {
			view, err := BuildDataSourceView(mustArtifact(twoTagFiltersOperation()))
			Expect(err).NotTo(HaveOccurred())

			seen := map[string]int{}
			for _, n := range modelNames(view) {
				seen[n]++
			}
			for name, count := range seen {
				Expect(count).To(Equal(1), "%s is declared %d times; the file would not compile", name, count)
			}
		})

		It("points each tag_filters field at its own struct", func() {
			view, err := BuildDataSourceView(mustArtifact(twoTagFiltersOperation()))
			Expect(err).NotTo(HaveOccurred())

			logSource := modelByName(view, "datadogIntegrationAwsAccountAWSLogSourceConfigModel")
			Expect(logSource.Fields).To(Equal([]ModelFieldView{
				{GoField: "TagFilters", GoType: "[]*datadogIntegrationAwsAccountAWSLogSourceTagFilterModel", TFName: "tag_filters"},
			}))

			metrics := modelByName(view, "datadogIntegrationAwsAccountAWSMetricsConfigModel")
			Expect(metrics.Fields).To(Equal([]ModelFieldView{
				{GoField: "TagFilters", GoType: "[]*datadogIntegrationAwsAccountAWSNamespaceTagFilterModel", TFName: "tag_filters"},
			}))
		})

		It("renders gofmt-clean Go", func() {
			view, err := BuildDataSourceView(mustArtifact(twoTagFiltersOperation()))
			Expect(err).NotTo(HaveOccurred())
			// RenderDataSource runs format.Source, so a duplicate declaration or a
			// mangled identifier fails here rather than at the provider's build.
			_, err = RenderDataSource(view)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("scoping to the artifact", func() {
		It("prefixes every model with the artifact base", func() {
			// Two data sources sharing a component must not each declare it: they land
			// in one package, so an unprefixed name is a redeclaration across files.
			// 54 components in the mini-OAS corpus appear in more than one slice.
			view, err := BuildDataSourceView(mustArtifact(twoTagFiltersOperation()))
			Expect(err).NotTo(HaveOccurred())

			for _, n := range modelNames(view) {
				Expect(n).To(HavePrefix("datadogIntegrationAwsAccount"),
					"%s is not scoped to its artifact", n)
			}
		})

		It("keeps nested models unexported, as the hand-written data sources do", func() {
			view, err := BuildDataSourceView(mustArtifact(twoTagFiltersOperation()))
			Expect(err).NotTo(HaveOccurred())

			for _, n := range modelNames(view) {
				Expect(n[:1]).To(Equal(strings.ToLower(n[:1])), "%s is exported", n)
			}
		})
	})

	Context("one component reached twice", func() {
		It("collapses to a single declaration both sites point at", func() {
			view, err := BuildDataSourceView(mustArtifact(sharedComponentOperation()))
			Expect(err).NotTo(HaveOccurred())

			var creators int
			for _, n := range modelNames(view) {
				if n == "datadogWidgetCreatorModel" {
					creators++
				}
			}
			Expect(creators).To(Equal(1))

			root := modelByName(view, "datadogWidgetDataSourceModel")
			var types []string
			for _, f := range root.Fields {
				if f.TFName == "primary_contact" || f.TFName == "secondary_contact" {
					types = append(types, f.GoType)
				}
			}
			Expect(types).To(Equal([]string{"*datadogWidgetCreatorModel", "*datadogWidgetCreatorModel"}))
		})
	})

	DescribeTable("nestedStem",
		func(stem string, a *model.Attribute, want string) {
			Expect(nestedStem(stem, a)).To(Equal(want))
		},
		Entry("prefers the component name, verbatim",
			"Parent", &model.Attribute{Path: "response.tag_filters", ModelRefName: "AWSLogSourceTagFilter"},
			"AWSLogSourceTagFilter"),
		Entry("restarts the accumulation at the component, ignoring the enclosing stem",
			"LogsConfigLogSourceConfig", &model.Attribute{Path: "response.tag_filters", ModelRefName: "AWSNamespaceTagFilter"},
			"AWSNamespaceTagFilter"),
		Entry("accumulates from the enclosing stem for an inline schema",
			"LogSourceConfig", &model.Attribute{Path: "response.logs.log_source_config.tag_filters"},
			"LogSourceConfigTagFilters"),
		Entry("converts a snake_case property, since only a component name is already Go-style",
			"", &model.Attribute{Path: "response.http_token_auth"},
			"HttpTokenAuth"),
	)

	Describe("dedupeModels", func() {
		It("collapses identical declarations without complaint", func() {
			fields := []ModelFieldView{{GoField: "Name", GoType: "types.String", TFName: "name"}}
			out, conflicts := dedupeModels([]ModelStructView{
				{Name: "aModel", Fields: fields},
				{Name: "aModel", Fields: fields},
			})
			Expect(conflicts).To(BeEmpty())
			Expect(out).To(HaveLen(1))
		})

		It("preserves first-use order", func() {
			out, conflicts := dedupeModels([]ModelStructView{
				{Name: "rootModel"}, {Name: "aModel"}, {Name: "rootModel"}, {Name: "bModel"},
			})
			Expect(conflicts).To(BeEmpty())
			Expect([]string{out[0].Name, out[1].Name, out[2].Name}).To(Equal([]string{"rootModel", "aModel", "bModel"}))
		})

		It("fails when one name covers two shapes, naming both", func() {
			// The guard applied to artifact names has no counterpart at the Go
			// identifier level, so this is the emitter's own defence: emitting the
			// redeclaration or silently dropping one shape are both worse.
			out, conflicts := dedupeModels([]ModelStructView{
				{Name: "tagFiltersModel", Fields: []ModelFieldView{{GoField: "Source", GoType: "types.String", TFName: "source"}}},
				{Name: "tagFiltersModel", Fields: []ModelFieldView{{GoField: "Namespace", GoType: "types.String", TFName: "namespace"}}},
			})
			Expect(out).To(HaveLen(1))
			Expect(conflicts).To(HaveLen(1))
			Expect(conflicts[0].Path).To(Equal("tagFiltersModel"))
			Expect(conflicts[0].Reason).To(ContainSubstring("Source types.String"))
			Expect(conflicts[0].Reason).To(ContainSubstring("Namespace types.String"))
			Expect(conflicts[0].Reason).To(ContainSubstring("redeclare"))
		})
	})
})
