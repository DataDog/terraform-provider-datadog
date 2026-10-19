package parser

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// -------------------------------------------------------------------
//  Helpers
// -------------------------------------------------------------------

func loadComponents(fixture string) *v3.Components {
	GinkgoHelper()
	data, err := os.ReadFile(filepath.Join("../testdata/parser", fixture))
	Expect(err).To(Succeed(), "read fixture %s", fixture)
	doc, err := libopenapi.NewDocument(data)
	Expect(err).To(Succeed(), "parse fixture %s", fixture)
	model, err := doc.BuildV3Model()
	Expect(err).To(Succeed(), "build v3 model for %s", fixture)
	return model.Model.Components
}

func componentProxy(comps *v3.Components, name string) *base.SchemaProxy {
	GinkgoHelper()
	for k, v := range comps.Schemas.FromOldest() {
		if k == name {
			return v
		}
	}
	Fail("component " + name + " not found")
	return nil
}

// -------------------------------------------------------------------
//  cycleWalker unit tests
// -------------------------------------------------------------------

var _ = Describe("cycleWalker", func() {

	Context("enter and leave", func() {
		It("maintains the stack and depth across a matched enter/leave pair", func() {
			w := newCycleWalker(0)

			Expect(w.enter("a", true)).To(BeTrue())
			Expect(w.onStack["a"]).To(BeTrue())
			Expect(w.stack).To(HaveLen(1))
			Expect(w.depth).To(Equal(1))

			w.leave("a", true, true)
			Expect(w.onStack["a"]).To(BeFalse())
			Expect(w.stack).To(BeEmpty())
			Expect(w.depth).To(Equal(0))
			Expect(w.done["a"]).To(BeTrue())
		})

		It("does not mark a ref done when leave is called with completed=false", func() {
			w := newCycleWalker(0)
			Expect(w.enter("a", true)).To(BeTrue())
			w.leave("a", false, true)

			Expect(w.done["a"]).To(BeFalse())
			Expect(w.depth).To(Equal(0))
			Expect(w.stack).To(BeEmpty())
		})

		It("seed entry does not consume a depth slot", func() {
			w := newCycleWalker(2)

			Expect(w.enter("seed", false)).To(BeTrue())
			Expect(w.depth).To(Equal(0), "seed must not increment depth")

			// Two real edges fit exactly within maxDepth=2 because the seed is free.
			Expect(w.enter("a", true)).To(BeTrue())
			Expect(w.enter("b", true)).To(BeTrue())
			_, err := w.enter("c", true)
			Expect(err).To(HaveOccurred(), "third real edge should exceed --max-depth 2")
		})
	})

	Context("depth bounding", func() {
		It("returns a *MaxDepthError naming the ref and bound when exceeded", func() {
			w := newCycleWalker(1)
			Expect(w.enter("a", true)).To(BeTrue())

			entered, err := w.enter("b", true)
			Expect(entered).To(BeFalse())
			Expect(err).To(HaveOccurred())

			var depthErr *MaxDepthError
			Expect(err).To(BeAssignableToTypeOf(depthErr))
			// unwrap manually so we can inspect fields
			Expect(err.(*MaxDepthError).MaxDepth).To(Equal(1))
			Expect(err.(*MaxDepthError).Ref).To(Equal("b"))
			Expect(w.stack).To(HaveLen(1), "a failed enter must not push")
		})

		It("never errors when the depth bound is 0 (disabled)", func() {
			w := newCycleWalker(0)
			for _, ref := range []string{"a", "b", "c", "d", "e"} {
				Expect(w.enter(ref, true)).To(BeTrue(), "enter %s", ref)
			}
		})

		It("reports a cycle rather than a depth error on re-entry at the bound", func() {
			w := newCycleWalker(1)
			Expect(w.enter("a", true)).To(BeTrue())

			entered, err := w.enter("a", true)
			Expect(entered).To(BeFalse())
			Expect(err).To(Succeed(), "re-entry at the depth bound must be a cycle, not an error")
			Expect(w.cycles).To(HaveLen(1))
		})
	})

	Context("cycle recording", func() {
		It("records a cycle when re-entering a ref already on the stack", func() {
			w := newCycleWalker(0)
			Expect(w.enter("a", true)).To(BeTrue())
			Expect(w.enter("b", true)).To(BeTrue())

			entered, err := w.enter("a", true)
			Expect(entered).To(BeFalse())
			Expect(err).To(Succeed())

			Expect(w.cycles).To(HaveLen(1))
			Expect(w.cycles[0].Ref).To(Equal("a"))
			Expect(w.cycles[0].Path).To(Equal([]string{"a", "b", "a"}))
		})

		It("uses the sub-path from the actual cycle root when closing mid-stack", func() {
			w := newCycleWalker(0)
			Expect(w.enter("a", true)).To(BeTrue())
			Expect(w.enter("b", true)).To(BeTrue())
			Expect(w.enter("c", true)).To(BeTrue())

			// c -> b closes a cycle on b, not a; the path must start at b.
			_, _ = w.enter("b", true)
			Expect(w.cycles).To(HaveLen(1))
			Expect(w.cycles[0].Ref).To(Equal("b"))
			Expect(w.cycles[0].Path).To(Equal([]string{"b", "c", "b"}))
		})

		It("deduplicates repeated recordCycle calls for the same ref", func() {
			w := newCycleWalker(0)
			Expect(w.enter("a", true)).To(BeTrue())

			w.recordCycle("a")
			w.recordCycle("a")
			Expect(w.cycles).To(HaveLen(1))
		})
	})

	Context("pruning", func() {
		It("skips a ref already marked done without pushing or recording a cycle", func() {
			w := newCycleWalker(0)
			w.done["x"] = true

			entered, err := w.enter("x", true)
			Expect(entered).To(BeFalse())
			Expect(err).To(Succeed())
			Expect(w.stack).To(BeEmpty())
			Expect(w.cycles).To(BeEmpty())
		})
	})
})

