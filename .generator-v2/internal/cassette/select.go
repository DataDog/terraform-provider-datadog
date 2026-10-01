// Package cassette turns the example candidates the parser extracted into a
// replayable fixture: selecting one coherent set of values, materializing them
// into provider configuration and request bodies, and rendering the generated
// test and its cassette from a single scenario.
package cassette

import (
	"fmt"
	"slices"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Candidate selection
//
// A target's interactions must agree with one another: the name a create
// request sends has to be the name its response returns and the name the
// following read returns, or the replayed apply fails Terraform's own
// consistency check. OpenAPI gives no guarantee that two examples picked
// independently agree, so selection never mixes them — it chooses one
// *scenario* spanning every value the target needs, and declines when no such
// scenario exists rather than assembling an incoherent one.
// ----------------------------------------------------------------------------

// SetRole names which part of an operation a selected example set belongs to.
type SetRole string

const (
	SetRoleRequest   SetRole = "request"
	SetRoleResponse  SetRole = "response"
	SetRoleParameter SetRole = "parameter"
)

// SetKey identifies one example set within a target. Parameters are keyed by
// location and name because the same name in path and query is two distinct
// parameters.
type SetKey struct {
	OperationId string
	Role        SetRole
	// Detail is the parameter's "in:name" for a parameter set, the status code
	// for a response, and empty for a request body.
	Detail string
}

// String renders the key for diagnostics.
func (k SetKey) String() string {
	if k.Detail == "" {
		return fmt.Sprintf("%s %s", k.OperationId, k.Role)
	}
	return fmt.Sprintf("%s %s %s", k.OperationId, k.Role, k.Detail)
}

// SelectedSet is how one example set resolved. Exactly one of Candidate and
// Fallbacks is populated: a whole value a human wrote, or the property-level
// pieces the materializer must assemble into one. Keeping them distinct is
// what lets materialization prefer the former without re-deriving precedence.
type SelectedSet struct {
	Key SetKey
	// Candidate is the chosen whole-value example.
	Candidate *model.ExampleCandidate
	// Fallbacks are the schema-property candidates to assemble, used only when
	// no whole value was declared anywhere for this set.
	Fallbacks []*model.ExampleCandidate
}

// Resolved reports whether this set yielded anything usable.
func (s SelectedSet) Resolved() bool {
	return s.Candidate != nil || len(s.Fallbacks) > 0
}

// TargetSelection is the deterministic, traceable choice for one target.
type TargetSelection struct {
	// ScenarioName is the resolved common example name, or "single" when the
	// values came from singular `example` fields that carry no name.
	ScenarioName string
	// Sets are the resolved example sets, ordered by operation and role so the
	// result is stable across runs.
	Sets []SelectedSet
	// Provenance records every selected value's origin, in the same order.
	Provenance []model.ExampleProvenance
}

// Model returns the selection in the form the run report and the scenario
// carry, so provenance reaches the report without the cassette package's
// internals leaking into it.
func (t TargetSelection) Model() model.ExampleSelection {
	return model.ExampleSelection{
		ScenarioName: t.ScenarioName,
		Provenance:   slices.Clone(t.Provenance),
	}
}

// Set returns the resolution for one key.
func (t TargetSelection) Set(key SetKey) (SelectedSet, bool) {
	for _, set := range t.Sets {
		if set.Key == key {
			return set, true
		}
	}
	return SelectedSet{}, false
}

// IneligibleError reports that no coherent scenario spans a target's example
// sets. It names the sets involved rather than just failing, because the fix
// is always a change to the description and the author needs to know where.
type IneligibleError struct {
	// Reason is the rule that was not satisfied.
	Reason string
	// Sets are the keys the reason concerns, in stable order.
	Sets []SetKey
}

func (e *IneligibleError) Error() string {
	if len(e.Sets) == 0 {
		return e.Reason
	}
	names := make([]string, 0, len(e.Sets))
	for _, key := range e.Sets {
		names = append(names, key.String())
	}
	return fmt.Sprintf("%s (%s)", e.Reason, joinAnd(names))
}

// Select chooses one coherent scenario across every example set the given
// operations need. Operations are taken in the order supplied, which is the
// target's lifecycle order, so the resulting sets and provenance read the way
// the generated test will run.
//
// This is the initial selection: it resolves a target whose sets use singular
// examples, and one whose named sets share exactly one common name. A target
// offering several common names is ambiguous and declined here — preferring
// "default" and falling back to the lexically smallest common name is a
// documented rule that belongs with explicit per-operation overrides, not
// guessed at now.
func Select(operations []*model.Operation) (TargetSelection, error) {
	required := requiredSets(operations)
	if len(required) == 0 {
		return TargetSelection{}, &IneligibleError{Reason: "the target declares no example sets"}
	}

	if malformed := malformedKeys(required); len(malformed) > 0 {
		return TargetSelection{}, &IneligibleError{
			Reason: "example sets declare both `example` and `examples`, which are mutually exclusive",
			Sets:   malformed,
		}
	}

	scenarioName, err := resolveScenarioName(required)
	if err != nil {
		return TargetSelection{}, err
	}

	selection := TargetSelection{ScenarioName: scenarioName}
	var unresolved []SetKey
	for _, candidate := range required {
		resolved := resolveSet(candidate, scenarioName)
		if !resolved.Resolved() {
			unresolved = append(unresolved, candidate.key)
			continue
		}
		selection.Sets = append(selection.Sets, resolved)
		selection.Provenance = append(selection.Provenance, provenanceFor(resolved)...)
	}
	if len(unresolved) > 0 {
		return TargetSelection{}, &IneligibleError{
			Reason: "no usable example is declared for every interaction the test performs",
			Sets:   unresolved,
		}
	}
	return selection, nil
}

// ----------------------------------------------------------------------------
// Required sets
// ----------------------------------------------------------------------------

// requiredSet pairs one example set with the key identifying it.
type requiredSet struct {
	key SetKey
	set *model.ExampleSet
}

// requiredSets enumerates the example sets a target's interactions cannot be
// built without, in stable order.
//
// Two kinds of set are deliberately absent. A declared-bodyless response is a
// complete contract that needs no example, so demanding one would make every
// DELETE ineligible. And a path parameter carrying the target's identity is
// satisfied from the create response's minted id rather than from the
// description — requiring an example for it would reject almost every real
// resource, whose id parameter is a server-assigned value no author would
// write an example for.
func requiredSets(operations []*model.Operation) []requiredSet {
	identity := identityPathParameters(operations)
	var out []requiredSet
	for _, op := range operations {
		if op == nil {
			continue
		}
		if op.RequestExamples != nil && op.RequestExamples.Present {
			out = append(out, requiredSet{
				key: SetKey{OperationId: op.OperationId, Role: SetRoleRequest},
				set: &op.RequestExamples.Examples,
			})
		}
		if success := op.SuccessResponseExample(); success != nil && success.BodyPresent {
			out = append(out, requiredSet{
				key: SetKey{OperationId: op.OperationId, Role: SetRoleResponse, Detail: success.Status},
				set: &success.Examples,
			})
		}
		for i := range op.ParameterExamples {
			parameter := &op.ParameterExamples[i]
			// An optional parameter the description gives no example for is
			// simply omitted from the request; only a required one blocks.
			if !parameter.Required && parameter.Examples.Empty() {
				continue
			}
			// The identity parameter's value comes from the create response,
			// so it imposes no requirement on the description.
			if parameter.In == model.ParameterInPath && identity[parameter.Name] {
				continue
			}
			out = append(out, requiredSet{
				key: SetKey{
					OperationId: op.OperationId,
					Role:        SetRoleParameter,
					Detail:      fmt.Sprintf("%s:%s", parameter.In, parameter.Name),
				},
				set: &parameter.Examples,
			})
		}
	}
	return out
}

// identityPathParameters names the path parameters whose values the scenario
// mints rather than reads from the description.
//
// A create operation posts to a collection, so its path carries no identity;
// read, update and delete address one object and carry the id the create
// response returned. Any path parameter on those operations that the create
// path does not also declare is therefore identity-derived.
//
// A target with no create operation — a data source — mints nothing, so every
// path parameter there is genuinely user-supplied and does need an example.
func identityPathParameters(operations []*model.Operation) map[string]bool {
	var create *model.Operation
	for _, op := range operations {
		if op != nil && op.HasLifecycleRole(model.GroupRoleCreate) {
			create = op
			break
		}
	}
	if create == nil {
		return nil
	}
	onCreatePath := make(map[string]bool)
	for i := range create.ParameterExamples {
		parameter := create.ParameterExamples[i]
		if parameter.In == model.ParameterInPath {
			onCreatePath[parameter.Name] = true
		}
	}
	identity := make(map[string]bool)
	for _, op := range operations {
		if op == nil || op == create {
			continue
		}
		for i := range op.ParameterExamples {
			parameter := op.ParameterExamples[i]
			if parameter.In == model.ParameterInPath && !onCreatePath[parameter.Name] {
				identity[parameter.Name] = true
			}
		}
	}
	return identity
}

// malformedKeys returns the keys of sets declaring both example forms.
func malformedKeys(required []requiredSet) []SetKey {
	var out []SetKey
	for _, candidate := range required {
		if candidate.set.Malformed() {
			out = append(out, candidate.key)
		}
	}
	return out
}

// ----------------------------------------------------------------------------
// Scenario naming
// ----------------------------------------------------------------------------

// resolveScenarioName decides which named example the whole target uses, or
// reports why no single name serves it.
//
// Sets using only the singular form impose no name, so they never constrain
// the choice; a target with no named set at all is a "single" scenario.
func resolveScenarioName(required []requiredSet) (string, error) {
	var namedKeys []SetKey
	var common []string
	first := true
	for _, candidate := range required {
		names := candidate.set.SortedNames()
		if len(names) == 0 {
			continue
		}
		namedKeys = append(namedKeys, candidate.key)
		if first {
			common = names
			first = false
			continue
		}
		common = model.IntersectStrings(common, names)
	}
	if first {
		return model.ScenarioNameSingle, nil
	}
	switch len(common) {
	case 0:
		// Combining independently named candidates would invent a scenario the
		// description never described.
		return "", &IneligibleError{
			Reason: "named example sets share no common name, and independently named examples are never combined",
			Sets:   namedKeys,
		}
	case 1:
		return common[0], nil
	default:
		return "", &IneligibleError{
			Reason: fmt.Sprintf("named example sets share several common names (%s) and no selection rule has been applied",
				joinAnd(common)),
			Sets: namedKeys,
		}
	}
}

// ----------------------------------------------------------------------------
// Per-set resolution
// ----------------------------------------------------------------------------

// resolveSet picks the candidate for one set under the chosen scenario name,
// applying the precedence the source kinds encode: a named example for this
// scenario, then the singular form, then a whole-body schema example, and only
// then the property-level pieces. Each step is something a human wrote more
// deliberately than the next.
func resolveSet(candidate requiredSet, scenarioName string) SelectedSet {
	resolved := SelectedSet{Key: candidate.key}
	set := candidate.set

	if scenarioName != model.ScenarioNameSingle {
		if named, ok := set.Named[scenarioName]; ok && named.Eligible() {
			resolved.Candidate = named
			return resolved
		}
	}
	if set.Single.Eligible() {
		resolved.Candidate = set.Single
		return resolved
	}
	var properties []*model.ExampleCandidate
	for _, fallback := range set.SchemaFallback {
		if !fallback.Eligible() {
			continue
		}
		if fallback.SourceKind == model.ExampleSourceSchema {
			resolved.Candidate = fallback
			return resolved
		}
		properties = append(properties, fallback)
	}
	resolved.Fallbacks = properties
	return resolved
}

// provenanceFor records where a resolved set's values came from, so a reviewer
// can trace committed fixture bytes back to the description.
func provenanceFor(resolved SelectedSet) []model.ExampleProvenance {
	entries := make([]model.ExampleProvenance, 0, 1+len(resolved.Fallbacks))
	if resolved.Candidate != nil {
		entries = append(entries, provenanceOf(resolved.Candidate))
	}
	for _, fallback := range resolved.Fallbacks {
		entries = append(entries, provenanceOf(fallback))
	}
	return entries
}

func provenanceOf(candidate *model.ExampleCandidate) model.ExampleProvenance {
	return model.ExampleProvenance{
		Location:      candidate.Location,
		CandidateName: candidate.Name,
		Reference:     candidate.Reference,
	}
}

// joinAnd renders a short list the way prose would, for diagnostics a human
// reads rather than parses.
func joinAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}
