package model

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// ----------------------------------------------------------------------------
// OpenAPI example contract
//
// These types carry the values an OpenAPI description declares for an
// operation's inputs and outputs. They are kept separate from the schema model
// in types.go because only cassette generation consumes them: provider-schema
// normalization cares about an operation's shape, cassette generation cares
// about concrete values it can replay. Nothing here retains a libopenapi
// handle, so a scenario can be described entirely from this package.
// ----------------------------------------------------------------------------

// ExampleSourceKind records which OpenAPI location declared a candidate. The
// kinds form a fallback order rather than a flat set: a media example beats a
// schema example, which beats per-property examples, because a whole-body
// example is something a human wrote for this operation while assembled
// property values are the generator's reconstruction.
type ExampleSourceKind string

const (
	// ExampleSourceParameter is a path or query parameter example.
	ExampleSourceParameter ExampleSourceKind = "parameter"
	// ExampleSourceMedia is a request/response media-type example, the most
	// specific and therefore most trusted form.
	ExampleSourceMedia ExampleSourceKind = "media"
	// ExampleSourceSchema is a whole-body example declared on the schema.
	ExampleSourceSchema ExampleSourceKind = "schema"
	// ExampleSourceSchemaProperty is a single property's example, only ever
	// meaningful assembled with its siblings.
	ExampleSourceSchemaProperty ExampleSourceKind = "schema_property"
)

// ExampleComponent names the part of an operation a location belongs to, so a
// diagnostic can say "the request body" or "the 404 response" without the
// reader decoding a path expression.
type ExampleComponent string

const (
	ExampleComponentParameter   ExampleComponent = "parameter"
	ExampleComponentRequestBody ExampleComponent = "request_body"
	ExampleComponentResponse    ExampleComponent = "response"
)

// ExampleLocation is a stable source anchor for one example declaration. It
// exists so every selection choice and every ineligibility reason can name
// exactly where it came from: reviewers trace committed fixture bytes back to
// the OpenAPI description through this, and it is the only provenance the
// cassette pipeline keeps once libopenapi is released.
//
// Anchors are addresses, never values. A location is safe to log even when the
// candidate it points at is sensitive.
type ExampleLocation struct {
	// Path is the OpenAPI path template, e.g. /api/v2/users/{user_id}.
	Path string
	// Method is the HTTP method, uppercase.
	Method string
	// OperationId is the OpenAPI operationId.
	OperationId string
	// Component is the part of the operation this anchor addresses.
	Component ExampleComponent
	// Detail qualifies Component: the parameter name for a parameter, the
	// status code for a response, empty for a request body.
	Detail string
	// MediaType is the media type, e.g. application/json. Empty for a
	// parameter, whose examples are not media-scoped.
	MediaType string
	// Field is the OpenAPI field that declared the example: "example" for the
	// singular form, "examples" for the named map, or the schema keyword for a
	// fallback.
	Field string
	// ExampleName is the key within a named examples map; empty for the
	// singular form.
	ExampleName string
	// PropertyPath is the dotted path to the property for a schema-property
	// candidate, e.g. data.attributes.name. Empty for whole-body candidates.
	PropertyPath string
}

// String renders the anchor in the same "spec:"-prefixed form the run report
// uses for Diagnostic.Location, so cassette diagnostics read like every other
// diagnostic the generator emits.
func (l ExampleLocation) String() string {
	var b strings.Builder
	b.WriteString("spec:")
	b.WriteString(l.Path)
	if l.Method != "" {
		b.WriteString(".")
		b.WriteString(strings.ToLower(l.Method))
	}
	switch l.Component {
	case ExampleComponentParameter:
		b.WriteString(".parameters.")
		b.WriteString(l.Detail)
	case ExampleComponentRequestBody:
		b.WriteString(".requestBody")
	case ExampleComponentResponse:
		b.WriteString(".responses.")
		b.WriteString(l.Detail)
	}
	if l.MediaType != "" {
		b.WriteString(".content.")
		b.WriteString(l.MediaType)
	}
	if l.Field != "" {
		b.WriteString(".")
		b.WriteString(l.Field)
	}
	if l.ExampleName != "" {
		b.WriteString(".")
		b.WriteString(l.ExampleName)
	}
	if l.PropertyPath != "" {
		b.WriteString(" (")
		b.WriteString(l.PropertyPath)
		b.WriteString(")")
	}
	return b.String()
}

