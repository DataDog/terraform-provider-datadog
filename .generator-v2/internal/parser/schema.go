package parser

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/index"
	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// jsonMediaType is the only request/response content type the generator
// normalizes; selecting one keeps the projected schema deterministic.
const jsonMediaType = "application/json"

// paginationExtension is the OpenAPI vendor extension describing a list
// endpoint's page/limit query parameters and result-array property.
const paginationExtension = "x-pagination"

// defaultResultsPath is the JSON:API convention for the response property
// holding a list's elements, used when no x-pagination resultsPath is declared.
const defaultResultsPath = "data"

// UnresolvableRefError reports a $ref whose target component is missing from
// #/components/schemas (or whose form is unsupported).
type UnresolvableRefError struct {
	// Ref is the offending $ref string, e.g. "#/components/schemas/DoesNotExist".
	Ref string
}

func (e *UnresolvableRefError) Error() string {
	return fmt.Sprintf("parser: unresolvable $ref %q: target not found in #/components/schemas", e.Ref)
}

// asUnresolvableRefError converts a BuildV3Model error reporting a missing $ref
// target into a typed *UnresolvableRefError, or returns nil for any other error.
// libopenapi drops the body that references a missing local component, so the
// dangling ref never reaches NormalizeSchemas — this catches it at the build step.
func asUnresolvableRefError(err error) *UnresolvableRefError {
	var idxErr *index.IndexingError
	if !errors.As(err, &idxErr) {
		return nil
	}
	// libopenapi phrases it as: component `<ref>` does not exist in the specification
	msg := idxErr.Error()
	if !strings.Contains(msg, "does not exist in the specification") {
		return nil
	}
	ref, ok := backtickedToken(msg)
	if !ok {
		return nil
	}
	return &UnresolvableRefError{Ref: ref}
}

// backtickedToken returns the substring between the first pair of backticks in s.
func backtickedToken(s string) (string, bool) {
	_, afterFirst, found := strings.Cut(s, "`")
	if !found {
		return "", false
	}
	token, _, found := strings.Cut(afterFirst, "`")
	if !found {
		return "", false
	}
	return token, true
}

// NormalizeSchemas fills RequestSchema and ResponseSchema on every tracked
// operation's CRUD group, resolving the create/read/update/delete operationIds
// in Tracking.Group to operations and extracting their bodies from rawOps.
//
// Request uses the application/json requestBody; response uses the
// application/json body of the lowest-numbered 2xx code that has one. A missing
// body leaves the field nil. $refs resolve through spec.Components up to maxDepth
// edges, beyond which a node is SchemaKindRefCycle; a missing target yields
// *UnresolvableRefError. Untracked, ungrouped operations are left untouched.
func NormalizeSchemas(spec *model.Spec, rawOps map[*model.Operation]*v3.Operation, maxDepth int, trackingFieldName string) error {
	if spec == nil {
		return nil
	}
	n := &schemaNormalizer{
		components:        spec.Components,
		maxDepth:          maxDepth,
		trackingFieldName: trackingFieldName,
	}

	// operationId → *model.Operation, to resolve a group's CRUD references.
	byID := make(map[string]*model.Operation, len(spec.Operations))
	for _, op := range spec.Operations {
		if op == nil || op.OperationId == "" {
			continue
		}
		byID[op.OperationId] = op
	}

	// Roots are tracked operations; each fills its group's operations, which may
	// themselves be untracked. filled dedups operations shared across groups.
	filled := make(map[*model.Operation]bool)
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil {
			continue
		}
		for _, id := range groupOperationIds(op.Tracking) {
			target := byID[id]
			if target == nil || filled[target] {
				continue
			}
			filled[target] = true
			if err := n.fillOperation(target, rawOps[target]); err != nil {
				return err
			}
		}
	}

	// Resolve each tracked op's search reference to the list operation it points
	// at, so BuildArtifact can reach the search op's filters and list call. An
	// unknown operationId leaves SearchOp nil for BuildArtifact to fail-slow on.
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil || op.Tracking.Group == nil {
			continue
		}
		if id := op.Tracking.Group.Search; id != "" {
			op.SearchOp = byID[id]
		}
	}
	return nil
}

