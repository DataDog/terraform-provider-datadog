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

// jsonMediaType is the only request/response content type normalized; picking
// exactly one keeps the projected schema deterministic.
const jsonMediaType = "application/json"

// paginationExtension is the OpenAPI vendor extension describing a list
// endpoint's page/limit query parameters and result-array property.
const paginationExtension = "x-pagination"

// secretExtension is Datadog's schema-level credential marker. It defaults an
// attribute to sensitive unless the tracking extension overrides it.
const secretExtension = "x-secret"

// defaultResultsPath is the JSON:API response property holding a list's
// elements, used when x-pagination declares no resultsPath.
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

// asUnresolvableRefError converts a BuildV3Model indexing error about a missing
// $ref target into a typed *UnresolvableRefError, returning nil for any other
// error. libopenapi drops the body referencing a missing local component, so
// the dangling ref has to be caught here rather than during normalization.
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
// operation and on every operation its resolved group names, reading from
// rawOps the application/json requestBody and the application/json body of the
// lowest-numbered 2xx code that has one. A missing body leaves the field nil, a
// missing $ref target yields *UnresolvableRefError, and a local oneOf naming
// failure becomes an Unsupported node. Requires ResolveOperationGroups first.
func NormalizeSchemas(spec *model.Spec, rawOps map[*model.Operation]*v3.Operation, maxDepth int, trackingFieldName string) error {
	if spec == nil {
		return nil
	}
	n := &schemaNormalizer{
		components:        spec.Components,
		maxDepth:          maxDepth,
		trackingFieldName: trackingFieldName,
	}

	// Each tracked operation fills its own bodies and its group's, which may
	// themselves be untracked, so an operation whose group does not name it —
	// or that has no group — still gets its own trees populated. filled dedups
	// operations shared across groups.
	filled := make(map[*model.Operation]bool)
	fill := func(target *model.Operation) error {
		if filled[target] {
			return nil
		}
		filled[target] = true
		return n.fillOperation(target, rawOps[target])
	}
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil {
			continue
		}
		if err := fill(op); err != nil {
			return err
		}
		for _, target := range op.ResolvedGroup.Operations() {
			if err := fill(target); err != nil {
				return err
			}
		}
	}
	return nil
}

// schemaNormalizer holds the per-pass state: the component set for $ref
// resolution, the depth bound, and the tracking extension key.
type schemaNormalizer struct {
	components        *v3.Components
	maxDepth          int
	trackingFieldName string
	// refStack is the $ref chain currently being expanded, in order. A $ref
	// re-entering the chain closes a cycle, which becomes a terminal node
	// instead of recursing forever.
	refStack []string
}

// schemaContext carries what a resolved SchemaProxy no longer knows: the node's
// path, whether its parent declared it required, and the component name it came
// from. oneOf normalization needs all three.
type schemaContext struct {
	path     string
	required bool
	refName  string
}

// fillOperation normalizes raw's request and 2xx response bodies into op,
// leaving a field nil when its body is absent, and also captures path/query
// parameters, pagination, and the list element type.
func (n *schemaNormalizer) fillOperation(op *model.Operation, raw *v3.Operation) error {
	if op == nil || raw == nil {
		return nil
	}
	if reqProxy := requestBodySchemaProxy(raw); reqProxy != nil {
		required := raw.RequestBody.Required != nil && *raw.RequestBody.Required
		// Retain the component name before normalizeProxy follows the $ref and
		// discards it.
		if ref, ok := schemaRef(reqProxy); ok {
			op.RequestRefName = lastRefSegment(ref)
		}
		req, err := n.normalizeProxyAt(reqProxy, 0, schemaContext{
			path:     "request",
			required: required,
		})
		if err != nil {
			return err
		}
		op.RequestSchema = req
	}

	if err := n.fillParameters(op, raw); err != nil {
		return err
	}
	op.Pagination = decodePagination(raw)

	respProxy := responseBodySchemaProxy(raw)
	if respProxy == nil {
		return nil
	}
	// The same capture on the response side. A 3.0 $ref-with-siblings arrives
	// as a synthetic single-structural-branch allOf, so look through that
	// overlay too; an inline, composed or absent body leaves the name empty.
	responseRefName, err := n.referenceNameThroughOverlay(respProxy)
	if err != nil {
		return err
	}
	op.ResponseRefName = responseRefName
	resp, err := n.normalizeProxyAt(respProxy, 0, schemaContext{path: "response", required: true})
	if err != nil {
		return err
	}
	op.ResponseSchema = resp

	if err := n.retainItemRef(op, respProxy); err != nil {
		return err
	}
	return n.retainResponseDataRef(op, respProxy)
}

