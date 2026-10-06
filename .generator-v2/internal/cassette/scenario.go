package cassette

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Scenario construction
//
// A scenario is the single source of truth for one example-backed target: the
// generated test and the cassette it replays are both rendered from it, so the
// request the test sends and the request the cassette expects cannot drift.
//
// Two values tie the whole trace together. The identity is minted once from
// the create response and reused by every later URL — a cassette whose create
// response and subsequent request targets disagree cannot replay at all. The
// freeze time is fixed, which is what makes the generated unique name
// reproducible without a recording.
// ----------------------------------------------------------------------------

// How many reads one applied step contributes to the trace.
//
// The framework plans against the API once after each apply, and the following
// step refreshes again before its own apply. So a step with another after it
// contributes two reads, and the last step contributes one. Destroy does not
// refresh: the read after the delete is the destroy verification's own, added
// separately.
//
// This is framework behavior rather than anything the description implies, so
// offline replay is the arbiter and these constants are what change if the
// pinned framework disagrees. Replay is also what corrected them: giving the
// final step two reads leaves one unconsumed, the destroy verification matches
// that stale 200 instead of the 404, and the failure reads "still exists after
// destroy" — naming the symptom and not the count.
//
// The 344 single-resource CRUD cassettes already in this repository record more
// than this — five or six interactions for a create-only flow where a generated
// one records four. The difference is that a hand-written test's check
// functions call the API themselves, which a generated test never does: its
// checks read state only, because the cassette is the oracle.
//
// Letting the recorder replay an interaction more than once looks like it would
// make all of this irrelevant. It does not, and the idea is a trap worth naming
// here so it is not retried. go-vcr's lookup is
//
//	for _, i := range c.Interactions {
//		if (c.ReplayableInteractions || !i.replayed) && c.Matcher(r, i.Request) {
//
// so setting ReplayableInteractions bypasses the consumed check and the lookup
// always returns the *first* match, never advancing. Every read in this trace
// shares a method and URL while returning a different body — create state, then
// updated state, then a 404 once destroyed — so with repeats enabled the
// post-update refresh serves stale state and the destroy verification can never
// reach its 404. Measured against a recorded cassette whose reads are 200, 200,
// 404: strict playback yields exactly that, and repeats yield 200, 200, 200.
//
// The count being predictable therefore matters, and the deliberately
// duplicated refreshes are load-bearing rather than redundant.
const (
	readsPerStep         = 1
	readsBeforeNextStep  = 1
	readsAfterMiddleStep = readsPerStep + readsBeforeNextStep
)

// ResourceTarget is everything scenario construction needs about one resource.
type ResourceTarget struct {
	// ArtifactName is the Terraform artifact name without the datadog_ prefix.
	ArtifactName string
	// ServerURL is the resolved origin recorded request URLs are built on.
	ServerURL string
	// IdStrategy says where in a response the canonical identifier lives.
	IdStrategy model.IdStrategy
	// FreezeTime is the fixed UTC instant the replay harness restores.
	FreezeTime time.Time
	// Create, Read, Update and Delete are the lifecycle operations. Update may
	// be nil; the others are required.
	Create, Read, Update, Delete *model.Operation
	// Selection is the chosen scenario these interactions are built from.
	Selection TargetSelection
}

// BuildResourceScenario assembles the ordered interaction trace and Terraform
// steps for one resource target.
//
// The trace follows the standard generated flow — create, refresh, destroy,
// verify — and adds an update step only when the examples describe a distinct
// updated state. An update whose response equals the create response asserts
// nothing, so a scenario is not given a step it cannot check.
func BuildResourceScenario(target ResourceTarget) (*model.GeneratedTestScenario, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}

	createRequest, createResponse, err := target.materializePair(target.Create)
	if err != nil {
		return nil, err
	}

	identity, ok := IdentityFrom(createResponse.Body, target.IdStrategy)
	if !ok {
		return nil, fmt.Errorf(
			"scenario %q: the create response example carries no %s to use as the resource identity",
			target.ArtifactName, identityPathLabel(target.IdStrategy))
	}

	readResponse, err := target.materializeResponse(target.Read)
	if err != nil {
		return nil, err
	}

	testName := resourceTestFuncName(target.ArtifactName)
	scenario := &model.GeneratedTestScenario{
		ArtifactName:     target.ArtifactName,
		ArtifactKind:     model.ArtifactKindResource,
		TestFilePath:     resourceTestFilePath(target.ArtifactName),
		TestFuncName:     testName,
		TerraformAddress: fmt.Sprintf("datadog_%s.foo", target.ArtifactName),
		CassetteBaseName: testName,
		FreezeTime:       target.FreezeTime.UTC(),
		Selection:        target.Selection.Model(),
	}

	// Materialized before the create's reads are added, because how many reads
	// the create step contributes depends on whether a step follows it.
	updated, hasUpdate, err := target.updateStep(createRequest, createResponse)
	if err != nil {
		return nil, err
	}

	builder := &traceBuilder{target: target, identity: identity}

	// Create, then the reads the framework issues after its apply.
	builder.add(model.InteractionRoleCreate, target.Create, createRequest, createResponse, http.StatusCreated)
	createReads := readsPerStep
	if hasUpdate {
		createReads = readsAfterMiddleStep
	}
	builder.addRefreshes(target.Read, readResponse, createReads)
	scenario.Steps = append(scenario.Steps, model.ScenarioStep{State: createRequest.configuration()})

	// An update step, only when the examples give a distinct state to assert.
	if hasUpdate {
		builder.add(model.InteractionRoleUpdate, target.Update, updated.request, updated.response, http.StatusOK)
		builder.addRefreshes(target.Read, updated.response, readsPerStep)
		scenario.Steps = append(scenario.Steps, model.ScenarioStep{State: updated.request.configuration()})
	}

	builder.addDelete()
	builder.addDestroyVerification()

	scenario.Interactions = builder.interactions
	if err := scenario.Validate(); err != nil {
		return nil, err
	}
	return scenario, nil
}

