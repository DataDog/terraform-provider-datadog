package parser

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// These specs drive the shared selection helpers through the full LoadSpec
// pipeline against the example-form matrix fixture, so the refactor is checked
// by observed parser output rather than by the helpers in isolation.
var _ = Describe("path-item and operation parameter merging", func() {
	var spec *model.Spec

	BeforeEach(func() {
		spec = loadSpecMust("openapi_examples.yaml")
	})

	findParam := func(params []model.QueryParam, name string) *model.QueryParam {
		GinkgoHelper()
		for i := range params {
			if params[i].Name == name {
				return &params[i]
			}
		}
		return nil
	}

	// Without the merge an operation silently loses the path parameter naming
	// the object it acts on, which makes the recorded URL unbuildable.
	It("inherits a path parameter declared once on the path item", func() {
		op := opByID(spec, "GetWithParameters")
		widgetId := findParam(op.PathParams, "widget_id")
		Expect(widgetId).NotTo(BeNil(), "widget_id should be inherited from the path item")
		Expect(widgetId.Required).To(BeTrue())
		Expect(widgetId.In).To(Equal(model.ParameterInPath))
	})

	It("inherits path-item parameters for every operation on the path", func() {
		for _, id := range []string{"GetBodyless", "DeleteBodyless"} {
			op := opByID(spec, id)
			Expect(findParam(op.PathParams, "widget_id")).NotTo(BeNil(), "for %s", id)
		}
	})

	It("lets the operation's declaration win on a (name, location) collision", func() {
		op := opByID(spec, "GetParameterMerge")
		shared := findParam(op.QueryParams, "shared")
		Expect(shared).NotTo(BeNil())
		// The path item declares this parameter without a description; the
		// operation redeclares it with one. The operation must win.
		Expect(shared.Description).To(ContainSubstring("this wins"))
	})

	It("keeps a path-item-only parameter the operation never mentions", func() {
		op := opByID(spec, "GetParameterMerge")
		Expect(findParam(op.QueryParams, "only_path_item")).NotTo(BeNil())
	})

	It("records the location on both path and query parameters", func() {
		op := opByID(spec, "GetWithParameters")
		Expect(findParam(op.PathParams, "widget_id").In).To(Equal(model.ParameterInPath))
		Expect(findParam(op.QueryParams, "tags").In).To(Equal(model.ParameterInQuery))
	})
})

var _ = Describe("parameter serialization metadata", func() {
	var op *model.Operation

	BeforeEach(func() {
		op = opByID(loadSpecMust("openapi_examples.yaml"), "GetWithParameters")
	})

	// Style and explode together decide the recorded request target, so an
	// explicit explode:false must survive parsing rather than being defaulted
	// back to form style's true.
	It("retains a declared style and an explicit explode:false", func() {
		var tags *model.QueryParam
		for i := range op.QueryParams {
			if op.QueryParams[i].Name == "tags" {
				tags = &op.QueryParams[i]
			}
		}
		Expect(tags).NotTo(BeNil())
		Expect(tags.ResolvedStyle()).To(Equal(model.ParameterStyleForm))
		Expect(tags.Explode).NotTo(BeNil(), "explicit explode:false must be retained, not defaulted")
		Expect(tags.ResolvedExplode()).To(BeFalse())
	})

	It("defaults explode to true for a form parameter that omits it", func() {
		var archived *model.QueryParam
		for i := range op.QueryParams {
			if op.QueryParams[i].Name == "include_archived" {
				archived = &op.QueryParams[i]
			}
		}
		Expect(archived).NotTo(BeNil())
		Expect(archived.Explode).To(BeNil(), "the fixture declares no explode here")
		Expect(archived.ResolvedExplode()).To(BeTrue())
	})

	It("defaults a path parameter to simple style", func() {
		var widgetId *model.QueryParam
		for i := range op.PathParams {
			if op.PathParams[i].Name == "widget_id" {
				widgetId = &op.PathParams[i]
			}
		}
		Expect(widgetId).NotTo(BeNil())
		Expect(widgetId.ResolvedStyle()).To(Equal(model.ParameterStyleSimple))
		Expect(widgetId.ResolvedExplode()).To(BeFalse())
	})
})

// The helpers below take libopenapi operations directly, so these specs build
// minimal ones by hand rather than going through a fixture. They cover the
// branches the fixture-driven specs above cannot reach — notably the failure
// and range outcomes DeclaredResponses must sort out, which T018 relies on.

func response(mediaType string, withSchema bool) *v3.Response {
	resp := &v3.Response{}
	if mediaType == "" {
		return resp
	}
	content := orderedmap.New[string, *v3.MediaType]()
	mt := &v3.MediaType{}
	if withSchema {
		mt.Schema = base.CreateSchemaProxy(&base.Schema{Type: []string{"object"}})
	}
	content.Set(mediaType, mt)
	resp.Content = content
	return resp
}

func operationWithResponses(codes map[string]*v3.Response) *v3.Operation {
	ordered := orderedmap.New[string, *v3.Response]()
	// Set in a deliberately unsorted order so the helpers' own ordering is
	// what the assertions observe.
	for _, code := range []string{"429", "404", "201", "200", "204", "4XX", "400"} {
		if resp, ok := codes[code]; ok {
			ordered.Set(code, resp)
		}
	}
	for code, resp := range codes {
		if _, ok := ordered.Get(code); !ok {
			ordered.Set(code, resp)
		}
	}
	return &v3.Operation{Responses: &v3.Responses{Codes: ordered}}
}

