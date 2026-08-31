package model

import (
	"fmt"
	"strings"
	"unicode"
)

// BuildArtifact wraps a tracked Operation's response tree into an *Artifact and
// resolves its SDK call bindings. It sets Name/Kind/SourceFile/Schema and the
// data-source SDK calls (Lifecycle.Read for by-id, Lifecycle.Search for the list
// call) plus Lifecycle.IdStrategy. The request side (Create/Update/Delete,
// GoRequestType) stays empty.
//
// A singular data source resolves its one record three ways, selected by which
// group operations are declared: read only (by-id, the original behavior),
// search only (find one in a list), or both (by-id when an id is given, else
// search). Plural is unchanged.
func BuildArtifact(op *Operation) (*Artifact, error) {
	if op == nil || op.Tracking == nil {
		return nil, fmt.Errorf("model: BuildArtifact requires a tracked operation")
	}
	artifact, err := buildArtifact(op)
	if err != nil {
		return nil, err
	}
	// A group reference naming an operationId the spec does not declare is an
	// author error the run report must show, even when the artifact still builds
	// without that role.
	if diags := unresolvedGroupDiagnostics(op); len(diags) > 0 {
		artifact.Diagnostics = append(diags, artifact.Diagnostics...)
	}
	return artifact, nil
}

// buildArtifact selects the builder for op's cardinality and lookup shape.
func buildArtifact(op *Operation) (*Artifact, error) {
	if op.Tracking.Cardinality == CardinalityPlural {
		return buildPluralArtifact(op)
	}

	g := op.Tracking.Group
	hasRead := g != nil && g.Read != ""
	hasSearch := g != nil && g.Search != ""
	switch {
	case hasSearch && !hasRead:
		return buildSingularSearchArtifact(op)
	case hasSearch && hasRead:
		return buildSingularBothArtifact(op)
	default:
		// Read-only (group.read) or groupless: resolve the single record by id.
		return buildSingularByIdArtifact(op)
	}
}

// buildSingularByIdArtifact resolves the one record by direct id lookup: its
// Schema is the by-id response tree and Lifecycle.Read the get-by-id call.
func buildSingularByIdArtifact(op *Operation) (*Artifact, error) {
	schema, diags, err := BuildResponseTree(op.ResponseSchema)
	if err != nil {
		return nil, err
	}
	read := readCall(op, true)
	inputs, err := buildRequiredInputLeaves(read.Arguments)
	if err != nil {
		return nil, err
	}
	schema.Attributes = append(inputs, schema.Attributes...)
	return &Artifact{
		Name:        op.Tracking.ArtifactName,
		Kind:        op.Tracking.ArtifactKind,
		Cardinality: CardinalitySingular,
		Description: op.Tracking.TfDescription,
		Schema:      schema,
		SourceFile:  sourceFileFor(op.Tracking.ArtifactName),
		Lifecycle: &LifecycleBindings{
			Read:       read,
			IdStrategy: op.Tracking.IdStrategy,
		},
		Diagnostics: diags,
	}, nil
}

// buildSingularSearchArtifact resolves the one record by searching a list: op is
// the list operation itself (group.search names it). The flat record is the list
// element reshaped into a singular {data:{…}} envelope so the by-id envelope
// flattener serves it unchanged; the search side adds Optional filters from the
// query parameters and the list call. No list/items block is emitted.
func buildSingularSearchArtifact(op *Operation) (*Artifact, error) {
	element := listElementSchema(op)
	if element == nil {
		return nil, fmt.Errorf("model: search data source %q has no result-array element to flatten", op.Tracking.ArtifactName)
	}
	record, recDiags, err := BuildResponseTree(singularEnvelope(element))
	if err != nil {
		return nil, err
	}
	search := listCall(op)
	inputs, err := buildRequiredInputLeaves(search.Arguments)
	if err != nil {
		return nil, err
	}
	filters, diags := buildFilterLeaves(op)
	diags = append(diags, recDiags...)
	return &Artifact{
		Name:        op.Tracking.ArtifactName,
		Kind:        op.Tracking.ArtifactKind,
		Cardinality: CardinalitySingular,
		Description: op.Tracking.TfDescription,
		Schema:      &AttributeTree{Attributes: append(append(inputs, filters...), record.Attributes...)},
		SourceFile:  sourceFileFor(op.Tracking.ArtifactName),
		Lifecycle: &LifecycleBindings{
			Search:     search,
			IdStrategy: op.Tracking.IdStrategy,
		},
		Diagnostics: diags,
	}, nil
}

