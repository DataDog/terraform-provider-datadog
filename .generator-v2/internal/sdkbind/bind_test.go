package sdkbind

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

var placementSpec = filepath.Join("../testdata/sdkbind", "bind_placement.yaml")

// loadBound loads the placement fixture and binds every operation, returning the
// operations keyed by operationId together with each one's binding error.
func loadBound() (map[string]*model.Operation, map[string]error) {
	GinkgoHelper()
	spec, err := parser.LoadSpec(placementSpec)
	Expect(err).To(Succeed())

	ops := make(map[string]*model.Operation, len(spec.Operations))
	errs := make(map[string]error, len(spec.Operations))
	for _, op := range spec.Operations {
		ops[op.OperationId] = op
		if bindErr := BindOperation(op); bindErr != nil {
			errs[op.OperationId] = bindErr
		}
	}
	return ops, errs
}

// unionAt walks a normalized schema down a slash-delimited path of steps, where a
// step is a property name, "[]" for an array element, or "{}" for a map value, and
// returns the OneOfSpec it lands on.
func unionAt(root *model.Schema, steps ...string) *model.OneOfSpec {
	GinkgoHelper()
	node := root
	for _, step := range steps {
		Expect(node).NotTo(BeNil(), "walked off the tree before %q", step)
		switch step {
		case "[]", "{}":
			node = node.Items
		default:
			Expect(node.Properties).To(HaveKey(step))
			node = node.Properties[step]
		}
	}
	Expect(node).NotTo(BeNil())
	Expect(node.Kind).To(Equal(model.SchemaKindOneOf), "node is %s, not a union", node.Kind)
	Expect(node.OneOf).NotTo(BeNil())
	return node.OneOf
}

// variant returns the named alternative of a union.
func variant(spec *model.OneOfSpec, tfName string) model.OneOfVariant {
	GinkgoHelper()
	for _, v := range spec.Variants {
		if v.TFName == tfName {
			return v
		}
	}
	Fail("union " + spec.Path + " has no variant " + tfName)
	return model.OneOfVariant{}
}