// ExampleCandidate is one value declared by one OpenAPI example location.
//
// A candidate is a claim, not a decision: extraction keeps every candidate it
// finds, including ones it cannot use, because an ineligible candidate is the
// evidence a diagnostic needs. Selection happens later and separately.
type ExampleCandidate struct {
	// Name is the named-example key, empty for a singular example.
	Name string
	// Value is the decoded example. Nil is meaningful: it is a declared JSON
	// null, not absence — absence is the candidate not existing at all.
	Value any
	// SourceKind records which OpenAPI location declared this value.
	SourceKind ExampleSourceKind
	// Location anchors the declaration for diagnostics and provenance.
	Location ExampleLocation
	// Reference is the local reusable-example reference the value was resolved
	// through, e.g. #/components/examples/ReusableWidget. Empty when the value
	// was declared inline. Kept so a reviewer can find the shared declaration
	// rather than the use site.
	Reference string
	// External records that the declaration used externalValue. The URL is
	// never fetched: a generated fixture must be reproducible from the spec
	// alone, and fetching would make generation depend on network state.
	External bool
	// Sensitive records that the declaring schema classified this value as
	// write-only, secret, or credential material. A sensitive candidate is
	// usable only after replacement and its value must never reach a
	// diagnostic or the run report.
	Sensitive bool
}

// Eligible reports whether this candidate can contribute a value to a
// generated cassette. An external candidate is retained for diagnostics but is
// never eligible, because its value is not in the description.
func (c *ExampleCandidate) Eligible() bool {
	return c != nil && !c.External
}

// IneligibleReason explains a non-eligible candidate in reviewer-facing terms,
// or returns the empty string when the candidate is usable. It never includes
// the candidate's value, which may be sensitive.
func (c *ExampleCandidate) IneligibleReason() string {
	switch {
	case c == nil:
		return "no example is declared"
	case c.External:
		return fmt.Sprintf("example at %s declares an external value, which is never fetched", c.Location)
	default:
		return ""
	}
}

// ExampleSet is every candidate declared at one semantic input or output
// location: one media type of one request body, one media type of one
// response, or one parameter.
type ExampleSet struct {
	// Single is the singular `example` candidate, if declared.
	Single *ExampleCandidate
	// Named holds `examples` candidates keyed by their OpenAPI name. Keys are
	// compared as exact strings; iterate through SortedNames so behavior never
	// depends on map order.
	Named map[string]*ExampleCandidate
	// SchemaFallback holds schema-level and schema-property candidates. They
	// are consulted only when neither Single nor Named yields a usable value,
	// so a whole-body example a human wrote always wins over one the generator
	// assembles.
	SchemaFallback []*ExampleCandidate
	// Location anchors the set itself, for diagnostics about the location
	// rather than about any single candidate in it.
	Location ExampleLocation
}

// SortedNames returns the named-example keys in lexical order. Selection reads
// names through this so a spec's declaration order cannot influence which
// example a run picks — byte-identical regeneration depends on it.
func (s *ExampleSet) SortedNames() []string {
	if s == nil || len(s.Named) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(s.Named))
}

// Empty reports whether the location declared no candidate of any kind.
func (s *ExampleSet) Empty() bool {
	return s == nil || (s.Single == nil && len(s.Named) == 0 && len(s.SchemaFallback) == 0)
}

// Malformed reports whether this location declares both the singular and the
// named form. OpenAPI treats `example` and `examples` as mutually exclusive, so
// a location declaring both has no determinate value and makes its target
// ineligible rather than being silently resolved one way.
func (s *ExampleSet) Malformed() bool {
	return s != nil && s.Single != nil && len(s.Named) > 0
}

