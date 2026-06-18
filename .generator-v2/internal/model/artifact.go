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
	schema, err := BuildResponseTree(op.ResponseSchema)
	if err != nil {
		return nil, err
	}
	name := op.Tracking.ArtifactName
	return &Artifact{
		Name:       name,
		Kind:       op.Tracking.ArtifactKind,
		Schema:     schema,
		SourceFile: "datadog/fwprovider/data_source_datadog_" + name + ".go",
		Lifecycle: &LifecycleBindings{
			Read: &SDKCall{
				GoPackage:      "datadog" + strings.ToUpper(versionSegment(op.Path)),
				GoApiStruct:    tagToClassName(op.Tag) + "Api",
				GoMethod:       op.OperationId,
				GoResponseType: op.ResponseRefName,
			},
			IdStrategy: op.Tracking.IdStrategy,
		},
	}, nil
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