var _ = Describe("BindOperation", func() {

	Context("wrapper placement", func() {
		It("binds a component-backed union to its own component name", func() {
			ops, errs := loadBound()
			Expect(errs).NotTo(HaveKey("GetWidget"))

			union := unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "integration")
			Expect(union.SDKType).To(Equal("WidgetIntegration"))
			Expect(union.RefName).To(Equal("WidgetIntegration"))
		})

		It("binds an inline union to the enclosing model plus the property name", func() {
			// child_models names an inline model <parent><CamelProperty>, so the union
			// at WidgetAttributes.inline_choice is WidgetAttributesInlineChoice — a name
			// the Terraform envelope identity (path-derived) would never produce.
			ops, _ := loadBound()
			union := unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "inline_choice")
			Expect(union.SDKType).To(Equal("WidgetAttributesInlineChoice"))
			Expect(union.RefName).To(BeEmpty())
			Expect(union.Name).NotTo(Equal(union.SDKType),
				"the Terraform envelope name must not be reused as the SDK wrapper")
		})

		It("appends Item for an inline union in a list", func() {
			ops, _ := loadBound()
			union := unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "choices", "[]")
			Expect(union.SDKType).To(Equal("WidgetAttributesChoicesItem"))
		})

		It("binds a union at the response root to the response type", func() {
			ops, errs := loadBound()
			Expect(errs).NotTo(HaveKey("GetRoot"))

			root := ops["GetRoot"].ResponseSchema
			Expect(root.Kind).To(Equal(model.SchemaKindOneOf))
			Expect(root.OneOf.SDKType).To(Equal("RootUnion"))
		})

		It("binds a union nested inside another union's alternative", func() {
			ops, _ := loadBound()
			outer := unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "integration")
			aws := variant(outer, "aws_integration")
			inner := unionAt(aws.Schema, "credentials")
			Expect(inner.SDKType).To(Equal("AWSCredentials"))
		})

		It("binds a union on the request side from the request root", func() {
			ops, errs := loadBound()
			Expect(errs).NotTo(HaveKey("CreateWidget"))

			op := ops["CreateWidget"]
			Expect(op.RequestRefName).To(Equal("WidgetCreateRequest"),
				"the request $ref name is the SDK root the request-side walk starts from")
			union := unionAt(op.RequestSchema, "data", "attributes", "payload")
			Expect(union.SDKType).To(Equal("WidgetCreateAttributesPayload"))
		})
	})

	Context("members and constructors", func() {
		It("binds each referenced alternative to its component member and constructor", func() {
			ops, _ := loadBound()
			union := unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "integration")

			aws := variant(union, "aws_integration")
			Expect(aws.SDKField).To(Equal("AWSIntegration"))
			Expect(aws.SDKConstructor).To(Equal("AWSIntegrationAsWidgetIntegration"))
			Expect(aws.SDKPointer).To(BeTrue())

			http := variant(union, "http_integration")
			Expect(http.SDKField).To(Equal("HTTPIntegration"))
			Expect(http.SDKConstructor).To(Equal("HTTPIntegrationAsWidgetIntegration"))
		})

		It("binds a primitive alternative to its upperfirst Go spelling", func() {
			ops, _ := loadBound()
			union := unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "inline_choice")

			str := variant(union, "string")
			Expect(str.SDKField).To(Equal("String"))
			Expect(str.SDKConstructor).To(Equal("StringAsWidgetAttributesInlineChoice"))

			b := variant(union, "boolean")
			Expect(b.SDKField).To(Equal("Bool"))
			Expect(b.SDKConstructor).To(Equal("BoolAsWidgetAttributesInlineChoice"))
		})

		It("emits a constructor for every alternative, never only some", func() {
			// model_oneof.j2 emits <Member>As<Union> unconditionally; treating it as
			// optional silently degrades request mapping to direct member assignment.
			ops, _ := loadBound()
			for _, union := range []*model.OneOfSpec{
				unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "integration"),
				unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "inline_choice"),
				unionAt(ops["GetWidget"].ResponseSchema, "data", "attributes", "choices", "[]"),
				unionAt(ops["CreateWidget"].RequestSchema, "data", "attributes", "payload"),
			} {
				Expect(union.Variants).NotTo(BeEmpty())
				for _, v := range union.Variants {
					Expect(v.SDKConstructor).To(Equal(v.SDKField+"As"+union.SDKType),
						"union %s variant %s", union.Path, v.TFName)
				}
			}
		})

		It("binds an integer alternative to Int32, not Int64", func() {
			ops, _ := loadBound()
			union := unionAt(ops["CreateWidget"].RequestSchema, "data", "attributes", "payload")
			Expect(variant(union, "integer").SDKField).To(Equal("Int32"))
		})
	})

	Context("unresolvable positions", func() {
		It("fails the union whose position no SDK model names", func() {
			_, errs := loadBound()

			err := errs["GetWidgetUnnamed"]
			Expect(err).To(HaveOccurred())

			var unresolved *UnresolvedBindingError
			Expect(errors.As(err, &unresolved)).To(BeTrue())
			Expect(unresolved.Artifact).To(Equal("widget_unnamed"))
			Expect(unresolved.Operation).To(Equal("GetWidgetUnnamed"))
			Expect(unresolved.Failures).To(HaveLen(1))
			Expect(unresolved.Failures[0].Path).To(Equal("response.by_key{}"))
			Expect(unresolved.Failures[0].Reason).To(ContainSubstring("$ref"))
		})

		It("does not fail an unrelated artifact for it", func() {
			// Binding is per operation, so one artifact's unnameable union leaves the
			// others generable.
			_, errs := loadBound()
			Expect(errs).NotTo(HaveKey("GetWidget"))
			Expect(errs).NotTo(HaveKey("CreateWidget"))
			Expect(errs).NotTo(HaveKey("GetRoot"))
		})

		It("names the artifact, operation and union path in its diagnostic", func() {
			_, errs := loadBound()
			msg := errs["GetWidgetUnnamed"].Error()
			Expect(msg).To(ContainSubstring("widget_unnamed"))
			Expect(msg).To(ContainSubstring("GetWidgetUnnamed"))
			Expect(msg).To(ContainSubstring("response.by_key{}"))
		})

		It("leaves no half-formed constructor on an unbound union", func() {
			ops, _ := loadBound()
			union := unionAt(ops["GetWidgetUnnamed"].ResponseSchema, "by_key", "{}")
			Expect(union.SDKType).To(BeEmpty())
			Expect(union.Variants).NotTo(BeEmpty())
			for _, v := range union.Variants {
				Expect(v.SDKConstructor).To(BeEmpty(),
					"variant %s got a constructor without a wrapper to build", v.TFName)
			}
		})
	})

	Context("contract", func() {
		It("is idempotent", func() {
			ops, _ := loadBound()
			op := ops["GetWidget"]
			union := unionAt(op.ResponseSchema, "data", "attributes", "integration")
			before := *union

			Expect(BindOperation(op)).To(Succeed())
			after := unionAt(op.ResponseSchema, "data", "attributes", "integration")
			Expect(after.SDKType).To(Equal(before.SDKType))
			Expect(after.Variants).To(Equal(before.Variants))
		})

		It("tolerates a nil operation and empty schemas", func() {
			Expect(BindOperation(nil)).To(Succeed())
			Expect(BindOperation(&model.Operation{OperationId: "Empty"})).To(Succeed())
		})

		It("binds every operation of a spec through BindSpec", func() {
			spec, err := parser.LoadSpec(placementSpec)
			Expect(err).To(Succeed())

			failed := BindSpec(spec)
			var failedIDs []string
			for op := range failed {
				failedIDs = append(failedIDs, op.OperationId)
			}
			Expect(failedIDs).To(ConsistOf("GetWidgetUnnamed"))

			for _, op := range spec.Operations {
				if op.OperationId != "GetRoot" {
					continue
				}
				Expect(op.ResponseSchema.OneOf.SDKType).To(Equal("RootUnion"))
			}
		})
	})
})
