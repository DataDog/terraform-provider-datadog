package parser

import (
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Raw document context
//
// The normalized model deliberately does not retain libopenapi objects, so
// anything that must read the description again — schema normalization, and
// example extraction after it — needs the raw handles passed to it explicitly.
// RawContext is that channel, and it stays inside the parser: LoadSpec returns
// only the normalized model.Spec, and the phases consuming this write their
// results onto model.Operation. So nothing downstream of LoadSpec receives a
// RawContext, and the only libopenapi handle reaching the model remains the
// documented Spec.Components exception.
// ----------------------------------------------------------------------------

// RawContext carries the parts of the OpenAPI document the normalized model
// drops, keyed by the operations they belong to.
type RawContext struct {
	// Operations maps each projected operation back to the libopenapi
	// operation it came from, reaching request and response bodies.
	Operations map[*model.Operation]*v3.Operation
	// PathItemParams holds the parameters each operation inherits from its
	// path item. The high-level operation does not expose them, so they are
	// captured while the document is walked.
	PathItemParams map[*model.Operation][]*v3.Parameter
	// ServerURL is the resolved default server origin.
	ServerURL string
}

// newRawContext returns an empty context with its maps ready to fill.
func newRawContext() *RawContext {
	return &RawContext{
		Operations:     make(map[*model.Operation]*v3.Operation),
		PathItemParams: make(map[*model.Operation][]*v3.Parameter),
	}
}

// add records one operation's raw handles. Path-item parameters are stored only
// when the path item declares some, keeping the map empty for the common case.
func (c *RawContext) add(op *model.Operation, raw *v3.Operation, pathItem *v3.PathItem) {
	c.Operations[op] = raw
	if pathItem != nil && len(pathItem.Parameters) > 0 {
		c.PathItemParams[op] = pathItem.Parameters
	}
}

// Raw returns the libopenapi operation behind a normalized one, and whether it
// is known.
func (c *RawContext) Raw(op *model.Operation) (*v3.Operation, bool) {
	if c == nil {
		return nil, false
	}
	raw, ok := c.Operations[op]
	return raw, ok
}

// MergedParameters returns an operation's path-item and operation parameters
// merged by (name, location), with the operation's declaration winning. It is
// the one place both schema normalization and example extraction read
// parameters from, so the two cannot disagree about which an operation takes.
func (c *RawContext) MergedParameters(op *model.Operation) []*v3.Parameter {
	raw, ok := c.Raw(op)
	if !ok {
		return nil
	}
	return MergeParameters(c.PathItemParams[op], raw.Parameters)
}

// resolveServerURL returns the first declared server's origin with every
// template variable replaced by its declared default, e.g.
// https://{subdomain}.{site} becomes https://api.datadoghq.com.
//
// Cassette generation records request URLs against this, so it must be the
// origin the generated provider client actually calls: a different host fails
// the replay matcher exactly as a wrong path does. A server whose variables
// have no defaults is left with its placeholders intact rather than guessed
// at, so the mismatch surfaces as an obviously wrong URL instead of a
// plausible one.
func resolveServerURL(servers []*v3.Server) string {
	for _, server := range servers {
		if server == nil || server.URL == "" {
			continue
		}
		url := server.URL
		for name, variable := range server.Variables.FromOldest() {
			if variable == nil || variable.Default == "" {
				continue
			}
			url = strings.ReplaceAll(url, "{"+name+"}", variable.Default)
		}
		if strings.Contains(url, "{") {
			// Unresolved placeholders remain; try the next server, which may
			// be a fully concrete one.
			continue
		}
		return strings.TrimSuffix(url, "/")
	}
	return ""
}