// fillParameters normalizes raw's in:path and in:query parameters onto op,
// sorted by name. libopenapi has already resolved $ref parameters, so each
// arrives with Name and Schema populated; the inner schema is normalized like a
// body. Raw bracketed names (filter[keyword]) are preserved.
func (n *schemaNormalizer) fillParameters(op *model.Operation, raw *v3.Operation) error {
	for index, p := range raw.Parameters {
		if p == nil || (p.In != "query" && p.In != "path") || p.Name == "" {
			continue
		}
		schema, err := n.normalizeProxyAt(p.Schema, 0, schemaContext{
			path:     "query." + p.Name,
			required: p.Required != nil && *p.Required,
		})
		if err != nil {
			return err
		}
		parameter := model.QueryParam{
			Name:             p.Name,
			Required:         p.Required != nil && *p.Required,
			Schema:           schema,
			Description:      p.Description,
			DeclarationOrder: index + 1,
		}
		if p.In == "path" {
			op.PathParams = append(op.PathParams, parameter)
		} else {
			op.QueryParams = append(op.QueryParams, parameter)
		}
	}
	sort.Slice(op.PathParams, func(i, j int) bool {
		return op.PathParams[i].Name < op.PathParams[j].Name
	})
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
// results-array element schema. The results property is
// op.Pagination.ResultsPath when set, else "data". A non-array property (e.g. a
// by-id "data" object) leaves ItemRefName empty.
func (n *schemaNormalizer) retainItemRef(op *model.Operation, respProxy *base.SchemaProxy) error {
	resultsPath := defaultResultsPath
	if op.Pagination != nil && op.Pagination.ResultsPath != "" {
		resultsPath = op.Pagination.ResultsPath
	}
	propProxy, err := n.findPropertyProxy(respProxy, resultsPath)
	if err != nil || propProxy == nil {
		return err
	}
	prop, err := n.resolveOverlayToSchema(propProxy)
	if err != nil || prop == nil {
		return err
	}
	if !hasType(prop, "array") || prop.Items == nil || !prop.Items.IsA() {
		return nil
	}
	itemRefName, err := n.referenceNameThroughOverlay(prop.Items.A)
	if err != nil {
		return err
	}
	op.ItemRefName = itemRefName
	return nil
}

// retainResponseDataRef records op.ResponseDataRefName: the last $ref segment
// of a by-id response's "data" property when that property is a single object
// reference (e.g. "FullAPIKey"). A "data" resolving to an array leaves it
// empty even when the array is referenced; retainItemRef covers that case.
func (n *schemaNormalizer) retainResponseDataRef(op *model.Operation, respProxy *base.SchemaProxy) error {
	data, err := n.findPropertyProxy(respProxy, defaultResultsPath)
	if err != nil || data == nil {
		return err
	}
	dataSchema, err := n.resolveOverlayToSchema(data)
	if err != nil {
		return err
	}
	if dataSchema != nil && hasType(dataSchema, "array") {
		return nil
	}
	dataRefName, err := n.referenceNameThroughOverlay(data)
	if err != nil {
		return err
	}
	op.ResponseDataRefName = dataRefName
	return nil
}

// referenceNameThroughOverlay returns the last segment of a direct $ref, or of
// the sole structural branch of an allOf metadata overlay. A multi-branch
// composition has no single type identity, so it returns "".
func (n *schemaNormalizer) referenceNameThroughOverlay(proxy *base.SchemaProxy) (string, error) {
	for proxy != nil {
		if proxy.IsReference() {
			return lastRefSegment(proxy.GetReference()), nil
		}
		branch, ok, err := n.singleAllOfStructuralBranch(proxy.Schema())
		if err != nil || !ok {
			return "", err
		}
		proxy = branch
	}
	return "", nil
}

// resolveOverlayToSchema follows refs and unwraps single-structural-branch
// allOf overlays until it reaches the schema that owns the shape.
func (n *schemaNormalizer) resolveOverlayToSchema(proxy *base.SchemaProxy) (*base.Schema, error) {
	return n.resolveOverlayToSchemaAt(proxy, 0, make(map[string]bool))
}

func (n *schemaNormalizer) resolveOverlayToSchemaAt(proxy *base.SchemaProxy, depth int, onStack map[string]bool) (*base.Schema, error) {
	if proxy == nil {
		return nil, nil
	}
	if proxy.IsReference() {
		ref := proxy.GetReference()
		if onStack[ref] || (n.maxDepth > 0 && depth >= n.maxDepth) {
			return nil, nil
		}
		target, err := n.resolveRef(ref)
		if err != nil {
			return nil, err
		}
		onStack[ref] = true
		defer delete(onStack, ref)
		return n.resolveOverlayToSchemaAt(target, depth+1, onStack)
	}

	schema := proxy.Schema()
	branch, ok, err := n.singleAllOfStructuralBranchAt(schema, depth, onStack)
	if err != nil {
		return nil, err
	}
	if !ok {
		return schema, nil
	}
	return n.resolveOverlayToSchemaAt(branch, depth, onStack)
}

// singleAllOfStructuralBranch returns the unique non-annotation branch of an
// allOf. ok=false for anything that is not such an overlay, including a genuine
// multi-branch composition, which has no single identity.
func (n *schemaNormalizer) singleAllOfStructuralBranch(s *base.Schema) (*base.SchemaProxy, bool, error) {
	return n.singleAllOfStructuralBranchAt(s, 0, make(map[string]bool))
}

func (n *schemaNormalizer) singleAllOfStructuralBranchAt(s *base.Schema, depth int, onStack map[string]bool) (*base.SchemaProxy, bool, error) {
	if s == nil || len(s.AllOf) == 0 {
		return nil, false, nil
	}
	var structural *base.SchemaProxy
	for _, branch := range s.AllOf {
		raw, err := n.resolveToSchemaAt(branch, depth, onStack)
		if err != nil {
			return nil, false, err
		}
		if isAnnotationOnlySchema(raw) {
			continue
		}
		if structural != nil {
			return nil, false, nil
		}
		structural = branch
	}
	return structural, structural != nil, nil
}

// findPropertyProxy finds a named property through refs and allOf object
// composition. It stays conservative: if more than one branch exposes the
// property it returns nothing. Ref depth and the active ref path are bounded so
// a malformed recursive composition cannot loop here.
func (n *schemaNormalizer) findPropertyProxy(proxy *base.SchemaProxy, name string) (*base.SchemaProxy, error) {
	return n.findPropertyProxyAt(proxy, name, 0, make(map[string]bool))
}

func (n *schemaNormalizer) findPropertyProxyAt(proxy *base.SchemaProxy, name string, depth int, onStack map[string]bool) (*base.SchemaProxy, error) {
	for proxy != nil {
		if proxy.IsReference() {
			ref := proxy.GetReference()
			if onStack[ref] || (n.maxDepth > 0 && depth >= n.maxDepth) {
				return nil, nil
			}
			target, err := n.resolveRef(ref)
			if err != nil {
				return nil, err
			}
			onStack[ref] = true
			defer delete(onStack, ref)
			return n.findPropertyProxyAt(target, name, depth+1, onStack)
		}
		break
	}
	if proxy == nil {
		return nil, nil
	}

	schema := proxy.Schema()
	if schema == nil {
		return nil, nil
	}
	var found *base.SchemaProxy
	if schema.Properties != nil {
		found = schema.Properties.GetOrZero(name)
	}
	for _, branch := range schema.AllOf {
		candidate, err := n.findPropertyProxyAt(branch, name, depth, onStack)
		if err != nil {
			return nil, err
		}
		if candidate == nil {
			continue
		}
		if found != nil {
			return nil, nil
		}
		found = candidate
	}
	return found, nil
}

// resolveToSchema follows a proxy through its $ref hops to the underlying
// *base.Schema, returning the raw libopenapi node so callers can read $ref
// names the normalized model discards.
func (n *schemaNormalizer) resolveToSchema(proxy *base.SchemaProxy) (*base.Schema, error) {
	return n.resolveToSchemaAt(proxy, 0, make(map[string]bool))
}

func (n *schemaNormalizer) resolveToSchemaAt(proxy *base.SchemaProxy, depth int, onStack map[string]bool) (*base.Schema, error) {
	if proxy == nil {
		return nil, nil
	}
	if proxy.IsReference() {
		ref := proxy.GetReference()
		if onStack[ref] || (n.maxDepth > 0 && depth >= n.maxDepth) {
			return nil, nil
		}
		target, err := n.resolveRef(ref)
		if err != nil {
			return nil, err
		}
		onStack[ref] = true
		defer delete(onStack, ref)
		return n.resolveToSchemaAt(target, depth+1, onStack)
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

// schemaRef returns the $ref a proxy points at, and whether it has one at all.
// IsReference/GetReference cover a bare $ref, but a node writing $ref alongside
// other keywords — which the Datadog spec does — is reported as a non-reference
// with an empty reference, leaving a schema that carries only the siblings and
// so has no type. Per OpenAPI 3.0 those siblings are ignored, so the reference
// recorded in the low-level model is the node's whole meaning.
func schemaRef(proxy *base.SchemaProxy) (string, bool) {
	if proxy == nil {
		return "", false
	}
	if proxy.IsReference() {
		if ref := proxy.GetReference(); ref != "" {
			return ref, true
		}
	}
	if low := proxy.GoLow(); low != nil && low.IsTransformedRefWithSiblings() {
		if ref := low.GetTransformedRefReference(); ref != "" {
			return ref, true
		}
	}
	return "", false
}

// normalizeProxyAt normalizes one schema proxy, resolving a $ref through the
// component set and counting it against the depth budget. The first component
// name of the chain is retained in ctx before resolution, since the resolved
// schema no longer identifies the component it came from.
func (n *schemaNormalizer) normalizeProxyAt(proxy *base.SchemaProxy, depth int, ctx schemaContext) (*model.Schema, error) {
	if proxy == nil {
		return nil, nil
	}
	if proxy.IsReference() {
		ref := proxy.GetReference()
		if ctx.refName == "" {
			ctx.refName = lastRefSegment(ref)
		}
		// depth counts $ref edges already followed. A $ref already being
		// expanded closes a cycle: terminate here, before the depth check, so a
		// genuine cycle is never misreported as merely running out of budget.
		// The two get different kinds, and only the depth one is fixable by
		// raising the flag.
		if i := slices.Index(n.refStack, ref); i >= 0 {
			cycle := append(append([]string{}, n.refStack[i:]...), ref)
			return &model.Schema{
				Kind: model.SchemaKindRefCycle,
				UnsupportedReason: fmt.Sprintf(
					"circular $ref: %s re-enters a schema already being expanded (cycle: %s)",
					ref, strings.Join(cycle, " -> "),
				),
			}, nil
		}
		if n.maxDepth > 0 && depth >= n.maxDepth {
			return &model.Schema{
				Kind: model.SchemaKindDepthExceeded,
				UnsupportedReason: fmt.Sprintf(
					"$ref expansion stopped at --max-depth=%d before reaching %q; this is a depth limit, not a $ref cycle — re-run with a higher --max-depth if the chain is legitimately this deep",
					n.maxDepth, ref,
				),
			}, nil
		}
		target, err := n.resolveRef(ref)
		if err != nil {
			return nil, err
		}
		// Popped by the defer, so the chain unwinds even when a branch errors.
		n.refStack = append(n.refStack, ref)
		defer func() { n.refStack = n.refStack[:len(n.refStack)-1] }()
		return n.normalizeProxyAt(target, depth+1, ctx)
	}
	// A $ref with sibling keywords is a synthetic allOf: retain the referenced
	// component identity, but still normalize the composition so supported
	// sibling metadata is not discarded.
	if ref, ok := schemaRef(proxy); ok && ctx.refName == "" {
		ctx.refName = lastRefSegment(ref)
	}
	return n.normalizeSchema(proxy.Schema(), depth, ctx)
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

// normalizeSchema converts a resolved *base.Schema into a model.Schema,
// classifying its kind from structure and carrying Type, Format, Enum, Required
// and Sensitive. Children recurse at the same depth — only $refs cost depth.
func (n *schemaNormalizer) normalizeSchema(s *base.Schema, depth int, ctx schemaContext) (*model.Schema, error) {
	if s == nil {
		return nil, nil
	}
	// allOf flattens here only without an adjacent oneOf; when both are
	// present oneOf wins, and its sibling merge applies the allOf intersection
	// to every alternative instead.
	if len(s.OneOf) == 0 && len(s.AllOf) > 0 {
		return n.normalizeAllOf(s, depth, ctx)
	}
	out := &model.Schema{
		Kind:            classifyKind(s),
		Type:            firstType(s),
		Format:          s.Format,
		Enum:            enumValues(s),
		HasDefault:      s.Default != nil,
		ReadOnly:        schemaReadOnly(s),
		WriteOnlySecret: schemaWriteOnly(s),
		Sensitive:       n.isSensitive(s),
		Description:     s.Description,
		// The component name that led here — the go-sdk names its generated
		// model after it. ctx carries the first $ref of the chain; a node
		// reached inline leaves this empty.
		RefName: ctx.refName,
	}

	// Kind decides which children to recurse into and where to store them.
	// Primitive and Unsupported have none, so they have no case and keep only
	// the scalar fields set above.
	switch out.Kind {
	case model.SchemaKindObject:
		// Object: walk every named property into out.Properties, keyed by name.
		out.Properties = make(map[string]*model.Schema)
		// Sorted iteration keeps recursion (and any surfaced error) deterministic.
		for _, key := range sortedPropertyKeys(s) {
			child, err := n.normalizeProxyAt(s.Properties.GetOrZero(key), depth, schemaContext{
				path:     model.ChildPath(ctx.path, key),
				required: slices.Contains(s.Required, key),
			})
			if err != nil {
				return nil, err
			}
			out.Properties[key] = child
		}
		out.Required = sortedRequired(s)

	case model.SchemaKindArray:
		// Array: a single element schema, carried in out.Items.
		if s.Items != nil && s.Items.IsA() {
			item, err := n.normalizeProxyAt(s.Items.A, depth, schemaContext{
				path:     model.ChildPath(ctx.path, "[]"),
				required: true,
			})
			if err != nil {
				return nil, err
			}
			out.Items = collectionElement(item)
		}

	case model.SchemaKindMap:
		// Map: dynamic keys sharing one value schema (additionalProperties),
		// carried in out.Items. A boolean `additionalProperties: true` declares
		// no value schema, so its values are unconstrained — which TF cannot
		// represent, hence an Unsupported sentinel.
		if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
			value, err := n.normalizeProxyAt(s.AdditionalProperties.A, depth, schemaContext{
				path:     model.ChildPath(ctx.path, "{}"),
				required: true,
			})
			if err != nil {
				return nil, err
			}
			out.Items = collectionElement(value)
		} else {
			out.Items = &model.Schema{Kind: model.SchemaKindUnsupported}
		}

	case model.SchemaKindOneOf:
		union, err := n.normalizeOneOf(s, depth, ctx)
		if err != nil {
			if isLocalOneOfNamingError(err) {
				out.Kind = model.SchemaKindUnsupported
				out.UnsupportedReason = err.Error()
				return out, nil
			}
			return nil, err
		}
		out.OneOf = union
	}

	return out, nil
}

func isLocalOneOfNamingError(err error) bool {
	var unresolved *model.OneOfVariantNameResolutionError
	if errors.As(err, &unresolved) {
		return true
	}
	var collision *model.OneOfVariantNameCollisionError
	return errors.As(err, &collision)
}

type oneOfAlternativeSource struct {
	schema  *base.Schema
	ref     string
	refName string
}

// normalizeOneOf builds the union model while the raw OpenAPI proxies are
// still available — the only point at which component references, discriminator
// mappings and sibling constraints can be tied to the same alternative without
// guessing. Null-only alternatives are dropped onto the union's Nullable flag.
func (n *schemaNormalizer) normalizeOneOf(s *base.Schema, depth int, ctx schemaContext) (*model.OneOfSpec, error) {
	union := &model.OneOfSpec{
		Name: oneOfEnvelopeName(ctx),
		Path: ctx.path,
		// Kept apart from Name so a component-backed union (wrapper = this
		// name) stays distinguishable from an inline one (wrapper derived from
		// the SDK root) without inspecting Name's spelling.
		RefName:       ctx.refName,
		Optional:      !ctx.required,
		Nullable:      schemaAllowsNull(s),
		Discriminator: normalizeDiscriminator(s.Discriminator),
	}

	for alternativeIndex, proxy := range s.OneOf {
		source, err := n.inspectOneOfVariants(proxy, depth)
		if err != nil {
			return nil, err
		}
		if source.schema != nil && schemaAllowsNull(source.schema) {
			union.Nullable = true
		}
		if source.schema != nil && isNullOnlySchema(source.schema) {
			continue
		}

		tfName, err := oneOfVariantName(
			s.Discriminator,
			source,
			ctx.path,
			alternativeIndex+1,
		)
		if err != nil {
			return nil, fmt.Errorf("parser: %w", err)
		}
		variantPath := model.ChildPath(ctx.path, tfName)
		variantSchema, err := n.normalizeProxyAt(proxy, depth, schemaContext{
			path:     variantPath,
			required: true,
			refName:  source.refName,
		})
		if err != nil {
			return nil, err
		}
		variantSchema, err = n.mergeOneOfSiblings(s, variantSchema, depth, schemaContext{path: variantPath})
		if err != nil {
			return nil, err
		}
		if variantSchema == nil {
			variantSchema = &model.Schema{Kind: model.SchemaKindUnsupported}
		}

		union.Variants = append(union.Variants, model.OneOfVariant{
			TFName:       tfName,
			GoName:       model.SdkName(tfName),
			Schema:       variantSchema,
			RefName:      source.refName,
			ValueWrapped: model.OneOfValueWrapped(variantSchema),
		})
	}

	sort.Slice(union.Variants, func(i, j int) bool {
		return union.Variants[i].TFName < union.Variants[j].TFName
	})
	if err := model.ValidateOneOfVariantNames(ctx.path, union.Variants); err != nil {
		return nil, fmt.Errorf("parser: %w", err)
	}
	return union, nil
}

// inspectOneOfVariants follows reference chains just far enough to retain the
// outer component name and read naming/nullability metadata; depth handling and
// schema conversion stay with the normal normalization pass.
func (n *schemaNormalizer) inspectOneOfVariants(proxy *base.SchemaProxy, depth int) (oneOfAlternativeSource, error) {
	var source oneOfAlternativeSource
	for proxy != nil {
		if proxy.IsReference() {
			if source.ref == "" {
				source.ref = proxy.GetReference()
				source.refName = lastRefSegment(source.ref)
			}
			if n.maxDepth > 0 && depth >= n.maxDepth {
				return source, nil
			}
			target, err := n.resolveRef(proxy.GetReference())
			if err != nil {
				return source, err
			}
			proxy = target
			depth++
			continue
		}

		// A transformed $ref-with-siblings is not a high-level reference:
		// retain its component identity for variant naming, then look through
		// the synthetic allOf so nullability comes from the referenced shape.
		if ref, ok := schemaRef(proxy); ok {
			if source.ref == "" {
				source.ref = ref
				source.refName = lastRefSegment(ref)
			}
			schema, err := n.resolveOverlayToSchemaAt(proxy, depth, make(map[string]bool))
			if err != nil {
				return source, err
			}
			source.schema = schema
			return source, nil
		}

		source.schema = proxy.Schema()
		return source, nil
	}
	return source, nil
}

func oneOfEnvelopeName(ctx schemaContext) string {
	if ctx.refName != "" {
		return ctx.refName
	}
	name := model.SdkName(ctx.path)
	if name == "" {
		name = "Inline"
	}
	return name + "OneOf"
}

func normalizeDiscriminator(discriminator *base.Discriminator) *model.OneOfDiscriminator {
	if discriminator == nil {
		return nil
	}
	out := &model.OneOfDiscriminator{PropertyName: discriminator.PropertyName}
	if discriminator.Mapping != nil && orderedmap.Len(discriminator.Mapping) > 0 {
		out.Mapping = make(map[string]string, orderedmap.Len(discriminator.Mapping))
		for key, value := range discriminator.Mapping.FromOldest() {
			out.Mapping[key] = value
		}
	}
	return out
}

func oneOfVariantName(
	discriminator *base.Discriminator,
	source oneOfAlternativeSource,
	path string,
	alternative int,
) (string, error) {
	candidates := model.OneOfVariantNameCandidates{
		DiscriminatorKey: discriminatorNameForRef(discriminator, source.ref, source.refName),
		RefName:          source.refName,
	}
	if source.schema != nil {
		if isRepresentablePrimitive(source.schema) {
			candidates.PrimitiveType = firstType(source.schema)
			candidates.PrimitiveFormat = source.schema.Format
		}
	}
	return model.ResolveOneOfVariantName(path, alternative, candidates)
}

func discriminatorNameForRef(discriminator *base.Discriminator, ref, refName string) string {
	if discriminator == nil || discriminator.Mapping == nil || (ref == "" && refName == "") {
		return ""
	}
	keys := make([]string, 0, orderedmap.Len(discriminator.Mapping))
	for key := range discriminator.Mapping.KeysFromOldest() {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		mapped := discriminator.Mapping.GetOrZero(key)
		if mapped == ref || (refName != "" && lastRefSegment(mapped) == refName) {
			return key
		}
	}
	return ""
}

// mergeOneOfSiblings applies the constraints adjacent to oneOf to one
// alternative: it normalizes a shallow copy of the parent with oneOf stripped —
// so the ordinary object/array/map/primitive code handles the siblings, at the
// variant's own path — and merges the result into the variant.
func (n *schemaNormalizer) mergeOneOfSiblings(
	parent *base.Schema,
	variant *model.Schema,
	depth int,
	ctx schemaContext,
) (*model.Schema, error) {
	commonRaw, ok := oneOfSiblingSchema(parent)
	if !ok {
		return variant, nil
	}
	common, err := n.normalizeSchema(commonRaw, depth, ctx)
	if err != nil {
		return nil, err
	}
	// A type:object sibling carrying only required names is a constraint
	// carrier, not a standalone object. Normalization classifies a propertyless
	// object as Unsupported, and MergeNormalizedSchemas would then drop the
	// constraint, so give it an empty object shape instead and its required
	// names reach every object alternative.
	if common.Kind == model.SchemaKindUnsupported && isRequiredOnlyObjectConstraint(commonRaw) {
		common.Kind = model.SchemaKindObject
		common.Properties = make(map[string]*model.Schema)
		common.Required = sortedRequired(commonRaw)
	}
	return model.MergeNormalizedSchemas(variant, common), nil
}

func isRequiredOnlyObjectConstraint(s *base.Schema) bool {
	return s != nil &&
		hasType(s, "object") &&
		len(s.Required) > 0 &&
		(s.Properties == nil || orderedmap.Len(s.Properties) == 0) &&
		s.Items == nil &&
		s.AdditionalProperties == nil &&
		len(s.Enum) == 0 &&
		s.Format == ""
}

func oneOfSiblingSchema(s *base.Schema) (*base.Schema, bool) {
	if s == nil {
		return nil, false
	}
	common := *s
	common.OneOf = nil
	common.AnyOf = nil
	common.Discriminator = nil
	common.Nullable = nil
	common.Description = ""
	common.Type = nonNullTypes(s.Type)

	hasConstraints := len(common.Type) > 0 ||
		len(common.AllOf) > 0 ||
		(common.Properties != nil && orderedmap.Len(common.Properties) > 0) ||
		len(common.Required) > 0 ||
		common.Items != nil ||
		common.AdditionalProperties != nil ||
		len(common.Enum) > 0 ||
		common.Format != ""
	return &common, hasConstraints
}

// allOfAnnotations accumulates the metadata an allOf node carries outside its
// structural branches: the outer node's own, unioned with each annotation-only
// branch's before that branch is skipped. A skipped branch asserts nothing
// about the value but still declares metadata that would otherwise be lost.
type allOfAnnotations struct {
	// outerDescription is the allOf node's own and outranks any branch's,
	// hence two fields rather than one.
	outerDescription string
	// branchDescription is the first absorbed branch's, used only as a fallback.
	branchDescription string
	sensitive         bool
	readOnly          bool
	writeOnlySecret   bool
	hasDefault        bool
}

func (n *schemaNormalizer) outerAnnotations(s *base.Schema) allOfAnnotations {
	return allOfAnnotations{
		outerDescription: s.Description,
		sensitive:        n.isSensitive(s),
		readOnly:         schemaReadOnly(s),
		writeOnlySecret:  schemaWriteOnly(s),
		hasDefault:       s.Default != nil,
	}
}

// absorb unions one skipped annotation-only branch into the carrier, reading
// the branch's normalized form: normalizeSchema sets these four fields before
// dispatching on kind, and an annotation-only branch has no kind case that
// would overwrite them.
func (a *allOfAnnotations) absorb(branch *model.Schema) {
	if a.branchDescription == "" {
		a.branchDescription = branch.Description
	}
	a.sensitive = a.sensitive || branch.Sensitive
	a.readOnly = a.readOnly || branch.ReadOnly
	a.writeOnlySecret = a.writeOnlySecret || branch.WriteOnlySecret
	a.hasDefault = a.hasDefault || branch.HasDefault
}

// applyTo unions the collected metadata onto a built result.
func (a allOfAnnotations) applyTo(out *model.Schema) {
	out.Sensitive = out.Sensitive || a.sensitive
	out.ReadOnly = out.ReadOnly || a.readOnly
	out.WriteOnlySecret = out.WriteOnlySecret || a.writeOnlySecret
	out.HasDefault = out.HasDefault || a.hasDefault
	switch {
	case a.outerDescription != "":
		out.Description = a.outerDescription
	case a.branchDescription != "":
		out.Description = a.branchDescription
	}
}

// normalizeAllOf flattens the supported allOf subset: a single structural
// branch is a metadata overlay; several structural branches must all be objects
// with disjoint properties. Annotation-only branches are skipped structurally
// but their metadata is unioned in. Anything outside the subset becomes an
// Unsupported schema carrying a reason rather than an error.
func (n *schemaNormalizer) normalizeAllOf(s *base.Schema, depth int, ctx schemaContext) (*model.Schema, error) {
	annotations := n.outerAnnotations(s)
	unsupported := func(reason string) *model.Schema {
		out := unsupportedSchema(reason)
		annotations.applyTo(out)
		return out
	}
	if reason := unsupportedAllOfOuterStructure(s); reason != "" {
		return unsupported(reason), nil
	}

	type structuralBranch struct {
		index  int
		schema *model.Schema
	}
	branches := make([]structuralBranch, 0, len(s.AllOf))
	for i, proxy := range s.AllOf {
		branch, err := n.normalizeProxyAt(proxy, depth, ctx)
		if err != nil {
			return nil, err
		}
		if branch == nil {
			return unsupported(fmt.Sprintf("allOf branch %d has no schema", i+1)), nil
		}
		annotations.writeOnlySecret = annotations.writeOnlySecret || branch.WriteOnlySecret

		raw, err := n.resolveToSchema(proxy)
		if err != nil {
			return nil, err
		}
		if branch.Kind == model.SchemaKindUnsupported && branch.UnsupportedReason == "" && isAnnotationOnlySchema(raw) {
			annotations.absorb(branch)
			continue
		}

		if branch.Kind == model.SchemaKindUnsupported {
			reason := branch.UnsupportedReason
			if reason == "" {
				reason = fmt.Sprintf("allOf branch %d has unsupported schema kind %q", i+1, branch.Kind)
			}
			return unsupported(reason), nil
		}
		branches = append(branches, structuralBranch{index: i + 1, schema: branch})
	}

	if len(branches) == 0 {
		return unsupported("allOf has no structural branches"), nil
	}

	outerType := firstType(s)
	if len(s.Type) > 1 {
		return unsupported("allOf declares multiple outer types"), nil
	}

	if len(branches) == 1 {
		out := model.CloneSchema(branches[0].schema)
		if outerType != "" && !schemaKindMatchesType(out, outerType) {
			return unsupported(fmt.Sprintf("allOf outer type %q conflicts with branch %d schema kind %q", outerType, branches[0].index, out.Kind)), nil
		}
		if reason := applyAllOfScalarConstraints(out, s, branches[0].index); reason != "" {
			return unsupported(reason), nil
		}
		if len(s.Required) > 0 {
			if out.Kind != model.SchemaKindObject {
				return unsupported(fmt.Sprintf("allOf declares outer required fields for branch %d schema kind %q", branches[0].index, out.Kind)), nil
			}
			out.Required = unionRequired(out.Required, s.Required)
		}
		annotations.applyTo(out)
		return out, nil
	}

	if outerType != "" && outerType != "object" {
		return unsupported(fmt.Sprintf("allOf object composition conflicts with outer type %q", outerType)), nil
	}
	if s.Format != "" || len(s.Enum) > 0 {
		return unsupported("allOf object composition declares scalar outer constraints"), nil
	}

	out := &model.Schema{
		Kind:       model.SchemaKindObject,
		Type:       "object",
		Properties: make(map[string]*model.Schema),
		RefName:    ctx.refName,
	}
	annotations.applyTo(out)

	propertyBranch := make(map[string]int)
	required := make(map[string]bool, len(s.Required))
	for _, name := range s.Required {
		required[name] = true
	}
	for _, branch := range branches {
		if branch.schema.Kind != model.SchemaKindObject {
			return unsupported(fmt.Sprintf(
				"allOf branch %d has schema kind %q; multi-branch composition supports objects only",
				branch.index, branch.schema.Kind,
			)), nil
		}
		out.Sensitive = out.Sensitive || branch.schema.Sensitive
		out.WriteOnlySecret = out.WriteOnlySecret || branch.schema.WriteOnlySecret
		for name, child := range branch.schema.Properties {
			if previous, exists := propertyBranch[name]; exists {
				return unsupported(fmt.Sprintf(
					"allOf property %q is declared by branches %d and %d",
					name, previous, branch.index,
				)), nil
			}
			propertyBranch[name] = branch.index
			out.Properties[name] = model.CloneSchema(child)
		}
		for _, name := range branch.schema.Required {
			required[name] = true
		}
	}

	out.Required = make([]string, 0, len(required))
	for name := range required {
		out.Required = append(out.Required, name)
	}
	sort.Strings(out.Required)
	return out, nil
}

// applyAllOfScalarConstraints folds the outer format and enum into out,
// returning a reason string when they cannot apply. Two enums are intersected;
// a conflicting format or an empty intersection is unsupported rather than a
// silently widened schema.
func applyAllOfScalarConstraints(out *model.Schema, outer *base.Schema, branchIndex int) string {
	if outer.Format == "" && len(outer.Enum) == 0 {
		return ""
	}
	if out.Kind != model.SchemaKindPrimitive {
		return fmt.Sprintf("allOf declares scalar outer constraints for branch %d schema kind %q", branchIndex, out.Kind)
	}

	if outer.Format != "" {
		if out.Format != "" && out.Format != outer.Format {
			return fmt.Sprintf("allOf outer format %q conflicts with branch %d format %q", outer.Format, branchIndex, out.Format)
		}
		out.Format = outer.Format
	}

	outerEnum := enumValues(outer)
	if len(outerEnum) == 0 {
		return ""
	}
	if len(out.Enum) == 0 {
		out.Enum = outerEnum
		return ""
	}
	intersection := intersectEnums(out.Enum, outerEnum)
	if len(intersection) == 0 {
		return fmt.Sprintf("allOf outer enum has no values in common with branch %d enum", branchIndex)
	}
	out.Enum = intersection
	return ""
}

// intersectEnums returns the unique common values in the left-hand declaration's
// order, keeping generated validators stable.
func intersectEnums(left, right []string) []string {
	rightValues := make(map[string]bool, len(right))
	for _, value := range right {
		rightValues[value] = true
	}
	seen := make(map[string]bool, len(left))
	var intersection []string
	for _, value := range left {
		if rightValues[value] && !seen[value] {
			intersection = append(intersection, value)
			seen[value] = true
		}
	}
	return intersection
}

// unionRequired returns the sorted union of two required-property lists.
func unionRequired(left, right []string) []string {
	required := make(map[string]bool, len(left)+len(right))
	for _, name := range left {
		required[name] = true
	}
	for _, name := range right {
		required[name] = true
	}
	out := make([]string, 0, len(required))
	for name := range required {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// unsupportedAllOfOuterStructure rejects nodes whose structure lives both
// outside and inside the allOf. Only an outer type assertion plus
// annotation/constraint metadata is supported; merging outer properties,
// variants, arrays or maps would need full JSON Schema intersection semantics.
func unsupportedAllOfOuterStructure(s *base.Schema) string {
	switch {
	case len(s.OneOf) > 0:
		return "allOf combined with outer oneOf is not supported"
	case len(s.AnyOf) > 0:
		return "allOf combined with outer anyOf is not supported"
	case s.Properties != nil && orderedmap.Len(s.Properties) > 0:
		return "allOf combined with outer properties is not supported"
	case s.Items != nil:
		return "allOf combined with outer items is not supported"
	case s.AdditionalProperties != nil:
		return "allOf combined with outer additionalProperties is not supported"
	default:
		return ""
	}
}

// isAnnotationOnlySchema recognizes a branch that asserts nothing about the
// value — libopenapi's sibling-only branch for a 3.0 $ref with siblings, or the
// same form authored by hand. It is the residual of
// hasStructuralOrConstraintKeywords, not a second allow-list, since the
// qualifying keywords (nullable, readOnly, deprecated, xml, …) are open-ended.
// declaresKeyword excludes the empty schema: an untyped value, not annotation.
func isAnnotationOnlySchema(s *base.Schema) bool {
	if s == nil || hasStructuralOrConstraintKeywords(s) {
		return false
	}
	return declaresKeyword(s)
}

// declaresKeyword reports whether a schema writes any keyword at all, read off
// the node the document was parsed from so no keyword needs naming here. An
// absent node counts as declaring nothing, which keeps the conservative answer
// (structural, therefore Unsupported).
func declaresKeyword(s *base.Schema) bool {
	low := s.GoLow()
	if low == nil || low.RootNode == nil {
		return false
	}
	return len(low.RootNode.Content) > 0
}

// hasStructuralOrConstraintKeywords reports whether a schema narrows values, as
// opposed to being a metadata-only overlay. Deliberately conservative: missing
// an assertion here would discard it and widen the generated schema.
func hasStructuralOrConstraintKeywords(s *base.Schema) bool {
	// Type and composition keywords.
	if len(s.Type) > 0 || len(s.AllOf) > 0 || len(s.OneOf) > 0 || len(s.AnyOf) > 0 {
		return true
	}
	if s.Not != nil || s.Discriminator != nil {
		return true
	}

	// Object and collection structure.
	if s.Properties != nil && orderedmap.Len(s.Properties) > 0 {
		return true
	}
	if s.Items != nil || s.AdditionalProperties != nil || len(s.PrefixItems) > 0 {
		return true
	}
	if s.Contains != nil || s.MinContains != nil || s.MaxContains != nil {
		return true
	}

	// Conditional and dependent schemas.
	if s.If != nil || s.Then != nil || s.Else != nil {
		return true
	}
	if s.DependentSchemas != nil && orderedmap.Len(s.DependentSchemas) > 0 {
		return true
	}
	if s.DependentRequired != nil && orderedmap.Len(s.DependentRequired) > 0 {
		return true
	}
	if s.PatternProperties != nil && orderedmap.Len(s.PatternProperties) > 0 {
		return true
	}
	if s.PropertyNames != nil || s.UnevaluatedItems != nil || s.UnevaluatedProperties != nil {
		return true
	}

	// Scalar and collection constraints.
	if s.MultipleOf != nil || s.Maximum != nil || s.Minimum != nil {
		return true
	}
	if s.ExclusiveMaximum != nil || s.ExclusiveMinimum != nil {
		return true
	}
	if s.MaxLength != nil || s.MinLength != nil || s.Pattern != "" || s.Format != "" {
		return true
	}
	if s.MaxItems != nil || s.MinItems != nil || s.UniqueItems != nil {
		return true
	}
	if s.MaxProperties != nil || s.MinProperties != nil || len(s.Required) > 0 {
		return true
	}

	// Exact-value and reference constraints.
	if len(s.Enum) > 0 || s.Const != nil {
		return true
	}
	// String-content keywords. JSON Schema 2019-09 calls these annotations,
	// but they say how to decode the value, so treat them as constraints.
	if s.ContentEncoding != "" || s.ContentMediaType != "" {
		return true
	}
	return s.DynamicRef != "" || s.ContentSchema != nil
}

// schemaReadOnly reads readOnly through its nil pointer, which libopenapi uses
// to separate "declared false" from "not declared".
func schemaReadOnly(s *base.Schema) bool {
	return s.ReadOnly != nil && *s.ReadOnly
}

func unsupportedSchema(reason string) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindUnsupported, UnsupportedReason: reason}
}

func schemaKindMatchesType(s *model.Schema, typ string) bool {
	if s == nil {
		return false
	}
	switch typ {
	case "object":
		return s.Kind == model.SchemaKindObject || s.Kind == model.SchemaKindMap
	case "array":
		return s.Kind == model.SchemaKindArray
	case "string", "integer", "number", "boolean":
		return s.Kind == model.SchemaKindPrimitive && s.Type == typ
	default:
		return false
	}
}

func schemaAllowsNull(s *base.Schema) bool {
	return s != nil && ((s.Nullable != nil && *s.Nullable) || hasType(s, "null"))
}

func isNullOnlySchema(s *base.Schema) bool {
	if !schemaAllowsNull(s) {
		return false
	}
	return len(nonNullTypes(s.Type)) == 0 &&
		len(s.AllOf) == 0 &&
		len(s.OneOf) == 0 &&
		len(s.AnyOf) == 0 &&
		(s.Properties == nil || orderedmap.Len(s.Properties) == 0) &&
		s.Items == nil &&
		s.AdditionalProperties == nil
}

func nonNullTypes(types []string) []string {
	out := make([]string, 0, len(types))
	for _, typ := range types {
		if typ != "null" {
			out = append(out, typ)
		}
	}
	return out
}

// classifyKind derives the SchemaKind from structure, not type alone. First
// match wins, since a node can satisfy several at once: anyOf → unsupported;
// oneOf → one_of; properties → object; type:array with items → array;
// additionalProperties → map; a concrete scalar type → primitive; anything else
// (free-form object, typeless leaf, itemless array) → unsupported. anyOf stays
// unsupported rather than inheriting oneOf's exactly-one semantics.
func classifyKind(s *base.Schema) model.SchemaKind {
	switch {
	case len(s.AnyOf) > 0:
		// anyOf has no Terraform equivalent; reject rather than drop or guess.
		return model.SchemaKindUnsupported
	case len(s.OneOf) > 0:
		return model.SchemaKindOneOf
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
		// No representable type or structure: a free-form object (type:object
		// with no properties), a typeless leaf (empty schema {}), an itemless
		// array. TF cannot emit these — reject rather than guess.
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

// collectionElement passes a collection's element schema through unchanged —
// nested chains like list(list(string)) are representable — substituting an
// Unsupported sentinel only when the element schema is missing.
func collectionElement(elem *model.Schema) *model.Schema {
	if elem == nil {
		return &model.Schema{Kind: model.SchemaKindUnsupported}
	}
	return elem
}

// hasType reports whether t is in the schema's type set (a slice, since 3.1
// allows multiple types).
func hasType(s *base.Schema, t string) bool {
	return slices.Contains(s.Type, t)
}

// isRepresentablePrimitive reports whether the declared type is a concrete
// scalar Terraform can emit. A node with no type, or one reaching here with a
// non-scalar type (e.g. an itemless array), is not a primitive.
func isRepresentablePrimitive(s *base.Schema) bool {
	switch firstType(s) {
	case "string", "integer", "number", "boolean":
		return true
	default:
		return false
	}
}

// firstType returns the schema's first non-null declared type, or "" when the
// schema is untyped or null-only; nullability is carried on OneOfSpec instead.
func firstType(s *base.Schema) string {
	for _, typ := range s.Type {
		if typ != "null" {
			return typ
		}
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

// isSensitive derives Terraform sensitivity from the explicit tracking
// annotation when present; otherwise writeOnly or x-secret default the field to
// sensitive, so an omitted annotation cannot expose a credential in a plan.
func (n *schemaNormalizer) isSensitive(s *base.Schema) bool {
	if explicit, ok := n.explicitSensitivity(s); ok {
		return explicit
	}
	return schemaWriteOnly(s) || boolExtension(s, secretExtension)
}

// explicitSensitivity distinguishes an omitted annotation from an explicit
// false, so the derived default can be overridden in either direction. A
// malformed annotation counts as absent, leaving x-secret/writeOnly in force.
func (n *schemaNormalizer) explicitSensitivity(s *base.Schema) (bool, bool) {
	if s.Extensions == nil {
		return false, false
	}
	node := s.Extensions.GetOrZero(n.trackingFieldName)
	if node == nil {
		return false, false
	}
	var ext struct {
		Sensitive *bool `yaml:"sensitive"`
	}
	if err := node.Decode(&ext); err != nil || ext.Sensitive == nil {
		return false, false
	}
	return *ext.Sensitive, true
}

func schemaWriteOnly(s *base.Schema) bool {
	return s.WriteOnly != nil && *s.WriteOnly
}

// boolExtension reads a boolean Schema Object extension. Missing, false and
// malformed values all return false; only an explicit true opts in.
func boolExtension(s *base.Schema, name string) bool {
	if s.Extensions == nil {
		return false
	}
	node := s.Extensions.GetOrZero(name)
	if node == nil {
		return false
	}
	var value bool
	return node.Decode(&value) == nil && value
}
