package cassette

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/yaml.v3"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Cassette rendering
//
// The cassette format belongs to go-vcr, not to this generator, so the fixture
// is built from go-vcr's own types and marshalled with the YAML library go-vcr
// reads them back with. A hand-rolled copy of the format would let a go-vcr
// upgrade silently desync generated fixtures from the replayer; this way the
// mismatch is a compile error instead.
// ----------------------------------------------------------------------------

// cassetteFormatVersion is the on-disk format the provider's recorder reads.
const cassetteFormatVersion = 2

// freezeLayout is how the provider harness parses a freeze file (restoreClock
// reads it with time.RFC3339Nano). A generated cassette has no wall-clock
// origin, so the scenario's fixed UTC instant is written in that form.
const freezeLayout = time.RFC3339Nano

// RenderCassette renders a scenario's interactions as a go-vcr cassette.
//
// Output is a pure function of the scenario: interaction order is the
// scenario's, bodies are already canonical JSON, and every duration is zero.
// Recording a real latency would make regeneration non-deterministic while
// changing nothing about replay.
func RenderCassette(scenario *model.GeneratedTestScenario) ([]byte, error) {
	if err := scenario.Validate(); err != nil {
		return nil, err
	}

	fixture := &cassette.Cassette{
		Version:      cassetteFormatVersion,
		Interactions: make([]*cassette.Interaction, 0, len(scenario.Interactions)),
	}
	for i := range scenario.Interactions {
		interaction, err := renderInteraction(&scenario.Interactions[i])
		if err != nil {
			return nil, fmt.Errorf("cassette %q: %w", scenario.CassetteBaseName, err)
		}
		fixture.Interactions = append(fixture.Interactions, interaction)
	}

	out, err := yaml.Marshal(fixture)
	if err != nil {
		return nil, fmt.Errorf("cassette %q: marshalling: %w", scenario.CassetteBaseName, err)
	}
	// The recorder writes a leading document separator, so a generated fixture
	// is indistinguishable in shape from a recorded one.
	return append([]byte("---\n"), out...), nil
}

// renderInteraction converts one scenario interaction into a recorded pair.
func renderInteraction(in *model.ScenarioInteraction) (*cassette.Interaction, error) {
	host, err := requestHost(in.Request.URL)
	if err != nil {
		return nil, fmt.Errorf("interaction %d: %w", in.Index, err)
	}
	return &cassette.Interaction{
		ID: in.Index,
		Request: cassette.Request{
			// A provider request goes out over HTTP/1.1 and the response comes
			// back over HTTP/2, mirroring what the recorder captures, so a
			// generated fixture deserializes exactly as a recorded one does.
			Proto:            "HTTP/1.1",
			ProtoMajor:       1,
			ProtoMinor:       1,
			ContentLength:    int64(len(in.Request.Body)),
			TransferEncoding: []string{},
			Trailer:          http.Header{},
			Host:             host,
			Body:             in.Request.Body,
			Form:             map[string][]string{},
			Headers:          header(in.Request.Headers),
			URL:              in.Request.URL,
			Method:           in.Request.Method,
		},
		Response: cassette.Response{
			Proto:            "HTTP/2.0",
			ProtoMajor:       2,
			ProtoMinor:       0,
			TransferEncoding: []string{},
			Trailer:          http.Header{},
			// A recorded response reports an unknown length and an
			// already-decompressed body; a bodyless one reports neither.
			ContentLength: responseContentLength(in.Response.Body),
			Uncompressed:  in.Response.Body != "",
			Body:          in.Response.Body,
			Headers:       header(in.Response.Headers),
			Status:        in.Response.StatusText,
			Code:          in.Response.StatusCode,
			Duration:      0,
		},
	}, nil
}

// responseContentLength reports an unknown length for a body the recorder would
// have streamed, and zero for a response that declares none.
func responseContentLength(body string) int64 {
	if body == "" {
		return 0
	}
	return -1
}

// header converts retained headers into the canonical http.Header form go-vcr
// serializes. An absent set becomes an empty map rather than null, which is
// what a recorded interaction carries.
func header(in map[string][]string) http.Header {
	out := http.Header{}
	for name, values := range in {
		out[http.CanonicalHeaderKey(name)] = values
	}
	return out
}

// requestHost extracts the host a recorded request carries alongside its URL.
func requestHost(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing request URL %q: %w", rawURL, err)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("request URL %q carries no host", rawURL)
	}
	return parsed.Host, nil
}

// RenderFreeze renders the cassette's time companion: the single fixed instant
// the replay harness restores, so a generated unique name is reproducible
// without a recording.
func RenderFreeze(scenario *model.GeneratedTestScenario) ([]byte, error) {
	if err := scenario.Validate(); err != nil {
		return nil, err
	}
	return []byte(scenario.FreezeTime.UTC().Format(freezeLayout)), nil
}
