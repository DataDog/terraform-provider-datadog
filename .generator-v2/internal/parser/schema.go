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
// *UnresolvableRefError. Local oneOf naming failures become
// SchemaKindUnsupported nodes carrying their reason so unrelated operations can
// still normalize. Untracked, ungrouped operations are left untouched.
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

// schemaContext carries information that is not available after a SchemaProxy
// has been resolved. oneOf normalization needs all three values to retain
// component identity and to give inline/nested unions deterministic paths.
type schemaContext struct {
	path     string
	required bool
	refName  string
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
		required := raw.RequestBody.Required != nil && *raw.RequestBody.Required
		// Capture the request type name before normalizeProxy follows the $ref and
		// discards it — the mirror of ResponseRefName below. It is the SDK root the
		// oneOf binding pass walks from on the request side.
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
	// Capture the response type name from the top-level body proxy before
	// normalizeProxy follows the $ref and discards it. OpenAPI 3.0 $ref siblings
	// arrive as a synthetic single-structural-branch allOf, so look through that
	// overlay as well. An inline/composed/absent body leaves ResponseRefName empty.
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

// fillParameters normalizes raw's in:path and in:query parameters onto the
// corresponding operation fields, sorted by name. libopenapi resolves $ref parameters (e.g.
// #/components/parameters/PageNumber) during its build, so each parameter
// arrives with Name and Schema populated; the inner schema is normalized like a
// body so type/format/enum/array come through. Raw bracketed names
// (filter[keyword]) are preserved.
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
// results-array element schema. The results property is op.Pagination.ResultsPath
// when present, else the JSON:API default "data". A property that is not an array
// (e.g. a get-by-id "data" object) leaves ItemRefName empty.
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

// retainResponseDataRef records op.ResponseDataRefName: the last $ref segment of a
// by-id response's "data" property when that property is a single object
// reference (e.g. "FullAPIKey"). A list response whose "data" resolves to an
// array leaves the field empty even when that array is referenced (retainItemRef
// covers it). This lets the model detect a "both" data source whose by-id record
// shape diverges from its list element shape.
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

// referenceNameThroughOverlay returns the last segment of a direct $ref or of
// the sole structural branch in an allOf metadata overlay. Multi-branch object
// compositions deliberately have no single SDK type identity.
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

// resolveOverlayToSchema resolves refs and unwraps single-structural-branch
// allOf metadata overlays until it reaches the schema that owns the shape.
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

// singleAllOfStructuralBranch finds a unique non-annotation branch. It returns
// ok=false for schemas that are not overlays (including genuine multi-branch
// compositions), so callers never invent a single SDK identity for a composite.
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
		if n.isAnnotationOnlySchema(raw) {
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
// composition. Strict duplicate-property rejection happens during normalization;
// this helper stays conservative and returns no identity if raw branches expose
// the property more than once. Ref depth and the active path are bounded like
// schema normalization so malformed recursive compositions cannot loop here.
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

// resolveToSchema follows a schema proxy through one or more $ref hops to its
// underlying *base.Schema, mirroring normalizeProxy's resolution but returning
// the raw libopenapi node so callers can read $ref names the model discards.
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
//
// libopenapi's high-level IsReference/GetReference cover a bare $ref, but a node
// that writes $ref alongside other keywords — {$ref: TokenName, example: "x"},
// which the Datadog spec does — is reported as a non-reference with an empty
// reference, leaving a schema carrying only the siblings and therefore no type.
// OpenAPI 3.0 says keywords beside a $ref are ignored, so the reference is the
// whole meaning of such a node; the low-level model still records it.
//
// Every "is this a reference?" test in this file goes through here, so a sibling
// can never cost a node its type, its component name, or its SDK type name.
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
// name is retained in ctx before resolution, since libopenapi's resolved schema
// no longer identifies the component that supplied a oneOf envelope.
func (n *schemaNormalizer) normalizeProxyAt(proxy *base.SchemaProxy, depth int, ctx schemaContext) (*model.Schema, error) {
	if proxy == nil {
		return nil, nil
	}
	if proxy.IsReference() {
		ref := proxy.GetReference()
		if ctx.refName == "" {
			ctx.refName = lastRefSegment(ref)
		}
		// depth counts $ref edges already followed; the >= bound matches cycles.go.
		// Exhausting the budget is not a cycle — cycles.go finds those independently
		// of depth — so it gets its own kind and says how to lift the limit.
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
		return n.normalizeProxyAt(target, depth+1, ctx)
	}
	// libopenapi represents a $ref with sibling keywords as a synthetic allOf.
	// Retain the referenced component identity, but normalize that synthetic
	// composition so supported sibling metadata is not discarded.
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

// normalizeSchema converts a resolved *base.Schema into a model.Schema:
// classifying its kind from structure and carrying Type, Format, Enum, Required
// and Sensitive. Children recurse at the same depth — only $refs cost depth.
func (n *schemaNormalizer) normalizeSchema(s *base.Schema, depth int, ctx schemaContext) (*model.Schema, error) {
	if s == nil {
		return nil, nil
	}
	// oneOf is the Terraform envelope when composition keywords are adjacent.
	// Its sibling merge normalizes the bounded allOf subset once and applies the
	// resulting intersection to every alternative.
	if len(s.OneOf) == 0 && len(s.AllOf) > 0 {
		return n.normalizeAllOf(s, depth, ctx)
	}
	out := &model.Schema{
		Kind:        classifyKind(s),
		Type:        firstType(s),
		Format:      s.Format,
		Enum:        enumValues(s),
		Sensitive:   n.isSensitive(s),
		Description: s.Description,
		// The component name that led here, retained because the Datadog go-sdk
		// names its generated model after it. ctx carries the first $ref of the
		// chain; a node reached inline leaves this empty.
		RefName: ctx.refName,
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
			child, err := n.normalizeProxyAt(s.Properties.GetOrZero(key), depth, schemaContext{
				path:     childPath(ctx.path, key),
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
				path:     childPath(ctx.path, "[]"),
				required: true,
			})
			if err != nil {
				return nil, err
			}
			out.Items = collectionElement(item)
			if out.Items.Kind == model.SchemaKindJSON {
				// Preserve heterogeneous or otherwise unconstrained elements by
				// encoding the complete collection as normalized JSON.
				out.Kind, out.Items = model.SchemaKindJSON, nil
			}
		}

	case model.SchemaKindMap:
		// Map: dynamic keys sharing one value schema (additionalProperties),
		// carried in out.Items. A boolean `additionalProperties: true` declares
		// no value schema, so preserve the complete map as normalized JSON.
		if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
			value, err := n.normalizeProxyAt(s.AdditionalProperties.A, depth, schemaContext{
				path:     childPath(ctx.path, "{}"),
				required: true,
			})
			if err != nil {
				return nil, err
			}
			out.Items = collectionElement(value)
			if out.Items.Kind == model.SchemaKindJSON {
				out.Kind, out.Items = model.SchemaKindJSON, nil
			}
		} else {
			out.Kind = model.SchemaKindJSON
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

// normalizeOneOf builds the parser-facing union model while the raw OpenAPI
// proxies are still available. This is the only point at which component
// references, discriminator mappings and sibling constraints can all be
// associated with the same alternative without guessing downstream.
func (n *schemaNormalizer) normalizeOneOf(s *base.Schema, depth int, ctx schemaContext) (*model.OneOfSpec, error) {
	union := &model.OneOfSpec{
		Name: oneOfEnvelopeName(ctx),
		Path: ctx.path,
		// Kept apart from Name so the SDK binding pass can tell a component-backed
		// union (whose wrapper is this name) from an inline one (whose wrapper it
		// must derive from the SDK root), without inspecting Name's spelling.
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
		variantPath := childPath(ctx.path, tfName)
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
			ValueWrapped: oneOfValueWrapped(variantSchema),
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

// inspectOneOfVariants follows reference chains only far enough to retain
// the outer component name and inspect naming/nullability metadata. The normal
// normalization pass still owns depth handling and schema conversion.
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

		// A transformed $ref-with-siblings is not a high-level reference. Retain
		// its component identity for variant naming, then inspect through the
		// synthetic allOf so nullability comes from the referenced shape while
		// normal normalization still preserves supported sibling metadata.
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

// mergeOneOfSiblings applies the constraints adjacent to oneOf to an
// alternative. Normalizing a shallow copy without oneOf lets the existing
// object/array/map/primitive code process those siblings using the variant's
// canonical path.
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
	// carrier, not a standalone Terraform object. General schema normalization
	// correctly classifies an object with no properties as Unsupported, but
	// retaining that sentinel here would cause mergeNormalizedSchemas to discard
	// the required constraint. Give this merge-only schema an empty object shape
	// so its required names are applied to every object alternative.
	if (common.Kind == model.SchemaKindUnsupported || common.Kind == model.SchemaKindJSON) && isRequiredOnlyObjectConstraint(commonRaw) {
		common.Kind = model.SchemaKindObject
		common.Properties = make(map[string]*model.Schema)
		common.Required = sortedRequired(commonRaw)
	}
	return mergeNormalizedSchemas(variant, common), nil
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

// mergeNormalizedSchemas intersects the subset of OpenAPI constraints retained
// by model.Schema. A kind mismatch remains present as an Unsupported variant so
// later validation can report that precise alternative instead of dropping it.
func mergeNormalizedSchemas(variant, common *model.Schema) *model.Schema {
	if variant == nil {
		return common
	}
	if common == nil {
		return variant
	}
	// A bare type: object sibling constrains every oneOf alternative to an
	// object without adding any fields. A normalized free-form object uses the
	// JSON kind, but an explicitly shaped object alternative is already a
	// narrower representation of that constraint.
	if common.Kind == model.SchemaKindJSON && common.Type == "object" && variant.Kind == model.SchemaKindObject {
		variant.Sensitive = variant.Sensitive || common.Sensitive
		return variant
	}
	if common.Kind == model.SchemaKindUnsupported {
		out := cloneSchema(common)
		if out.Description == "" {
			out.Description = variant.Description
		}
		return out
	}
	if variant.Kind == model.SchemaKindOneOf && variant.OneOf != nil {
		for i := range variant.OneOf.Variants {
			variant.OneOf.Variants[i].Schema = mergeNormalizedSchemas(variant.OneOf.Variants[i].Schema, common)
			variant.OneOf.Variants[i].ValueWrapped = oneOfValueWrapped(variant.OneOf.Variants[i].Schema)
		}
		return variant
	}
	if variant.Kind == model.SchemaKindUnsupported {
		if variant.UnsupportedReason != "" {
			return variant
		}
		if common.Description == "" {
			common.Description = variant.Description
		}
		return common
	}
	if variant.Kind != common.Kind {
		return &model.Schema{
			Kind:              model.SchemaKindUnsupported,
			Description:       variant.Description,
			UnsupportedReason: fmt.Sprintf("oneOf alternative kind %q conflicts with adjacent schema kind %q", variant.Kind, common.Kind),
		}
	}

	switch variant.Kind {
	case model.SchemaKindObject:
		if variant.Properties == nil {
			variant.Properties = make(map[string]*model.Schema)
		}
		for key, commonProperty := range common.Properties {
			if property, exists := variant.Properties[key]; exists {
				variant.Properties[key] = mergeNormalizedSchemas(property, commonProperty)
			} else {
				variant.Properties[key] = commonProperty
			}
		}
		variant.Required = sortedUniqueStrings(append(variant.Required, common.Required...))
	case model.SchemaKindArray, model.SchemaKindMap:
		variant.Items = mergeNormalizedSchemas(variant.Items, common.Items)
	case model.SchemaKindPrimitive:
		if variant.Type != "" && common.Type != "" && variant.Type != common.Type {
			return &model.Schema{
				Kind:              model.SchemaKindUnsupported,
				Description:       variant.Description,
				UnsupportedReason: fmt.Sprintf("oneOf alternative type %q conflicts with adjacent type %q", variant.Type, common.Type),
			}
		}
		if variant.Type == "" {
			variant.Type = common.Type
		}
		if variant.Format != "" && common.Format != "" && variant.Format != common.Format {
			return &model.Schema{
				Kind:              model.SchemaKindUnsupported,
				Description:       variant.Description,
				UnsupportedReason: fmt.Sprintf("oneOf alternative format %q conflicts with adjacent format %q", variant.Format, common.Format),
			}
		}
		if variant.Format == "" {
			variant.Format = common.Format
		}
		switch {
		case len(variant.Enum) == 0:
			variant.Enum = append([]string(nil), common.Enum...)
		case len(common.Enum) > 0:
			intersection := intersectStrings(variant.Enum, common.Enum)
			if len(intersection) == 0 {
				return &model.Schema{
					Kind:              model.SchemaKindUnsupported,
					Description:       variant.Description,
					UnsupportedReason: "oneOf alternative enum has no values in common with adjacent enum",
				}
			}
			variant.Enum = intersection
		}
	}
	variant.Sensitive = variant.Sensitive || common.Sensitive
	return variant
}

func oneOfValueWrapped(schema *model.Schema) bool {
	if schema == nil {
		return false
	}
	switch schema.Kind {
	case model.SchemaKindPrimitive, model.SchemaKindArray, model.SchemaKindMap:
		return true
	default:
		return false
	}
}

// normalizeAllOf flattens the bounded allOf subset used by the Datadog API
// spec. A single structural branch is a metadata overlay; multiple structural
// branches must all be objects with disjoint properties. Annotation-only
// branches and empty identity branches are ignored structurally; annotations may
// still supply a local description or sensitive marker. Anything outside that
// subset becomes an Unsupported schema with a reason so only the affected
// artifact fails later in the model layer.
func (n *schemaNormalizer) normalizeAllOf(s *base.Schema, depth int, ctx schemaContext) (*model.Schema, error) {
	if reason := unsupportedAllOfOuterStructure(s); reason != "" {
		return unsupportedSchema(reason), nil
	}

	type structuralBranch struct {
		index  int
		schema *model.Schema
	}
	branches := make([]structuralBranch, 0, len(s.AllOf))
	annotationDescription := ""
	sensitive := n.isSensitive(s)

	for i, proxy := range s.AllOf {
		branch, err := n.normalizeProxyAt(proxy, depth, ctx)
		if err != nil {
			return nil, err
		}
		if branch == nil {
			return unsupportedSchema(fmt.Sprintf("allOf branch %d has no schema", i+1)), nil
		}

		raw, err := n.resolveToSchema(proxy)
		if err != nil {
			return nil, err
		}
		// An unconstrained schema is the identity element of allOf intersection.
		// This is context-specific: the same {} used as a property value remains
		// arbitrary JSON, but inside allOf it adds no structural constraint.
		if branch.UnsupportedReason == "" && !hasStructuralOrConstraintKeywords(raw) {
			if annotationDescription == "" && raw.Description != "" {
				annotationDescription = raw.Description
			}
			sensitive = sensitive || n.isSensitive(raw)
			continue
		}

		if branch.Kind == model.SchemaKindUnsupported {
			reason := branch.UnsupportedReason
			if reason == "" {
				reason = fmt.Sprintf("allOf branch %d has unsupported schema kind %q", i+1, branch.Kind)
			}
			return unsupportedSchema(reason), nil
		}
		branches = append(branches, structuralBranch{index: i + 1, schema: branch})
	}

	if len(branches) == 0 {
		return unsupportedSchema("allOf has no structural branches"), nil
	}

	outerType := firstType(s)
	if len(s.Type) > 1 {
		return unsupportedSchema("allOf declares multiple outer types"), nil
	}

	if len(branches) == 1 {
		out := cloneSchema(branches[0].schema)
		if outerType != "" && !schemaKindMatchesType(out, outerType) {
			return unsupportedSchema(fmt.Sprintf("allOf outer type %q conflicts with branch %d schema kind %q", outerType, branches[0].index, out.Kind)), nil
		}
		if reason := applyAllOfScalarConstraints(out, s, branches[0].index); reason != "" {
			return unsupportedSchema(reason), nil
		}
		if len(s.Required) > 0 {
			if out.Kind != model.SchemaKindObject {
				return unsupportedSchema(fmt.Sprintf("allOf declares outer required fields for branch %d schema kind %q", branches[0].index, out.Kind)), nil
			}
			out.Required = unionRequired(out.Required, s.Required)
		}
		out.Sensitive = out.Sensitive || sensitive
		switch {
		case s.Description != "":
			out.Description = s.Description
		case annotationDescription != "":
			out.Description = annotationDescription
		}
		return out, nil
	}

	if outerType != "" && outerType != "object" {
		return unsupportedSchema(fmt.Sprintf("allOf object composition conflicts with outer type %q", outerType)), nil
	}
	if s.Format != "" || len(s.Enum) > 0 {
		return unsupportedSchema("allOf object composition declares scalar outer constraints"), nil
	}

	out := &model.Schema{
		Kind:        model.SchemaKindObject,
		Type:        "object",
		Properties:  make(map[string]*model.Schema),
		Sensitive:   sensitive,
		Description: s.Description,
		RefName:     ctx.refName,
	}
	if out.Description == "" {
		out.Description = annotationDescription
	}

	propertyBranch := make(map[string]int)
	required := make(map[string]bool, len(s.Required))
	for _, name := range s.Required {
		required[name] = true
	}
	for _, branch := range branches {
		if branch.schema.Kind != model.SchemaKindObject {
			return unsupportedSchema(fmt.Sprintf(
				"allOf branch %d has schema kind %q; multi-branch composition supports objects only",
				branch.index, branch.schema.Kind,
			)), nil
		}
		out.Sensitive = out.Sensitive || branch.schema.Sensitive
		for name, child := range branch.schema.Properties {
			if previous, exists := propertyBranch[name]; exists {
				return unsupportedSchema(fmt.Sprintf(
					"allOf property %q is declared by branches %d and %d",
					name, previous, branch.index,
				)), nil
			}
			propertyBranch[name] = branch.index
			out.Properties[name] = cloneSchema(child)
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

// applyAllOfScalarConstraints applies the scalar constraints the normalized
// model understands. Two enum declarations are intersected; incompatible
// formats and empty enum intersections remain unsupported rather than silently
// widening the generated Terraform schema.
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

// unsupportedAllOfOuterStructure rejects combinations whose structure lives
// both outside and inside allOf. The current Datadog subset uses only an outer
// type assertion plus annotation/constraint metadata; merging outer properties,
// variants, arrays, or maps would require broader JSON Schema intersection
// semantics.
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

// isAnnotationOnlySchema recognizes the sibling-only branch libopenapi creates
// for an OpenAPI 3.0 $ref with description/example siblings, plus the same form
// when authored explicitly. A completely empty schema is not an annotation: it
// remains an arbitrary/untyped value and must not be discarded.
func (n *schemaNormalizer) isAnnotationOnlySchema(s *base.Schema) bool {
	if s == nil || hasStructuralOrConstraintKeywords(s) {
		return false
	}
	hasExtension := s.Extensions != nil && orderedmap.Len(s.Extensions) > 0
	return s.Title != "" || s.Description != "" || s.Example != nil || len(s.Examples) > 0 ||
		s.Default != nil || hasExtension || n.isSensitive(s)
}

// hasStructuralOrConstraintKeywords distinguishes schemas that narrow values
// from genuine metadata-only overlays. Keep this deliberately conservative:
// accepting an unknown assertion here would discard it and widen the generated
// Terraform schema.
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
	return s.DynamicRef != "" || s.ContentSchema != nil
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

func childPath(parent, child string) string {
	if parent == "" {
		return child
	}
	if child == "[]" || child == "{}" {
		return parent + child
	}
	return parent + "." + child
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	return slices.Compact(values)
}

func intersectStrings(left, right []string) []string {
	allowed := make(map[string]struct{}, len(right))
	for _, value := range right {
		allowed[value] = struct{}{}
	}
	var intersection []string
	for _, value := range left {
		if _, ok := allowed[value]; ok {
			intersection = append(intersection, value)
		}
	}
	return sortedUniqueStrings(intersection)
}

// cloneSchema returns a deep copy so applying allOf metadata never mutates a
// normalized branch or any child reachable from it.
func cloneSchema(s *model.Schema) *model.Schema {
	if s == nil {
		return nil
	}
	out := *s
	out.Enum = append([]string(nil), s.Enum...)
	out.Required = append([]string(nil), s.Required...)
	out.Items = cloneSchema(s.Items)
	if s.Properties != nil {
		out.Properties = make(map[string]*model.Schema, len(s.Properties))
		for name, child := range s.Properties {
			out.Properties[name] = cloneSchema(child)
		}
	}
	if s.Variants != nil {
		out.Variants = make([]*model.Schema, len(s.Variants))
		for i, variant := range s.Variants {
			out.Variants[i] = cloneSchema(variant)
		}
	}
	if s.OneOf != nil {
		oneOf := *s.OneOf
		oneOf.Variants = make([]model.OneOfVariant, len(s.OneOf.Variants))
		for i, variant := range s.OneOf.Variants {
			oneOf.Variants[i] = variant
			oneOf.Variants[i].Schema = cloneSchema(variant.Schema)
		}
		if s.OneOf.Discriminator != nil {
			discriminator := *s.OneOf.Discriminator
			if s.OneOf.Discriminator.Mapping != nil {
				discriminator.Mapping = make(map[string]string, len(s.OneOf.Discriminator.Mapping))
				for key, value := range s.OneOf.Discriminator.Mapping {
					discriminator.Mapping[key] = value
				}
			}
			oneOf.Discriminator = &discriminator
		}
		out.OneOf = &oneOf
	}
	return &out
}

// classifyKind derives the SchemaKind from structure, not type alone. Precedence
// (first match wins, since a node can satisfy several at once):
// anyOf → unsupported; oneOf → one_of; properties → object; type:array+items →
// array; additionalProperties → map; a concrete scalar type → primitive; anything
// else (free-form/empty object, typeless leaf, itemless array) → json.
//
// oneOf and anyOf are intentionally distinct: oneOf is normalized into a typed
// envelope, while anyOf remains unsupported so its artifact fails rather than
// inheriting oneOf's exactly-one semantics.
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
		// no items has an unknown element type and falls through to JSON.
		return model.SchemaKindArray
	case isMap(s):
		// additionalProperties (and no declared properties) → dynamic-key map.
		return model.SchemaKindMap
	case isRepresentablePrimitive(s):
		// A concrete scalar leaf (string, integer, number, boolean).
		return model.SchemaKindPrimitive
	default:
		// A genuinely unconstrained value can be preserved losslessly as normalized
		// JSON. Constraint-bearing schemas without a usable type remain unsupported:
		// encoding them as JSON would silently discard their validation semantics.
		if isFreeFormSchema(s) {
			return model.SchemaKindJSON
		}
		return model.SchemaKindUnsupported
	}
}

func isFreeFormSchema(s *base.Schema) bool {
	if s == nil {
		return false
	}
	switch firstType(s) {
	case "object":
		return s.Properties == nil || orderedmap.Len(s.Properties) == 0
	case "array":
		return s.Items == nil
	case "":
		return !hasStructuralOrConstraintKeywords(s)
	default:
		return false
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

// collectionElement preserves recursively typed collection shapes. Terraform
// attr.Type values can represent list/map chains such as list(list(string)) and
// map(list(string)); a missing element schema is arbitrary JSON.
func collectionElement(elem *model.Schema) *model.Schema {
	if elem == nil {
		return &model.Schema{Kind: model.SchemaKindJSON}
	}
	return elem
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

// firstType returns the schema's first non-null declared type, or "" when the
// schema is untyped/null-only. Nullability is represented on OneOfSpec rather
// than as a primitive alternative.
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