// validate reports a target that cannot describe a resource lifecycle.
func (t ResourceTarget) validate() error {
	switch {
	case t.ArtifactName == "":
		return fmt.Errorf("scenario: no artifact name")
	case t.ServerURL == "":
		return fmt.Errorf("scenario %q: no server URL to record request targets against", t.ArtifactName)
	case t.FreezeTime.IsZero():
		return fmt.Errorf("scenario %q: no freeze time", t.ArtifactName)
	case t.Create == nil:
		return fmt.Errorf("scenario %q: no create operation", t.ArtifactName)
	case t.Read == nil:
		return fmt.Errorf("scenario %q: no read operation", t.ArtifactName)
	case t.Delete == nil:
		return fmt.Errorf("scenario %q: no delete operation", t.ArtifactName)
	}
	return nil
}

// materializePair materializes one operation's request and success response.
func (t ResourceTarget) materializePair(op *model.Operation) (MaterializedSet, MaterializedSet, error) {
	request, err := t.materializeRequest(op)
	if err != nil {
		return MaterializedSet{}, MaterializedSet{}, err
	}
	response, err := t.materializeResponse(op)
	if err != nil {
		return MaterializedSet{}, MaterializedSet{}, err
	}
	return request, response, nil
}

func (t ResourceTarget) materializeRequest(op *model.Operation) (MaterializedSet, error) {
	// An operation that declares no request body sends none. Unusual for a
	// create, but legitimate when every input travels in the path or query,
	// and a hard failure here would reject a shape the API permits.
	if op.RequestExamples == nil || !op.RequestExamples.Present {
		return MaterializedSet{}, nil
	}
	key := SetKey{OperationId: op.OperationId, Role: SetRoleRequest}
	set, ok := t.Selection.Set(key)
	if !ok {
		return MaterializedSet{}, fmt.Errorf("scenario %q: no selected request example for %s",
			t.ArtifactName, op.OperationId)
	}
	var schema *model.Schema
	if op.RequestExamples != nil {
		schema = op.RequestExamples.Schema
	}
	return MaterializeSet(set, schema)
}

func (t ResourceTarget) materializeResponse(op *model.Operation) (MaterializedSet, error) {
	success := op.SuccessResponseExample()
	if success == nil {
		return MaterializedSet{}, fmt.Errorf("scenario %q: %s declares no success response",
			t.ArtifactName, op.OperationId)
	}
	key := SetKey{OperationId: op.OperationId, Role: SetRoleResponse, Detail: success.Status}
	set, ok := t.Selection.Set(key)
	if !ok {
		return MaterializedSet{}, fmt.Errorf("scenario %q: no selected response example for %s",
			t.ArtifactName, op.OperationId)
	}
	return MaterializeSet(set, success.Schema)
}

// updatedState is one update step's materialized request and response.
type updatedState struct {
	request  MaterializedSet
	response MaterializedSet
}

