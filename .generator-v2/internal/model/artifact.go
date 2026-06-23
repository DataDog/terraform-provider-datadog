package model

import (
	"fmt"
	"strings"
	"unicode"
)

// BuildArtifact wraps a tracked Operation's response tree into an *Artifact and
// resolves its SDK call bindings. It sets Name/Kind/SourceFile/Schema and, for
// the data-source read, Lifecycle.Read (the datadog-api-client-go call) and
// Lifecycle.IdStrategy. The request side (Create/Update/Delete, GoRequestType)
// stays empty.
func BuildArtifact(op *Operation) (*Artifact, error) {
	if op == nil || op.Tracking == nil {
		return nil, fmt.Errorf("model: BuildArtifact requires a tracked operation")
	}
	if op.Tracking.Cardinality == CardinalityPlural {
		return buildPluralArtifact(op)
	}

	schema, err := BuildResponseTree(op.ResponseSchema)
	if err != nil {
		return nil, err
	}
	name := op.Tracking.ArtifactName
	return &Artifact{
		Name:        name,
		Kind:        op.Tracking.ArtifactKind,
		Cardinality: CardinalitySingular,
		Description: op.Tracking.TfDescription,
		Schema:      schema,
		SourceFile:  "datadog/fwprovider/data_source_datadog_" + name + ".go",
		Lifecycle: &LifecycleBindings{
			Read:       readCall(op),
			IdStrategy: op.Tracking.IdStrategy,
		},
	}, nil
}

// buildPluralArtifact derives a plural data-source artifact: its Schema is the
// scalar-filter leaves (from the query parameters) followed by the response
// tree, whose single top-level results array the emit builder turns into the
// items block. It records the list-call bindings (item type, optional-params
// type, pagination) and any dropped filters as info Diagnostics.
func buildPluralArtifact(op *Operation) (*Artifact, error) {
	itemsBlock, err := buildItemsBlock(op)
	if err != nil {
		return nil, err
	}
	filters, diags := buildFilterLeaves(op)
	attrs := filters
	if itemsBlock != nil {
		attrs = append(attrs, itemsBlock)
	}

	read := readCall(op)
	read.ItemType = op.ItemRefName
	read.Paginated = op.Pagination != nil
	// The SDK generates an optional-parameters struct iff the endpoint declares
	// query parameters (pagination params are query parameters); without one the
	// list call takes no optional argument.
	if len(op.QueryParams) > 0 {
		read.OptionalParamsType = op.OperationId + "OptionalParameters"
	}

	name := op.Tracking.ArtifactName
	return &Artifact{
		Name:        name,
		Kind:        op.Tracking.ArtifactKind,
		Cardinality: CardinalityPlural,
		Description: op.Tracking.TfDescription,
		Schema:      &AttributeTree{Attributes: attrs},
		SourceFile:  "datadog/fwprovider/data_source_datadog_" + name + ".go",
		Lifecycle: &LifecycleBindings{
			Read:       read,
			IdStrategy: op.Tracking.IdStrategy,
		},
		Diagnostics: diags,
	}, nil
}

// defaultResultsPath is the JSON:API response property holding a list's
// elements, used when no x-pagination resultsPath is declared.
const defaultResultsPath = "data"

// buildItemsBlock builds the plural items block from the results array alone
// (op.Pagination.ResultsPath, else "data"), so response siblings such as
// meta/links/included are dropped rather than emitted. Returns nil when the
// response declares no such array.
func buildItemsBlock(op *Operation) (*Attribute, error) {
	resultsPath := defaultResultsPath
	if op.Pagination != nil && op.Pagination.ResultsPath != "" {
		resultsPath = op.Pagination.ResultsPath
	}
	if op.ResponseSchema == nil || op.ResponseSchema.Kind != SchemaKindObject {
		return nil, nil
	}
	arr := op.ResponseSchema.Properties[resultsPath]
	if arr == nil {
		return nil, nil
	}
	return buildAttribute(arr, "response."+resultsPath, nestBlock)
}

// readCall resolves the datadog-api-client-go binding for op's read.
func readCall(op *Operation) *SDKCall {
	return &SDKCall{
		GoPackage:      "datadog" + strings.ToUpper(versionSegment(op.Path)),
		GoApiStruct:    tagToClassName(op.Tag) + "Api",
		GoMethod:       op.OperationId,
		GoResponseType: op.ResponseRefName,
	}
}

// buildFilterLeaves converts op's scalar query parameters into Optional
// top-level filter attributes. Pagination params are excluded (the SDK's
// pagination form handles them); array- and enum-valued params are dropped with
// an info Diagnostic rather than failing the build. The result preserves
// QueryParams' name order.
func buildFilterLeaves(op *Operation) ([]*Attribute, []Diagnostic) {
	var leaves []*Attribute
	var diags []Diagnostic
	for _, p := range op.QueryParams {
		if op.Pagination != nil && (p.Name == op.Pagination.LimitParam || p.Name == op.Pagination.PageParam) {
			continue
		}
		if reason := unsupportedFilterReason(p.Schema); reason != "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("dropped query parameter %q from filters: %s", p.Name, reason),
			})
			continue
		}
		tfType, goType, err := FrameworkType(p.Schema)
		if err != nil {
			diags = append(diags, Diagnostic{
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("dropped query parameter %q from filters: %v", p.Name, err),
			})
			continue
		}
		leaves = append(leaves, &Attribute{
			Path:        snakeCase(p.Name),
			TfType:      tfType,
			GoType:      goType,
			Format:      p.Schema.Format,
			Optional:    true,
			Description: p.Description,
		})
	}
	return leaves, diags
}

// unsupportedFilterReason reports why a query parameter cannot become a filter,
// or "" when it can. Array- and enum-valued params are deferred (their SDK
// optional-params field is a slice or named enum type a string filter cannot
// set); everything non-scalar is unsupported.
func unsupportedFilterReason(s *Schema) string {
	if s == nil {
		return "parameter has no schema"
	}
	switch s.Kind {
	case SchemaKindArray:
		return "array-valued query parameters are not supported as filters"
	case SchemaKindPrimitive:
		if len(s.Enum) > 0 {
			return "enum-valued query parameters are not supported as filters"
		}
		return ""
	default:
		return fmt.Sprintf("query parameter kind %q is not supported as a filter", s.Kind)
	}
}

// versionSegment returns the API version path segment immediately after "/api/",
// e.g. "/api/v2/incidents/config/types/{id}" → "v2". It returns "" when the path
// has no segment after "api", leaving the resolved GoPackage incomplete so the
// emit builder fail-slows on it rather than emitting a broken import.
func versionSegment(path string) string {
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for i, s := range segs {
		if s == "api" && i+1 < len(segs) {
			return segs[i+1]
		}
	}
	return ""
}

// tagToClassName converts an OpenAPI tag into the datadog-api-client-go API
// struct base name: non-alphanumeric runs become word breaks, each word is
// capitalized on its first rune, and in-word casing is preserved. So "org
// groups" → "OrgGroups" and "APM" → "APM". This deliberately differs from
// SdkName, which lower-cases acronyms ("APM" → "Apm").
func tagToClassName(tag string) string {
	var b strings.Builder
	for _, word := range strings.FieldsFunc(tag, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}) {
		runes := []rune(word)
		b.WriteRune(unicode.ToUpper(runes[0]))
		b.WriteString(string(runes[1:]))
	}
	return b.String()
}