// -------------------------------------------------------------------
//  DetectComponentRefCycles fixture tests
// -------------------------------------------------------------------

var _ = Describe("DetectComponentRefCycles", func() {

	It("detects a self-referential cycle with the correct ref and path", func() {
		cycles, err := DetectComponentRefCycles(loadComponents("cycle_self.yaml"), 16)
		Expect(err).To(Succeed())
		Expect(cycles).To(HaveLen(1))

		ref := componentSchemaPrefix + "Node"
		Expect(cycles[0].Ref).To(Equal(ref))
		Expect(cycles[0].Path).To(Equal([]string{ref, ref}))
	})

	It("detects an indirect A→B→A cycle with the correct path", func() {
		cycles, err := DetectComponentRefCycles(loadComponents("cycle_indirect.yaml"), 16)
		Expect(err).To(Succeed())
		Expect(cycles).To(HaveLen(1))

		// Components iterate in document order (A before B).
		Expect(cycles[0].Path).To(Equal([]string{
			componentSchemaPrefix + "A",
			componentSchemaPrefix + "B",
			componentSchemaPrefix + "A",
		}))
	})

	It("reports no cycles for an acyclic diamond graph", func() {
		cycles, err := DetectComponentRefCycles(loadComponents("acyclic_diamond.yaml"), 16)
		Expect(err).To(Succeed())
		Expect(cycles).To(BeEmpty())
	})

	It("still finds cycles when the depth bound is disabled (0)", func() {
		cycles, err := DetectComponentRefCycles(loadComponents("cycle_self.yaml"), 0)
		Expect(err).To(Succeed())
		Expect(cycles).To(HaveLen(1))
	})

	DescribeTable("detects a single cycle reached via",
		func(fixture, wantComponent string) {
			cycles, err := DetectComponentRefCycles(loadComponents(fixture), 16)
			Expect(err).To(Succeed())
			Expect(cycles).To(HaveLen(1))
			Expect(cycles[0].Ref).To(Equal(componentSchemaPrefix + wantComponent))
		},
		Entry("array items edge", "cycle_array_items.yaml", "Tree"),
		Entry("allOf edge", "cycle_allof.yaml", "Loop"),
		Entry("additionalProperties edge", "cycle_additional_properties.yaml", "Dict"),
		Entry("oneOf edge", "cycle_oneof.yaml", "Tree"),
		Entry("anyOf edge", "cycle_anyof.yaml", "Graph"),
		Entry("not edge", "cycle_not.yaml", "Excluded"),
		Entry("prefixItems edge", "cycle_prefixitems.yaml", "Tuple"),
	)
})

// -------------------------------------------------------------------
//  DetectRefCycles fixture tests
// -------------------------------------------------------------------

var _ = Describe("DetectRefCycles", func() {

	Context("on an 8-deep acyclic chain (S0→…→S8)", func() {
		var root *base.SchemaProxy

		BeforeEach(func() {
			root = componentProxy(loadComponents("deep_chain.yaml"), "S0")
		})

		It("errors when the chain exceeds the depth limit", func() {
			_, err := DetectRefCycles(root, 4)
			Expect(err).To(HaveOccurred())
		})

		It("succeeds when depth equals chain length", func() {
			cycles, err := DetectRefCycles(root, 8)
			Expect(err).To(Succeed())
			Expect(cycles).To(BeEmpty())
		})

		It("never errors when the depth bound is disabled (0)", func() {
			_, err := DetectRefCycles(root, 0)
			Expect(err).To(Succeed())
		})
	})
})