// updateStep materializes an update step, and reports whether the examples
// justify having one. They do only when the update's response differs from the
// create's: an update that changes nothing gives the step nothing to assert.
//
// The returned request is the create request overlaid with the update's own
// values. See MaterializedSet.overlaidWith for why the delta alone cannot
// match what the provider sends.
func (t ResourceTarget) updateStep(createRequest, createResponse MaterializedSet) (updatedState, bool, error) {
	if t.Update == nil {
		return updatedState{}, false, nil
	}
	request, response, err := t.materializePair(t.Update)
	if err != nil {
		// An update the description does not fully describe is not fatal: the
		// scenario simply keeps the create-only flow it can support.
		var incomplete *IncompleteError
		if errors.As(err, &incomplete) {
			return updatedState{}, false, nil
		}
		return updatedState{}, false, err
	}
	if equivalentJSON(response.Body, createResponse.Body) {
		return updatedState{}, false, nil
	}
	// An update whose every value was invented describes no state the author
	// wrote down, so it earns no step: the recorded PATCH would assert values
	// the description never gave. Synthesis exists so a target can be recorded
	// over, not to manufacture a lifecycle nobody described.
	if whollySynthesized(request) || whollySynthesized(response) {
		return updatedState{}, false, nil
	}
	// The recorded request is what the provider will send: the created
	// resource with the update's changes applied, not the example's delta.
	return updatedState{request: createRequest.overlaidWith(request), response: response}, true, nil
}

// ----------------------------------------------------------------------------
// Trace building
// ----------------------------------------------------------------------------

// traceBuilder accumulates the ordered interactions, keeping indexes dense.
type traceBuilder struct {
	target       ResourceTarget
	identity     string
	interactions []model.ScenarioInteraction
}

// appendInteraction is the one place a recorded interaction is shaped. The
// bodyless cases differ in what they carry, not in how it is assembled:
// requestHeaders and responseHeaders already decide Accept-versus-Content-Type
// from whether there is a body, and a zero content length follows from an
// empty one.
func (b *traceBuilder) appendInteraction(
	role model.InteractionRole,
	op *model.Operation,
	requestBody, responseBody string,
	status int,
	provenance []model.ExampleProvenance,
) {
	b.interactions = append(b.interactions, model.ScenarioInteraction{
		Index:       len(b.interactions),
		Role:        role,
		OperationId: op.OperationId,
		Request: model.InteractionRequest{
			Method:        op.Method,
			URL:           b.url(op),
			Body:          requestBody,
			Headers:       requestHeaders(requestBody),
			ContentLength: len(requestBody),
		},
		Response: model.InteractionResponse{
			StatusCode:    status,
			StatusText:    fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:          responseBody,
			Headers:       responseHeaders(responseBody),
			ContentLength: len(responseBody),
		},
		SourceExamples: provenance,
	})
}

// add appends an interaction whose request carries a body.
func (b *traceBuilder) add(
	role model.InteractionRole,
	op *model.Operation,
	request, response MaterializedSet,
	status int,
) {
	b.appendInteraction(role, op,
		encodeBody(request.Body), encodeBody(response.Body), status, b.provenanceFor(op))
}

// addRefreshes appends the reads the framework issues after an apply.
// Identical requests are repeated rather than shared, because a replayed
// interaction is consumed once.
func (b *traceBuilder) addRefreshes(read *model.Operation, response MaterializedSet, times int) {
	for range times {
		role := model.InteractionRoleRefresh
		if len(b.interactions) == 1 {
			// The first read after create is the resource's own Read, not a
			// framework refresh.
			role = model.InteractionRoleRead
		}
		b.addRead(role, read, response, http.StatusOK)
	}
}

// addRead appends a bodyless request returning a representation.
func (b *traceBuilder) addRead(
	role model.InteractionRole,
	op *model.Operation,
	response MaterializedSet,
	status int,
) {
	b.appendInteraction(role, op, "", encodeBody(response.Body), status, b.provenanceFor(op))
}

// addDelete appends the delete, which neither sends nor returns a body. It
// carries no provenance: no declared example contributed a value to it.
func (b *traceBuilder) addDelete() {
	b.appendInteraction(model.InteractionRoleDelete, b.target.Delete,
		"", "", deleteStatus(b.target.Delete), nil)
}

// addDestroyVerification appends the post-destroy read the harness uses to
// confirm the object is gone. It expects a 404 and is why response extraction
// keeps failure outcomes, not only the success one.
func (b *traceBuilder) addDestroyVerification() {
	b.appendInteraction(model.InteractionRoleDestroyVerification, b.target.Read,
		"", "", http.StatusNotFound, nil)
}

// provenanceFor collects the origins of the values this operation's interaction
// carries, so a reviewer can trace recorded bytes back to the description.
func (b *traceBuilder) provenanceFor(op *model.Operation) []model.ExampleProvenance {
	var out []model.ExampleProvenance
	for _, set := range b.target.Selection.Sets {
		if set.Key.OperationId != op.OperationId {
			continue
		}
		out = append(out, provenanceFor(set)...)
	}
	return out
}

