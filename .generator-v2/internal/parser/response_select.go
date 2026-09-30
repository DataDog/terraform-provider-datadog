package parser

import (
	"sort"
	"strconv"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ----------------------------------------------------------------------------
// Shared response and parameter selection
//
// Schema normalization and example extraction must agree on two questions:
// which response an operation succeeds with, and which parameters it takes. If
// they answered independently they could diverge — a cassette built from the
// 201 while the provider model came from the 200 would replay a body the
// generated code cannot decode, and the mismatch would surface as an
// inscrutable replay failure rather than a generation error. These helpers are
// the single answer both paths use.
// ----------------------------------------------------------------------------

// SelectedResponse is one chosen response outcome and the detail needed to
// describe it: the status as written, whether it declares a body, and the
// media type and schema when it does.
type SelectedResponse struct {
	// Status is the status code as written in the description, e.g. "201".
	Status string
	// Code is Status parsed as an integer.
	Code int
	// Response is the underlying libopenapi response.
	Response *v3.Response
	// MediaType is the selected media type, empty when no body is declared.
	MediaType string
	// Schema is the selected media type's schema proxy, nil when no body is
	// declared.
	Schema *base.SchemaProxy
}

// BodyPresent reports whether this response declares a body at all. A false
// value is a complete contract for a 204 or 205 and a gap for anything else —
// the distinction the cassette pipeline turns on when deciding whether a
// missing example is an error.
func (r SelectedResponse) BodyPresent() bool {
	return r.Schema != nil
}

// jsonMediaTypeOf returns the application/json media type of a content map, or
// nil when the map is absent or declares no JSON representation. Selecting
// exactly one media type is what keeps the projected schema deterministic.
func jsonMediaTypeOf(content *orderedmap.Map[string, *v3.MediaType]) *v3.MediaType {
	if content == nil {
		return nil
	}
	return content.GetOrZero(jsonMediaType)
}

// successResponses returns every declared 2xx outcome in ascending numeric
// order. Ordering by number rather than declaration makes the choice
// deterministic across spec reorderings.
func successResponses(op *v3.Operation) []SelectedResponse {
	if op == nil || op.Responses == nil || op.Responses.Codes == nil {
		return nil
	}
	var out []SelectedResponse
	for code, resp := range op.Responses.Codes.FromOldest() {
		num, err := strconv.Atoi(code)
		if err != nil || num < 200 || num > 299 || resp == nil {
			continue
		}
		out = append(out, newSelectedResponse(code, num, resp))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// DeclaredResponses returns every declared response outcome with a parseable
// numeric status, in ascending order — success and failure alike. The failure
// outcomes matter to cassette generation: a generated resource test reads once
// more after destroy and expects a 404. Range codes such as "4XX" are skipped,
// since a range names no single status a recorded interaction could carry.
func DeclaredResponses(op *v3.Operation) []SelectedResponse {
	if op == nil || op.Responses == nil || op.Responses.Codes == nil {
		return nil
	}
	var out []SelectedResponse
	for code, resp := range op.Responses.Codes.FromOldest() {
		num, err := strconv.Atoi(code)
		if err != nil || resp == nil {
			continue
		}
		out = append(out, newSelectedResponse(code, num, resp))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// newSelectedResponse resolves one response's body presence and schema.
func newSelectedResponse(status string, code int, resp *v3.Response) SelectedResponse {
	selected := SelectedResponse{Status: status, Code: code, Response: resp}
	if mt := jsonMediaTypeOf(resp.Content); mt != nil && mt.Schema != nil {
		selected.MediaType = jsonMediaType
		selected.Schema = mt.Schema
	}
	return selected
}

// SelectSuccessResponse returns the lowest-numbered 2xx response that declares
// a JSON body, or the zero value when none does. Codes without a body are
// skipped rather than chosen, because a bodyless success carries no schema for
// the provider model to be built from.
//
// This is the one selection both schema normalization and example extraction
// use; keeping it in one place is what guarantees they agree.
func SelectSuccessResponse(op *v3.Operation) (SelectedResponse, bool) {
	for _, candidate := range successResponses(op) {
		if candidate.BodyPresent() {
			return candidate, true
		}
	}
	return SelectedResponse{}, false
}

// SelectRequestBody returns an operation's application/json request-body media
// type along with whether the body is required, and whether one is declared at
// all. Presence is reported separately from the schema so a caller can tell
// "this operation sends nothing" from "its body has no usable schema".
func SelectRequestBody(op *v3.Operation) (mt *v3.MediaType, required bool, present bool) {
	if op == nil || op.RequestBody == nil {
		return nil, false, false
	}
	required = op.RequestBody.Required != nil && *op.RequestBody.Required
	mt = jsonMediaTypeOf(op.RequestBody.Content)
	return mt, required, mt != nil
}

// ----------------------------------------------------------------------------
// Parameter merging
// ----------------------------------------------------------------------------

// parameterKey identifies a parameter for merging. OpenAPI makes (name,
// location) the identity, so the same name in path and query is two distinct
// parameters rather than a conflict.
type parameterKey struct {
	name string
	in   string
}

// MergeParameters combines a path item's parameters with one operation's,
// keyed by (name, location), with the operation's declaration overriding the
// path item's. Path-item parameters keep their position ahead of
// operation-only ones so declaration order stays stable and reproducible.
//
// Without this an operation silently loses every parameter its path item
// declares — including the path parameter naming the object it acts on, which
// would make the recorded URL unbuildable.
func MergeParameters(pathItem, operation []*v3.Parameter) []*v3.Parameter {
	overrides := make(map[parameterKey]*v3.Parameter, len(operation))
	for _, p := range operation {
		if p == nil || p.Name == "" {
			continue
		}
		overrides[parameterKey{name: p.Name, in: p.In}] = p
	}

	var merged []*v3.Parameter
	seen := make(map[parameterKey]bool, len(pathItem)+len(operation))
	for _, p := range pathItem {
		if p == nil || p.Name == "" {
			continue
		}
		key := parameterKey{name: p.Name, in: p.In}
		if seen[key] {
			continue
		}
		seen[key] = true
		if override, ok := overrides[key]; ok {
			merged = append(merged, override)
			continue
		}
		merged = append(merged, p)
	}
	for _, p := range operation {
		if p == nil || p.Name == "" {
			continue
		}
		key := parameterKey{name: p.Name, in: p.In}
		if seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, p)
	}
	return merged
}