// buildSingularBothArtifact resolves the one record by id when given one and by
// search otherwise. The flat record is the canonical by-id response tree (the
// same element shape the search returns); the search side adds Optional filters
// from the list op's query parameters and the list call, alongside the by-id Read.
func buildSingularBothArtifact(op *Operation) (*Artifact, error) {
	searchOp := op.ResolvedGroup.Op(GroupRoleSearch)
	if searchOp == nil {
		return nil, fmt.Errorf("model: data source %q declares group.search %q but no such operation exists",
			op.Tracking.ArtifactName, op.Tracking.Group.Search)
	}

	// One state mapper serves both lookups only when the by-id record and the
	// list element are the same shape. Stay "both" only when both $ref names are
	// known and agree; otherwise — an inline schema on either side (empty ref) or
	// names that differ (e.g. api_key: FullAPIKey vs PartialAPIKey) — degrade to
	// by-id-only (the full shape, id required) rather than risk a mapper that
	// silently reads the wrong fields, and record why.
	if op.ResponseDataRefName == "" || searchOp.ItemRefName == "" ||
		op.ResponseDataRefName != searchOp.ItemRefName {
		art, err := buildSingularByIdArtifact(op)
		if err != nil {
			return nil, err
		}
		art.Diagnostics = append(art.Diagnostics, Diagnostic{
			Severity: SeverityInfo,
			Message: fmt.Sprintf(
				"data source %q: search lookup dropped — could not confirm the by-id record shape matches the list element (by-id ref %q, list ref %q); generated by-id-only",
				op.Tracking.ArtifactName, op.ResponseDataRefName, searchOp.ItemRefName),
		})
		return art, nil
	}

	record, recDiags, err := BuildResponseTree(op.ResponseSchema)
	if err != nil {
		return nil, err
	}
	read := readCall(op, true)
	search := listCall(searchOp)
	inputs, err := mergeRequiredInputLeaves(read.Arguments, search.Arguments)
	if err != nil {
		return nil, err
	}
	filters, diags := buildFilterLeaves(searchOp)
	diags = append(diags, recDiags...)
	return &Artifact{
		Name:        op.Tracking.ArtifactName,
		Kind:        op.Tracking.ArtifactKind,
		Cardinality: CardinalitySingular,
		Description: op.Tracking.TfDescription,
		Schema:      &AttributeTree{Attributes: append(append(inputs, filters...), record.Attributes...)},
		SourceFile:  sourceFileFor(op.Tracking.ArtifactName),
		Lifecycle: &LifecycleBindings{
			Read:       read,
			Search:     search,
			IdStrategy: op.Tracking.IdStrategy,
		},
		Diagnostics: diags,
	}, nil
}

// unresolvedGroupDiagnostics reports every group reference whose operationId
// matches no operation in the spec, so an author's typo is never dropped
// silently.
//
// It reports rather than fails, because whether a dangling reference is fatal
// depends on the role: a builder that cannot proceed without one fails the
// artifact itself with a message naming the role (buildSingularBothArtifact
// does exactly that for a dangling group.search, and the resource lifecycle
// builder will for the CRUD roles), and those returns never reach here. What is
// left for this to surface is the remainder — a reference to a role this
// artifact shape does not consume, which would otherwise vanish unnoticed.
func unresolvedGroupDiagnostics(op *Operation) []Diagnostic {
	if op.ResolvedGroup == nil || len(op.ResolvedGroup.Unresolved) == 0 {
		return nil
	}
	diags := make([]Diagnostic, 0, len(op.ResolvedGroup.Unresolved))
	for _, ref := range op.ResolvedGroup.Unresolved {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Message: fmt.Sprintf("artifact %q: group.%s names operationId %q, which no operation in the spec declares; that role is unbound",
				op.Tracking.ArtifactName, ref.Role, ref.OperationId),
		})
	}
	return diags
}

// sourceFileFor is the output path for a data-source artifact name.
func sourceFileFor(name string) string {
	return "datadog/fwprovider/data_source_datadog_" + name + ".go"
}

// singularEnvelope wraps a list element schema in a one-property {data: element}
// object, mimicking a by-id response so the singular envelope flattener can hoist
// the element's leaves regardless of where the record was fetched.
func singularEnvelope(element *Schema) *Schema {
	return &Schema{
		Kind:       SchemaKindObject,
		Properties: map[string]*Schema{defaultResultsPath: element},
	}
}