// url builds one operation's recorded request target: the resolved server
// origin, the path template, and every path parameter substituted. The identity
// fills the parameter the create response minted; any other path parameter
// takes its selected example.
func (b *traceBuilder) url(op *model.Operation) string {
	path := op.Path
	for i := range op.ParameterExamples {
		parameter := op.ParameterExamples[i]
		if parameter.In != model.ParameterInPath {
			continue
		}
		placeholder := "{" + parameter.Name + "}"
		if !strings.Contains(path, placeholder) {
			continue
		}
		path = strings.ReplaceAll(path, placeholder, b.pathValue(op, parameter))
	}
	return strings.TrimSuffix(b.target.ServerURL, "/") + path
}

// pathValue resolves one path parameter: the minted identity when the
// description gives no example for it, otherwise the selected example.
func (b *traceBuilder) pathValue(op *model.Operation, parameter model.ParameterExamples) string {
	key := SetKey{
		OperationId: op.OperationId,
		Role:        SetRoleParameter,
		Detail:      fmt.Sprintf("%s:%s", parameter.In, parameter.Name),
	}
	if set, ok := b.target.Selection.Set(key); ok && set.Candidate != nil {
		if text, isString := set.Candidate.Value.(string); isString && text != "" {
			return text
		}
	}
	return b.identity
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

const jsonContentType = "application/json"

// configuration turns a materialized request into the configuration a test
// step applies. Config rendering stays in the emitter, which owns the mapping
// from these paths to HCL; this carries only the values.
func (m MaterializedSet) configuration() *model.MaterializedConfiguration {
	return &model.MaterializedConfiguration{
		RequestValues:         m.Values,
		SensitiveReplacements: m.SensitiveReplacements,
	}
}

// responseHeaders returns the retained headers for a response. A bodyless
// outcome carries none, which is what a recorded delete looks like.
func responseHeaders(body string) map[string][]string {
	if body == "" {
		return nil
	}
	return model.FilterRetainedHeaders(map[string][]string{"Content-Type": {jsonContentType}})
}

// requestHeaders returns the retained headers for a request, declaring a
// content type only when there is a body to describe.
func requestHeaders(body string) map[string][]string {
	headers := map[string][]string{"Accept": {jsonContentType}}
	if body != "" {
		headers["Content-Type"] = []string{jsonContentType}
	}
	return model.FilterRetainedHeaders(headers)
}

// encodeBody renders a materialized body as canonical JSON. Go's encoder sorts
// map keys, so the same values always produce the same bytes — which is what
// byte-identical regeneration depends on.
func encodeBody(value any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// whollySynthesized reports a set none of whose values the description
// supplied. Compared as a set rather than by count: a synthesized collection
// records both its own path and its elements', so the two lists are not
// one-to-one.
func whollySynthesized(set MaterializedSet) bool {
	if len(set.Values) == 0 {
		// Nothing described, but nothing invented either — a bodyless
		// operation is legitimate.
		return false
	}
	invented := make(map[string]bool, len(set.SynthesizedPaths))
	for _, path := range set.SynthesizedPaths {
		invented[path] = true
	}
	for _, value := range set.Values {
		if !invented[value.Path] {
			return false
		}
	}
	return true
}

// equivalentJSON reports whether two bodies carry the same values, used to tell
// a real update from one that changes nothing.
func equivalentJSON(left, right any) bool {
	return encodeBody(left) == encodeBody(right)
}

// deleteStatus returns the success status the delete declares, defaulting to
// 204. A delete that declares 200 must record 200: the matcher compares codes.
func deleteStatus(op *model.Operation) int {
	if success := op.SuccessResponseExample(); success != nil {
		var code int
		if _, err := fmt.Sscanf(success.Status, "%d", &code); err == nil && code > 0 {
			return code
		}
	}
	return http.StatusNoContent
}

// resourceTestFuncName builds the generated test's name, which doubles as the
// cassette basename because the harness derives the cassette from t.Name().
func resourceTestFuncName(artifact string) string {
	return "TestAccDatadog" + camelCase(artifact) + "OpenAPIExample"
}

func resourceTestFilePath(artifact string) string {
	return fmt.Sprintf("resource_datadog_%s_openapi_example_test.go", artifact)
}

// camelCase converts a snake_case artifact name to UpperCamelCase.
func camelCase(in string) string {
	var out strings.Builder
	for _, part := range strings.Split(in, "_") {
		if part == "" {
			continue
		}
		out.WriteString(strings.ToUpper(part[:1]))
		out.WriteString(part[1:])
	}
	return out.String()
}

// identityPathLabel names where an id strategy expects to find the identifier.
func identityPathLabel(strategy model.IdStrategy) string {
	if strategy == "" {
		return string(model.IdStrategyDataID)
	}
	return string(strategy)
}
