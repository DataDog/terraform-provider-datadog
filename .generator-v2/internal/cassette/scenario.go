package cassette

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

// defaultRefreshesAfterApply is how many reads the test framework issues after
// an apply step: one to refresh state, one for the follow-up plan.
//
// This number is framework behavior, not something the description implies, and
// it is the least certain part of the scenario. Surveying the 344 single-
// resource CRUD cassettes already in this repository, a create-only flow
// records either five or six interactions for what is logically the same
// sequence, and a one-update flow either nine or ten. The extra read comes from
// a test's own check functions calling the API, which a generated test does not
// do — its checks read state only, because the cassette is the oracle.
//
// So two is the right default for the shape this generator emits, and offline
// replay is the arbiter: if the pinned framework disagrees, this one constant
// is what changes.
//
// Letting the recorder replay an interaction more than once looks like it would
// make the count irrelevant. It does not, and the idea is a trap worth naming
// here so it is not retried. go-vcr's lookup is
//
//	for _, i := range c.Interactions {
//		if (c.ReplayableInteractions || !i.replayed) && c.Matcher(r, i.Request) {
//
// so setting ReplayableInteractions bypasses the consumed check and the lookup
// always returns the *first* match, never advancing. Every read in this trace
// shares a method and URL while returning a different body — create state,
// then updated state, then a 404 once destroyed — so with repeats enabled the
// post-update refresh serves stale state and the destroy verification can
// never reach its 404. Measured against a recorded cassette whose reads are
// 200, 200, 404: strict playback yields exactly that, and repeats yield
// 200, 200, 200.
//
// The count being predictable therefore matters, and the deliberately
// duplicated refreshes are load-bearing rather than redundant.
const defaultRefreshesAfterApply = 2

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
	// RefreshesAfterApply overrides defaultRefreshesAfterApply. Zero means the
	// default.
	RefreshesAfterApply int
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

	builder := &traceBuilder{target: target, identity: identity}

	// Create, then the reads the framework issues after an apply.
	builder.add(model.InteractionRoleCreate, target.Create, createRequest, createResponse, http.StatusCreated)
	builder.addRefreshes(target.Read, readResponse)
	scenario.Steps = append(scenario.Steps, model.ScenarioStep{State: createRequest.configuration()})

	// An update step, only when the examples give a distinct state to assert.
	if updated, has, err := target.updateStep(createResponse); err != nil {
		return nil, err
	} else if has {
		// The framework refreshes before applying the second step.
		builder.addRefreshes(target.Read, readResponse, 1)
		builder.add(model.InteractionRoleUpdate, target.Update, updated.request, updated.response, http.StatusOK)
		builder.addRefreshes(target.Read, updated.response)
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

// refreshCount returns how many reads follow an apply.
func (t ResourceTarget) refreshCount() int {
	if t.RefreshesAfterApply > 0 {
		return t.RefreshesAfterApply
	}
	return defaultRefreshesAfterApply
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
func (t ResourceTarget) updateStep(createResponse MaterializedSet) (updatedState, bool, error) {
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
	return updatedState{request: request, response: response}, true, nil
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

func (b *traceBuilder) add(
	role model.InteractionRole,
	op *model.Operation,
	request, response MaterializedSet,
	status int,
) {
	body := encodeBody(request.Body)
	responseBody := encodeBody(response.Body)
	b.interactions = append(b.interactions, model.ScenarioInteraction{
		Index:       len(b.interactions),
		Role:        role,
		OperationId: op.OperationId,
		Request: model.InteractionRequest{
			Method:        op.Method,
			URL:           b.url(op),
			Body:          body,
			Headers:       requestHeaders(body),
			ContentLength: len(body),
		},
		Response: model.InteractionResponse{
			StatusCode:    status,
			StatusText:    fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:          responseBody,
			Headers:       model.FilterRetainedHeaders(map[string][]string{"Content-Type": {jsonContentType}}),
			ContentLength: len(responseBody),
		},
		SourceExamples: b.provenanceFor(op),
	})
}

// addRefreshes appends the reads the framework issues after an apply.
// Identical requests are repeated rather than shared, because a replayed
// interaction is consumed once.
func (b *traceBuilder) addRefreshes(read *model.Operation, response MaterializedSet, count ...int) {
	times := b.target.refreshCount()
	if len(count) == 1 {
		times = count[0]
	}
	for i := 0; i < times; i++ {
		role := model.InteractionRoleRefresh
		if len(b.interactions) == 1 {
			// The first read after create is the resource's own Read, not a
			// framework refresh.
			role = model.InteractionRoleRead
		}
		b.addRead(role, read, response, http.StatusOK)
	}
}

func (b *traceBuilder) addRead(
	role model.InteractionRole,
	op *model.Operation,
	response MaterializedSet,
	status int,
) {
	responseBody := encodeBody(response.Body)
	b.interactions = append(b.interactions, model.ScenarioInteraction{
		Index:       len(b.interactions),
		Role:        role,
		OperationId: op.OperationId,
		Request: model.InteractionRequest{
			Method:  op.Method,
			URL:     b.url(op),
			Headers: model.FilterRetainedHeaders(map[string][]string{"Accept": {jsonContentType}}),
		},
		Response: model.InteractionResponse{
			StatusCode:    status,
			StatusText:    fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:          responseBody,
			Headers:       model.FilterRetainedHeaders(map[string][]string{"Content-Type": {jsonContentType}}),
			ContentLength: len(responseBody),
		},
		SourceExamples: b.provenanceFor(op),
	})
}

func (b *traceBuilder) addDelete() {
	status := deleteStatus(b.target.Delete)
	b.interactions = append(b.interactions, model.ScenarioInteraction{
		Index:       len(b.interactions),
		Role:        model.InteractionRoleDelete,
		OperationId: b.target.Delete.OperationId,
		Request: model.InteractionRequest{
			Method:  b.target.Delete.Method,
			URL:     b.url(b.target.Delete),
			Headers: model.FilterRetainedHeaders(map[string][]string{"Accept": {jsonContentType}}),
		},
		Response: model.InteractionResponse{
			StatusCode: status,
			StatusText: fmt.Sprintf("%d %s", status, http.StatusText(status)),
		},
	})
}

// addDestroyVerification appends the post-destroy read the harness uses to
// confirm the object is gone. It expects a 404 and is why response extraction
// keeps failure outcomes, not only the success one.
func (b *traceBuilder) addDestroyVerification() {
	b.interactions = append(b.interactions, model.ScenarioInteraction{
		Index:       len(b.interactions),
		Role:        model.InteractionRoleDestroyVerification,
		OperationId: b.target.Read.OperationId,
		Request: model.InteractionRequest{
			Method:  b.target.Read.Method,
			URL:     b.url(b.target.Read),
			Headers: model.FilterRetainedHeaders(map[string][]string{"Accept": {jsonContentType}}),
		},
		Response: model.InteractionResponse{
			StatusCode: http.StatusNotFound,
			StatusText: "404 Not Found",
		},
	})
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
	// resolveServerURL already trimmed any trailing slash, so the origin is
	// concatenated as given — one place owns that rule.
	return b.target.ServerURL + path
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

// equivalentJSON reports whether two bodies carry the same values, used to tell
// a real update from one that changes nothing.
func equivalentJSON(left, right any) bool {
	return encodeBody(left) == encodeBody(right)
}

// deleteStatus returns the success status the delete declares, defaulting to
// 204. A delete that declares 200 must record 200: the matcher compares codes.
func deleteStatus(op *model.Operation) int {
	if success := op.SuccessResponseExample(); success != nil {
		// Parsed the same way model.SuccessResponseExample parses it, so the
		// two cannot disagree about what counts as a numeric status.
		if code, err := strconv.Atoi(success.Status); err == nil && code > 0 {
			return code
		}
	}
	return http.StatusNoContent
}

// resourceTestFuncName builds the generated test's name, which doubles as the
// cassette basename because the harness derives the cassette from t.Name().
func resourceTestFuncName(artifact string) string {
	return "TestAccDatadog" + model.SdkName(artifact) + "OpenAPIExample"
}

func resourceTestFilePath(artifact string) string {
	return fmt.Sprintf("resource_datadog_%s_openapi_example_test.go", artifact)
}

// identityPathLabel names where an id strategy expects to find the identifier.
func identityPathLabel(strategy model.IdStrategy) string {
	if strategy == "" {
		return string(model.IdStrategyDataID)
	}
	return string(strategy)
}