var _ = Describe("DeclaredResponses", func() {
	It("returns success and failure outcomes together, in ascending order", func() {
		op := operationWithResponses(map[string]*v3.Response{
			"201": response("application/json", true),
			"404": response("application/json", true),
			"429": response("", false),
		})
		got := DeclaredResponses(op)
		Expect(got).To(HaveLen(3))
		Expect([]string{got[0].Status, got[1].Status, got[2].Status}).
			To(Equal([]string{"201", "404", "429"}))
	})

	// A range names no single status a recorded interaction could carry.
	It("skips a range code such as 4XX", func() {
		op := operationWithResponses(map[string]*v3.Response{
			"200": response("application/json", true),
			"4XX": response("application/json", true),
		})
		got := DeclaredResponses(op)
		Expect(got).To(HaveLen(1))
		Expect(got[0].Status).To(Equal("200"))
	})

	It("reports body presence per outcome", func() {
		op := operationWithResponses(map[string]*v3.Response{
			"200": response("application/json", true),
			"204": response("", false),
		})
		got := DeclaredResponses(op)
		Expect(got[0].Status).To(Equal("200"))
		Expect(got[0].BodyPresent()).To(BeTrue())
		Expect(got[0].MediaType).To(Equal("application/json"))
		Expect(got[1].Status).To(Equal("204"))
		Expect(got[1].BodyPresent()).To(BeFalse())
		Expect(got[1].MediaType).To(BeEmpty())
	})

	It("returns nothing for an operation declaring no responses", func() {
		Expect(DeclaredResponses(nil)).To(BeEmpty())
		Expect(DeclaredResponses(&v3.Operation{})).To(BeEmpty())
	})
})

var _ = Describe("SelectSuccessResponse", func() {
	// A bodyless success carries no schema for the provider model, so it is
	// skipped rather than chosen — otherwise a 204 would mask a usable 200.
	It("skips a bodyless success in favour of one that declares a body", func() {
		op := operationWithResponses(map[string]*v3.Response{
			"200": response("", false),
			"201": response("application/json", true),
		})
		got, ok := SelectSuccessResponse(op)
		Expect(ok).To(BeTrue())
		Expect(got.Status).To(Equal("201"))
	})

	It("ignores a non-JSON body", func() {
		op := operationWithResponses(map[string]*v3.Response{
			"200": response("text/plain", true),
		})
		_, ok := SelectSuccessResponse(op)
		Expect(ok).To(BeFalse())
	})

	It("reports no selection when only failure outcomes are declared", func() {
		op := operationWithResponses(map[string]*v3.Response{
			"400": response("application/json", true),
			"404": response("application/json", true),
		})
		_, ok := SelectSuccessResponse(op)
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("MergeParameters", func() {
	param := func(name, in, description string) *v3.Parameter {
		return &v3.Parameter{Name: name, In: in, Description: description}
	}

	It("puts path-item parameters ahead of operation-only ones", func() {
		got := MergeParameters(
			[]*v3.Parameter{param("widget_id", "path", "")},
			[]*v3.Parameter{param("tags", "query", "")},
		)
		Expect(got).To(HaveLen(2))
		Expect(got[0].Name).To(Equal("widget_id"))
		Expect(got[1].Name).To(Equal("tags"))
	})

	It("lets the operation override the path item in place", func() {
		got := MergeParameters(
			[]*v3.Parameter{param("shared", "query", "from path item"), param("other", "query", "")},
			[]*v3.Parameter{param("shared", "query", "from operation")},
		)
		Expect(got).To(HaveLen(2))
		// Position is the path item's; content is the operation's.
		Expect(got[0].Name).To(Equal("shared"))
		Expect(got[0].Description).To(Equal("from operation"))
		Expect(got[1].Name).To(Equal("other"))
	})

	// (name, location) is the identity, so the same name in two locations is
	// two parameters rather than a collision.
	It("treats the same name in different locations as distinct", func() {
		got := MergeParameters(
			[]*v3.Parameter{param("id", "path", "")},
			[]*v3.Parameter{param("id", "query", "")},
		)
		Expect(got).To(HaveLen(2))
	})

	It("drops nil entries and nameless parameters from either side", func() {
		got := MergeParameters(
			[]*v3.Parameter{nil, param("", "query", ""), param("kept", "query", "")},
			[]*v3.Parameter{nil, param("", "path", "")},
		)
		Expect(got).To(HaveLen(1))
		Expect(got[0].Name).To(Equal("kept"))
	})

	It("keeps only the first of a duplicated path-item key", func() {
		got := MergeParameters(
			[]*v3.Parameter{param("dup", "query", "first"), param("dup", "query", "second")},
			nil,
		)
		Expect(got).To(HaveLen(1))
		Expect(got[0].Description).To(Equal("first"))
	})

	It("returns nothing when both sides are empty", func() {
		Expect(MergeParameters(nil, nil)).To(BeEmpty())
	})
})