// Validate reports why the set cannot be used, or nil when it is well-formed.
// An empty set is not an error here: whether a missing example is fatal
// depends on whether the interaction needs one, which only the scenario knows.
func (s *ExampleSet) Validate() error {
	if s.Malformed() {
		return fmt.Errorf("%s declares both `example` and `examples`, which are mutually exclusive", s.Location)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Parameter serialization
// ----------------------------------------------------------------------------

// ParameterStyle is an OpenAPI parameter serialization style. Style and explode
// together decide the bytes that land in a recorded request target, so a
// cassette cannot be built for a parameter whose style the generator does not
// implement — a near-miss URL fails the matcher exactly like a wrong one.
type ParameterStyle string

const (
	ParameterStyleSimple         ParameterStyle = "simple"
	ParameterStyleLabel          ParameterStyle = "label"
	ParameterStyleMatrix         ParameterStyle = "matrix"
	ParameterStyleForm           ParameterStyle = "form"
	ParameterStyleSpaceDelimited ParameterStyle = "spaceDelimited"
	ParameterStylePipeDelimited  ParameterStyle = "pipeDelimited"
	ParameterStyleDeepObject     ParameterStyle = "deepObject"
)

// ParameterIn distinguishes the parameter locations cassette generation can
// represent. Header and cookie parameters are deliberately absent: recorded
// request headers are filtered to an allowlist, so a header parameter cannot
// participate in matching.
type ParameterIn string

const (
	ParameterInPath  ParameterIn = "path"
	ParameterInQuery ParameterIn = "query"
)

// DefaultParameterStyle returns the OpenAPI default style for a location, used
// when a parameter omits `style`.
func DefaultParameterStyle(in ParameterIn) ParameterStyle {
	if in == ParameterInPath {
		return ParameterStyleSimple
	}
	return ParameterStyleForm
}

// DefaultParameterExplode returns the OpenAPI default for `explode`, which is
// true for form style and false everywhere else. It is a function rather than a
// constant because the default is style-dependent, and defaulting it wrongly
// silently changes a recorded URL.
func DefaultParameterExplode(style ParameterStyle) bool {
	return style == ParameterStyleForm
}

// ParameterExamples is the cassette-relevant contract for one path or query
// parameter: enough to serialize its value into a request target byte for byte.
type ParameterExamples struct {
	// Name is the parameter name as declared.
	Name string
	// In is the parameter location.
	In ParameterIn
	// Required mirrors the OpenAPI `required` flag.
	Required bool
	// Schema is the normalized parameter schema.
	Schema *Schema
	// Serialization detail is deliberately absent. QueryParam already carries
	// Style/Explode/AllowReserved for the same merged parameters, and exposes
	// ResolvedStyle/ResolvedExplode as the single authority for the location-
	// and style-dependent defaults. Duplicating it here gave two contracts to
	// keep in sync for values nothing read.
	// Examples are the candidates declared for this parameter.
	Examples ExampleSet
	// DeclarationOrder is the one-based position among the operation's merged
	// parameters, matching the convention in QueryParam. Zero means a
	// hand-built fixture carried no source order.
	DeclarationOrder int
}

// ----------------------------------------------------------------------------
// Request and response bodies
// ----------------------------------------------------------------------------

// RequestBodyExamples is the cassette-relevant contract for one operation's
// request body.
type RequestBodyExamples struct {
	// Present records that the operation declares a request body at all.
	// Distinguishing this from an empty example set is what lets generation
	// tell "this operation sends nothing" from "this operation's body has no
	// example".
	Present bool
	// Required mirrors the OpenAPI `required` flag on the request body.
	Required bool
	// MediaType is the selected media type, e.g. application/json.
	MediaType string
	// Schema is the normalized request body schema.
	Schema *Schema
	// Examples are the candidates declared for the selected media type.
	Examples ExampleSet
}

// ResponseExamples is the cassette-relevant contract for one declared response
// outcome.
type ResponseExamples struct {
	// Status is the declared status code as written in the description, e.g.
	// "201". Kept as a string because OpenAPI also permits ranges like "2XX".
	Status string
	// BodyPresent records that the response declares content. A false value on
	// a 204 or 205 is a complete contract, not a gap — see Bodyless.
	BodyPresent bool
	// MediaType is the selected media type; empty when the response is
	// bodyless.
	MediaType string
	// Schema is the normalized response body schema; nil when bodyless.
	Schema *Schema
	// Examples are the candidates declared for the selected media type.
	Examples ExampleSet
}

// Bodyless reports whether this response legitimately carries no body. A 204 or
// 205 declaring no content is a complete contract and must never be reported as
// a missing example, which is the distinction FR-017 turns on.
func (r *ResponseExamples) Bodyless() bool {
	if r == nil {
		return false
	}
	if r.BodyPresent {
		return false
	}
	return r.Status == "204" || r.Status == "205"
}

// Complete reports whether this response can be rendered into a cassette
// interaction: either it is legitimately bodyless, or it declares a body and
// has at least one candidate to fill it.
func (r *ResponseExamples) Complete() bool {
	if r == nil {
		return false
	}
	if r.Bodyless() {
		return true
	}
	return r.BodyPresent && !r.Examples.Empty()
}
