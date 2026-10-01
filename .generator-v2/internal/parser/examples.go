package parser

import (
	"fmt"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Example extraction
//
// Reads the concrete values an OpenAPI description declares for an operation's
// inputs and outputs onto the normalized operation. Extraction is deliberately
// credulous: it records every candidate it finds, including ones it cannot use,
// because an external or malformed declaration is the evidence a later
// diagnostic needs. Deciding which candidate to use is selection's job, and
// deciding whether the set is sufficient is the scenario's.
//
// Candidate values are read through libopenapi's resolved model, so an example
// reached through a local reusable reference arrives already inlined and is
// represented exactly as a directly declared one would be.
// ----------------------------------------------------------------------------

// maxPropertyExampleDepth bounds the schema-property walk. Property examples
// are a fallback for assembling a body the description never spelled out
// whole, and a deeply nested one is not usefully reconstructable anyway.
const maxPropertyExampleDepth = 8

// ExtractExamples populates each tracked operation's request, response and
// parameter example contracts. It mirrors NormalizeSchemas' traversal so an
// operation reached only through another's tracking group is filled too.
//
// Extraction never fails the run: a value it cannot decode is dropped with the
// rest of the operation's candidates left intact, so one malformed example
// costs its own target and nothing else.
func ExtractExamples(spec *model.Spec, raw *RawContext) {
	if spec == nil || raw == nil {
		return
	}
	filled := make(map[*model.Operation]bool)
	fill := func(target *model.Operation) {
		if target == nil || filled[target] {
			return
		}
		filled[target] = true
		extractOperationExamples(target, raw)
	}
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil {
			continue
		}
		fill(op)
		for _, target := range op.ResolvedGroup.Operations() {
			fill(target)
		}
	}
}

// extractOperationExamples fills one operation's three example contracts.
func extractOperationExamples(op *model.Operation, raw *RawContext) {
	rawOp, ok := raw.Raw(op)
	if !ok {
		return
	}
	op.RequestExamples = extractRequestExamples(op, rawOp)
	op.ResponseExamples = extractResponseExamples(op, rawOp)
	op.ParameterExamples = extractParameterExamples(op, raw)
}

// extractRequestExamples reads the request body's contract. Presence is
// recorded separately from the example set so a caller can tell "this
// operation sends nothing" from "its body has no example".
func extractRequestExamples(op *model.Operation, rawOp *v3.Operation) *model.RequestBodyExamples {
	mt, required, present := SelectRequestBody(rawOp)
	if !present {
		if rawOp.RequestBody == nil {
			return nil
		}
		// A body is declared but carries no JSON representation; record that
		// rather than reporting the operation as bodyless.
		return &model.RequestBodyExamples{Required: required}
	}
	location := model.ExampleLocation{
		Path:        op.Path,
		Method:      op.Method,
		OperationId: op.OperationId,
		Component:   model.ExampleComponentRequestBody,
		MediaType:   jsonMediaType,
	}
	return &model.RequestBodyExamples{
		Present:   true,
		Required:  required,
		MediaType: jsonMediaType,
		Schema:    op.RequestSchema,
		Examples:  mediaExampleSet(mt, location),
	}
}