// groupOperationIds returns the non-empty operationIds backing a tracking group
// (create/read/search/update/delete), so their schemas get normalized.
func groupOperationIds(t *model.TrackingFieldMetadata) []string {
	if t == nil || t.Group == nil {
		return nil
	}
	g := t.Group
	ids := make([]string, 0, 5)
	for _, id := range []string{g.Create, g.Read, g.Search, g.Update, g.Delete} {
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// schemaNormalizer holds the per-pass state: the component set for $ref
// resolution, the depth bound, and the sensitive-marking extension key.
type schemaNormalizer struct {
	components        *v3.Components
	maxDepth          int
	trackingFieldName string
}

// fillOperation normalizes raw's request and 2xx response bodies into op,
// leaving a field nil when its body is absent. It also captures query
// parameters, pagination, and the list element type, which feed the plural
// data-source path and are harmless to the singular and resource paths.
func (n *schemaNormalizer) fillOperation(op *model.Operation, raw *v3.Operation) error {
	if op == nil || raw == nil {
		return nil
	}
	if reqProxy := requestBodySchemaProxy(raw); reqProxy != nil {
		req, err := n.normalizeProxy(reqProxy, 0)
		if err != nil {
			return err
		}
		op.RequestSchema = req
	}

	if err := n.fillQueryParams(op, raw); err != nil {
		return err
	}
	op.Pagination = decodePagination(raw)

	respProxy := responseBodySchemaProxy(raw)
	if respProxy == nil {
		return nil
	}
	// Capture the response type name from the top-level body proxy before
	// normalizeProxy follows the $ref and discards it. An inline/absent body
	// leaves ResponseRefName empty.
	if respProxy.IsReference() {
		op.ResponseRefName = lastRefSegment(respProxy.GetReference())
	}
	resp, err := n.normalizeProxy(respProxy, 0)
	if err != nil {
		return err
	}
	op.ResponseSchema = resp

	return n.retainItemRef(op, respProxy)
}

// fillQueryParams normalizes raw's in:query parameters onto op.QueryParams,
// sorted by name. libopenapi resolves $ref parameters (e.g.
// #/components/parameters/PageNumber) during its build, so each parameter
// arrives with Name and Schema populated; the inner schema is normalized like a
// body so type/format/enum/array come through. Raw bracketed names
// (filter[keyword]) are preserved.
func (n *schemaNormalizer) fillQueryParams(op *model.Operation, raw *v3.Operation) error {
	for _, p := range raw.Parameters {
		if p == nil || p.In != "query" || p.Name == "" {
			continue
		}
		schema, err := n.normalizeProxy(p.Schema, 0)
		if err != nil {
			return err
		}
		op.QueryParams = append(op.QueryParams, model.QueryParam{
			Name:        p.Name,
			Required:    p.Required != nil && *p.Required,
			Schema:      schema,
			Description: p.Description,
		})
	}
	sort.Slice(op.QueryParams, func(i, j int) bool {
		return op.QueryParams[i].Name < op.QueryParams[j].Name
	})
	return nil
}

// decodePagination decodes raw's x-pagination extension, or returns nil when the
// operation declares none (or the extension is malformed).
func decodePagination(raw *v3.Operation) *model.Pagination {
	if raw.Extensions == nil {
		return nil
	}
	node := raw.Extensions.GetOrZero(paginationExtension)
	if node == nil {
		return nil
	}
	var pg struct {
		LimitParam  string `yaml:"limitParam"`
		PageParam   string `yaml:"pageParam"`
		ResultsPath string `yaml:"resultsPath"`
	}
	if err := node.Decode(&pg); err != nil {
		return nil
	}
	return &model.Pagination{LimitParam: pg.LimitParam, PageParam: pg.PageParam, ResultsPath: pg.ResultsPath}
}

// retainItemRef records op.ItemRefName: the last $ref segment of the
// results-array element schema. The results property is op.Pagination.ResultsPath
// when present, else the JSON:API default "data". A property that is not an array
// (e.g. a get-by-id "data" object) leaves ItemRefName empty.
func (n *schemaNormalizer) retainItemRef(op *model.Operation, respProxy *base.SchemaProxy) error {
	resultsPath := defaultResultsPath
	if op.Pagination != nil && op.Pagination.ResultsPath != "" {
		resultsPath = op.Pagination.ResultsPath
	}
	body, err := n.resolveToSchema(respProxy)
	if err != nil || body == nil || body.Properties == nil {
		return err
	}
	prop, err := n.resolveToSchema(body.Properties.GetOrZero(resultsPath))
	if err != nil || prop == nil {
		return err
	}
	if !hasType(prop, "array") || prop.Items == nil || !prop.Items.IsA() {
		return nil
	}
	if elem := prop.Items.A; elem != nil && elem.IsReference() {
		op.ItemRefName = lastRefSegment(elem.GetReference())
	}
	return nil
}

// resolveToSchema follows a schema proxy through one or more $ref hops to its
// underlying *base.Schema, mirroring normalizeProxy's resolution but returning
// the raw libopenapi node so callers can read $ref names the model discards.
func (n *schemaNormalizer) resolveToSchema(proxy *base.SchemaProxy) (*base.Schema, error) {
	if proxy == nil {
		return nil, nil
	}
	if proxy.IsReference() {
		target, err := n.resolveRef(proxy.GetReference())
		if err != nil {
			return nil, err
		}
		return n.resolveToSchema(target)
	}
	return proxy.Schema(), nil
}

// lastRefSegment returns the component name after the final "/" of a $ref,
// e.g. "#/components/schemas/IncidentTypeResponse" → "IncidentTypeResponse".
func lastRefSegment(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

// requestBodySchemaProxy returns op's application/json request body schema, or nil.
func requestBodySchemaProxy(op *v3.Operation) *base.SchemaProxy {
	if op.RequestBody == nil || op.RequestBody.Content == nil {
		return nil
	}
	mt := op.RequestBody.Content.GetOrZero(jsonMediaType)
	if mt == nil {
		return nil
	}
	return mt.Schema
}

// responseBodySchemaProxy returns the application/json schema of the
// lowest-numbered 2xx response code that has one, or nil. Codes without a JSON
// body are skipped; numeric ordering makes the choice deterministic.
func responseBodySchemaProxy(op *v3.Operation) *base.SchemaProxy {
	if op.Responses == nil || op.Responses.Codes == nil {
		return nil
	}
	type codedResponse struct {
		code int
		resp *v3.Response
	}
	var twoXX []codedResponse
	for code, resp := range op.Responses.Codes.FromOldest() {
		num, err := strconv.Atoi(code)
		if err != nil || num < 200 || num > 299 || resp == nil {
			continue
		}
		twoXX = append(twoXX, codedResponse{code: num, resp: resp})
	}
	sort.Slice(twoXX, func(i, j int) bool { return twoXX[i].code < twoXX[j].code })

	for _, cr := range twoXX {
		if cr.resp.Content == nil {
			continue
		}
		mt := cr.resp.Content.GetOrZero(jsonMediaType)
		if mt == nil || mt.Schema == nil {
			continue
		}
		return mt.Schema
	}
	return nil
}

// normalizeProxy normalizes one schema proxy, resolving a $ref through the
// component set and counting it against the depth budget. At the boundary the
// node is classified SchemaKindRefCycle rather than expanded.
func (n *schemaNormalizer) normalizeProxy(proxy *base.SchemaProxy, depth int) (*model.Schema, error) {
	if proxy == nil {
		return nil, nil
	}
	if proxy.IsReference() {
		// depth counts $ref edges already followed; the >= bound matches cycles.go.
		if n.maxDepth > 0 && depth >= n.maxDepth {
			return &model.Schema{Kind: model.SchemaKindRefCycle}, nil
		}
		target, err := n.resolveRef(proxy.GetReference())
		if err != nil {
			return nil, err
		}
		return n.normalizeProxy(target, depth+1)
	}
	return n.normalizeSchema(proxy.Schema(), depth)
}

// resolveRef returns the proxy a "#/components/schemas/<name>" ref points to, or
// *UnresolvableRefError when the form is unsupported or the target is absent.
func (n *schemaNormalizer) resolveRef(ref string) (*base.SchemaProxy, error) {
	name, ok := strings.CutPrefix(ref, componentSchemaPrefix)
	if !ok || name == "" {
		return nil, &UnresolvableRefError{Ref: ref}
	}
	if n.components == nil || n.components.Schemas == nil {
		return nil, &UnresolvableRefError{Ref: ref}
	}
	target := n.components.Schemas.GetOrZero(name)
	if target == nil {
		return nil, &UnresolvableRefError{Ref: ref}
	}
	return target, nil
}

// normalizeSchema converts a resolved *base.Schema into a model.Schema:
// classifying its kind from structure and carrying Type, Format, Enum, Required
// and Sensitive. Children recurse at the same depth — only $refs cost depth.
func (n *schemaNormalizer) normalizeSchema(s *base.Schema, depth int) (*model.Schema, error) {
	if s == nil {
		return nil, nil
	}
	out := &model.Schema{
		Kind:        classifyKind(s),
		Type:        firstType(s),
		Format:      s.Format,
		Enum:        enumValues(s),
		Sensitive:   n.isSensitive(s),
		Description: s.Description,
	}

	// The kind (set above by classifyKind) decides which children to recurse into
	// and where to store them. Primitive and Unsupported have no children, so they
	// have no case and fall through with only the scalar fields already set.
	switch out.Kind {
	case model.SchemaKindObject:
		// Object: walk every named property into out.Properties, keyed by name.
		out.Properties = make(map[string]*model.Schema)
		// Sorted iteration keeps recursion (and any surfaced error) deterministic.
		for _, key := range sortedPropertyKeys(s) {
			child, err := n.normalizeProxy(s.Properties.GetOrZero(key), depth)
			if err != nil {
				return nil, err
			}
			out.Properties[key] = child
		}
		out.Required = sortedRequired(s)

	case model.SchemaKindArray:
		// Array: a single element schema, carried in out.Items.
		if s.Items != nil && s.Items.IsA() {
			item, err := n.normalizeProxy(s.Items.A, depth)
			if err != nil {
				return nil, err
			}
			out.Items = elementOrUnsupported(item)
		}

	case model.SchemaKindMap:
		// Map: dynamic keys sharing one value schema (additionalProperties),
		// carried in out.Items. A boolean `additionalProperties: true` declares
		// no value schema — its values are unconstrained, which TF cannot
		// represent, so carry an Unsupported sentinel for the check to reject.
		if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
			value, err := n.normalizeProxy(s.AdditionalProperties.A, depth)
			if err != nil {
				return nil, err
			}
			out.Items = elementOrUnsupported(value)
		} else {
			out.Items = &model.Schema{Kind: model.SchemaKindUnsupported}
		}

	case model.SchemaKindVariant:
		// Variant: one of several alternative shapes (oneOf/anyOf). Every
		// alternative is normalized and collected into out.Variants.
		for _, group := range [][]*base.SchemaProxy{s.OneOf, s.AnyOf} {
			for _, variant := range group {
				v, err := n.normalizeProxy(variant, depth)
				if err != nil {
					return nil, err
				}
				out.Variants = append(out.Variants, v)
			}
		}
	}

	return out, nil
}

// classifyKind derives the SchemaKind from structure, not type alone. Precedence
// (first match wins, since a node can satisfy several at once):
// oneOf/anyOf → variant; properties → object; type:array+items → array;
// additionalProperties → map; a concrete scalar type → primitive; anything else
// (free-form/empty object, typeless leaf, itemless array) → unsupported.
func classifyKind(s *base.Schema) model.SchemaKind {
	switch {
	case len(s.OneOf) > 0 || len(s.AnyOf) > 0:
		// "one of several shapes" outranks everything else.
		return model.SchemaKindVariant
	case s.Properties != nil && orderedmap.Len(s.Properties) > 0:
		// Declared named fields → object, regardless of the type keyword.
		return model.SchemaKindObject
	case hasType(s, "array") && s.Items != nil:
		// A list, but only if it says what its elements are. A type:array with
		// no items has an unknown element type and falls through to unsupported.
		return model.SchemaKindArray
	case isMap(s):
		// additionalProperties (and no declared properties) → dynamic-key map.
		return model.SchemaKindMap
	case isRepresentablePrimitive(s):
		// A concrete scalar leaf (string, integer, number, boolean).
		return model.SchemaKindPrimitive
	default:
		// No representable type or structure: a free-form/empty object
		// (type:object with no properties), a typeless leaf (empty schema {}),
		// an itemless array, etc. TF cannot emit these — reject, don't guess.
		return model.SchemaKindUnsupported
	}
}

// isMap reports whether additionalProperties defines a value schema (or is true);
// the caller has already ruled out declared properties.
func isMap(s *base.Schema) bool {
	ap := s.AdditionalProperties
	if ap == nil {
		return false
	}
	return ap.IsA() || (ap.IsB() && ap.B)
}

// elementOrUnsupported returns elem unless it is itself a collection (array or
// map). A Terraform list/map element type must be a primitive or object, so a
// collection-of-collection has no representable element; returning the Unsupported
// sentinel lets CheckSchemaRepresentability reject it like any other unrepresentable
// node and keeps the model builder's matching error unreachable in the pipeline.
func elementOrUnsupported(elem *model.Schema) *model.Schema {
	if elem == nil {
		return &model.Schema{Kind: model.SchemaKindUnsupported}
	}
	switch elem.Kind {
	case model.SchemaKindArray, model.SchemaKindMap:
		return &model.Schema{Kind: model.SchemaKindUnsupported}
	default:
		return elem
	}
}

// hasType reports whether t is in the schema's type set (a slice, since 3.1
// allows multiple types).
func hasType(s *base.Schema, t string) bool {
	return slices.Contains(s.Type, t)
}

// isRepresentablePrimitive reports whether the schema's declared type is a
// concrete scalar Terraform can emit. A node with no type — or one that reached
// here with a non-scalar type, e.g. an itemless array — is not a primitive; it
// is unsupported. The source spec is 3.0, so a single firstType suffices.
func isRepresentablePrimitive(s *base.Schema) bool {
	switch firstType(s) {
	case "string", "integer", "number", "boolean":
		return true
	default:
		return false
	}
}

// firstType returns the schema's first declared type, or "" when untyped.
func firstType(s *base.Schema) string {
	if len(s.Type) > 0 {
		return s.Type[0]
	}
	return ""
}

// enumValues returns the schema's enum values as strings in spec order, or nil.
func enumValues(s *base.Schema) []string {
	if len(s.Enum) == 0 {
		return nil
	}
	vals := make([]string, 0, len(s.Enum))
	for _, node := range s.Enum {
		if node == nil {
			continue
		}
		vals = append(vals, node.Value)
	}
	return vals
}

// sortedPropertyKeys returns an object's property names sorted alphabetically.
func sortedPropertyKeys(s *base.Schema) []string {
	if s.Properties == nil {
		return nil
	}
	keys := make([]string, 0, orderedmap.Len(s.Properties))
	for k := range s.Properties.KeysFromOldest() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedRequired returns a sorted copy of the schema's required list.
func sortedRequired(s *base.Schema) []string {
	if len(s.Required) == 0 {
		return nil
	}
	req := append([]string(nil), s.Required...)
	sort.Strings(req)
	return req
}

// isSensitive reports whether the schema node's tracking extension sets
// sensitive: true. A malformed value is treated as not-sensitive.
func (n *schemaNormalizer) isSensitive(s *base.Schema) bool {
	if s.Extensions == nil {
		return false
	}
	node := s.Extensions.GetOrZero(n.trackingFieldName)
	if node == nil {
		return false
	}
	var ext struct {
		Sensitive bool `yaml:"sensitive"`
	}
	if err := node.Decode(&ext); err != nil {
		return false
	}
	return ext.Sensitive
}