// listElementSchema returns the element schema of op's results array
// (op.Pagination.ResultsPath, else "data"), or nil when that property is absent
// or not an array.
func listElementSchema(op *Operation) *Schema {
	resultsPath := defaultResultsPath
	if op.Pagination != nil && op.Pagination.ResultsPath != "" {
		resultsPath = op.Pagination.ResultsPath
	}
	if op.ResponseSchema == nil || op.ResponseSchema.Kind != SchemaKindObject {
		return nil
	}
	arr := op.ResponseSchema.Properties[resultsPath]
	if arr == nil || arr.Kind != SchemaKindArray {
		return nil
	}
	return arr.Items
}

// buildPluralArtifact derives a plural data-source artifact: its Schema is the
// scalar-filter leaves (from the query parameters) followed by the response
// tree, whose single top-level results array the emit builder turns into the
// items block. It records the list-call bindings (item type, optional-params
// type, pagination) and any dropped filters as info Diagnostics.
func buildPluralArtifact(op *Operation) (*Artifact, error) {
	itemsBlock, itemDiags, err := buildItemsBlock(op)
	if err != nil {
		return nil, err
	}
	read := listCall(op)
	inputs, err := buildRequiredInputLeaves(read.Arguments)
	if err != nil {
		return nil, err
	}
	filters, diags := buildFilterLeaves(op)
	diags = append(diags, itemDiags...)
	attrs := append(inputs, filters...)
	if itemsBlock != nil {
		attrs = append(attrs, itemsBlock)
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
// meta/links/included are dropped rather than emitted. Returns a nil Attribute
// when the response declares no such array, plus the diagnostics raised while
// building the element (none today: an element either projects or fails).
func buildItemsBlock(op *Operation) (*Attribute, []Diagnostic, error) {
	resultsPath := defaultResultsPath
	if op.Pagination != nil && op.Pagination.ResultsPath != "" {
		resultsPath = op.Pagination.ResultsPath
	}
	if op.ResponseSchema == nil || op.ResponseSchema.Kind != SchemaKindObject {
		return nil, nil, nil
	}
	arr := op.ResponseSchema.Properties[resultsPath]
	if arr == nil {
		return nil, nil, nil
	}
	required := false
	for _, name := range op.ResponseSchema.Required {
		if name == resultsPath {
			required = true
			break
		}
	}
	attr, err := (&treeBuilder{kind: responseTree}).attribute(arr, "response."+resultsPath, nestBlock, required)
	return attr, nil, err
}

// SDKPackageForPath returns the versioned datadog-api-client-go package an
// operation path belongs to, e.g. "/api/v2/teams/{id}" → "datadogV2". A path with
// no version segment yields the bare "datadog" prefix, which is deliberately not
// a real package: the emit builder fail-slows on it rather than emitting a broken
// import.
//
// It is the one place this derivation lives, so the emitter, the SDK oneOf
// binding pass and its corroboration test all agree on which package a union's
// wrapper is declared in.
func SDKPackageForPath(path string) string {
	return "datadog" + strings.ToUpper(versionSegment(path))
}

// readCall resolves the datadog-api-client-go binding for op's read.
func readCall(op *Operation, aliasTerminalID bool) *SDKCall {
	call := &SDKCall{
		GoPackage:      SDKPackageForPath(op.Path),
		GoApiStruct:    tagToClassName(op.Tag) + "Api",
		GoMethod:       op.OperationId,
		GoResponseType: op.ResponseRefName,
	}
	applySDKBinding(call, op, aliasTerminalID)
	return call
}

// listCall resolves the datadog-api-client-go binding for op's list call: the
// base read binding plus the element type, the optional-parameters struct (the
// SDK generates one iff the endpoint declares query parameters — pagination
// params are query parameters), and the pagination flag. Shared by the plural
// path and the singular search path.
func listCall(op *Operation) *SDKCall {
	c := readCall(op, false)
	c.ItemType = op.ItemRefName
	c.Paginated = op.Pagination != nil
	if op.SDKBinding == nil && len(op.QueryParams) > 0 {
		c.OptionalParamsType = op.OperationId + "OptionalParameters"
	}
	return c
}

func applySDKBinding(call *SDKCall, op *Operation, aliasTerminalID bool) {
	if op.SDKBinding == nil {
		return
	}
	call.BindingResolved = true
	call.OptionalParamsType = op.SDKBinding.OptionalParamsType
	for _, raw := range op.SDKBinding.Required {
		arg := raw
		arg.TFName = SnakeCase(arg.Name)
		if aliasTerminalID && arg.Location == "path" && strings.HasSuffix(strings.TrimRight(op.Path, "/"), "/{"+arg.Name+"}") {
			arg.TFName = "id"
		}
		call.Arguments = append(call.Arguments, arg)
	}
	for _, raw := range op.SDKBinding.Optional {
		arg := raw
		arg.TFName = SnakeCase(arg.Name)
		call.OptionalArguments = append(call.OptionalArguments, arg)
	}
}

func buildRequiredInputLeaves(arguments []SDKArgument) ([]*Attribute, error) {
	var leaves []*Attribute
	for _, arg := range arguments {
		if arg.TFName == "id" {
			continue
		}
		if arg.Schema == nil || arg.Schema.Kind != SchemaKindPrimitive {
			return nil, fmt.Errorf("model: SDK argument %s has unsupported scalar-first type %s", arg.Name, arg.GoType)
		}
		tfType, goType, err := FrameworkType(arg.Schema)
		if err != nil {
			return nil, fmt.Errorf("model: SDK argument %s: %w", arg.Name, err)
		}
		description := arg.Description
		if description == "" {
			description = fmt.Sprintf("The %s argument.", arg.Name)
		}
		leaves = append(leaves, &Attribute{
			Path: arg.TFName, TfType: tfType, GoType: goType, Format: arg.Schema.Format,
			Required: true, Description: description,
		})
	}
	return leaves, nil
}

func mergeRequiredInputLeaves(argumentGroups ...[]SDKArgument) ([]*Attribute, error) {
	seen := map[string]*Attribute{}
	var merged []*Attribute
	for _, arguments := range argumentGroups {
		leaves, err := buildRequiredInputLeaves(arguments)
		if err != nil {
			return nil, err
		}
		for _, leaf := range leaves {
			if prior, ok := seen[leaf.Path]; ok {
				if prior.TfType != leaf.TfType {
					return nil, fmt.Errorf("model: SDK argument %s has conflicting Terraform types %s and %s", leaf.Path, prior.TfType, leaf.TfType)
				}
				continue
			}
			seen[leaf.Path] = leaf
			merged = append(merged, leaf)
		}
	}
	return merged, nil
}

// buildFilterLeaves converts op's scalar query parameters into Optional
// top-level filter attributes. Pagination params are excluded (the SDK's
// pagination form handles them); array- and enum-valued params are dropped with
// an info Diagnostic rather than failing the build. Required query parameters
// are surfaced as info diagnostics because the current SDK-call binding cannot
// represent required query arguments. The result preserves QueryParams' order.
func buildFilterLeaves(op *Operation) ([]*Attribute, []Diagnostic) {
	var leaves []*Attribute
	var diags []Diagnostic
	boundRequired := map[string]bool{}
	if op.SDKBinding != nil {
		for _, arg := range op.SDKBinding.Required {
			if arg.Location == "query" {
				boundRequired[arg.Name] = true
			}
		}
	}
	for _, p := range op.QueryParams {
		if boundRequired[p.Name] {
			continue
		}
		if op.Pagination != nil && (p.Name == op.Pagination.LimitParam || p.Name == op.Pagination.PageParam) {
			continue
		}
		if reason := unsupportedFilterReason(p.Schema); reason != "" {
			required := ""
			if p.Required {
				required = "required "
			}
			diags = append(diags, Diagnostic{
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("dropped %squery parameter %q from filters: %s", required, p.Name, reason),
			})
			continue
		}
		tfType, goType, err := FrameworkType(p.Schema)
		if err != nil {
			required := ""
			if p.Required {
				required = "required "
			}
			diags = append(diags, Diagnostic{
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("dropped %squery parameter %q from filters: %v", required, p.Name, err),
			})
			continue
		}
		if p.Required {
			diags = append(diags, Diagnostic{
				Severity: SeverityInfo,
				Message: fmt.Sprintf(
					"required query parameter %q is represented as an optional Terraform filter; generated example may be incomplete and requires review",
					p.Name,
				),
			})
		}
		leaves = append(leaves, &Attribute{
			Path:        SnakeCase(p.Name),
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