// extractResponseExamples reads every declared outcome, not just the success
// one. The failure outcomes matter: a generated resource test reads once more
// after destroy and expects a 404, so that response needs an interaction too.
//
// Only the success response carries a normalized schema, since that is the one
// schema normalization built the provider model from. A 404's body is replayed
// verbatim from its example and never decoded into provider state.
func extractResponseExamples(op *model.Operation, rawOp *v3.Operation) []model.ResponseExamples {
	declared := DeclaredResponses(rawOp)
	if len(declared) == 0 {
		return nil
	}
	success, hasSuccess := SelectSuccessResponse(rawOp)
	out := make([]model.ResponseExamples, 0, len(declared))
	for _, response := range declared {
		location := model.ExampleLocation{
			Path:        op.Path,
			Method:      op.Method,
			OperationId: op.OperationId,
			Component:   model.ExampleComponentResponse,
			Detail:      response.Status,
		}
		entry := model.ResponseExamples{
			Status:      response.Status,
			BodyPresent: response.BodyPresent(),
			Headers:     declaredHeaderNames(response.Response),
		}
		if response.BodyPresent() {
			entry.MediaType = response.MediaType
			location.MediaType = response.MediaType
			if mt := jsonMediaTypeOf(response.Response.Content); mt != nil {
				entry.Examples = mediaExampleSet(mt, location)
			}
			if hasSuccess && response.Status == success.Status {
				entry.Schema = op.ResponseSchema
			}
		} else {
			entry.Examples = model.ExampleSet{Location: location}
		}
		out = append(out, entry)
	}
	return out
}

// declaredHeaderNames returns the response's declared header names, sorted.
// Only names are kept: recorded headers are filtered to an allowlist, so a
// declared header value cannot affect replay.
func declaredHeaderNames(response *v3.Response) []string {
	if response == nil || response.Headers == nil {
		return nil
	}
	var names []string
	for name := range response.Headers.FromOldest() {
		names = append(names, name)
	}
	return sortedStrings(names)
}

// extractParameterExamples reads the merged path and query parameter
// contracts, carrying the serialization detail needed to rebuild a request
// target byte for byte. Parameters come through RawContext.MergedParameters,
// the same entry point schema normalization uses, so the two cannot disagree
// about which parameters an operation takes.
func extractParameterExamples(op *model.Operation, raw *RawContext) []model.ParameterExamples {
	merged := raw.MergedParameters(op)
	if len(merged) == 0 {
		return nil
	}
	var out []model.ParameterExamples
	for index, p := range merged {
		if p == nil || p.Name == "" {
			continue
		}
		var in model.ParameterIn
		switch p.In {
		case "path":
			in = model.ParameterInPath
		case "query":
			in = model.ParameterInQuery
		default:
			// Header and cookie parameters cannot participate in matching,
			// since recorded headers are filtered to an allowlist.
			continue
		}
		location := model.ExampleLocation{
			Path:        op.Path,
			Method:      op.Method,
			OperationId: op.OperationId,
			Component:   model.ExampleComponentParameter,
			Detail:      p.Name,
		}
		// Style and explode are stored resolved. Their defaults are
		// location- and style-dependent, and explode defaults to true for
		// form style, so leaving them unresolved would let a renderer apply
		// the wrong one and silently change the recorded request target.
		style := model.ParameterStyle(p.Style)
		if style == "" {
			style = model.DefaultParameterStyle(in)
		}
		explode := model.DefaultParameterExplode(style)
		if p.Explode != nil {
			explode = *p.Explode
		}
		entry := model.ParameterExamples{
			Name:             p.Name,
			In:               in,
			Required:         p.Required != nil && *p.Required,
			Schema:           normalizedParamSchema(op, p.Name, in),
			Style:            style,
			Explode:          explode,
			AllowReserved:    p.AllowReserved,
			DeclarationOrder: index + 1,
			Examples:         parameterExampleSet(p, location),
		}
		out = append(out, entry)
	}
	return out
}

