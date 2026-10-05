package model

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// Generated test scenarios and cassette bundles
//
// A GeneratedTestScenario is the single source of truth for one example-backed
// target: the generated acceptance test and the cassette it replays are both
// rendered from it. That is deliberate. If the test's request and the
// cassette's expected request were derived independently they could drift, the
// matcher would still pass, and the test would go green while asserting
// nothing about the code under test.
// ----------------------------------------------------------------------------

// RedactedPlaceholder is the only spelling a sensitive value may take once it
// leaves the materializer. It appears in rendered fixtures and diagnostics
// alike so a leaked secret is a visible diff rather than a silent one.
const RedactedPlaceholder = "[redacted]"

// retainedHeaders are the only headers a recorded interaction may carry. The
// provider's replay harness filters everything else, so retaining more would
// record bytes that can never participate in matching — and would risk
// committing an authorization header.
//
// This list must stay in parity with allowedHeaders in
// datadog/tests/provider_test.go, which is the harness's own allowlist. That
// one lives in a _test.go file in the root module, so it cannot be imported
// here and the two are kept equal by hand.
var retainedHeaders = []string{"Accept", "Content-Type"}

// RetainedHeaders returns the header allowlist in canonical order.
func RetainedHeaders() []string { return slices.Clone(retainedHeaders) }

