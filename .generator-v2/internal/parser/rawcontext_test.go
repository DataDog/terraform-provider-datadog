package parser

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func serverVariables(pairs map[string]string) *orderedmap.Map[string, *v3.ServerVariable] {
	if pairs == nil {
		return nil
	}
	m := orderedmap.New[string, *v3.ServerVariable]()
	for name, def := range pairs {
		m.Set(name, &v3.ServerVariable{Default: def})
	}
	return m
}

var _ = Describe("resolveServerURL", func() {
	// Cassette generation records request URLs against this origin, so it must
	// match what the generated client actually calls.
	It("substitutes each template variable with its declared default", func() {
		got := resolveServerURL([]*v3.Server{{
			URL:       "https://{subdomain}.{site}",
			Variables: serverVariables(map[string]string{"subdomain": "api", "site": "datadoghq.com"}),
		}})
		Expect(got).To(Equal("https://api.datadoghq.com"))
	})

	It("returns a concrete URL unchanged", func() {
		Expect(resolveServerURL([]*v3.Server{{URL: "https://api.datadoghq.com"}})).
			To(Equal("https://api.datadoghq.com"))
	})

	It("trims a trailing slash so recorded URLs do not double it", func() {
		Expect(resolveServerURL([]*v3.Server{{URL: "https://api.datadoghq.com/"}})).
			To(Equal("https://api.datadoghq.com"))
	})

	// Guessing at an unresolvable placeholder would produce a plausible-looking
	// wrong host; skipping to a concrete server keeps the failure obvious.
	It("skips a server whose variables have no defaults for a concrete one", func() {
		got := resolveServerURL([]*v3.Server{
			{URL: "https://{unresolved}.example.com"},
			{URL: "https://api.datadoghq.com"},
		})
		Expect(got).To(Equal("https://api.datadoghq.com"))
	})

	It("returns empty when nothing resolves", func() {
		Expect(resolveServerURL(nil)).To(BeEmpty())
		Expect(resolveServerURL([]*v3.Server{{URL: ""}})).To(BeEmpty())
		Expect(resolveServerURL([]*v3.Server{{URL: "https://{nope}.example.com"}})).To(BeEmpty())
	})

	// A declared variable with no default cannot be substituted, so the
	// placeholder survives and the server is skipped rather than half-resolved.
	It("skips a variable that declares no default", func() {
		vars := orderedmap.New[string, *v3.ServerVariable]()
		vars.Set("site", &v3.ServerVariable{Default: ""})
		vars.Set("missing", nil)
		got := resolveServerURL([]*v3.Server{
			{URL: "https://api.{site}", Variables: vars},
			{URL: "https://api.datadoghq.com"},
		})
		Expect(got).To(Equal("https://api.datadoghq.com"))
	})

	It("skips a nil server entry", func() {
		got := resolveServerURL([]*v3.Server{nil, {URL: "https://api.datadoghq.com"}})
		Expect(got).To(Equal("https://api.datadoghq.com"))
	})

	It("tolerates a server with no variables map at all", func() {
		got := resolveServerURL([]*v3.Server{{URL: "https://api.datadoghq.com", Variables: nil}})
		Expect(got).To(Equal("https://api.datadoghq.com"))
	})
})

var _ = Describe("RawContext", func() {
	Describe("Raw and MergedParameters", func() {
		It("returns the libopenapi operation behind a normalized one", func() {
			op := &model.Operation{OperationId: "GetThing"}
			rawOp := &v3.Operation{OperationId: "GetThing"}
			ctx := newRawContext()
			ctx.add(op, rawOp, nil)

			got, ok := ctx.Raw(op)
			Expect(ok).To(BeTrue())
			Expect(got).To(BeIdenticalTo(rawOp))

			_, ok = ctx.Raw(&model.Operation{})
			Expect(ok).To(BeFalse())
		})

		It("stores path-item parameters only when the path item declares some", func() {
			op := &model.Operation{OperationId: "GetThing"}
			ctx := newRawContext()
			ctx.add(op, &v3.Operation{}, &v3.PathItem{})
			Expect(ctx.PathItemParams).To(BeEmpty())

			ctx.add(op, &v3.Operation{}, &v3.PathItem{
				Parameters: []*v3.Parameter{{Name: "thing_id", In: "path"}},
			})
			Expect(ctx.PathItemParams).To(HaveLen(1))
		})

		// Both schema normalization and example extraction read parameters
		// through here, so they cannot disagree about what an operation takes.
		It("merges path-item and operation parameters with the operation winning", func() {
			op := &model.Operation{OperationId: "GetThing"}
			ctx := newRawContext()
			ctx.add(op,
				&v3.Operation{Parameters: []*v3.Parameter{{Name: "shared", In: "query", Description: "op"}}},
				&v3.PathItem{Parameters: []*v3.Parameter{
					{Name: "shared", In: "query", Description: "path item"},
					{Name: "thing_id", In: "path"},
				}})

			merged := ctx.MergedParameters(op)
			Expect(merged).To(HaveLen(2))
			Expect(merged[0].Name).To(Equal("shared"))
			Expect(merged[0].Description).To(Equal("op"))
			Expect(merged[1].Name).To(Equal("thing_id"))
		})

		It("returns nothing for an operation it does not know", func() {
			Expect(newRawContext().MergedParameters(&model.Operation{})).To(BeEmpty())
			var nilCtx *RawContext
			Expect(nilCtx.MergedParameters(&model.Operation{})).To(BeEmpty())
		})
	})
})

var _ = Describe("Spec.ServerURL", func() {
	// The Datadog descriptions declare a templated default server; the
	// resolved origin must be the host the cassettes were recorded against.
	It("is resolved from the fixture's declared servers", func() {
		spec := loadSpecMust("openapi_examples.yaml")
		Expect(spec.ServerURL).To(BeEmpty(), "the matrix fixture declares no servers")

		twilio, err := LoadSpec("../testdata/mini-oas/mini-datadog_integration_twilio_account.yaml")
		Expect(err).To(Succeed())
		Expect(twilio.ServerURL).To(Equal("https://api.datadoghq.com"))
	})
})