// normalizedParamSchema finds the normalized schema normalization already
// produced for this parameter, rather than normalizing it a second time.
func normalizedParamSchema(op *model.Operation, name string, in model.ParameterIn) *model.Schema {
	params := op.QueryParams
	if in == model.ParameterInPath {
		params = op.PathParams
	}
	for i := range params {
		if params[i].Name == name {
			return params[i].Schema
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Candidate collection
// ----------------------------------------------------------------------------

// mediaExampleSet collects the candidates declared at one media type: the
// singular form, the named form, and the schema fallbacks behind them.
func mediaExampleSet(mt *v3.MediaType, location model.ExampleLocation) model.ExampleSet {
	set := model.ExampleSet{Location: location}
	if mt == nil {
		return set
	}
	set.Single = singularCandidate(mt.Example, location, model.ExampleSourceMedia)
	set.Named = namedCandidates(mt.Examples, location, model.ExampleSourceMedia)
	set.SchemaFallback = schemaFallbackCandidates(mt.Schema, location)
	return set
}

// parameterExampleSet collects a parameter's candidates. A parameter's
// examples are not media-scoped, so its location carries no media type.
func parameterExampleSet(p *v3.Parameter, location model.ExampleLocation) model.ExampleSet {
	set := model.ExampleSet{Location: location}
	set.Single = singularCandidate(p.Example, location, model.ExampleSourceParameter)
	set.Named = namedCandidates(p.Examples, location, model.ExampleSourceParameter)
	set.SchemaFallback = schemaFallbackCandidates(p.Schema, location)
	return set
}

// singularCandidate decodes an `example` field into a candidate, or returns
// nil when the field is absent or cannot be decoded.
func singularCandidate(node *yaml.Node, location model.ExampleLocation, kind model.ExampleSourceKind) *model.ExampleCandidate {
	if node == nil {
		return nil
	}
	value, err := decodeExampleValue(node)
	if err != nil {
		return nil
	}
	location.Field = "example"
	return &model.ExampleCandidate{
		Value:      value,
		SourceKind: kind,
		Location:   location,
	}
}

// namedCandidates decodes an `examples` map. An entry declaring an external
// value is retained with a nil value and External set: it is ineligible, but a
// diagnostic needs to be able to say so and name where it came from.
func namedCandidates(
	examples *orderedmap.Map[string, *base.Example],
	location model.ExampleLocation,
	kind model.ExampleSourceKind,
) map[string]*model.ExampleCandidate {
	if examples == nil || examples.Len() == 0 {
		return nil
	}
	out := make(map[string]*model.ExampleCandidate, examples.Len())
	for name, example := range examples.FromOldest() {
		if example == nil {
			continue
		}
		entryLocation := location
		entryLocation.Field = "examples"
		entryLocation.ExampleName = name

		candidate := &model.ExampleCandidate{
			Name:       name,
			SourceKind: kind,
			Location:   entryLocation,
			Reference:  example.Reference,
		}
		if example.ExternalValue != "" {
			candidate.External = true
			out[name] = candidate
			continue
		}
		value, err := decodeExampleValue(example.Value)
		if err != nil {
			continue
		}
		candidate.Value = value
		out[name] = candidate
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// schemaFallbackCandidates collects the schema-side candidates behind a
// location's own examples: the whole-body `example` first, then each scalar
// property's. They are consulted only when the media or parameter examples
// yield nothing, so a whole-body example a human wrote always wins over one
// the generator would assemble.
func schemaFallbackCandidates(proxy *base.SchemaProxy, location model.ExampleLocation) []*model.ExampleCandidate {
	if proxy == nil {
		return nil
	}
	schema, err := proxy.BuildSchema()
	if err != nil || schema == nil {
		return nil
	}
	var out []*model.ExampleCandidate
	if whole := singularCandidate(schema.Example, location, model.ExampleSourceSchema); whole != nil {
		out = append(out, whole)
	}
	out = append(out, propertyExampleCandidates(schema, location, "", 0, map[*base.Schema]bool{})...)
	return out
}

// propertyExampleCandidates walks a schema collecting each scalar property's
// own example, keyed by its dotted path. A property example is only ever
// meaningful assembled with its siblings, so the path is what makes it usable.
func propertyExampleCandidates(
	schema *base.Schema,
	location model.ExampleLocation,
	prefix string,
	depth int,
	onStack map[*base.Schema]bool,
) []*model.ExampleCandidate {
	if schema == nil || depth > maxPropertyExampleDepth || onStack[schema] {
		return nil
	}
	onStack[schema] = true
	defer delete(onStack, schema)

	var out []*model.ExampleCandidate
	for _, branch := range schema.AllOf {
		out = append(out, resolveAndWalk(branch, location, prefix, depth+1, onStack)...)
	}
	if schema.Items != nil && schema.Items.IsA() {
		out = append(out, resolveAndWalk(schema.Items.A, location, prefix+"[]", depth+1, onStack)...)
	}
	if schema.Properties == nil {
		return out
	}
	for _, name := range sortedPropertyNames(schema) {
		property := schema.Properties.GetOrZero(name)
		if property == nil {
			continue
		}
		resolved, err := property.BuildSchema()
		if err != nil || resolved == nil {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if resolved.Properties != nil || len(resolved.AllOf) > 0 || resolved.Items != nil {
			out = append(out, propertyExampleCandidates(resolved, location, path, depth+1, onStack)...)
			continue
		}
		if resolved.Example == nil {
			continue
		}
		value, err := decodeExampleValue(resolved.Example)
		if err != nil {
			continue
		}
		propertyLocation := location
		propertyLocation.Field = "example"
		propertyLocation.PropertyPath = path
		out = append(out, &model.ExampleCandidate{
			Value:      value,
			SourceKind: model.ExampleSourceSchemaProperty,
			Location:   propertyLocation,
			Sensitive:  schemaDeclaresSecret(resolved),
		})
	}
	return out
}

// resolveAndWalk builds a proxy and continues the property walk through it.
func resolveAndWalk(
	proxy *base.SchemaProxy,
	location model.ExampleLocation,
	prefix string,
	depth int,
	onStack map[*base.Schema]bool,
) []*model.ExampleCandidate {
	if proxy == nil {
		return nil
	}
	resolved, err := proxy.BuildSchema()
	if err != nil || resolved == nil {
		return nil
	}
	return propertyExampleCandidates(resolved, location, prefix, depth, onStack)
}

// schemaDeclaresSecret reports whether a schema node marks its value as
// credential material, by either writeOnly or Datadog's x-secret extension. A
// candidate so marked is usable only after replacement.
func schemaDeclaresSecret(schema *base.Schema) bool {
	if schema == nil {
		return false
	}
	if schema.WriteOnly != nil && *schema.WriteOnly {
		return true
	}
	return boolExtension(schema, secretExtension)
}

// decodeExampleValue decodes a YAML example node into the supported semantic
// value set: a scalar, a map, a slice, or nil for a declared null. A node that
// decodes to anything else is rejected rather than carried, since a value the
// renderer cannot represent would surface as a replay mismatch rather than a
// generation error.
func decodeExampleValue(node *yaml.Node) (any, error) {
	if node == nil {
		return nil, fmt.Errorf("no example node")
	}
	if node.ShortTag() == "!!null" {
		return nil, nil
	}
	var value any
	if err := node.Decode(&value); err != nil {
		return nil, fmt.Errorf("decoding example at line %d: %w", node.Line, err)
	}
	if err := checkSupportedValue(value, ""); err != nil {
		return nil, err
	}
	return value, nil
}

// checkSupportedValue rejects a decoded value carrying a type the cassette
// renderer cannot represent, naming the path at which it appears.
func checkSupportedValue(value any, path string) error {
	switch typed := value.(type) {
	case nil, bool, string, int, int64, float64:
		return nil
	case map[string]any:
		for _, key := range sortedStrings(mapKeys(typed)) {
			if err := checkSupportedValue(typed[key], joinPath(path, key)); err != nil {
				return err
			}
		}
		return nil
	case []any:
		for i, item := range typed {
			if err := checkSupportedValue(item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
		return nil
	default:
		where := path
		if where == "" {
			where = "the example root"
		}
		return fmt.Errorf("unsupported example value type %T at %s", value, where)
	}
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func sortedPropertyNames(schema *base.Schema) []string {
	if schema == nil || schema.Properties == nil {
		return nil
	}
	var names []string
	for name := range schema.Properties.FromOldest() {
		names = append(names, name)
	}
	return sortedStrings(names)
}