// FilterRetainedHeaders copies only the allowlisted headers, canonicalizing
// their names. Callers assembling headers are expected to pass them through
// here; it is not a chokepoint, since rendering forwards whatever headers an
// interaction already carries.
func FilterRetainedHeaders(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := map[string][]string{}
	for name, values := range in {
		canonical := http.CanonicalHeaderKey(name)
		if slices.Contains(retainedHeaders, canonical) && len(values) > 0 {
			out[canonical] = slices.Clone(values)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ----------------------------------------------------------------------------
// Example selection
// ----------------------------------------------------------------------------

// ScenarioNameSingle is the scenario name recorded when a target's values came
// from singular `example` fields, which carry no name of their own.
const ScenarioNameSingle = "single"

// ScenarioNameDefault is the named example automatic selection prefers, being
// the convention the Datadog description follows.
const ScenarioNameDefault = "default"

// ExampleProvenance is one selected value's traceable origin. Every value that
// reaches a committed fixture has one, so a reviewer can follow bytes back to
// the description rather than trusting the generator.
type ExampleProvenance struct {
	// Location anchors the declaration.
	Location ExampleLocation
	// CandidateName is the named-example key, empty for a singular example.
	CandidateName string
	// Reference is the reusable-example reference the value resolved through,
	// if any.
	Reference string
}

// ExampleSelection is the deterministic, traceable choice of candidates for one
// target. It is recorded on the scenario rather than discarded after use
// because "which example did this come from" is a question reviewers ask of
// every generated fixture.
type ExampleSelection struct {
	// ExplicitNames maps operationId to a maintainer-supplied example name.
	// An explicit name must exist in every named set that operation requires;
	// a missing one is an error rather than a silent fallback.
	ExplicitNames map[string]string
	// ScenarioName is the resolved common name: an explicit name, "default", the
	// lexically smallest name common to all required sets, or "single".
	ScenarioName string
	// Provenance lists every selected value's origin, ordered stably so the
	// run report is byte-identical across runs.
	Provenance []ExampleProvenance
}

// Explicit reports whether the maintainer pinned a name for this operation.
func (s *ExampleSelection) Explicit(operationId string) (string, bool) {
	if s == nil || len(s.ExplicitNames) == 0 {
		return "", false
	}
	name, ok := s.ExplicitNames[operationId]
	return name, ok
}

// SortedExplicitOperations returns the operationIds carrying an explicit name,
// lexically ordered, so diagnostics about explicit selection never vary by map
// iteration order.
func (s *ExampleSelection) SortedExplicitOperations() []string {
	if s == nil || len(s.ExplicitNames) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(s.ExplicitNames))
}

// ----------------------------------------------------------------------------
// Materialized configuration
// ----------------------------------------------------------------------------

// MaterializedValue is one typed value bound to an attribute path.
type MaterializedValue struct {
	// Path is the dotted attribute path, e.g. data.attributes.name for a
	// request value or attributes.name for a Terraform value.
	Path string
	// Value is the typed value. Nil is a declared null, not absence.
	Value any
	// Sensitive records that this value was replaced and now holds a safe
	// stand-in rather than what the description declared.
	Sensitive bool
}

// MaterializedConfiguration is the provider-facing configuration and request
// values for one test state, built from a selected example scenario.
type MaterializedConfiguration struct {
	// TerraformValues render the HCL configuration block.
	TerraformValues []MaterializedValue
	// RequestValues construct the expected SDK request body and parameters.
	// They are kept apart from TerraformValues because the two shapes differ:
	// read-only values are excluded from both, but write-only values appear in
	// the request while taking the provider's write-only form in HCL.
	RequestValues []MaterializedValue
	// Identity is the canonical resource or data-source identifier used
	// throughout the scenario. One value, minted once: a cassette whose create
	// response and subsequent request URLs disagree cannot replay at all.
	Identity string
	// SensitiveReplacements maps an attribute path to the safe value
	// substituted there, so the writer can prove no declared secret survived.
	SensitiveReplacements map[string]string
}

// SortedSensitivePaths returns the replaced paths in lexical order, for stable
// reporting.
func (c *MaterializedConfiguration) SortedSensitivePaths() []string {
	if c == nil || len(c.SensitiveReplacements) == 0 {
		return nil
	}
	return slices.Sorted(maps.Keys(c.SensitiveReplacements))
}

// ----------------------------------------------------------------------------
// Scenario interactions
// ----------------------------------------------------------------------------

// InteractionRole names why an interaction exists in the trace. Refresh and
// destroy verification are distinct roles even though they issue the same read
// request, because the reason a request repeats is what a reviewer needs to
// understand an apparently duplicated interaction.
type InteractionRole string

const (
	InteractionRoleCreate  InteractionRole = "create"
	InteractionRoleRead    InteractionRole = "read"
	InteractionRoleSearch  InteractionRole = "search"
	InteractionRoleUpdate  InteractionRole = "update"
	InteractionRoleDelete  InteractionRole = "delete"
	InteractionRoleRefresh InteractionRole = "refresh"
	// InteractionRoleDestroyVerification is the post-destroy read the test
	// harness issues to confirm the object is gone; it expects a 404.
	InteractionRoleDestroyVerification InteractionRole = "destroy_verification"
)

// InteractionRequest is the expected request side of one interaction.
type InteractionRequest struct {
	// Method is the HTTP method, uppercase.
	Method string
	// URL is the canonical absolute URL with credential query parameters
	// already removed, matching the provider harness's comparison form.
	URL string
	// Body is the canonicalized request body, empty for a bodyless request.
	Body string
	// Headers holds only the retained allowlist.
	Headers map[string][]string
	// ContentLength is the byte length of Body.
	ContentLength int
}

// InteractionResponse is the expected response side of one interaction.
type InteractionResponse struct {
	// StatusCode is the numeric HTTP status.
	StatusCode int
	// StatusText is the recorded status line text, e.g. "200 OK".
	StatusText string
	// Body is the response body; empty is legitimate for a declared bodyless
	// response and is not the same as a missing example.
	Body string
	// Headers holds only the retained allowlist.
	Headers map[string][]string
	// ContentLength is the byte length of Body.
	ContentLength int
	// Duration is always zero. A recorded latency would make regeneration
	// non-deterministic while changing nothing about replay.
	Duration time.Duration
}

// ScenarioInteraction is one expected request/response pair in the generated
// test's trace.
type ScenarioInteraction struct {
	// Index is the zero-based stable position in the cassette.
	Index int
	// Role records why this interaction exists.
	Role InteractionRole
	// OperationId is the source OpenAPI operation.
	OperationId string
	// Request is the expected request.
	Request InteractionRequest
	// Response is the expected response.
	Response InteractionResponse
	// SourceExamples are the provenance entries that contributed values here.
	SourceExamples []ExampleProvenance
}

// ----------------------------------------------------------------------------
// Generated test scenario
// ----------------------------------------------------------------------------

// ScenarioStep is one Terraform apply step in the generated test.
type ScenarioStep struct {
	// Config is the rendered HCL for this step.
	Config string
	// Checks are the resource.TestCheckFunc expressions asserted after apply,
	// rendered as Go source fragments.
	Checks []string
	// State is the materialized configuration this step applies.
	State *MaterializedConfiguration
}

// GeneratedTestScenario is the validated, single source of truth rendered into
// both the generated acceptance test and its cassette bundle.
type GeneratedTestScenario struct {
	// ArtifactName is the Terraform artifact name without the datadog_ prefix.
	ArtifactName string
	// ArtifactKind distinguishes a resource from a data source.
	ArtifactKind ArtifactKind
	// TestFilePath is the generated test's path.
	TestFilePath string
	// TestFuncName is the Go test function, which doubles as the cassette
	// basename because the replay harness derives the cassette from t.Name().
	TestFuncName string
	// TerraformAddress is the address the test asserts against, e.g.
	// datadog_integration_twilio_account.foo.
	TerraformAddress string
	// Steps are the ordered Terraform steps: the initial apply, plus an update
	// step only when the examples provide a distinct coherent updated state.
	Steps []ScenarioStep
	// Selection is the example selection this scenario was built from.
	Selection ExampleSelection
}

// HasUpdateStep reports whether the examples supported a distinct updated
// state. An update step is added only when they do: an update whose response
// equals the create response asserts nothing.
func (s *GeneratedTestScenario) HasUpdateStep() bool {
	return s != nil && len(s.Steps) > 1
}

// Validate reports why the scenario cannot be rendered, or nil when it is
// sound. It enforces the invariants the renderers then rely on rather than
// re-checking, so a malformed scenario fails before any path is touched.
func (s *GeneratedTestScenario) Validate() error {
	switch {
	case s == nil:
		return fmt.Errorf("scenario is nil")
	case s.ArtifactName == "":
		return fmt.Errorf("scenario has no artifact name")
	case s.TestFuncName == "":
		return fmt.Errorf("scenario %q has no test function name", s.ArtifactName)
	case s.TerraformAddress == "":
		return fmt.Errorf("scenario %q has no Terraform address to assert against", s.ArtifactName)
	case len(s.Steps) == 0:
		return fmt.Errorf("scenario %q has no Terraform steps", s.ArtifactName)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Cassette bundle
// ----------------------------------------------------------------------------

// GeneratedMarker identifies a file tfgen owns. The writer will replace a file
// carrying it under the generated policy but never one without it, so the
// marker is what separates "regenerate my own output" from "overwrite someone's
// hand-recorded fixture".
const GeneratedMarker = "Generated by tfgen from OpenAPI examples. Do not edit by hand."

// CassetteOwnership classifies what already exists at a bundle's target paths.
type CassetteOwnership string

const (
	// CassetteOwnershipMissing means no bundle member exists; the default
	// policy writes only in this state.
	CassetteOwnershipMissing CassetteOwnership = "missing"
	// CassetteOwnershipGenerated means every member exists and carries the
	// tfgen marker.
	CassetteOwnershipGenerated CassetteOwnership = "generated"
	// CassetteOwnershipHandwritten means members exist without the marker;
	// replacing them takes an explicit all-files policy.
	CassetteOwnershipHandwritten CassetteOwnership = "handwritten"
	// CassetteOwnershipIncomplete means only some members exist, or their
	// markers disagree. Ambiguous ownership is never resolved by guessing.
	CassetteOwnershipIncomplete CassetteOwnership = "incomplete"
)

// CassetteBundle is the generated test, cassette, and freeze companion treated
// as one logical unit. All three are rendered and validated before any target
// path is modified, and a failed commit restores all of them: a half-written
// bundle is worse than none, because it replays as a confusing failure rather
// than an obvious absence.
type CassetteBundle struct {
	// TestPath and TestContent are the dedicated example-backed test.
	TestPath    string
	TestContent []byte
	// CassettePath and CassetteContent are the go-vcr YAML fixture.
	CassettePath    string
	CassetteContent []byte
	// FreezePath and FreezeContent are the RFC3339Nano timestamp companion.
	// It carries no marker of its own; its ownership derives from the cassette.
	FreezePath    string
	FreezeContent []byte
	// ContentHashes maps each path to a stable content hash, used for
	// unchanged detection and to verify a rollback restored the original.
	ContentHashes map[string]string
	// Ownership is what was found at the target paths before writing.
	Ownership CassetteOwnership
}

// Paths returns the bundle's three target paths in a stable order.
func (b *CassetteBundle) Paths() []string {
	if b == nil {
		return nil
	}
	return []string{b.TestPath, b.CassettePath, b.FreezePath}
}

// Complete reports whether all three members were rendered. The writer checks
// this before touching the filesystem.
func (b *CassetteBundle) Complete() bool {
	return b != nil &&
		b.TestPath != "" && len(b.TestContent) > 0 &&
		b.CassettePath != "" && len(b.CassetteContent) > 0 &&
		b.FreezePath != "" && len(b.FreezeContent) > 0
}

// ----------------------------------------------------------------------------
// Cassette result
// ----------------------------------------------------------------------------

// CassetteStatus is the per-target outcome.
type CassetteStatus string

const (
	CassetteStatusGenerated  CassetteStatus = "generated"
	CassetteStatusPreserved  CassetteStatus = "preserved"
	CassetteStatusIneligible CassetteStatus = "ineligible"
)

// CassetteWriteAction records what the writer did, kept separate from
// CassetteStatus so "generated" can be distinguished from "generated and it
// replaced a recording", which a reviewer must be told about explicitly.
type CassetteWriteAction string

const (
	CassetteWriteNone              CassetteWriteAction = "none"
	CassetteWriteCreated           CassetteWriteAction = "created"
	CassetteWriteUnchanged         CassetteWriteAction = "unchanged"
	CassetteWriteReplacedGenerated CassetteWriteAction = "replaced_generated"
	CassetteWriteReplacedRecorded  CassetteWriteAction = "replaced_recorded"
)

// InteractionSummary is the safe, reviewer-facing description of one
// interaction in the run report. It carries roles and anchors, never bodies:
// a request body may hold a replaced secret, and the report is committed.
type InteractionSummary struct {
	Index       int      `json:"index"`
	Role        string   `json:"role"`
	OperationId string   `json:"operation_id"`
	Method      string   `json:"method"`
	URL         string   `json:"url"`
	Status      int      `json:"status"`
	SourcePaths []string `json:"source_paths,omitempty"`
}

// CassetteResult is the structured per-target outcome reported by
// tfgen generate.
type CassetteResult struct {
	Name         string              `json:"name"`
	Kind         ArtifactKind        `json:"kind"`
	TestName     string              `json:"test_name"`
	Status       CassetteStatus      `json:"status"`
	WriteAction  CassetteWriteAction `json:"write_action"`
	TestPath     string              `json:"test_path,omitempty"`
	CassettePath string              `json:"cassette_path,omitempty"`
	FreezePath   string              `json:"freeze_path,omitempty"`
	// SelectedExample is the resolved scenario name, for traceability.
	SelectedExample string               `json:"selected_example,omitempty"`
	Interactions    []InteractionSummary `json:"interactions,omitempty"`
	Diagnostics     []Diagnostic         `json:"diagnostics,omitempty"`
}

// NewCassetteDiagnostic builds a diagnostic anchored at an example location.
// Callers pass a location rather than formatting one into the message so the
// anchor stays machine-readable and the message stays value-free: locations are
// addresses and are always safe to commit, candidate values are not.
func NewCassetteDiagnostic(severity DiagnosticSeverity, message string, loc ExampleLocation) Diagnostic {
	return Diagnostic{Severity: severity, Message: message, Location: loc.String()}
}

// Redact replaces every occurrence of a known sensitive value with
// RedactedPlaceholder. It is the last line of defense before a message reaches
// the run report, for the cases where a value reached a string despite the
// materializer having replaced it.
func Redact(message string, sensitiveValues []string) string {
	for _, value := range sensitiveValues {
		if value == "" {
			continue
		}
		message = strings.ReplaceAll(message, value, RedactedPlaceholder)
	}
	return message
}
